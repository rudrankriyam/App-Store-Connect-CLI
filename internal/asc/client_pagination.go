package asc

import (
	"context"
	"fmt"
)

// PaginatedResponse represents a response that supports pagination
type PaginatedResponse interface {
	GetLinks() *Links
	GetData() interface{}
}

// GetLinks returns the links field for pagination
func (r *Response[T]) GetLinks() *Links {
	return &r.Links
}

// GetData returns the data field for aggregation
func (r *Response[T]) GetData() interface{} {
	return r.Data
}

// GetLinks returns the links field for pagination.
func (r *PreReleaseVersionsResponse) GetLinks() *Links {
	return &r.Links
}

// GetData returns the data field for aggregation.
func (r *PreReleaseVersionsResponse) GetData() interface{} {
	return r.Data
}

// PaginateFunc is a function that fetches a page of results
type PaginateFunc func(ctx context.Context, nextURL string) (PaginatedResponse, error)

// PaginateAll fetches all pages and aggregates results
func PaginateAll(ctx context.Context, firstPage PaginatedResponse, fetchNext PaginateFunc) (PaginatedResponse, error) {
	if firstPage == nil {
		return nil, nil
	}

	// Determine the response type from the first page
	var result PaginatedResponse
	switch firstPage.(type) {
	case *FeedbackResponse:
		result = &FeedbackResponse{Links: Links{}}
	case *CrashesResponse:
		result = &CrashesResponse{Links: Links{}}
	case *ReviewsResponse:
		result = &ReviewsResponse{Links: Links{}}
	case *AppsResponse:
		result = &AppsResponse{Links: Links{}}
	case *TerritoriesResponse:
		result = &TerritoriesResponse{Links: Links{}}
	case *AppPricePointsV3Response:
		result = &AppPricePointsV3Response{Links: Links{}}
	case *BuildsResponse:
		result = &BuildsResponse{Links: Links{}}
	case *AppStoreVersionsResponse:
		result = &AppStoreVersionsResponse{Links: Links{}}
	case *PreReleaseVersionsResponse:
		result = &PreReleaseVersionsResponse{Links: Links{}}
	case *AppStoreVersionLocalizationsResponse:
		result = &AppStoreVersionLocalizationsResponse{Links: Links{}}
	case *AppInfoLocalizationsResponse:
		result = &AppInfoLocalizationsResponse{Links: Links{}}
	case *BetaGroupsResponse:
		result = &BetaGroupsResponse{Links: Links{}}
	case *BetaTestersResponse:
		result = &BetaTestersResponse{Links: Links{}}
	case *UsersResponse:
		result = &UsersResponse{Links: Links{}}
	case *UserInvitationsResponse:
		result = &UserInvitationsResponse{Links: Links{}}
	case *SandboxTestersResponse:
		result = &SandboxTestersResponse{Links: Links{}}
	case *AnalyticsReportRequestsResponse:
		result = &AnalyticsReportRequestsResponse{Links: Links{}}
	case *CiProductsResponse:
		result = &CiProductsResponse{Links: Links{}}
	case *CiWorkflowsResponse:
		result = &CiWorkflowsResponse{Links: Links{}}
	case *ScmGitReferencesResponse:
		result = &ScmGitReferencesResponse{Links: Links{}}
	case *CiBuildRunsResponse:
		result = &CiBuildRunsResponse{Links: Links{}}
	default:
		return nil, fmt.Errorf("unsupported response type for pagination")
	}

	page := 1
	for {
		// Aggregate data from current page
		switch p := firstPage.(type) {
		case *FeedbackResponse:
			result.(*FeedbackResponse).Data = append(result.(*FeedbackResponse).Data, p.Data...)
		case *CrashesResponse:
			result.(*CrashesResponse).Data = append(result.(*CrashesResponse).Data, p.Data...)
		case *ReviewsResponse:
			result.(*ReviewsResponse).Data = append(result.(*ReviewsResponse).Data, p.Data...)
		case *AppsResponse:
			result.(*AppsResponse).Data = append(result.(*AppsResponse).Data, p.Data...)
		case *TerritoriesResponse:
			result.(*TerritoriesResponse).Data = append(result.(*TerritoriesResponse).Data, p.Data...)
		case *AppPricePointsV3Response:
			result.(*AppPricePointsV3Response).Data = append(result.(*AppPricePointsV3Response).Data, p.Data...)
		case *BuildsResponse:
			result.(*BuildsResponse).Data = append(result.(*BuildsResponse).Data, p.Data...)
		case *AppStoreVersionsResponse:
			result.(*AppStoreVersionsResponse).Data = append(result.(*AppStoreVersionsResponse).Data, p.Data...)
		case *PreReleaseVersionsResponse:
			result.(*PreReleaseVersionsResponse).Data = append(result.(*PreReleaseVersionsResponse).Data, p.Data...)
		case *AppStoreVersionLocalizationsResponse:
			result.(*AppStoreVersionLocalizationsResponse).Data = append(result.(*AppStoreVersionLocalizationsResponse).Data, p.Data...)
		case *AppInfoLocalizationsResponse:
			result.(*AppInfoLocalizationsResponse).Data = append(result.(*AppInfoLocalizationsResponse).Data, p.Data...)
		case *BetaGroupsResponse:
			result.(*BetaGroupsResponse).Data = append(result.(*BetaGroupsResponse).Data, p.Data...)
		case *BetaTestersResponse:
			result.(*BetaTestersResponse).Data = append(result.(*BetaTestersResponse).Data, p.Data...)
		case *UsersResponse:
			result.(*UsersResponse).Data = append(result.(*UsersResponse).Data, p.Data...)
		case *UserInvitationsResponse:
			result.(*UserInvitationsResponse).Data = append(result.(*UserInvitationsResponse).Data, p.Data...)
		case *SandboxTestersResponse:
			result.(*SandboxTestersResponse).Data = append(result.(*SandboxTestersResponse).Data, p.Data...)
		case *AnalyticsReportRequestsResponse:
			result.(*AnalyticsReportRequestsResponse).Data = append(result.(*AnalyticsReportRequestsResponse).Data, p.Data...)
		case *CiProductsResponse:
			result.(*CiProductsResponse).Data = append(result.(*CiProductsResponse).Data, p.Data...)
		case *CiWorkflowsResponse:
			result.(*CiWorkflowsResponse).Data = append(result.(*CiWorkflowsResponse).Data, p.Data...)
		case *ScmGitReferencesResponse:
			result.(*ScmGitReferencesResponse).Data = append(result.(*ScmGitReferencesResponse).Data, p.Data...)
		case *CiBuildRunsResponse:
			result.(*CiBuildRunsResponse).Data = append(result.(*CiBuildRunsResponse).Data, p.Data...)
		}

		// Check for next page
		links := firstPage.GetLinks()
		if links == nil || links.Next == "" {
			break
		}

		page++

		// Fetch next page
		nextPage, err := fetchNext(ctx, links.Next)
		if err != nil {
			return result, fmt.Errorf("page %d: %w", page, err)
		}

		// Validate that the response type matches
		if typeOf(nextPage) != typeOf(firstPage) {
			return result, fmt.Errorf("page %d: unexpected response type (expected %T, got %T)", page, firstPage, nextPage)
		}

		firstPage = nextPage
	}

	return result, nil
}

