package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

// AppSetupCommand returns the app-setup command group.
func AppSetupCommand() *ffcli.Command {
	fs := flag.NewFlagSet("app-setup", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "app-setup",
		ShortUsage: "asc app-setup <subcommand> [flags]",
		ShortHelp:  "Post-create app setup automation.",
		LongHelp: `Post-create app setup automation using public App Store Connect APIs.

Examples:
  asc app-setup info set --app "APP_ID" --primary-locale "en-US" --bundle-id "com.example.app"
  asc app-setup categories set --app "APP_ID" --primary GAMES
  asc app-setup availability set --app "APP_ID" --territory "USA,GBR" --available true
  asc app-setup pricing set --app "APP_ID" --price-point "PRICE_POINT_ID" --base-territory "USA"
  asc app-setup localizations upload --version "VERSION_ID" --path "./localizations"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AppSetupInfoCommand(),
			AppSetupCategoriesCommand(),
			AppSetupAvailabilityCommand(),
			AppSetupPricingCommand(),
			AppSetupLocalizationsCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// AppSetupInfoCommand returns the info subcommand group.
func AppSetupInfoCommand() *ffcli.Command {
	return &ffcli.Command{
		Name:       "info",
		ShortUsage: "asc app-setup info <subcommand> [flags]",
		ShortHelp:  "Update app info and app info localizations.",
		LongHelp: `Update app attributes and app info localizations.

Examples:
  asc app-setup info set --app "APP_ID" --primary-locale "en-US" --bundle-id "com.example.app"
  asc app-setup info set --app "APP_ID" --locale "en-US" --name "My App" --subtitle "Great app"
  asc app-setup info set --app "APP_ID" --primary-locale "en-US" --privacy-policy-url "https://example.com/privacy"`,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AppSetupInfoSetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// AppSetupInfoSetCommand returns the info set subcommand.
func AppSetupInfoSetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("app-setup info set", flag.ExitOnError)

	appID := fs.String("app", os.Getenv("ASC_APP_ID"), "App Store Connect app ID (required)")
	bundleID := fs.String("bundle-id", "", "Bundle ID to set")
	primaryLocale := fs.String("primary-locale", "", "Primary locale (e.g., en-US)")
	locale := fs.String("locale", "", "Locale for app info localization (defaults to --primary-locale)")
	appInfoID := fs.String("app-info", "", "App Info ID (optional override)")
	name := fs.String("name", "", "Localized app name")
	subtitle := fs.String("subtitle", "", "Localized app subtitle")
	privacyPolicyURL := fs.String("privacy-policy-url", "", "Localized privacy policy URL")
	privacyChoicesURL := fs.String("privacy-choices-url", "", "Localized privacy choices URL")
	privacyPolicyText := fs.String("privacy-policy-text", "", "Localized privacy policy text")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "set",
		ShortUsage: "asc app-setup info set [flags]",
		ShortHelp:  "Set app attributes and app info localizations.",
		LongHelp: `Set app attributes (bundle ID, primary locale) and app info localizations.

Examples:
  asc app-setup info set --app "APP_ID" --primary-locale "en-US" --bundle-id "com.example.app"
  asc app-setup info set --app "APP_ID" --locale "en-US" --name "My App" --subtitle "Great app"
  asc app-setup info set --app "APP_ID" --primary-locale "en-US" --privacy-policy-url "https://example.com/privacy"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			appIDValue := strings.TrimSpace(*appID)
			if appIDValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required")
				return flag.ErrHelp
			}

			bundleIDValue := strings.TrimSpace(*bundleID)
			primaryLocaleValue := strings.TrimSpace(*primaryLocale)

			localeValue := strings.TrimSpace(*locale)
			if localeValue == "" && primaryLocaleValue != "" {
				localeValue = primaryLocaleValue
			}

			nameValue := strings.TrimSpace(*name)
			subtitleValue := strings.TrimSpace(*subtitle)
			privacyPolicyURLValue := strings.TrimSpace(*privacyPolicyURL)
			privacyChoicesURLValue := strings.TrimSpace(*privacyChoicesURL)
			privacyPolicyTextValue := strings.TrimSpace(*privacyPolicyText)

			hasAppUpdate := bundleIDValue != "" || primaryLocaleValue != ""
			hasLocalization := nameValue != "" ||
				subtitleValue != "" ||
				privacyPolicyURLValue != "" ||
				privacyChoicesURLValue != "" ||
				privacyPolicyTextValue != ""

			if !hasAppUpdate && !hasLocalization {
				fmt.Fprintln(os.Stderr, "Error: provide at least one update flag")
				return flag.ErrHelp
			}
			if primaryLocaleValue != "" {
				if err := validateBuildLocalizationLocale(primaryLocaleValue); err != nil {
					return fmt.Errorf("app-setup info set: %w", err)
				}
			}
			if hasLocalization && localeValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --locale is required for app info localization updates")
				return flag.ErrHelp
			}
			if localeValue != "" {
				if err := validateBuildLocalizationLocale(localeValue); err != nil {
					return fmt.Errorf("app-setup info set: %w", err)
				}
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("app-setup info set: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			var appResp *asc.AppResponse
			if hasAppUpdate {
				attrs := asc.AppUpdateAttributes{}
				if bundleIDValue != "" {
					attrs.BundleID = &bundleIDValue
				}
				if primaryLocaleValue != "" {
					attrs.PrimaryLocale = &primaryLocaleValue
				}
				appResp, err = client.UpdateApp(requestCtx, appIDValue, attrs)
				if err != nil {
					return fmt.Errorf("app-setup info set: %w", err)
				}
			}

			var appInfoResp *asc.AppInfoLocalizationResponse
			if hasLocalization {
				resolvedAppInfoID, err := resolveAppInfoID(requestCtx, client, appIDValue, strings.TrimSpace(*appInfoID))
				if err != nil {
					return fmt.Errorf("app-setup info set: %w", err)
				}

				localizations, err := client.GetAppInfoLocalizations(
					requestCtx,
					resolvedAppInfoID,
					asc.WithAppInfoLocalizationsLimit(200),
					asc.WithAppInfoLocalizationLocales([]string{localeValue}),
				)
				if err != nil {
					return fmt.Errorf("app-setup info set: failed to fetch app info localizations: %w", err)
				}

				attrs := asc.AppInfoLocalizationAttributes{}
				if nameValue != "" {
					attrs.Name = nameValue
				}
				if subtitleValue != "" {
					attrs.Subtitle = subtitleValue
				}
				if privacyPolicyURLValue != "" {
					attrs.PrivacyPolicyURL = privacyPolicyURLValue
				}
				if privacyChoicesURLValue != "" {
					attrs.PrivacyChoicesURL = privacyChoicesURLValue
				}
				if privacyPolicyTextValue != "" {
					attrs.PrivacyPolicyText = privacyPolicyTextValue
				}

				if len(localizations.Data) == 0 {
					attrs.Locale = localeValue
					appInfoResp, err = client.CreateAppInfoLocalization(requestCtx, resolvedAppInfoID, attrs)
					if err != nil {
						return fmt.Errorf("app-setup info set: %w", err)
					}
				} else {
					localizationID := strings.TrimSpace(localizations.Data[0].ID)
					if localizationID == "" {
						return fmt.Errorf("app-setup info set: localization id is empty")
					}
					appInfoResp, err = client.UpdateAppInfoLocalization(requestCtx, localizationID, attrs)
					if err != nil {
						return fmt.Errorf("app-setup info set: %w", err)
					}
				}
			}

			result := &asc.AppSetupInfoResult{
				AppID:               appIDValue,
				App:                 appResp,
				AppInfoLocalization: appInfoResp,
			}

			return printOutput(result, *output, *pretty)
		},
	}
}

