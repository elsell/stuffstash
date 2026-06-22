package httpserver

import (
	"net/http"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/ports"
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
	expectedOperations := []struct {
		path   string
		method string
	}{
		{"/me/tenants", "get"},
		{"/tenants/{tenantId}", "get"},
		{"/tenants/{tenantId}", "patch"},
		{"/tenants/{tenantId}", "delete"},
		{"/tenants/{tenantId}/archive", "patch"},
		{"/tenants/{tenantId}/restore", "patch"},
		{"/tenants/{tenantId}/inventories", "post"},
		{"/tenants/{tenantId}/inventories", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}", "delete"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/archive", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/restore", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets", "post"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}", "delete"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/archive", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/restore", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments", "post"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads", "post"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads/{uploadId}/complete", "post"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}", "delete"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/archive", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/restore", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/content", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/thumbnail", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/access-grants", "post"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/access-grants", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/access-grants/{principalId}/{relationship}", "delete"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/access-invitations", "post"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/access-invitations", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}", "delete"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/accept", "post"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/expiration", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/access-invitations/{invitationId}/cancel", "patch"},
		{"/tenants/{tenantId}/custom-asset-types", "post"},
		{"/tenants/{tenantId}/custom-asset-types", "get"},
		{"/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}", "get"},
		{"/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}", "patch"},
		{"/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}", "delete"},
		{"/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}/archive", "patch"},
		{"/tenants/{tenantId}/custom-asset-types/{customAssetTypeId}/restore", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types", "post"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}", "delete"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}/archive", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-asset-types/{customAssetTypeId}/restore", "patch"},
		{"/tenants/{tenantId}/custom-field-definitions", "post"},
		{"/tenants/{tenantId}/custom-field-definitions", "get"},
		{"/tenants/{tenantId}/custom-field-definitions/{definitionId}", "get"},
		{"/tenants/{tenantId}/custom-field-definitions/{definitionId}", "patch"},
		{"/tenants/{tenantId}/custom-field-definitions/{definitionId}", "delete"},
		{"/tenants/{tenantId}/custom-field-definitions/{definitionId}/archive", "patch"},
		{"/tenants/{tenantId}/custom-field-definitions/{definitionId}/restore", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions", "post"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}", "delete"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}/archive", "patch"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/custom-field-definitions/{definitionId}/restore", "patch"},
		{"/tenants/{tenantId}/audit-records", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/audit-records", "get"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/undoable-operations/{operationId}/undo", "post"},
		{"/tenants/{tenantId}/inventories/{inventoryId}/undoable-operations/{operationId}/redo", "post"},
	}
	for _, operation := range expectedOperations {
		assertOpenAPIPathMethod(t, body.Paths, operation.path, operation.method)
	}
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
