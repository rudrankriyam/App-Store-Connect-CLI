package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

const (
	localizationTypeVersion = "version"
	localizationTypeAppInfo = "app-info"
)

var (
	versionLocalizationKeys = []string{
		"description",
		"keywords",
		"marketingUrl",
		"promotionalText",
		"supportUrl",
		"whatsNew",
	}
	appInfoLocalizationKeys = []string{
		"name",
		"subtitle",
		"privacyPolicyUrl",
		"privacyChoicesUrl",
		"privacyPolicyText",
	}
)

// LocalizationsCommand returns the localizations command with subcommands.
func LocalizationsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("localizations", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "localizations",
		ShortUsage: "asc localizations <subcommand> [flags]",
		ShortHelp:  "Manage App Store localization metadata.",
		LongHelp: `Manage App Store localization metadata.

Examples:
  asc localizations list --version "VERSION_ID"
  asc localizations download --version "VERSION_ID" --path "./localizations"
  asc localizations upload --version "VERSION_ID" --path "./localizations"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			LocalizationsListCommand(),
			LocalizationsDownloadCommand(),
			LocalizationsUploadCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// LocalizationsListCommand returns the list localizations subcommand.
func LocalizationsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	versionID := fs.String("version", "", "App Store version ID")
	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	appInfoID := fs.String("app-info", "", "App Info ID (optional override)")
	locType := fs.String("type", localizationTypeVersion, "Localization type: version (default) or app-info")
	locale := fs.String("locale", "", "Filter by locale(s), comma-separated")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc localizations list [flags]",
		ShortHelp:  "List localization metadata for an app or version.",
		LongHelp: `List localization metadata for an app or version.

Examples:
  asc localizations list --version "VERSION_ID"
  asc localizations list --app "APP_ID" --type app-info
  asc localizations list --version "VERSION_ID" --locale "en-US,ja"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("localizations list: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("localizations list: %w", err)
			}

			normalizedType, err := normalizeLocalizationType(*locType)
			if err != nil {
				return fmt.Errorf("localizations list: %w", err)
			}

			locales := splitCSV(*locale)

			switch normalizedType {
			case localizationTypeVersion:
				if strings.TrimSpace(*versionID) == "" {
					fmt.Fprintln(os.Stderr, "Error: --version is required for version localizations")
					return flag.ErrHelp
				}

				client, err := getASCClient()
				if err != nil {
					return fmt.Errorf("localizations list: %w", err)
				}

				requestCtx, cancel := contextWithTimeout(ctx)
				defer cancel()

				opts := []asc.AppStoreVersionLocalizationsOption{
					asc.WithAppStoreVersionLocalizationsLimit(*limit),
					asc.WithAppStoreVersionLocalizationsNextURL(*next),
				}
				if len(locales) > 0 {
					opts = append(opts, asc.WithAppStoreVersionLocalizationLocales(locales))
				}

				resp, err := client.GetAppStoreVersionLocalizations(requestCtx, strings.TrimSpace(*versionID), opts...)
				if err != nil {
					return fmt.Errorf("localizations list: failed to fetch: %w", err)
				}
				return printOutput(resp, *output, *pretty)
			case localizationTypeAppInfo:
				resolvedAppID := resolveAppID(*appID)
				if resolvedAppID == "" {
					fmt.Fprintln(os.Stderr, "Error: --app is required for app-info localizations")
					return flag.ErrHelp
				}

				client, err := getASCClient()
				if err != nil {
					return fmt.Errorf("localizations list: %w", err)
				}

				requestCtx, cancel := contextWithTimeout(ctx)
				defer cancel()

				appInfo, err := resolveAppInfoID(requestCtx, client, resolvedAppID, strings.TrimSpace(*appInfoID))
				if err != nil {
					return fmt.Errorf("localizations list: %w", err)
				}

				opts := []asc.AppInfoLocalizationsOption{
					asc.WithAppInfoLocalizationsLimit(*limit),
					asc.WithAppInfoLocalizationsNextURL(*next),
				}
				if len(locales) > 0 {
					opts = append(opts, asc.WithAppInfoLocalizationLocales(locales))
				}

				resp, err := client.GetAppInfoLocalizations(requestCtx, appInfo, opts...)
				if err != nil {
					return fmt.Errorf("localizations list: failed to fetch: %w", err)
				}
				return printOutput(resp, *output, *pretty)
			default:
				return fmt.Errorf("localizations list: unsupported type %q", normalizedType)
			}
		},
	}
}

