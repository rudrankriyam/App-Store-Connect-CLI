package cmd

import (
	"context"
	"flag"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

const assetUploadDefaultTimeout = 10 * time.Minute
const assetPollInterval = 2 * time.Second

// AssetsCommand returns the assets command with subcommands.
func AssetsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("assets", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "assets",
		ShortUsage: "asc assets <subcommand> [flags]",
		ShortHelp:  "Manage App Store assets (screenshots, previews).",
		LongHelp: `Manage App Store metadata assets (screenshots and app previews).

Examples:
  asc assets screenshots list --version-localization "LOC_ID"
  asc assets screenshots upload --version-localization "LOC_ID" --path "./screenshots" --device-type "IPHONE_65"
  asc assets previews upload --version-localization "LOC_ID" --path "./previews" --device-type "IPHONE_65"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AssetsScreenshotsCommand(),
			AssetsPreviewsCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// AssetsScreenshotsCommand returns the screenshots subcommand group.
func AssetsScreenshotsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("screenshots", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "screenshots",
		ShortUsage: "asc assets screenshots <subcommand> [flags]",
		ShortHelp:  "Manage App Store screenshots.",
		LongHelp: `Manage App Store screenshots.

Examples:
  asc assets screenshots list --version-localization "LOC_ID"
  asc assets screenshots upload --version-localization "LOC_ID" --path "./screenshots" --device-type "IPHONE_65"
  asc assets screenshots delete --id "SCREENSHOT_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AssetsScreenshotsListCommand(),
			AssetsScreenshotsUploadCommand(),
			AssetsScreenshotsDeleteCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// AssetsPreviewsCommand returns the previews subcommand group.
func AssetsPreviewsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("previews", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "previews",
		ShortUsage: "asc assets previews <subcommand> [flags]",
		ShortHelp:  "Manage App Store app previews.",
		LongHelp: `Manage App Store app previews.

Examples:
  asc assets previews list --version-localization "LOC_ID"
  asc assets previews upload --version-localization "LOC_ID" --path "./previews" --device-type "IPHONE_65"
  asc assets previews delete --id "PREVIEW_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AssetsPreviewsListCommand(),
			AssetsPreviewsUploadCommand(),
			AssetsPreviewsDeleteCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// AssetsScreenshotsListCommand returns the screenshots list subcommand.
func AssetsScreenshotsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	localizationID := fs.String("version-localization", "", "App Store version localization ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc assets screenshots list --version-localization \"LOC_ID\"",
		ShortHelp:  "List screenshots for a localization.",
		LongHelp: `List screenshots for a localization.

Examples:
  asc assets screenshots list --version-localization "LOC_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			locID := strings.TrimSpace(*localizationID)
			if locID == "" {
				fmt.Fprintln(os.Stderr, "Error: --version-localization is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("assets screenshots list: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			setsResp, err := client.GetAppScreenshotSets(requestCtx, locID)
			if err != nil {
				return fmt.Errorf("assets screenshots list: failed to fetch sets: %w", err)
			}

			result := asc.AppScreenshotListResult{
				VersionLocalizationID: locID,
				Sets:                  make([]asc.AppScreenshotSetWithScreenshots, 0, len(setsResp.Data)),
			}

			for _, set := range setsResp.Data {
				screenshots, err := client.GetAppScreenshots(requestCtx, set.ID)
				if err != nil {
					return fmt.Errorf("assets screenshots list: failed to fetch screenshots for set %s: %w", set.ID, err)
				}
				result.Sets = append(result.Sets, asc.AppScreenshotSetWithScreenshots{
					Set:         set,
					Screenshots: screenshots.Data,
				})
			}

			return printOutput(&result, *output, *pretty)
		},
	}
}

// AssetsPreviewsListCommand returns the previews list subcommand.
func AssetsPreviewsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)

	localizationID := fs.String("version-localization", "", "App Store version localization ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc assets previews list --version-localization \"LOC_ID\"",
		ShortHelp:  "List previews for a localization.",
		LongHelp: `List previews for a localization.

Examples:
  asc assets previews list --version-localization "LOC_ID"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			locID := strings.TrimSpace(*localizationID)
			if locID == "" {
				fmt.Fprintln(os.Stderr, "Error: --version-localization is required")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("assets previews list: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			setsResp, err := client.GetAppPreviewSets(requestCtx, locID)
			if err != nil {
				return fmt.Errorf("assets previews list: failed to fetch sets: %w", err)
			}

			result := asc.AppPreviewListResult{
				VersionLocalizationID: locID,
				Sets:                  make([]asc.AppPreviewSetWithPreviews, 0, len(setsResp.Data)),
			}

			for _, set := range setsResp.Data {
				previews, err := client.GetAppPreviews(requestCtx, set.ID)
				if err != nil {
					return fmt.Errorf("assets previews list: failed to fetch previews for set %s: %w", set.ID, err)
				}
				result.Sets = append(result.Sets, asc.AppPreviewSetWithPreviews{
					Set:      set,
					Previews: previews.Data,
				})
			}

			return printOutput(&result, *output, *pretty)
		},
	}
}

// AssetsScreenshotsUploadCommand returns the screenshots upload subcommand.
func AssetsScreenshotsUploadCommand() *ffcli.Command {
	fs := flag.NewFlagSet("upload", flag.ExitOnError)

	localizationID := fs.String("version-localization", "", "App Store version localization ID")
	path := fs.String("path", "", "Path to screenshot file or directory")
	deviceType := fs.String("device-type", "", "Device type (e.g., IPHONE_65)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "upload",
		ShortUsage: "asc assets screenshots upload --version-localization \"LOC_ID\" --path \"./screenshots\" --device-type \"IPHONE_65\"",
		ShortHelp:  "Upload screenshots for a localization.",
		LongHelp: `Upload screenshots for a localization.

Examples:
  asc assets screenshots upload --version-localization "LOC_ID" --path "./screenshots" --device-type "IPHONE_65"
  asc assets screenshots upload --version-localization "LOC_ID" --path "./screenshots/en-US.png" --device-type "IPHONE_65"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			locID := strings.TrimSpace(*localizationID)
			if locID == "" {
				fmt.Fprintln(os.Stderr, "Error: --version-localization is required")
				return flag.ErrHelp
			}
			pathValue := strings.TrimSpace(*path)
			if pathValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --path is required")
				return flag.ErrHelp
			}
			deviceValue := strings.TrimSpace(*deviceType)
			if deviceValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --device-type is required")
				return flag.ErrHelp
			}

			displayType, err := normalizeScreenshotDisplayType(deviceValue)
			if err != nil {
				return fmt.Errorf("assets screenshots upload: %w", err)
			}

			files, err := collectAssetFiles(pathValue)
			if err != nil {
				return fmt.Errorf("assets screenshots upload: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("assets screenshots upload: %w", err)
			}

			requestCtx, cancel := contextWithAssetUploadTimeout(ctx)
			defer cancel()

			set, err := ensureScreenshotSet(requestCtx, client, locID, displayType)
			if err != nil {
				return fmt.Errorf("assets screenshots upload: %w", err)
			}

			results := make([]asc.AssetUploadResultItem, 0, len(files))
			for _, filePath := range files {
				item, err := uploadScreenshotAsset(requestCtx, client, set.ID, filePath)
				if err != nil {
					return fmt.Errorf("assets screenshots upload: %w", err)
				}
				results = append(results, item)
			}

			result := asc.AppScreenshotUploadResult{
				VersionLocalizationID: locID,
				SetID:                 set.ID,
				DisplayType:           set.Attributes.ScreenshotDisplayType,
				Results:               results,
			}

			return printOutput(&result, *output, *pretty)
		},
	}
}

