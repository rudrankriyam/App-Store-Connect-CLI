package asc

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/auth"
)

const (
	// BaseURL is the App Store Connect API base URL
	BaseURL = "https://api.appstoreconnect.apple.com"
	// DefaultTimeout is the default request timeout
	DefaultTimeout = 30 * time.Second
	tokenLifetime  = 20 * time.Minute
)

// ResolveTimeout returns the request timeout, optionally overridden by env vars.
func ResolveTimeout() time.Duration {
	timeout := DefaultTimeout
	if override := strings.TrimSpace(os.Getenv("ASC_TIMEOUT")); override != "" {
		if parsed, err := time.ParseDuration(override); err == nil && parsed > 0 {
			timeout = parsed
		}
	} else if override := strings.TrimSpace(os.Getenv("ASC_TIMEOUT_SECONDS")); override != "" {
		if parsed, err := time.ParseDuration(override + "s"); err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	return timeout
}

// Client is an App Store Connect API client
type Client struct {
	httpClient *http.Client
	keyID      string
	issuerID   string
	privateKey *ecdsa.PrivateKey
}

// ResourceType represents an ASC resource type.
type ResourceType string

const (
	ResourceTypeApps                         ResourceType = "apps"
	ResourceTypeBuilds                       ResourceType = "builds"
	ResourceTypeBuildUploads                 ResourceType = "buildUploads"
	ResourceTypeBuildUploadFiles             ResourceType = "buildUploadFiles"
	ResourceTypeAppStoreVersions             ResourceType = "appStoreVersions"
	ResourceTypeAppStoreVersionSubmissions   ResourceType = "appStoreVersionSubmissions"
	ResourceTypeBetaGroups                   ResourceType = "betaGroups"
	ResourceTypeBetaTesters                  ResourceType = "betaTesters"
	ResourceTypeBetaTesterInvitations        ResourceType = "betaTesterInvitations"
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

// FeedbackAttributes describes beta feedback screenshot submissions.
type FeedbackAttributes struct {
	CreatedDate    string                    `json:"createdDate"`
	Comment        string                    `json:"comment"`
	Email          string                    `json:"email"`
	DeviceModel    string                    `json:"deviceModel,omitempty"`
	OSVersion      string                    `json:"osVersion,omitempty"`
	AppPlatform    string                    `json:"appPlatform,omitempty"`
	DevicePlatform string                    `json:"devicePlatform,omitempty"`
	Screenshots    []FeedbackScreenshotImage `json:"screenshots,omitempty"`
}

// FeedbackScreenshotImage describes a screenshot attached to feedback.
type FeedbackScreenshotImage struct {
	URL            string `json:"url"`
	Width          int    `json:"width,omitempty"`
	Height         int    `json:"height,omitempty"`
	ExpirationDate string `json:"expirationDate,omitempty"`
}

// CrashAttributes describes beta feedback crash submissions.
type CrashAttributes struct {
	CreatedDate    string `json:"createdDate"`
	Comment        string `json:"comment"`
	Email          string `json:"email"`
	DeviceModel    string `json:"deviceModel,omitempty"`
	OSVersion      string `json:"osVersion,omitempty"`
	AppPlatform    string `json:"appPlatform,omitempty"`
	DevicePlatform string `json:"devicePlatform,omitempty"`
	CrashLog       string `json:"crashLog,omitempty"`
}

// ReviewAttributes describes App Store customer reviews.
type ReviewAttributes struct {
	Rating           int    `json:"rating"`
	Title            string `json:"title"`
	Body             string `json:"body"`
	ReviewerNickname string `json:"reviewerNickname"`
	CreatedDate      string `json:"createdDate"`
	Territory        string `json:"territory"`
}

// FeedbackResponse is the response from beta feedback screenshots endpoint.
type FeedbackResponse = Response[FeedbackAttributes]

// CrashesResponse is the response from beta feedback crashes endpoint.
type CrashesResponse = Response[CrashAttributes]

// ReviewsResponse is the response from customer reviews endpoint.
type ReviewsResponse = Response[ReviewAttributes]

// AppsResponse is the response from apps endpoint.
type AppsResponse = Response[AppAttributes]

// BuildsResponse is the response from builds endpoint.
type BuildsResponse = Response[BuildAttributes]

// AppStoreVersionsResponse is the response from app store versions endpoints.
type AppStoreVersionsResponse = Response[AppStoreVersionAttributes]

// AppStoreVersionResponse is the response from app store version detail.
type AppStoreVersionResponse = SingleResponse[AppStoreVersionAttributes]

// BuildResponse is the response from build detail/updates.
type BuildResponse = SingleResponse[BuildAttributes]

// AppStoreVersionLocalizationsResponse is the response from app store version localizations endpoints.
type AppStoreVersionLocalizationsResponse = Response[AppStoreVersionLocalizationAttributes]

// AppStoreVersionLocalizationResponse is the response from app store version localization detail/creates.
type AppStoreVersionLocalizationResponse = SingleResponse[AppStoreVersionLocalizationAttributes]

// AppInfoLocalizationsResponse is the response from app info localizations endpoints.
type AppInfoLocalizationsResponse = Response[AppInfoLocalizationAttributes]

// AppInfoLocalizationResponse is the response from app info localization detail/creates.
type AppInfoLocalizationResponse = SingleResponse[AppInfoLocalizationAttributes]

// AppInfosResponse is the response from app info endpoints.
type AppInfosResponse = Response[AppInfoAttributes]

// BetaGroupsResponse is the response from beta groups endpoints.
type BetaGroupsResponse = Response[BetaGroupAttributes]

// BetaGroupResponse is the response from beta group detail/creates.
type BetaGroupResponse = SingleResponse[BetaGroupAttributes]

// BetaTestersResponse is the response from beta testers endpoints.
type BetaTestersResponse = Response[BetaTesterAttributes]

// BetaTesterResponse is the response from beta tester detail/creates.
type BetaTesterResponse = SingleResponse[BetaTesterAttributes]

// BetaTesterInvitationResponse is the response from beta tester invitations.
type BetaTesterInvitationResponse = SingleResponse[struct{}]

type listQuery struct {
	limit   int
	nextURL string
}

type feedbackQuery struct {
	listQuery
	deviceModels              []string
	osVersions                []string
	appPlatforms              []string
	devicePlatforms           []string
	buildIDs                  []string
	buildPreReleaseVersionIDs []string
	testerIDs                 []string
	sort                      string
	includeScreenshots        bool
}

type crashQuery struct {
	listQuery
	deviceModels              []string
	osVersions                []string
	appPlatforms              []string
	devicePlatforms           []string
	buildIDs                  []string
	buildPreReleaseVersionIDs []string
	testerIDs                 []string
	sort                      string
}

type reviewQuery struct {
	listQuery
	rating    int
	territory string
	sort      string
}

type appsQuery struct {
	listQuery
	sort string
}

type buildsQuery struct {
	listQuery
	sort string
}

type appStoreVersionsQuery struct {
	listQuery
	platforms      []string
	versionStrings []string
	states         []string
}

type appStoreVersionLocalizationsQuery struct {
	listQuery
	locales []string
}

type appInfoLocalizationsQuery struct {
	listQuery
	locales []string
}

type betaGroupsQuery struct {
	listQuery
}

type betaTestersQuery struct {
	listQuery
	email    string
	groupIDs []string
}

// AppAttributes describes an app resource.
type AppAttributes struct {
	Name          string `json:"name"`
	BundleID      string `json:"bundleId"`
	SKU           string `json:"sku"`
	PrimaryLocale string `json:"primaryLocale,omitempty"`
}

// BuildAttributes describes a build resource.
type BuildAttributes struct {
	Version                 string `json:"version"`
	UploadedDate            string `json:"uploadedDate"`
	ExpirationDate          string `json:"expirationDate,omitempty"`
	ProcessingState         string `json:"processingState,omitempty"`
	MinOSVersion            string `json:"minOsVersion,omitempty"`
	UsesNonExemptEncryption bool   `json:"usesNonExemptEncryption,omitempty"`
	Expired                 bool   `json:"expired,omitempty"`
}

// AppStoreVersionAttributes describes app store version metadata.
type AppStoreVersionAttributes struct {
	Platform        Platform `json:"platform,omitempty"`
	VersionString   string   `json:"versionString,omitempty"`
	AppStoreState   string   `json:"appStoreState,omitempty"`
	AppVersionState string   `json:"appVersionState,omitempty"`
	CreatedDate     string   `json:"createdDate,omitempty"`
}

// AppStoreVersionLocalizationAttributes describes app store version localization metadata.
type AppStoreVersionLocalizationAttributes struct {
	Locale          string `json:"locale,omitempty"`
	Description     string `json:"description,omitempty"`
	Keywords        string `json:"keywords,omitempty"`
	MarketingURL    string `json:"marketingUrl,omitempty"`
	PromotionalText string `json:"promotionalText,omitempty"`
	SupportURL      string `json:"supportUrl,omitempty"`
	WhatsNew        string `json:"whatsNew,omitempty"`
}

// AppInfoLocalizationAttributes describes app info localization metadata.
type AppInfoLocalizationAttributes struct {
	Locale            string `json:"locale,omitempty"`
	Name              string `json:"name,omitempty"`
	Subtitle          string `json:"subtitle,omitempty"`
	PrivacyPolicyURL  string `json:"privacyPolicyUrl,omitempty"`
	PrivacyChoicesURL string `json:"privacyChoicesUrl,omitempty"`
	PrivacyPolicyText string `json:"privacyPolicyText,omitempty"`
}

// AppInfoAttributes describes app info resources.
type AppInfoAttributes struct{}

// BetaGroupAttributes describes a beta group resource.
type BetaGroupAttributes struct {
	Name                   string `json:"name"`
	CreatedDate            string `json:"createdDate,omitempty"`
	IsInternalGroup        bool   `json:"isInternalGroup,omitempty"`
	HasAccessToAllBuilds   bool   `json:"hasAccessToAllBuilds,omitempty"`
	PublicLinkEnabled      bool   `json:"publicLinkEnabled,omitempty"`
	PublicLinkLimitEnabled bool   `json:"publicLinkLimitEnabled,omitempty"`
	PublicLinkLimit        int    `json:"publicLinkLimit,omitempty"`
	PublicLink             string `json:"publicLink,omitempty"`
	FeedbackEnabled        bool   `json:"feedbackEnabled,omitempty"`
}

// BetaTesterAttributes describes a beta tester resource.
type BetaTesterAttributes struct {
	FirstName  string          `json:"firstName,omitempty"`
	LastName   string          `json:"lastName,omitempty"`
	Email      string          `json:"email,omitempty"`
	InviteType BetaInviteType  `json:"inviteType,omitempty"`
	State      BetaTesterState `json:"state,omitempty"`
}

// BetaInviteType represents the invitation type for a beta tester.
type BetaInviteType string

const (
	BetaInviteTypeEmail      BetaInviteType = "EMAIL"
	BetaInviteTypePublicLink BetaInviteType = "PUBLIC_LINK"
)

// BetaTesterState represents the invitation state for a beta tester.
type BetaTesterState string

const (
	BetaTesterStateNotInvited BetaTesterState = "NOT_INVITED"
	BetaTesterStateInvited    BetaTesterState = "INVITED"
	BetaTesterStateAccepted   BetaTesterState = "ACCEPTED"
	BetaTesterStateInstalled  BetaTesterState = "INSTALLED"
	BetaTesterStateRevoked    BetaTesterState = "REVOKED"
)

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

// ResourceData represents the data portion of a resource.
type ResourceData struct {
	Type ResourceType `json:"type"`
	ID   string       `json:"id"`
}

// BuildUploadAttributes describes a build upload resource.
type BuildUploadAttributes struct {
	CFBundleShortVersionString string   `json:"cfBundleShortVersionString"`
	CFBundleVersion            string   `json:"cfBundleVersion"`
	Platform                   Platform `json:"platform"`
	CreatedDate                *string  `json:"createdDate,omitempty"`
	State                      *string  `json:"state,omitempty"`
}

// BuildUploadRelationships describes the relationships for a build upload.
type BuildUploadRelationships struct {
	App   *Relationship `json:"app,omitempty"`
	Build *Relationship `json:"build,omitempty"`
}

// BuildUploadCreateData is the data portion of a build upload create request.
type BuildUploadCreateData struct {
	Type          ResourceType              `json:"type"`
	Attributes    BuildUploadAttributes     `json:"attributes"`
	Relationships *BuildUploadRelationships `json:"relationships,omitempty"`
}

// BuildUploadCreateRequest is a request to create a build upload.
type BuildUploadCreateRequest struct {
	Data BuildUploadCreateData `json:"data"`
}

// SingleResourceResponse is a response with a single resource (not an array).
type SingleResourceResponse[T any] struct {
	Data Resource[T] `json:"data"`
}

// BuildUploadResponse is the response from build upload endpoint.
type BuildUploadResponse = SingleResourceResponse[BuildUploadAttributes]

// BuildUploadFileAttributes describes a build upload file resource.
type BuildUploadFileAttributes struct {
	AssetDeliveryState  *AppMediaAssetState `json:"assetDeliveryState,omitempty"`
	AssetToken          *string             `json:"assetToken,omitempty"`
	AssetType           AssetType           `json:"assetType,omitempty"`
	FileName            string              `json:"fileName"`
	FileSize            int64               `json:"fileSize"`
	SourceFileChecksums *Checksums          `json:"sourceFileChecksums,omitempty"`
	UploadOperations    []UploadOperation   `json:"uploadOperations,omitempty"`
	UTI                 UTI                 `json:"uti"`
	Uploaded            *bool               `json:"uploaded,omitempty"`
}

// BuildUploadFileRelationships describes the relationships for a build upload file.
type BuildUploadFileRelationships struct {
	BuildUpload *Relationship `json:"buildUpload"`
}

// BuildUploadFileCreateData is the data portion of a build upload file create request.
type BuildUploadFileCreateData struct {
	Type          ResourceType                  `json:"type"`
	Attributes    BuildUploadFileAttributes     `json:"attributes"`
	Relationships *BuildUploadFileRelationships `json:"relationships"`
}

// BuildUploadFileCreateRequest is a request to create a build upload file.
type BuildUploadFileCreateRequest struct {
	Data BuildUploadFileCreateData `json:"data"`
}

// BuildUploadFileResponse is the response from build upload file endpoint.
type BuildUploadFileResponse = SingleResourceResponse[BuildUploadFileAttributes]

// BuildUploadFileUpdateAttributes describes the attributes to update on a build upload file.
type BuildUploadFileUpdateAttributes struct {
	SourceFileChecksums *Checksums `json:"sourceFileChecksums,omitempty"`
	Uploaded            *bool      `json:"uploaded,omitempty"`
}

// BuildUploadFileUpdateData is the data portion of a build upload file update request.
type BuildUploadFileUpdateData struct {
	Type       ResourceType                     `json:"type"`
	ID         string                           `json:"id"`
	Attributes *BuildUploadFileUpdateAttributes `json:"attributes,omitempty"`
}

// BuildUploadFileUpdateRequest is a request to update a build upload file.
type BuildUploadFileUpdateRequest struct {
	Data BuildUploadFileUpdateData `json:"data"`
}

// UploadOperation represents a file upload operation with presigned URL.
type UploadOperation struct {
	Method         string       `json:"method"`
	URL            string       `json:"url"`
	Length         int64        `json:"length"`
	Offset         int64        `json:"offset"`
	RequestHeaders []HTTPHeader `json:"requestHeaders,omitempty"`
	Expiration     *string      `json:"expiration,omitempty"`
}

// HTTPHeader represents an HTTP header.
type HTTPHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Checksums represents file checksums.
type Checksums struct {
	File      *Checksum `json:"file,omitempty"`
	Composite *Checksum `json:"composite,omitempty"`
}

// Checksum represents a single checksum.
type Checksum struct {
	Hash      string            `json:"hash"`
	Algorithm ChecksumAlgorithm `json:"algorithm"`
}

// AppMediaAssetState represents the state of an asset.
type AppMediaAssetState struct {
	State    *string       `json:"state,omitempty"`
	Errors   []StateDetail `json:"errors,omitempty"`
	Warnings []StateDetail `json:"warnings,omitempty"`
	Infos    []StateDetail `json:"infos,omitempty"`
}

// StateDetail represents details about a state (errors, warnings, infos).
type StateDetail struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// AppStoreVersionSubmissionCreateData is the data portion of an app store version submission create request.
type AppStoreVersionSubmissionCreateData struct {
	Type          ResourceType                            `json:"type"`
	Relationships *AppStoreVersionSubmissionRelationships `json:"relationships"`
}

// AppStoreVersionSubmissionCreateRequest is a request to create an app store version submission.
type AppStoreVersionSubmissionCreateRequest struct {
	Data AppStoreVersionSubmissionCreateData `json:"data"`
}

// AppStoreVersionSubmissionRelationships describes the relationships for an app store version submission.
type AppStoreVersionSubmissionRelationships struct {
	AppStoreVersion *Relationship `json:"appStoreVersion"`
}

// AppStoreVersionSubmissionAttributes describes an app store version submission resource.
type AppStoreVersionSubmissionAttributes struct {
	CreatedDate *string `json:"createdDate,omitempty"`
}

// AppStoreVersionSubmissionResponse is the response from app store version submission endpoint.
type AppStoreVersionSubmissionResponse = SingleResourceResponse[AppStoreVersionSubmissionAttributes]

// AppStoreVersionSubmissionResource represents a submission with relationships.
type AppStoreVersionSubmissionResource struct {
	Type       ResourceType `json:"type"`
	ID         string       `json:"id"`
	Attributes struct {
		CreatedDate *string `json:"createdDate,omitempty"`
	} `json:"attributes,omitempty"`
	Relationships struct {
		AppStoreVersion *Relationship `json:"appStoreVersion,omitempty"`
	} `json:"relationships,omitempty"`
}

// AppStoreVersionSubmissionResourceResponse is a response containing a submission resource.
type AppStoreVersionSubmissionResourceResponse struct {
	Data AppStoreVersionSubmissionResource `json:"data"`
}

// AppStoreVersionBuildRelationshipUpdateRequest is a request to attach a build to a version.
type AppStoreVersionBuildRelationshipUpdateRequest struct {
	Data ResourceData `json:"data"`
}

// AppStoreVersionLocalizationCreateData is the data portion of a version localization create request.
type AppStoreVersionLocalizationCreateData struct {
	Type          ResourceType                              `json:"type"`
	Attributes    AppStoreVersionLocalizationAttributes     `json:"attributes"`
	Relationships *AppStoreVersionLocalizationRelationships `json:"relationships"`
}

// AppStoreVersionLocalizationCreateRequest is a request to create a version localization.
type AppStoreVersionLocalizationCreateRequest struct {
	Data AppStoreVersionLocalizationCreateData `json:"data"`
}

// AppStoreVersionLocalizationUpdateData is the data portion of a version localization update request.
type AppStoreVersionLocalizationUpdateData struct {
	Type       ResourceType                          `json:"type"`
	ID         string                                `json:"id"`
	Attributes AppStoreVersionLocalizationAttributes `json:"attributes"`
}

// AppStoreVersionLocalizationUpdateRequest is a request to update a version localization.
type AppStoreVersionLocalizationUpdateRequest struct {
	Data AppStoreVersionLocalizationUpdateData `json:"data"`
}

// AppStoreVersionLocalizationRelationships describes relationships for version localizations.
type AppStoreVersionLocalizationRelationships struct {
	AppStoreVersion *Relationship `json:"appStoreVersion"`
}

// AppInfoLocalizationCreateData is the data portion of an app info localization create request.
type AppInfoLocalizationCreateData struct {
	Type          ResourceType                      `json:"type"`
	Attributes    AppInfoLocalizationAttributes     `json:"attributes"`
	Relationships *AppInfoLocalizationRelationships `json:"relationships"`
}

// AppInfoLocalizationCreateRequest is a request to create an app info localization.
type AppInfoLocalizationCreateRequest struct {
	Data AppInfoLocalizationCreateData `json:"data"`
}

// AppInfoLocalizationUpdateData is the data portion of an app info localization update request.
type AppInfoLocalizationUpdateData struct {
	Type       ResourceType                  `json:"type"`
	ID         string                        `json:"id"`
	Attributes AppInfoLocalizationAttributes `json:"attributes"`
}

// AppInfoLocalizationUpdateRequest is a request to update an app info localization.
type AppInfoLocalizationUpdateRequest struct {
	Data AppInfoLocalizationUpdateData `json:"data"`
}

// AppInfoLocalizationRelationships describes relationships for app info localizations.
type AppInfoLocalizationRelationships struct {
	AppInfo *Relationship `json:"appInfo"`
}

// BetaGroupCreateData is the data portion of a beta group create request.
type BetaGroupCreateData struct {
	Type          ResourceType            `json:"type"`
	Attributes    BetaGroupAttributes     `json:"attributes"`
	Relationships *BetaGroupRelationships `json:"relationships"`
}

// BetaGroupCreateRequest is a request to create a beta group.
type BetaGroupCreateRequest struct {
	Data BetaGroupCreateData `json:"data"`
}

// BetaGroupRelationships describes relationships for beta groups.
type BetaGroupRelationships struct {
	App *Relationship `json:"app"`
}

// BetaTesterCreateAttributes describes attributes for creating a beta tester.
type BetaTesterCreateAttributes struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email"`
}

// BetaTesterCreateRelationships describes relationships for beta tester creation.
type BetaTesterCreateRelationships struct {
	BetaGroups *RelationshipList `json:"betaGroups,omitempty"`
}

// BetaTesterCreateData is the data portion of a beta tester create request.
type BetaTesterCreateData struct {
	Type          ResourceType                   `json:"type"`
	Attributes    BetaTesterCreateAttributes     `json:"attributes"`
	Relationships *BetaTesterCreateRelationships `json:"relationships,omitempty"`
}

// BetaTesterCreateRequest is a request to create a beta tester.
type BetaTesterCreateRequest struct {
	Data BetaTesterCreateData `json:"data"`
}

// BetaTesterInvitationCreateRelationships describes relationships for invitations.
type BetaTesterInvitationCreateRelationships struct {
	App        *Relationship `json:"app"`
	BetaTester *Relationship `json:"betaTester,omitempty"`
}

// BetaTesterInvitationCreateData is the data portion of an invitation create request.
type BetaTesterInvitationCreateData struct {
	Type          ResourceType                             `json:"type"`
	Relationships *BetaTesterInvitationCreateRelationships `json:"relationships"`
}

// BetaTesterInvitationCreateRequest is a request to create a beta tester invitation.
type BetaTesterInvitationCreateRequest struct {
	Data BetaTesterInvitationCreateData `json:"data"`
}

// BuildUploadResult represents CLI output for build upload preparation.
type BuildUploadResult struct {
	UploadID   string            `json:"uploadId"`
	FileID     string            `json:"fileId"`
	FileName   string            `json:"fileName"`
	FileSize   int64             `json:"fileSize"`
	Operations []UploadOperation `json:"operations,omitempty"`
}

// AppStoreVersionSubmissionResult represents CLI output for submissions.
type AppStoreVersionSubmissionResult struct {
	SubmissionID string  `json:"submissionId"`
	CreatedDate  *string `json:"createdDate,omitempty"`
}

// AppStoreVersionSubmissionCreateResult represents CLI output for submission creation.
type AppStoreVersionSubmissionCreateResult struct {
	SubmissionID string  `json:"submissionId"`
	VersionID    string  `json:"versionId"`
	BuildID      string  `json:"buildId"`
	CreatedDate  *string `json:"createdDate,omitempty"`
}

// AppStoreVersionSubmissionStatusResult represents CLI output for submission status.
type AppStoreVersionSubmissionStatusResult struct {
	ID            string  `json:"id"`
	VersionID     string  `json:"versionId,omitempty"`
	VersionString string  `json:"versionString,omitempty"`
	Platform      string  `json:"platform,omitempty"`
	State         string  `json:"state,omitempty"`
	CreatedDate   *string `json:"createdDate,omitempty"`
}

// AppStoreVersionSubmissionCancelResult represents CLI output for submission cancellation.
type AppStoreVersionSubmissionCancelResult struct {
	ID        string `json:"id"`
	Cancelled bool   `json:"cancelled"`
}

// AppStoreVersionDetailResult represents CLI output for version details.
type AppStoreVersionDetailResult struct {
	ID            string `json:"id"`
	VersionString string `json:"versionString,omitempty"`
	Platform      string `json:"platform,omitempty"`
	State         string `json:"state,omitempty"`
	BuildID       string `json:"buildId,omitempty"`
	BuildVersion  string `json:"buildVersion,omitempty"`
	SubmissionID  string `json:"submissionId,omitempty"`
}

// AppStoreVersionAttachBuildResult represents CLI output for build attachment.
type AppStoreVersionAttachBuildResult struct {
	VersionID string `json:"versionId"`
	BuildID   string `json:"buildId"`
	Attached  bool   `json:"attached"`
}

// BetaTesterInvitationResult represents CLI output for invitations.
type BetaTesterInvitationResult struct {
	InvitationID string `json:"invitationId"`
	TesterID     string `json:"testerId,omitempty"`
	AppID        string `json:"appId,omitempty"`
	Email        string `json:"email,omitempty"`
}

// BetaTesterDeleteResult represents CLI output for deletions.
type BetaTesterDeleteResult struct {
	ID      string `json:"id"`
	Email   string `json:"email,omitempty"`
	Deleted bool   `json:"deleted"`
}

// LocalizationFileResult represents a localization file written or read.
type LocalizationFileResult struct {
	Locale string `json:"locale"`
	Path   string `json:"path"`
}

// LocalizationDownloadResult represents CLI output for localization downloads.
type LocalizationDownloadResult struct {
	Type       string                   `json:"type"`
	VersionID  string                   `json:"versionId,omitempty"`
	AppID      string                   `json:"appId,omitempty"`
	AppInfoID  string                   `json:"appInfoId,omitempty"`
	OutputPath string                   `json:"outputPath"`
	Files      []LocalizationFileResult `json:"files"`
}

// LocalizationUploadLocaleResult represents a per-locale upload result.
type LocalizationUploadLocaleResult struct {
	Locale         string `json:"locale"`
	Action         string `json:"action"`
	LocalizationID string `json:"localizationId,omitempty"`
}

// LocalizationUploadResult represents CLI output for localization uploads.
type LocalizationUploadResult struct {
	Type      string                           `json:"type"`
	VersionID string                           `json:"versionId,omitempty"`
	AppID     string                           `json:"appId,omitempty"`
	AppInfoID string                           `json:"appInfoId,omitempty"`
	DryRun    bool                             `json:"dryRun"`
	Results   []LocalizationUploadLocaleResult `json:"results"`
}

// FeedbackOption is a functional option for GetFeedback.
type FeedbackOption func(*feedbackQuery)

// CrashOption is a functional option for GetCrashes.
type CrashOption func(*crashQuery)

// ReviewOption is a functional option for GetReviews.
type ReviewOption func(*reviewQuery)

// AppsOption is a functional option for GetApps.
type AppsOption func(*appsQuery)

// BuildsOption is a functional option for GetBuilds.
type BuildsOption func(*buildsQuery)

// AppStoreVersionsOption is a functional option for GetAppStoreVersions.
type AppStoreVersionsOption func(*appStoreVersionsQuery)

// BetaGroupsOption is a functional option for GetBetaGroups.
type BetaGroupsOption func(*betaGroupsQuery)

// BetaTestersOption is a functional option for GetBetaTesters.
type BetaTestersOption func(*betaTestersQuery)

// AppStoreVersionLocalizationsOption is a functional option for version localizations.
type AppStoreVersionLocalizationsOption func(*appStoreVersionLocalizationsQuery)

// AppInfoLocalizationsOption is a functional option for app info localizations.
type AppInfoLocalizationsOption func(*appInfoLocalizationsQuery)

// WithFeedbackDeviceModels filters feedback by device model(s).
func WithFeedbackDeviceModels(models []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.deviceModels = normalizeList(models)
	}
}

