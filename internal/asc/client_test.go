package asc

import (
	"bytes"
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

func TestBuildRequestBody(t *testing.T) {
	body, err := BuildRequestBody(map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("BuildRequestBody() error: %v", err)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		t.Fatalf("read body error: %v", err)
	}

	if !strings.Contains(buf.String(), `"hello":"world"`) {
		t.Fatalf("unexpected body: %s", buf.String())
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
		WithAppsSort("-uploadedDate"),
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
	if query.sort != "-uploadedDate" {
		t.Fatalf("expected sort=-uploadedDate, got %q", query.sort)
	}
}

func TestBuildBuildsQuery(t *testing.T) {
	query := &buildsQuery{}
	opts := []BuildsOption{
		WithBuildsLimit(25),
		WithBuildsApp("1234567890"),
		WithBuildsSort("uploadedDate"),
	}
	for _, opt := range opts {
		opt(query)
	}

	if query.limit != 25 {
		t.Fatalf("expected limit=25, got %d", query.limit)
	}
	if query.appID != "1234567890" {
		t.Fatalf("expected appID=1234567890, got %q", query.appID)
	}
	if query.sort != "uploadedDate" {
		t.Fatalf("expected sort=uploadedDate, got %q", query.sort)
	}
}
