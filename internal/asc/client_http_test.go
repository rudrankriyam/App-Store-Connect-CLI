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

func TestGetBetaTesters_WithFilters(t *testing.T) {
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
