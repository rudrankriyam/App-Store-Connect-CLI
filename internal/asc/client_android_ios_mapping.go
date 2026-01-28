package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GetAndroidToIosAppMappingDetails retrieves mappings for an app.
func (c *Client) GetAndroidToIosAppMappingDetails(ctx context.Context, appID string, opts ...AndroidToIosAppMappingDetailsOption) (*AndroidToIosAppMappingDetailsResponse, error) {
	appID = strings.TrimSpace(appID)
	if appID == "" {
		return nil, fmt.Errorf("app ID is required")
	}

	query := &androidToIosAppMappingDetailsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/apps/%s/androidToIosAppMappingDetails", appID)
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("androidToIosAppMappingDetails: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildAndroidToIosAppMappingDetailsQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response AndroidToIosAppMappingDetailsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetAndroidToIosAppMappingDetail retrieves a mapping by ID.
func (c *Client) GetAndroidToIosAppMappingDetail(ctx context.Context, id string, opts ...AndroidToIosAppMappingDetailsOption) (*AndroidToIosAppMappingDetailResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("mapping ID is required")
	}

	query := &androidToIosAppMappingDetailsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/androidToIosAppMappingDetails/%s", id)
	if queryString := buildAndroidToIosAppMappingDetailsQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response AndroidToIosAppMappingDetailResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreateAndroidToIosAppMappingDetail creates a new mapping.
func (c *Client) CreateAndroidToIosAppMappingDetail(ctx context.Context, appID string, attrs AndroidToIosAppMappingDetailCreateAttributes) (*AndroidToIosAppMappingDetailResponse, error) {
	appID = strings.TrimSpace(appID)
	if appID == "" {
		return nil, fmt.Errorf("app ID is required")
	}

	request := AndroidToIosAppMappingDetailCreateRequest{
		Data: AndroidToIosAppMappingDetailCreateData{
			Type:       ResourceTypeAndroidToIosAppMappingDetails,
			Attributes: attrs,
			Relationships: AndroidToIosAppMappingDetailCreateRelationships{
				App: &Relationship{
					Data: ResourceData{
						Type: ResourceTypeApps,
						ID:   appID,
					},
				},
			},
		},
	}

	body, err := BuildRequestBody(request)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/androidToIosAppMappingDetails", body)
	if err != nil {
		return nil, err
	}

	var response AndroidToIosAppMappingDetailResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// UpdateAndroidToIosAppMappingDetail updates an existing mapping.
func (c *Client) UpdateAndroidToIosAppMappingDetail(ctx context.Context, id string, attrs AndroidToIosAppMappingDetailUpdateAttributes) (*AndroidToIosAppMappingDetailResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("mapping ID is required")
	}

	request := AndroidToIosAppMappingDetailUpdateRequest{
		Data: AndroidToIosAppMappingDetailUpdateData{
			Type:       ResourceTypeAndroidToIosAppMappingDetails,
			ID:         id,
			Attributes: &attrs,
		},
	}

	body, err := BuildRequestBody(request)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/androidToIosAppMappingDetails/%s", id)
	data, err := c.do(ctx, "PATCH", path, body)
	if err != nil {
		return nil, err
	}

	var response AndroidToIosAppMappingDetailResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// DeleteAndroidToIosAppMappingDetail deletes a mapping by ID.
func (c *Client) DeleteAndroidToIosAppMappingDetail(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("mapping ID is required")
	}

	path := fmt.Sprintf("/v1/androidToIosAppMappingDetails/%s", id)
	_, err := c.do(ctx, "DELETE", path, nil)
	return err
}
