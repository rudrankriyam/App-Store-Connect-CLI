package asc

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func newTestClient(t *testing.T, check func(*http.Request), response *http.Response) *Client {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error: %v", err)
	}

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if check != nil {
			check(req)
		}
		return response, nil
	})

	return &Client{
		httpClient: &http.Client{Transport: transport},
		keyID:      "KEY123",
		issuerID:   "ISS456",
		privateKey: key,
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func assertAuthorized(t *testing.T, req *http.Request) {
	t.Helper()

	auth := req.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		t.Fatalf("expected Authorization bearer token, got %q", auth)
	}
}

func TestGetApps_RateLimitedIncludesRetryAfter(t *testing.T) {
	t.Setenv("ASC_MAX_RETRIES", "0")

	response := jsonResponse(http.StatusTooManyRequests, `{"errors":[{"title":"Rate limit","detail":"Too many requests"}]}`)
	response.Header.Set("Retry-After", "120")
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		assertAuthorized(t, req)
	}, response)

	_, err := client.GetApps(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsRetryable(err) {
		t.Fatalf("expected retryable error, got %v", err)
	}
	if got := GetRetryAfter(err); got != 2*time.Minute {
		t.Fatalf("expected retry-after 2m, got %s", got)
	}
	if !strings.Contains(err.Error(), "retry after") {
		t.Fatalf("expected retry-after in error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "status 429") {
		t.Fatalf("expected status 429 in error, got %q", err.Error())
	}
}

func TestGetApps_WithSortAndLimit(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"apps","id":"1","attributes":{"name":"Demo","bundleId":"com.example.demo","sku":"SKU1"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/apps" {
			t.Fatalf("expected path /v1/apps, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("sort") != "-name" {
			t.Fatalf("expected sort=-name, got %q", values.Get("sort"))
		}
		if values.Get("limit") != "10" {
			t.Fatalf("expected limit=10, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetApps(context.Background(), WithAppsLimit(10), WithAppsSort("-name")); err != nil {
		t.Fatalf("GetApps() error: %v", err)
	}
}

func TestGetApps_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/apps?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetApps(context.Background(), WithAppsLimit(5), WithAppsSort("name"), WithAppsNextURL(next)); err != nil {
		t.Fatalf("GetApps() error: %v", err)
	}
}

func TestGetApps_WithFilters(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"apps","id":"1","attributes":{"name":"Demo","bundleId":"com.example.demo","sku":"SKU1"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/apps" {
			t.Fatalf("expected path /v1/apps, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[bundleId]") != "com.example.demo,com.example.other" {
			t.Fatalf("expected filter[bundleId] to be set, got %q", values.Get("filter[bundleId]"))
		}
		if values.Get("filter[name]") != "Demo App" {
			t.Fatalf("expected filter[name]=Demo App, got %q", values.Get("filter[name]"))
		}
		if values.Get("filter[sku]") != "SKU1,SKU2" {
			t.Fatalf("expected filter[sku]=SKU1,SKU2, got %q", values.Get("filter[sku]"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetApps(
		context.Background(),
		WithAppsBundleIDs([]string{"com.example.demo", "com.example.other"}),
		WithAppsNames([]string{"Demo App"}),
		WithAppsSKUs([]string{"SKU1", "SKU2"}),
	); err != nil {
		t.Fatalf("GetApps() error: %v", err)
	}
}

func TestGetApp(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"apps","id":"123","attributes":{"name":"Demo","bundleId":"com.example.demo","sku":"SKU1"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/apps/123" {
			t.Fatalf("expected path /v1/apps/123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetApp(context.Background(), "123"); err != nil {
		t.Fatalf("GetApp() error: %v", err)
	}
}

func TestGetBuilds_WithSortAndLimit(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"builds","id":"1","attributes":{"version":"1.0","uploadedDate":"2026-01-20T00:00:00Z"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		// When sorting or limiting, we use /v1/builds?filter[app]=APP_ID
		if req.URL.Path != "/v1/builds" {
			t.Fatalf("expected path /v1/builds, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[app]") != "123" {
			t.Fatalf("expected filter[app]=123, got %q", values.Get("filter[app]"))
		}
		if values.Get("sort") != "-uploadedDate" {
			t.Fatalf("expected sort=-uploadedDate, got %q", values.Get("sort"))
		}
		if values.Get("limit") != "5" {
			t.Fatalf("expected limit=5, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBuilds(context.Background(), "123", WithBuildsLimit(5), WithBuildsSort("-uploadedDate")); err != nil {
		t.Fatalf("GetBuilds() error: %v", err)
	}
}

func TestGetBuilds_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/builds?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBuilds(context.Background(), "123", WithBuildsLimit(5), WithBuildsSort("uploadedDate"), WithBuildsNextURL(next)); err != nil {
		t.Fatalf("GetBuilds() error: %v", err)
	}
}

func TestGetBuilds_WithPreReleaseVersion(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"builds","id":"build-1","attributes":{"version":"1.0","uploadedDate":"2026-01-20T00:00:00Z"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		// When filtering by preReleaseVersion, we use /v1/builds endpoint
		if req.URL.Path != "/v1/builds" {
			t.Fatalf("expected path /v1/builds, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[app]") != "123" {
			t.Fatalf("expected filter[app]=123, got %q", values.Get("filter[app]"))
		}
		if values.Get("filter[preReleaseVersion]") != "prv-456" {
			t.Fatalf("expected filter[preReleaseVersion]=prv-456, got %q", values.Get("filter[preReleaseVersion]"))
		}
		if values.Get("sort") != "-uploadedDate" {
			t.Fatalf("expected sort=-uploadedDate, got %q", values.Get("sort"))
		}
		if values.Get("limit") != "1" {
			t.Fatalf("expected limit=1, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	builds, err := client.GetBuilds(context.Background(), "123",
		WithBuildsLimit(1),
		WithBuildsSort("-uploadedDate"),
		WithBuildsPreReleaseVersion("prv-456"),
	)
	if err != nil {
		t.Fatalf("GetBuilds() error: %v", err)
	}
	if len(builds.Data) != 1 {
		t.Fatalf("expected 1 build, got %d", len(builds.Data))
	}
	if builds.Data[0].ID != "build-1" {
		t.Fatalf("expected build ID build-1, got %s", builds.Data[0].ID)
	}
}

func TestGetAppStoreVersions_WithFilters(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"appStoreVersions","id":"1","attributes":{"versionString":"1.0.0","platform":"IOS"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/apps/123/appStoreVersions" {
			t.Fatalf("expected path /v1/apps/123/appStoreVersions, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[platform]") != "IOS" {
			t.Fatalf("expected filter[platform]=IOS, got %q", values.Get("filter[platform]"))
		}
		if values.Get("filter[versionString]") != "1.0.0" {
			t.Fatalf("expected filter[versionString]=1.0.0, got %q", values.Get("filter[versionString]"))
		}
		if values.Get("filter[appStoreState]") != "READY_FOR_REVIEW" {
			t.Fatalf("expected filter[appStoreState]=READY_FOR_REVIEW, got %q", values.Get("filter[appStoreState]"))
		}
		if values.Get("limit") != "5" {
			t.Fatalf("expected limit=5, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAppStoreVersions(
		context.Background(),
		"123",
		WithAppStoreVersionsLimit(5),
		WithAppStoreVersionsPlatforms([]string{"IOS"}),
		WithAppStoreVersionsVersionStrings([]string{"1.0.0"}),
		WithAppStoreVersionsStates([]string{"READY_FOR_REVIEW"}),
	); err != nil {
		t.Fatalf("GetAppStoreVersions() error: %v", err)
	}
}

func TestGetAppStoreVersions_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/apps/123/appStoreVersions?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAppStoreVersions(context.Background(), "123", WithAppStoreVersionsNextURL(next)); err != nil {
		t.Fatalf("GetAppStoreVersions() error: %v", err)
	}
}

func TestGetPreReleaseVersions_WithFilters(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"preReleaseVersions","id":"1","attributes":{"version":"1.0.0","platform":"IOS"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/preReleaseVersions" {
			t.Fatalf("expected path /v1/preReleaseVersions, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[app]") != "123" {
			t.Fatalf("expected filter[app]=123, got %q", values.Get("filter[app]"))
		}
		if values.Get("filter[platform]") != "IOS" {
			t.Fatalf("expected filter[platform]=IOS, got %q", values.Get("filter[platform]"))
		}
		if values.Get("filter[version]") != "1.0.0" {
			t.Fatalf("expected filter[version]=1.0.0, got %q", values.Get("filter[version]"))
		}
		if values.Get("limit") != "5" {
			t.Fatalf("expected limit=5, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetPreReleaseVersions(
		context.Background(),
		"123",
		WithPreReleaseVersionsLimit(5),
		WithPreReleaseVersionsPlatform("ios"),
		WithPreReleaseVersionsVersion("1.0.0"),
	); err != nil {
		t.Fatalf("GetPreReleaseVersions() error: %v", err)
	}
}

func TestGetPreReleaseVersions_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/preReleaseVersions?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetPreReleaseVersions(context.Background(), "123", WithPreReleaseVersionsNextURL(next)); err != nil {
		t.Fatalf("GetPreReleaseVersions() error: %v", err)
	}
}

func TestGetPreReleaseVersion(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"preReleaseVersions","id":"pr-1","attributes":{"version":"1.0.0","platform":"IOS"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/preReleaseVersions/pr-1" {
			t.Fatalf("expected path /v1/preReleaseVersions/pr-1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetPreReleaseVersion(context.Background(), "pr-1"); err != nil {
		t.Fatalf("GetPreReleaseVersion() error: %v", err)
	}
}

func TestGetAppStoreVersion(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"appStoreVersions","id":"1","attributes":{"versionString":"1.0.0","platform":"IOS"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersions/1" {
			t.Fatalf("expected path /v1/appStoreVersions/1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAppStoreVersion(context.Background(), "1"); err != nil {
		t.Fatalf("GetAppStoreVersion() error: %v", err)
	}
}

func TestAttachBuildToVersion(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, ``)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersions/1/relationships/build" {
			t.Fatalf("expected path /v1/appStoreVersions/1/relationships/build, got %s", req.URL.Path)
		}
		var body AppStoreVersionBuildRelationshipUpdateRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if body.Data.Type != ResourceTypeBuilds || body.Data.ID != "BUILD_123" {
			t.Fatalf("unexpected request payload: %+v", body.Data)
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.AttachBuildToVersion(context.Background(), "1", "BUILD_123"); err != nil {
		t.Fatalf("AttachBuildToVersion() error: %v", err)
	}
}

func TestAddBetaGroupsToBuild_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, ``)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/builds/build-1/relationships/betaGroups" {
			t.Fatalf("expected path /v1/builds/build-1/relationships/betaGroups, got %s", req.URL.Path)
		}
		var payload RelationshipRequest
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(payload.Data) != 2 {
			t.Fatalf("expected 2 relationships, got %d", len(payload.Data))
		}
		if payload.Data[0].Type != ResourceTypeBetaGroups || payload.Data[0].ID != "group-1" {
			t.Fatalf("unexpected relationship[0]: %+v", payload.Data[0])
		}
		if payload.Data[1].Type != ResourceTypeBetaGroups || payload.Data[1].ID != "group-2" {
			t.Fatalf("unexpected relationship[1]: %+v", payload.Data[1])
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.AddBetaGroupsToBuild(context.Background(), "build-1", []string{"group-1", "group-2"}); err != nil {
		t.Fatalf("AddBetaGroupsToBuild() error: %v", err)
	}
}

func TestRemoveBetaGroupsFromBuild_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, ``)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", req.Method)
		}
		if req.URL.Path != "/v1/builds/build-1/relationships/betaGroups" {
			t.Fatalf("expected path /v1/builds/build-1/relationships/betaGroups, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		if len(body) == 0 {
			t.Fatalf("expected request body for delete")
		}
		var payload RelationshipRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if len(payload.Data) != 1 {
			t.Fatalf("expected 1 relationship, got %d", len(payload.Data))
		}
		if payload.Data[0].Type != ResourceTypeBetaGroups || payload.Data[0].ID != "group-1" {
			t.Fatalf("unexpected relationship: %+v", payload.Data[0])
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.RemoveBetaGroupsFromBuild(context.Background(), "build-1", []string{"group-1"}); err != nil {
		t.Fatalf("RemoveBetaGroupsFromBuild() error: %v", err)
	}
}

func TestGetAppStoreVersionBuild(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"builds","id":"BUILD_123","attributes":{"version":"1.0.0","uploadedDate":"2026-01-20T00:00:00Z"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersions/1/build" {
			t.Fatalf("expected path /v1/appStoreVersions/1/build, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAppStoreVersionBuild(context.Background(), "1"); err != nil {
		t.Fatalf("GetAppStoreVersionBuild() error: %v", err)
	}
}

func TestGetAppStoreVersionSubmissionResource(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"appStoreVersionSubmissions","id":"SUBMIT_123","relationships":{"appStoreVersion":{"data":{"type":"appStoreVersions","id":"VERSION_123"}}}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersionSubmissions/SUBMIT_123" {
			t.Fatalf("expected path /v1/appStoreVersionSubmissions/SUBMIT_123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAppStoreVersionSubmissionResource(context.Background(), "SUBMIT_123"); err != nil {
		t.Fatalf("GetAppStoreVersionSubmissionResource() error: %v", err)
	}
}

func TestGetAppStoreVersionSubmissionForVersion(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"appStoreVersionSubmissions","id":"SUBMIT_123","relationships":{"appStoreVersion":{"data":{"type":"appStoreVersions","id":"VERSION_123"}}}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersions/VERSION_123/appStoreVersionSubmission" {
			t.Fatalf("expected path /v1/appStoreVersions/VERSION_123/appStoreVersionSubmission, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAppStoreVersionSubmissionForVersion(context.Background(), "VERSION_123"); err != nil {
		t.Fatalf("GetAppStoreVersionSubmissionForVersion() error: %v", err)
	}
}

func TestGetBetaGroups_WithLimit(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"betaGroups","id":"1","attributes":{"name":"Beta"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/apps/123/betaGroups" {
			t.Fatalf("expected path /v1/apps/123/betaGroups, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("limit") != "10" {
			t.Fatalf("expected limit=10, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBetaGroups(context.Background(), "123", WithBetaGroupsLimit(10)); err != nil {
		t.Fatalf("GetBetaGroups() error: %v", err)
	}
}

func TestGetBetaGroups_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/apps/123/betaGroups?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBetaGroups(context.Background(), "123", WithBetaGroupsLimit(5), WithBetaGroupsNextURL(next)); err != nil {
		t.Fatalf("GetBetaGroups() error: %v", err)
	}
}

func TestGetBetaGroupBuilds_WithLimit(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"builds","id":"build-1","attributes":{"version":"1.0","uploadedDate":"2026-01-20T00:00:00Z"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaGroups/group-1/builds" {
			t.Fatalf("expected path /v1/betaGroups/group-1/builds, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("limit") != "50" {
			t.Fatalf("expected limit=50, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBetaGroupBuilds(context.Background(), "group-1", WithBetaGroupBuildsLimit(50)); err != nil {
		t.Fatalf("GetBetaGroupBuilds() error: %v", err)
	}
}

func TestGetBetaGroupBuilds_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/betaGroups/group-1/builds?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBetaGroupBuilds(context.Background(), "group-1", WithBetaGroupBuildsNextURL(next)); err != nil {
		t.Fatalf("GetBetaGroupBuilds() error: %v", err)
	}
}

func TestGetBetaGroupTesters_WithLimit(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"betaTesters","id":"tester-1","attributes":{"email":"tester@example.com"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaGroups/group-1/betaTesters" {
			t.Fatalf("expected path /v1/betaGroups/group-1/betaTesters, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("limit") != "20" {
			t.Fatalf("expected limit=20, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBetaGroupTesters(context.Background(), "group-1", WithBetaGroupTestersLimit(20)); err != nil {
		t.Fatalf("GetBetaGroupTesters() error: %v", err)
	}
}

func TestGetBetaGroupTesters_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/betaGroups/group-1/betaTesters?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBetaGroupTesters(context.Background(), "group-1", WithBetaGroupTestersNextURL(next)); err != nil {
		t.Fatalf("GetBetaGroupTesters() error: %v", err)
	}
}

func TestGetBetaTesters_WithAppFilter(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"betaTesters","id":"1","attributes":{"email":"tester@example.com"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaTesters" {
			t.Fatalf("expected path /v1/betaTesters, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[apps]") != "123" {
			t.Fatalf("expected filter[apps]=123, got %q", values.Get("filter[apps]"))
		}
		if values.Get("filter[email]") != "tester@example.com" {
			t.Fatalf("expected filter[email]=tester@example.com, got %q", values.Get("filter[email]"))
		}
		if values.Get("filter[betaGroups]") != "group-1,group-2" {
			t.Fatalf("expected filter[betaGroups]=group-1,group-2, got %q", values.Get("filter[betaGroups]"))
		}
		if values.Get("limit") != "5" {
			t.Fatalf("expected limit=5, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBetaTesters(
		context.Background(),
		"123",
		WithBetaTestersEmail("tester@example.com"),
		WithBetaTestersGroupIDs([]string{"group-1", "group-2"}),
		WithBetaTestersLimit(5),
	); err != nil {
		t.Fatalf("GetBetaTesters() error: %v", err)
	}
}

func TestGetBetaTesters_WithBuildFilter(t *testing.T) {
	// API only allows one relationship filter, so builds takes precedence over apps
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"betaTesters","id":"1","attributes":{"email":"tester@example.com"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaTesters" {
			t.Fatalf("expected path /v1/betaTesters, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		// When build filter is provided, apps filter should NOT be present
		if values.Get("filter[apps]") != "" {
			t.Fatalf("expected no filter[apps] when filter[builds] is set, got %q", values.Get("filter[apps]"))
		}
		if values.Get("filter[builds]") != "build-1" {
			t.Fatalf("expected filter[builds]=build-1, got %q", values.Get("filter[builds]"))
		}
		if values.Get("filter[email]") != "tester@example.com" {
			t.Fatalf("expected filter[email]=tester@example.com, got %q", values.Get("filter[email]"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBetaTesters(
		context.Background(),
		"123", // appID provided but should be ignored when build filter is set
		WithBetaTestersEmail("tester@example.com"),
		WithBetaTestersBuildID("build-1"),
	); err != nil {
		t.Fatalf("GetBetaTesters() error: %v", err)
	}
}

func TestGetBetaTesters_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/betaTesters?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBetaTesters(context.Background(), "123", WithBetaTestersLimit(5), WithBetaTestersNextURL(next)); err != nil {
		t.Fatalf("GetBetaTesters() error: %v", err)
	}
}

func TestGetBuild_ByID(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"builds","id":"123","attributes":{"version":"1.0","uploadedDate":"2026-01-20T00:00:00Z","expired":false}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/builds/123" {
			t.Fatalf("expected path /v1/builds/123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBuild(context.Background(), "123"); err != nil {
		t.Fatalf("GetBuild() error: %v", err)
	}
}

func TestGetApp_ByID(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"apps","id":"app-1","attributes":{"name":"Demo","bundleId":"com.example.demo","sku":"SKU1"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/apps/app-1" {
			t.Fatalf("expected path /v1/apps/app-1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetApp(context.Background(), "app-1"); err != nil {
		t.Fatalf("GetApp() error: %v", err)
	}
}

func TestGetBuildAppStoreVersion_ByBuildID(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"appStoreVersions","id":"version-1","attributes":{"versionString":"1.0"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/builds/123/appStoreVersion" {
			t.Fatalf("expected path /v1/builds/123/appStoreVersion, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBuildAppStoreVersion(context.Background(), "123"); err != nil {
		t.Fatalf("GetBuildAppStoreVersion() error: %v", err)
	}
}

func TestExpireBuild_SendsPatch(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"builds","id":"123","attributes":{"version":"1.0","uploadedDate":"2026-01-20T00:00:00Z","expired":true}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", req.Method)
		}
		if req.URL.Path != "/v1/builds/123" {
			t.Fatalf("expected path /v1/builds/123, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		var payload struct {
			Data struct {
				Type       string `json:"type"`
				ID         string `json:"id"`
				Attributes struct {
					Expired bool `json:"expired"`
				} `json:"attributes"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if payload.Data.Type != "builds" {
			t.Fatalf("expected type builds, got %q", payload.Data.Type)
		}
		if payload.Data.ID != "123" {
			t.Fatalf("expected id 123, got %q", payload.Data.ID)
		}
		if !payload.Data.Attributes.Expired {
			t.Fatalf("expected expired true")
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.ExpireBuild(context.Background(), "123"); err != nil {
		t.Fatalf("ExpireBuild() error: %v", err)
	}
}

func TestCreateBetaGroup_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"betaGroups","id":"bg1","attributes":{"name":"Beta"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaGroups" {
			t.Fatalf("expected path /v1/betaGroups, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		var payload BetaGroupCreateRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if payload.Data.Type != ResourceTypeBetaGroups {
			t.Fatalf("expected type betaGroups, got %q", payload.Data.Type)
		}
		if payload.Data.Attributes.Name != "Beta" {
			t.Fatalf("expected name Beta, got %q", payload.Data.Attributes.Name)
		}
		if payload.Data.Relationships == nil || payload.Data.Relationships.App == nil {
			t.Fatalf("expected app relationship to be set")
		}
		if payload.Data.Relationships.App.Data.ID != "app-1" {
			t.Fatalf("expected app id app-1, got %q", payload.Data.Relationships.App.Data.ID)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.CreateBetaGroup(context.Background(), "app-1", "Beta"); err != nil {
		t.Fatalf("CreateBetaGroup() error: %v", err)
	}
}

func TestGetBetaGroup_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"betaGroups","id":"bg1","attributes":{"name":"Beta Testers","isInternalGroup":true}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaGroups/bg1" {
			t.Fatalf("expected path /v1/betaGroups/bg1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetBetaGroup(context.Background(), "bg1"); err != nil {
		t.Fatalf("GetBetaGroup() error: %v", err)
	}
}

func TestGetBetaTester_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"betaTesters","id":"bt1","attributes":{"email":"tester@example.com","firstName":"Test","lastName":"User","state":"INVITED","inviteType":"EMAIL"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaTesters/bt1" {
			t.Fatalf("expected path /v1/betaTesters/bt1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	tester, err := client.GetBetaTester(context.Background(), "bt1")
	if err != nil {
		t.Fatalf("GetBetaTester() error: %v", err)
	}
	if tester.Data.ID != "bt1" {
		t.Fatalf("expected tester id bt1, got %q", tester.Data.ID)
	}
	if tester.Data.Attributes.Email != "tester@example.com" {
		t.Fatalf("expected tester email tester@example.com, got %q", tester.Data.Attributes.Email)
	}
	if tester.Data.Attributes.State != BetaTesterStateInvited {
		t.Fatalf("expected state %q, got %q", BetaTesterStateInvited, tester.Data.Attributes.State)
	}
	if tester.Data.Attributes.InviteType != BetaInviteTypeEmail {
		t.Fatalf("expected invite type %q, got %q", BetaInviteTypeEmail, tester.Data.Attributes.InviteType)
	}
}

func TestUpdateBetaGroup_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"betaGroups","id":"bg1","attributes":{"name":"Updated Beta Testers","publicLinkEnabled":true}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaGroups/bg1" {
			t.Fatalf("expected path /v1/betaGroups/bg1, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		var payload BetaGroupUpdateRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if payload.Data.Type != ResourceTypeBetaGroups {
			t.Fatalf("expected type betaGroups, got %q", payload.Data.Type)
		}
		if payload.Data.ID != "bg1" {
			t.Fatalf("expected id bg1, got %q", payload.Data.ID)
		}
		if payload.Data.Attributes == nil {
			t.Fatalf("expected attributes to be set")
		}
		if payload.Data.Attributes.Name != "Updated Beta Testers" {
			t.Fatalf("expected name Updated Beta Testers, got %q", payload.Data.Attributes.Name)
		}
		assertAuthorized(t, req)
	}, response)

	req := BetaGroupUpdateRequest{
		Data: BetaGroupUpdateData{
			Type:       ResourceTypeBetaGroups,
			ID:         "bg1",
			Attributes: &BetaGroupUpdateAttributes{Name: "Updated Beta Testers"},
		},
	}
	if _, err := client.UpdateBetaGroup(context.Background(), "bg1", req); err != nil {
		t.Fatalf("UpdateBetaGroup() error: %v", err)
	}
}

func TestDeleteBetaGroup_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, "")
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaGroups/bg1" {
			t.Fatalf("expected path /v1/betaGroups/bg1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.DeleteBetaGroup(context.Background(), "bg1"); err != nil {
		t.Fatalf("DeleteBetaGroup() error: %v", err)
	}
}

func TestAddBetaTestersToGroup_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, "")
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaGroups/bg1/relationships/betaTesters" {
			t.Fatalf("expected path /v1/betaGroups/bg1/relationships/betaTesters, got %s", req.URL.Path)
		}
		var payload RelationshipRequest
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(payload.Data) != 2 {
			t.Fatalf("expected 2 tester relationships, got %d", len(payload.Data))
		}
		if payload.Data[0].Type != ResourceTypeBetaTesters || payload.Data[0].ID != "tester-1" {
			t.Fatalf("unexpected tester data: %+v", payload.Data[0])
		}
		if payload.Data[1].Type != ResourceTypeBetaTesters || payload.Data[1].ID != "tester-2" {
			t.Fatalf("unexpected tester data: %+v", payload.Data[1])
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.AddBetaTestersToGroup(context.Background(), "bg1", []string{"tester-1", "tester-2"}); err != nil {
		t.Fatalf("AddBetaTestersToGroup() error: %v", err)
	}
}

func TestRemoveBetaTestersFromGroup_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, "")
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaGroups/bg1/relationships/betaTesters" {
			t.Fatalf("expected path /v1/betaGroups/bg1/relationships/betaTesters, got %s", req.URL.Path)
		}
		var payload RelationshipRequest
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(payload.Data) != 1 {
			t.Fatalf("expected 1 tester relationship, got %d", len(payload.Data))
		}
		if payload.Data[0].Type != ResourceTypeBetaTesters || payload.Data[0].ID != "tester-1" {
			t.Fatalf("unexpected tester data: %+v", payload.Data[0])
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.RemoveBetaTestersFromGroup(context.Background(), "bg1", []string{"tester-1"}); err != nil {
		t.Fatalf("RemoveBetaTestersFromGroup() error: %v", err)
	}
}

func TestCreateBetaTester_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"betaTesters","id":"bt1","attributes":{"email":"tester@example.com"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaTesters" {
			t.Fatalf("expected path /v1/betaTesters, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		var payload BetaTesterCreateRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if payload.Data.Type != ResourceTypeBetaTesters {
			t.Fatalf("expected type betaTesters, got %q", payload.Data.Type)
		}
		if payload.Data.Attributes.Email != "tester@example.com" {
			t.Fatalf("expected email tester@example.com, got %q", payload.Data.Attributes.Email)
		}
		if payload.Data.Relationships == nil || payload.Data.Relationships.BetaGroups == nil {
			t.Fatalf("expected betaGroups relationship to be set")
		}
		if len(payload.Data.Relationships.BetaGroups.Data) != 2 {
			t.Fatalf("expected 2 beta group relationships, got %d", len(payload.Data.Relationships.BetaGroups.Data))
		}
		if payload.Data.Relationships.BetaGroups.Data[0].ID != "group-1" {
			t.Fatalf("expected group-1, got %q", payload.Data.Relationships.BetaGroups.Data[0].ID)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.CreateBetaTester(context.Background(), "tester@example.com", "Test", "User", []string{"group-1", "group-2"}); err != nil {
		t.Fatalf("CreateBetaTester() error: %v", err)
	}
}

func TestBetaGroupTesterRelationshipMethods_ErrorResponse(t *testing.T) {
	ctx := context.Background()
	errorBody := `{"errors":[{"code":"BAD_REQUEST","title":"Bad Request","detail":"nope"}]}`

	tests := []struct {
		name string
		call func(*Client) error
	}{
		{
			name: "AddBetaTestersToGroup",
			call: func(c *Client) error {
				return c.AddBetaTestersToGroup(ctx, "bg1", []string{"tester-1"})
			},
		},
		{
			name: "RemoveBetaTestersFromGroup",
			call: func(c *Client) error {
				return c.RemoveBetaTestersFromGroup(ctx, "bg1", []string{"tester-1"})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := newTestClient(t, nil, jsonResponse(http.StatusBadRequest, errorBody))
			if err := test.call(client); err == nil {
				t.Fatalf("expected error")
			} else if !strings.Contains(err.Error(), "Bad Request") {
				t.Fatalf("expected error to contain title, got %v", err)
			}
		})
	}
}

func TestDeleteBetaTester_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, ``)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaTesters/bt-1" {
			t.Fatalf("expected path /v1/betaTesters/bt-1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.DeleteBetaTester(context.Background(), "bt-1"); err != nil {
		t.Fatalf("DeleteBetaTester() error: %v", err)
	}
}

func TestAddBetaTesterToGroups_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, ``)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaTesters/bt-1/relationships/betaGroups" {
			t.Fatalf("expected path /v1/betaTesters/bt-1/relationships/betaGroups, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		var payload RelationshipList
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if len(payload.Data) != 2 {
			t.Fatalf("expected 2 beta group relationships, got %d", len(payload.Data))
		}
		if payload.Data[0].Type != ResourceTypeBetaGroups {
			t.Fatalf("expected type betaGroups, got %q", payload.Data[0].Type)
		}
		if payload.Data[0].ID != "group-1" {
			t.Fatalf("expected group-1, got %q", payload.Data[0].ID)
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.AddBetaTesterToGroups(context.Background(), "bt-1", []string{"group-1", "group-2"}); err != nil {
		t.Fatalf("AddBetaTesterToGroups() error: %v", err)
	}
}

func TestRemoveBetaTesterFromGroups_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, ``)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaTesters/bt-1/relationships/betaGroups" {
			t.Fatalf("expected path /v1/betaTesters/bt-1/relationships/betaGroups, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		var payload RelationshipList
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if len(payload.Data) != 2 {
			t.Fatalf("expected 2 beta group relationships, got %d", len(payload.Data))
		}
		if payload.Data[0].Type != ResourceTypeBetaGroups {
			t.Fatalf("expected type betaGroups, got %q", payload.Data[0].Type)
		}
		if payload.Data[1].ID != "group-2" {
			t.Fatalf("expected group-2, got %q", payload.Data[1].ID)
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.RemoveBetaTesterFromGroups(context.Background(), "bt-1", []string{"group-1", "group-2"}); err != nil {
		t.Fatalf("RemoveBetaTesterFromGroups() error: %v", err)
	}
}

func TestCreateBetaTesterInvitation_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"betaTesterInvitations","id":"invite-1"}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/betaTesterInvitations" {
			t.Fatalf("expected path /v1/betaTesterInvitations, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		var payload BetaTesterInvitationCreateRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if payload.Data.Type != ResourceTypeBetaTesterInvitations {
			t.Fatalf("expected type betaTesterInvitations, got %q", payload.Data.Type)
		}
		if payload.Data.Relationships == nil || payload.Data.Relationships.App == nil {
			t.Fatalf("expected app relationship to be set")
		}
		if payload.Data.Relationships.App.Data.ID != "app-1" {
			t.Fatalf("expected app id app-1, got %q", payload.Data.Relationships.App.Data.ID)
		}
		if payload.Data.Relationships.BetaTester == nil || payload.Data.Relationships.BetaTester.Data.ID != "tester-1" {
			t.Fatalf("expected beta tester id tester-1")
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.CreateBetaTesterInvitation(context.Background(), "app-1", "tester-1"); err != nil {
		t.Fatalf("CreateBetaTesterInvitation() error: %v", err)
	}
}

func TestGetAppStoreVersionLocalizations_WithFilters(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"appStoreVersionLocalizations","id":"loc-1","attributes":{"locale":"en-US"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersions/version-1/appStoreVersionLocalizations" {
			t.Fatalf("expected path /v1/appStoreVersions/version-1/appStoreVersionLocalizations, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[locale]") != "en-US" {
			t.Fatalf("expected filter[locale]=en-US, got %q", values.Get("filter[locale]"))
		}
		if values.Get("limit") != "10" {
			t.Fatalf("expected limit=10, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAppStoreVersionLocalizations(
		context.Background(),
		"version-1",
		WithAppStoreVersionLocalizationLocales([]string{"en-US"}),
		WithAppStoreVersionLocalizationsLimit(10),
	); err != nil {
		t.Fatalf("GetAppStoreVersionLocalizations() error: %v", err)
	}
}

func TestGetAppStoreVersionLocalization_ByID(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"appStoreVersionLocalizations","id":"loc-1","attributes":{"locale":"en-US"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersionLocalizations/loc-1" {
			t.Fatalf("expected path /v1/appStoreVersionLocalizations/loc-1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAppStoreVersionLocalization(context.Background(), "loc-1"); err != nil {
		t.Fatalf("GetAppStoreVersionLocalization() error: %v", err)
	}
}

func TestCreateAppStoreVersionLocalization_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"appStoreVersionLocalizations","id":"loc-1","attributes":{"locale":"en-US"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersionLocalizations" {
			t.Fatalf("expected path /v1/appStoreVersionLocalizations, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		var payload AppStoreVersionLocalizationCreateRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if payload.Data.Type != ResourceTypeAppStoreVersionLocalizations {
			t.Fatalf("expected type appStoreVersionLocalizations, got %q", payload.Data.Type)
		}
		if payload.Data.Attributes.Locale != "en-US" {
			t.Fatalf("expected locale en-US, got %q", payload.Data.Attributes.Locale)
		}
		if payload.Data.Relationships == nil || payload.Data.Relationships.AppStoreVersion == nil {
			t.Fatalf("expected appStoreVersion relationship")
		}
		if payload.Data.Relationships.AppStoreVersion.Data.ID != "version-1" {
			t.Fatalf("expected version id version-1, got %q", payload.Data.Relationships.AppStoreVersion.Data.ID)
		}
		assertAuthorized(t, req)
	}, response)

	attrs := AppStoreVersionLocalizationAttributes{
		Locale:      "en-US",
		Description: "Hello",
	}
	if _, err := client.CreateAppStoreVersionLocalization(context.Background(), "version-1", attrs); err != nil {
		t.Fatalf("CreateAppStoreVersionLocalization() error: %v", err)
	}
}

func TestUpdateAppStoreVersionLocalization_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"appStoreVersionLocalizations","id":"loc-1","attributes":{"description":"Updated"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersionLocalizations/loc-1" {
			t.Fatalf("expected path /v1/appStoreVersionLocalizations/loc-1, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		var payload AppStoreVersionLocalizationUpdateRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if payload.Data.Type != ResourceTypeAppStoreVersionLocalizations {
			t.Fatalf("expected type appStoreVersionLocalizations, got %q", payload.Data.Type)
		}
		if payload.Data.ID != "loc-1" {
			t.Fatalf("expected id loc-1, got %q", payload.Data.ID)
		}
		assertAuthorized(t, req)
	}, response)

	attrs := AppStoreVersionLocalizationAttributes{
		Description: "Updated",
	}
	if _, err := client.UpdateAppStoreVersionLocalization(context.Background(), "loc-1", attrs); err != nil {
		t.Fatalf("UpdateAppStoreVersionLocalization() error: %v", err)
	}
}

func TestDeleteAppStoreVersionLocalization_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, "")
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersionLocalizations/loc-1" {
			t.Fatalf("expected path /v1/appStoreVersionLocalizations/loc-1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.DeleteAppStoreVersionLocalization(context.Background(), "loc-1"); err != nil {
		t.Fatalf("DeleteAppStoreVersionLocalization() error: %v", err)
	}
}

func TestGetAppInfoLocalizations_WithFilters(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"appInfoLocalizations","id":"loc-1","attributes":{"locale":"en-US"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appInfos/app-info-1/appInfoLocalizations" {
			t.Fatalf("expected path /v1/appInfos/app-info-1/appInfoLocalizations, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[locale]") != "en-US" {
			t.Fatalf("expected filter[locale]=en-US, got %q", values.Get("filter[locale]"))
		}
		if values.Get("limit") != "5" {
			t.Fatalf("expected limit=5, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAppInfoLocalizations(
		context.Background(),
		"app-info-1",
		WithAppInfoLocalizationLocales([]string{"en-US"}),
		WithAppInfoLocalizationsLimit(5),
	); err != nil {
		t.Fatalf("GetAppInfoLocalizations() error: %v", err)
	}
}

func TestCreateAppInfoLocalization_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"appInfoLocalizations","id":"loc-1","attributes":{"locale":"en-US"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appInfoLocalizations" {
			t.Fatalf("expected path /v1/appInfoLocalizations, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		var payload AppInfoLocalizationCreateRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if payload.Data.Type != ResourceTypeAppInfoLocalizations {
			t.Fatalf("expected type appInfoLocalizations, got %q", payload.Data.Type)
		}
		if payload.Data.Attributes.Locale != "en-US" {
			t.Fatalf("expected locale en-US, got %q", payload.Data.Attributes.Locale)
		}
		if payload.Data.Relationships == nil || payload.Data.Relationships.AppInfo == nil {
			t.Fatalf("expected appInfo relationship")
		}
		if payload.Data.Relationships.AppInfo.Data.ID != "app-info-1" {
			t.Fatalf("expected appInfo id app-info-1, got %q", payload.Data.Relationships.AppInfo.Data.ID)
		}
		assertAuthorized(t, req)
	}, response)

	attrs := AppInfoLocalizationAttributes{
		Locale: "en-US",
		Name:   "Demo App",
	}
	if _, err := client.CreateAppInfoLocalization(context.Background(), "app-info-1", attrs); err != nil {
		t.Fatalf("CreateAppInfoLocalization() error: %v", err)
	}
}

func TestUpdateAppInfoLocalization_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"appInfoLocalizations","id":"loc-1","attributes":{"name":"Updated"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appInfoLocalizations/loc-1" {
			t.Fatalf("expected path /v1/appInfoLocalizations/loc-1, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		var payload AppInfoLocalizationUpdateRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if payload.Data.Type != ResourceTypeAppInfoLocalizations {
			t.Fatalf("expected type appInfoLocalizations, got %q", payload.Data.Type)
		}
		if payload.Data.ID != "loc-1" {
			t.Fatalf("expected id loc-1, got %q", payload.Data.ID)
		}
		assertAuthorized(t, req)
	}, response)

	attrs := AppInfoLocalizationAttributes{
		Name: "Updated",
	}
	if _, err := client.UpdateAppInfoLocalization(context.Background(), "loc-1", attrs); err != nil {
		t.Fatalf("UpdateAppInfoLocalization() error: %v", err)
	}
}

func TestGetAppInfos(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"appInfos","id":"info-1"}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/apps/app-1/appInfos" {
			t.Fatalf("expected path /v1/apps/app-1/appInfos, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAppInfos(context.Background(), "app-1"); err != nil {
		t.Fatalf("GetAppInfos() error: %v", err)
	}
}

func TestGetFeedback_BuildsQuery(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"betaFeedbackScreenshotSubmissions","id":"1","attributes":{"createdDate":"2026-01-20T00:00:00Z","comment":"Nice","email":"tester@example.com"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.Path != "/v1/apps/123/betaFeedbackScreenshotSubmissions" {
			t.Fatalf("expected feedback path, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[deviceModel]") != "iPhone15,3" {
			t.Fatalf("expected deviceModel filter, got %q", values.Get("filter[deviceModel]"))
		}
		if values.Get("filter[osVersion]") != "17.2" {
			t.Fatalf("expected osVersion filter, got %q", values.Get("filter[osVersion]"))
		}
		if values.Get("filter[appPlatform]") != "IOS,MAC_OS" {
			t.Fatalf("expected appPlatform filter, got %q", values.Get("filter[appPlatform]"))
		}
		if values.Get("filter[devicePlatform]") != "TV_OS" {
			t.Fatalf("expected devicePlatform filter, got %q", values.Get("filter[devicePlatform]"))
		}
		if values.Get("filter[build]") != "build-1" {
			t.Fatalf("expected build filter, got %q", values.Get("filter[build]"))
		}
		if values.Get("filter[build.preReleaseVersion]") != "pre-1" {
			t.Fatalf("expected preRelease filter, got %q", values.Get("filter[build.preReleaseVersion]"))
		}
		if values.Get("filter[tester]") != "tester-1" {
			t.Fatalf("expected tester filter, got %q", values.Get("filter[tester]"))
		}
		if values.Get("sort") != "-createdDate" {
			t.Fatalf("expected sort=-createdDate, got %q", values.Get("sort"))
		}
		if values.Get("limit") != "10" {
			t.Fatalf("expected limit=10, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	opts := []FeedbackOption{
		WithFeedbackDeviceModels([]string{"iPhone15,3"}),
		WithFeedbackOSVersions([]string{"17.2"}),
		WithFeedbackAppPlatforms([]string{"ios", "mac_os"}),
		WithFeedbackDevicePlatforms([]string{"tv_os"}),
		WithFeedbackBuildIDs([]string{"build-1"}),
		WithFeedbackBuildPreReleaseVersionIDs([]string{"pre-1"}),
		WithFeedbackTesterIDs([]string{"tester-1"}),
		WithFeedbackLimit(10),
		WithFeedbackSort("-createdDate"),
	}

	if _, err := client.GetFeedback(context.Background(), "123", opts...); err != nil {
		t.Fatalf("GetFeedback() error: %v", err)
	}
}

func TestGetFeedback_IncludesScreenshots(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"betaFeedbackScreenshotSubmissions","id":"1","attributes":{"createdDate":"2026-01-20T00:00:00Z","comment":"Nice","email":"tester@example.com","screenshots":[{"url":"https://example.com/shot.png","width":320,"height":640,"expirationDate":"2026-01-21T00:00:00Z"}]}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		values := req.URL.Query()
		expected := "createdDate,comment,email,deviceModel,osVersion,appPlatform,devicePlatform,screenshots"
		if values.Get("fields[betaFeedbackScreenshotSubmissions]") != expected {
			t.Fatalf("expected screenshot fields, got %q", values.Get("fields[betaFeedbackScreenshotSubmissions]"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetFeedback(context.Background(), "123", WithFeedbackIncludeScreenshots()); err != nil {
		t.Fatalf("GetFeedback() error: %v", err)
	}
}

func TestGetCrashes_BuildsQuery(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"betaFeedbackCrashSubmissions","id":"1","attributes":{"createdDate":"2026-01-20T00:00:00Z","comment":"Crash","email":"tester@example.com","crashLog":"stack"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.Path != "/v1/apps/123/betaFeedbackCrashSubmissions" {
			t.Fatalf("expected crashes path, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[deviceModel]") != "iPhone16,1" {
			t.Fatalf("expected deviceModel filter, got %q", values.Get("filter[deviceModel]"))
		}
		if values.Get("filter[osVersion]") != "18.0" {
			t.Fatalf("expected osVersion filter, got %q", values.Get("filter[osVersion]"))
		}
		if values.Get("filter[appPlatform]") != "IOS" {
			t.Fatalf("expected appPlatform filter, got %q", values.Get("filter[appPlatform]"))
		}
		if values.Get("filter[devicePlatform]") != "MAC_OS" {
			t.Fatalf("expected devicePlatform filter, got %q", values.Get("filter[devicePlatform]"))
		}
		if values.Get("filter[build]") != "build-2" {
			t.Fatalf("expected build filter, got %q", values.Get("filter[build]"))
		}
		if values.Get("filter[build.preReleaseVersion]") != "pre-2" {
			t.Fatalf("expected preRelease filter, got %q", values.Get("filter[build.preReleaseVersion]"))
		}
		if values.Get("filter[tester]") != "tester-2" {
			t.Fatalf("expected tester filter, got %q", values.Get("filter[tester]"))
		}
		if values.Get("sort") != "createdDate" {
			t.Fatalf("expected sort=createdDate, got %q", values.Get("sort"))
		}
		if values.Get("limit") != "5" {
			t.Fatalf("expected limit=5, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

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

	if _, err := client.GetCrashes(context.Background(), "123", opts...); err != nil {
		t.Fatalf("GetCrashes() error: %v", err)
	}
}

func TestGetReviews_BuildsQuery(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"customerReviews","id":"1","attributes":{"rating":5,"title":"Great","body":"Nice","reviewerNickname":"Tester","createdDate":"2026-01-20T00:00:00Z","territory":"US"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.Path != "/v1/apps/123/customerReviews" {
			t.Fatalf("expected reviews path, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[rating]") != "5" {
			t.Fatalf("expected rating filter, got %q", values.Get("filter[rating]"))
		}
		if values.Get("filter[territory]") != "US" {
			t.Fatalf("expected territory filter, got %q", values.Get("filter[territory]"))
		}
		if values.Get("sort") != "-createdDate" {
			t.Fatalf("expected sort=-createdDate, got %q", values.Get("sort"))
		}
		if values.Get("limit") != "25" {
			t.Fatalf("expected limit=25, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	opts := []ReviewOption{
		WithRating(5),
		WithTerritory("us"),
		WithReviewSort("-createdDate"),
		WithLimit(25),
	}

	if _, err := client.GetReviews(context.Background(), "123", opts...); err != nil {
		t.Fatalf("GetReviews() error: %v", err)
	}
}

func TestGetEndpoints_ReturnsAPIError(t *testing.T) {
	tests := []struct {
		name string
		call func(*Client) error
	}{
		{
			name: "apps",
			call: func(c *Client) error {
				_, err := c.GetApps(context.Background())
				return err
			},
		},
		{
			name: "builds",
			call: func(c *Client) error {
				_, err := c.GetBuilds(context.Background(), "123")
				return err
			},
		},
		{
			name: "feedback",
			call: func(c *Client) error {
				_, err := c.GetFeedback(context.Background(), "123")
				return err
			},
		},
		{
			name: "crashes",
			call: func(c *Client) error {
				_, err := c.GetCrashes(context.Background(), "123")
				return err
			},
		},
		{
			name: "reviews",
			call: func(c *Client) error {
				_, err := c.GetReviews(context.Background(), "123")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := jsonResponse(http.StatusForbidden, `{"errors":[{"code":"FORBIDDEN","title":"Forbidden","detail":"not allowed"}]}`)
			client := newTestClient(t, nil, response)
			err := tt.call(client)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "Forbidden") {
				t.Fatalf("expected Forbidden error, got %v", err)
			}
		})
	}
}

func TestGetEndpoints_ReturnsParseError(t *testing.T) {
	tests := []struct {
		name string
		call func(*Client) error
	}{
		{
			name: "apps",
			call: func(c *Client) error {
				_, err := c.GetApps(context.Background())
				return err
			},
		},
		{
			name: "builds",
			call: func(c *Client) error {
				_, err := c.GetBuilds(context.Background(), "123")
				return err
			},
		},
		{
			name: "feedback",
			call: func(c *Client) error {
				_, err := c.GetFeedback(context.Background(), "123")
				return err
			},
		},
		{
			name: "crashes",
			call: func(c *Client) error {
				_, err := c.GetCrashes(context.Background(), "123")
				return err
			},
		},
		{
			name: "reviews",
			call: func(c *Client) error {
				_, err := c.GetReviews(context.Background(), "123")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := jsonResponse(http.StatusOK, `{"data":[}`)
			client := newTestClient(t, nil, response)
			err := tt.call(client)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), "failed to parse response") {
				t.Fatalf("expected parse error, got %v", err)
			}
		})
	}
}

func TestIsNotFoundAndUnauthorized(t *testing.T) {
	if !IsNotFound(fmt.Errorf("NOT_FOUND: missing")) {
		t.Fatal("expected IsNotFound to return true")
	}
	if !IsNotFound(fmt.Errorf("The specified resource does not exist")) {
		t.Fatal("expected IsNotFound to return true for resource does not exist")
	}
	if IsNotFound(fmt.Errorf("something else")) {
		t.Fatal("expected IsNotFound to return false")
	}
	if !IsUnauthorized(fmt.Errorf("UNAUTHORIZED: missing")) {
		t.Fatal("expected IsUnauthorized to return true")
	}
	if IsUnauthorized(fmt.Errorf("something else")) {
		t.Fatal("expected IsUnauthorized to return false")
	}
}

func TestCreateBuildUpload(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"buildUploads","id":"UPLOAD_123","attributes":{"cfBundleShortVersionString":"1.0.0","cfBundleVersion":"123","platform":"IOS"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/buildUploads" {
			t.Fatalf("expected path /v1/buildUploads, got %s", req.URL.Path)
		}
		if !strings.Contains(req.Header.Get("Content-Type"), "application/json") {
			t.Fatalf("expected Content-Type application/json")
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.CreateBuildUpload(context.Background(), BuildUploadCreateRequest{
		Data: BuildUploadCreateData{
			Type: ResourceTypeBuildUploads,
			Attributes: BuildUploadAttributes{
				CFBundleShortVersionString: "1.0.0",
				CFBundleVersion:            "123",
				Platform:                   PlatformIOS,
			},
			Relationships: &BuildUploadRelationships{
				App: &Relationship{
					Data: ResourceData{Type: ResourceTypeApps, ID: "APP_123"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateBuildUpload() error: %v", err)
	}
	if result.Data.ID != "UPLOAD_123" {
		t.Fatalf("expected upload ID UPLOAD_123, got %s", result.Data.ID)
	}
}

func TestGetBuildUpload(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"buildUploads","id":"UPLOAD_123","attributes":{"cfBundleShortVersionString":"1.0.0","cfBundleVersion":"123","platform":"IOS"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/buildUploads/UPLOAD_123" {
			t.Fatalf("expected path /v1/buildUploads/UPLOAD_123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.GetBuildUpload(context.Background(), "UPLOAD_123")
	if err != nil {
		t.Fatalf("GetBuildUpload() error: %v", err)
	}
	if result.Data.ID != "UPLOAD_123" {
		t.Fatalf("expected upload ID UPLOAD_123, got %s", result.Data.ID)
	}
}

func TestCreateBuildUploadFile(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"buildUploadFiles","id":"FILE_123","attributes":{"fileName":"app.ipa","fileSize":1024000,"uti":"com.apple.ipa","assetType":"ASSET","uploadOperations":[{"method":"PUT","url":"https://example.com/upload","length":1024000,"offset":0}]}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/buildUploadFiles" {
			t.Fatalf("expected path /v1/buildUploadFiles, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.CreateBuildUploadFile(context.Background(), BuildUploadFileCreateRequest{
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
					Data: ResourceData{Type: ResourceTypeBuildUploads, ID: "UPLOAD_123"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateBuildUploadFile() error: %v", err)
	}
	if result.Data.ID != "FILE_123" {
		t.Fatalf("expected file ID FILE_123, got %s", result.Data.ID)
	}
}

func TestUpdateBuildUploadFile(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"buildUploadFiles","id":"FILE_123","attributes":{"uploaded":true}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", req.Method)
		}
		if req.URL.Path != "/v1/buildUploadFiles/FILE_123" {
			t.Fatalf("expected path /v1/buildUploadFiles/FILE_123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	uploaded := true
	result, err := client.UpdateBuildUploadFile(context.Background(), "FILE_123", BuildUploadFileUpdateRequest{
		Data: BuildUploadFileUpdateData{
			Type: ResourceTypeBuildUploadFiles,
			ID:   "FILE_123",
			Attributes: &BuildUploadFileUpdateAttributes{
				Uploaded: &uploaded,
			},
		},
	})
	if err != nil {
		t.Fatalf("UpdateBuildUploadFile() error: %v", err)
	}
	if result.Data.ID != "FILE_123" {
		t.Fatalf("expected file ID FILE_123, got %s", result.Data.ID)
	}
}

func TestGetAppScreenshotSets(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"appScreenshotSets","id":"SET_123","attributes":{"screenshotDisplayType":"APP_IPHONE_65"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersionLocalizations/LOC_123/appScreenshotSets" {
			t.Fatalf("expected path /v1/appStoreVersionLocalizations/LOC_123/appScreenshotSets, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.GetAppScreenshotSets(context.Background(), "LOC_123")
	if err != nil {
		t.Fatalf("GetAppScreenshotSets() error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 set, got %d", len(result.Data))
	}
}

func TestCreateAppScreenshotSet(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"appScreenshotSets","id":"SET_123","attributes":{"screenshotDisplayType":"APP_IPHONE_65"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appScreenshotSets" {
			t.Fatalf("expected path /v1/appScreenshotSets, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.CreateAppScreenshotSet(context.Background(), "LOC_123", "APP_IPHONE_65")
	if err != nil {
		t.Fatalf("CreateAppScreenshotSet() error: %v", err)
	}
	if result.Data.ID != "SET_123" {
		t.Fatalf("expected set ID SET_123, got %s", result.Data.ID)
	}
}

func TestDeleteAppScreenshotSet(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, "")
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appScreenshotSets/SET_123" {
			t.Fatalf("expected path /v1/appScreenshotSets/SET_123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.DeleteAppScreenshotSet(context.Background(), "SET_123"); err != nil {
		t.Fatalf("DeleteAppScreenshotSet() error: %v", err)
	}
}

func TestGetAppScreenshots(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"appScreenshots","id":"SHOT_123","attributes":{"fileName":"shot.png","fileSize":1024}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appScreenshotSets/SET_123/appScreenshots" {
			t.Fatalf("expected path /v1/appScreenshotSets/SET_123/appScreenshots, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.GetAppScreenshots(context.Background(), "SET_123")
	if err != nil {
		t.Fatalf("GetAppScreenshots() error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 screenshot, got %d", len(result.Data))
	}
}

func TestGetAppScreenshot(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"appScreenshots","id":"SHOT_123","attributes":{"fileName":"shot.png","fileSize":1024}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appScreenshots/SHOT_123" {
			t.Fatalf("expected path /v1/appScreenshots/SHOT_123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.GetAppScreenshot(context.Background(), "SHOT_123")
	if err != nil {
		t.Fatalf("GetAppScreenshot() error: %v", err)
	}
	if result.Data.ID != "SHOT_123" {
		t.Fatalf("expected screenshot ID SHOT_123, got %s", result.Data.ID)
	}
}

func TestCreateAppScreenshot(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"appScreenshots","id":"SHOT_123","attributes":{"fileName":"shot.png","fileSize":1024,"uploadOperations":[{"method":"PUT","url":"https://example.com/upload","length":1024,"offset":0}]}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appScreenshots" {
			t.Fatalf("expected path /v1/appScreenshots, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.CreateAppScreenshot(context.Background(), "SET_123", "shot.png", 1024)
	if err != nil {
		t.Fatalf("CreateAppScreenshot() error: %v", err)
	}
	if result.Data.ID != "SHOT_123" {
		t.Fatalf("expected screenshot ID SHOT_123, got %s", result.Data.ID)
	}
}

func TestUpdateAppScreenshot(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"appScreenshots","id":"SHOT_123","attributes":{"uploaded":true}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appScreenshots/SHOT_123" {
			t.Fatalf("expected path /v1/appScreenshots/SHOT_123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	uploaded := true
	result, err := client.UpdateAppScreenshot(context.Background(), "SHOT_123", uploaded, &Checksum{
		Hash:      "abc123",
		Algorithm: ChecksumAlgorithmMD5,
	})
	if err != nil {
		t.Fatalf("UpdateAppScreenshot() error: %v", err)
	}
	if result.Data.ID != "SHOT_123" {
		t.Fatalf("expected screenshot ID SHOT_123, got %s", result.Data.ID)
	}
}

func TestDeleteAppScreenshot(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, "")
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appScreenshots/SHOT_123" {
			t.Fatalf("expected path /v1/appScreenshots/SHOT_123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.DeleteAppScreenshot(context.Background(), "SHOT_123"); err != nil {
		t.Fatalf("DeleteAppScreenshot() error: %v", err)
	}
}

func TestGetAppPreviewSets(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"appPreviewSets","id":"SET_123","attributes":{"previewType":"IPHONE_65"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersionLocalizations/LOC_123/appPreviewSets" {
			t.Fatalf("expected path /v1/appStoreVersionLocalizations/LOC_123/appPreviewSets, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.GetAppPreviewSets(context.Background(), "LOC_123")
	if err != nil {
		t.Fatalf("GetAppPreviewSets() error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 set, got %d", len(result.Data))
	}
}

func TestCreateAppPreviewSet(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"appPreviewSets","id":"SET_123","attributes":{"previewType":"IPHONE_65"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appPreviewSets" {
			t.Fatalf("expected path /v1/appPreviewSets, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.CreateAppPreviewSet(context.Background(), "LOC_123", "IPHONE_65")
	if err != nil {
		t.Fatalf("CreateAppPreviewSet() error: %v", err)
	}
	if result.Data.ID != "SET_123" {
		t.Fatalf("expected set ID SET_123, got %s", result.Data.ID)
	}
}

func TestDeleteAppPreviewSet(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, "")
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appPreviewSets/SET_123" {
			t.Fatalf("expected path /v1/appPreviewSets/SET_123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.DeleteAppPreviewSet(context.Background(), "SET_123"); err != nil {
		t.Fatalf("DeleteAppPreviewSet() error: %v", err)
	}
}

func TestGetAppPreviews(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"appPreviews","id":"PREVIEW_123","attributes":{"fileName":"preview.mov","fileSize":2048}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appPreviewSets/SET_123/appPreviews" {
			t.Fatalf("expected path /v1/appPreviewSets/SET_123/appPreviews, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.GetAppPreviews(context.Background(), "SET_123")
	if err != nil {
		t.Fatalf("GetAppPreviews() error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 preview, got %d", len(result.Data))
	}
}

func TestGetAppPreview(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"appPreviews","id":"PREVIEW_123","attributes":{"fileName":"preview.mov","fileSize":2048}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appPreviews/PREVIEW_123" {
			t.Fatalf("expected path /v1/appPreviews/PREVIEW_123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.GetAppPreview(context.Background(), "PREVIEW_123")
	if err != nil {
		t.Fatalf("GetAppPreview() error: %v", err)
	}
	if result.Data.ID != "PREVIEW_123" {
		t.Fatalf("expected preview ID PREVIEW_123, got %s", result.Data.ID)
	}
}

func TestCreateAppPreview(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"appPreviews","id":"PREVIEW_123","attributes":{"fileName":"preview.mov","fileSize":2048,"mimeType":"video/quicktime","uploadOperations":[{"method":"PUT","url":"https://example.com/upload","length":2048,"offset":0}]}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appPreviews" {
			t.Fatalf("expected path /v1/appPreviews, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.CreateAppPreview(context.Background(), "SET_123", "preview.mov", 2048, "video/quicktime")
	if err != nil {
		t.Fatalf("CreateAppPreview() error: %v", err)
	}
	if result.Data.ID != "PREVIEW_123" {
		t.Fatalf("expected preview ID PREVIEW_123, got %s", result.Data.ID)
	}
}

func TestUpdateAppPreview(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"appPreviews","id":"PREVIEW_123","attributes":{"uploaded":true}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appPreviews/PREVIEW_123" {
			t.Fatalf("expected path /v1/appPreviews/PREVIEW_123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	uploaded := true
	result, err := client.UpdateAppPreview(context.Background(), "PREVIEW_123", uploaded, &Checksum{
		Hash:      "def456",
		Algorithm: ChecksumAlgorithmMD5,
	})
	if err != nil {
		t.Fatalf("UpdateAppPreview() error: %v", err)
	}
	if result.Data.ID != "PREVIEW_123" {
		t.Fatalf("expected preview ID PREVIEW_123, got %s", result.Data.ID)
	}
}

func TestDeleteAppPreview(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, "")
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appPreviews/PREVIEW_123" {
			t.Fatalf("expected path /v1/appPreviews/PREVIEW_123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.DeleteAppPreview(context.Background(), "PREVIEW_123"); err != nil {
		t.Fatalf("DeleteAppPreview() error: %v", err)
	}
}

func TestCreateAppStoreVersionSubmission(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"appStoreVersionSubmissions","id":"SUBMIT_123","attributes":{"createdDate":"2026-01-20T00:00:00Z"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersionSubmissions" {
			t.Fatalf("expected path /v1/appStoreVersionSubmissions, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.CreateAppStoreVersionSubmission(context.Background(), AppStoreVersionSubmissionCreateRequest{
		Data: AppStoreVersionSubmissionCreateData{
			Type: ResourceTypeAppStoreVersionSubmissions,
			Relationships: &AppStoreVersionSubmissionRelationships{
				AppStoreVersion: &Relationship{
					Data: ResourceData{Type: ResourceTypeAppStoreVersions, ID: "VERSION_123"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateAppStoreVersionSubmission() error: %v", err)
	}
	if result.Data.ID != "SUBMIT_123" {
		t.Fatalf("expected submission ID SUBMIT_123, got %s", result.Data.ID)
	}
}

func TestGetAppStoreVersionSubmission(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"appStoreVersionSubmissions","id":"SUBMIT_123","attributes":{"createdDate":"2026-01-20T00:00:00Z"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersionSubmissions/SUBMIT_123" {
			t.Fatalf("expected path /v1/appStoreVersionSubmissions/SUBMIT_123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	result, err := client.GetAppStoreVersionSubmission(context.Background(), "SUBMIT_123")
	if err != nil {
		t.Fatalf("GetAppStoreVersionSubmission() error: %v", err)
	}
	if result.Data.ID != "SUBMIT_123" {
		t.Fatalf("expected submission ID SUBMIT_123, got %s", result.Data.ID)
	}
}

func TestDeleteAppStoreVersionSubmission(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, "")
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appStoreVersionSubmissions/SUBMIT_123" {
			t.Fatalf("expected path /v1/appStoreVersionSubmissions/SUBMIT_123, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	err := client.DeleteAppStoreVersionSubmission(context.Background(), "SUBMIT_123")
	if err != nil {
		t.Fatalf("DeleteAppStoreVersionSubmission() error: %v", err)
	}
}

func TestBuildUploadMethods_ErrorResponse(t *testing.T) {
	ctx := context.Background()
	errorBody := `{"errors":[{"code":"BAD_REQUEST","title":"Bad Request","detail":"nope"}]}`

	tests := []struct {
		name string
		call func(*Client) error
	}{
		{
			name: "CreateBuildUpload",
			call: func(c *Client) error {
				_, err := c.CreateBuildUpload(ctx, BuildUploadCreateRequest{
					Data: BuildUploadCreateData{
						Type: ResourceTypeBuildUploads,
						Attributes: BuildUploadAttributes{
							CFBundleShortVersionString: "1.0.0",
							CFBundleVersion:            "123",
							Platform:                   PlatformIOS,
						},
						Relationships: &BuildUploadRelationships{
							App: &Relationship{
								Data: ResourceData{Type: ResourceTypeApps, ID: "APP_123"},
							},
						},
					},
				})
				return err
			},
		},
		{
			name: "GetBuildUpload",
			call: func(c *Client) error {
				_, err := c.GetBuildUpload(ctx, "UPLOAD_123")
				return err
			},
		},
		{
			name: "CreateBuildUploadFile",
			call: func(c *Client) error {
				_, err := c.CreateBuildUploadFile(ctx, BuildUploadFileCreateRequest{
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
								Data: ResourceData{Type: ResourceTypeBuildUploads, ID: "UPLOAD_123"},
							},
						},
					},
				})
				return err
			},
		},
		{
			name: "UpdateBuildUploadFile",
			call: func(c *Client) error {
				uploaded := true
				_, err := c.UpdateBuildUploadFile(ctx, "FILE_123", BuildUploadFileUpdateRequest{
					Data: BuildUploadFileUpdateData{
						Type: ResourceTypeBuildUploadFiles,
						ID:   "FILE_123",
						Attributes: &BuildUploadFileUpdateAttributes{
							Uploaded: &uploaded,
						},
					},
				})
				return err
			},
		},
		{
			name: "CreateAppStoreVersionSubmission",
			call: func(c *Client) error {
				_, err := c.CreateAppStoreVersionSubmission(ctx, AppStoreVersionSubmissionCreateRequest{
					Data: AppStoreVersionSubmissionCreateData{
						Type: ResourceTypeAppStoreVersionSubmissions,
						Relationships: &AppStoreVersionSubmissionRelationships{
							AppStoreVersion: &Relationship{
								Data: ResourceData{Type: ResourceTypeAppStoreVersions, ID: "VERSION_123"},
							},
						},
					},
				})
				return err
			},
		},
		{
			name: "GetAppStoreVersionSubmission",
			call: func(c *Client) error {
				_, err := c.GetAppStoreVersionSubmission(ctx, "SUBMIT_123")
				return err
			},
		},
		{
			name: "DeleteAppStoreVersionSubmission",
			call: func(c *Client) error {
				return c.DeleteAppStoreVersionSubmission(ctx, "SUBMIT_123")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := newTestClient(t, nil, jsonResponse(http.StatusBadRequest, errorBody))
			if err := test.call(client); err == nil {
				t.Fatalf("expected error")
			} else if !strings.Contains(err.Error(), "Bad Request") {
				t.Fatalf("expected error to contain title, got %v", err)
			}
		})
	}
}

func TestGetCiProducts_WithAppFilterAndLimit(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"ciProducts","id":"prod-1"}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/ciProducts" {
			t.Fatalf("expected path /v1/ciProducts, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[app]") != "app-1" {
			t.Fatalf("expected filter[app]=app-1, got %q", values.Get("filter[app]"))
		}
		if values.Get("limit") != "25" {
			t.Fatalf("expected limit=25, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetCiProducts(context.Background(), WithCiProductsAppID("app-1"), WithCiProductsLimit(25)); err != nil {
		t.Fatalf("GetCiProducts() error: %v", err)
	}
}

func TestGetCiWorkflows_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/ciProducts/prod-1/workflows?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetCiWorkflows(context.Background(), "prod-1", WithCiWorkflowsNextURL(next)); err != nil {
		t.Fatalf("GetCiWorkflows() error: %v", err)
	}
}

func TestGetScmGitReferences_WithLimit(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"scmGitReferences","id":"ref-1"}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/scmRepositories/repo-1/gitReferences" {
			t.Fatalf("expected path /v1/scmRepositories/repo-1/gitReferences, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("limit") != "100" {
			t.Fatalf("expected limit=100, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetScmGitReferences(context.Background(), "repo-1", WithScmGitReferencesLimit(100)); err != nil {
		t.Fatalf("GetScmGitReferences() error: %v", err)
	}
}

func TestGetCiBuildRuns_WithLimit(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"ciBuildRuns","id":"run-1","attributes":{"number":1}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/ciWorkflows/wf-1/buildRuns" {
			t.Fatalf("expected path /v1/ciWorkflows/wf-1/buildRuns, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("limit") != "50" {
			t.Fatalf("expected limit=50, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetCiBuildRuns(context.Background(), "wf-1", WithCiBuildRunsLimit(50)); err != nil {
		t.Fatalf("GetCiBuildRuns() error: %v", err)
	}
}

func TestGetCiBuildRun(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"ciBuildRuns","id":"run-1","attributes":{"number":1}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/ciBuildRuns/run-1" {
			t.Fatalf("expected path /v1/ciBuildRuns/run-1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetCiBuildRun(context.Background(), "run-1"); err != nil {
		t.Fatalf("GetCiBuildRun() error: %v", err)
	}
}

func TestCreateCiBuildRun(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"ciBuildRuns","id":"run-1","attributes":{"number":1}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/ciBuildRuns" {
			t.Fatalf("expected path /v1/ciBuildRuns, got %s", req.URL.Path)
		}
		var payload CiBuildRunCreateRequest
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if payload.Data.Type != ResourceTypeCiBuildRuns {
			t.Fatalf("expected type ciBuildRuns, got %q", payload.Data.Type)
		}
		if payload.Data.Relationships == nil || payload.Data.Relationships.Workflow == nil || payload.Data.Relationships.SourceBranchOrTag == nil {
			t.Fatalf("expected workflow and sourceBranchOrTag relationships")
		}
		assertAuthorized(t, req)
	}, response)

	req := CiBuildRunCreateRequest{
		Data: CiBuildRunCreateData{
			Type: ResourceTypeCiBuildRuns,
			Relationships: &CiBuildRunCreateRelationships{
				Workflow: &Relationship{
					Data: ResourceData{Type: ResourceTypeCiWorkflows, ID: "wf-1"},
				},
				SourceBranchOrTag: &Relationship{
					Data: ResourceData{Type: ResourceTypeScmGitReferences, ID: "ref-1"},
				},
			},
		},
	}
	if _, err := client.CreateCiBuildRun(context.Background(), req); err != nil {
		t.Fatalf("CreateCiBuildRun() error: %v", err)
	}
}

func TestResolveCiWorkflowByName_CaseInsensitive(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"ciWorkflows","id":"wf-1","attributes":{"name":"CI Build"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.Path != "/v1/ciProducts/prod-1/workflows" {
			t.Fatalf("expected path /v1/ciProducts/prod-1/workflows, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	workflow, err := client.ResolveCiWorkflowByName(context.Background(), "prod-1", "ci build")
	if err != nil {
		t.Fatalf("ResolveCiWorkflowByName() error: %v", err)
	}
	if workflow.ID != "wf-1" {
		t.Fatalf("expected workflow ID wf-1, got %q", workflow.ID)
	}
}

func TestResolveCiWorkflowByName_NoMatch(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"ciWorkflows","id":"wf-1","attributes":{"name":"Deploy"}}]}`)
	client := newTestClient(t, nil, response)

	if _, err := client.ResolveCiWorkflowByName(context.Background(), "prod-1", "ci"); err == nil {
		t.Fatal("expected error")
	} else if !strings.Contains(err.Error(), "no workflow named") {
		t.Fatalf("expected no workflow named error, got %v", err)
	}
}

func TestResolveGitReferenceByName_CanonicalMatch(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"scmGitReferences","id":"ref-1","attributes":{"name":"main","canonicalName":"refs/heads/main","isDeleted":false}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.Path != "/v1/scmRepositories/repo-1/gitReferences" {
			t.Fatalf("expected path /v1/scmRepositories/repo-1/gitReferences, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	ref, err := client.ResolveGitReferenceByName(context.Background(), "repo-1", "main")
	if err != nil {
		t.Fatalf("ResolveGitReferenceByName() error: %v", err)
	}
	if ref.ID != "ref-1" {
		t.Fatalf("expected git reference ID ref-1, got %q", ref.ID)
	}
}

func TestResolveGitReferenceByName_SuffixMatchNotAllowed(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"scmGitReferences","id":"ref-1","attributes":{"name":"feature/main","canonicalName":"refs/heads/feature/main","isDeleted":false}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.Path != "/v1/scmRepositories/repo-1/gitReferences" {
			t.Fatalf("expected path /v1/scmRepositories/repo-1/gitReferences, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.ResolveGitReferenceByName(context.Background(), "repo-1", "main"); err == nil {
		t.Fatal("expected error")
	} else if !strings.Contains(err.Error(), "no git reference named") {
		t.Fatalf("expected no git reference named error, got %v", err)
	}
}

func TestResolveGitReferenceByName_NoMatch(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"scmGitReferences","id":"ref-1","attributes":{"name":"develop","canonicalName":"refs/heads/develop","isDeleted":false}}]}`)
	client := newTestClient(t, nil, response)

	if _, err := client.ResolveGitReferenceByName(context.Background(), "repo-1", "main"); err == nil {
		t.Fatal("expected error")
	} else if !strings.Contains(err.Error(), "no git reference named") {
		t.Fatalf("expected no git reference named error, got %v", err)
	}
}
