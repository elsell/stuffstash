package httpserver

import (
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"net/http"
	"testing"
)

func TestHealthEndpointReturnsHealthyStatus(t *testing.T) {
	observer := &fakeObserver{}
	server := NewServer(":0", newTestApp(observer, "unused-id"))

	response := performRequest(server, http.MethodGet, "/healthz", "", nil)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var body struct {
		Service string `json:"service"`
		Status  string `json:"status"`
	}
	decodeBody(t, response, &body)

	if body.Service != string(app.ServiceNameStuffStash) {
		t.Fatalf("expected service %q, got %q", app.ServiceNameStuffStash, body.Service)
	}
	if body.Status != string(app.HealthStatusHealthy) {
		t.Fatalf("expected status %q, got %q", app.HealthStatusHealthy, body.Status)
	}

	if len(observer.events) != 1 {
		t.Fatalf("expected 1 observability event, got %d", len(observer.events))
	}
	if observer.events[0].Name != ports.EventHealthChecked {
		t.Fatalf("expected event %q, got %q", ports.EventHealthChecked, observer.events[0].Name)
	}
}

func TestIndexEndpointReturnsHelpfulLinks(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{}, "unused-id"))

	response := performRequest(server, http.MethodGet, "/", "", nil)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body struct {
		Data indexResponse `json:"data"`
		Meta responseMeta  `json:"meta"`
	}
	decodeBody(t, response, &body)

	if body.Data.Service != "stuff-stash" {
		t.Fatalf("expected service stuff-stash, got %q", body.Data.Service)
	}
	if body.Data.Links.Health != "/healthz" || body.Data.Links.OpenAPI != "/openapi.json" || body.Data.Links.Docs != "/docs" {
		t.Fatalf("unexpected index links: %+v", body.Data.Links)
	}
}

func TestUnknownGetPathStillReturnsNotFound(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{}, "unused-id"))

	response := performRequest(server, http.MethodGet, "/missing", "", nil)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusNotFound, response.Code, response.Body.String())
	}
}

func TestOpenAPIIsGenerated(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{}, "unused-id"))

	response := performRequest(server, http.MethodGet, "/openapi.json", "", nil)

	if response.Code != http.StatusOK {
		t.Fatalf("expected OpenAPI status %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body struct {
		Paths      map[string]any `json:"paths"`
		Components struct {
			SecuritySchemes map[string]any `json:"securitySchemes"`
		} `json:"components"`
	}
	decodeBody(t, response, &body)
	if _, ok := body.Paths["/tenants/{tenantId}/inventories"]; !ok {
		t.Fatalf("expected OpenAPI to include inventory path, got %s", response.Body.String())
	}
	if _, ok := body.Paths["/tenants/{tenantId}/inventories/{inventoryId}/assets"]; !ok {
		t.Fatalf("expected OpenAPI to include asset path, got %s", response.Body.String())
	}
	if _, ok := body.Paths["/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}"]; !ok {
		t.Fatalf("expected OpenAPI to include asset update path, got %s", response.Body.String())
	}
	if _, ok := body.Paths["/tenants/{tenantId}/inventories/{inventoryId}/access-grants"]; !ok {
		t.Fatalf("expected OpenAPI to include inventory access grant path, got %s", response.Body.String())
	}
	if _, ok := body.Paths["/tenants/{tenantId}/custom-field-definitions"]; !ok {
		t.Fatalf("expected OpenAPI to include tenant custom field definition path, got %s", response.Body.String())
	}
	if _, ok := body.Paths["/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions"]; !ok {
		t.Fatalf("expected OpenAPI to include inventory custom field definition path, got %s", response.Body.String())
	}
	if _, ok := body.Paths["/tenants/{tenantId}/audit-records"]; !ok {
		t.Fatalf("expected OpenAPI to include tenant audit records path, got %s", response.Body.String())
	}
	if _, ok := body.Paths["/tenants/{tenantId}/inventories/{inventoryId}/audit-records"]; !ok {
		t.Fatalf("expected OpenAPI to include inventory audit records path, got %s", response.Body.String())
	}
	if _, ok := body.Paths["/"]; ok {
		t.Fatalf("expected OpenAPI to omit local API index path, got %s", response.Body.String())
	}
	if _, ok := body.Components.SecuritySchemes["bearerAuth"]; !ok {
		t.Fatalf("expected OpenAPI to include bearer auth, got %+v", body.Components.SecuritySchemes)
	}
}
