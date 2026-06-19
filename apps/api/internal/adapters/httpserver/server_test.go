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
		Paths      map[string]map[string]any `json:"paths"`
		Components struct {
			SecuritySchemes map[string]any `json:"securitySchemes"`
		} `json:"components"`
	}
	decodeBody(t, response, &body)
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/inventories", "post")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/inventories", "get")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/inventories/{inventoryId}/assets", "post")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/inventories/{inventoryId}/assets", "get")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}", "patch")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/inventories/{inventoryId}/access-grants", "post")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/inventories/{inventoryId}/access-grants", "get")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/custom-field-definitions", "post")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/custom-field-definitions", "get")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/custom-field-definitions/{definitionId}", "patch")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions", "post")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions", "get")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}", "patch")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/audit-records", "get")
	assertOpenAPIPathMethod(t, body.Paths, "/tenants/{tenantId}/inventories/{inventoryId}/audit-records", "get")
	if _, ok := body.Paths["/"]; ok {
		t.Fatalf("expected OpenAPI to omit local API index path, got %s", response.Body.String())
	}
	if _, ok := body.Components.SecuritySchemes["bearerAuth"]; !ok {
		t.Fatalf("expected OpenAPI to include bearer auth, got %+v", body.Components.SecuritySchemes)
	}
}

func assertOpenAPIPathMethod(t *testing.T, paths map[string]map[string]any, path string, method string) {
	t.Helper()

	operations, ok := paths[path]
	if !ok {
		t.Fatalf("expected OpenAPI to include path %s", path)
	}
	if _, ok := operations[method]; !ok {
		t.Fatalf("expected OpenAPI path %s to include method %s, got %+v", path, method, operations)
	}
}
