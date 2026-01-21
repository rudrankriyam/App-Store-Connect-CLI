package asc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func rawResponse(status int, body string) *http.Response {
	return &http.Response{
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/a-gzip"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestBuildSalesReportQuery(t *testing.T) {
	query := buildSalesReportQuery(SalesReportParams{
		VendorNumber:  "12345678",
		ReportType:    SalesReportTypeSales,
		ReportSubType: SalesReportSubTypeSummary,
		Frequency:     SalesReportFrequencyDaily,
		ReportDate:    "2024-01-20",
		Version:       SalesReportVersion1_0,
	})

	values, err := url.ParseQuery(query)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}
	if got := values.Get("filter[vendorNumber]"); got != "12345678" {
		t.Fatalf("expected vendorNumber filter, got %q", got)
	}
	if got := values.Get("filter[reportType]"); got != "SALES" {
		t.Fatalf("expected reportType filter, got %q", got)
	}
	if got := values.Get("filter[reportSubType]"); got != "SUMMARY" {
		t.Fatalf("expected reportSubType filter, got %q", got)
	}
	if got := values.Get("filter[frequency]"); got != "DAILY" {
		t.Fatalf("expected frequency filter, got %q", got)
	}
	if got := values.Get("filter[reportDate]"); got != "2024-01-20" {
		t.Fatalf("expected reportDate filter, got %q", got)
	}
	if got := values.Get("filter[version]"); got != "1_0" {
		t.Fatalf("expected version filter, got %q", got)
	}
}

func TestGetSalesReport_SendsRequest(t *testing.T) {
	response := rawResponse(http.StatusOK, "gzdata")
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/salesReports" {
			t.Fatalf("expected path /v1/salesReports, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[vendorNumber]") != "12345678" {
			t.Fatalf("expected vendorNumber filter, got %q", values.Get("filter[vendorNumber]"))
		}
		if req.Header.Get("Accept") != "application/a-gzip" {
			t.Fatalf("expected gzip Accept header, got %q", req.Header.Get("Accept"))
		}
		assertAuthorized(t, req)
	}, response)

	download, err := client.GetSalesReport(context.Background(), SalesReportParams{
		VendorNumber:  "12345678",
		ReportType:    SalesReportTypeSales,
		ReportSubType: SalesReportSubTypeSummary,
		Frequency:     SalesReportFrequencyDaily,
		ReportDate:    "2024-01-20",
		Version:       SalesReportVersion1_0,
	})
	if err != nil {
		t.Fatalf("GetSalesReport() error: %v", err)
	}
	_ = download.Body.Close()
}

func TestGetSalesReport_ErrorResponse(t *testing.T) {
	response := jsonResponse(http.StatusForbidden, `{"errors":[{"code":"FORBIDDEN","title":"Forbidden","detail":"nope"}]}`)
	client := newTestClient(t, nil, response)
	_, err := client.GetSalesReport(context.Background(), SalesReportParams{
		VendorNumber:  "12345678",
		ReportType:    SalesReportTypeSales,
		ReportSubType: SalesReportSubTypeSummary,
		Frequency:     SalesReportFrequencyDaily,
		ReportDate:    "2024-01-20",
		Version:       SalesReportVersion1_0,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Forbidden") {
		t.Fatalf("expected Forbidden error, got %v", err)
	}
}

func TestCreateAnalyticsReportRequest_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"analyticsReportRequests","id":"req-1","attributes":{"accessType":"ONGOING","state":"PROCESSING"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/analyticsReportRequests" {
			t.Fatalf("expected path /v1/analyticsReportRequests, got %s", req.URL.Path)
		}
		var payload AnalyticsReportRequestCreateRequest
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if payload.Data.Type != ResourceTypeAnalyticsReportRequests {
			t.Fatalf("expected type analyticsReportRequests, got %q", payload.Data.Type)
		}
		if payload.Data.Attributes.AccessType != AnalyticsAccessTypeOngoing {
			t.Fatalf("expected accessType ONGOING, got %q", payload.Data.Attributes.AccessType)
		}
		if payload.Data.Relationships.App.Data.ID != "app-1" {
			t.Fatalf("expected app id app-1, got %q", payload.Data.Relationships.App.Data.ID)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.CreateAnalyticsReportRequest(context.Background(), "app-1", AnalyticsAccessTypeOngoing); err != nil {
		t.Fatalf("CreateAnalyticsReportRequest() error: %v", err)
	}
}

func TestGetAnalyticsReportRequests_WithFilters(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/apps/app-1/analyticsReportRequests" {
			t.Fatalf("expected path /v1/apps/app-1/analyticsReportRequests, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[state]") != "COMPLETED" {
			t.Fatalf("expected filter[state]=COMPLETED, got %q", values.Get("filter[state]"))
		}
		if values.Get("limit") != "10" {
			t.Fatalf("expected limit=10, got %q", values.Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAnalyticsReportRequests(
		context.Background(),
		"app-1",
		WithAnalyticsReportRequestsLimit(10),
		WithAnalyticsReportRequestsState("COMPLETED"),
	); err != nil {
		t.Fatalf("GetAnalyticsReportRequests() error: %v", err)
	}
}

func TestGetAnalyticsReportRequest_ByID(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"analyticsReportRequests","id":"req-1"}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/analyticsReportRequests/req-1" {
			t.Fatalf("expected path /v1/analyticsReportRequests/req-1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAnalyticsReportRequest(context.Background(), "req-1"); err != nil {
		t.Fatalf("GetAnalyticsReportRequest() error: %v", err)
	}
}

func TestGetAnalyticsReports_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/analyticsReportRequests/req-1/reports?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAnalyticsReports(context.Background(), "req-1", WithAnalyticsReportsNextURL(next)); err != nil {
		t.Fatalf("GetAnalyticsReports() error: %v", err)
	}
}

func TestGetAnalyticsReportInstances_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/analyticsReports/report-1/instances?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAnalyticsReportInstances(context.Background(), "report-1", WithAnalyticsReportInstancesNextURL(next)); err != nil {
		t.Fatalf("GetAnalyticsReportInstances() error: %v", err)
	}
}

func TestGetAnalyticsReportSegments_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/analyticsReportInstances/inst-1/segments?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetAnalyticsReportSegments(context.Background(), "inst-1", WithAnalyticsReportSegmentsNextURL(next)); err != nil {
		t.Fatalf("GetAnalyticsReportSegments() error: %v", err)
	}
}

func TestDownloadAnalyticsReport_NoAuthHeader(t *testing.T) {
	downloadURL := "https://example.com/report.gz"
	response := rawResponse(http.StatusOK, "gzdata")
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != downloadURL {
			t.Fatalf("expected URL %q, got %q", downloadURL, req.URL.String())
		}
		if req.Header.Get("Authorization") != "" {
			t.Fatalf("expected no Authorization header")
		}
	}, response)

	download, err := client.DownloadAnalyticsReport(context.Background(), downloadURL)
	if err != nil {
		t.Fatalf("DownloadAnalyticsReport() error: %v", err)
	}
	_ = download.Body.Close()
}
