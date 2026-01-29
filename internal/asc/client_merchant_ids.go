package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GetMerchantIDs retrieves the list of merchant IDs.
func (c *Client) GetMerchantIDs(ctx context.Context, opts ...MerchantIDsOption) (*MerchantIDsResponse, error) {
	query := &merchantIDsQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := "/v1/merchantIds"
	if query.nextURL != "" {
		// Validate nextURL to prevent credential exfiltration
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("merchantIds: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildMerchantIDsQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response MerchantIDsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetMerchantID retrieves a single merchant ID by ID.
func (c *Client) GetMerchantID(ctx context.Context, id string) (*MerchantIDResponse, error) {
	id = strings.TrimSpace(id)
	path := fmt.Sprintf("/v1/merchantIds/%s", id)
	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response MerchantIDResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreateMerchantID creates a new merchant ID.
func (c *Client) CreateMerchantID(ctx context.Context, attrs MerchantIDCreateAttributes) (*MerchantIDResponse, error) {
	request := MerchantIDCreateRequest{
		Data: MerchantIDCreateData{
			Type:       ResourceTypeMerchantIds,
			Attributes: attrs,
		},
	}

	body, err := BuildRequestBody(request)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "POST", "/v1/merchantIds", body)
	if err != nil {
		return nil, err
	}

	var response MerchantIDResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// UpdateMerchantID updates an existing merchant ID.
func (c *Client) UpdateMerchantID(ctx context.Context, id string, attrs MerchantIDUpdateAttributes) (*MerchantIDResponse, error) {
	id = strings.TrimSpace(id)
	request := MerchantIDUpdateRequest{
		Data: MerchantIDUpdateData{
			Type:       ResourceTypeMerchantIds,
			ID:         id,
			Attributes: &attrs,
		},
	}

	body, err := BuildRequestBody(request)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, "PATCH", fmt.Sprintf("/v1/merchantIds/%s", id), body)
	if err != nil {
		return nil, err
	}

	var response MerchantIDResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// DeleteMerchantID deletes a merchant ID by ID.
func (c *Client) DeleteMerchantID(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	path := fmt.Sprintf("/v1/merchantIds/%s", id)
	_, err := c.do(ctx, "DELETE", path, nil)
	return err
}

// GetMerchantIDCertificates retrieves certificates for a merchant ID.
func (c *Client) GetMerchantIDCertificates(ctx context.Context, merchantID string, opts ...MerchantIDCertificatesOption) (*CertificatesResponse, error) {
	query := &merchantIDCertificatesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	merchantID = strings.TrimSpace(merchantID)
	path := fmt.Sprintf("/v1/merchantIds/%s/certificates", merchantID)
	if query.nextURL != "" {
		// Validate nextURL to prevent credential exfiltration
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("merchantIdCertificates: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildMerchantIDCertificatesQuery(query); queryString != "" {
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

// GetMerchantIDCertificatesRelationships retrieves certificate linkages for a merchant ID.
func (c *Client) GetMerchantIDCertificatesRelationships(ctx context.Context, merchantID string, opts ...LinkagesOption) (*MerchantIDCertificatesLinkagesResponse, error) {
	query := &linkagesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	merchantID = strings.TrimSpace(merchantID)
	path := fmt.Sprintf("/v1/merchantIds/%s/relationships/certificates", merchantID)
	if query.nextURL != "" {
		// Validate nextURL to prevent credential exfiltration
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("merchantIdCertificatesRelationships: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildLinkagesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response MerchantIDCertificatesLinkagesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}
