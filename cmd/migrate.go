package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

// MigrateCommand returns the migrate command with subcommands.
func MigrateCommand() *ffcli.Command {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "migrate",
		ShortUsage: "asc migrate <subcommand> [flags]",
		ShortHelp:  "Migrate metadata from/to fastlane format.",
		LongHelp: `Migrate metadata from/to fastlane directory structure.

This enables transitioning from fastlane's deliver tool to asc.

Examples:
  asc migrate import --app "APP_ID" --version "VERSION_ID" --fastlane-dir ./fastlane
  asc migrate export --app "APP_ID" --version "VERSION_ID" --output-dir ./fastlane`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			MigrateImportCommand(),
			MigrateExportCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// MigrateImportCommand returns the migrate import subcommand.
func MigrateImportCommand() *ffcli.Command {
	fs := flag.NewFlagSet("migrate import", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID)")
	versionID := fs.String("version-id", "", "App Store version ID (required)")
	fastlaneDir := fs.String("fastlane-dir", "", "Path to fastlane directory (required)")
	metadataOnly := fs.Bool("metadata-only", false, "Import only metadata (skip screenshots)")
	dryRun := fs.Bool("dry-run", false, "Preview changes without uploading")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "import",
		ShortUsage: "asc migrate import [flags]",
		ShortHelp:  "Import metadata from fastlane directory structure.",
		LongHelp: `Import metadata from fastlane directory structure.

Reads from the standard fastlane structure:
  fastlane/
  ├── metadata/
  │   ├── en-US/
  │   │   ├── description.txt
  │   │   ├── keywords.txt
  │   │   ├── name.txt
  │   │   ├── subtitle.txt
  │   │   ├── release_notes.txt
  │   │   ├── promotional_text.txt
  │   │   ├── support_url.txt
  │   │   ├── marketing_url.txt
  │   │   └── privacy_url.txt
  │   └── de-DE/
  │       └── ...
  └── screenshots/
      └── en-US/
          └── ...

Examples:
  asc migrate import --app "APP_ID" --version-id "VERSION_ID" --fastlane-dir ./fastlane
  asc migrate import --app "APP_ID" --version-id "VERSION_ID" --fastlane-dir ./fastlane --dry-run
  asc migrate import --app "APP_ID" --version-id "VERSION_ID" --fastlane-dir ./fastlane --metadata-only`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*versionID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --version-id is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*fastlaneDir) == "" {
				fmt.Fprintln(os.Stderr, "Error: --fastlane-dir is required")
				return flag.ErrHelp
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required (or set ASC_APP_ID)")
				return flag.ErrHelp
			}

			// Check if directory exists
			metadataDir := filepath.Join(*fastlaneDir, "metadata")
			if _, err := os.Stat(metadataDir); os.IsNotExist(err) {
				return fmt.Errorf("migrate import: metadata directory not found: %s", metadataDir)
			}

			// Read metadata from fastlane structure
			localizations, err := readFastlaneMetadata(metadataDir)
			if err != nil {
				return fmt.Errorf("migrate import: %w", err)
			}

			if *dryRun {
				result := &MigrateImportResult{
					DryRun:        true,
					VersionID:     strings.TrimSpace(*versionID),
					Localizations: localizations,
				}
				return printMigrateOutput(result, *output, *pretty)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("migrate import: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			// Fetch existing localizations to get their IDs
			existingLocs, err := client.GetAppStoreVersionLocalizations(requestCtx, strings.TrimSpace(*versionID))
			if err != nil {
				return fmt.Errorf("migrate import: failed to fetch existing localizations: %w", err)
			}

			// Build a map of locale -> localization ID
			localeToID := make(map[string]string)
			for _, loc := range existingLocs.Data {
				localeToID[loc.Attributes.Locale] = loc.ID
			}

			// Upload each localization
			uploaded := make([]LocalizationUploadItem, 0, len(localizations))
			for _, loc := range localizations {
				attrs := asc.AppStoreVersionLocalizationAttributes{
					Locale:          loc.Locale,
					Description:     loc.Description,
					Keywords:        loc.Keywords,
					WhatsNew:        loc.WhatsNew,
					PromotionalText: loc.PromotionalText,
					SupportURL:      loc.SupportURL,
					MarketingURL:    loc.MarketingURL,
				}

				// Check if localization already exists
				if existingID, exists := localeToID[loc.Locale]; exists {
					// Update existing localization
					_, err := client.UpdateAppStoreVersionLocalization(requestCtx, existingID, attrs)
					if err != nil {
						return fmt.Errorf("migrate import: failed to update %s: %w", loc.Locale, err)
					}
				} else {
					// Create new localization
					_, err := client.CreateAppStoreVersionLocalization(requestCtx, strings.TrimSpace(*versionID), attrs)
					if err != nil {
						return fmt.Errorf("migrate import: failed to create %s: %w", loc.Locale, err)
					}
				}

				uploaded = append(uploaded, LocalizationUploadItem{
					Locale: loc.Locale,
					Fields: countNonEmptyFields(loc),
				})
			}

			result := &MigrateImportResult{
				DryRun:        false,
				VersionID:     strings.TrimSpace(*versionID),
				Localizations: localizations,
				Uploaded:      uploaded,
			}

			// Note: Screenshot import not implemented in this version
			_ = metadataOnly

			return printMigrateOutput(result, *output, *pretty)
		},
	}
}

