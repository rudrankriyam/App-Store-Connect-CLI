package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestGetTerritories(t *testing.T) {
	resp := TerritoriesResponse{
		Data: []Resource[TerritoryAttributes]{
			{Type: ResourceTypeTerritories, ID: "USA", Attributes: TerritoryAttributes{Currency: "USD"}},
		},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/territories" {
			t.Fatalf("expected path /v1/territories, got %s", req.URL.Path)
		}
		if got := req.URL.Query().Get("limit"); got != "5" {
			t.Fatalf("expected limit=5, got %q", got)
		}
	}, jsonResponse(http.StatusOK, string(body)))

	result, err := client.GetTerritories(context.Background(), WithTerritoriesLimit(5))
	if err != nil {
		t.Fatalf("GetTerritories() error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 territory, got %d", len(result.Data))
	}
	if result.Data[0].ID != "USA" {
		t.Fatalf("expected territory USA, got %q", result.Data[0].ID)
	}
}

func TestGetAppPricePoints_WithTerritory(t *testing.T) {
	resp := AppPricePointsV3Response{
		Data: []Resource[AppPricePointV3Attributes]{
			{Type: ResourceTypeAppPricePoints, ID: "pp-1", Attributes: AppPricePointV3Attributes{CustomerPrice: "0.99", Proceeds: "0.70"}},
		},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.URL.Path != "/v1/apps/app-1/appPricePoints" {
			t.Fatalf("expected path /v1/apps/app-1/appPricePoints, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[territory]") != "USA" {
			t.Fatalf("expected territory filter USA, got %q", values.Get("filter[territory]"))
		}
		if values.Get("limit") != "10" {
			t.Fatalf("expected limit=10, got %q", values.Get("limit"))
		}
	}, jsonResponse(http.StatusOK, string(body)))

	result, err := client.GetAppPricePoints(context.Background(), "app-1",
		WithPricePointsTerritory("usa"),
		WithPricePointsLimit(10),
	)
	if err != nil {
		t.Fatalf("GetAppPricePoints() error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 price point, got %d", len(result.Data))
	}
}

func TestGetAppPricePoint(t *testing.T) {
	single := SingleResponse[AppPricePointV3Attributes]{
		Data: Resource[AppPricePointV3Attributes]{
			Type: ResourceTypeAppPricePoints,
			ID:   "pp-1",
			Attributes: AppPricePointV3Attributes{
				CustomerPrice: "0.99",
				Proceeds:      "0.70",
			},
		},
	}
	body, _ := json.Marshal(single)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v3/appPricePoints/pp-1" {
			t.Fatalf("expected path /v3/appPricePoints/pp-1, got %s", req.URL.Path)
		}
	}, jsonResponse(http.StatusOK, string(body)))

	result, err := client.GetAppPricePoint(context.Background(), "pp-1")
	if err != nil {
		t.Fatalf("GetAppPricePoint() error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 price point, got %d", len(result.Data))
	}
	if result.Data[0].ID != "pp-1" {
		t.Fatalf("expected price point pp-1, got %q", result.Data[0].ID)
	}
}

func TestGetAppPricePointEqualizations(t *testing.T) {
	resp := AppPricePointsV3Response{
		Data: []Resource[AppPricePointV3Attributes]{
			{Type: ResourceTypeAppPricePoints, ID: "pp-eq-1"},
		},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.URL.Path != "/v3/appPricePoints/pp-1/equalizations" {
			t.Fatalf("expected path /v3/appPricePoints/pp-1/equalizations, got %s", req.URL.Path)
		}
	}, jsonResponse(http.StatusOK, string(body)))

	if _, err := client.GetAppPricePointEqualizations(context.Background(), "pp-1"); err != nil {
		t.Fatalf("GetAppPricePointEqualizations() error: %v", err)
	}
}

func TestGetAppPriceSchedule(t *testing.T) {
	resp := AppPriceScheduleResponse{
		Data: Resource[AppPriceScheduleAttributes]{
			Type: ResourceTypeAppPriceSchedules,
			ID:   "schedule-1",
		},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.URL.Path != "/v1/apps/app-1/appPriceSchedule" {
			t.Fatalf("expected path /v1/apps/app-1/appPriceSchedule, got %s", req.URL.Path)
		}
	}, jsonResponse(http.StatusOK, string(body)))

	result, err := client.GetAppPriceSchedule(context.Background(), "app-1")
	if err != nil {
		t.Fatalf("GetAppPriceSchedule() error: %v", err)
	}
	if result.Data.ID != "schedule-1" {
		t.Fatalf("expected schedule ID schedule-1, got %q", result.Data.ID)
	}
}

