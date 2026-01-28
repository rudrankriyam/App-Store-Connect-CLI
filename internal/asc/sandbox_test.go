package asc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestBuildSandboxTestersQuery(t *testing.T) {
	query := &sandboxTestersQuery{}
	opts := []SandboxTestersOption{
		WithSandboxTestersEmail(" tester@example.com "),
		WithSandboxTestersTerritory("usa"),
		WithSandboxTestersLimit(10),
	}
	for _, opt := range opts {
		opt(query)
	}

	values, err := url.ParseQuery(buildSandboxTestersQuery(query))
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("limit"); got != "10" {
		t.Fatalf("expected limit=10, got %q", got)
	}
}

func TestGetSandboxTesters_WithFilters(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"sandboxTesters","id":"1","attributes":{"acAccountName":"tester@example.com","firstName":"Test","lastName":"User","territory":"USA"}},{"type":"sandboxTesters","id":"2","attributes":{"acAccountName":"other@example.com","firstName":"Other","lastName":"User","territory":"JPN"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v2/sandboxTesters" {
			t.Fatalf("expected path /v2/sandboxTesters, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("limit") != "10" {
			t.Fatalf("expected limit=10, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	resp, err := client.GetSandboxTesters(context.Background(),
		WithSandboxTestersEmail("tester@example.com"),
		WithSandboxTestersTerritory("usa"),
		WithSandboxTestersLimit(10),
	)
	if err != nil {
		t.Fatalf("GetSandboxTesters() error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 sandbox tester, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "1" {
		t.Fatalf("expected tester ID 1, got %q", resp.Data[0].ID)
	}
}

func TestGetSandboxTesters_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/sandboxTesters?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetSandboxTesters(context.Background(), WithSandboxTestersNextURL(next)); err != nil {
		t.Fatalf("GetSandboxTesters() error: %v", err)
	}
}

func TestGetSandboxTester_ByID(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"sandboxTesters","id":"tester-1","attributes":{"acAccountName":"tester@example.com","territory":"USA"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v2/sandboxTesters" {
			t.Fatalf("expected path /v2/sandboxTesters, got %s", req.URL.Path)
		}
		if got := req.URL.Query().Get("limit"); got != "200" {
			t.Fatalf("expected limit=200, got %q", got)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetSandboxTester(context.Background(), "tester-1"); err != nil {
		t.Fatalf("GetSandboxTester() error: %v", err)
	}
}

func TestUpdateSandboxTester_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"sandboxTesters","id":"tester-1","attributes":{"acAccountName":"tester@example.com","territory":"USA"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", req.Method)
		}
		if req.URL.Path != "/v2/sandboxTesters/tester-1" {
			t.Fatalf("expected path /v2/sandboxTesters/tester-1, got %s", req.URL.Path)
		}

		var body map[string]map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		data, ok := body["data"]
		if !ok {
			t.Fatalf("expected data in request body")
		}
		if data["type"] != "sandboxTesters" {
			t.Fatalf("expected type sandboxTesters, got %v", data["type"])
		}
		if data["id"] != "tester-1" {
			t.Fatalf("expected id tester-1, got %v", data["id"])
		}
		attributes, ok := data["attributes"].(map[string]any)
		if !ok {
			t.Fatalf("expected attributes in request body")
		}
		if attributes["territory"] != "USA" {
			t.Fatalf("expected territory USA, got %v", attributes["territory"])
		}
		if attributes["interruptPurchases"] != true {
			t.Fatalf("expected interruptPurchases true, got %v", attributes["interruptPurchases"])
		}
		if attributes["subscriptionRenewalRate"] != string(SandboxTesterRenewalEveryOneHour) {
			t.Fatalf("expected subscriptionRenewalRate %s, got %v", SandboxTesterRenewalEveryOneHour, attributes["subscriptionRenewalRate"])
		}
		assertAuthorized(t, req)
	}, response)

	territory := "USA"
	interrupt := true
	rate := SandboxTesterRenewalEveryOneHour
	attrs := SandboxTesterUpdateAttributes{
		Territory:               &territory,
		InterruptPurchases:      &interrupt,
		SubscriptionRenewalRate: &rate,
	}
	if _, err := client.UpdateSandboxTester(context.Background(), "tester-1", attrs); err != nil {
		t.Fatalf("UpdateSandboxTester() error: %v", err)
	}
}

func TestCreateSandboxTester_NotSupported(t *testing.T) {
	client := newTestClient(t, func(req *http.Request) {
		t.Fatalf("unexpected request to %s %s", req.Method, req.URL.Path)
	}, jsonResponse(http.StatusOK, `{}`))

	_, err := client.CreateSandboxTester(context.Background(), SandboxTesterCreateAttributes{})
	if err == nil {
		t.Fatal("expected CreateSandboxTester to return error")
	}
	if !strings.Contains(err.Error(), "not supported by OpenAPI v2") {
		t.Fatalf("expected OpenAPI v2 unsupported error, got %v", err)
	}
}

func TestClearSandboxTesterPurchaseHistory(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"sandboxTestersClearPurchaseHistoryRequest","id":"request-1"}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v2/sandboxTestersClearPurchaseHistoryRequest" {
			t.Fatalf("expected path /v2/sandboxTestersClearPurchaseHistoryRequest, got %s", req.URL.Path)
		}

		var body map[string]map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		data, ok := body["data"]
		if !ok {
			t.Fatalf("expected data in request body")
		}
		if data["type"] != "sandboxTestersClearPurchaseHistoryRequest" {
			t.Fatalf("expected type sandboxTestersClearPurchaseHistoryRequest, got %v", data["type"])
		}
		relationships, ok := data["relationships"].(map[string]any)
		if !ok {
			t.Fatalf("expected relationships in request body")
		}
		sandboxTesters, ok := relationships["sandboxTesters"].(map[string]any)
		if !ok {
			t.Fatalf("expected sandboxTesters relationship in request body")
		}
		dataList, ok := sandboxTesters["data"].([]any)
		if !ok || len(dataList) != 1 {
			t.Fatalf("expected sandboxTesters data array with one element")
		}
		item, ok := dataList[0].(map[string]any)
		if !ok {
			t.Fatalf("expected sandboxTesters data item to be object")
		}
		if item["id"] != "tester-1" || item["type"] != "sandboxTesters" {
			t.Fatalf("expected sandbox tester id/type, got %v", item)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.ClearSandboxTesterPurchaseHistory(context.Background(), "tester-1"); err != nil {
		t.Fatalf("ClearSandboxTesterPurchaseHistory() error: %v", err)
	}
}

func TestDeleteSandboxTester_NotSupported(t *testing.T) {
	client := newTestClient(t, func(req *http.Request) {
		t.Fatalf("unexpected request to %s %s", req.Method, req.URL.Path)
	}, jsonResponse(http.StatusOK, `{}`))

	if err := client.DeleteSandboxTester(context.Background(), "tester-1"); err == nil {
		t.Fatal("expected DeleteSandboxTester to return error")
	} else if !strings.Contains(err.Error(), "not supported by OpenAPI v2") {
		t.Fatalf("expected OpenAPI v2 unsupported error, got %v", err)
	}
}
