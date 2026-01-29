package asc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestGetPassTypeIDs_WithFilters(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"passTypeIds","id":"p1","attributes":{"name":"Wallet","identifier":"pass.com.example"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/passTypeIds" {
			t.Fatalf("expected path /v1/passTypeIds, got %s", req.URL.Path)
		}
		values := req.URL.Query()
		if values.Get("filter[identifier]") != "pass.com.example" {
			t.Fatalf("expected filter[identifier]=pass.com.example, got %q", values.Get("filter[identifier]"))
		}
		if values.Get("filter[name]") != "Wallet" {
			t.Fatalf("expected filter[name]=Wallet, got %q", values.Get("filter[name]"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetPassTypeIDs(
		context.Background(),
		WithPassTypeIDsFilterIdentifier("pass.com.example"),
		WithPassTypeIDsFilterName("Wallet"),
	); err != nil {
		t.Fatalf("GetPassTypeIDs() error: %v", err)
	}
}

func TestGetPassTypeIDs_WithLimit(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"passTypeIds","id":"p1","attributes":{"name":"Wallet","identifier":"pass.com.example"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/passTypeIds" {
			t.Fatalf("expected path /v1/passTypeIds, got %s", req.URL.Path)
		}
		if req.URL.Query().Get("limit") != "10" {
			t.Fatalf("expected limit=10, got %q", req.URL.Query().Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetPassTypeIDs(context.Background(), WithPassTypeIDsLimit(10)); err != nil {
		t.Fatalf("GetPassTypeIDs() error: %v", err)
	}
}

func TestGetPassTypeIDs_UsesNextURL(t *testing.T) {
	next := "https://api.appstoreconnect.apple.com/v1/passTypeIds?cursor=abc"
	response := jsonResponse(http.StatusOK, `{"data":[]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.URL.String() != next {
			t.Fatalf("expected next URL %q, got %q", next, req.URL.String())
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetPassTypeIDs(context.Background(), WithPassTypeIDsNextURL(next)); err != nil {
		t.Fatalf("GetPassTypeIDs() error: %v", err)
	}
}

func TestGetPassTypeID_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"passTypeIds","id":"p1","attributes":{"name":"Wallet","identifier":"pass.com.example"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/passTypeIds/p1" {
			t.Fatalf("expected path /v1/passTypeIds/p1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetPassTypeID(context.Background(), "p1"); err != nil {
		t.Fatalf("GetPassTypeID() error: %v", err)
	}
}

func TestCreatePassTypeID_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusCreated, `{"data":{"type":"passTypeIds","id":"p1","attributes":{"name":"Wallet","identifier":"pass.com.example"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/passTypeIds" {
			t.Fatalf("expected path /v1/passTypeIds, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		var payload PassTypeIDCreateRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if payload.Data.Type != ResourceTypePassTypeIds {
			t.Fatalf("expected type passTypeIds, got %q", payload.Data.Type)
		}
		if payload.Data.Attributes.Identifier != "pass.com.example" {
			t.Fatalf("expected identifier pass.com.example, got %q", payload.Data.Attributes.Identifier)
		}
		if payload.Data.Attributes.Name != "Wallet" {
			t.Fatalf("expected name Wallet, got %q", payload.Data.Attributes.Name)
		}
		assertAuthorized(t, req)
	}, response)

	attrs := PassTypeIDCreateAttributes{
		Name:       "Wallet",
		Identifier: "pass.com.example",
	}
	if _, err := client.CreatePassTypeID(context.Background(), attrs); err != nil {
		t.Fatalf("CreatePassTypeID() error: %v", err)
	}
}

func TestUpdatePassTypeID_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"passTypeIds","id":"p1","attributes":{"name":"Updated","identifier":"pass.com.example"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", req.Method)
		}
		if req.URL.Path != "/v1/passTypeIds/p1" {
			t.Fatalf("expected path /v1/passTypeIds/p1, got %s", req.URL.Path)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body error: %v", err)
		}
		var payload PassTypeIDUpdateRequest
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body error: %v", err)
		}
		if payload.Data.Type != ResourceTypePassTypeIds {
			t.Fatalf("expected type passTypeIds, got %q", payload.Data.Type)
		}
		if payload.Data.ID != "p1" {
			t.Fatalf("expected id p1, got %q", payload.Data.ID)
		}
		if payload.Data.Attributes == nil || payload.Data.Attributes.Name != "Updated" {
			t.Fatalf("expected name Updated, got %v", payload.Data.Attributes)
		}
		assertAuthorized(t, req)
	}, response)

	attrs := PassTypeIDUpdateAttributes{Name: "Updated"}
	if _, err := client.UpdatePassTypeID(context.Background(), "p1", attrs); err != nil {
		t.Fatalf("UpdatePassTypeID() error: %v", err)
	}
}

func TestDeletePassTypeID_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusNoContent, ``)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", req.Method)
		}
		if req.URL.Path != "/v1/passTypeIds/p1" {
			t.Fatalf("expected path /v1/passTypeIds/p1, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if err := client.DeletePassTypeID(context.Background(), "p1"); err != nil {
		t.Fatalf("DeletePassTypeID() error: %v", err)
	}
}

func TestGetPassTypeIDCertificates_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"certificates","id":"c1","attributes":{"name":"Cert","certificateType":"APPLE_PAY"}}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/passTypeIds/p1/certificates" {
			t.Fatalf("expected path /v1/passTypeIds/p1/certificates, got %s", req.URL.Path)
		}
		if req.URL.Query().Get("limit") != "10" {
			t.Fatalf("expected limit=10, got %q", req.URL.Query().Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetPassTypeIDCertificates(context.Background(), "p1", WithPassTypeIDCertificatesLimit(10)); err != nil {
		t.Fatalf("GetPassTypeIDCertificates() error: %v", err)
	}
}

func TestGetPassTypeIDCertificatesRelationships_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":[{"type":"certificates","id":"c1"}]}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/passTypeIds/p1/relationships/certificates" {
			t.Fatalf("expected path /v1/passTypeIds/p1/relationships/certificates, got %s", req.URL.Path)
		}
		if req.URL.Query().Get("limit") != "5" {
			t.Fatalf("expected limit=5, got %q", req.URL.Query().Get("limit"))
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetPassTypeIDCertificatesRelationships(context.Background(), "p1", WithLinkagesLimit(5)); err != nil {
		t.Fatalf("GetPassTypeIDCertificatesRelationships() error: %v", err)
	}
}

func TestGetCertificatePassTypeID_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"passTypeIds","id":"p1","attributes":{"name":"Wallet","identifier":"pass.com.example"}}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/certificates/c1/passTypeId" {
			t.Fatalf("expected path /v1/certificates/c1/passTypeId, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetCertificatePassTypeID(context.Background(), "c1"); err != nil {
		t.Fatalf("GetCertificatePassTypeID() error: %v", err)
	}
}

func TestGetCertificatePassTypeIDRelationship_SendsRequest(t *testing.T) {
	response := jsonResponse(http.StatusOK, `{"data":{"type":"passTypeIds","id":"p1"}}`)
	client := newTestClient(t, func(req *http.Request) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/certificates/c1/relationships/passTypeId" {
			t.Fatalf("expected path /v1/certificates/c1/relationships/passTypeId, got %s", req.URL.Path)
		}
		assertAuthorized(t, req)
	}, response)

	if _, err := client.GetCertificatePassTypeIDRelationship(context.Background(), "c1"); err != nil {
		t.Fatalf("GetCertificatePassTypeIDRelationship() error: %v", err)
	}
}