// MigrateExportCommand returns the migrate export subcommand.
func MigrateExportCommand() *ffcli.Command {
	fs := flag.NewFlagSet("migrate export", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID)")
	versionID := fs.String("version-id", "", "App Store version ID (required)")
	outputDir := fs.String("output-dir", "", "Output directory for fastlane structure (required)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "export",
		ShortUsage: "asc migrate export [flags]",
		ShortHelp:  "Export metadata to fastlane directory structure.",
		LongHelp: `Export current App Store metadata to fastlane directory structure.

Creates the standard fastlane structure with all localizations.

Examples:
  asc migrate export --app "APP_ID" --version-id "VERSION_ID" --output-dir ./fastlane`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*versionID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --version-id is required")
				return flag.ErrHelp
			}
			if strings.TrimSpace(*outputDir) == "" {
				fmt.Fprintln(os.Stderr, "Error: --output-dir is required")
				return flag.ErrHelp
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required (or set ASC_APP_ID)")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("migrate export: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			// Fetch all localizations
			resp, err := client.GetAppStoreVersionLocalizations(requestCtx, strings.TrimSpace(*versionID))
			if err != nil {
				return fmt.Errorf("migrate export: %w", err)
			}

			// Create output directory structure
			metadataDir := filepath.Join(*outputDir, "metadata")
			if err := os.MkdirAll(metadataDir, 0755); err != nil {
				return fmt.Errorf("migrate export: failed to create directory: %w", err)
			}

			// Write each localization
			exported := make([]string, 0, len(resp.Data))
			for _, loc := range resp.Data {
				locale := loc.Attributes.Locale
				localeDir := filepath.Join(metadataDir, locale)
				if err := os.MkdirAll(localeDir, 0755); err != nil {
					return fmt.Errorf("migrate export: failed to create locale directory: %w", err)
				}

				// Write files
				writeIfNotEmpty(filepath.Join(localeDir, "description.txt"), loc.Attributes.Description)
				writeIfNotEmpty(filepath.Join(localeDir, "keywords.txt"), loc.Attributes.Keywords)
				writeIfNotEmpty(filepath.Join(localeDir, "release_notes.txt"), loc.Attributes.WhatsNew)
				writeIfNotEmpty(filepath.Join(localeDir, "promotional_text.txt"), loc.Attributes.PromotionalText)
				writeIfNotEmpty(filepath.Join(localeDir, "support_url.txt"), loc.Attributes.SupportURL)
				writeIfNotEmpty(filepath.Join(localeDir, "marketing_url.txt"), loc.Attributes.MarketingURL)

				exported = append(exported, locale)
			}

			result := &MigrateExportResult{
				VersionID:   strings.TrimSpace(*versionID),
				OutputDir:   *outputDir,
				Locales:     exported,
				TotalFiles:  len(exported) * 6, // 6 files per locale
			}

			return printMigrateOutput(result, *output, *pretty)
		},
	}
}

