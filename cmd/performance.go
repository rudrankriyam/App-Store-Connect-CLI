package cmd

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

// PerformanceCommand returns the performance command group.
func PerformanceCommand() *ffcli.Command {
	fs := flag.NewFlagSet("performance", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "performance",
		ShortUsage: "asc performance <subcommand> [flags]",
		ShortHelp:  "Access performance metrics and diagnostic logs.",
		LongHelp: `Access performance metrics and diagnostic logs.

Examples:
  asc performance metrics list --app "APP_ID"
  asc performance metrics get --build "BUILD_ID"
  asc performance diagnostics list --build "BUILD_ID"
  asc performance diagnostics get --id "SIGNATURE_ID"
  asc performance download --build "BUILD_ID" --output ./metrics.json`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			PerformanceMetricsCommand(),
			PerformanceDiagnosticsCommand(),
			PerformanceDownloadCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// PerformanceMetricsCommand returns the metrics subcommand group.
func PerformanceMetricsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("metrics", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "metrics",
		ShortUsage: "asc performance metrics <subcommand> [flags]",
		ShortHelp:  "Work with performance/power metrics.",
		LongHelp: `Work with performance/power metrics.

Examples:
  asc performance metrics list --app "APP_ID"
  asc performance metrics get --build "BUILD_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			PerformanceMetricsListCommand(),
			PerformanceMetricsGetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// PerformanceMetricsListCommand returns the metrics list subcommand.
func PerformanceMetricsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("metrics list", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID)")
	platform := fs.String("platform", "", "Platform filter (IOS)")
	metricType := fs.String("metric-type", "", "Metric types (comma-separated: "+strings.Join(perfPowerMetricTypeList(), ", ")+")")
	deviceType := fs.String("device-type", "", "Device types (comma-separated, e.g., iPhone15,2)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc performance metrics list --app \"APP_ID\"",
		ShortHelp:  "List performance/power metrics for an app.",
		LongHelp: `List performance/power metrics for an app.

Examples:
  asc performance metrics list --app "APP_ID"
  asc performance metrics list --app "APP_ID" --metric-type "MEMORY,DISK" --device-type "iPhone15,2"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required (or set ASC_APP_ID)")
				return flag.ErrHelp
			}

			platforms, err := normalizePerfPowerMetricPlatforms(splitCSVUpper(*platform), "--platform")
			if err != nil {
				return fmt.Errorf("performance metrics list: %w", err)
			}
			metricTypes, err := normalizePerfPowerMetricTypes(splitCSVUpper(*metricType))
			if err != nil {
				return fmt.Errorf("performance metrics list: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("performance metrics list: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resp, err := client.GetPerfPowerMetricsForApp(requestCtx, resolvedAppID,
				asc.WithPerfPowerMetricsPlatforms(platforms),
				asc.WithPerfPowerMetricsMetricTypes(metricTypes),
				asc.WithPerfPowerMetricsDeviceTypes(splitCSV(*deviceType)),
			)
			if err != nil {
				return fmt.Errorf("performance metrics list: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// PerformanceMetricsGetCommand returns the metrics get subcommand.
func PerformanceMetricsGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("metrics get", flag.ExitOnError)

	buildID := fs.String("build", "", "Build ID to fetch metrics for")
	platform := fs.String("platform", "", "Platform filter (IOS)")
	metricType := fs.String("metric-type", "", "Metric types (comma-separated: "+strings.Join(perfPowerMetricTypeList(), ", ")+")")
	deviceType := fs.String("device-type", "", "Device types (comma-separated, e.g., iPhone15,2)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc performance metrics get --build \"BUILD_ID\"",
		ShortHelp:  "Get performance/power metrics for a build.",
		LongHelp: `Get performance/power metrics for a build.

Examples:
  asc performance metrics get --build "BUILD_ID"
  asc performance metrics get --build "BUILD_ID" --metric-type "MEMORY" --device-type "iPhone15,2"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			trimmedBuildID := strings.TrimSpace(*buildID)
			if trimmedBuildID == "" {
				fmt.Fprintln(os.Stderr, "Error: --build is required")
				return flag.ErrHelp
			}

			platforms, err := normalizePerfPowerMetricPlatforms(splitCSVUpper(*platform), "--platform")
			if err != nil {
				return fmt.Errorf("performance metrics get: %w", err)
			}
			metricTypes, err := normalizePerfPowerMetricTypes(splitCSVUpper(*metricType))
			if err != nil {
				return fmt.Errorf("performance metrics get: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("performance metrics get: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resp, err := client.GetPerfPowerMetricsForBuild(requestCtx, trimmedBuildID,
				asc.WithPerfPowerMetricsPlatforms(platforms),
				asc.WithPerfPowerMetricsMetricTypes(metricTypes),
				asc.WithPerfPowerMetricsDeviceTypes(splitCSV(*deviceType)),
			)
			if err != nil {
				return fmt.Errorf("performance metrics get: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// PerformanceDiagnosticsCommand returns the diagnostics subcommand group.
func PerformanceDiagnosticsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("diagnostics", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "diagnostics",
		ShortUsage: "asc performance diagnostics <subcommand> [flags]",
		ShortHelp:  "Work with diagnostic signatures and logs.",
		LongHelp: `Work with diagnostic signatures and logs.

Examples:
  asc performance diagnostics list --build "BUILD_ID"
  asc performance diagnostics get --id "SIGNATURE_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			PerformanceDiagnosticsListCommand(),
			PerformanceDiagnosticsGetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// PerformanceDiagnosticsListCommand returns the diagnostics list subcommand.
func PerformanceDiagnosticsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("diagnostics list", flag.ExitOnError)

	buildID := fs.String("build", "", "Build ID to list diagnostics for")
	diagnosticType := fs.String("diagnostic-type", "", "Diagnostic type filter (comma-separated: "+strings.Join(diagnosticSignatureTypeList(), ", ")+")")
	fields := fs.String("fields", "", "Fields to return (comma-separated: "+strings.Join(diagnosticSignatureFieldList(), ", ")+")")
	limit := fs.Int("limit", 0, "Limit number of signatures (max 200)")
	next := fs.String("next", "", "Next page URL")
	paginate := fs.Bool("paginate", false, "Fetch all pages")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc performance diagnostics list --build \"BUILD_ID\"",
		ShortHelp:  "List diagnostic signatures for a build.",
		LongHelp: `List diagnostic signatures for a build.

Examples:
  asc performance diagnostics list --build "BUILD_ID"
  asc performance diagnostics list --build "BUILD_ID" --diagnostic-type "HANGS" --limit 50`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			trimmedBuildID := strings.TrimSpace(*buildID)
			if trimmedBuildID == "" {
				fmt.Fprintln(os.Stderr, "Error: --build is required")
				return flag.ErrHelp
			}
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("performance diagnostics list: --limit must be between 1 and 200")
			}

			diagnosticTypes, err := normalizeDiagnosticSignatureTypes(splitCSVUpper(*diagnosticType))
			if err != nil {
				return fmt.Errorf("performance diagnostics list: %w", err)
			}
			fieldValues, err := normalizeDiagnosticSignatureFields(*fields)
			if err != nil {
				return fmt.Errorf("performance diagnostics list: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("performance diagnostics list: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.DiagnosticSignaturesOption{
				asc.WithDiagnosticSignaturesDiagnosticTypes(diagnosticTypes),
				asc.WithDiagnosticSignaturesFields(fieldValues),
				asc.WithDiagnosticSignaturesLimit(*limit),
				asc.WithDiagnosticSignaturesNextURL(*next),
			}

			if *paginate {
				paginateOpts := append(opts, asc.WithDiagnosticSignaturesLimit(200))
				firstPage, err := client.GetDiagnosticSignaturesForBuild(requestCtx, trimmedBuildID, paginateOpts...)
				if err != nil {
					return fmt.Errorf("performance diagnostics list: failed to fetch: %w", err)
				}

				paginated, err := asc.PaginateAll(requestCtx, firstPage, func(ctx context.Context, nextURL string) (asc.PaginatedResponse, error) {
					return client.GetDiagnosticSignaturesForBuild(ctx, trimmedBuildID, asc.WithDiagnosticSignaturesNextURL(nextURL))
				})
				if err != nil {
					return fmt.Errorf("performance diagnostics list: %w", err)
				}

				return printOutput(paginated, *output, *pretty)
			}

			resp, err := client.GetDiagnosticSignaturesForBuild(requestCtx, trimmedBuildID, opts...)
			if err != nil {
				return fmt.Errorf("performance diagnostics list: failed to fetch: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// PerformanceDiagnosticsGetCommand returns the diagnostics get subcommand.
func PerformanceDiagnosticsGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("diagnostics get", flag.ExitOnError)

	signatureID := fs.String("id", "", "Diagnostic signature ID")
	limit := fs.Int("limit", 0, "Limit number of logs (max 200)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc performance diagnostics get --id \"SIGNATURE_ID\"",
		ShortHelp:  "Get diagnostic logs for a signature.",
		LongHelp: `Get diagnostic logs for a signature.

Examples:
  asc performance diagnostics get --id "SIGNATURE_ID"
  asc performance diagnostics get --id "SIGNATURE_ID" --limit 50`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			trimmedID := strings.TrimSpace(*signatureID)
			if trimmedID == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("performance diagnostics get: --limit must be between 1 and 200")
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("performance diagnostics get: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			resp, err := client.GetDiagnosticSignatureLogs(requestCtx, trimmedID, asc.WithDiagnosticLogsLimit(*limit))
			if err != nil {
				return fmt.Errorf("performance diagnostics get: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// PerformanceDownloadCommand returns the download subcommand.
func PerformanceDownloadCommand() *ffcli.Command {
	fs := flag.NewFlagSet("download", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID)")
	buildID := fs.String("build", "", "Build ID to download metrics for")
	diagnosticID := fs.String("diagnostic-id", "", "Diagnostic signature ID to download logs for")
	platform := fs.String("platform", "", "Platform filter (IOS)")
	metricType := fs.String("metric-type", "", "Metric types (comma-separated: "+strings.Join(perfPowerMetricTypeList(), ", ")+")")
	deviceType := fs.String("device-type", "", "Device types (comma-separated, e.g., iPhone15,2)")
	limit := fs.Int("limit", 0, "Limit number of logs (max 200, diagnostic logs only)")
	output := fs.String("output", "", "Output file path (default: metrics/diagnostic file name)")
	decompress := fs.Bool("decompress", false, "Decompress gzip output (if compressed)")
	outputFormat := fs.String("output-format", "json", "Output format for metadata: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "download",
		ShortUsage: "asc performance download [flags]",
		ShortHelp:  "Download metrics or diagnostic logs.",
		LongHelp: `Download metrics or diagnostic logs.

Examples:
  asc performance download --app "APP_ID" --output ./metrics.json
  asc performance download --build "BUILD_ID" --output ./metrics.json
  asc performance download --diagnostic-id "SIGNATURE_ID" --output ./diagnostic.json --decompress`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			appFlag := strings.TrimSpace(*appID)
			trimmedBuildID := strings.TrimSpace(*buildID)
			trimmedDiagnosticID := strings.TrimSpace(*diagnosticID)

			selectionCount := 0
			if appFlag != "" {
				selectionCount++
			}
			if trimmedBuildID != "" {
				selectionCount++
			}
			if trimmedDiagnosticID != "" {
				selectionCount++
			}
			if selectionCount == 0 {
				appFlag = resolveAppID(*appID)
				if appFlag == "" {
					fmt.Fprintln(os.Stderr, "Error: --app, --build, or --diagnostic-id is required")
					return flag.ErrHelp
				}
				selectionCount = 1
			}
			if selectionCount > 1 {
				return fmt.Errorf("performance download: --app, --build, and --diagnostic-id are mutually exclusive")
			}
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("performance download: --limit must be between 1 and 200")
			}
			if trimmedDiagnosticID != "" && (strings.TrimSpace(*platform) != "" || strings.TrimSpace(*metricType) != "" || strings.TrimSpace(*deviceType) != "") {
				return fmt.Errorf("performance download: metric filters are not valid with --diagnostic-id")
			}
			if trimmedDiagnosticID == "" && *limit > 0 {
				return fmt.Errorf("performance download: --limit is only valid with --diagnostic-id")
			}

			platforms, err := normalizePerfPowerMetricPlatforms(splitCSVUpper(*platform), "--platform")
			if err != nil {
				return fmt.Errorf("performance download: %w", err)
			}
			metricTypes, err := normalizePerfPowerMetricTypes(splitCSVUpper(*metricType))
			if err != nil {
				return fmt.Errorf("performance download: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("performance download: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			switch {
			case trimmedDiagnosticID != "":
				defaultOutput := fmt.Sprintf("diagnostic_logs_%s.json", trimmedDiagnosticID)

				download, err := client.DownloadDiagnosticSignatureLogs(requestCtx, trimmedDiagnosticID, asc.WithDiagnosticLogsLimit(*limit))
				if err != nil {
					return fmt.Errorf("performance download: %w", err)
				}
				defer download.Body.Close()

				reader, isGzip, err := preparePerformanceDownloadReader(download.Body, *decompress)
				if err != nil {
					return fmt.Errorf("performance download: %w", err)
				}
				shouldDecompress := *decompress && isGzip
				compressedPath, decompressedPath := shared.ResolveReportOutputPaths(*output, defaultOutput, ".json", shouldDecompress)

				compressedSize, err := shared.WriteStreamToFile(compressedPath, reader)
				if err != nil {
					return fmt.Errorf("performance download: %w", err)
				}

				var decompressedSize int64
				if shouldDecompress {
					decompressedSize, err = shared.DecompressGzipFile(compressedPath, decompressedPath)
					if err != nil {
						return fmt.Errorf("performance download: %w", err)
					}
				}

				result := &asc.PerformanceDownloadResult{
					DownloadType:          "diagnostic-logs",
					DiagnosticSignatureID: trimmedDiagnosticID,
					FilePath:              compressedPath,
					FileSize:              compressedSize,
					Decompressed:          shouldDecompress,
					DecompressedPath:      decompressedPath,
					DecompressedSize:      decompressedSize,
				}

				return printOutput(result, *outputFormat, *pretty)
			case trimmedBuildID != "":
				defaultOutput := fmt.Sprintf("perf_power_metrics_%s.json", trimmedBuildID)

				download, err := client.DownloadPerfPowerMetricsForBuild(requestCtx, trimmedBuildID,
					asc.WithPerfPowerMetricsPlatforms(platforms),
					asc.WithPerfPowerMetricsMetricTypes(metricTypes),
					asc.WithPerfPowerMetricsDeviceTypes(splitCSV(*deviceType)),
				)
				if err != nil {
					return fmt.Errorf("performance download: %w", err)
				}
				defer download.Body.Close()

				reader, isGzip, err := preparePerformanceDownloadReader(download.Body, *decompress)
				if err != nil {
					return fmt.Errorf("performance download: %w", err)
				}
				shouldDecompress := *decompress && isGzip
				compressedPath, decompressedPath := shared.ResolveReportOutputPaths(*output, defaultOutput, ".json", shouldDecompress)

				compressedSize, err := shared.WriteStreamToFile(compressedPath, reader)
				if err != nil {
					return fmt.Errorf("performance download: %w", err)
				}

				var decompressedSize int64
				if shouldDecompress {
					decompressedSize, err = shared.DecompressGzipFile(compressedPath, decompressedPath)
					if err != nil {
						return fmt.Errorf("performance download: %w", err)
					}
				}

				result := &asc.PerformanceDownloadResult{
					DownloadType:     "metrics",
					BuildID:          trimmedBuildID,
					FilePath:         compressedPath,
					FileSize:         compressedSize,
					Decompressed:     shouldDecompress,
					DecompressedPath: decompressedPath,
					DecompressedSize: decompressedSize,
				}

				return printOutput(result, *outputFormat, *pretty)
			default:
				defaultOutput := fmt.Sprintf("perf_power_metrics_%s.json", appFlag)

				download, err := client.DownloadPerfPowerMetricsForApp(requestCtx, appFlag,
					asc.WithPerfPowerMetricsPlatforms(platforms),
					asc.WithPerfPowerMetricsMetricTypes(metricTypes),
					asc.WithPerfPowerMetricsDeviceTypes(splitCSV(*deviceType)),
				)
				if err != nil {
					return fmt.Errorf("performance download: %w", err)
				}
				defer download.Body.Close()

				reader, isGzip, err := preparePerformanceDownloadReader(download.Body, *decompress)
				if err != nil {
					return fmt.Errorf("performance download: %w", err)
				}
				shouldDecompress := *decompress && isGzip
				compressedPath, decompressedPath := shared.ResolveReportOutputPaths(*output, defaultOutput, ".json", shouldDecompress)

				compressedSize, err := shared.WriteStreamToFile(compressedPath, reader)
				if err != nil {
					return fmt.Errorf("performance download: %w", err)
				}

				var decompressedSize int64
				if shouldDecompress {
					decompressedSize, err = shared.DecompressGzipFile(compressedPath, decompressedPath)
					if err != nil {
						return fmt.Errorf("performance download: %w", err)
					}
				}

				result := &asc.PerformanceDownloadResult{
					DownloadType:     "metrics",
					AppID:            appFlag,
					FilePath:         compressedPath,
					FileSize:         compressedSize,
					Decompressed:     shouldDecompress,
					DecompressedPath: decompressedPath,
					DecompressedSize: decompressedSize,
				}

				return printOutput(result, *outputFormat, *pretty)
			}
		},
	}
}

var perfPowerMetricTypes = map[string]struct{}{
	string(asc.PerfPowerMetricTypeDisk):        {},
	string(asc.PerfPowerMetricTypeHang):        {},
	string(asc.PerfPowerMetricTypeBattery):     {},
	string(asc.PerfPowerMetricTypeLaunch):      {},
	string(asc.PerfPowerMetricTypeMemory):      {},
	string(asc.PerfPowerMetricTypeAnimation):   {},
	string(asc.PerfPowerMetricTypeTermination): {},
}

func perfPowerMetricTypeList() []string {
	return []string{
		string(asc.PerfPowerMetricTypeAnimation),
		string(asc.PerfPowerMetricTypeBattery),
		string(asc.PerfPowerMetricTypeDisk),
		string(asc.PerfPowerMetricTypeHang),
		string(asc.PerfPowerMetricTypeLaunch),
		string(asc.PerfPowerMetricTypeMemory),
		string(asc.PerfPowerMetricTypeTermination),
	}
}

func normalizePerfPowerMetricTypes(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	for _, value := range values {
		if _, ok := perfPowerMetricTypes[value]; !ok {
			return nil, fmt.Errorf("--metric-type must be one of: %s", strings.Join(perfPowerMetricTypeList(), ", "))
		}
	}
	return values, nil
}

var perfPowerMetricPlatforms = map[string]struct{}{
	string(asc.PlatformIOS): {},
}

func perfPowerMetricPlatformList() []string {
	return []string{
		string(asc.PlatformIOS),
	}
}

func normalizePerfPowerMetricPlatforms(values []string, flagName string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	for _, value := range values {
		if _, ok := perfPowerMetricPlatforms[value]; !ok {
			return nil, fmt.Errorf("%s must be one of: %s", flagName, strings.Join(perfPowerMetricPlatformList(), ", "))
		}
	}
	return values, nil
}

var diagnosticSignatureTypes = map[string]struct{}{
	string(asc.DiagnosticSignatureTypeDiskWrites): {},
	string(asc.DiagnosticSignatureTypeHangs):      {},
	string(asc.DiagnosticSignatureTypeLaunches):   {},
}

func diagnosticSignatureTypeList() []string {
	return []string{
		string(asc.DiagnosticSignatureTypeDiskWrites),
		string(asc.DiagnosticSignatureTypeHangs),
		string(asc.DiagnosticSignatureTypeLaunches),
	}
}

func normalizeDiagnosticSignatureTypes(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	for _, value := range values {
		if _, ok := diagnosticSignatureTypes[value]; !ok {
			return nil, fmt.Errorf("--diagnostic-type must be one of: %s", strings.Join(diagnosticSignatureTypeList(), ", "))
		}
	}
	return values, nil
}

func preparePerformanceDownloadReader(reader io.Reader, decompress bool) (io.Reader, bool, error) {
	if !decompress {
		return reader, false, nil
	}

	buffered := bufio.NewReader(reader)
	header, err := buffered.Peek(2)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, false, err
	}
	isGzip := len(header) >= 2 && header[0] == 0x1f && header[1] == 0x8b
	return buffered, isGzip, nil
}

func diagnosticSignatureFieldList() []string {
	return []string{
		"diagnosticType",
		"signature",
		"weight",
		"insight",
		"logs",
	}
}

func normalizeDiagnosticSignatureFields(value string) ([]string, error) {
	fields := splitCSV(value)
	if len(fields) == 0 {
		return nil, nil
	}

	allowed := map[string]struct{}{}
	for _, field := range diagnosticSignatureFieldList() {
		allowed[field] = struct{}{}
	}
	for _, field := range fields {
		if _, ok := allowed[field]; !ok {
			return nil, fmt.Errorf("--fields must be one of: %s", strings.Join(diagnosticSignatureFieldList(), ", "))
		}
	}

	return fields, nil
}
