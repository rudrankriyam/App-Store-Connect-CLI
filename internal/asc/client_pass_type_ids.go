package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GetPassTypeIDs retrieves the list of pass type IDs.
func (c *Client) GetPassTypeIDs(ctx context.Context, opts ...PassTypeIDsOption) (*PassTypeIDsResponse, error) {
	query := &passTypeIDsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := "/v1/passTypeIds"
	if query.nextURL != "" {
		// Validate nextURL to prevent credential exfiltration
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("passTypeIds: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildPassTypeIDsQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response PassTypeIDsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetPassTypeID retrieves a single pass type ID by ID.
func (c *Client) GetPassTypeID(ctx context.Context, id string) (*PassTypeIDResponse, error) {
	id = strings.TrimSpace(id)
	path := fmt.Sprintf("/v1/passTypeIds/%s", id)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response PassTypeIDResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreatePassTypeID creates a new pass type ID.
func (c *Client) CreatePassTypeID(ctx context.Context, attrs PassTypeIDCreateAttributes) (*PassTypeIDResponse, error) {
	request := PassTypeIDCreateRequest{
		Data: PassTypeIDCreateData{
			Type:       ResourceTypePassTypeIds,
			Attributes: attrs,
		},
	}

	body, err := BuildRequestBody(request)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/passTypeIds", body)
	if err != nil {
		return nil, err
	}

	var response PassTypeIDResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// UpdatePassTypeID updates an existing pass type ID.
func (c *Client) UpdatePassTypeID(ctx context.Context, id string, attrs PassTypeIDUpdateAttributes) (*PassTypeIDResponse, error) {
	id = strings.TrimSpace(id)
	request := PassTypeIDUpdateRequest{
		Data: PassTypeIDUpdateData{
			Type:       ResourceTypePassTypeIds,
			ID:         id,
			Attributes: &attrs,
		},
	}

	body, err := BuildRequestBody(request)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "PATCH", fmt.Sprintf("/v1/passTypeIds/%s", id), body)
	if err != nil {
		return nil, err
	}

	var response PassTypeIDResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// DeletePassTypeID deletes a pass type ID by ID.
func (c *Client) DeletePassTypeID(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	path := fmt.Sprintf("/v1/passTypeIds/%s", id)
	_, err := c.do(ctx, "DELETE", path, nil)
	return err
}

// GetPassTypeIDCertificates retrieves certificates for a pass type ID.
func (c *Client) GetPassTypeIDCertificates(ctx context.Context, passTypeID string, opts ...PassTypeIDCertificatesOption) (*CertificatesResponse, error) {
	passTypeID = strings.TrimSpace(passTypeID)
	query := &passTypeIDCertificatesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/passTypeIds/%s/certificates", passTypeID)
	if query.nextURL != "" {
		// Validate nextURL to prevent credential exfiltration
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("passTypeIdCertificates: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildPassTypeIDCertificatesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response CertificatesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetPassTypeIDCertificatesRelationships retrieves certificate linkages for a pass type ID.
func (c *Client) GetPassTypeIDCertificatesRelationships(ctx context.Context, passTypeID string, opts ...LinkagesOption) (*PassTypeIDCertificatesLinkagesResponse, error) {
	passTypeID = strings.TrimSpace(passTypeID)
	query := &linkagesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/passTypeIds/%s/relationships/certificates", passTypeID)
	if query.nextURL != "" {
		// Validate nextURL to prevent credential exfiltration
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("passTypeIdCertificateRelationships: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildLinkagesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response PassTypeIDCertificatesLinkagesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetCertificatePassTypeID retrieves the pass type ID for a certificate.
func (c *Client) GetCertificatePassTypeID(ctx context.Context, certificateID string) (*PassTypeIDResponse, error) {
	certificateID = strings.TrimSpace(certificateID)
	path := fmt.Sprintf("/v1/certificates/%s/passTypeId", certificateID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response PassTypeIDResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetCertificatePassTypeIDRelationship retrieves the pass type ID linkage for a certificate.
func (c *Client) GetCertificatePassTypeIDRelationship(ctx context.Context, certificateID string) (*CertificatePassTypeIDLinkageResponse, error) {
	certificateID = strings.TrimSpace(certificateID)
	path := fmt.Sprintf("/v1/certificates/%s/relationships/passTypeId", certificateID)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response CertificatePassTypeIDLinkageResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}
