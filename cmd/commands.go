package cmd

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/auth"
)

// ANSI escape codes for bold text
var (
	bold  = "\033[1m"
	reset = "\033[22m"
)

// Bold returns the string wrapped in ANSI bold codes
func Bold(s string) string {
	return bold + s + reset
}

// DefaultUsageFunc returns a usage string with bold section headers
func DefaultUsageFunc(c *ffcli.Command) string {
	var b strings.Builder

	shortHelp := strings.TrimSpace(c.ShortHelp)
	longHelp := strings.TrimSpace(c.LongHelp)
	if shortHelp == "" && longHelp != "" {
		shortHelp = longHelp
		longHelp = ""
	}

	// DESCRIPTION
	if shortHelp != "" {
		b.WriteString(Bold("DESCRIPTION"))
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(shortHelp)
		b.WriteString("\n\n")
	}

	// USAGE / ShortUsage
	usage := strings.TrimSpace(c.ShortUsage)
	if usage == "" {
		usage = strings.TrimSpace(c.Name)
	}
	if usage != "" {
		b.WriteString(Bold("USAGE"))
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(usage)
		b.WriteString("\n\n")
	}

	// LongHelp (additional description)
	if longHelp != "" {
		if shortHelp != "" && strings.HasPrefix(longHelp, shortHelp) {
			longHelp = strings.TrimSpace(strings.TrimPrefix(longHelp, shortHelp))
		}
		if longHelp != "" {
			b.WriteString(longHelp)
			b.WriteString("\n\n")
		}
	}

	// SUBCOMMANDS
	if len(c.Subcommands) > 0 {
		b.WriteString(Bold("SUBCOMMANDS"))
		b.WriteString("\n")
		tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
		for _, sub := range c.Subcommands {
			fmt.Fprintf(tw, "  %-12s %s\n", sub.Name, sub.ShortHelp)
		}
		tw.Flush()
		b.WriteString("\n")
	}

	// FLAGS
	if c.FlagSet != nil {
		hasFlags := false
		c.FlagSet.VisitAll(func(*flag.Flag) {
			hasFlags = true
		})
		if hasFlags {
			b.WriteString(Bold("FLAGS"))
			b.WriteString("\n")
			tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
			c.FlagSet.VisitAll(func(f *flag.Flag) {
				def := f.DefValue
				if def != "" {
					fmt.Fprintf(tw, "  --%-12s %s (default: %s)\n", f.Name, f.Usage, def)
					return
				}
				fmt.Fprintf(tw, "  --%-12s %s\n", f.Name, f.Usage)
			})
			tw.Flush()
			b.WriteString("\n")
		}
	}

	return b.String()
}