func TestGetAppPriceScheduleManualPrices(t *testing.T) {
	resp := AppPricesResponse{
		Data: []Resource[AppPriceAttributes]{{Type: ResourceTypeAppPrices, ID: "price-1"}},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.URL.Path != "/v1/appPriceSchedules/schedule-1/manualPrices" {
			t.Fatalf("expected path /v1/appPriceSchedules/schedule-1/manualPrices, got %s", req.URL.Path)
		}
	}, jsonResponse(http.StatusOK, string(body)))

	if _, err := client.GetAppPriceScheduleManualPrices(context.Background(), "schedule-1"); err != nil {
		t.Fatalf("GetAppPriceScheduleManualPrices() error: %v", err)
	}
}

func TestGetAppPriceScheduleAutomaticPrices(t *testing.T) {
	resp := AppPricesResponse{
		Data: []Resource[AppPriceAttributes]{{Type: ResourceTypeAppPrices, ID: "price-1"}},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.URL.Path != "/v1/appPriceSchedules/schedule-1/automaticPrices" {
			t.Fatalf("expected path /v1/appPriceSchedules/schedule-1/automaticPrices, got %s", req.URL.Path)
		}
	}, jsonResponse(http.StatusOK, string(body)))

	if _, err := client.GetAppPriceScheduleAutomaticPrices(context.Background(), "schedule-1"); err != nil {
		t.Fatalf("GetAppPriceScheduleAutomaticPrices() error: %v", err)
	}
}

func TestCreateAppPriceSchedule(t *testing.T) {
	resp := AppPriceScheduleResponse{
		Data: Resource[AppPriceScheduleAttributes]{
			Type: ResourceTypeAppPriceSchedules,
			ID:   "schedule-1",
		},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/appPriceSchedules" {
			t.Fatalf("expected path /v1/appPriceSchedules, got %s", req.URL.Path)
		}

		var createReq AppPriceScheduleCreateRequest
		if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if createReq.Data.Type != ResourceTypeAppPriceSchedules {
			t.Fatalf("expected type appPriceSchedules, got %v", createReq.Data.Type)
		}
		if createReq.Data.Relationships.App.Data.ID != "app-1" {
			t.Fatalf("expected app ID app-1, got %q", createReq.Data.Relationships.App.Data.ID)
		}
		if len(createReq.Data.Relationships.ManualPrices.Data) != 1 {
			t.Fatalf("expected 1 manual price, got %d", len(createReq.Data.Relationships.ManualPrices.Data))
		}
		if len(createReq.Included) != 1 {
			t.Fatalf("expected 1 included price, got %d", len(createReq.Included))
		}
		if createReq.Included[0].Attributes.StartDate != "2024-03-01" {
			t.Fatalf("expected start date 2024-03-01, got %q", createReq.Included[0].Attributes.StartDate)
		}
		if createReq.Included[0].Relationships.AppPricePoint.Data.ID != "pp-1" {
			t.Fatalf("expected price point pp-1, got %q", createReq.Included[0].Relationships.AppPricePoint.Data.ID)
		}
		if createReq.Data.Relationships.ManualPrices.Data[0].ID != createReq.Included[0].ID {
			t.Fatalf("expected manual price relationship to match included id")
		}
	}, jsonResponse(http.StatusCreated, string(body)))

	_, err := client.CreateAppPriceSchedule(context.Background(), "app-1", AppPriceScheduleCreateAttributes{
		PricePointID: "pp-1",
		StartDate:    "2024-03-01",
	})
	if err != nil {
		t.Fatalf("CreateAppPriceSchedule() error: %v", err)
	}
}

func TestGetAppAvailabilityV2(t *testing.T) {
	resp := AppAvailabilityV2Response{
		Data: Resource[AppAvailabilityV2Attributes]{
			Type: ResourceTypeAppAvailabilities,
			ID:   "availability-1",
			Attributes: AppAvailabilityV2Attributes{
				AvailableInNewTerritories: true,
			},
		},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.URL.Path != "/v1/apps/app-1/appAvailabilityV2" {
			t.Fatalf("expected path /v1/apps/app-1/appAvailabilityV2, got %s", req.URL.Path)
		}
	}, jsonResponse(http.StatusOK, string(body)))

	if _, err := client.GetAppAvailabilityV2(context.Background(), "app-1"); err != nil {
		t.Fatalf("GetAppAvailabilityV2() error: %v", err)
	}
}

