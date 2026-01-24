package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// PhasedReleaseState represents the state of a phased release.
type PhasedReleaseState string

const (
	PhasedReleaseStateInactive PhasedReleaseState = "INACTIVE"
	PhasedReleaseStateActive   PhasedReleaseState = "ACTIVE"
	PhasedReleaseStatePaused   PhasedReleaseState = "PAUSED"
	PhasedReleaseStateComplete PhasedReleaseState = "COMPLETE"
)

// ResourceTypeAppStoreVersionPhasedReleases is the resource type for phased releases.
const ResourceTypeAppStoreVersionPhasedReleases ResourceType = "appStoreVersionPhasedReleases"

// AppStoreVersionPhasedReleaseAttributes represents the attributes of a phased release.
type AppStoreVersionPhasedReleaseAttributes struct {
	PhasedReleaseState PhasedReleaseState `json:"phasedReleaseState,omitempty"`
	StartDate          string             `json:"startDate,omitempty"`
	TotalPauseDuration int                `json:"totalPauseDuration,omitempty"`
	CurrentDayNumber   int                `json:"currentDayNumber,omitempty"`
}

// AppStoreVersionPhasedReleaseResponse represents an API response for a phased release.
type AppStoreVersionPhasedReleaseResponse struct {
	Data  Resource[AppStoreVersionPhasedReleaseAttributes] `json:"data"`
	Links Links                                            `json:"links,omitempty"`
}

// AppStoreVersionPhasedReleaseCreateRequest represents a request to create a phased release.
type AppStoreVersionPhasedReleaseCreateRequest struct {
	Data AppStoreVersionPhasedReleaseCreateData `json:"data"`
}

// AppStoreVersionPhasedReleaseCreateData represents the data for creating a phased release.
type AppStoreVersionPhasedReleaseCreateData struct {
	Type          ResourceType                                  `json:"type"`
	Attributes    *AppStoreVersionPhasedReleaseCreateAttributes `json:"attributes,omitempty"`
	Relationships AppStoreVersionPhasedReleaseRelationships     `json:"relationships"`
}

// AppStoreVersionPhasedReleaseCreateAttributes represents attributes for creating a phased release.
type AppStoreVersionPhasedReleaseCreateAttributes struct {
	PhasedReleaseState PhasedReleaseState `json:"phasedReleaseState,omitempty"`
}

// AppStoreVersionPhasedReleaseRelationships represents the relationships for a phased release.
type AppStoreVersionPhasedReleaseRelationships struct {
	AppStoreVersion *Relationship `json:"appStoreVersion"`
}

// AppStoreVersionPhasedReleaseUpdateRequest represents a request to update a phased release.
type AppStoreVersionPhasedReleaseUpdateRequest struct {
	Data AppStoreVersionPhasedReleaseUpdateData `json:"data"`
}

// AppStoreVersionPhasedReleaseUpdateData represents the data for updating a phased release.
type AppStoreVersionPhasedReleaseUpdateData struct {
	Type       ResourceType                                  `json:"type"`
	ID         string                                        `json:"id"`
	Attributes *AppStoreVersionPhasedReleaseUpdateAttributes `json:"attributes,omitempty"`
}

// AppStoreVersionPhasedReleaseUpdateAttributes represents attributes for updating a phased release.
type AppStoreVersionPhasedReleaseUpdateAttributes struct {
	PhasedReleaseState PhasedReleaseState `json:"phasedReleaseState,omitempty"`
}

// AppStoreVersionPhasedReleaseDeleteResult represents the result of deleting a phased release.
type AppStoreVersionPhasedReleaseDeleteResult struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// GetAppStoreVersionPhasedRelease fetches the phased release for an app store version.
func (c *Client) GetAppStoreVersionPhasedRelease(ctx context.Context, versionID string) (*AppStoreVersionPhasedReleaseResponse, error) {
	path := fmt.Sprintf("/v1/appStoreVersions/%s/appStoreVersionPhasedRelease", versionID)

	data, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response AppStoreVersionPhasedReleaseResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreateAppStoreVersionPhasedRelease creates a phased release for an app store version.
// If state is empty, defaults to INACTIVE per API spec.
func (c *Client) CreateAppStoreVersionPhasedRelease(ctx context.Context, versionID string, state PhasedReleaseState) (*AppStoreVersionPhasedReleaseResponse, error) {
	// Default to INACTIVE if no state provided (API spec default)
	if state == "" {
		state = PhasedReleaseStateInactive
	}

	payload := AppStoreVersionPhasedReleaseCreateRequest{
		Data: AppStoreVersionPhasedReleaseCreateData{
			Type: ResourceTypeAppStoreVersionPhasedReleases,
			Attributes: &AppStoreVersionPhasedReleaseCreateAttributes{
				PhasedReleaseState: state,
			},
			Relationships: AppStoreVersionPhasedReleaseRelationships{
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

	data, err := c.do(ctx, http.MethodPost, "/v1/appStoreVersionPhasedReleases", body)
	if err != nil {
		return nil, err
	}

	var response AppStoreVersionPhasedReleaseResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// UpdateAppStoreVersionPhasedRelease updates a phased release.
// State must be provided (ACTIVE, PAUSED, or COMPLETE).
func (c *Client) UpdateAppStoreVersionPhasedRelease(ctx context.Context, phasedReleaseID string, state PhasedReleaseState) (*AppStoreVersionPhasedReleaseResponse, error) {
	if state == "" {
		return nil, fmt.Errorf("state is required for update")
	}

	payload := AppStoreVersionPhasedReleaseUpdateRequest{
		Data: AppStoreVersionPhasedReleaseUpdateData{
			Type: ResourceTypeAppStoreVersionPhasedReleases,
			ID:   phasedReleaseID,
			Attributes: &AppStoreVersionPhasedReleaseUpdateAttributes{
				PhasedReleaseState: state,
			},
		},
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/appStoreVersionPhasedReleases/%s", phasedReleaseID)

	data, err := c.do(ctx, http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}

	var response AppStoreVersionPhasedReleaseResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// DeleteAppStoreVersionPhasedRelease deletes a phased release.
func (c *Client) DeleteAppStoreVersionPhasedRelease(ctx context.Context, phasedReleaseID string) error {
	path := fmt.Sprintf("/v1/appStoreVersionPhasedReleases/%s", phasedReleaseID)

	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