// AssetsPreviewsUploadCommand returns the previews upload subcommand.
func AssetsPreviewsUploadCommand() *ffcli.Command {
	fs := flag.NewFlagSet("upload", flag.ExitOnError)

	localizationID := fs.String("version-localization", "", "App Store version localization ID")
	path := fs.String("path", "", "Path to preview file or directory")
	deviceType := fs.String("device-type", "", "Device type (e.g., IPHONE_65)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "upload",
		ShortUsage: "asc assets previews upload --version-localization \"LOC_ID\" --path \"./previews\" --device-type \"IPHONE_65\"",
		ShortHelp:  "Upload previews for a localization.",
		LongHelp: `Upload previews for a localization.

Examples:
  asc assets previews upload --version-localization "LOC_ID" --path "./previews" --device-type "IPHONE_65"
  asc assets previews upload --version-localization "LOC_ID" --path "./previews/preview.mov" --device-type "IPHONE_65"`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			locID := strings.TrimSpace(*localizationID)
			if locID == "" {
				fmt.Fprintln(os.Stderr, "Error: --version-localization is required")
				return flag.ErrHelp
			}
			pathValue := strings.TrimSpace(*path)
			if pathValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --path is required")
				return flag.ErrHelp
			}
			deviceValue := strings.TrimSpace(*deviceType)
			if deviceValue == "" {
				fmt.Fprintln(os.Stderr, "Error: --device-type is required")
				return flag.ErrHelp
			}

			previewType, err := normalizePreviewType(deviceValue)
			if err != nil {
				return fmt.Errorf("assets previews upload: %w", err)
			}

			files, err := collectAssetFiles(pathValue)
			if err != nil {
				return fmt.Errorf("assets previews upload: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("assets previews upload: %w", err)
			}

			requestCtx, cancel := contextWithAssetUploadTimeout(ctx)
			defer cancel()

			set, err := ensurePreviewSet(requestCtx, client, locID, previewType)
			if err != nil {
				return fmt.Errorf("assets previews upload: %w", err)
			}

			results := make([]asc.AssetUploadResultItem, 0, len(files))
			for _, filePath := range files {
				item, err := uploadPreviewAsset(requestCtx, client, set.ID, filePath)
				if err != nil {
					return fmt.Errorf("assets previews upload: %w", err)
				}
				results = append(results, item)
			}

			result := asc.AppPreviewUploadResult{
				VersionLocalizationID: locID,
				SetID:                 set.ID,
				PreviewType:           set.Attributes.PreviewType,
				Results:               results,
			}

			return printOutput(&result, *output, *pretty)
		},
	}
}