func TestGetTerritoryAvailabilities(t *testing.T) {
	resp := TerritoryAvailabilitiesResponse{
		Data: []Resource[TerritoryAvailabilityAttributes]{
			{Type: ResourceTypeTerritoryAvailabilities, ID: "ta-1"},
		},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.URL.Path != "/v2/appAvailabilities/availability-1/territoryAvailabilities" {
			t.Fatalf("expected path /v2/appAvailabilities/availability-1/territoryAvailabilities, got %s", req.URL.Path)
		}
	}, jsonResponse(http.StatusOK, string(body)))

	if _, err := client.GetTerritoryAvailabilities(context.Background(), "availability-1"); err != nil {
		t.Fatalf("GetTerritoryAvailabilities() error: %v", err)
	}
}

func TestCreateAppAvailabilityV2(t *testing.T) {
	resp := AppAvailabilityV2Response{
		Data: Resource[AppAvailabilityV2Attributes]{
			Type: ResourceTypeAppAvailabilities,
			ID:   "availability-1",
		},
	}
	body, _ := json.Marshal(resp)

	client := newTestClient(t, func(req *http.Request) {
		assertAuthorized(t, req)
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v2/appAvailabilities" {
			t.Fatalf("expected path /v2/appAvailabilities, got %s", req.URL.Path)
		}

		var createReq AppAvailabilityV2CreateRequest
		if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if createReq.Data.Relationships.App.Data.ID != "app-1" {
			t.Fatalf("expected app ID app-1, got %q", createReq.Data.Relationships.App.Data.ID)
		}
		if len(createReq.Data.Relationships.TerritoryAvailabilities.Data) != 2 {
			t.Fatalf("expected 2 territory availabilities, got %d", len(createReq.Data.Relationships.TerritoryAvailabilities.Data))
		}
		if len(createReq.Included) != 2 {
			t.Fatalf("expected 2 included items, got %d", len(createReq.Included))
		}
		if createReq.Included[0].Relationships.Territory.Data.ID == "" {
			t.Fatalf("expected territory relationship to be set")
		}
		if createReq.Included[0].Attributes.Available == createReq.Included[1].Attributes.Available {
			t.Fatalf("expected available values to differ for test coverage")
		}
	}, jsonResponse(http.StatusCreated, string(body)))

	_, err := client.CreateAppAvailabilityV2(context.Background(), "app-1", AppAvailabilityV2CreateAttributes{
		TerritoryAvailabilities: []TerritoryAvailabilityCreate{
			{TerritoryID: "usa", Available: true},
			{TerritoryID: "gbr", Available: false},
		},
	})
	if err != nil {
		t.Fatalf("CreateAppAvailabilityV2() error: %v", err)
	}
}

func TestPaginateAll_Territories(t *testing.T) {
	makePage := func(page int) *TerritoriesResponse {
		links := Links{}
		if page < 2 {
			links.Next = fmt.Sprintf("page=%d", page+1)
		}
		return &TerritoriesResponse{
			Data: []Resource[TerritoryAttributes]{
				{Type: ResourceTypeTerritories, ID: fmt.Sprintf("territory-%d", page)},
			},
			Links: links,
		}
	}

	firstPage := makePage(1)
	response, err := PaginateAll(context.Background(), firstPage, func(ctx context.Context, nextURL string) (PaginatedResponse, error) {
		if !strings.HasPrefix(nextURL, "page=") {
			return nil, fmt.Errorf("unexpected next URL %q", nextURL)
		}
		return makePage(2), nil
	})
	if err != nil {
		t.Fatalf("PaginateAll() error: %v", err)
	}

	territories, ok := response.(*TerritoriesResponse)
	if !ok {
		t.Fatalf("expected TerritoriesResponse, got %T", response)
	}
	if len(territories.Data) != 2 {
		t.Fatalf("expected 2 territories, got %d", len(territories.Data))
	}
	if territories.Links.Next != "" {
		t.Fatalf("expected next link to be cleared, got %q", territories.Links.Next)
	}
}
