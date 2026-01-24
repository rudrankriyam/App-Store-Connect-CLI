package asc

// AppScreenshotSetAttributes describes a screenshot set resource.
type AppScreenshotSetAttributes struct {
	ScreenshotDisplayType string `json:"screenshotDisplayType"`
}

// AppScreenshotAttributes describes a screenshot asset resource.
type AppScreenshotAttributes struct {
	FileSize           int64               `json:"fileSize"`
	FileName           string              `json:"fileName"`
	SourceFileChecksum *Checksum           `json:"sourceFileChecksum,omitempty"`
	ImageAsset         *ImageAsset         `json:"imageAsset,omitempty"`
	AssetToken         string              `json:"assetToken,omitempty"`
	AssetType          string              `json:"assetType,omitempty"`
	UploadOperations   []UploadOperation   `json:"uploadOperations,omitempty"`
	AssetDeliveryState *AssetDeliveryState `json:"assetDeliveryState,omitempty"`
}

// ImageAsset describes an image asset.
type ImageAsset struct {
	TemplateURL string `json:"templateUrl"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

// AssetDeliveryState describes the delivery state of an asset.
type AssetDeliveryState struct {
	State  string        `json:"state"`
	Errors []ErrorDetail `json:"errors,omitempty"`
}

// ErrorDetail describes an asset error detail.
type ErrorDetail struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// AppPreviewSetAttributes describes a preview set resource.
type AppPreviewSetAttributes struct {
	PreviewType string `json:"previewType"`
}

// AppPreviewAttributes describes a preview asset resource.
type AppPreviewAttributes struct {
	FileSize             int64               `json:"fileSize"`
	FileName             string              `json:"fileName"`
	SourceFileChecksum   *Checksum           `json:"sourceFileChecksum,omitempty"`
	PreviewFrameTimeCode string              `json:"previewFrameTimeCode,omitempty"`
	MimeType             string              `json:"mimeType,omitempty"`
	VideoURL             string              `json:"videoUrl,omitempty"`
	PreviewImage         *ImageAsset         `json:"previewImage,omitempty"`
	UploadOperations     []UploadOperation   `json:"uploadOperations,omitempty"`
	AssetDeliveryState   *AssetDeliveryState `json:"assetDeliveryState,omitempty"`
}

// Response types
type AppScreenshotSetsResponse = Response[AppScreenshotSetAttributes]
type AppScreenshotSetResponse = SingleResponse[AppScreenshotSetAttributes]
type AppScreenshotsResponse = Response[AppScreenshotAttributes]
type AppScreenshotResponse = SingleResponse[AppScreenshotAttributes]
type AppPreviewSetsResponse = Response[AppPreviewSetAttributes]
type AppPreviewSetResponse = SingleResponse[AppPreviewSetAttributes]
type AppPreviewsResponse = Response[AppPreviewAttributes]
type AppPreviewResponse = SingleResponse[AppPreviewAttributes]

// Valid screenshot display types for validation.
var ValidScreenshotDisplayTypes = []string{
	"APP_IPHONE_67",
	"APP_IPHONE_65",
	"APP_IPHONE_61",
	"APP_IPHONE_58",
	"APP_IPHONE_55",
	"APP_IPAD_PRO_129",
	"APP_IPAD_PRO_3RD_GEN_129",
	"APP_IPAD_PRO_11",
	"APP_APPLE_TV",
	"APP_APPLE_WATCH_ULTRA",
	"APP_APPLE_WATCH_SERIES_10",
}

// Valid preview types for validation.
var ValidPreviewTypes = []string{
	"IPHONE_67",
	"IPHONE_65",
	"IPHONE_61",
	"IPHONE_58",
	"IPHONE_55",
	"IPAD_PRO_129",
	"IPAD_PRO_3RD_GEN_129",
	"IPAD_PRO_11",
	"APPLE_TV",
	"APPLE_WATCH_ULTRA",
	"APPLE_WATCH_SERIES_10",
}

// IsValidScreenshotDisplayType checks if a screenshot display type is supported.
func IsValidScreenshotDisplayType(value string) bool {
	for _, item := range ValidScreenshotDisplayTypes {
		if item == value {
			return true
		}
	}
	return false
}

// IsValidPreviewType checks if a preview type is supported.
func IsValidPreviewType(value string) bool {
	for _, item := range ValidPreviewTypes {
		if item == value {
			return true
		}
	}
	return false
}
