package asc

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
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
