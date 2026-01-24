package asc

import (
	"context"
	"encoding/json"
	"fmt"
)

// AppCategoryAttributes describes app category metadata.
type AppCategoryAttributes struct {
	Platforms []Platform `json:"platforms,omitempty"`
}

// AppCategory represents an app category resource.
type AppCategory struct {
	Type       ResourceType          `json:"type"`
	ID         string                `json:"id"`
	Attributes AppCategoryAttributes `json:"attributes,omitempty"`
}

// AppCategoriesResponse is the response from app categories endpoint.
type AppCategoriesResponse struct {
	Data  []AppCategory `json:"data"`
	Links Links         `json:"links,omitempty"`
}

// GetAppCategories retrieves all app categories.
func (c *Client) GetAppCategories(ctx context.Context, opts ...AppCategoriesOption) (*AppCategoriesResponse, error) {
	query := &appCategoriesQuery{}
	for _, opt := range opts {
		opt(query)
	}

	path := "/v1/appCategories"
	if query.limit > 0 {
		path += fmt.Sprintf("?limit=%d", query.limit)
	}

	data, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response AppCategoriesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// appCategoriesQuery holds query parameters for app categories.
type appCategoriesQuery struct {
	limit int
}

// AppCategoriesOption configures app categories queries.
type AppCategoriesOption func(*appCategoriesQuery)

// WithAppCategoriesLimit sets the limit for app categories queries.
func WithAppCategoriesLimit(limit int) AppCategoriesOption {
	return func(q *appCategoriesQuery) {
		q.limit = limit
	}
}
