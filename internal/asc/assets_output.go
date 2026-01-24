package asc

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// AppScreenshotSetWithScreenshots groups a set with its screenshots.
type AppScreenshotSetWithScreenshots struct {
	Set         Resource[AppScreenshotSetAttributes] `json:"set"`
	Screenshots []Resource[AppScreenshotAttributes]  `json:"screenshots"`
}

// AppScreenshotListResult represents screenshot list output by localization.
type AppScreenshotListResult struct {
	VersionLocalizationID string                            `json:"versionLocalizationId"`
	Sets                  []AppScreenshotSetWithScreenshots `json:"sets"`
}

// AppPreviewSetWithPreviews groups a set with its previews.
type AppPreviewSetWithPreviews struct {
	Set      Resource[AppPreviewSetAttributes] `json:"set"`
	Previews []Resource[AppPreviewAttributes]  `json:"previews"`
}

// AppPreviewListResult represents preview list output by localization.
type AppPreviewListResult struct {
	VersionLocalizationID string                      `json:"versionLocalizationId"`
	Sets                  []AppPreviewSetWithPreviews `json:"sets"`
}

// AssetUploadResultItem represents a single uploaded asset.
type AssetUploadResultItem struct {
	FileName string `json:"fileName"`
	FilePath string `json:"filePath"`
	AssetID  string `json:"assetId"`
	State    string `json:"state,omitempty"`
}

// AppScreenshotUploadResult represents screenshot upload output.
type AppScreenshotUploadResult struct {
	VersionLocalizationID string                  `json:"versionLocalizationId"`
	SetID                 string                  `json:"setId"`
	DisplayType           string                  `json:"displayType"`
	Results               []AssetUploadResultItem `json:"results"`
}

// AppPreviewUploadResult represents preview upload output.
type AppPreviewUploadResult struct {
	VersionLocalizationID string                  `json:"versionLocalizationId"`
	SetID                 string                  `json:"setId"`
	PreviewType           string                  `json:"previewType"`
	Results               []AssetUploadResultItem `json:"results"`
}

// AssetDeleteResult represents deletion output for assets.
type AssetDeleteResult struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

func printAppScreenshotSetsTable(resp *AppScreenshotSetsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDisplay Type")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\n", item.ID, item.Attributes.ScreenshotDisplayType)
	}
	return w.Flush()
}

func printAppScreenshotSetsMarkdown(resp *AppScreenshotSetsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Display Type |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.ScreenshotDisplayType),
		)
	}
	return nil
}

func printAppScreenshotsTable(resp *AppScreenshotsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tFile Name\tFile Size\tState")
	for _, item := range resp.Data {
		state := ""
		if item.Attributes.AssetDeliveryState != nil {
			state = item.Attributes.AssetDeliveryState.State
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
			item.ID,
			item.Attributes.FileName,
			item.Attributes.FileSize,
			state,
		)
	}
	return w.Flush()
}

func printAppScreenshotsMarkdown(resp *AppScreenshotsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | File Name | File Size | State |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		state := ""
		if item.Attributes.AssetDeliveryState != nil {
			state = item.Attributes.AssetDeliveryState.State
		}
		fmt.Fprintf(os.Stdout, "| %s | %s | %d | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.FileName),
			item.Attributes.FileSize,
			escapeMarkdown(state),
		)
	}
	return nil
}

func printAppPreviewSetsTable(resp *AppPreviewSetsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tPreview Type")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\n", item.ID, item.Attributes.PreviewType)
	}
	return w.Flush()
}

func printAppPreviewSetsMarkdown(resp *AppPreviewSetsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Preview Type |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.PreviewType),
		)
	}
	return nil
}

func printAppPreviewsTable(resp *AppPreviewsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tFile Name\tFile Size\tState")
	for _, item := range resp.Data {
		state := ""
		if item.Attributes.AssetDeliveryState != nil {
			state = item.Attributes.AssetDeliveryState.State
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
			item.ID,
			item.Attributes.FileName,
			item.Attributes.FileSize,
			state,
		)
	}
	return w.Flush()
}

func printAppPreviewsMarkdown(resp *AppPreviewsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | File Name | File Size | State |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		state := ""
		if item.Attributes.AssetDeliveryState != nil {
			state = item.Attributes.AssetDeliveryState.State
		}
		fmt.Fprintf(os.Stdout, "| %s | %s | %d | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.FileName),
			item.Attributes.FileSize,
			escapeMarkdown(state),
		)
	}
	return nil
}

func printAppScreenshotListResultTable(result *AppScreenshotListResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Set ID\tDisplay Type\tScreenshot ID\tFile Name\tFile Size\tState")
	for _, set := range result.Sets {
		displayType := set.Set.Attributes.ScreenshotDisplayType
		if len(set.Screenshots) == 0 {
			fmt.Fprintf(w, "%s\t%s\t\t\t\t\n", set.Set.ID, displayType)
			continue
		}
		for _, item := range set.Screenshots {
			state := ""
			if item.Attributes.AssetDeliveryState != nil {
				state = item.Attributes.AssetDeliveryState.State
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
				set.Set.ID,
				displayType,
				item.ID,
				item.Attributes.FileName,
				item.Attributes.FileSize,
				state,
			)
		}
	}
	return w.Flush()
}

