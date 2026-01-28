package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// GetSubscriptionWinBackOffers retrieves win-back offers for a subscription.
func (c *Client) GetSubscriptionWinBackOffers(ctx context.Context, subscriptionID string, opts ...WinBackOffersOption) (*WinBackOffersResponse, error) {
	query := &winBackOffersQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/subscriptions/%s/winBackOffers", strings.TrimSpace(subscriptionID))
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("winBackOffers: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildWinBackOffersQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response WinBackOffersResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetSubscriptionWinBackOffersRelationships retrieves win-back offer relationships for a subscription.
func (c *Client) GetSubscriptionWinBackOffersRelationships(ctx context.Context, subscriptionID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	query := &linkagesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/subscriptions/%s/relationships/winBackOffers", strings.TrimSpace(subscriptionID))
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("winBackOffers relationships: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildLinkagesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response LinkagesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetWinBackOffer retrieves a win-back offer by ID.
func (c *Client) GetWinBackOffer(ctx context.Context, offerID string) (*WinBackOfferResponse, error) {
	path := fmt.Sprintf("/v1/winBackOffers/%s", strings.TrimSpace(offerID))
	data, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response WinBackOfferResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// CreateWinBackOffer creates a win-back offer.
func (c *Client) CreateWinBackOffer(ctx context.Context, req WinBackOfferCreateRequest) (*WinBackOfferResponse, error) {
	body, err := BuildRequestBody(req)
	if err != nil {
		return nil, err
	}

	data, err := c.do(ctx, http.MethodPost, "/v1/winBackOffers", body)
	if err != nil {
		return nil, err
	}

	var response WinBackOfferResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// UpdateWinBackOffer updates a win-back offer.
func (c *Client) UpdateWinBackOffer(ctx context.Context, offerID string, attrs WinBackOfferUpdateAttributes) (*WinBackOfferResponse, error) {
	payload := WinBackOfferUpdateRequest{
		Data: WinBackOfferUpdateData{
			Type:       ResourceTypeWinBackOffers,
			ID:         strings.TrimSpace(offerID),
			Attributes: attrs,
		},
	}

	body, err := BuildRequestBody(payload)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/winBackOffers/%s", strings.TrimSpace(offerID))
	data, err := c.do(ctx, http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}

	var response WinBackOfferResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// DeleteWinBackOffer deletes a win-back offer.
func (c *Client) DeleteWinBackOffer(ctx context.Context, offerID string) error {
	path := fmt.Sprintf("/v1/winBackOffers/%s", strings.TrimSpace(offerID))
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}

// GetWinBackOfferPrices retrieves prices for a win-back offer.
func (c *Client) GetWinBackOfferPrices(ctx context.Context, offerID string, opts ...WinBackOfferPricesOption) (*WinBackOfferPricesResponse, error) {
	query := &winBackOfferPricesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/winBackOffers/%s/prices", strings.TrimSpace(offerID))
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("winBackOfferPrices: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildWinBackOfferPricesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response WinBackOfferPricesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// GetWinBackOfferPricesRelationships retrieves price relationships for a win-back offer.
func (c *Client) GetWinBackOfferPricesRelationships(ctx context.Context, offerID string, opts ...LinkagesOption) (*LinkagesResponse, error) {
	query := &linkagesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := fmt.Sprintf("/v1/winBackOffers/%s/relationships/prices", strings.TrimSpace(offerID))
	if query.nextURL != "" {
		if err := validateNextURL(query.nextURL); err != nil {
			return nil, fmt.Errorf("winBackOfferPrices relationships: %w", err)
		}
		path = query.nextURL
	} else if queryString := buildLinkagesQuery(query); queryString != "" {
		path += "?" + queryString
	}

	data, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response LinkagesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}