// typeOf returns the runtime type of a PaginatedResponse
func typeOf(p PaginatedResponse) string {
	switch p.(type) {
	case *FeedbackResponse:
		return "FeedbackResponse"
	case *CrashesResponse:
		return "CrashesResponse"
	case *ReviewsResponse:
		return "ReviewsResponse"
	case *AppsResponse:
		return "AppsResponse"
	case *TerritoriesResponse:
		return "TerritoriesResponse"
	case *AppPricePointsV3Response:
		return "AppPricePointsV3Response"
	case *BuildsResponse:
		return "BuildsResponse"
	case *AppStoreVersionsResponse:
		return "AppStoreVersionsResponse"
	case *PreReleaseVersionsResponse:
		return "PreReleaseVersionsResponse"
	case *AppStoreVersionLocalizationsResponse:
		return "AppStoreVersionLocalizationsResponse"
	case *AppInfoLocalizationsResponse:
		return "AppInfoLocalizationsResponse"
	case *BetaGroupsResponse:
		return "BetaGroupsResponse"
	case *BetaTestersResponse:
		return "BetaTestersResponse"
	case *UsersResponse:
		return "UsersResponse"
	case *UserInvitationsResponse:
		return "UserInvitationsResponse"
	case *SandboxTestersResponse:
		return "SandboxTestersResponse"
	case *AnalyticsReportRequestsResponse:
		return "AnalyticsReportRequestsResponse"
	case *CiProductsResponse:
		return "CiProductsResponse"
	case *CiWorkflowsResponse:
		return "CiWorkflowsResponse"
	case *ScmGitReferencesResponse:
		return "ScmGitReferencesResponse"
	case *CiBuildRunsResponse:
		return "CiBuildRunsResponse"
	default:
		return "unknown"
	}
}
