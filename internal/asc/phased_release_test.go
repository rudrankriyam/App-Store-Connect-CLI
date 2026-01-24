package asc

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestGetAppStoreVersionPhasedRelease(t *testing.T) {
	resp := AppStoreVersionPhasedReleaseResponse{
		Data: Resource[AppStoreVersionPhasedReleaseAttributes]{
			Type: "appStoreVersionPhasedReleases",
			ID:   "phased-123",
			Attributes: AppStoreVersionPhasedReleaseAttributes{
				PhasedReleaseState: PhasedReleaseStateActive,
				CurrentDayNumber:   3,
				StartDate:          "2026-01-20",
				TotalPauseDuration: 0,
			},
		},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", req.Method)
		}
		if !strings.HasSuffix(req.URL.Path, "/v1/appStoreVersions/version-123/appStoreVersionPhasedRelease") {
			t.Errorf("unexpected path: %s", req.URL.Path)
		}
	}, jsonResponse(http.StatusOK, string(body)))

	result, err := client.GetAppStoreVersionPhasedRelease(context.Background(), "version-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Data.ID != "phased-123" {
		t.Errorf("expected ID phased-123, got %s", result.Data.ID)
	}
	if result.Data.Attributes.PhasedReleaseState != PhasedReleaseStateActive {
		t.Errorf("expected state ACTIVE, got %s", result.Data.Attributes.PhasedReleaseState)
	}
	if result.Data.Attributes.CurrentDayNumber != 3 {
		t.Errorf("expected day 3, got %d", result.Data.Attributes.CurrentDayNumber)
	}
}

func TestCreateAppStoreVersionPhasedRelease(t *testing.T) {
	resp := AppStoreVersionPhasedReleaseResponse{
		Data: Resource[AppStoreVersionPhasedReleaseAttributes]{
			Type: "appStoreVersionPhasedReleases",
			ID:   "phased-new",
			Attributes: AppStoreVersionPhasedReleaseAttributes{
				PhasedReleaseState: PhasedReleaseStateInactive,
			},
		},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", req.Method)
		}
		if !strings.HasSuffix(req.URL.Path, "/v1/appStoreVersionPhasedReleases") {
			t.Errorf("unexpected path: %s", req.URL.Path)
		}

		var createReq AppStoreVersionPhasedReleaseCreateRequest
		if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if createReq.Data.Type != "appStoreVersionPhasedReleases" {
			t.Errorf("expected type appStoreVersionPhasedReleases, got %s", createReq.Data.Type)
		}
		if createReq.Data.Relationships.AppStoreVersion.Data.ID != "version-123" {
			t.Errorf("expected version ID version-123, got %s", createReq.Data.Relationships.AppStoreVersion.Data.ID)
		}
		// When no state provided, should default to INACTIVE
		if createReq.Data.Attributes == nil {
			t.Fatal("expected attributes to be set (default to INACTIVE)")
		}
		if createReq.Data.Attributes.PhasedReleaseState != PhasedReleaseStateInactive {
			t.Errorf("expected default state INACTIVE, got %s", createReq.Data.Attributes.PhasedReleaseState)
		}
	}, jsonResponse(http.StatusCreated, string(body)))

	result, err := client.CreateAppStoreVersionPhasedRelease(context.Background(), "version-123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Data.ID != "phased-new" {
		t.Errorf("expected ID phased-new, got %s", result.Data.ID)
	}
}

func TestCreateAppStoreVersionPhasedRelease_WithState(t *testing.T) {
	resp := AppStoreVersionPhasedReleaseResponse{
		Data: Resource[AppStoreVersionPhasedReleaseAttributes]{
			Type: "appStoreVersionPhasedReleases",
			ID:   "phased-new",
			Attributes: AppStoreVersionPhasedReleaseAttributes{
				PhasedReleaseState: PhasedReleaseStateActive,
			},
		},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)

		var createReq AppStoreVersionPhasedReleaseCreateRequest
		if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if createReq.Data.Attributes == nil {
			t.Fatal("expected attributes to be set")
		}
		if createReq.Data.Attributes.PhasedReleaseState != PhasedReleaseStateActive {
			t.Errorf("expected state ACTIVE, got %s", createReq.Data.Attributes.PhasedReleaseState)
		}
	}, jsonResponse(http.StatusCreated, string(body)))

	result, err := client.CreateAppStoreVersionPhasedRelease(context.Background(), "version-123", PhasedReleaseStateActive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Data.Attributes.PhasedReleaseState != PhasedReleaseStateActive {
		t.Errorf("expected state ACTIVE, got %s", result.Data.Attributes.PhasedReleaseState)
	}
}

func TestUpdateAppStoreVersionPhasedRelease(t *testing.T) {
	resp := AppStoreVersionPhasedReleaseResponse{
		Data: Resource[AppStoreVersionPhasedReleaseAttributes]{
			Type: "appStoreVersionPhasedReleases",
			ID:   "phased-123",
			Attributes: AppStoreVersionPhasedReleaseAttributes{
				PhasedReleaseState: PhasedReleaseStatePaused,
				CurrentDayNumber:   3,
			},
		},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", req.Method)
		}
		if !strings.HasSuffix(req.URL.Path, "/v1/appStoreVersionPhasedReleases/phased-123") {
			t.Errorf("unexpected path: %s", req.URL.Path)
		}

		var updateReq AppStoreVersionPhasedReleaseUpdateRequest
		if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if updateReq.Data.ID != "phased-123" {
			t.Errorf("expected ID phased-123, got %s", updateReq.Data.ID)
		}
		if updateReq.Data.Attributes.PhasedReleaseState != PhasedReleaseStatePaused {
			t.Errorf("expected state PAUSED, got %s", updateReq.Data.Attributes.PhasedReleaseState)
		}
	}, jsonResponse(http.StatusOK, string(body)))

	result, err := client.UpdateAppStoreVersionPhasedRelease(context.Background(), "phased-123", PhasedReleaseStatePaused)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Data.Attributes.PhasedReleaseState != PhasedReleaseStatePaused {
		t.Errorf("expected state PAUSED, got %s", result.Data.Attributes.PhasedReleaseState)
	}
}

func TestDeleteAppStoreVersionPhasedRelease(t *testing.T) {
	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", req.Method)
		}
		if !strings.HasSuffix(req.URL.Path, "/v1/appStoreVersionPhasedReleases/phased-123") {
			t.Errorf("unexpected path: %s", req.URL.Path)
		}
	}, jsonResponse(http.StatusNoContent, ""))

	err := client.DeleteAppStoreVersionPhasedRelease(context.Background(), "phased-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateAppStoreVersionPhasedRelease_EmptyState(t *testing.T) {
	client := newTestClient(t, nil, nil)

	_, err := client.UpdateAppStoreVersionPhasedRelease(context.Background(), "phased-123", "")
	if err == nil {
		t.Fatal("expected error for empty state")
	}
	if !strings.Contains(err.Error(), "state is required") {
		t.Errorf("expected 'state is required' error, got: %v", err)
	}
}

func TestPhasedReleaseState_Values(t *testing.T) {
	// Verify enum values match API spec
	tests := []struct {
		state PhasedReleaseState
		want  string
	}{
		{PhasedReleaseStateInactive, "INACTIVE"},
		{PhasedReleaseStateActive, "ACTIVE"},
		{PhasedReleaseStatePaused, "PAUSED"},
		{PhasedReleaseStateComplete, "COMPLETE"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.state) != tt.want {
				t.Errorf("expected %s, got %s", tt.want, tt.state)
			}
		})
	}
}
