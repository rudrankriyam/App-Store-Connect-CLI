package asc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestBuildReviewQuery(t *testing.T) {
	query := buildReviewQuery([]ReviewOption{
		WithRating(5),
		WithTerritory("us"),
		WithLimit(25),
		WithReviewSort("-createdDate"),
	})

	values, err := url.ParseQuery(query)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	if got := values.Get("filter[rating]"); got != "5" {
		t.Fatalf("expected filter[rating]=5, got %q", got)
	}

	if got := values.Get("filter[territory]"); got != "US" {
		t.Fatalf("expected filter[territory]=US, got %q", got)
	}

	if got := values.Get("limit"); got != "25" {
		t.Fatalf("expected limit=25, got %q", got)
	}

	if got := values.Get("sort"); got != "-createdDate" {
		t.Fatalf("expected sort=-createdDate, got %q", got)
	}
}

func TestBuildReviewQuery_InvalidRating(t *testing.T) {
	query := buildReviewQuery([]ReviewOption{
		WithRating(9),
	})

	values, err := url.ParseQuery(query)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	if got := values.Get("filter[rating]"); got != "" {
		t.Fatalf("expected empty filter[rating], got %q", got)
	}
}

func TestBuildFeedbackQuery(t *testing.T) {
	query := &feedbackQuery{}
	opts := []FeedbackOption{
		WithFeedbackDeviceModels([]string{"iPhone15,3", " iPhone15,2 "}),
		WithFeedbackOSVersions([]string{"17.2", ""}),
		WithFeedbackAppPlatforms([]string{"ios", "mac_os"}),
		WithFeedbackDevicePlatforms([]string{"tv_os"}),
		WithFeedbackBuildIDs([]string{"build-1"}),
		WithFeedbackBuildPreReleaseVersionIDs([]string{"pre-1"}),
		WithFeedbackTesterIDs([]string{"tester-1"}),
		WithFeedbackLimit(10),
		WithFeedbackSort("-createdDate"),
	}
	for _, opt := range opts {
		opt(query)
	}

	values, err := url.ParseQuery(buildFeedbackQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	if got := values.Get("filter[deviceModel]"); got != "iPhone15,3,iPhone15,2" {
		t.Fatalf("expected filter[deviceModel] to be CSV, got %q", got)
	}
	if got := values.Get("filter[osVersion]"); got != "17.2" {
		t.Fatalf("expected filter[osVersion]=17.2, got %q", got)
	}
	if got := values.Get("limit"); got != "10" {
		t.Fatalf("expected limit=10, got %q", got)
	}
	if got := values.Get("filter[appPlatform]"); got != "IOS,MAC_OS" {
		t.Fatalf("expected filter[appPlatform]=IOS,MAC_OS, got %q", got)
	}
	if got := values.Get("filter[devicePlatform]"); got != "TV_OS" {
		t.Fatalf("expected filter[devicePlatform]=TV_OS, got %q", got)
	}
	if got := values.Get("filter[build]"); got != "build-1" {
		t.Fatalf("expected filter[build]=build-1, got %q", got)
	}
	if got := values.Get("filter[build.preReleaseVersion]"); got != "pre-1" {
		t.Fatalf("expected filter[build.preReleaseVersion]=pre-1, got %q", got)
	}
	if got := values.Get("filter[tester]"); got != "tester-1" {
		t.Fatalf("expected filter[tester]=tester-1, got %q", got)
	}
	if got := values.Get("sort"); got != "-createdDate" {
		t.Fatalf("expected sort=-createdDate, got %q", got)
	}
}

func TestBuildFeedbackQuery_IncludesScreenshots(t *testing.T) {
	query := &feedbackQuery{}
	WithFeedbackIncludeScreenshots()(query)

	values, err := url.ParseQuery(buildFeedbackQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	expected := "createdDate,comment,email,deviceModel,osVersion,appPlatform,devicePlatform,screenshots"
	if got := values.Get("fields[betaFeedbackScreenshotSubmissions]"); got != expected {
		t.Fatalf("expected fields to be %q, got %q", expected, got)
	}
}

func TestBuildCrashQuery(t *testing.T) {
	query := &crashQuery{}
	opts := []CrashOption{
		WithCrashDeviceModels([]string{"iPhone16,1"}),
		WithCrashOSVersions([]string{"18.0"}),
		WithCrashAppPlatforms([]string{"ios"}),
		WithCrashDevicePlatforms([]string{"mac_os"}),
		WithCrashBuildIDs([]string{"build-2"}),
		WithCrashBuildPreReleaseVersionIDs([]string{"pre-2"}),
		WithCrashTesterIDs([]string{"tester-2"}),
		WithCrashLimit(5),
		WithCrashSort("createdDate"),
	}
	for _, opt := range opts {
		opt(query)
	}

	values, err := url.ParseQuery(buildCrashQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	if got := values.Get("filter[deviceModel]"); got != "iPhone16,1" {
		t.Fatalf("expected filter[deviceModel]=iPhone16,1, got %q", got)
	}
	if got := values.Get("filter[osVersion]"); got != "18.0" {
		t.Fatalf("expected filter[osVersion]=18.0, got %q", got)
	}
	if got := values.Get("limit"); got != "5" {
		t.Fatalf("expected limit=5, got %q", got)
	}
	if got := values.Get("filter[appPlatform]"); got != "IOS" {
		t.Fatalf("expected filter[appPlatform]=IOS, got %q", got)
	}
	if got := values.Get("filter[devicePlatform]"); got != "MAC_OS" {
		t.Fatalf("expected filter[devicePlatform]=MAC_OS, got %q", got)
	}
	if got := values.Get("filter[build]"); got != "build-2" {
		t.Fatalf("expected filter[build]=build-2, got %q", got)
	}
	if got := values.Get("filter[build.preReleaseVersion]"); got != "pre-2" {
		t.Fatalf("expected filter[build.preReleaseVersion]=pre-2, got %q", got)
	}
	if got := values.Get("filter[tester]"); got != "tester-2" {
		t.Fatalf("expected filter[tester]=tester-2, got %q", got)
	}
	if got := values.Get("sort"); got != "createdDate" {
		t.Fatalf("expected sort=createdDate, got %q", got)
	}
}

func TestBuildBetaGroupsQuery(t *testing.T) {
	query := &betaGroupsQuery{}
	WithBetaGroupsLimit(10)(query)

	values, err := url.ParseQuery(buildBetaGroupsQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("limit"); got != "10" {
		t.Fatalf("expected limit=10, got %q", got)
	}
}

func TestBuildBetaTestersQuery(t *testing.T) {
	query := &betaTestersQuery{}
	opts := []BetaTestersOption{
		WithBetaTestersLimit(25),
		WithBetaTestersEmail("tester@example.com"),
		WithBetaTestersGroupIDs([]string{"group-1", " group-2 "}),
	}
	for _, opt := range opts {
		opt(query)
	}

	values, err := url.ParseQuery(buildBetaTestersQuery("APP_ID", query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("filter[apps]"); got != "APP_ID" {
		t.Fatalf("expected filter[apps]=APP_ID, got %q", got)
	}
	if got := values.Get("filter[email]"); got != "tester@example.com" {
		t.Fatalf("expected filter[email]=tester@example.com, got %q", got)
	}
	if got := values.Get("filter[betaGroups]"); got != "group-1,group-2" {
		t.Fatalf("expected filter[betaGroups]=group-1,group-2, got %q", got)
	}
	if got := values.Get("limit"); got != "25" {
		t.Fatalf("expected limit=25, got %q", got)
	}
}

func TestBuildAppStoreVersionsQuery(t *testing.T) {
	query := &appStoreVersionsQuery{}
	opts := []AppStoreVersionsOption{
		WithAppStoreVersionsLimit(20),
		WithAppStoreVersionsPlatforms([]string{"ios", "MAC_OS"}),
		WithAppStoreVersionsVersionStrings([]string{"1.0.0", "1.1.0"}),
		WithAppStoreVersionsStates([]string{"ready_for_review"}),
	}
	for _, opt := range opts {
		opt(query)
	}

	values, err := url.ParseQuery(buildAppStoreVersionsQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("filter[platform]"); got != "IOS,MAC_OS" {
		t.Fatalf("expected filter[platform]=IOS,MAC_OS, got %q", got)
	}
	if got := values.Get("filter[versionString]"); got != "1.0.0,1.1.0" {
		t.Fatalf("expected filter[versionString]=1.0.0,1.1.0, got %q", got)
	}
	if got := values.Get("filter[appStoreState]"); got != "READY_FOR_REVIEW" {
		t.Fatalf("expected filter[appStoreState]=READY_FOR_REVIEW, got %q", got)
	}
	if got := values.Get("limit"); got != "20" {
		t.Fatalf("expected limit=20, got %q", got)
	}
}

func TestBuildPreReleaseVersionsQuery(t *testing.T) {
	query := &preReleaseVersionsQuery{}
	opts := []PreReleaseVersionsOption{
		WithPreReleaseVersionsLimit(15),
		WithPreReleaseVersionsPlatform(" ios, MAC_OS "),
		WithPreReleaseVersionsVersion("1.0.0, 1.1.0"),
	}
	for _, opt := range opts {
		opt(query)
	}

	values, err := url.ParseQuery(buildPreReleaseVersionsQuery("APP_ID", query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("filter[app]"); got != "APP_ID" {
		t.Fatalf("expected filter[app]=APP_ID, got %q", got)
	}
	if got := values.Get("filter[platform]"); got != "IOS,MAC_OS" {
		t.Fatalf("expected filter[platform]=IOS,MAC_OS, got %q", got)
	}
	if got := values.Get("filter[version]"); got != "1.0.0,1.1.0" {
		t.Fatalf("expected filter[version]=1.0.0,1.1.0, got %q", got)
	}
	if got := values.Get("limit"); got != "15" {
		t.Fatalf("expected limit=15, got %q", got)
	}
}

func TestBuildAppStoreVersionLocalizationsQuery(t *testing.T) {
	query := &appStoreVersionLocalizationsQuery{}
	opts := []AppStoreVersionLocalizationsOption{
		WithAppStoreVersionLocalizationsLimit(10),
		WithAppStoreVersionLocalizationLocales([]string{"en-US", "ja"}),
	}
	for _, opt := range opts {
		opt(query)
	}

	values, err := url.ParseQuery(buildAppStoreVersionLocalizationsQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("filter[locale]"); got != "en-US,ja" {
		t.Fatalf("expected filter[locale]=en-US,ja, got %q", got)
	}
	if got := values.Get("limit"); got != "10" {
		t.Fatalf("expected limit=10, got %q", got)
	}
}

func TestBuildAppInfoLocalizationsQuery(t *testing.T) {
	query := &appInfoLocalizationsQuery{}
	opts := []AppInfoLocalizationsOption{
		WithAppInfoLocalizationsLimit(5),
		WithAppInfoLocalizationLocales([]string{"en-US"}),
	}
	for _, opt := range opts {
		opt(query)
	}

	values, err := url.ParseQuery(buildAppInfoLocalizationsQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("filter[locale]"); got != "en-US" {
		t.Fatalf("expected filter[locale]=en-US, got %q", got)
	}
	if got := values.Get("limit"); got != "5" {
		t.Fatalf("expected limit=5, got %q", got)
	}
}

func TestBuildRequestBody(t *testing.T) {
	body, err := BuildRequestBody(map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("BuildRequestBody() error: %v", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		t.Fatalf("read body error: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if parsed["hello"] != "world" {
		t.Fatalf("expected hello=world, got %q", parsed["hello"])
	}
}

func TestParseError(t *testing.T) {
	payload := []byte(`{"errors":[{"code":"FORBIDDEN","title":"Forbidden","detail":"not allowed"}]}`)
	err := ParseError(payload)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Forbidden") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestBuildAppsQuery(t *testing.T) {
	query := &appsQuery{}
	opts := []AppsOption{
		WithAppsLimit(10),
		WithAppsNextURL("https://api.appstoreconnect.apple.com/v1/apps?cursor=abc123"),
		WithAppsSort("-name"),
		WithAppsBundleIDs([]string{"com.example.app", " ", "com.example.other"}),
		WithAppsNames([]string{"Demo", " "}),
		WithAppsSKUs([]string{"SKU1", "SKU2"}),
	}
	for _, opt := range opts {
		opt(query)
	}

	if query.limit != 10 {
		t.Fatalf("expected limit=10, got %d", query.limit)
	}
	if query.nextURL != "https://api.appstoreconnect.apple.com/v1/apps?cursor=abc123" {
		t.Fatalf("expected nextURL, got %q", query.nextURL)
	}
	if query.sort != "-name" {
		t.Fatalf("expected sort=-name, got %q", query.sort)
	}
	if len(query.bundleIDs) != 2 || query.bundleIDs[0] != "com.example.app" || query.bundleIDs[1] != "com.example.other" {
		t.Fatalf("expected bundleIDs to be normalized, got %v", query.bundleIDs)
	}
	if len(query.names) != 1 || query.names[0] != "Demo" {
		t.Fatalf("expected names to be normalized, got %v", query.names)
	}
	if len(query.skus) != 2 || query.skus[0] != "SKU1" || query.skus[1] != "SKU2" {
		t.Fatalf("expected skus to be set, got %v", query.skus)
	}
}

func TestBuildBuildsQuery(t *testing.T) {
	query := &buildsQuery{}
	opts := []BuildsOption{
		WithBuildsLimit(25),
		WithBuildsNextURL("https://api.appstoreconnect.apple.com/v1/apps/123/builds?cursor=abc"),
		WithBuildsSort("-uploadedDate"),
	}
	for _, opt := range opts {
		opt(query)
	}

	if query.limit != 25 {
		t.Fatalf("expected limit=25, got %d", query.limit)
	}
	if query.nextURL != "https://api.appstoreconnect.apple.com/v1/apps/123/builds?cursor=abc" {
		t.Fatalf("expected nextURL to be set, got %q", query.nextURL)
	}
	if query.sort != "-uploadedDate" {
		t.Fatalf("expected sort=-uploadedDate, got %q", query.sort)
	}
}

func TestBuildUploadCreateRequest_JSON(t *testing.T) {
	req := BuildUploadCreateRequest{
		Data: BuildUploadCreateData{
			Type: ResourceTypeBuildUploads,
			Attributes: BuildUploadAttributes{
				CFBundleShortVersionString: "1.0.0",
				CFBundleVersion:            "123",
				Platform:                   PlatformIOS,
			},
			Relationships: &BuildUploadRelationships{
				App: &Relationship{
					Data: ResourceData{
						Type: ResourceTypeApps,
						ID:   "APP_ID_123",
					},
				},
			},
		},
	}

	body, err := BuildRequestBody(req)
	if err != nil {
		t.Fatalf("BuildRequestBody() error: %v", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		t.Fatalf("read body error: %v", err)
	}

	// Unmarshal and verify structure
	var parsed struct {
		Data struct {
			Type       string `json:"type"`
			Attributes struct {
				CFBundleShortVersionString string `json:"cfBundleShortVersionString"`
				CFBundleVersion            string `json:"cfBundleVersion"`
				Platform                   string `json:"platform"`
			} `json:"attributes"`
			Relationships struct {
				App struct {
					Data struct {
						Type string `json:"type"`
						ID   string `json:"id"`
					} `json:"data"`
				} `json:"app"`
			} `json:"relationships"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if parsed.Data.Type != "buildUploads" {
		t.Fatalf("expected type=buildUploads, got %q", parsed.Data.Type)
	}
	if parsed.Data.Attributes.CFBundleShortVersionString != "1.0.0" {
		t.Fatalf("expected cfBundleShortVersionString=1.0.0, got %q", parsed.Data.Attributes.CFBundleShortVersionString)
	}
	if parsed.Data.Attributes.CFBundleVersion != "123" {
		t.Fatalf("expected cfBundleVersion=123, got %q", parsed.Data.Attributes.CFBundleVersion)
	}
	if parsed.Data.Attributes.Platform != "IOS" {
		t.Fatalf("expected platform=IOS, got %q", parsed.Data.Attributes.Platform)
	}
	if parsed.Data.Relationships.App.Data.Type != "apps" {
		t.Fatalf("expected app type=apps, got %q", parsed.Data.Relationships.App.Data.Type)
	}
	if parsed.Data.Relationships.App.Data.ID != "APP_ID_123" {
		t.Fatalf("expected app id=APP_ID_123, got %q", parsed.Data.Relationships.App.Data.ID)
	}
}

func TestBuildUploadFileCreateRequest_JSON(t *testing.T) {
	req := BuildUploadFileCreateRequest{
		Data: BuildUploadFileCreateData{
			Type: ResourceTypeBuildUploadFiles,
			Attributes: BuildUploadFileAttributes{
				FileName:  "app.ipa",
				FileSize:  1024000,
				UTI:       UTIIPA,
				AssetType: AssetTypeAsset,
			},
			Relationships: &BuildUploadFileRelationships{
				BuildUpload: &Relationship{
					Data: ResourceData{
						Type: ResourceTypeBuildUploads,
						ID:   "UPLOAD_ID_123",
					},
				},
			},
		},
	}

	body, err := BuildRequestBody(req)
	if err != nil {
		t.Fatalf("BuildRequestBody() error: %v", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		t.Fatalf("read body error: %v", err)
	}

	// Unmarshal and verify structure
	var parsed struct {
		Data struct {
			Type       string `json:"type"`
			ID         string `json:"id,omitempty"`
			Attributes struct {
				FileName  string `json:"fileName"`
				FileSize  int64  `json:"fileSize"`
				UTI       string `json:"uti"`
				AssetType string `json:"assetType"`
			} `json:"attributes"`
			Relationships struct {
				BuildUpload struct {
					Data struct {
						Type string `json:"type"`
						ID   string `json:"id"`
					} `json:"data"`
				} `json:"buildUpload"`
			} `json:"relationships"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if parsed.Data.Type != "buildUploadFiles" {
		t.Fatalf("expected type=buildUploadFiles, got %q", parsed.Data.Type)
	}
	if parsed.Data.Attributes.FileName != "app.ipa" {
		t.Fatalf("expected fileName=app.ipa, got %q", parsed.Data.Attributes.FileName)
	}
	if parsed.Data.Attributes.FileSize != 1024000 {
		t.Fatalf("expected fileSize=1024000, got %d", parsed.Data.Attributes.FileSize)
	}
	if parsed.Data.Attributes.UTI != "com.apple.ipa" {
		t.Fatalf("expected uti=com.apple.ipa, got %q", parsed.Data.Attributes.UTI)
	}
	if parsed.Data.Attributes.AssetType != "ASSET" {
		t.Fatalf("expected assetType=ASSET, got %q", parsed.Data.Attributes.AssetType)
	}
	if parsed.Data.Relationships.BuildUpload.Data.Type != "buildUploads" {
		t.Fatalf("expected buildUpload type=buildUploads, got %q", parsed.Data.Relationships.BuildUpload.Data.Type)
	}
	if parsed.Data.Relationships.BuildUpload.Data.ID != "UPLOAD_ID_123" {
		t.Fatalf("expected buildUpload id=UPLOAD_ID_123, got %q", parsed.Data.Relationships.BuildUpload.Data.ID)
	}
}

func TestBuildUploadFileUpdateRequest_JSON(t *testing.T) {
	uploaded := true
	req := BuildUploadFileUpdateRequest{
		Data: BuildUploadFileUpdateData{
			Type: ResourceTypeBuildUploadFiles,
			ID:   "FILE_ID_123",
			Attributes: &BuildUploadFileUpdateAttributes{
				SourceFileChecksums: &Checksums{
					File: &Checksum{
						Hash:      "abc123def456",
						Algorithm: ChecksumAlgorithmSHA256,
					},
				},
				Uploaded: &uploaded,
			},
		},
	}

	body, err := BuildRequestBody(req)
	if err != nil {
		t.Fatalf("BuildRequestBody() error: %v", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		t.Fatalf("read body error: %v", err)
	}

	// Unmarshal and verify structure
	var parsed struct {
		Data struct {
			Type       string `json:"type"`
			ID         string `json:"id"`
			Attributes struct {
				SourceFileChecksums struct {
					File struct {
						Hash      string `json:"hash"`
						Algorithm string `json:"algorithm"`
					} `json:"file"`
				} `json:"sourceFileChecksums"`
				Uploaded bool `json:"uploaded"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if parsed.Data.ID != "FILE_ID_123" {
		t.Fatalf("expected id=FILE_ID_123, got %q", parsed.Data.ID)
	}
	if !parsed.Data.Attributes.Uploaded {
		t.Fatalf("expected uploaded=true, got false")
	}
	if parsed.Data.Attributes.SourceFileChecksums.File.Hash != "abc123def456" {
		t.Fatalf("expected checksum hash=abc123def456, got %q", parsed.Data.Attributes.SourceFileChecksums.File.Hash)
	}
	if parsed.Data.Attributes.SourceFileChecksums.File.Algorithm != "SHA_256" {
		t.Fatalf("expected algorithm=SHA_256, got %q", parsed.Data.Attributes.SourceFileChecksums.File.Algorithm)
	}
}

func TestAppScreenshotSetCreateRequest_JSON(t *testing.T) {
	req := AppScreenshotSetCreateRequest{
		Data: AppScreenshotSetCreateData{
			Type:       ResourceTypeAppScreenshotSets,
			Attributes: AppScreenshotSetAttributes{ScreenshotDisplayType: "APP_IPHONE_65"},
			Relationships: &AppScreenshotSetRelationships{
				AppStoreVersionLocalization: &Relationship{
					Data: ResourceData{
						Type: ResourceTypeAppStoreVersionLocalizations,
						ID:   "LOC_ID_123",
					},
				},
			},
		},
	}

	body, err := BuildRequestBody(req)
	if err != nil {
		t.Fatalf("BuildRequestBody() error: %v", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		t.Fatalf("read body error: %v", err)
	}

	var parsed struct {
		Data struct {
			Type       string `json:"type"`
			Attributes struct {
				ScreenshotDisplayType string `json:"screenshotDisplayType"`
			} `json:"attributes"`
			Relationships struct {
				AppStoreVersionLocalization struct {
					Data struct {
						Type string `json:"type"`
						ID   string `json:"id"`
					} `json:"data"`
				} `json:"appStoreVersionLocalization"`
			} `json:"relationships"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if parsed.Data.Type != "appScreenshotSets" {
		t.Fatalf("expected type=appScreenshotSets, got %q", parsed.Data.Type)
	}
	if parsed.Data.Attributes.ScreenshotDisplayType != "APP_IPHONE_65" {
		t.Fatalf("expected screenshotDisplayType=APP_IPHONE_65, got %q", parsed.Data.Attributes.ScreenshotDisplayType)
	}
	if parsed.Data.Relationships.AppStoreVersionLocalization.Data.Type != "appStoreVersionLocalizations" {
		t.Fatalf("expected relationship type=appStoreVersionLocalizations, got %q", parsed.Data.Relationships.AppStoreVersionLocalization.Data.Type)
	}
	if parsed.Data.Relationships.AppStoreVersionLocalization.Data.ID != "LOC_ID_123" {
		t.Fatalf("expected relationship id=LOC_ID_123, got %q", parsed.Data.Relationships.AppStoreVersionLocalization.Data.ID)
	}
}

func TestAppScreenshotCreateRequest_JSON(t *testing.T) {
	req := AppScreenshotCreateRequest{
		Data: AppScreenshotCreateData{
			Type: ResourceTypeAppScreenshots,
			Attributes: AppScreenshotAttributes{
				FileName: "shot.png",
				FileSize: 1024,
			},
			Relationships: &AppScreenshotRelationships{
				AppScreenshotSet: &Relationship{
					Data: ResourceData{
						Type: ResourceTypeAppScreenshotSets,
						ID:   "SET_ID_123",
					},
				},
			},
		},
	}

	body, err := BuildRequestBody(req)
	if err != nil {
		t.Fatalf("BuildRequestBody() error: %v", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		t.Fatalf("read body error: %v", err)
	}

	var parsed struct {
		Data struct {
			Type       string `json:"type"`
			Attributes struct {
				FileName string `json:"fileName"`
				FileSize int64  `json:"fileSize"`
			} `json:"attributes"`
			Relationships struct {
				AppScreenshotSet struct {
					Data struct {
						Type string `json:"type"`
						ID   string `json:"id"`
					} `json:"data"`
				} `json:"appScreenshotSet"`
			} `json:"relationships"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if parsed.Data.Type != "appScreenshots" {
		t.Fatalf("expected type=appScreenshots, got %q", parsed.Data.Type)
	}
	if parsed.Data.Attributes.FileName != "shot.png" {
		t.Fatalf("expected fileName=shot.png, got %q", parsed.Data.Attributes.FileName)
	}
	if parsed.Data.Attributes.FileSize != 1024 {
		t.Fatalf("expected fileSize=1024, got %d", parsed.Data.Attributes.FileSize)
	}
	if parsed.Data.Relationships.AppScreenshotSet.Data.Type != "appScreenshotSets" {
		t.Fatalf("expected appScreenshotSet type=appScreenshotSets, got %q", parsed.Data.Relationships.AppScreenshotSet.Data.Type)
	}
	if parsed.Data.Relationships.AppScreenshotSet.Data.ID != "SET_ID_123" {
		t.Fatalf("expected appScreenshotSet id=SET_ID_123, got %q", parsed.Data.Relationships.AppScreenshotSet.Data.ID)
	}
}

func TestAppScreenshotUpdateRequest_JSON(t *testing.T) {
	uploaded := true
	req := AppScreenshotUpdateRequest{
		Data: AppScreenshotUpdateData{
			Type: ResourceTypeAppScreenshots,
			ID:   "SCREENSHOT_ID_123",
			Attributes: &AppScreenshotUpdateAttributes{
				Uploaded: &uploaded,
				SourceFileChecksum: &Checksum{
					Hash:      "abc123",
					Algorithm: ChecksumAlgorithmMD5,
				},
			},
		},
	}

	body, err := BuildRequestBody(req)
	if err != nil {
		t.Fatalf("BuildRequestBody() error: %v", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		t.Fatalf("read body error: %v", err)
	}

	var parsed struct {
		Data struct {
			Type       string `json:"type"`
			ID         string `json:"id"`
			Attributes struct {
				Uploaded           bool `json:"uploaded"`
				SourceFileChecksum struct {
					Hash      string `json:"hash"`
					Algorithm string `json:"algorithm"`
				} `json:"sourceFileChecksum"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if parsed.Data.Type != "appScreenshots" {
		t.Fatalf("expected type=appScreenshots, got %q", parsed.Data.Type)
	}
	if parsed.Data.ID != "SCREENSHOT_ID_123" {
		t.Fatalf("expected id=SCREENSHOT_ID_123, got %q", parsed.Data.ID)
	}
	if !parsed.Data.Attributes.Uploaded {
		t.Fatalf("expected uploaded=true")
	}
	if parsed.Data.Attributes.SourceFileChecksum.Algorithm != "MD5" {
		t.Fatalf("expected checksum algorithm=MD5, got %q", parsed.Data.Attributes.SourceFileChecksum.Algorithm)
	}
}

func TestAppPreviewSetCreateRequest_JSON(t *testing.T) {
	req := AppPreviewSetCreateRequest{
		Data: AppPreviewSetCreateData{
			Type:       ResourceTypeAppPreviewSets,
			Attributes: AppPreviewSetAttributes{PreviewType: "IPHONE_65"},
			Relationships: &AppPreviewSetRelationships{
				AppStoreVersionLocalization: &Relationship{
					Data: ResourceData{
						Type: ResourceTypeAppStoreVersionLocalizations,
						ID:   "LOC_ID_123",
					},
				},
			},
		},
	}

	body, err := BuildRequestBody(req)
	if err != nil {
		t.Fatalf("BuildRequestBody() error: %v", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		t.Fatalf("read body error: %v", err)
	}

	var parsed struct {
		Data struct {
			Type       string `json:"type"`
			Attributes struct {
				PreviewType string `json:"previewType"`
			} `json:"attributes"`
			Relationships struct {
				AppStoreVersionLocalization struct {
					Data struct {
						Type string `json:"type"`
						ID   string `json:"id"`
					} `json:"data"`
				} `json:"appStoreVersionLocalization"`
			} `json:"relationships"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if parsed.Data.Type != "appPreviewSets" {
		t.Fatalf("expected type=appPreviewSets, got %q", parsed.Data.Type)
	}
	if parsed.Data.Attributes.PreviewType != "IPHONE_65" {
		t.Fatalf("expected previewType=IPHONE_65, got %q", parsed.Data.Attributes.PreviewType)
	}
	if parsed.Data.Relationships.AppStoreVersionLocalization.Data.Type != "appStoreVersionLocalizations" {
		t.Fatalf("expected relationship type=appStoreVersionLocalizations, got %q", parsed.Data.Relationships.AppStoreVersionLocalization.Data.Type)
	}
	if parsed.Data.Relationships.AppStoreVersionLocalization.Data.ID != "LOC_ID_123" {
		t.Fatalf("expected relationship id=LOC_ID_123, got %q", parsed.Data.Relationships.AppStoreVersionLocalization.Data.ID)
	}
}

func TestAppPreviewCreateRequest_JSON(t *testing.T) {
	req := AppPreviewCreateRequest{
		Data: AppPreviewCreateData{
			Type: ResourceTypeAppPreviews,
			Attributes: AppPreviewAttributes{
				FileName: "preview.mov",
				FileSize: 2048,
				MimeType: "video/quicktime",
			},
			Relationships: &AppPreviewRelationships{
				AppPreviewSet: &Relationship{
					Data: ResourceData{
						Type: ResourceTypeAppPreviewSets,
						ID:   "SET_ID_123",
					},
				},
			},
		},
	}

	body, err := BuildRequestBody(req)
	if err != nil {
		t.Fatalf("BuildRequestBody() error: %v", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		t.Fatalf("read body error: %v", err)
	}

	var parsed struct {
		Data struct {
			Type       string `json:"type"`
			Attributes struct {
				FileName string `json:"fileName"`
				FileSize int64  `json:"fileSize"`
				MimeType string `json:"mimeType"`
			} `json:"attributes"`
			Relationships struct {
				AppPreviewSet struct {
					Data struct {
						Type string `json:"type"`
						ID   string `json:"id"`
					} `json:"data"`
				} `json:"appPreviewSet"`
			} `json:"relationships"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if parsed.Data.Type != "appPreviews" {
		t.Fatalf("expected type=appPreviews, got %q", parsed.Data.Type)
	}
	if parsed.Data.Attributes.FileName != "preview.mov" {
		t.Fatalf("expected fileName=preview.mov, got %q", parsed.Data.Attributes.FileName)
	}
	if parsed.Data.Attributes.FileSize != 2048 {
		t.Fatalf("expected fileSize=2048, got %d", parsed.Data.Attributes.FileSize)
	}
	if parsed.Data.Attributes.MimeType != "video/quicktime" {
		t.Fatalf("expected mimeType=video/quicktime, got %q", parsed.Data.Attributes.MimeType)
	}
	if parsed.Data.Relationships.AppPreviewSet.Data.Type != "appPreviewSets" {
		t.Fatalf("expected appPreviewSet type=appPreviewSets, got %q", parsed.Data.Relationships.AppPreviewSet.Data.Type)
	}
	if parsed.Data.Relationships.AppPreviewSet.Data.ID != "SET_ID_123" {
		t.Fatalf("expected appPreviewSet id=SET_ID_123, got %q", parsed.Data.Relationships.AppPreviewSet.Data.ID)
	}
}

func TestAppPreviewUpdateRequest_JSON(t *testing.T) {
	uploaded := true
	req := AppPreviewUpdateRequest{
		Data: AppPreviewUpdateData{
			Type: ResourceTypeAppPreviews,
			ID:   "PREVIEW_ID_123",
			Attributes: &AppPreviewUpdateAttributes{
				Uploaded: &uploaded,
				SourceFileChecksum: &Checksum{
					Hash:      "def456",
					Algorithm: ChecksumAlgorithmMD5,
				},
			},
		},
	}

	body, err := BuildRequestBody(req)
	if err != nil {
		t.Fatalf("BuildRequestBody() error: %v", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		t.Fatalf("read body error: %v", err)
	}

	var parsed struct {
		Data struct {
			Type       string `json:"type"`
			ID         string `json:"id"`
			Attributes struct {
				Uploaded           bool `json:"uploaded"`
				SourceFileChecksum struct {
					Hash      string `json:"hash"`
					Algorithm string `json:"algorithm"`
				} `json:"sourceFileChecksum"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if parsed.Data.Type != "appPreviews" {
		t.Fatalf("expected type=appPreviews, got %q", parsed.Data.Type)
	}
	if parsed.Data.ID != "PREVIEW_ID_123" {
		t.Fatalf("expected id=PREVIEW_ID_123, got %q", parsed.Data.ID)
	}
	if !parsed.Data.Attributes.Uploaded {
		t.Fatalf("expected uploaded=true")
	}
	if parsed.Data.Attributes.SourceFileChecksum.Algorithm != "MD5" {
		t.Fatalf("expected checksum algorithm=MD5, got %q", parsed.Data.Attributes.SourceFileChecksum.Algorithm)
	}
}

func TestAppStoreVersionSubmissionCreateRequest_JSON(t *testing.T) {
	req := AppStoreVersionSubmissionCreateRequest{
		Data: AppStoreVersionSubmissionCreateData{
			Type: ResourceTypeAppStoreVersionSubmissions,
			Relationships: &AppStoreVersionSubmissionRelationships{
				AppStoreVersion: &Relationship{
					Data: ResourceData{
						Type: ResourceTypeAppStoreVersions,
						ID:   "VERSION_ID_123",
					},
				},
			},
		},
	}

	body, err := BuildRequestBody(req)
	if err != nil {
		t.Fatalf("BuildRequestBody() error: %v", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		t.Fatalf("read body error: %v", err)
	}

	// Unmarshal and verify structure
	var parsed struct {
		Data struct {
			Type          string `json:"type"`
			Relationships struct {
				AppStoreVersion struct {
					Data struct {
						Type string `json:"type"`
						ID   string `json:"id"`
					} `json:"data"`
				} `json:"appStoreVersion"`
			} `json:"relationships"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if parsed.Data.Type != "appStoreVersionSubmissions" {
		t.Fatalf("expected type=appStoreVersionSubmissions, got %q", parsed.Data.Type)
	}
	if parsed.Data.Relationships.AppStoreVersion.Data.Type != "appStoreVersions" {
		t.Fatalf("expected version type=appStoreVersions, got %q", parsed.Data.Relationships.AppStoreVersion.Data.Type)
	}
	if parsed.Data.Relationships.AppStoreVersion.Data.ID != "VERSION_ID_123" {
		t.Fatalf("expected version id=VERSION_ID_123, got %q", parsed.Data.Relationships.AppStoreVersion.Data.ID)
	}
}

func TestWithRetry_ZeroRetries(t *testing.T) {
	callCount := 0
	wantErr := fmt.Errorf("transient error")

	_, err := WithRetry(context.Background(), func() (string, error) {
		callCount++
		return "", wantErr
	}, RetryOptions{MaxRetries: 0})

	// Should fail immediately without retries
	if callCount != 1 {
		t.Fatalf("expected 1 call (no retries), got %d", callCount)
	}
	if err != wantErr {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}

func TestWithRetry_NegativeRetriesUsesDefault(t *testing.T) {
	callCount := 0

	WithRetry(context.Background(), func() (string, error) {
		callCount++
		if callCount < DefaultMaxRetries+1 {
			return "", &RetryableError{RetryAfter: time.Millisecond}
		}
		return "success", nil
	}, RetryOptions{MaxRetries: -1})

	// Should use default retries (3)
	if callCount != DefaultMaxRetries+1 {
		t.Fatalf("expected %d calls (default retries), got %d", DefaultMaxRetries+1, callCount)
	}
}

func TestWithRetry_AttemptCountInErrorMessage(t *testing.T) {
	const maxRetries = 2
	callCount := 0

	_, err := WithRetry(context.Background(), func() (string, error) {
		callCount++
		return "", &RetryableError{RetryAfter: time.Millisecond}
	}, RetryOptions{MaxRetries: maxRetries})

	// Should have exhausted all retries
	if callCount != maxRetries+1 {
		t.Fatalf("expected %d calls, got %d", maxRetries+1, callCount)
	}

	// Error message should report the correct number of retries
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errMsg := err.Error()
	// With MaxRetries=2, we try 3 times total (initial + 2 retries)
	// Error message should say "after 3 retries"
	if !strings.Contains(errMsg, "after 3 retries") {
		t.Fatalf("error message should mention 'after 3 retries', got: %s", errMsg)
	}
}

func TestWithRetry_NonRetryableErrorFailsFast(t *testing.T) {
	callCount := 0
	wantErr := fmt.Errorf("non-retryable error")

	_, err := WithRetry(context.Background(), func() (string, error) {
		callCount++
		return "", wantErr
	}, RetryOptions{MaxRetries: 3, BaseDelay: time.Millisecond})

	// Should fail immediately without retries
	if callCount != 1 {
		t.Fatalf("expected 1 call (no retries for non-retryable error), got %d", callCount)
	}
	if err != wantErr {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}

func TestWithRetry_SuccessOnFirstTry(t *testing.T) {
	callCount := 0

	result, err := WithRetry(context.Background(), func() (string, error) {
		callCount++
		return "success", nil
	}, RetryOptions{MaxRetries: 3})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != "success" {
		t.Fatalf("expected 'success', got %q", result)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 call, got %d", callCount)
	}
}

func TestPaginateAll_CiBuildRuns_ManyPages(t *testing.T) {
	const totalPages = 20
	const perPage = 50

	makePage := func(page int) *CiBuildRunsResponse {
		data := make([]CiBuildRunResource, 0, perPage)
		for i := 0; i < perPage; i++ {
			data = append(data, CiBuildRunResource{
				Type: ResourceTypeCiBuildRuns,
				ID:   fmt.Sprintf("run-%d-%d", page, i),
			})
		}
		links := Links{}
		if page < totalPages {
			links.Next = fmt.Sprintf("page=%d", page+1)
		}
		return &CiBuildRunsResponse{
			Data:  data,
			Links: links,
		}
	}

	firstPage := makePage(1)
	response, err := PaginateAll(context.Background(), firstPage, func(ctx context.Context, nextURL string) (PaginatedResponse, error) {
		pageStr := strings.TrimPrefix(nextURL, "page=")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, fmt.Errorf("invalid next URL %q", nextURL)
		}
		return makePage(page), nil
	})
	if err != nil {
		t.Fatalf("PaginateAll() error: %v", err)
	}

	buildRuns, ok := response.(*CiBuildRunsResponse)
	if !ok {
		t.Fatalf("expected CiBuildRunsResponse, got %T", response)
	}
	expected := totalPages * perPage
	if len(buildRuns.Data) != expected {
		t.Fatalf("expected %d build runs, got %d", expected, len(buildRuns.Data))
	}
	if buildRuns.Links.Next != "" {
		t.Fatalf("expected next link to be cleared, got %q", buildRuns.Links.Next)
	}
}
