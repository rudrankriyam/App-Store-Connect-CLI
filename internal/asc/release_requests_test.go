package asc

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCreateAppStoreVersionReleaseRequest(t *testing.T) {
	resp := AppStoreVersionReleaseRequestResponse{
		Data: AppStoreVersionReleaseRequest{
			Type: ResourceTypeAppStoreVersionReleaseRequests,
			ID:   "release-123",
		},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", req.Method)
		}
		if !strings.HasSuffix(req.URL.Path, "/v1/appStoreVersionReleaseRequests") {
			t.Errorf("unexpected path: %s", req.URL.Path)
		}

		var createReq AppStoreVersionReleaseRequestCreateRequest
		if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if createReq.Data.Type != ResourceTypeAppStoreVersionReleaseRequests {
			t.Errorf("expected type %s, got %s", ResourceTypeAppStoreVersionReleaseRequests, createReq.Data.Type)
		}
		if createReq.Data.Relationships.AppStoreVersion.Data.Type != ResourceTypeAppStoreVersions {
			t.Errorf("expected version type %s, got %s", ResourceTypeAppStoreVersions, createReq.Data.Relationships.AppStoreVersion.Data.Type)
		}
		if createReq.Data.Relationships.AppStoreVersion.Data.ID != "version-123" {
			t.Errorf("expected version ID version-123, got %s", createReq.Data.Relationships.AppStoreVersion.Data.ID)
		}
	}, jsonResponse(http.StatusCreated, string(body)))

	result, err := client.CreateAppStoreVersionReleaseRequest(context.Background(), "version-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Data.ID != "release-123" {
		t.Errorf("expected ID release-123, got %s", result.Data.ID)
	}
}

func TestCreateAppStoreVersionReleaseRequestRequiresVersionID(t *testing.T) {
	client := newTestClient(t, func(req *http.Request) {
		t.Fatal("unexpected request")
	}, jsonResponse(http.StatusOK, `{"data":{}}`))

	_, err := client.CreateAppStoreVersionReleaseRequest(context.Background(), " ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