// WithFeedbackOSVersions filters feedback by OS version(s).
func WithFeedbackOSVersions(versions []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.osVersions = normalizeList(versions)
	}
}

// WithFeedbackAppPlatforms filters feedback by app platform(s).
func WithFeedbackAppPlatforms(platforms []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.appPlatforms = normalizeUpperList(platforms)
	}
}

// WithFeedbackDevicePlatforms filters feedback by device platform(s).
func WithFeedbackDevicePlatforms(platforms []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.devicePlatforms = normalizeUpperList(platforms)
	}
}

// WithFeedbackBuildIDs filters feedback by build ID(s).
func WithFeedbackBuildIDs(ids []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.buildIDs = normalizeList(ids)
	}
}

// WithFeedbackBuildPreReleaseVersionIDs filters feedback by pre-release version ID(s).
func WithFeedbackBuildPreReleaseVersionIDs(ids []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.buildPreReleaseVersionIDs = normalizeList(ids)
	}
}

// WithFeedbackTesterIDs filters feedback by tester ID(s).
func WithFeedbackTesterIDs(ids []string) FeedbackOption {
	return func(q *feedbackQuery) {
		q.testerIDs = normalizeList(ids)
	}
}

// WithFeedbackLimit sets the max number of feedback items to return.
func WithFeedbackLimit(limit int) FeedbackOption {
	return func(q *feedbackQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithFeedbackNextURL uses a next page URL directly.
func WithFeedbackNextURL(next string) FeedbackOption {
	return func(q *feedbackQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithFeedbackSort sets the sort order for feedback.
func WithFeedbackSort(sort string) FeedbackOption {
	return func(q *feedbackQuery) {
		if strings.TrimSpace(sort) != "" {
			q.sort = strings.TrimSpace(sort)
		}
	}
}

// WithFeedbackIncludeScreenshots includes screenshot URLs in feedback responses.
func WithFeedbackIncludeScreenshots() FeedbackOption {
	return func(q *feedbackQuery) {
		q.includeScreenshots = true
	}
}

// WithCrashDeviceModels filters crashes by device model(s).
func WithCrashDeviceModels(models []string) CrashOption {
	return func(q *crashQuery) {
		q.deviceModels = normalizeList(models)
	}
}

// WithCrashOSVersions filters crashes by OS version(s).
func WithCrashOSVersions(versions []string) CrashOption {
	return func(q *crashQuery) {
		q.osVersions = normalizeList(versions)
	}
}

// WithCrashAppPlatforms filters crashes by app platform(s).
func WithCrashAppPlatforms(platforms []string) CrashOption {
	return func(q *crashQuery) {
		q.appPlatforms = normalizeUpperList(platforms)
	}
}

// WithCrashDevicePlatforms filters crashes by device platform(s).
func WithCrashDevicePlatforms(platforms []string) CrashOption {
	return func(q *crashQuery) {
		q.devicePlatforms = normalizeUpperList(platforms)
	}
}

// WithCrashBuildIDs filters crashes by build ID(s).
func WithCrashBuildIDs(ids []string) CrashOption {
	return func(q *crashQuery) {
		q.buildIDs = normalizeList(ids)
	}
}

// WithCrashBuildPreReleaseVersionIDs filters crashes by pre-release version ID(s).
func WithCrashBuildPreReleaseVersionIDs(ids []string) CrashOption {
	return func(q *crashQuery) {
		q.buildPreReleaseVersionIDs = normalizeList(ids)
	}
}

// WithCrashTesterIDs filters crashes by tester ID(s).
func WithCrashTesterIDs(ids []string) CrashOption {
	return func(q *crashQuery) {
		q.testerIDs = normalizeList(ids)
	}
}

// WithCrashLimit sets the max number of crash items to return.
func WithCrashLimit(limit int) CrashOption {
	return func(q *crashQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithCrashNextURL uses a next page URL directly.
func WithCrashNextURL(next string) CrashOption {
	return func(q *crashQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithCrashSort sets the sort order for crashes.
func WithCrashSort(sort string) CrashOption {
	return func(q *crashQuery) {
		if strings.TrimSpace(sort) != "" {
			q.sort = strings.TrimSpace(sort)
		}
	}
}

// WithRating filters reviews by star rating (1-5).
func WithRating(rating int) ReviewOption {
	return func(r *reviewQuery) {
		if rating >= 1 && rating <= 5 {
			r.rating = rating
		}
	}
}

// WithTerritory filters reviews by territory code (e.g. US, GBR).
func WithTerritory(territory string) ReviewOption {
	return func(r *reviewQuery) {
		if territory != "" {
			r.territory = strings.ToUpper(territory)
		}
	}
}

// WithReviewSort sets the sort order for reviews.
func WithReviewSort(sort string) ReviewOption {
	return func(r *reviewQuery) {
		if strings.TrimSpace(sort) != "" {
			r.sort = strings.TrimSpace(sort)
		}
	}
}

// WithLimit sets the max number of reviews to return.
func WithLimit(limit int) ReviewOption {
	return func(r *reviewQuery) {
		if limit > 0 {
			r.limit = limit
		}
	}
}

// WithNextURL uses a next page URL directly.
func WithNextURL(next string) ReviewOption {
	return func(r *reviewQuery) {
		if strings.TrimSpace(next) != "" {
			r.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppsLimit sets the max number of apps to return.
func WithAppsLimit(limit int) AppsOption {
	return func(q *appsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithAppsNextURL uses a next page URL directly.
func WithAppsNextURL(next string) AppsOption {
	return func(q *appsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppsSort sets the sort order for apps.
func WithAppsSort(sort string) AppsOption {
	return func(q *appsQuery) {
		if strings.TrimSpace(sort) != "" {
			q.sort = strings.TrimSpace(sort)
		}
	}
}

// WithBuildsLimit sets the max number of builds to return.
func WithBuildsLimit(limit int) BuildsOption {
	return func(q *buildsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBuildsNextURL uses a next page URL directly.
func WithBuildsNextURL(next string) BuildsOption {
	return func(q *buildsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBuildsSort sets the sort order for builds.
func WithBuildsSort(sort string) BuildsOption {
	return func(q *buildsQuery) {
		if strings.TrimSpace(sort) != "" {
			q.sort = strings.TrimSpace(sort)
		}
	}
}

// WithAppStoreVersionsLimit sets the max number of versions to return.
func WithAppStoreVersionsLimit(limit int) AppStoreVersionsOption {
	return func(q *appStoreVersionsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithAppStoreVersionsNextURL uses a next page URL directly.
func WithAppStoreVersionsNextURL(next string) AppStoreVersionsOption {
	return func(q *appStoreVersionsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppStoreVersionsPlatforms filters versions by platform.
func WithAppStoreVersionsPlatforms(platforms []string) AppStoreVersionsOption {
	return func(q *appStoreVersionsQuery) {
		q.platforms = normalizeUpperList(platforms)
	}
}

// WithAppStoreVersionsVersionStrings filters versions by version string.
func WithAppStoreVersionsVersionStrings(versions []string) AppStoreVersionsOption {
	return func(q *appStoreVersionsQuery) {
		q.versionStrings = normalizeList(versions)
	}
}

// WithAppStoreVersionsStates filters versions by app store state.
func WithAppStoreVersionsStates(states []string) AppStoreVersionsOption {
	return func(q *appStoreVersionsQuery) {
		q.states = normalizeUpperList(states)
	}
}

// WithBetaGroupsLimit sets the max number of beta groups to return.
func WithBetaGroupsLimit(limit int) BetaGroupsOption {
	return func(q *betaGroupsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBetaGroupsNextURL uses a next page URL directly.
func WithBetaGroupsNextURL(next string) BetaGroupsOption {
	return func(q *betaGroupsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBetaTestersLimit sets the max number of beta testers to return.
func WithBetaTestersLimit(limit int) BetaTestersOption {
	return func(q *betaTestersQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithBetaTestersNextURL uses a next page URL directly.
func WithBetaTestersNextURL(next string) BetaTestersOption {
	return func(q *betaTestersQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithBetaTestersEmail filters beta testers by email.
func WithBetaTestersEmail(email string) BetaTestersOption {
	return func(q *betaTestersQuery) {
		q.email = strings.TrimSpace(email)
	}
}

// WithBetaTestersGroupIDs filters beta testers by beta group ID(s).
func WithBetaTestersGroupIDs(ids []string) BetaTestersOption {
	return func(q *betaTestersQuery) {
		q.groupIDs = normalizeList(ids)
	}
}

// WithAppStoreVersionLocalizationsLimit sets the max number of localizations to return.
func WithAppStoreVersionLocalizationsLimit(limit int) AppStoreVersionLocalizationsOption {
	return func(q *appStoreVersionLocalizationsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithAppStoreVersionLocalizationsNextURL uses a next page URL directly.
func WithAppStoreVersionLocalizationsNextURL(next string) AppStoreVersionLocalizationsOption {
	return func(q *appStoreVersionLocalizationsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppStoreVersionLocalizationLocales filters version localizations by locale.
func WithAppStoreVersionLocalizationLocales(locales []string) AppStoreVersionLocalizationsOption {
	return func(q *appStoreVersionLocalizationsQuery) {
		q.locales = normalizeList(locales)
	}
}

// WithAppInfoLocalizationsLimit sets the max number of app info localizations to return.
func WithAppInfoLocalizationsLimit(limit int) AppInfoLocalizationsOption {
	return func(q *appInfoLocalizationsQuery) {
		if limit > 0 {
			q.limit = limit
		}
	}
}

// WithAppInfoLocalizationsNextURL uses a next page URL directly.
func WithAppInfoLocalizationsNextURL(next string) AppInfoLocalizationsOption {
	return func(q *appInfoLocalizationsQuery) {
		if strings.TrimSpace(next) != "" {
			q.nextURL = strings.TrimSpace(next)
		}
	}
}

// WithAppInfoLocalizationLocales filters app info localizations by locale.
func WithAppInfoLocalizationLocales(locales []string) AppInfoLocalizationsOption {
	return func(q *appInfoLocalizationsQuery) {
		q.locales = normalizeList(locales)
	}
}

// NewClient creates a new ASC client
func NewClient(keyID, issuerID, privateKeyPath string) (*Client, error) {
	if err := auth.ValidateKeyFile(privateKeyPath); err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	key, err := auth.LoadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: ResolveTimeout(),
		},
		keyID:      keyID,
		issuerID:   issuerID,
		privateKey: key,
	}, nil
}

// newRequest creates a new HTTP request with JWT authentication
func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	// Generate JWT token
	token, err := c.generateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	url := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		url = BaseURL + path
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// generateJWT generates a JWT for ASC API authentication
func (c *Client) generateJWT() (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    c.issuerID,
		Audience:  jwt.ClaimStrings{"appstoreconnect-v1"},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenLifetime)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = c.keyID

	// Sign with the private key
	signedToken, err := token.SignedString(c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// do performs an HTTP request and returns the response
func (c *Client) do(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		if err := ParseError(respBody); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) doStream(ctx context.Context, method, path string, body io.Reader, accept string) (*http.Response, error) {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(accept) != "" {
		req.Header.Set("Accept", accept)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err := ParseError(respBody); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}
	return resp, nil
}

func (c *Client) doStreamNoAuth(ctx context.Context, method, rawURL, accept string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if strings.TrimSpace(accept) != "" {
		req.Header.Set("Accept", accept)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err := ParseError(respBody); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}
	return resp, nil
}

func buildReviewQuery(opts []ReviewOption) string {
	query := &reviewQuery{}
	for _, opt := range opts {
		opt(query)
	}

	values := url.Values{}
	if query.territory != "" {
		values.Set("filter[territory]", query.territory)
	}
	if query.rating >= 1 && query.rating <= 5 {
		values.Set("filter[rating]", fmt.Sprintf("%d", query.rating))
	}
	if query.sort != "" {
		values.Set("sort", query.sort)
	}
	addLimit(values, query.limit)

	return values.Encode()
}

func buildFeedbackQuery(query *feedbackQuery) string {
	values := url.Values{}
	if query.includeScreenshots {
		values.Set("fields[betaFeedbackScreenshotSubmissions]", strings.Join([]string{
			"createdDate",
			"comment",
			"email",
			"deviceModel",
			"osVersion",
			"appPlatform",
			"devicePlatform",
			"screenshots",
		}, ","))
	}
	addCSV(values, "filter[deviceModel]", query.deviceModels)
	addCSV(values, "filter[osVersion]", query.osVersions)
	addCSV(values, "filter[appPlatform]", query.appPlatforms)
	addCSV(values, "filter[devicePlatform]", query.devicePlatforms)
	addCSV(values, "filter[build]", query.buildIDs)
	addCSV(values, "filter[build.preReleaseVersion]", query.buildPreReleaseVersionIDs)
	addCSV(values, "filter[tester]", query.testerIDs)
	if query.sort != "" {
		values.Set("sort", query.sort)
	}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildCrashQuery(query *crashQuery) string {
	values := url.Values{}
	addCSV(values, "filter[deviceModel]", query.deviceModels)
	addCSV(values, "filter[osVersion]", query.osVersions)
	addCSV(values, "filter[appPlatform]", query.appPlatforms)
	addCSV(values, "filter[devicePlatform]", query.devicePlatforms)
	addCSV(values, "filter[build]", query.buildIDs)
	addCSV(values, "filter[build.preReleaseVersion]", query.buildPreReleaseVersionIDs)
	addCSV(values, "filter[tester]", query.testerIDs)
	if query.sort != "" {
		values.Set("sort", query.sort)
	}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildBetaGroupsQuery(query *betaGroupsQuery) string {
	values := url.Values{}
	addLimit(values, query.limit)
	return values.Encode()
}

func buildBetaTestersQuery(appID string, query *betaTestersQuery) string {
	values := url.Values{}
	if strings.TrimSpace(appID) != "" {
		values.Set("filter[apps]", strings.TrimSpace(appID))
	}
	if strings.TrimSpace(query.email) != "" {
		values.Set("filter[email]", strings.TrimSpace(query.email))
	}
	addCSV(values, "filter[betaGroups]", query.groupIDs)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildAppStoreVersionsQuery(query *appStoreVersionsQuery) string {
	values := url.Values{}
	addCSV(values, "filter[platform]", query.platforms)
	addCSV(values, "filter[versionString]", query.versionStrings)
	addCSV(values, "filter[appStoreState]", query.states)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildAppStoreVersionLocalizationsQuery(query *appStoreVersionLocalizationsQuery) string {
	values := url.Values{}
	addCSV(values, "filter[locale]", query.locales)
	addLimit(values, query.limit)
	return values.Encode()
}

func buildAppInfoLocalizationsQuery(query *appInfoLocalizationsQuery) string {
	values := url.Values{}
	addCSV(values, "filter[locale]", query.locales)
	addLimit(values, query.limit)
	return values.Encode()
}

func normalizeList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeUpperList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized = append(normalized, strings.ToUpper(value))
	}
	return normalized
}

func addCSV(values url.Values, key string, items []string) {
	items = normalizeList(items)
	if len(items) == 0 {
		return
	}
	values.Set(key, strings.Join(items, ","))
}

func addLimit(values url.Values, limit int) {
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}
}

// GetFeedback retrieves TestFlight feedback
func (c *Client) GetFeedback(ctx context.Context, appID string, opts ...FeedbackOption) (*FeedbackResponse, error) {
	query := &feedbackQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/apps/%s/betaFeedbackScreenshotSubmissions", appID)
	if query.nextURL != "" {
		path = query.nextURL
	} else if queryString := buildFeedbackQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response FeedbackResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetCrashes retrieves TestFlight crash reports
func (c *Client) GetCrashes(ctx context.Context, appID string, opts ...CrashOption) (*CrashesResponse, error) {
	query := &crashQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/apps/%s/betaFeedbackCrashSubmissions", appID)
	if query.nextURL != "" {
		path = query.nextURL
	} else if queryString := buildCrashQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response CrashesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetReviews retrieves App Store reviews
func (c *Client) GetReviews(ctx context.Context, appID string, opts ...ReviewOption) (*ReviewsResponse, error) {
	query := &reviewQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/apps/%s/customerReviews", appID)
	if query.nextURL != "" {
		path = query.nextURL
	} else if queryString := buildReviewQuery(opts); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response ReviewsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetApps retrieves the list of apps
func (c *Client) GetApps(ctx context.Context, opts ...AppsOption) (*AppsResponse, error) {
	query := &appsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := "/v1/apps"
	if query.nextURL != "" {
		path = query.nextURL
	} else {
		values := url.Values{}
		if query.sort != "" {
			values.Set("sort", query.sort)
		}
		if query.limit > 0 {
			values.Set("limit", strconv.Itoa(query.limit))
		}
		if queryString := values.Encode(); queryString != "" {
			path += "?" + queryString
		}
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response AppsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetBuilds retrieves the list of builds for an app
func (c *Client) GetBuilds(ctx context.Context, appID string, opts ...BuildsOption) (*BuildsResponse, error) {
	query := &buildsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/apps/%s/builds", appID)
	if query.nextURL != "" {
		path = query.nextURL
	} else {
		values := url.Values{}
		// Use /v1/builds endpoint when sorting or limiting, since /v1/apps/{id}/builds doesn't support these
		if query.sort != "" || query.limit > 0 {
			path = "/v1/builds"
			values.Set("filter[app]", appID)
			if query.sort != "" {
				values.Set("sort", query.sort)
			}
			if query.limit > 0 {
				values.Set("limit", strconv.Itoa(query.limit))
			}
		}
		if queryString := values.Encode(); queryString != "" {
			path += "?" + queryString
		}
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response BuildsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetAppStoreVersions retrieves app store versions for an app.
func (c *Client) GetAppStoreVersions(ctx context.Context, appID string, opts ...AppStoreVersionsOption) (*AppStoreVersionsResponse, error) {
	query := &appStoreVersionsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/apps/%s/appStoreVersions", appID)
	if query.nextURL != "" {
		path = query.nextURL
	} else if queryString := buildAppStoreVersionsQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response AppStoreVersionsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetAppStoreVersion retrieves an app store version by ID.
func (c *Client) GetAppStoreVersion(ctx context.Context, versionID string) (*AppStoreVersionResponse, error) {
	path := fmt.Sprintf("/v1/appStoreVersions/%s", versionID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response AppStoreVersionResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// AttachBuildToVersion attaches a build to an app store version.
func (c *Client) AttachBuildToVersion(ctx context.Context, versionID, buildID string) error {
	request := AppStoreVersionBuildRelationshipUpdateRequest{
		Data: ResourceData{
			Type: ResourceTypeBuilds,
			ID:   buildID,
		},
	}
	body, err := BuildRequestBody(request)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v1/appStoreVersions/%s/relationships/build", versionID)
	if _, err := c.do(ctx, "PATCH", path, body); err != nil {
		return err
	}
	return nil
}

// GetAppStoreVersionBuild retrieves the build attached to a version.
func (c *Client) GetAppStoreVersionBuild(ctx context.Context, versionID string) (*BuildResponse, error) {
	path := fmt.Sprintf("/v1/appStoreVersions/%s/build", versionID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response BuildResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetAppStoreVersionSubmissionResource retrieves a submission by submission ID.
func (c *Client) GetAppStoreVersionSubmissionResource(ctx context.Context, submissionID string) (*AppStoreVersionSubmissionResourceResponse, error) {
	path := fmt.Sprintf("/v1/appStoreVersionSubmissions/%s", submissionID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response AppStoreVersionSubmissionResourceResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetAppStoreVersionSubmissionForVersion retrieves a submission by version ID.
func (c *Client) GetAppStoreVersionSubmissionForVersion(ctx context.Context, versionID string) (*AppStoreVersionSubmissionResourceResponse, error) {
	path := fmt.Sprintf("/v1/appStoreVersions/%s/appStoreVersionSubmission", versionID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response AppStoreVersionSubmissionResourceResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetBetaGroups retrieves the list of beta groups for an app.
func (c *Client) GetBetaGroups(ctx context.Context, appID string, opts ...BetaGroupsOption) (*BetaGroupsResponse, error) {
	query := &betaGroupsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/apps/%s/betaGroups", appID)
	if query.nextURL != "" {
		path = query.nextURL
	} else if queryString := buildBetaGroupsQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response BetaGroupsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreateBetaGroup creates a beta group for an app.
func (c *Client) CreateBetaGroup(ctx context.Context, appID, name string) (*BetaGroupResponse, error) {
	payload := BetaGroupCreateRequest{
		Data: BetaGroupCreateData{
			Type:       ResourceTypeBetaGroups,
			Attributes: BetaGroupAttributes{Name: name},
			Relationships: &BetaGroupRelationships{
				App: &Relationship{
					Data: ResourceData{
						Type: ResourceTypeApps,
						ID:   appID,
					},
				},
			},
		},
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/betaGroups", body)
	if err != nil {
		return nil, err
	}

	var response BetaGroupResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetBetaTesters retrieves beta testers for an app.
func (c *Client) GetBetaTesters(ctx context.Context, appID string, opts ...BetaTestersOption) (*BetaTestersResponse, error) {
	query := &betaTestersQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := "/v1/betaTesters"
	if query.nextURL != "" {
		path = query.nextURL
	} else if queryString := buildBetaTestersQuery(appID, query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response BetaTestersResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreateBetaTester creates a beta tester.
func (c *Client) CreateBetaTester(ctx context.Context, email, firstName, lastName string, groupIDs []string) (*BetaTesterResponse, error) {
	groupIDs = normalizeList(groupIDs)
	var relationships *BetaTesterCreateRelationships
	if len(groupIDs) > 0 {
		relData := make([]ResourceData, 0, len(groupIDs))
		for _, groupID := range groupIDs {
			relData = append(relData, ResourceData{
				Type: ResourceTypeBetaGroups,
				ID:   groupID,
			})
		}
		relationships = &BetaTesterCreateRelationships{
			BetaGroups: &RelationshipList{Data: relData},
		}
	}

	payload := BetaTesterCreateRequest{
		Data: BetaTesterCreateData{
			Type: ResourceTypeBetaTesters,
			Attributes: BetaTesterCreateAttributes{
				FirstName: strings.TrimSpace(firstName),
				LastName:  strings.TrimSpace(lastName),
				Email:     strings.TrimSpace(email),
			},
			Relationships: relationships,
		},
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/betaTesters", body)
	if err != nil {
		return nil, err
	}

	var response BetaTesterResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// DeleteBetaTester deletes a beta tester by ID.
func (c *Client) DeleteBetaTester(ctx context.Context, testerID string) error {
	path := fmt.Sprintf("/v1/betaTesters/%s", testerID)
	_, err := c.do(ctx, "DELETE", path, nil)
	return err
}

// CreateBetaTesterInvitation creates a beta tester invitation.
func (c *Client) CreateBetaTesterInvitation(ctx context.Context, appID, testerID string) (*BetaTesterInvitationResponse, error) {
	payload := BetaTesterInvitationCreateRequest{
		Data: BetaTesterInvitationCreateData{
			Type: ResourceTypeBetaTesterInvitations,
			Relationships: &BetaTesterInvitationCreateRelationships{
				App: &Relationship{
					Data: ResourceData{
						Type: ResourceTypeApps,
						ID:   appID,
					},
				},
				BetaTester: &Relationship{
					Data: ResourceData{
						Type: ResourceTypeBetaTesters,
						ID:   testerID,
					},
				},
			},
		},
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/betaTesterInvitations", body)
	if err != nil {
		return nil, err
	}

	var response BetaTesterInvitationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetAppStoreVersionLocalizations retrieves localizations for an app store version.
func (c *Client) GetAppStoreVersionLocalizations(ctx context.Context, versionID string, opts ...AppStoreVersionLocalizationsOption) (*AppStoreVersionLocalizationsResponse, error) {
	query := &appStoreVersionLocalizationsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/appStoreVersions/%s/appStoreVersionLocalizations", versionID)
	if query.nextURL != "" {
		path = query.nextURL
	} else if queryString := buildAppStoreVersionLocalizationsQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response AppStoreVersionLocalizationsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreateAppStoreVersionLocalization creates a localization for an app store version.
func (c *Client) CreateAppStoreVersionLocalization(ctx context.Context, versionID string, attributes AppStoreVersionLocalizationAttributes) (*AppStoreVersionLocalizationResponse, error) {
	payload := AppStoreVersionLocalizationCreateRequest{
		Data: AppStoreVersionLocalizationCreateData{
			Type:       ResourceTypeAppStoreVersionLocalizations,
			Attributes: attributes,
			Relationships: &AppStoreVersionLocalizationRelationships{
				AppStoreVersion: &Relationship{
					Data: ResourceData{
						Type: ResourceTypeAppStoreVersions,
						ID:   versionID,
					},
				},
			},
		},
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/appStoreVersionLocalizations", body)
	if err != nil {
		return nil, err
	}

	var response AppStoreVersionLocalizationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// UpdateAppStoreVersionLocalization updates a localization for an app store version.
func (c *Client) UpdateAppStoreVersionLocalization(ctx context.Context, localizationID string, attributes AppStoreVersionLocalizationAttributes) (*AppStoreVersionLocalizationResponse, error) {
	payload := AppStoreVersionLocalizationUpdateRequest{
		Data: AppStoreVersionLocalizationUpdateData{
			Type:       ResourceTypeAppStoreVersionLocalizations,
			ID:         localizationID,
			Attributes: attributes,
		},
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/appStoreVersionLocalizations/%s", localizationID)
	data, err := c.do(ctx, "PATCH", path, body)
	if err != nil {
		return nil, err
	}

	var response AppStoreVersionLocalizationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetAppInfoLocalizations retrieves localizations for an app info resource.
func (c *Client) GetAppInfoLocalizations(ctx context.Context, appInfoID string, opts ...AppInfoLocalizationsOption) (*AppInfoLocalizationsResponse, error) {
	query := &appInfoLocalizationsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/appInfos/%s/appInfoLocalizations", appInfoID)
	if query.nextURL != "" {
		path = query.nextURL
	} else if queryString := buildAppInfoLocalizationsQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response AppInfoLocalizationsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreateAppInfoLocalization creates a localization for an app info resource.
func (c *Client) CreateAppInfoLocalization(ctx context.Context, appInfoID string, attributes AppInfoLocalizationAttributes) (*AppInfoLocalizationResponse, error) {
	payload := AppInfoLocalizationCreateRequest{
		Data: AppInfoLocalizationCreateData{
			Type:       ResourceTypeAppInfoLocalizations,
			Attributes: attributes,
			Relationships: &AppInfoLocalizationRelationships{
				AppInfo: &Relationship{
					Data: ResourceData{
						Type: ResourceTypeAppInfos,
						ID:   appInfoID,
					},
				},
			},
		},
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/appInfoLocalizations", body)
	if err != nil {
		return nil, err
	}

	var response AppInfoLocalizationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// UpdateAppInfoLocalization updates a localization for an app info resource.
func (c *Client) UpdateAppInfoLocalization(ctx context.Context, localizationID string, attributes AppInfoLocalizationAttributes) (*AppInfoLocalizationResponse, error) {
	payload := AppInfoLocalizationUpdateRequest{
		Data: AppInfoLocalizationUpdateData{
			Type:       ResourceTypeAppInfoLocalizations,
			ID:         localizationID,
			Attributes: attributes,
		},
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/appInfoLocalizations/%s", localizationID)
	data, err := c.do(ctx, "PATCH", path, body)
	if err != nil {
		return nil, err
	}

	var response AppInfoLocalizationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetAppInfos retrieves app info records for an app.
func (c *Client) GetAppInfos(ctx context.Context, appID string) (*AppInfosResponse, error) {
	path := fmt.Sprintf("/v1/apps/%s/appInfos", appID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response AppInfosResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetBuild retrieves a single build by ID.
func (c *Client) GetBuild(ctx context.Context, buildID string) (*BuildResponse, error) {
	path := fmt.Sprintf("/v1/builds/%s", buildID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response BuildResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// ExpireBuild expires a build for TestFlight testing.
func (c *Client) ExpireBuild(ctx context.Context, buildID string) (*BuildResponse, error) {
	payload := struct {
		Data struct {
			Type       ResourceType `json:"type"`
			ID         string       `json:"id"`
			Attributes struct {
				Expired bool `json:"expired"`
			} `json:"attributes"`
		} `json:"data"`
	}{}
	payload.Data.Type = ResourceTypeBuilds
	payload.Data.ID = buildID
	payload.Data.Attributes.Expired = true

	body, err := BuildRequestBody(payload)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/builds/%s", buildID)
	data, err := c.do(ctx, "PATCH", path, body)
	if err != nil {
		return nil, err
	}

	var response BuildResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreateBuildUpload creates a new build upload record.
func (c *Client) CreateBuildUpload(ctx context.Context, req BuildUploadCreateRequest) (*BuildUploadResponse, error) {
	body, err := BuildRequestBody(req)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/buildUploads", body)
	if err != nil {
		return nil, err
	}

	var response BuildUploadResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetBuildUpload retrieves a build upload by ID.
func (c *Client) GetBuildUpload(ctx context.Context, id string) (*BuildUploadResponse, error) {
	data, err := c.do(ctx, "GET", fmt.Sprintf("/v1/buildUploads/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var response BuildUploadResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreateBuildUploadFile creates a new build upload file reservation.
func (c *Client) CreateBuildUploadFile(ctx context.Context, req BuildUploadFileCreateRequest) (*BuildUploadFileResponse, error) {
	body, err := BuildRequestBody(req)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/buildUploadFiles", body)
	if err != nil {
		return nil, err
	}

	var response BuildUploadFileResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// UpdateBuildUploadFile updates a build upload file (used to commit upload).
func (c *Client) UpdateBuildUploadFile(ctx context.Context, id string, req BuildUploadFileUpdateRequest) (*BuildUploadFileResponse, error) {
	body, err := BuildRequestBody(req)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "PATCH", fmt.Sprintf("/v1/buildUploadFiles/%s", id), body)
	if err != nil {
		return nil, err
	}

	var response BuildUploadFileResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreateAppStoreVersionSubmission creates a new app store version submission.
func (c *Client) CreateAppStoreVersionSubmission(ctx context.Context, req AppStoreVersionSubmissionCreateRequest) (*AppStoreVersionSubmissionResponse, error) {
	body, err := BuildRequestBody(req)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/appStoreVersionSubmissions", body)
	if err != nil {
		return nil, err
	}

	var response AppStoreVersionSubmissionResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetAppStoreVersionSubmission retrieves an app store version submission by ID.
func (c *Client) GetAppStoreVersionSubmission(ctx context.Context, id string) (*AppStoreVersionSubmissionResponse, error) {
	data, err := c.do(ctx, "GET", fmt.Sprintf("/v1/appStoreVersionSubmissions/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var response AppStoreVersionSubmissionResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// DeleteAppStoreVersionSubmission deletes an app store version submission.
func (c *Client) DeleteAppStoreVersionSubmission(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", fmt.Sprintf("/v1/appStoreVersionSubmissions/%s", id), nil)
	return err
}

// Links represents pagination links
type Links struct {
	Self string `json:"self,omitempty"`
	Next string `json:"next,omitempty"`
	Prev string `json:"prev,omitempty"`
}

// PrintJSON prints data as minified JSON (best for AI agents)
func PrintJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	return enc.Encode(data)
}

// PrintPrettyJSON prints data as indented JSON (best for debugging).
func PrintPrettyJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// PrintMarkdown prints data as Markdown table
func PrintMarkdown(data interface{}) error {
	switch v := data.(type) {
	case *FeedbackResponse:
		return printFeedbackMarkdown(v)
	case *CrashesResponse:
		return printCrashesMarkdown(v)
	case *ReviewsResponse:
		return printReviewsMarkdown(v)
	case *AppsResponse:
		return printAppsMarkdown(v)
	case *BuildsResponse:
		return printBuildsMarkdown(v)
	case *AppStoreVersionsResponse:
		return printAppStoreVersionsMarkdown(v)
	case *BuildResponse:
		return printBuildsMarkdown(&BuildsResponse{Data: []Resource[BuildAttributes]{v.Data}})
	case *AppStoreVersionLocalizationsResponse:
		return printAppStoreVersionLocalizationsMarkdown(v)
	case *AppInfoLocalizationsResponse:
		return printAppInfoLocalizationsMarkdown(v)
	case *BetaGroupsResponse:
		return printBetaGroupsMarkdown(v)
	case *BetaGroupResponse:
		return printBetaGroupsMarkdown(&BetaGroupsResponse{Data: []Resource[BetaGroupAttributes]{v.Data}})
	case *BetaTestersResponse:
		return printBetaTestersMarkdown(v)
	case *BetaTesterResponse:
		return printBetaTestersMarkdown(&BetaTestersResponse{Data: []Resource[BetaTesterAttributes]{v.Data}})
	case *LocalizationDownloadResult:
		return printLocalizationDownloadResultMarkdown(v)
	case *LocalizationUploadResult:
		return printLocalizationUploadResultMarkdown(v)
	case *BuildUploadResult:
		return printBuildUploadResultMarkdown(v)
	case *SalesReportResult:
		return printSalesReportResultMarkdown(v)
	case *AnalyticsReportRequestResult:
		return printAnalyticsReportRequestResultMarkdown(v)
	case *AnalyticsReportRequestsResponse:
		return printAnalyticsReportRequestsMarkdown(v)
	case *AnalyticsReportRequestResponse:
		return printAnalyticsReportRequestsMarkdown(&AnalyticsReportRequestsResponse{Data: []AnalyticsReportRequestResource{v.Data}, Links: v.Links})
	case *AnalyticsReportDownloadResult:
		return printAnalyticsReportDownloadResultMarkdown(v)
	case *AnalyticsReportGetResult:
		return printAnalyticsReportGetResultMarkdown(v)
	case *AppStoreVersionSubmissionResult:
		return printAppStoreVersionSubmissionMarkdown(v)
	case *AppStoreVersionSubmissionCreateResult:
		return printAppStoreVersionSubmissionCreateMarkdown(v)
	case *AppStoreVersionSubmissionStatusResult:
		return printAppStoreVersionSubmissionStatusMarkdown(v)
	case *AppStoreVersionSubmissionCancelResult:
		return printAppStoreVersionSubmissionCancelMarkdown(v)
	case *AppStoreVersionDetailResult:
		return printAppStoreVersionDetailMarkdown(v)
	case *AppStoreVersionAttachBuildResult:
		return printAppStoreVersionAttachBuildMarkdown(v)
	case *BetaTesterDeleteResult:
		return printBetaTesterDeleteResultMarkdown(v)
	case *BetaTesterInvitationResult:
		return printBetaTesterInvitationResultMarkdown(v)
	default:
		return PrintJSON(data)
	}
}

// PrintTable prints data as a formatted table
func PrintTable(data interface{}) error {
	switch v := data.(type) {
	case *FeedbackResponse:
		return printFeedbackTable(v)
	case *CrashesResponse:
		return printCrashesTable(v)
	case *ReviewsResponse:
		return printReviewsTable(v)
	case *AppsResponse:
		return printAppsTable(v)
	case *BuildsResponse:
		return printBuildsTable(v)
	case *AppStoreVersionsResponse:
		return printAppStoreVersionsTable(v)
	case *BuildResponse:
		return printBuildsTable(&BuildsResponse{Data: []Resource[BuildAttributes]{v.Data}})
	case *AppStoreVersionLocalizationsResponse:
		return printAppStoreVersionLocalizationsTable(v)
	case *AppInfoLocalizationsResponse:
		return printAppInfoLocalizationsTable(v)
	case *BetaGroupsResponse:
		return printBetaGroupsTable(v)
	case *BetaGroupResponse:
		return printBetaGroupsTable(&BetaGroupsResponse{Data: []Resource[BetaGroupAttributes]{v.Data}})
	case *BetaTestersResponse:
		return printBetaTestersTable(v)
	case *BetaTesterResponse:
		return printBetaTestersTable(&BetaTestersResponse{Data: []Resource[BetaTesterAttributes]{v.Data}})
	case *LocalizationDownloadResult:
		return printLocalizationDownloadResultTable(v)
	case *LocalizationUploadResult:
		return printLocalizationUploadResultTable(v)
	case *BuildUploadResult:
		return printBuildUploadResultTable(v)
	case *SalesReportResult:
		return printSalesReportResultTable(v)
	case *AnalyticsReportRequestResult:
		return printAnalyticsReportRequestResultTable(v)
	case *AnalyticsReportRequestsResponse:
		return printAnalyticsReportRequestsTable(v)
	case *AnalyticsReportRequestResponse:
		return printAnalyticsReportRequestsTable(&AnalyticsReportRequestsResponse{Data: []AnalyticsReportRequestResource{v.Data}, Links: v.Links})
	case *AnalyticsReportDownloadResult:
		return printAnalyticsReportDownloadResultTable(v)
	case *AnalyticsReportGetResult:
		return printAnalyticsReportGetResultTable(v)
	case *AppStoreVersionSubmissionResult:
		return printAppStoreVersionSubmissionTable(v)
	case *AppStoreVersionSubmissionCreateResult:
		return printAppStoreVersionSubmissionCreateTable(v)
	case *AppStoreVersionSubmissionStatusResult:
		return printAppStoreVersionSubmissionStatusTable(v)
	case *AppStoreVersionSubmissionCancelResult:
		return printAppStoreVersionSubmissionCancelTable(v)
	case *AppStoreVersionDetailResult:
		return printAppStoreVersionDetailTable(v)
	case *AppStoreVersionAttachBuildResult:
		return printAppStoreVersionAttachBuildTable(v)
	case *BetaTesterDeleteResult:
		return printBetaTesterDeleteResultTable(v)
	case *BetaTesterInvitationResult:
		return printBetaTesterInvitationResultTable(v)
	default:
		return PrintJSON(data)
	}
}

// BuildRequestBody builds a JSON request body
func BuildRequestBody(data interface{}) (io.Reader, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}
	return &buf, nil
}

// ParseError parses an error response
func ParseError(body []byte) error {
	var errResp struct {
		Errors []struct {
			Code   string `json:"code"`
			Title  string `json:"title"`
			Detail string `json:"detail"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil && len(errResp.Errors) > 0 {
		return fmt.Errorf("%s: %s", errResp.Errors[0].Title, errResp.Errors[0].Detail)
	}

	return fmt.Errorf("unknown error: %s", string(body))
}

// IsNotFound checks if the error is a "not found" error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not_found") ||
		strings.Contains(message, "not found") ||
		strings.Contains(message, "resource does not exist")
}

// IsUnauthorized checks if the error is an "unauthorized" error
func IsUnauthorized(err error) bool {
	return strings.Contains(err.Error(), "UNAUTHORIZED")
}

func compactWhitespace(input string) string {
	return strings.Join(strings.Fields(input), " ")
}

func escapeMarkdown(input string) string {
	clean := compactWhitespace(input)
	return strings.ReplaceAll(clean, "|", "\\|")
}

func feedbackHasScreenshots(resp *FeedbackResponse) bool {
	for _, item := range resp.Data {
		if len(item.Attributes.Screenshots) > 0 {
			return true
		}
	}
	return false
}

func formatScreenshotURLs(images []FeedbackScreenshotImage) string {
	if len(images) == 0 {
		return ""
	}
	urls := make([]string, 0, len(images))
	for _, image := range images {
		if strings.TrimSpace(image.URL) == "" {
			continue
		}
		urls = append(urls, image.URL)
	}
	return strings.Join(urls, ", ")
}

func printFeedbackTable(resp *FeedbackResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	hasScreenshots := feedbackHasScreenshots(resp)
	if hasScreenshots {
		fmt.Fprintln(w, "Created\tEmail\tComment\tScreenshots")
	} else {
		fmt.Fprintln(w, "Created\tEmail\tComment")
	}
	for _, item := range resp.Data {
		if hasScreenshots {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				item.Attributes.CreatedDate,
				item.Attributes.Email,
				compactWhitespace(item.Attributes.Comment),
				formatScreenshotURLs(item.Attributes.Screenshots),
			)
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			item.Attributes.CreatedDate,
			item.Attributes.Email,
			compactWhitespace(item.Attributes.Comment),
		)
	}
	return w.Flush()
}

func printCrashesTable(resp *CrashesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Created\tEmail\tDevice\tOS\tComment")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.Attributes.CreatedDate,
			item.Attributes.Email,
			item.Attributes.DeviceModel,
			item.Attributes.OSVersion,
			compactWhitespace(item.Attributes.Comment),
		)
	}
	return w.Flush()
}

func printReviewsTable(resp *ReviewsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Created\tRating\tTerritory\tTitle")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n",
			item.Attributes.CreatedDate,
			item.Attributes.Rating,
			item.Attributes.Territory,
			compactWhitespace(item.Attributes.Title),
		)
	}
	return w.Flush()
}

func printFeedbackMarkdown(resp *FeedbackResponse) error {
	hasScreenshots := feedbackHasScreenshots(resp)
	if hasScreenshots {
		fmt.Fprintln(os.Stdout, "| Created | Email | Comment | Screenshots |")
		fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	} else {
		fmt.Fprintln(os.Stdout, "| Created | Email | Comment |")
		fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	}
	for _, item := range resp.Data {
		if hasScreenshots {
			fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s |\n",
				escapeMarkdown(item.Attributes.CreatedDate),
				escapeMarkdown(item.Attributes.Email),
				escapeMarkdown(item.Attributes.Comment),
				escapeMarkdown(formatScreenshotURLs(item.Attributes.Screenshots)),
			)
			continue
		}
		fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
			escapeMarkdown(item.Attributes.CreatedDate),
			escapeMarkdown(item.Attributes.Email),
			escapeMarkdown(item.Attributes.Comment),
		)
	}
	return nil
}

func printCrashesMarkdown(resp *CrashesResponse) error {
	fmt.Fprintln(os.Stdout, "| Created | Email | Device | OS | Comment |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.Attributes.CreatedDate),
			escapeMarkdown(item.Attributes.Email),
			escapeMarkdown(item.Attributes.DeviceModel),
			escapeMarkdown(item.Attributes.OSVersion),
			escapeMarkdown(item.Attributes.Comment),
		)
	}
	return nil
}

func printReviewsMarkdown(resp *ReviewsResponse) error {
	fmt.Fprintln(os.Stdout, "| Created | Rating | Territory | Title |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %d | %s | %s |\n",
			escapeMarkdown(item.Attributes.CreatedDate),
			item.Attributes.Rating,
			escapeMarkdown(item.Attributes.Territory),
			escapeMarkdown(item.Attributes.Title),
		)
	}
	return nil
}

func printAppsTable(resp *AppsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tBundle ID\tSKU")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			item.ID,
			compactWhitespace(item.Attributes.Name),
			item.Attributes.BundleID,
			item.Attributes.SKU,
		)
	}
	return w.Flush()
}

func printAppStoreVersionLocalizationsTable(resp *AppStoreVersionLocalizationsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Locale\tWhats New\tKeywords")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			item.Attributes.Locale,
			compactWhitespace(item.Attributes.WhatsNew),
			compactWhitespace(item.Attributes.Keywords),
		)
	}
	return w.Flush()
}

func printAppInfoLocalizationsTable(resp *AppInfoLocalizationsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Locale\tName\tSubtitle\tPrivacy Policy URL")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			item.Attributes.Locale,
			compactWhitespace(item.Attributes.Name),
			compactWhitespace(item.Attributes.Subtitle),
			item.Attributes.PrivacyPolicyURL,
		)
	}
	return w.Flush()
}

func printBetaGroupsTable(resp *BetaGroupsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tInternal\tPublic Link Enabled\tPublic Link")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%t\t%t\t%s\n",
			item.ID,
			compactWhitespace(item.Attributes.Name),
			item.Attributes.IsInternalGroup,
			item.Attributes.PublicLinkEnabled,
			item.Attributes.PublicLink,
		)
	}
	return w.Flush()
}

func formatBetaTesterName(attr BetaTesterAttributes) string {
	first := strings.TrimSpace(attr.FirstName)
	last := strings.TrimSpace(attr.LastName)
	switch {
	case first == "" && last == "":
		return ""
	case first == "":
		return last
	case last == "":
		return first
	default:
		return first + " " + last
	}
}

func printBetaTestersTable(resp *BetaTestersResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tEmail\tName\tState\tInvite")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.Attributes.Email,
			compactWhitespace(formatBetaTesterName(item.Attributes)),
			string(item.Attributes.State),
			string(item.Attributes.InviteType),
		)
	}
	return w.Flush()
}

func printBuildsTable(resp *BuildsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Version\tUploaded\tProcessing\tExpired")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%t\n",
			item.Attributes.Version,
			item.Attributes.UploadedDate,
			item.Attributes.ProcessingState,
			item.Attributes.Expired,
		)
	}
	return w.Flush()
}

func printAppStoreVersionsTable(resp *AppStoreVersionsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tVersion\tPlatform\tState\tCreated")
	for _, item := range resp.Data {
		state := item.Attributes.AppVersionState
		if state == "" {
			state = item.Attributes.AppStoreState
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.Attributes.VersionString,
			string(item.Attributes.Platform),
			state,
			item.Attributes.CreatedDate,
		)
	}
	return w.Flush()
}

func printAppsMarkdown(resp *AppsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Bundle ID | SKU |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s |\n",
			item.ID,
			escapeMarkdown(item.Attributes.Name),
			escapeMarkdown(item.Attributes.BundleID),
			escapeMarkdown(item.Attributes.SKU),
		)
	}
	return nil
}

func printAppStoreVersionsMarkdown(resp *AppStoreVersionsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Version | Platform | State | Created |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		state := item.Attributes.AppVersionState
		if state == "" {
			state = item.Attributes.AppStoreState
		}
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.VersionString),
			escapeMarkdown(string(item.Attributes.Platform)),
			escapeMarkdown(state),
			escapeMarkdown(item.Attributes.CreatedDate),
		)
	}
	return nil
}

func printAppStoreVersionLocalizationsMarkdown(resp *AppStoreVersionLocalizationsResponse) error {
	fmt.Fprintln(os.Stdout, "| Locale | Whats New | Keywords |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
			escapeMarkdown(item.Attributes.Locale),
			escapeMarkdown(item.Attributes.WhatsNew),
			escapeMarkdown(item.Attributes.Keywords),
		)
	}
	return nil
}

func printAppInfoLocalizationsMarkdown(resp *AppInfoLocalizationsResponse) error {
	fmt.Fprintln(os.Stdout, "| Locale | Name | Subtitle | Privacy Policy URL |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s |\n",
			escapeMarkdown(item.Attributes.Locale),
			escapeMarkdown(item.Attributes.Name),
			escapeMarkdown(item.Attributes.Subtitle),
			escapeMarkdown(item.Attributes.PrivacyPolicyURL),
		)
	}
	return nil
}

func printBetaGroupsMarkdown(resp *BetaGroupsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Internal | Public Link Enabled | Public Link |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %t | %t | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.Name),
			item.Attributes.IsInternalGroup,
			item.Attributes.PublicLinkEnabled,
			escapeMarkdown(item.Attributes.PublicLink),
		)
	}
	return nil
}

func printBetaTestersMarkdown(resp *BetaTestersResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Email | Name | State | Invite |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.Email),
			escapeMarkdown(formatBetaTesterName(item.Attributes)),
			escapeMarkdown(string(item.Attributes.State)),
			escapeMarkdown(string(item.Attributes.InviteType)),
		)
	}
	return nil
}

func printBuildsMarkdown(resp *BuildsResponse) error {
	fmt.Fprintln(os.Stdout, "| Version | Uploaded | Processing | Expired |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %t |\n",
			escapeMarkdown(item.Attributes.Version),
			escapeMarkdown(item.Attributes.UploadedDate),
			escapeMarkdown(item.Attributes.ProcessingState),
			item.Attributes.Expired,
		)
	}
	return nil
}

func printBuildUploadResultTable(result *BuildUploadResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Upload ID\tFile ID\tFile Name\tFile Size")
	fmt.Fprintf(w, "%s\t%s\t%s\t%d\n",
		result.UploadID,
		result.FileID,
		result.FileName,
		result.FileSize,
	)
	if err := w.Flush(); err != nil {
		return err
	}
	if len(result.Operations) == 0 {
		return nil
	}
	fmt.Fprintln(os.Stdout, "\nUpload Operations")
	opsWriter := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(opsWriter, "Method\tURL\tLength\tOffset")
	for _, op := range result.Operations {
		fmt.Fprintf(opsWriter, "%s\t%s\t%d\t%d\n",
			op.Method,
			op.URL,
			op.Length,
			op.Offset,
		)
	}
	return opsWriter.Flush()
}

func printAppStoreVersionSubmissionTable(result *AppStoreVersionSubmissionResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Submission ID\tCreated Date")
	createdDate := ""
	if result.CreatedDate != nil {
		createdDate = *result.CreatedDate
	}
	fmt.Fprintf(w, "%s\t%s\n", result.SubmissionID, createdDate)
	return w.Flush()
}

func printAppStoreVersionSubmissionCreateTable(result *AppStoreVersionSubmissionCreateResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Submission ID\tVersion ID\tBuild ID\tCreated Date")
	createdDate := ""
	if result.CreatedDate != nil {
		createdDate = *result.CreatedDate
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		result.SubmissionID,
		result.VersionID,
		result.BuildID,
		createdDate,
	)
	return w.Flush()
}

func printAppStoreVersionSubmissionStatusTable(result *AppStoreVersionSubmissionStatusResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Submission ID\tVersion ID\tVersion\tPlatform\tState\tCreated Date")
	createdDate := ""
	if result.CreatedDate != nil {
		createdDate = *result.CreatedDate
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
		result.ID,
		result.VersionID,
		result.VersionString,
		result.Platform,
		result.State,
		createdDate,
	)
	return w.Flush()
}

func printAppStoreVersionSubmissionCancelTable(result *AppStoreVersionSubmissionCancelResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Submission ID\tCancelled")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Cancelled)
	return w.Flush()
}

func printAppStoreVersionDetailTable(result *AppStoreVersionDetailResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Version ID\tVersion\tPlatform\tState\tBuild ID\tBuild Version\tSubmission ID")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		result.ID,
		result.VersionString,
		result.Platform,
		result.State,
		result.BuildID,
		result.BuildVersion,
		result.SubmissionID,
	)
	return w.Flush()
}

func printAppStoreVersionAttachBuildTable(result *AppStoreVersionAttachBuildResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Version ID\tBuild ID\tAttached")
	fmt.Fprintf(w, "%s\t%s\t%t\n", result.VersionID, result.BuildID, result.Attached)
	return w.Flush()
}

func printBuildUploadResultMarkdown(result *BuildUploadResult) error {
	fmt.Fprintln(os.Stdout, "| Upload ID | File ID | File Name | File Size |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %d |\n",
		escapeMarkdown(result.UploadID),
		escapeMarkdown(result.FileID),
		escapeMarkdown(result.FileName),
		result.FileSize,
	)
	if len(result.Operations) == 0 {
		return nil
	}
	fmt.Fprintln(os.Stdout, "\n| Method | URL | Length | Offset |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	for _, op := range result.Operations {
		fmt.Fprintf(os.Stdout, "| %s | %s | %d | %d |\n",
			escapeMarkdown(op.Method),
			escapeMarkdown(op.URL),
			op.Length,
			op.Offset,
		)
	}
	return nil
}

func printAppStoreVersionSubmissionMarkdown(result *AppStoreVersionSubmissionResult) error {
	fmt.Fprintln(os.Stdout, "| Submission ID | Created Date |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	createdDate := ""
	if result.CreatedDate != nil {
		createdDate = *result.CreatedDate
	}
	fmt.Fprintf(os.Stdout, "| %s | %s |\n",
		escapeMarkdown(result.SubmissionID),
		escapeMarkdown(createdDate),
	)
	return nil
}

func printAppStoreVersionSubmissionCreateMarkdown(result *AppStoreVersionSubmissionCreateResult) error {
	fmt.Fprintln(os.Stdout, "| Submission ID | Version ID | Build ID | Created Date |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	createdDate := ""
	if result.CreatedDate != nil {
		createdDate = *result.CreatedDate
	}
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s |\n",
		escapeMarkdown(result.SubmissionID),
		escapeMarkdown(result.VersionID),
		escapeMarkdown(result.BuildID),
		escapeMarkdown(createdDate),
	)
	return nil
}

func printAppStoreVersionSubmissionStatusMarkdown(result *AppStoreVersionSubmissionStatusResult) error {
	fmt.Fprintln(os.Stdout, "| Submission ID | Version ID | Version | Platform | State | Created Date |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- |")
	createdDate := ""
	if result.CreatedDate != nil {
		createdDate = *result.CreatedDate
	}
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s | %s |\n",
		escapeMarkdown(result.ID),
		escapeMarkdown(result.VersionID),
		escapeMarkdown(result.VersionString),
		escapeMarkdown(result.Platform),
		escapeMarkdown(result.State),
		escapeMarkdown(createdDate),
	)
	return nil
}

func printAppStoreVersionSubmissionCancelMarkdown(result *AppStoreVersionSubmissionCancelResult) error {
	fmt.Fprintln(os.Stdout, "| Submission ID | Cancelled |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Cancelled,
	)
	return nil
}

func printAppStoreVersionDetailMarkdown(result *AppStoreVersionDetailResult) error {
	fmt.Fprintln(os.Stdout, "| Version ID | Version | Platform | State | Build ID | Build Version | Submission ID |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s | %s | %s |\n",
		escapeMarkdown(result.ID),
		escapeMarkdown(result.VersionString),
		escapeMarkdown(result.Platform),
		escapeMarkdown(result.State),
		escapeMarkdown(result.BuildID),
		escapeMarkdown(result.BuildVersion),
		escapeMarkdown(result.SubmissionID),
	)
	return nil
}

func printAppStoreVersionAttachBuildMarkdown(result *AppStoreVersionAttachBuildResult) error {
	fmt.Fprintln(os.Stdout, "| Version ID | Build ID | Attached |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %t |\n",
		escapeMarkdown(result.VersionID),
		escapeMarkdown(result.BuildID),
		result.Attached,
	)
	return nil
}

func printLocalizationDownloadResultTable(result *LocalizationDownloadResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Locale\tPath")
	for _, file := range result.Files {
		fmt.Fprintf(w, "%s\t%s\n", file.Locale, file.Path)
	}
	return w.Flush()
}

func printLocalizationDownloadResultMarkdown(result *LocalizationDownloadResult) error {
	fmt.Fprintln(os.Stdout, "| Locale | Path |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	for _, file := range result.Files {
		fmt.Fprintf(os.Stdout, "| %s | %s |\n",
			escapeMarkdown(file.Locale),
			escapeMarkdown(file.Path),
		)
	}
	return nil
}

func printLocalizationUploadResultTable(result *LocalizationUploadResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Locale\tAction\tLocalization ID")
	for _, item := range result.Results {
		fmt.Fprintf(w, "%s\t%s\t%s\n", item.Locale, item.Action, item.LocalizationID)
	}
	return w.Flush()
}

func printLocalizationUploadResultMarkdown(result *LocalizationUploadResult) error {
	fmt.Fprintln(os.Stdout, "| Locale | Action | Localization ID |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	for _, item := range result.Results {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
			escapeMarkdown(item.Locale),
			escapeMarkdown(item.Action),
			escapeMarkdown(item.LocalizationID),
		)
	}
	return nil
}

func printBetaTesterDeleteResultTable(result *BetaTesterDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tEmail\tDeleted")
	fmt.Fprintf(w, "%s\t%s\t%t\n",
		result.ID,
		result.Email,
		result.Deleted,
	)
	return w.Flush()
}

func printBetaTesterDeleteResultMarkdown(result *BetaTesterDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Email | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %t |\n",
		escapeMarkdown(result.ID),
		escapeMarkdown(result.Email),
		result.Deleted,
	)
	return nil
}

func printBetaTesterInvitationResultTable(result *BetaTesterInvitationResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Invitation ID\tTester ID\tApp ID\tEmail")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		result.InvitationID,
		result.TesterID,
		result.AppID,
		result.Email,
	)
	return w.Flush()
}

func printBetaTesterInvitationResultMarkdown(result *BetaTesterInvitationResult) error {
	fmt.Fprintln(os.Stdout, "| Invitation ID | Tester ID | App ID | Email |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s |\n",
		escapeMarkdown(result.InvitationID),
		escapeMarkdown(result.TesterID),
		escapeMarkdown(result.AppID),
		escapeMarkdown(result.Email),
	)
	return nil
}