// AppSetupCategoriesCommand returns the categories subcommand group.
func AppSetupCategoriesCommand() *ffcli.Command {
	return &ffcli.Command{
		Name:       "categories",
		ShortUsage: "asc app-setup categories <subcommand> [flags]",
		ShortHelp:  "Set categories for an app.",
		LongHelp: `Set primary and secondary categories for an app.

Examples:
  asc app-setup categories set --app "APP_ID" --primary GAMES --secondary ENTERTAINMENT`,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AppSetupCategoriesSetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// AppSetupCategoriesSetCommand returns the categories set subcommand.
func AppSetupCategoriesSetCommand() *ffcli.Command {
	return newCategoriesSetCommand(categoriesSetCommandConfig{
		flagSetName: "app-setup categories set",
		shortUsage:  "asc app-setup categories set --app APP_ID --primary CATEGORY_ID [--secondary CATEGORY_ID] [--app-info APP_INFO_ID]",
		shortHelp:   "Set primary and secondary categories for an app.",
		longHelp: `Set the primary and secondary categories for an app.

Use 'asc categories list' to find valid category IDs.

Examples:
  asc app-setup categories set --app 123456789 --primary GAMES
  asc app-setup categories set --app 123456789 --primary GAMES --secondary ENTERTAINMENT`,
		errorPrefix:    "app-setup categories set",
		includeAppInfo: true,
	})
}

// AppSetupAvailabilityCommand returns the availability subcommand group.
func AppSetupAvailabilityCommand() *ffcli.Command {
	return &ffcli.Command{
		Name:       "availability",
		ShortUsage: "asc app-setup availability <subcommand> [flags]",
		ShortHelp:  "Set app availability.",
		LongHelp: `Set app availability for territories.

Examples:
  asc app-setup availability set --app "APP_ID" --territory "USA,GBR" --available true`,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AppSetupAvailabilitySetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// AppSetupAvailabilitySetCommand returns the availability set subcommand.
func AppSetupAvailabilitySetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("app-setup availability set", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID)")
	territory := fs.String("territory", "", "Territory IDs (comma-separated, e.g., USA,GBR)")
	var available optionalBool
	fs.Var(&available, "available", "Set availability: true or false")
	var availableInNewTerritories optionalBool
	fs.Var(&availableInNewTerritories, "available-in-new-territories", "Set availability for new territories: true or false")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "set",
		ShortUsage: "asc app-setup availability set [flags]",
		ShortHelp:  "Set app availability for territories.",
		LongHelp: `Set app availability for territories.

Examples:
  asc app-setup availability set --app "123456789" --territory "USA,GBR" --available true --available-in-new-territories true`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required (or set ASC_APP_ID)")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*territory) == "" {
				fmt.Fprintln(os.Stderr, "Error: --territory is required")
				return flag.ErrHelp
			}
			if !available.set {
				fmt.Fprintln(os.Stderr, "Error: --available is required (true or false)")
				return flag.ErrHelp
			}
			if !availableInNewTerritories.set {
				fmt.Fprintln(os.Stderr, "Error: --available-in-new-territories is required (true or false)")
				return flag.ErrHelp
			}

			territories := splitCSVUpper(*territory)
			if len(territories) == 0 {
				fmt.Fprintln(os.Stderr, "Error: --territory must include at least one value")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("app-setup availability set: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			availabilities := make([]asc.TerritoryAvailabilityCreate, 0, len(territories))
			for _, territoryID := range territories {
				availabilities = append(availabilities, asc.TerritoryAvailabilityCreate{
					TerritoryID: territoryID,
					Available:   available.value,
				})
			}

			resp, err := client.CreateAppAvailabilityV2(requestCtx, resolvedAppID, asc.AppAvailabilityV2CreateAttributes{
				AvailableInNewTerritories: &availableInNewTerritories.value,
				TerritoryAvailabilities:   availabilities,
			})
			if err != nil {
				return fmt.Errorf("app-setup availability set: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// AppSetupPricingCommand returns the pricing subcommand group.
func AppSetupPricingCommand() *ffcli.Command {
	return &ffcli.Command{
		Name:       "pricing",
		ShortUsage: "asc app-setup pricing <subcommand> [flags]",
		ShortHelp:  "Set app pricing.",
		LongHelp: `Set app pricing using a price point.

Examples:
  asc app-setup pricing set --app "APP_ID" --price-point "PRICE_POINT_ID"`,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AppSetupPricingSetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// AppSetupPricingSetCommand returns the pricing set subcommand.
func AppSetupPricingSetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("app-setup pricing set", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID)")
	pricePointID := fs.String("price-point", "", "App price point ID")
	baseTerritory := fs.String("base-territory", "", "Base territory ID (e.g., USA)")
	startDate := fs.String("start-date", "", "Start date (YYYY-MM-DD, default: today)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "set",
		ShortUsage: "asc app-setup pricing set [flags]",
		ShortHelp:  "Set app pricing using a price point.",
		LongHelp: `Set app pricing using a price point.

Examples:
  asc app-setup pricing set --app "APP_ID" --price-point "PRICE_POINT_ID" --base-territory "USA"
  asc app-setup pricing set --app "APP_ID" --price-point "PRICE_POINT_ID" --base-territory "USA" --start-date "2024-03-01"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required (or set ASC_APP_ID)")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*pricePointID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --price-point is required")
				return flag.ErrHelp
			}

			startDateValue := strings.TrimSpace(*startDate)
			if startDateValue == "" {
				startDateValue = time.Now().Format("2006-01-02")
			}
			normalizedStartDate, err := normalizePricingStartDate(startDateValue)
			if err != nil {
				return fmt.Errorf("app-setup pricing set: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("app-setup pricing set: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			baseTerritoryID, err := resolveBaseTerritoryID(requestCtx, client, resolvedAppID, *baseTerritory)
			if err != nil {
				return fmt.Errorf("app-setup pricing set: %w", err)
			}

			resp, err := client.CreateAppPriceSchedule(requestCtx, resolvedAppID, asc.AppPriceScheduleCreateAttributes{
				PricePointID:    strings.TrimSpace(*pricePointID),
				StartDate:       normalizedStartDate,
				BaseTerritoryID: baseTerritoryID,
			})
			if err != nil {
				return fmt.Errorf("app-setup pricing set: %w", err)
			}

			return printOutput(resp, *output, *pretty)
		},
	}
}

// AppSetupLocalizationsCommand returns the localizations subcommand group.
func AppSetupLocalizationsCommand() *ffcli.Command {
	return &ffcli.Command{
		Name:       "localizations",
		ShortUsage: "asc app-setup localizations <subcommand> [flags]",
		ShortHelp:  "Upload app store localizations.",
		LongHelp: `Upload app store localizations (version or app-info).

Examples:
  asc app-setup localizations upload --version "VERSION_ID" --path "./localizations"
  asc app-setup localizations upload --app "APP_ID" --type app-info --path "./localizations"`,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AppSetupLocalizationsUploadCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// AppSetupLocalizationsUploadCommand returns the localizations upload subcommand.
func AppSetupLocalizationsUploadCommand() *ffcli.Command {
	fs := flag.NewFlagSet("app-setup localizations upload", flag.ExitOnError)

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
		ShortUsage: "asc app-setup localizations upload [flags]",
		ShortHelp:  "Upload localizations from .strings files.",
		LongHelp: `Upload localizations from .strings files.

Examples:
  asc app-setup localizations upload --version "VERSION_ID" --path "./localizations"
  asc app-setup localizations upload --app "APP_ID" --type app-info --path "./localizations"
  asc app-setup localizations upload --version "VERSION_ID" --locale "en-US" --path "en-US.strings"
  asc app-setup localizations upload --version "VERSION_ID" --path "./localizations" --dry-run`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*path) == "" {
				fmt.Fprintln(os.Stderr, "Error: --path is required")
				return flag.ErrHelp
			}

			normalizedType, err := normalizeLocalizationType(*locType)
			if err != nil {
				return fmt.Errorf("app-setup localizations upload: %w", err)
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
					return fmt.Errorf("app-setup localizations upload: %w", err)
				}

				requestCtx, cancel := contextWithTimeout(ctx)
				defer cancel()

				valuesByLocale, err := readLocalizationStrings(*path, locales)
				if err != nil {
					return fmt.Errorf("app-setup localizations upload: %w", err)
				}

				results, err := uploadVersionLocalizations(requestCtx, client, strings.TrimSpace(*versionID), valuesByLocale, *dryRun)
				if err != nil {
					return fmt.Errorf("app-setup localizations upload: %w", err)
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

				client, err := getASCClient()
				if err != nil {
					return fmt.Errorf("app-setup localizations upload: %w", err)
				}

				requestCtx, cancel := contextWithTimeout(ctx)
				defer cancel()

				appInfo, err := resolveAppInfoID(requestCtx, client, resolvedAppID, strings.TrimSpace(*appInfoID))
				if err != nil {
					return fmt.Errorf("app-setup localizations upload: %w", err)
				}

				valuesByLocale, err := readLocalizationStrings(*path, locales)
				if err != nil {
					return fmt.Errorf("app-setup localizations upload: %w", err)
				}

				results, err := uploadAppInfoLocalizations(requestCtx, client, appInfo, valuesByLocale, *dryRun)
				if err != nil {
					return fmt.Errorf("app-setup localizations upload: %w", err)
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
				return fmt.Errorf("app-setup localizations upload: unsupported type %q", normalizedType)
			}
		},
	}
}