// FastlaneLocalization holds metadata read from fastlane structure.
type FastlaneLocalization struct {
	Locale          string `json:"locale"`
	Description     string `json:"description,omitempty"`
	Keywords        string `json:"keywords,omitempty"`
	WhatsNew        string `json:"whatsNew,omitempty"`
	PromotionalText string `json:"promotionalText,omitempty"`
	SupportURL      string `json:"supportUrl,omitempty"`
	MarketingURL    string `json:"marketingUrl,omitempty"`
	Name            string `json:"name,omitempty"`
	Subtitle        string `json:"subtitle,omitempty"`
}

// LocalizationUploadItem represents an uploaded localization.
type LocalizationUploadItem struct {
	Locale string `json:"locale"`
	Fields int    `json:"fields"`
}

// MigrateImportResult is the result of a migrate import operation.
type MigrateImportResult struct {
	DryRun        bool                   `json:"dryRun"`
	VersionID     string                 `json:"versionId"`
	Localizations []FastlaneLocalization `json:"localizations"`
	Uploaded      []LocalizationUploadItem `json:"uploaded,omitempty"`
}

// MigrateExportResult is the result of a migrate export operation.
type MigrateExportResult struct {
	VersionID  string   `json:"versionId"`
	OutputDir  string   `json:"outputDir"`
	Locales    []string `json:"locales"`
	TotalFiles int      `json:"totalFiles"`
}

// readFastlaneMetadata reads metadata from a fastlane metadata directory.
func readFastlaneMetadata(metadataDir string) ([]FastlaneLocalization, error) {
	entries, err := os.ReadDir(metadataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata directory: %w", err)
	}

	var localizations []FastlaneLocalization
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		locale := entry.Name()
		if locale == "review_information" || locale == "default" {
			continue // Skip special directories
		}

		localeDir := filepath.Join(metadataDir, locale)
		loc := FastlaneLocalization{Locale: locale}

		// Read each metadata file
		loc.Description = readFileIfExists(filepath.Join(localeDir, "description.txt"))
		loc.Keywords = readFileIfExists(filepath.Join(localeDir, "keywords.txt"))
		loc.WhatsNew = readFileIfExists(filepath.Join(localeDir, "release_notes.txt"))
		loc.PromotionalText = readFileIfExists(filepath.Join(localeDir, "promotional_text.txt"))
		loc.SupportURL = readFileIfExists(filepath.Join(localeDir, "support_url.txt"))
		loc.MarketingURL = readFileIfExists(filepath.Join(localeDir, "marketing_url.txt"))
		loc.Name = readFileIfExists(filepath.Join(localeDir, "name.txt"))
		loc.Subtitle = readFileIfExists(filepath.Join(localeDir, "subtitle.txt"))

		localizations = append(localizations, loc)
	}

	return localizations, nil
}

// readFileIfExists reads a file's contents if it exists, returning empty string otherwise.
func readFileIfExists(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// writeIfNotEmpty writes content to a file only if the content is not empty.
func writeIfNotEmpty(path, content string) error {
	if content == "" {
		return nil
	}
	return os.WriteFile(path, []byte(content+"\n"), 0644)
}

// printMigrateOutput handles output for migrate-specific result types.
func printMigrateOutput(data interface{}, format string, pretty bool) error {
	format = strings.ToLower(format)
	switch format {
	case "json":
		if pretty {
			return asc.PrintPrettyJSON(data)
		}
		return asc.PrintJSON(data)
	case "markdown", "md":
		switch v := data.(type) {
		case *MigrateImportResult:
			return printMigrateImportResultMarkdown(v)
		case *MigrateExportResult:
			return printMigrateExportResultMarkdown(v)
		default:
			return asc.PrintJSON(data)
		}
	case "table":
		switch v := data.(type) {
		case *MigrateImportResult:
			return printMigrateImportResultTable(v)
		case *MigrateExportResult:
			return printMigrateExportResultTable(v)
		default:
			return asc.PrintJSON(data)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// countNonEmptyFields counts the number of non-empty fields in a localization.
func countNonEmptyFields(loc FastlaneLocalization) int {
	count := 0
	if loc.Description != "" {
		count++
	}
	if loc.Keywords != "" {
		count++
	}
	if loc.WhatsNew != "" {
		count++
	}
	if loc.PromotionalText != "" {
		count++
	}
	if loc.SupportURL != "" {
		count++
	}
	if loc.MarketingURL != "" {
		count++
	}
	if loc.Name != "" {
		count++
	}
	if loc.Subtitle != "" {
		count++
	}
	return count
}
