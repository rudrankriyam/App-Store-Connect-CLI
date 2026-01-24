package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

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

// BuildsResponse is the response from builds endpoint.
type BuildsResponse = Response[BuildAttributes]

// BuildResponse is the response from build detail/updates.
type BuildResponse = SingleResponse[BuildAttributes]

// BuildUploadAttributes describes a build upload resource.
type BuildUploadAttributes struct {
	CFBundleShortVersionString string              `json:"cfBundleShortVersionString"`
	CFBundleVersion            string              `json:"cfBundleVersion"`
	Platform                   Platform            `json:"platform"`
	CreatedDate                *string             `json:"createdDate,omitempty"`
	State                      *AppMediaAssetState `json:"state,omitempty"`
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

// GetBuilds retrieves the list of builds for an app
func (c *Client) GetBuilds(ctx context.Context, appID string, opts ...BuildsOption) (*BuildsResponse, error) {
	query := &buildsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/apps/%s/builds", appID)
	if query.nextURL != "" {
		// Validate nextURL to prevent credential exfiltration
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("builds: %w", err)
		}
		path = query.nextURL
	} else {
		values := url.Values{}
		// Use /v1/builds endpoint when sorting, limiting, or filtering by preReleaseVersion,
		// since /v1/apps/{id}/builds doesn't support these
		if query.sort != "" || query.limit > 0 || query.preReleaseVersionID != "" {
			path = "/v1/builds"
			values.Set("filter[app]", appID)
			if query.sort != "" {
				values.Set("sort", query.sort)
			}
			if query.limit > 0 {
				values.Set("limit", strconv.Itoa(query.limit))
			}
			if query.preReleaseVersionID != "" {
				values.Set("filter[preReleaseVersion]", query.preReleaseVersionID)
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

// GetBuildAppStoreVersion retrieves the app store version for a build.
func (c *Client) GetBuildAppStoreVersion(ctx context.Context, buildID string) (*AppStoreVersionResponse, error) {
	path := fmt.Sprintf("/v1/builds/%s/appStoreVersion", buildID)
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

// AddBetaGroupsToBuild adds beta groups to a build for TestFlight distribution.
func (c *Client) AddBetaGroupsToBuild(ctx context.Context, buildID string, groupIDs []string) error {
	return c.AddBetaGroupsToBuildWithNotify(ctx, buildID, groupIDs, false)
}

// AddBetaGroupsToBuildWithNotify adds beta groups to a build with optional notifications.
func (c *Client) AddBetaGroupsToBuildWithNotify(ctx context.Context, buildID string, groupIDs []string, notify bool) error {
	payload := RelationshipRequest{
		Data: make([]RelationshipData, len(groupIDs)),
	}
	for i, id := range groupIDs {
		payload.Data[i] = RelationshipData{
			Type: ResourceTypeBetaGroups,
			ID:   id,
		}
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v1/builds/%s/relationships/betaGroups", buildID)
	if notify {
		path += "?notify=true"
	}
	if _, err := c.do(ctx, "POST", path, body); err != nil {
		return err
	}
	return nil
}

// RemoveBetaGroupsFromBuild removes beta groups from a build.
func (c *Client) RemoveBetaGroupsFromBuild(ctx context.Context, buildID string, groupIDs []string) error {
	payload := RelationshipRequest{
		Data: make([]RelationshipData, len(groupIDs)),
	}
	for i, id := range groupIDs {
		payload.Data[i] = RelationshipData{
			Type: ResourceTypeBetaGroups,
			ID:   id,
		}
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v1/builds/%s/relationships/betaGroups", buildID)
	if _, err := c.do(ctx, "DELETE", path, body); err != nil {
		return err
	}
	return nil
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