// AssetsScreenshotsDeleteCommand returns the screenshot delete subcommand.
func AssetsScreenshotsDeleteCommand() *ffcli.Command {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)

	id := fs.String("id", "", "Screenshot ID")
	confirm := fs.Bool("confirm", false, "Confirm deletion")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "delete",
		ShortUsage: "asc assets screenshots delete --id \"SCREENSHOT_ID\" --confirm",
		ShortHelp:  "Delete a screenshot by ID.",
		LongHelp: `Delete a screenshot by ID.

Examples:
  asc assets screenshots delete --id "SCREENSHOT_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			assetID := strings.TrimSpace(*id)
			if assetID == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}
			if !*confirm {
				fmt.Fprintln(os.Stderr, "Error: --confirm is required to delete")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("assets screenshots delete: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			if err := client.DeleteAppScreenshot(requestCtx, assetID); err != nil {
				return fmt.Errorf("assets screenshots delete: %w", err)
			}

			result := asc.AssetDeleteResult{
				ID:      assetID,
				Deleted: true,
			}

			return printOutput(&result, *output, *pretty)
		},
	}
}

// AssetsPreviewsDeleteCommand returns the preview delete subcommand.
func AssetsPreviewsDeleteCommand() *ffcli.Command {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)

	id := fs.String("id", "", "Preview ID")
	confirm := fs.Bool("confirm", false, "Confirm deletion")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "delete",
		ShortUsage: "asc assets previews delete --id \"PREVIEW_ID\" --confirm",
		ShortHelp:  "Delete a preview by ID.",
		LongHelp: `Delete a preview by ID.