func printAppScreenshotListResultMarkdown(result *AppScreenshotListResult) error {
	fmt.Fprintln(os.Stdout, "| Set ID | Display Type | Screenshot ID | File Name | File Size | State |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- |")
	for _, set := range result.Sets {
		displayType := set.Set.Attributes.ScreenshotDisplayType
		if len(set.Screenshots) == 0 {
			fmt.Fprintf(os.Stdout, "| %s | %s |  |  |  |  |\n",
				escapeMarkdown(set.Set.ID),
				escapeMarkdown(displayType),
			)
			continue
		}
		for _, item := range set.Screenshots {
			state := ""
			if item.Attributes.AssetDeliveryState != nil {
				state = item.Attributes.AssetDeliveryState.State
			}
			fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %d | %s |\n",
				escapeMarkdown(set.Set.ID),
				escapeMarkdown(displayType),
				escapeMarkdown(item.ID),
				escapeMarkdown(item.Attributes.FileName),
				item.Attributes.FileSize,
				escapeMarkdown(state),
			)
		}
	}
	return nil
}

func printAppPreviewListResultTable(result *AppPreviewListResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Set ID\tPreview Type\tPreview ID\tFile Name\tFile Size\tState")
	for _, set := range result.Sets {
		previewType := set.Set.Attributes.PreviewType
		if len(set.Previews) == 0 {
			fmt.Fprintf(w, "%s\t%s\t\t\t\t\n", set.Set.ID, previewType)
			continue
		}
		for _, item := range set.Previews {
			state := ""
			if item.Attributes.AssetDeliveryState != nil {
				state = item.Attributes.AssetDeliveryState.State
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
				set.Set.ID,
				previewType,
				item.ID,
				item.Attributes.FileName,
				item.Attributes.FileSize,
				state,
			)
		}
	}
	return w.Flush()
}

func printAppPreviewListResultMarkdown(result *AppPreviewListResult) error {
	fmt.Fprintln(os.Stdout, "| Set ID | Preview Type | Preview ID | File Name | File Size | State |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- |")
	for _, set := range result.Sets {
		previewType := set.Set.Attributes.PreviewType
		if len(set.Previews) == 0 {
			fmt.Fprintf(os.Stdout, "| %s | %s |  |  |  |  |\n",
				escapeMarkdown(set.Set.ID),
				escapeMarkdown(previewType),
			)
			continue
		}
		for _, item := range set.Previews {
			state := ""
			if item.Attributes.AssetDeliveryState != nil {
				state = item.Attributes.AssetDeliveryState.State
			}
			fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %d | %s |\n",
				escapeMarkdown(set.Set.ID),
				escapeMarkdown(previewType),
				escapeMarkdown(item.ID),
				escapeMarkdown(item.Attributes.FileName),
				item.Attributes.FileSize,
				escapeMarkdown(state),
			)
		}
	}
	return nil
}

func printAppScreenshotUploadResultTable(result *AppScreenshotUploadResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Localization ID\tSet ID\tDisplay Type")
	fmt.Fprintf(w, "%s\t%s\t%s\n", result.VersionLocalizationID, result.SetID, result.DisplayType)
	if err := w.Flush(); err != nil {
		return err
	}
	if len(result.Results) == 0 {
		return nil
	}
	fmt.Fprintln(os.Stdout, "\nScreenshots")
	items := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(items, "File Name\tAsset ID\tState")
	for _, item := range result.Results {
		fmt.Fprintf(items, "%s\t%s\t%s\n", item.FileName, item.AssetID, item.State)
	}
	return items.Flush()
}

func printAppScreenshotUploadResultMarkdown(result *AppScreenshotUploadResult) error {
	fmt.Fprintln(os.Stdout, "| Localization ID | Set ID | Display Type |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
		escapeMarkdown(result.VersionLocalizationID),
		escapeMarkdown(result.SetID),
		escapeMarkdown(result.DisplayType),
	)
	if len(result.Results) == 0 {
		return nil
	}
	fmt.Fprintln(os.Stdout, "\n| File Name | Asset ID | State |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	for _, item := range result.Results {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
			escapeMarkdown(item.FileName),
			escapeMarkdown(item.AssetID),
			escapeMarkdown(item.State),
		)
	}
	return nil
}

func printAppPreviewUploadResultTable(result *AppPreviewUploadResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Localization ID\tSet ID\tPreview Type")
	fmt.Fprintf(w, "%s\t%s\t%s\n", result.VersionLocalizationID, result.SetID, result.PreviewType)
	if err := w.Flush(); err != nil {
		return err
	}
	if len(result.Results) == 0 {
		return nil
	}
	fmt.Fprintln(os.Stdout, "\nPreviews")
	items := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(items, "File Name\tAsset ID\tState")
	for _, item := range result.Results {
		fmt.Fprintf(items, "%s\t%s\t%s\n", item.FileName, item.AssetID, item.State)
	}
	return items.Flush()
}

func printAppPreviewUploadResultMarkdown(result *AppPreviewUploadResult) error {
	fmt.Fprintln(os.Stdout, "| Localization ID | Set ID | Preview Type |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
		escapeMarkdown(result.VersionLocalizationID),
		escapeMarkdown(result.SetID),
		escapeMarkdown(result.PreviewType),
	)
	if len(result.Results) == 0 {
		return nil
	}
	fmt.Fprintln(os.Stdout, "\n| File Name | Asset ID | State |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	for _, item := range result.Results {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
			escapeMarkdown(item.FileName),
			escapeMarkdown(item.AssetID),
			escapeMarkdown(item.State),
		)
	}
	return nil
}

func printAssetDeleteResultTable(result *AssetDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printAssetDeleteResultMarkdown(result *AssetDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n", escapeMarkdown(result.ID), result.Deleted)
	return nil
}