// Feedback command factory
func FeedbackCommand() *ffcli.Command {
	fs := flag.NewFlagSet("feedback", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	includeScreenshots := fs.Bool("include-screenshots", false, "Include screenshot URLs in feedback output")
	deviceModel := fs.String("device-model", "", "Filter by device model(s), comma-separated")
	osVersion := fs.String("os-version", "", "Filter by OS version(s), comma-separated")
	appPlatform := fs.String("app-platform", "", "Filter by app platform(s), comma-separated (IOS, MAC_OS, TV_OS, VISION_OS)")
	devicePlatform := fs.String("device-platform", "", "Filter by device platform(s), comma-separated (IOS, MAC_OS, TV_OS, VISION_OS)")
	buildID := fs.String("build", "", "Filter by build ID(s), comma-separated")
	buildPreRelease := fs.String("build-pre-release-version", "", "Filter by pre-release version ID(s), comma-separated")
	tester := fs.String("tester", "", "Filter by tester ID(s), comma-separated")
	sort := fs.String("sort", "", "Sort by createdDate or -createdDate")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")

	return &ffcli.Command{
		Name:       "feedback",
		ShortUsage: "asc feedback [flags]",
		ShortHelp:  "List TestFlight feedback from beta testers.",
		LongHelp: `List TestFlight feedback from beta testers.

This command fetches beta feedback screenshot submissions and comments.

Examples:
  asc feedback --app "123456789"
  asc feedback --app "123456789" --include-screenshots
  asc feedback --app "123456789" --device-model "iPhone15,3" --os-version "17.2"
  asc feedback --app "123456789" --sort -createdDate --limit 5
  asc feedback --next "<links.next>"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("feedback: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("feedback: %w", err)
			}
			if err := validateSort(*sort, "createdDate", "-createdDate"); err != nil {
				return fmt.Errorf("feedback: %w", err)
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("feedback: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.FeedbackOption{
				asc.WithFeedbackDeviceModels(splitCSV(*deviceModel)),
				asc.WithFeedbackOSVersions(splitCSV(*osVersion)),
				asc.WithFeedbackAppPlatforms(splitCSVUpper(*appPlatform)),
				asc.WithFeedbackDevicePlatforms(splitCSVUpper(*devicePlatform)),
				asc.WithFeedbackBuildIDs(splitCSV(*buildID)),
				asc.WithFeedbackBuildPreReleaseVersionIDs(splitCSV(*buildPreRelease)),
				asc.WithFeedbackTesterIDs(splitCSV(*tester)),
				asc.WithFeedbackLimit(*limit),
				asc.WithFeedbackNextURL(*next),
			}
			if strings.TrimSpace(*sort) != "" {
				opts = append(opts, asc.WithFeedbackSort(*sort))
			}
			if *includeScreenshots {
				opts = append(opts, asc.WithFeedbackIncludeScreenshots())
			}

			feedback, err := client.GetFeedback(requestCtx, resolvedAppID, opts...)
			if err != nil {
				return fmt.Errorf("feedback: failed to fetch: %w", err)
			}

			format := *output

			return printOutput(feedback, format, *pretty)
		},
	}
}

// Crashes command factory
func CrashesCommand() *ffcli.Command {
	fs := flag.NewFlagSet("crashes", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	deviceModel := fs.String("device-model", "", "Filter by device model(s), comma-separated")
	osVersion := fs.String("os-version", "", "Filter by OS version(s), comma-separated")
	appPlatform := fs.String("app-platform", "", "Filter by app platform(s), comma-separated (IOS, MAC_OS, TV_OS, VISION_OS)")
	devicePlatform := fs.String("device-platform", "", "Filter by device platform(s), comma-separated (IOS, MAC_OS, TV_OS, VISION_OS)")
	buildID := fs.String("build", "", "Filter by build ID(s), comma-separated")
	buildPreRelease := fs.String("build-pre-release-version", "", "Filter by pre-release version ID(s), comma-separated")
	tester := fs.String("tester", "", "Filter by tester ID(s), comma-separated")
	sort := fs.String("sort", "", "Sort by createdDate or -createdDate")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")

	return &ffcli.Command{
		Name:       "crashes",
		ShortUsage: "asc crashes [flags]",
		ShortHelp:  "List and export TestFlight crash reports.",
		LongHelp: `List and export TestFlight crash reports.

This command fetches crash reports submitted by TestFlight beta testers,
helping you identify and fix issues in your app.

Examples:
  asc crashes --app "123456789"
  asc crashes --app "123456789" > crashes.json
  asc crashes --app "123456789" --device-model "iPhone15,3" --os-version "17.2"
  asc crashes --app "123456789" --sort -createdDate --limit 5
  asc crashes --next "<links.next>"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("crashes: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("crashes: %w", err)
			}
			if err := validateSort(*sort, "createdDate", "-createdDate"); err != nil {
				return fmt.Errorf("crashes: %w", err)
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("crashes: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.CrashOption{
				asc.WithCrashDeviceModels(splitCSV(*deviceModel)),
				asc.WithCrashOSVersions(splitCSV(*osVersion)),
				asc.WithCrashAppPlatforms(splitCSVUpper(*appPlatform)),
				asc.WithCrashDevicePlatforms(splitCSVUpper(*devicePlatform)),
				asc.WithCrashBuildIDs(splitCSV(*buildID)),
				asc.WithCrashBuildPreReleaseVersionIDs(splitCSV(*buildPreRelease)),
				asc.WithCrashTesterIDs(splitCSV(*tester)),
				asc.WithCrashLimit(*limit),
				asc.WithCrashNextURL(*next),
			}
			if strings.TrimSpace(*sort) != "" {
				opts = append(opts, asc.WithCrashSort(*sort))
			}

			crashes, err := client.GetCrashes(requestCtx, resolvedAppID, opts...)
			if err != nil {
				return fmt.Errorf("crashes: failed to fetch: %w", err)
			}

			format := *output

			return printOutput(crashes, format, *pretty)
		},
	}
}

// Reviews command factory
func ReviewsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("reviews", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	stars := fs.Int("stars", 0, "Filter by star rating (1-5)")
	territory := fs.String("territory", "", "Filter by territory (e.g., US, GBR)")
	sort := fs.String("sort", "", "Sort by rating, -rating, createdDate, or -createdDate")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")

	return &ffcli.Command{
		Name:       "reviews",
		ShortUsage: "asc reviews [flags]",
		ShortHelp:  "List App Store customer reviews.",
		LongHelp: `List App Store customer reviews.

This command fetches customer reviews from the App Store,
helping you understand user feedback and sentiment.

Examples:
  asc reviews --app "123456789"
  asc reviews --app "123456789" --stars 1 --territory US
  asc reviews --app "123456789" --sort -createdDate --limit 5
  asc reviews --next "<links.next>"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("reviews: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("reviews: %w", err)
			}
			if err := validateSort(*sort, "rating", "-rating", "createdDate", "-createdDate"); err != nil {
				return fmt.Errorf("reviews: %w", err)
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("reviews: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.ReviewOption{}
			if *stars != 0 {
				if *stars < 1 || *stars > 5 {
					return fmt.Errorf("reviews: --stars must be between 1 and 5")
				}
				opts = append(opts, asc.WithRating(*stars))
			}
			if *territory != "" {
				opts = append(opts, asc.WithTerritory(*territory))
			}
			if *limit != 0 {
				opts = append(opts, asc.WithLimit(*limit))
			}
			if strings.TrimSpace(*next) != "" {
				opts = append(opts, asc.WithNextURL(*next))
			}
			if strings.TrimSpace(*sort) != "" {
				opts = append(opts, asc.WithReviewSort(*sort))
			}

			reviews, err := client.GetReviews(requestCtx, resolvedAppID, opts...)
			if err != nil {
				return fmt.Errorf("reviews: failed to fetch: %w", err)
			}

			format := *output

			return printOutput(reviews, format, *pretty)
		},
	}
}

// Apps command factory
func AppsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("apps", flag.ExitOnError)

	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	sort := fs.String("sort", "", "Sort by name or -name, bundleId or -bundleId")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")

	return &ffcli.Command{
		Name:       "apps",
		ShortUsage: "asc apps [flags]",
		ShortHelp:  "List all apps in your App Store Connect account.",
		LongHelp: `List all apps in your App Store Connect account.

This command fetches all apps associated with your API key,
useful for finding app IDs when using other commands.

Examples:
  asc apps
  asc apps --limit 10
  asc apps --sort name
  asc apps --output table
  asc apps --next "<links.next>"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("apps: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("apps: %w", err)
			}
			if err := validateSort(*sort, "name", "-name", "bundleId", "-bundleId"); err != nil {
				return fmt.Errorf("apps: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("apps: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.AppsOption{
				asc.WithAppsLimit(*limit),
				asc.WithAppsNextURL(*next),
			}
			if strings.TrimSpace(*sort) != "" {
				opts = append(opts, asc.WithAppsSort(*sort))
			}

			apps, err := client.GetApps(requestCtx, opts...)
			if err != nil {
				return fmt.Errorf("apps: failed to fetch: %w", err)
			}

			format := *output

			return printOutput(apps, format, *pretty)
		},
	}
}

// BuildsUploadCommand returns a command to upload a build
func BuildsUploadCommand() *ffcli.Command {
	fs := flag.NewFlagSet("upload", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (required, or ASC_APP_ID env)")
	ipaPath := fs.String("ipa", "", "Path to .ipa file (required)")
	version := fs.String("version", "", "CFBundleShortVersionString (e.g., 1.0.0, auto-extracted from IPA if not provided)")
	buildNumber := fs.String("build-number", "", "CFBundleVersion (e.g., 123, auto-extracted from IPA if not provided)")
	platform := fs.String("platform", "IOS", "Platform: IOS, MAC_OS, TV_OS, VISION_OS")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "upload",
		ShortUsage: "asc builds upload [flags]",
		ShortHelp:  "Prepare a build upload in App Store Connect.",
		LongHelp: `Prepare a build upload in App Store Connect.

This command creates a build upload record and reserves upload operations
with presigned URLs. The actual file upload must be done separately.

Examples:
  asc builds upload --app "123456789" --ipa "path/to/app.ipa"
  asc builds upload --ipa "app.ipa" --version "1.0.0" --build-number "123"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			// Validate required flags
			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}
			if *ipaPath == "" {
				fmt.Fprintf(os.Stderr, "Error: --ipa is required\n\n")
				return flag.ErrHelp
			}

			// Validate IPA file exists
			fileInfo, err := os.Stat(*ipaPath)
			if err != nil {
				return fmt.Errorf("builds upload: failed to stat IPA: %w", err)
			}
			if fileInfo.IsDir() {
				return fmt.Errorf("builds upload: --ipa must be a file")
			}

			// Validate platform
			platformValue := asc.Platform(strings.ToUpper(*platform))
			switch platformValue {
			case asc.PlatformIOS, asc.PlatformMacOS, asc.PlatformTVOS, asc.PlatformVisionOS:
			default:
				return fmt.Errorf("builds upload: --platform must be IOS, MAC_OS, TV_OS, or VISION_OS")
			}

			// TODO: Extract version and build number from IPA if not provided
			if *version == "" {
				return fmt.Errorf("builds upload: --version is required (auto-extraction not yet implemented)")
			}
			if *buildNumber == "" {
				return fmt.Errorf("builds upload: --build-number is required (auto-extraction not yet implemented)")
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("builds upload: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			// Step 1: Create build upload record
			uploadReq := asc.BuildUploadCreateRequest{
				Data: asc.BuildUploadCreateData{
					Type: asc.ResourceTypeBuildUploads,
					Attributes: asc.BuildUploadAttributes{
						CFBundleShortVersionString: *version,
						CFBundleVersion:            *buildNumber,
						Platform:                   platformValue,
					},
					Relationships: &asc.BuildUploadRelationships{
						App: &asc.Relationship{
							Data: asc.ResourceData{Type: asc.ResourceTypeApps, ID: resolvedAppID},
						},
					},
				},
			}

			uploadResp, err := client.CreateBuildUpload(requestCtx, uploadReq)
			if err != nil {
				return fmt.Errorf("builds upload: failed to create upload record: %w", err)
			}

			// Step 2: Create build upload file reservation
			fileReq := asc.BuildUploadFileCreateRequest{
				Data: asc.BuildUploadFileCreateData{
					Type: asc.ResourceTypeBuildUploadFiles,
					Attributes: asc.BuildUploadFileAttributes{
						FileName:  fileInfo.Name(),
						FileSize:  fileInfo.Size(),
						UTI:       asc.UTIIPA,
						AssetType: asc.AssetTypeAsset,
					},
					Relationships: &asc.BuildUploadFileRelationships{
						BuildUpload: &asc.Relationship{
							Data: asc.ResourceData{Type: asc.ResourceTypeBuildUploads, ID: uploadResp.Data.ID},
						},
					},
				},
			}

			fileResp, err := client.CreateBuildUploadFile(requestCtx, fileReq)
			if err != nil {
				return fmt.Errorf("builds upload: failed to create file reservation: %w", err)
			}

			// Return upload info including presigned URL operations
			result := &asc.BuildUploadResult{
				UploadID:   uploadResp.Data.ID,
				FileID:     fileResp.Data.ID,
				FileName:   fileResp.Data.Attributes.FileName,
				FileSize:   fileResp.Data.Attributes.FileSize,
				Operations: fileResp.Data.Attributes.UploadOperations,
			}

			format := *output

			return printOutput(result, format, *pretty)
		},
	}
}

// BuildsCommand returns the builds command with subcommands
func BuildsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("builds", flag.ExitOnError)

	// Parent command has no flags - subcommands define their own
	listCmd := BuildsListCommand()

	return &ffcli.Command{
		Name:       "builds",
		ShortUsage: "asc builds <subcommand> [flags]",
		ShortHelp:  "Manage builds in App Store Connect.",
		LongHelp: `Manage builds in App Store Connect.

Examples:
  asc builds list --app "123456789"
  asc builds info --build "BUILD_ID"
  asc builds expire --build "BUILD_ID"
  asc builds upload --app "123456789" --ipa "app.ipa"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			listCmd,
			BuildsInfoCommand(),
			BuildsExpireCommand(),
			BuildsUploadCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// BuildsListCommand returns the builds list subcommand
func BuildsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	sort := fs.String("sort", "", "Sort by uploadedDate or -uploadedDate")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc builds list [flags]",
		ShortHelp:  "List builds for an app in App Store Connect.",
		LongHelp: `List builds for an app in App Store Connect.

This command fetches builds uploaded to App Store Connect,
including processing status and expiration dates.

Examples:
  asc builds list --app "123456789"
  asc builds list --app "123456789" --limit 10`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			BuildsInfoCommand(),
			BuildsExpireCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("builds: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("builds: %w", err)
			}
			if err := validateSort(*sort, "uploadedDate", "-uploadedDate"); err != nil {
				return fmt.Errorf("builds: %w", err)
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("builds: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.BuildsOption{
				asc.WithBuildsLimit(*limit),
				asc.WithBuildsNextURL(*next),
			}
			if strings.TrimSpace(*sort) != "" {
				opts = append(opts, asc.WithBuildsSort(*sort))
			}

			builds, err := client.GetBuilds(requestCtx, resolvedAppID, opts...)
			if err != nil {
				return fmt.Errorf("builds: failed to fetch: %w", err)
			}

			format := *output

			return printOutput(builds, format, *pretty)
		},
	}
}

// BuildsInfoCommand returns a build detail subcommand.
func BuildsInfoCommand() *ffcli.Command {
	fs := flag.NewFlagSet("builds info", flag.ExitOnError)

	buildID := fs.String("build", "", "Build ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "info",
		ShortUsage: "asc builds info [flags]",
		ShortHelp:  "Show build details.",
		LongHelp: `Show build details.

Examples:
  asc builds info --build "BUILD_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*buildID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --build is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("builds info: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			build, err := client.GetBuild(requestCtx, strings.TrimSpace(*buildID))
			if err != nil {
				return fmt.Errorf("builds info: failed to fetch: %w", err)
			}

			format := *output

			return printOutput(build, format, *pretty)
		},
	}
}

// BuildsExpireCommand returns a build expiration subcommand.
func BuildsExpireCommand() *ffcli.Command {
	fs := flag.NewFlagSet("builds expire", flag.ExitOnError)

	buildID := fs.String("build", "", "Build ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "expire",
		ShortUsage: "asc builds expire [flags]",
		ShortHelp:  "Expire a build for TestFlight.",
		LongHelp: `Expire a build for TestFlight.

This action is irreversible for the specified build.

Examples:
  asc builds expire --build "BUILD_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*buildID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --build is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("builds expire: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			build, err := client.ExpireBuild(requestCtx, strings.TrimSpace(*buildID))
			if err != nil {
				return fmt.Errorf("builds expire: failed to expire: %w", err)
			}

			format := *output

			return printOutput(build, format, *pretty)
		},
	}
}

// SubmitCommand returns a command to submit a build for review
func SubmitCommand() *ffcli.Command {
	fs := flag.NewFlagSet("submit", flag.ExitOnError)

	versionID := fs.String("version", "", "App Store Version ID (required)")
	confirm := fs.Bool("confirm", false, "Confirm submission (required)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "submit",
		ShortUsage: "asc submit [flags]",
		ShortHelp:  "Submit a build for App Store review.",
		LongHelp: `Submit a build for App Store review.

This command creates an appStoreVersionSubmission to submit
a version for review on the App Store.

Examples:
  asc submit --version "VERSION_ID" --confirm
  asc submit --version "VERSION_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			// Validate required flags
			if *versionID == "" {
				fmt.Fprintf(os.Stderr, "Error: --version is required\n\n")
				return flag.ErrHelp
			}
			if !*confirm {
				fmt.Fprintf(os.Stderr, "Error: --confirm is required to submit for review\n\n")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("submit: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			// Create submission
			submitReq := asc.AppStoreVersionSubmissionCreateRequest{
				Data: asc.AppStoreVersionSubmissionCreateData{
					Type: asc.ResourceTypeAppStoreVersionSubmissions,
					Relationships: &asc.AppStoreVersionSubmissionRelationships{
						AppStoreVersion: &asc.Relationship{
							Data: asc.ResourceData{Type: asc.ResourceTypeAppStoreVersions, ID: *versionID},
						},
					},
				},
			}

			submitResp, err := client.CreateAppStoreVersionSubmission(requestCtx, submitReq)
			if err != nil {
				return fmt.Errorf("submit: failed to create submission: %w", err)
			}

			result := &asc.AppStoreVersionSubmissionResult{
				SubmissionID: submitResp.Data.ID,
				CreatedDate:  submitResp.Data.Attributes.CreatedDate,
			}

			format := *output

			return printOutput(result, format, *pretty)
		},
	}
}

// VersionCommand returns a version subcommand
func VersionCommand(version string) *ffcli.Command {
	return &ffcli.Command{
		Name:       "version",
		ShortUsage: "asc version",
		ShortHelp:  "Print version information and exit.",
		UsageFunc:  DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			fmt.Println(version)
			return nil
		},
	}
}

// RootCommand returns the root command
func RootCommand(version string) *ffcli.Command {
	root := &ffcli.Command{
		Name:       "asc",
		ShortUsage: "asc <subcommand> [flags]",
		ShortHelp:  "A fast, AI-agent friendly CLI for App Store Connect.",
		LongHelp:   "ASC is a lightweight CLI for App Store Connect. Built for developers and AI agents.",
		FlagSet:    flag.NewFlagSet("asc", flag.ExitOnError),
		UsageFunc:  DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AuthCommand(),
			FeedbackCommand(),
			CrashesCommand(),
			ReviewsCommand(),
			AppsCommand(),
			BuildsCommand(),
			BetaGroupsCommand(),
			BetaTestersCommand(),
			SubmitCommand(),
			VersionCommand(version),
		},
	}

	versionFlag := root.FlagSet.Bool("version", false, "Print version and exit")

	root.Exec = func(ctx context.Context, args []string) error {
		if *versionFlag {
			fmt.Fprintln(os.Stdout, version)
			return nil
		}
		if len(args) > 0 {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		}
		return flag.ErrHelp
	}

	return root
}

func getASCClient() (*asc.Client, error) {
	var actualKeyID, actualIssuerID, actualKeyPath string

	// Priority 1: Keychain credentials (explicit user setup via 'asc auth login')
	cfg, err := auth.GetDefaultCredentials()
	if err == nil && cfg != nil {
		actualKeyID = cfg.KeyID
		actualIssuerID = cfg.IssuerID
		actualKeyPath = cfg.PrivateKeyPath
	}

	// Priority 2: Environment variables (fallback for CI/CD or when keychain unavailable)
	if actualKeyID == "" {
		actualKeyID = os.Getenv("ASC_KEY_ID")
	}
	if actualIssuerID == "" {
		actualIssuerID = os.Getenv("ASC_ISSUER_ID")
	}
	if actualKeyPath == "" {
		actualKeyPath = os.Getenv("ASC_PRIVATE_KEY_PATH")
	}

	if actualKeyID == "" || actualIssuerID == "" || actualKeyPath == "" {
		return nil, fmt.Errorf("missing authentication. Run 'asc auth login'")
	}

	return asc.NewClient(actualKeyID, actualIssuerID, actualKeyPath)
}

func printOutput(data interface{}, format string, pretty bool) error {
	format = strings.ToLower(format)
	switch format {
	case "json":
		if pretty {
			return asc.PrintPrettyJSON(data)
		}
		return asc.PrintJSON(data)
	case "markdown", "md":
		if pretty {
			return fmt.Errorf("--pretty is only valid with JSON output")
		}
		return asc.PrintMarkdown(data)
	case "table":
		if pretty {
			return fmt.Errorf("--pretty is only valid with JSON output")
		}
		return asc.PrintTable(data)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func resolveAppID(appID string) string {
	if appID != "" {
		return appID
	}
	return os.Getenv("ASC_APP_ID")
}

func contextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, asc.DefaultTimeout)
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		cleaned = append(cleaned, part)
	}
	return cleaned
}

func splitCSVUpper(value string) []string {
	values := splitCSV(value)
	if len(values) == 0 {
		return nil
	}
	upper := make([]string, 0, len(values))
	for _, item := range values {
		upper = append(upper, strings.ToUpper(item))
	}
	return upper
}

func validateNextURL(next string) error {
	next = strings.TrimSpace(next)
	if next == "" {
		return nil
	}
	parsed, err := url.Parse(next)
	if err != nil {
		return fmt.Errorf("--next must be a valid URL: %w", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "api.appstoreconnect.apple.com" {
		return fmt.Errorf("--next must be an App Store Connect URL")
	}
	return nil
}

func validateSort(value string, allowed ...string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, option := range allowed {
		if value == option {
			return nil
		}
	}
	return fmt.Errorf("--sort must be one of: %s", strings.Join(allowed, ", "))
}
