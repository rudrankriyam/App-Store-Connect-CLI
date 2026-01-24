package asc

// ResourceType represents an ASC resource type.
type ResourceType string

const (
	ResourceTypeApps                         ResourceType = "apps"
	ResourceTypeBuilds                       ResourceType = "builds"
	ResourceTypeBuildUploads                 ResourceType = "buildUploads"
	ResourceTypeBuildUploadFiles             ResourceType = "buildUploadFiles"
	ResourceTypeAppStoreVersions             ResourceType = "appStoreVersions"
	ResourceTypePreReleaseVersions           ResourceType = "preReleaseVersions"
	ResourceTypeAppStoreVersionSubmissions   ResourceType = "appStoreVersionSubmissions"
	ResourceTypeAppScreenshotSets            ResourceType = "appScreenshotSets"
	ResourceTypeAppScreenshots               ResourceType = "appScreenshots"
	ResourceTypeAppPreviewSets               ResourceType = "appPreviewSets"
	ResourceTypeAppPreviews                  ResourceType = "appPreviews"
	ResourceTypeBetaGroups                   ResourceType = "betaGroups"
	ResourceTypeBetaTesters                  ResourceType = "betaTesters"
	ResourceTypeBetaTesterInvitations        ResourceType = "betaTesterInvitations"
	ResourceTypeSandboxTesters               ResourceType = "sandboxTesters"
	ResourceTypeSandboxTestersClearHistory   ResourceType = "sandboxTestersClearPurchaseHistoryRequest"
	ResourceTypeAppStoreVersionLocalizations ResourceType = "appStoreVersionLocalizations"
	ResourceTypeAppInfoLocalizations         ResourceType = "appInfoLocalizations"
	ResourceTypeAppInfos                     ResourceType = "appInfos"
	ResourceTypeAnalyticsReportRequests      ResourceType = "analyticsReportRequests"
	ResourceTypeAnalyticsReports             ResourceType = "analyticsReports"
	ResourceTypeAnalyticsReportInstances     ResourceType = "analyticsReportInstances"
	ResourceTypeAnalyticsReportSegments      ResourceType = "analyticsReportSegments"
)

// Resource is a generic ASC API resource wrapper.
type Resource[T any] struct {
	Type       ResourceType `json:"type"`
	ID         string       `json:"id"`
	Attributes T            `json:"attributes"`
}

// Response is a generic ASC API response wrapper.
type Response[T any] struct {
	Data  []Resource[T] `json:"data"`
	Links Links         `json:"links,omitempty"`
}

// SingleResponse is a generic ASC API response wrapper for single resources.
type SingleResponse[T any] struct {
	Data  Resource[T] `json:"data"`
	Links Links       `json:"links,omitempty"`
}

// SingleResourceResponse is a response with a single resource (not an array).
type SingleResourceResponse[T any] struct {
	Data Resource[T] `json:"data"`
}

// Links represents pagination links
type Links struct {
	Self string `json:"self,omitempty"`
	Next string `json:"next,omitempty"`
	Prev string `json:"prev,omitempty"`
}

// Platform represents an Apple platform.
type Platform string

const (
	PlatformIOS      Platform = "IOS"
	PlatformMacOS    Platform = "MAC_OS"
	PlatformTVOS     Platform = "TV_OS"
	PlatformVisionOS Platform = "VISION_OS"
)

// ChecksumAlgorithm represents the algorithm used for checksums.
type ChecksumAlgorithm string

const (
	ChecksumAlgorithmMD5    ChecksumAlgorithm = "MD5"
	ChecksumAlgorithmSHA256 ChecksumAlgorithm = "SHA_256"
)

// AssetType represents the asset type for build uploads.
type AssetType string

const (
	AssetTypeAsset AssetType = "ASSET"
)

// UTI represents a Uniform Type Identifier used in uploads.
type UTI string

const (
	UTIIPA UTI = "com.apple.ipa"
)

// Relationship represents a generic API relationship.
type Relationship struct {
	Data ResourceData `json:"data"`
}

// RelationshipList represents a relationship containing multiple resources.
type RelationshipList struct {
	Data []ResourceData `json:"data"`
}

// RelationshipRequest represents a relationship list payload.
type RelationshipRequest struct {
	Data []RelationshipData `json:"data"`
}

// RelationshipData represents data in a relationship payload.
type RelationshipData struct {
	Type ResourceType `json:"type"`
	ID   string       `json:"id"`
}

// ResourceData represents the data portion of a resource.
type ResourceData struct {
	Type ResourceType `json:"type"`
	ID   string       `json:"id"`
}