Examples:
  asc assets previews delete --id "PREVIEW_ID" --confirm`,
		FlagSet:   fs,
		UsageFunc: DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			assetID := strings.TrimSpace(*id)
			if assetID == "" {
				fmt.Fprintln(os.Stderr, "Error: --id is required")
				return flag.ErrHelp
			}
			if !*confirm {
				fmt.Fprintln(os.Stderr, "Error: --confirm is required to delete")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("assets previews delete: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			if err := client.DeleteAppPreview(requestCtx, assetID); err != nil {
				return fmt.Errorf("assets previews delete: %w", err)
			}

			result := asc.AssetDeleteResult{
				ID:      assetID,
				Deleted: true,
			}

			return printOutput(&result, *output, *pretty)
		},
	}
}

func contextWithAssetUploadTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, asc.ResolveTimeoutWithDefault(assetUploadDefaultTimeout))
}

func normalizeScreenshotDisplayType(input string) (string, error) {
	value := strings.ToUpper(strings.TrimSpace(input))
	if value == "" {
		return "", fmt.Errorf("device type is required")
	}
	if !strings.HasPrefix(value, "APP_") {
		value = "APP_" + value
	}
	if !asc.IsValidScreenshotDisplayType(value) {
		return "", fmt.Errorf("unsupported screenshot display type %q", value)
	}
	return value, nil
}

func normalizePreviewType(input string) (string, error) {
	value := strings.ToUpper(strings.TrimSpace(input))
	if value == "" {
		return "", fmt.Errorf("device type is required")
	}
	value = strings.TrimPrefix(value, "APP_")
	if !asc.IsValidPreviewType(value) {
		return "", fmt.Errorf("unsupported preview type %q", value)
	}
	return value, nil
}

func collectAssetFiles(path string) ([]string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("refusing to read symlink %q", path)
	}
	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		files := make([]string, 0, len(entries))
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			fullPath := filepath.Join(path, entry.Name())
			if err := asc.ValidateImageFile(fullPath); err != nil {
				return nil, err
			}
			files = append(files, fullPath)
		}
		if len(files) == 0 {
			return nil, fmt.Errorf("no files found in %q", path)
		}
		sort.Strings(files)
		return files, nil
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("expected regular file: %q", path)
	}
	if err := asc.ValidateImageFile(path); err != nil {
		return nil, err
	}
	return []string{path}, nil
}

func ensureScreenshotSet(ctx context.Context, client *asc.Client, localizationID, displayType string) (asc.Resource[asc.AppScreenshotSetAttributes], error) {
	resp, err := client.GetAppScreenshotSets(ctx, localizationID)
	if err != nil {
		return asc.Resource[asc.AppScreenshotSetAttributes]{}, err
	}
	for _, set := range resp.Data {
		if strings.EqualFold(set.Attributes.ScreenshotDisplayType, displayType) {
			return set, nil
		}
	}
	created, err := client.CreateAppScreenshotSet(ctx, localizationID, displayType)
	if err != nil {
		return asc.Resource[asc.AppScreenshotSetAttributes]{}, err
	}
	return created.Data, nil
}

func ensurePreviewSet(ctx context.Context, client *asc.Client, localizationID, previewType string) (asc.Resource[asc.AppPreviewSetAttributes], error) {
	resp, err := client.GetAppPreviewSets(ctx, localizationID)
	if err != nil {
		return asc.Resource[asc.AppPreviewSetAttributes]{}, err
	}
	for _, set := range resp.Data {
		if strings.EqualFold(set.Attributes.PreviewType, previewType) {
			return set, nil
		}
	}
	created, err := client.CreateAppPreviewSet(ctx, localizationID, previewType)
	if err != nil {
		return asc.Resource[asc.AppPreviewSetAttributes]{}, err
	}
	return created.Data, nil
}

func uploadScreenshotAsset(ctx context.Context, client *asc.Client, setID, filePath string) (asc.AssetUploadResultItem, error) {
	if err := asc.ValidateImageFile(filePath); err != nil {
		return asc.AssetUploadResultItem{}, err
	}

	file, err := openExistingNoFollow(filePath)
	if err != nil {
		return asc.AssetUploadResultItem{}, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return asc.AssetUploadResultItem{}, err
	}

	checksum, err := asc.ComputeChecksumFromReader(file, asc.ChecksumAlgorithmMD5)
	if err != nil {
		return asc.AssetUploadResultItem{}, err
	}

	created, err := client.CreateAppScreenshot(ctx, setID, info.Name(), info.Size())
	if err != nil {
		return asc.AssetUploadResultItem{}, err
	}
	if len(created.Data.Attributes.UploadOperations) == 0 {
		return asc.AssetUploadResultItem{}, fmt.Errorf("no upload operations returned for %q", info.Name())
	}

	if err := asc.UploadAssetFromFile(ctx, file, info.Size(), created.Data.Attributes.UploadOperations); err != nil {
		return asc.AssetUploadResultItem{}, err
	}

	if _, err := client.UpdateAppScreenshot(ctx, created.Data.ID, true, checksum); err != nil {
		return asc.AssetUploadResultItem{}, err
	}

	state, err := waitForScreenshotDelivery(ctx, client, created.Data.ID)
	if err != nil {
		return asc.AssetUploadResultItem{}, err
	}

	return asc.AssetUploadResultItem{
		FileName: info.Name(),
		FilePath: filePath,
		AssetID:  created.Data.ID,
		State:    state,
	}, nil
}

func uploadPreviewAsset(ctx context.Context, client *asc.Client, setID, filePath string) (asc.AssetUploadResultItem, error) {
	if err := asc.ValidateImageFile(filePath); err != nil {
		return asc.AssetUploadResultItem{}, err
	}

	mimeType, err := detectPreviewMimeType(filePath)
	if err != nil {
		return asc.AssetUploadResultItem{}, err
	}

	file, err := openExistingNoFollow(filePath)
	if err != nil {
		return asc.AssetUploadResultItem{}, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return asc.AssetUploadResultItem{}, err
	}

	checksum, err := asc.ComputeChecksumFromReader(file, asc.ChecksumAlgorithmMD5)
	if err != nil {
		return asc.AssetUploadResultItem{}, err
	}

	created, err := client.CreateAppPreview(ctx, setID, info.Name(), info.Size(), mimeType)
	if err != nil {
		return asc.AssetUploadResultItem{}, err
	}
	if len(created.Data.Attributes.UploadOperations) == 0 {
		return asc.AssetUploadResultItem{}, fmt.Errorf("no upload operations returned for %q", info.Name())
	}

	if err := asc.UploadAssetFromFile(ctx, file, info.Size(), created.Data.Attributes.UploadOperations); err != nil {
		return asc.AssetUploadResultItem{}, err
	}

	if _, err := client.UpdateAppPreview(ctx, created.Data.ID, true, checksum); err != nil {
		return asc.AssetUploadResultItem{}, err
	}

	state, err := waitForPreviewDelivery(ctx, client, created.Data.ID)
	if err != nil {
		return asc.AssetUploadResultItem{}, err
	}

	return asc.AssetUploadResultItem{
		FileName: info.Name(),
		FilePath: filePath,
		AssetID:  created.Data.ID,
		State:    state,
	}, nil
}

func detectPreviewMimeType(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return "", fmt.Errorf("preview file %q is missing an extension", path)
	}
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "", fmt.Errorf("unsupported preview file extension %q", ext)
	}
	if idx := strings.Index(mimeType, ";"); idx > 0 {
		mimeType = mimeType[:idx]
	}
	return mimeType, nil
}

func waitForScreenshotDelivery(ctx context.Context, client *asc.Client, screenshotID string) (string, error) {
	return waitForAssetDeliveryState(ctx, screenshotID, func(ctx context.Context) (*asc.AssetDeliveryState, error) {
		resp, err := client.GetAppScreenshot(ctx, screenshotID)
		if err != nil {
			return nil, err
		}
		return resp.Data.Attributes.AssetDeliveryState, nil
	})
}

func waitForPreviewDelivery(ctx context.Context, client *asc.Client, previewID string) (string, error) {
	return waitForAssetDeliveryState(ctx, previewID, func(ctx context.Context) (*asc.AssetDeliveryState, error) {
		resp, err := client.GetAppPreview(ctx, previewID)
		if err != nil {
			return nil, err
		}
		return resp.Data.Attributes.AssetDeliveryState, nil
	})
}

func waitForAssetDeliveryState(ctx context.Context, assetID string, fetch func(context.Context) (*asc.AssetDeliveryState, error)) (string, error) {
	ticker := time.NewTicker(assetPollInterval)
	defer ticker.Stop()

	var lastState string
	for {
		state, err := fetch(ctx)
		if err != nil {
			return lastState, err
		}
		if state != nil {
			lastState = state.State
			switch strings.ToUpper(state.State) {
			case "COMPLETE":
				return state.State, nil
			case "FAILED":
				return state.State, fmt.Errorf("asset %s delivery failed: %s", assetID, formatAssetErrors(state.Errors))
			}
		}

		select {
		case <-ctx.Done():
			return lastState, fmt.Errorf("timed out waiting for asset %s delivery: %w", assetID, ctx.Err())
		case <-ticker.C:
		}
	}
}

func formatAssetErrors(errors []asc.ErrorDetail) string {
	if len(errors) == 0 {
		return "unknown error"
	}
	parts := make([]string, 0, len(errors))
	for _, item := range errors {
		if item.Code != "" && item.Message != "" {
			parts = append(parts, fmt.Sprintf("%s: %s", item.Code, item.Message))
			continue
		}
		if item.Message != "" {
			parts = append(parts, item.Message)
			continue
		}
		if item.Code != "" {
			parts = append(parts, item.Code)
		}
	}
	if len(parts) == 0 {
		return "unknown error"
	}
	return strings.Join(parts, "; ")
}