// LocalizationsDownloadCommand returns the download localizations subcommand.
func LocalizationsDownloadCommand() *ffcli.Command {
	fs := flag.NewFlagSet("download", flag.ExitOnError)

	versionID := fs.String("version", "", "App Store version ID")
	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	appInfoID := fs.String("app-info", "", "App Info ID (optional override)")
	locType := fs.String("type", localizationTypeVersion, "Localization type: version (default) or app-info")
	locale := fs.String("locale", "", "Filter by locale(s), comma-separated")
	path := fs.String("path", "localizations", "Output path (directory or .strings file)")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "download",
		ShortUsage: "asc localizations download [flags]",
		ShortHelp:  "Download localizations to .strings files.",
		LongHelp: `Download localizations to .strings files.

Examples:
  asc localizations download --version "VERSION_ID" --path "./localizations"
  asc localizations download --app "APP_ID" --type app-info --path "./localizations"
  asc localizations download --version "VERSION_ID" --locale "en-US" --path "en-US.strings"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("localizations download: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("localizations download: %w", err)
			}

			normalizedType, err := normalizeLocalizationType(*locType)
			if err != nil {
				return fmt.Errorf("localizations download: %w", err)
			}

			locales := splitCSV(*locale)

			switch normalizedType {
			case localizationTypeVersion:
				if strings.TrimSpace(*versionID) == "" {
					fmt.Fprintln(os.Stderr, "Error: --version is required for version localizations")
					return flag.ErrHelp
				}

				client, err := getASCClient()
				if err != nil {
					return fmt.Errorf("localizations download: %w", err)
				}

				requestCtx, cancel := contextWithTimeout(ctx)
				defer cancel()

				opts := []asc.AppStoreVersionLocalizationsOption{
					asc.WithAppStoreVersionLocalizationsLimit(*limit),
					asc.WithAppStoreVersionLocalizationsNextURL(*next),
				}
				if len(locales) > 0 {
					opts = append(opts, asc.WithAppStoreVersionLocalizationLocales(locales))
				}

				resp, err := client.GetAppStoreVersionLocalizations(requestCtx, strings.TrimSpace(*versionID), opts...)
				if err != nil {
					return fmt.Errorf("localizations download: failed to fetch: %w", err)
				}

				files, err := writeVersionLocalizationStrings(*path, resp.Data)
				if err != nil {
					return fmt.Errorf("localizations download: %w", err)
				}

				result := asc.LocalizationDownloadResult{
					Type:       normalizedType,
					VersionID:  strings.TrimSpace(*versionID),
					OutputPath: *path,
					Files:      files,
				}

				return printOutput(&result, *output, *pretty)
			case localizationTypeAppInfo:
				resolvedAppID := resolveAppID(*appID)
				if resolvedAppID == "" {
					fmt.Fprintln(os.Stderr, "Error: --app is required for app-info localizations")
					return flag.ErrHelp
				}

				client, err := getASCClient()
				if err != nil {
					return fmt.Errorf("localizations download: %w", err)
				}

				requestCtx, cancel := contextWithTimeout(ctx)
				defer cancel()

				appInfo, err := resolveAppInfoID(requestCtx, client, resolvedAppID, strings.TrimSpace(*appInfoID))
				if err != nil {
					return fmt.Errorf("localizations download: %w", err)
				}

				opts := []asc.AppInfoLocalizationsOption{
					asc.WithAppInfoLocalizationsLimit(*limit),
					asc.WithAppInfoLocalizationsNextURL(*next),
				}
				if len(locales) > 0 {
					opts = append(opts, asc.WithAppInfoLocalizationLocales(locales))
				}

				resp, err := client.GetAppInfoLocalizations(requestCtx, appInfo, opts...)
				if err != nil {
					return fmt.Errorf("localizations download: failed to fetch: %w", err)
				}

				files, err := writeAppInfoLocalizationStrings(*path, resp.Data)
				if err != nil {
					return fmt.Errorf("localizations download: %w", err)
				}

				result := asc.LocalizationDownloadResult{
					Type:       normalizedType,
					AppID:      resolvedAppID,
					AppInfoID:  appInfo,
					OutputPath: *path,
					Files:      files,
				}

				return printOutput(&result, *output, *pretty)
			default:
				return fmt.Errorf("localizations download: unsupported type %q", normalizedType)
			}
		},
	}
}

// LocalizationsUploadCommand returns the upload localizations subcommand.
func LocalizationsUploadCommand() *ffcli.Command {
	fs := flag.NewFlagSet("upload", flag.ExitOnError)

	versionID := fs.String("version", "", "App Store version ID")
	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	appInfoID := fs.String("app-info", "", "App Info ID (optional override)")
	locType := fs.String("type", localizationTypeVersion, "Localization type: version (default) or app-info")
	locale := fs.String("locale", "", "Filter by locale(s), comma-separated")
	path := fs.String("path", "", "Input path (directory or .strings file)")
	dryRun := fs.Bool("dry-run", false, "Validate file without uploading")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "upload",
		ShortUsage: "asc localizations upload [flags]",
		ShortHelp:  "Upload localizations from .strings files.",
		LongHelp: `Upload localizations from .strings files.

Examples:
  asc localizations upload --version "VERSION_ID" --path "./localizations"
  asc localizations upload --app "APP_ID" --type app-info --path "./localizations"
  asc localizations upload --version "VERSION_ID" --locale "en-US" --path "en-US.strings"
  asc localizations upload --version "VERSION_ID" --path "./localizations" --dry-run`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*path) == "" {
				fmt.Fprintln(os.Stderr, "Error: --path is required")
				return flag.ErrHelp
			}

			normalizedType, err := normalizeLocalizationType(*locType)
			if err != nil {
				return fmt.Errorf("localizations upload: %w", err)
			}

			locales := splitCSV(*locale)

			switch normalizedType {
			case localizationTypeVersion:
				if strings.TrimSpace(*versionID) == "" {
					fmt.Fprintln(os.Stderr, "Error: --version is required for version localizations")
					return flag.ErrHelp
				}

				valuesByLocale, err := readLocalizationStrings(*path, locales)
				if err != nil {
					return fmt.Errorf("localizations upload: %w", err)
				}

				client, err := getASCClient()
				if err != nil {
					return fmt.Errorf("localizations upload: %w", err)
				}

				requestCtx, cancel := contextWithTimeout(ctx)
				defer cancel()

				results, err := uploadVersionLocalizations(requestCtx, client, strings.TrimSpace(*versionID), valuesByLocale, *dryRun)
				if err != nil {
					return fmt.Errorf("localizations upload: %w", err)
				}

				result := asc.LocalizationUploadResult{
					Type:      normalizedType,
					VersionID: strings.TrimSpace(*versionID),
					DryRun:    *dryRun,
					Results:   results,
				}

				return printOutput(&result, *output, *pretty)
			case localizationTypeAppInfo:
				resolvedAppID := resolveAppID(*appID)
				if resolvedAppID == "" {
					fmt.Fprintln(os.Stderr, "Error: --app is required for app-info localizations")
					return flag.ErrHelp
				}

				valuesByLocale, err := readLocalizationStrings(*path, locales)
				if err != nil {
					return fmt.Errorf("localizations upload: %w", err)
				}

				client, err := getASCClient()
				if err != nil {
					return fmt.Errorf("localizations upload: %w", err)
				}

				requestCtx, cancel := contextWithTimeout(ctx)
				defer cancel()

				appInfo, err := resolveAppInfoID(requestCtx, client, resolvedAppID, strings.TrimSpace(*appInfoID))
				if err != nil {
					return fmt.Errorf("localizations upload: %w", err)
				}

				results, err := uploadAppInfoLocalizations(requestCtx, client, appInfo, valuesByLocale, *dryRun)
				if err != nil {
					return fmt.Errorf("localizations upload: %w", err)
				}

				result := asc.LocalizationUploadResult{
					Type:      normalizedType,
					AppID:     resolvedAppID,
					AppInfoID: appInfo,
					DryRun:    *dryRun,
					Results:   results,
				}

				return printOutput(&result, *output, *pretty)
			default:
				return fmt.Errorf("localizations upload: unsupported type %q", normalizedType)
			}
		},
	}
}

func normalizeLocalizationType(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case localizationTypeVersion, localizationTypeAppInfo:
		return normalized, nil
	default:
		return "", fmt.Errorf("--type must be %q or %q", localizationTypeVersion, localizationTypeAppInfo)
	}
}

func resolveAppInfoID(ctx context.Context, client *asc.Client, appID, appInfoID string) (string, error) {
	if appInfoID != "" {
		return appInfoID, nil
	}

	resp, err := client.GetAppInfos(ctx, appID)
	if err != nil {
		return "", err
	}
	if len(resp.Data) == 0 {
		return "", fmt.Errorf("no app info found for app %q", appID)
	}
	if len(resp.Data) > 1 {
		return "", fmt.Errorf("multiple app infos found for app %q; use --app-info", appID)
	}
	return resp.Data[0].ID, nil
}

func writeVersionLocalizationStrings(outputPath string, items []asc.Resource[asc.AppStoreVersionLocalizationAttributes]) ([]asc.LocalizationFileResult, error) {
	byLocale := make(map[string]map[string]string, len(items))
	for _, item := range items {
		locale := strings.TrimSpace(item.Attributes.Locale)
		if locale == "" {
			continue
		}
		byLocale[locale] = mapVersionLocalizationStrings(item.Attributes)
	}
	return writeLocalizationStrings(outputPath, byLocale, versionLocalizationKeys)
}

func writeAppInfoLocalizationStrings(outputPath string, items []asc.Resource[asc.AppInfoLocalizationAttributes]) ([]asc.LocalizationFileResult, error) {
	byLocale := make(map[string]map[string]string, len(items))
	for _, item := range items {
		locale := strings.TrimSpace(item.Attributes.Locale)
		if locale == "" {
			continue
		}
		byLocale[locale] = mapAppInfoLocalizationStrings(item.Attributes)
	}
	return writeLocalizationStrings(outputPath, byLocale, appInfoLocalizationKeys)
}

func writeLocalizationStrings(outputPath string, valuesByLocale map[string]map[string]string, order []string) ([]asc.LocalizationFileResult, error) {
	if len(valuesByLocale) == 0 {
		return nil, fmt.Errorf("no localizations returned")
	}

	locales := make([]string, 0, len(valuesByLocale))
	for locale := range valuesByLocale {
		locales = append(locales, locale)
	}
	sort.Strings(locales)

	paths, err := resolveLocalizationOutputPaths(outputPath, locales)
	if err != nil {
		return nil, err
	}

	results := make([]asc.LocalizationFileResult, 0, len(locales))
	for _, locale := range locales {
		path, ok := paths[locale]
		if !ok {
			continue
		}
		if err := writeStringsFile(path, valuesByLocale[locale], order); err != nil {
			return nil, err
		}
		results = append(results, asc.LocalizationFileResult{
			Locale: locale,
			Path:   path,
		})
	}
	return results, nil
}

func resolveLocalizationOutputPaths(outputPath string, locales []string) (map[string]string, error) {
	if strings.TrimSpace(outputPath) == "" {
		outputPath = "localizations"
	}

	result := make(map[string]string, len(locales))
	if strings.HasSuffix(outputPath, ".strings") {
		if len(locales) != 1 {
			return nil, fmt.Errorf("output path %q requires exactly one locale", outputPath)
		}
		path := outputPath
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		result[locales[0]] = path
		return result, nil
	}

	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		return nil, err
	}
	for _, locale := range locales {
		result[locale] = filepath.Join(outputPath, locale+".strings")
	}
	return result, nil
}

func mapVersionLocalizationStrings(attrs asc.AppStoreVersionLocalizationAttributes) map[string]string {
	values := make(map[string]string)
	setIfNotEmpty(values, "description", attrs.Description)
	setIfNotEmpty(values, "keywords", attrs.Keywords)
	setIfNotEmpty(values, "marketingUrl", attrs.MarketingURL)
	setIfNotEmpty(values, "promotionalText", attrs.PromotionalText)
	setIfNotEmpty(values, "supportUrl", attrs.SupportURL)
	setIfNotEmpty(values, "whatsNew", attrs.WhatsNew)
	return values
}

func mapAppInfoLocalizationStrings(attrs asc.AppInfoLocalizationAttributes) map[string]string {
	values := make(map[string]string)
	setIfNotEmpty(values, "name", attrs.Name)
	setIfNotEmpty(values, "subtitle", attrs.Subtitle)
	setIfNotEmpty(values, "privacyPolicyUrl", attrs.PrivacyPolicyURL)
	setIfNotEmpty(values, "privacyChoicesUrl", attrs.PrivacyChoicesURL)
	setIfNotEmpty(values, "privacyPolicyText", attrs.PrivacyPolicyText)
	return values
}

func setIfNotEmpty(values map[string]string, key, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	values[key] = value
}

func readLocalizationStrings(inputPath string, locales []string) (map[string]map[string]string, error) {
	info, err := os.Stat(inputPath)
	if err != nil {
		return nil, err
	}

	filter := make(map[string]bool)
	for _, locale := range locales {
		filter[locale] = true
	}

	if !info.IsDir() {
		if len(locales) > 1 {
			return nil, fmt.Errorf("single file input only supports one locale")
		}
		locale := ""
		if len(locales) == 1 {
			locale = locales[0]
		} else {
			locale = strings.TrimSuffix(filepath.Base(inputPath), ".strings")
			if locale == "" || locale == filepath.Base(inputPath) {
				return nil, fmt.Errorf("cannot infer locale from %q (use --locale)", inputPath)
			}
		}

		entries, err := readStringsFile(inputPath)
		if err != nil {
			return nil, err
		}
		return map[string]map[string]string{locale: entries}, nil
	}

	entries, err := os.ReadDir(inputPath)
	if err != nil {
		return nil, err
	}

	values := make(map[string]map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".strings" {
			continue
		}
		locale := strings.TrimSuffix(entry.Name(), ".strings")
		if locale == "" {
			continue
		}
		if len(filter) > 0 && !filter[locale] {
			continue
		}
		path := filepath.Join(inputPath, entry.Name())
		parsed, err := readStringsFile(path)
		if err != nil {
			return nil, err
		}
		if _, exists := values[locale]; exists {
			return nil, fmt.Errorf("duplicate locale %q in %s", locale, inputPath)
		}
		values[locale] = parsed
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("no .strings files found in %q", inputPath)
	}
	return values, nil
}

func uploadVersionLocalizations(ctx context.Context, client *asc.Client, versionID string, valuesByLocale map[string]map[string]string, dryRun bool) ([]asc.LocalizationUploadLocaleResult, error) {
	validateKeys := buildAllowedKeys(versionLocalizationKeys)
	for locale, values := range valuesByLocale {
		if err := validateLocalizationKeys(locale, values, validateKeys); err != nil {
			return nil, err
		}
	}

	existing, err := client.GetAppStoreVersionLocalizations(ctx, versionID, asc.WithAppStoreVersionLocalizationsLimit(200))
	if err != nil {
		return nil, err
	}
	existingByLocale := make(map[string]string, len(existing.Data))
	for _, item := range existing.Data {
		if strings.TrimSpace(item.Attributes.Locale) == "" {
			continue
		}
		existingByLocale[item.Attributes.Locale] = item.ID
	}

	return uploadLocalizationValues(ctx, valuesByLocale, existingByLocale, func(locale string, values map[string]string, existingID string) (asc.LocalizationUploadLocaleResult, error) {
		attributes := buildVersionLocalizationAttributes(locale, values, existingID == "")
		if existingID == "" {
			if dryRun {
				return asc.LocalizationUploadLocaleResult{Locale: locale, Action: "create"}, nil
			}
			resp, err := client.CreateAppStoreVersionLocalization(ctx, versionID, attributes)
			if err != nil {
				return asc.LocalizationUploadLocaleResult{}, err
			}
			return asc.LocalizationUploadLocaleResult{Locale: locale, Action: "create", LocalizationID: resp.Data.ID}, nil
		}
		if dryRun {
			return asc.LocalizationUploadLocaleResult{Locale: locale, Action: "update", LocalizationID: existingID}, nil
		}
		resp, err := client.UpdateAppStoreVersionLocalization(ctx, existingID, attributes)
		if err != nil {
			return asc.LocalizationUploadLocaleResult{}, err
		}
		return asc.LocalizationUploadLocaleResult{Locale: locale, Action: "update", LocalizationID: resp.Data.ID}, nil
	})
}

func uploadAppInfoLocalizations(ctx context.Context, client *asc.Client, appInfoID string, valuesByLocale map[string]map[string]string, dryRun bool) ([]asc.LocalizationUploadLocaleResult, error) {
	validateKeys := buildAllowedKeys(appInfoLocalizationKeys)
	for locale, values := range valuesByLocale {
		if err := validateLocalizationKeys(locale, values, validateKeys); err != nil {
			return nil, err
		}
	}

	existing, err := client.GetAppInfoLocalizations(ctx, appInfoID, asc.WithAppInfoLocalizationsLimit(200))
	if err != nil {
		return nil, err
	}
	existingByLocale := make(map[string]string, len(existing.Data))
	for _, item := range existing.Data {
		if strings.TrimSpace(item.Attributes.Locale) == "" {
			continue
		}
		existingByLocale[item.Attributes.Locale] = item.ID
	}

	return uploadLocalizationValues(ctx, valuesByLocale, existingByLocale, func(locale string, values map[string]string, existingID string) (asc.LocalizationUploadLocaleResult, error) {
		attributes := buildAppInfoLocalizationAttributes(locale, values, existingID == "")
		if existingID == "" {
			if dryRun {
				return asc.LocalizationUploadLocaleResult{Locale: locale, Action: "create"}, nil
			}
			resp, err := client.CreateAppInfoLocalization(ctx, appInfoID, attributes)
			if err != nil {
				return asc.LocalizationUploadLocaleResult{}, err
			}
			return asc.LocalizationUploadLocaleResult{Locale: locale, Action: "create", LocalizationID: resp.Data.ID}, nil
		}
		if dryRun {
			return asc.LocalizationUploadLocaleResult{Locale: locale, Action: "update", LocalizationID: existingID}, nil
		}
		resp, err := client.UpdateAppInfoLocalization(ctx, existingID, attributes)
		if err != nil {
			return asc.LocalizationUploadLocaleResult{}, err
		}
		return asc.LocalizationUploadLocaleResult{Locale: locale, Action: "update", LocalizationID: resp.Data.ID}, nil
	})
}

func uploadLocalizationValues(ctx context.Context, valuesByLocale map[string]map[string]string, existing map[string]string, handler func(locale string, values map[string]string, existingID string) (asc.LocalizationUploadLocaleResult, error)) ([]asc.LocalizationUploadLocaleResult, error) {
	locales := make([]string, 0, len(valuesByLocale))
	for locale := range valuesByLocale {
		locales = append(locales, locale)
	}
	sort.Strings(locales)

	results := make([]asc.LocalizationUploadLocaleResult, 0, len(locales))
	for _, locale := range locales {
		values := valuesByLocale[locale]
		if len(values) == 0 {
			return nil, fmt.Errorf("no localization values for locale %q", locale)
		}
		if !hasNonEmptyLocalizationValues(values) {
			return nil, fmt.Errorf("localization values for locale %q are empty", locale)
		}
		result, err := handler(locale, values, existing[locale])
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func hasNonEmptyLocalizationValues(values map[string]string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

func buildAllowedKeys(keys []string) map[string]bool {
	allowed := make(map[string]bool, len(keys))
	for _, key := range keys {
		allowed[key] = true
	}
	return allowed
}

func validateLocalizationKeys(locale string, values map[string]string, allowed map[string]bool) error {
	unknown := make([]string, 0)
	for key := range values {
		if !allowed[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return fmt.Errorf("unsupported keys for locale %q: %s", locale, strings.Join(unknown, ", "))
	}
	return nil
}

func buildVersionLocalizationAttributes(locale string, values map[string]string, includeLocale bool) asc.AppStoreVersionLocalizationAttributes {
	attrs := asc.AppStoreVersionLocalizationAttributes{}
	if includeLocale {
		attrs.Locale = locale
	}
	if value, ok := values["description"]; ok {
		attrs.Description = value
	}
	if value, ok := values["keywords"]; ok {
		attrs.Keywords = value
	}
	if value, ok := values["marketingUrl"]; ok {
		attrs.MarketingURL = value
	}
	if value, ok := values["promotionalText"]; ok {
		attrs.PromotionalText = value
	}
	if value, ok := values["supportUrl"]; ok {
		attrs.SupportURL = value
	}
	if value, ok := values["whatsNew"]; ok {
		attrs.WhatsNew = value
	}
	return attrs
}

func buildAppInfoLocalizationAttributes(locale string, values map[string]string, includeLocale bool) asc.AppInfoLocalizationAttributes {
	attrs := asc.AppInfoLocalizationAttributes{}
	if includeLocale {
		attrs.Locale = locale
	}
	if value, ok := values["name"]; ok {
		attrs.Name = value
	}
	if value, ok := values["subtitle"]; ok {
		attrs.Subtitle = value
	}
	if value, ok := values["privacyPolicyUrl"]; ok {
		attrs.PrivacyPolicyURL = value
	}
	if value, ok := values["privacyChoicesUrl"]; ok {
		attrs.PrivacyChoicesURL = value
	}
	if value, ok := values["privacyPolicyText"]; ok {
		attrs.PrivacyPolicyText = value
	}
	return attrs
}

type stringsParser struct {
	runes []rune
	pos   int
	line  int
}

func readStringsFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseStringsContent(string(data))
}

func parseStringsContent(content string) (map[string]string, error) {
	parser := &stringsParser{runes: []rune(content), line: 1}
	values := make(map[string]string)
	for {
		if err := parser.skipWhitespaceAndComments(); err != nil {
			return nil, err
		}
		if parser.eof() {
			break
		}
		key, err := parser.readQuoted()
		if err != nil {
			return nil, err
		}
		if err := parser.skipWhitespaceAndComments(); err != nil {
			return nil, err
		}
		if !parser.consume('=') {
			return nil, parser.errorf("expected '=' after key")
		}
		if err := parser.skipWhitespaceAndComments(); err != nil {
			return nil, err
		}
		value, err := parser.readQuoted()
		if err != nil {
			return nil, err
		}
		if err := parser.skipWhitespaceAndComments(); err != nil {
			return nil, err
		}
		if !parser.consume(';') {
			return nil, parser.errorf("expected ';' after value")
		}
		values[key] = value
	}
	return values, nil
}

func (p *stringsParser) eof() bool {
	return p.pos >= len(p.runes)
}

func (p *stringsParser) peek() rune {
	if p.eof() {
		return 0
	}
	return p.runes[p.pos]
}

func (p *stringsParser) peekNext() rune {
	if p.pos+1 >= len(p.runes) {
		return 0
	}
	return p.runes[p.pos+1]
}

func (p *stringsParser) next() rune {
	if p.eof() {
		return 0
	}
	ch := p.runes[p.pos]
	p.pos++
	if ch == '\n' {
		p.line++
	}
	return ch
}

func (p *stringsParser) consume(expected rune) bool {
	if p.peek() != expected {
		return false
	}
	p.next()
	return true
}

func (p *stringsParser) skipWhitespaceAndComments() error {
	for {
		for unicode.IsSpace(p.peek()) {
			p.next()
		}
		if p.peek() == '/' && p.peekNext() == '/' {
			for !p.eof() && p.next() != '\n' {
			}
			continue
		}
		if p.peek() == '/' && p.peekNext() == '*' {
			p.next()
			p.next()
			foundEnd := false
			for !p.eof() {
				if p.peek() == '*' && p.peekNext() == '/' {
					p.next()
					p.next()
					foundEnd = true
					break
				}
				p.next()
			}
			if !foundEnd {
				return p.errorf("unterminated block comment")
			}
			continue
		}
		break
	}
	return nil
}

func (p *stringsParser) readQuoted() (string, error) {
	if !p.consume('"') {
		return "", p.errorf("expected '\"'")
	}
	var b strings.Builder
	for !p.eof() {
		ch := p.next()
		if ch == '"' {
			return b.String(), nil
		}
		if ch == '\\' {
			if p.eof() {
				return "", p.errorf("unterminated escape sequence")
			}
			escaped := p.next()
			switch escaped {
			case '"', '\\':
				b.WriteRune(escaped)
			case 'n':
				b.WriteRune('\n')
			case 'r':
				b.WriteRune('\r')
			case 't':
				b.WriteRune('\t')
			case 'u':
				r, err := p.readHexRune(4)
				if err != nil {
					return "", err
				}
				b.WriteRune(r)
			case 'U':
				r, err := p.readHexRune(8)
				if err != nil {
					return "", err
				}
				b.WriteRune(r)
			default:
				b.WriteRune(escaped)
			}
			continue
		}
		b.WriteRune(ch)
	}
	return "", p.errorf("unterminated string")
}

func (p *stringsParser) readHexRune(length int) (rune, error) {
	if p.pos+length > len(p.runes) {
		return 0, p.errorf("invalid unicode escape")
	}
	hex := string(p.runes[p.pos : p.pos+length])
	p.pos += length
	value, err := strconv.ParseInt(hex, 16, 32)
	if err != nil {
		return 0, p.errorf("invalid unicode escape")
	}
	return rune(value), nil
}

func (p *stringsParser) errorf(message string) error {
	return fmt.Errorf("strings parse error on line %d: %s", p.line, message)
}

func writeStringsFile(path string, values map[string]string, order []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var b strings.Builder
	for _, key := range order {
		value, ok := values[key]
		if !ok {
			continue
		}
		fmt.Fprintf(&b, "\"%s\" = \"%s\";\n", key, escapeStringsValue(value))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func escapeStringsValue(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
	)
	return replacer.Replace(value)
}
