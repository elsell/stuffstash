package httpserver

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
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

func TestProtectedEndpointsRejectMissingAndMalformedAuthentication(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{}, "01ARZ3NDEKTSV4RRFFQ69G5FAV"))

	endpoints := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{name: "current principal", method: http.MethodGet, path: "/me"},
		{name: "create tenant", method: http.MethodPost, path: "/tenants", body: map[string]string{"name": "Home"}},
		{name: "create inventory", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories", body: map[string]string{"name": "Tools"}},
		{name: "list inventories", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories"},
		{name: "create tenant custom field definition", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/custom-field-definitions", body: map[string]string{"key": "serial", "displayName": "Serial", "type": "text"}},
		{name: "list tenant custom field definitions", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/custom-field-definitions"},
		{name: "list tenant audit records", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/audit-records"},
		{name: "create asset", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/assets", body: map[string]string{"kind": "item", "title": "Drill"}},
		{name: "list assets", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/assets"},
		{name: "update asset", method: http.MethodPatch, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/assets/01ARZ3NDEKTSV4RRFFQ69G5FAX", body: map[string]string{"title": "Drill"}},
		{name: "grant inventory access", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/access-grants", body: map[string]string{"principalId": "viewer", "relationship": "viewer"}},
		{name: "list inventory access grants", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/access-grants"},
		{name: "create inventory custom field definition", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/custom-field-definitions", body: map[string]any{"key": "condition", "displayName": "Condition", "type": "enum", "enumOptions": []string{"new", "used"}}},
		{name: "list inventory custom field definitions", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/custom-field-definitions"},
		{name: "list inventory audit records", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/audit-records"},
	}

	authCases := []struct {
		name          string
		authorization string
	}{
		{name: "missing token"},
		{name: "malformed token", authorization: "Bearer nope"},
		{name: "unsupported scheme", authorization: "Basic dev:user-one"},
		{name: "empty principal", authorization: "Bearer dev:"},
		{name: "unsafe principal", authorization: "Bearer dev:user/one"},
	}

	for _, endpoint := range endpoints {
		for _, authCase := range authCases {
			t.Run(endpoint.name+" "+authCase.name, func(t *testing.T) {
				response := performRequest(server, endpoint.method, endpoint.path, authCase.authorization, endpoint.body)

				if response.Code != http.StatusUnauthorized {
					t.Fatalf("expected status %d, got %d with body %s", http.StatusUnauthorized, response.Code, response.Body.String())
				}

				assertSafeError(t, response, "authentication_required", "Authentication required.")
			})
		}
	}
}

func TestSecureTenantInventoryFlow(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"

	server := NewServer(":0", newTestApp(&fakeObserver{},
		tenantID, "tenant-event", "audit-tenant", "tenant-claim",
		inventoryID, "inventory-event", "audit-inventory", "inventory-claim",
		"location-one", "audit-location", "asset-one", "audit-asset",
	))

	me := performRequest(server, http.MethodGet, "/me", "Bearer dev:user-one", nil)
	if me.Code != http.StatusOK {
		t.Fatalf("expected /me status %d, got %d with body %s", http.StatusOK, me.Code, me.Body.String())
	}

	createTenant := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:user-one", map[string]string{"name": "Home"})
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected create tenant status %d, got %d with body %s", http.StatusCreated, createTenant.Code, createTenant.Body.String())
	}

	var tenantBody struct {
		Data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	decodeBody(t, createTenant, &tenantBody)
	if tenantBody.Data.ID != tenantID {
		t.Fatalf("expected tenant ID %q, got %q", tenantID, tenantBody.Data.ID)
	}

	createInventory := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", "Bearer dev:user-one", map[string]string{"name": "Tools"})
	if createInventory.Code != http.StatusCreated {
		t.Fatalf("expected create inventory status %d, got %d with body %s", http.StatusCreated, createInventory.Code, createInventory.Body.String())
	}

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories", "Bearer dev:user-one", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, list.Code, list.Body.String())
	}

	var listBody struct {
		Data []struct {
			ID       string `json:"id"`
			TenantID string `json:"tenantId"`
			Name     string `json:"name"`
		} `json:"data"`
	}
	decodeBody(t, list, &listBody)
	if len(listBody.Data) != 1 {
		t.Fatalf("expected 1 inventory, got %d", len(listBody.Data))
	}
	if listBody.Data[0].ID != inventoryID {
		t.Fatalf("expected inventory ID %q, got %q", inventoryID, listBody.Data[0].ID)
	}

	createLocation := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:user-one", map[string]string{
		"kind":  "location",
		"title": "Garage",
	})
	if createLocation.Code != http.StatusCreated {
		t.Fatalf("expected create location status %d, got %d with body %s", http.StatusCreated, createLocation.Code, createLocation.Body.String())
	}
	locationBody := decodeAsset(t, createLocation)
	if locationBody.Data.Kind != "location" || locationBody.Data.LifecycleState != "active" {
		t.Fatalf("unexpected location response: %+v", locationBody.Data)
	}

	createAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:user-one", map[string]string{
		"kind":          "item",
		"title":         "Drill",
		"description":   "Cordless",
		"parentAssetId": locationBody.Data.ID,
	})
	if createAsset.Code != http.StatusCreated {
		t.Fatalf("expected create asset status %d, got %d with body %s", http.StatusCreated, createAsset.Code, createAsset.Body.String())
	}

	assets := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:user-one", nil)
	if assets.Code != http.StatusOK {
		t.Fatalf("expected list assets status %d, got %d with body %s", http.StatusOK, assets.Code, assets.Body.String())
	}
	assetList := decodeAssetList(t, assets)
	if len(assetList.Data) != 2 {
		t.Fatalf("expected 2 assets, got %+v", assetList.Data)
	}

	firstAssetPage := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?limit=1", "Bearer dev:user-one", nil)
	if firstAssetPage.Code != http.StatusOK {
		t.Fatalf("expected first asset page status %d, got %d with body %s", http.StatusOK, firstAssetPage.Code, firstAssetPage.Body.String())
	}
	firstPage := decodeAssetList(t, firstAssetPage)
	if len(firstPage.Data) != 1 || firstPage.Meta.Pagination == nil || firstPage.Meta.Pagination.Limit != 1 || !firstPage.Meta.Pagination.HasMore || firstPage.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected first paginated asset page, got %+v", firstPage)
	}

	secondAssetPage := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?limit=1&cursor="+*firstPage.Meta.Pagination.NextCursor, "Bearer dev:user-one", nil)
	if secondAssetPage.Code != http.StatusOK {
		t.Fatalf("expected second asset page status %d, got %d with body %s", http.StatusOK, secondAssetPage.Code, secondAssetPage.Body.String())
	}
	if !bytes.Contains(secondAssetPage.Body.Bytes(), []byte(`"nextCursor":null`)) {
		t.Fatalf("expected final asset page to include null nextCursor, got %s", secondAssetPage.Body.String())
	}
	secondPage := decodeAssetList(t, secondAssetPage)
	if len(secondPage.Data) != 1 || secondPage.Meta.Pagination == nil || secondPage.Meta.Pagination.Limit != 1 || secondPage.Meta.Pagination.HasMore || secondPage.Meta.Pagination.NextCursor != nil {
		t.Fatalf("expected final paginated asset page, got %+v", secondPage)
	}
}

func TestInventoryEndpointsDenyCrossUserAccess(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newTestApp(&fakeObserver{}, tenantID, "01ARZ3NDEKTSV4RRFFQ69G5FAW"))

	createTenant := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:owner", map[string]string{"name": "Home"})
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected create tenant status %d, got %d with body %s", http.StatusCreated, createTenant.Code, createTenant.Body.String())
	}

	createInventory := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", "Bearer dev:other-user", map[string]string{"name": "Tools"})
	if createInventory.Code != http.StatusForbidden {
		t.Fatalf("expected create inventory status %d, got %d with body %s", http.StatusForbidden, createInventory.Code, createInventory.Body.String())
	}

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories", "Bearer dev:other-user", nil)
	if list.Code != http.StatusForbidden {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusForbidden, list.Code, list.Body.String())
	}
	assertSafeError(t, list, "forbidden", "Forbidden.")
}

func TestAuthorizedUserCannotCrossTenantBoundaries(t *testing.T) {
	const tenantOneID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const tenantTwoID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantOneID, name: "Home", owner: "owner-one"},
			{id: tenantTwoID, name: "Cabin", owner: "owner-two"},
		},
		inventories: []seedInventory{
			{id: "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID: tenantOneID, name: "Tools", owner: "owner-one"},
			{id: "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID: tenantTwoID, name: "Supplies", owner: "owner-two"},
		},
	}))

	createInventory := performRequest(server, http.MethodPost, "/tenants/"+tenantOneID+"/inventories", "Bearer dev:owner-two", map[string]string{"name": "Cross Tenant"})
	if createInventory.Code != http.StatusForbidden {
		t.Fatalf("expected create status %d, got %d with body %s", http.StatusForbidden, createInventory.Code, createInventory.Body.String())
	}
	assertSafeError(t, createInventory, "forbidden", "Forbidden.")

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantOneID+"/inventories", "Bearer dev:owner-two", nil)
	if list.Code != http.StatusForbidden {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusForbidden, list.Code, list.Body.String())
	}
	assertSafeError(t, list, "forbidden", "Forbidden.")
}

func TestAssetEndpointsRejectCrossInventoryAndInvalidCustomFields(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryOneID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const inventoryTwoID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryOneID, tenantID: tenantID, name: "Tools", owner: "owner"},
			{id: inventoryTwoID, tenantID: tenantID, name: "Supplies", owner: "owner"},
		},
	}))

	createParent := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryOneID+"/assets", "Bearer dev:owner", map[string]string{
		"kind":  "location",
		"title": "Garage",
	})
	if createParent.Code != http.StatusCreated {
		t.Fatalf("expected parent status %d, got %d with body %s", http.StatusCreated, createParent.Code, createParent.Body.String())
	}
	parent := decodeAsset(t, createParent)

	crossInventory := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryTwoID+"/assets", "Bearer dev:owner", map[string]string{
		"kind":          "item",
		"title":         "Fertilizer",
		"parentAssetId": parent.Data.ID,
	})
	if crossInventory.Code != http.StatusNotFound {
		t.Fatalf("expected cross-inventory parent status %d, got %d with body %s", http.StatusNotFound, crossInventory.Code, crossInventory.Body.String())
	}

	customFields := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryOneID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":         "item",
		"title":        "Drill",
		"customFields": map[string]any{"serial": "abc"},
	})
	if customFields.Code != http.StatusBadRequest {
		t.Fatalf("expected custom fields status %d, got %d with body %s", http.StatusBadRequest, customFields.Code, customFields.Body.String())
	}
	assertSafeError(t, customFields, "invalid_request", "Invalid request.")
}

func TestUnrelatedUserCannotCreateOrListAssets(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
	}))

	createAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:intruder", map[string]string{
		"kind":  "item",
		"title": "Drill",
	})
	if createAsset.Code != http.StatusForbidden {
		t.Fatalf("expected create status %d, got %d with body %s", http.StatusForbidden, createAsset.Code, createAsset.Body.String())
	}
	assertSafeError(t, createAsset, "forbidden", "Forbidden.")

	listAssets := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:intruder", nil)
	if listAssets.Code != http.StatusForbidden {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusForbidden, listAssets.Code, listAssets.Body.String())
	}
	assertSafeError(t, listAssets, "forbidden", "Forbidden.")
}

func TestAssetUpdateFlowAndMovement(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const otherInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	const otherTenantInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAZ"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
			{id: otherTenantID, name: "Cabin", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
			{id: otherInventoryID, tenantID: tenantID, name: "Supplies", owner: "owner"},
			{id: otherTenantInventoryID, tenantID: otherTenantID, name: "Cabin Tools", owner: "other-owner"},
		},
		ids: []string{
			"garage", "audit-garage",
			"shelf", "audit-shelf",
			"box", "audit-box",
			"wrench", "audit-wrench",
			"audit-box-update", "audit-box-move",
			"other-inventory-location", "audit-other-inventory-location",
			"other-tenant-location", "audit-other-tenant-location",
			"audit-box-root-move",
			"viewer-grant-event", "audit-viewer-grant", "viewer-grant-claim",
			"editor-grant-event", "audit-editor-grant", "editor-grant-claim",
			"audit-editor-update",
		},
	}))

	garageResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "location",
		"title": "Garage",
	})
	if garageResponse.Code != http.StatusCreated {
		t.Fatalf("expected garage create status %d, got %d with body %s", http.StatusCreated, garageResponse.Code, garageResponse.Body.String())
	}
	garage := decodeAsset(t, garageResponse)

	shelfResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":          "location",
		"title":         "Shelf",
		"parentAssetId": garage.Data.ID,
	})
	if shelfResponse.Code != http.StatusCreated {
		t.Fatalf("expected shelf create status %d, got %d with body %s", http.StatusCreated, shelfResponse.Code, shelfResponse.Body.String())
	}
	shelf := decodeAsset(t, shelfResponse)

	boxResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":          "container",
		"title":         "Toolbox",
		"parentAssetId": shelf.Data.ID,
	})
	if boxResponse.Code != http.StatusCreated {
		t.Fatalf("expected box create status %d, got %d with body %s", http.StatusCreated, boxResponse.Code, boxResponse.Body.String())
	}
	box := decodeAsset(t, boxResponse)

	wrenchResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":          "item",
		"title":         "Wrench",
		"parentAssetId": box.Data.ID,
	})
	if wrenchResponse.Code != http.StatusCreated {
		t.Fatalf("expected wrench create status %d, got %d with body %s", http.StatusCreated, wrenchResponse.Code, wrenchResponse.Body.String())
	}
	wrench := decodeAsset(t, wrenchResponse)

	moveBox := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+box.Data.ID, "Bearer dev:owner", map[string]any{
		"title":         "Moved Toolbox",
		"description":   "Blue metal box",
		"parentAssetId": garage.Data.ID,
	})
	if moveBox.Code != http.StatusOK {
		t.Fatalf("expected move status %d, got %d with body %s", http.StatusOK, moveBox.Code, moveBox.Body.String())
	}
	movedBox := decodeAsset(t, moveBox)
	if movedBox.Data.Title != "Moved Toolbox" || movedBox.Data.Description != "Blue metal box" || movedBox.Data.ParentAssetID != garage.Data.ID {
		t.Fatalf("unexpected moved box response: %+v", movedBox.Data)
	}

	assetsAfterMove := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?limit=50", "Bearer dev:owner", nil)
	if assetsAfterMove.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, assetsAfterMove.Code, assetsAfterMove.Body.String())
	}
	listAfterMove := decodeAssetList(t, assetsAfterMove)
	foundWrench := false
	for _, item := range listAfterMove.Data {
		if item.ID == wrench.Data.ID {
			foundWrench = true
			if item.ParentAssetID != box.Data.ID {
				t.Fatalf("expected wrench to remain inside moved box, got %+v", item)
			}
		}
	}
	if !foundWrench {
		t.Fatalf("expected wrench in asset list, got %+v", listAfterMove.Data)
	}

	blankParent := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+box.Data.ID, "Bearer dev:owner", map[string]any{
		"parentAssetId": "   ",
	})
	if blankParent.Code != http.StatusBadRequest {
		t.Fatalf("expected blank parent status %d, got %d with body %s", http.StatusBadRequest, blankParent.Code, blankParent.Body.String())
	}
	assertSafeError(t, blankParent, "invalid_request", "Invalid request.")

	otherInventoryLocationResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+otherInventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "location",
		"title": "Other Inventory Shelf",
	})
	if otherInventoryLocationResponse.Code != http.StatusCreated {
		t.Fatalf("expected other inventory location status %d, got %d with body %s", http.StatusCreated, otherInventoryLocationResponse.Code, otherInventoryLocationResponse.Body.String())
	}
	otherInventoryLocation := decodeAsset(t, otherInventoryLocationResponse)
	crossInventoryMove := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+box.Data.ID, "Bearer dev:owner", map[string]any{
		"parentAssetId": otherInventoryLocation.Data.ID,
	})
	if crossInventoryMove.Code != http.StatusNotFound {
		t.Fatalf("expected cross-inventory move status %d, got %d with body %s", http.StatusNotFound, crossInventoryMove.Code, crossInventoryMove.Body.String())
	}
	assertSafeError(t, crossInventoryMove, "resource_not_found", "Resource not found.")

	otherTenantLocationResponse := performRequest(server, http.MethodPost, "/tenants/"+otherTenantID+"/inventories/"+otherTenantInventoryID+"/assets", "Bearer dev:other-owner", map[string]any{
		"kind":  "location",
		"title": "Cabin Shelf",
	})
	if otherTenantLocationResponse.Code != http.StatusCreated {
		t.Fatalf("expected other tenant location status %d, got %d with body %s", http.StatusCreated, otherTenantLocationResponse.Code, otherTenantLocationResponse.Body.String())
	}
	otherTenantLocation := decodeAsset(t, otherTenantLocationResponse)
	crossTenantMove := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+box.Data.ID, "Bearer dev:owner", map[string]any{
		"parentAssetId": otherTenantLocation.Data.ID,
	})
	if crossTenantMove.Code != http.StatusNotFound {
		t.Fatalf("expected cross-tenant move status %d, got %d with body %s", http.StatusNotFound, crossTenantMove.Code, crossTenantMove.Body.String())
	}
	assertSafeError(t, crossTenantMove, "resource_not_found", "Resource not found.")

	moveBoxToRoot := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+box.Data.ID, "Bearer dev:owner", map[string]any{
		"parentAssetId": nil,
	})
	if moveBoxToRoot.Code != http.StatusOK {
		t.Fatalf("expected root move status %d, got %d with body %s", http.StatusOK, moveBoxToRoot.Code, moveBoxToRoot.Body.String())
	}
	rootBox := decodeAsset(t, moveBoxToRoot)
	if rootBox.Data.ParentAssetID != "" {
		t.Fatalf("expected box at root, got %+v", rootBox.Data)
	}

	cycle := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+garage.Data.ID, "Bearer dev:owner", map[string]any{
		"parentAssetId": shelf.Data.ID,
	})
	if cycle.Code != http.StatusBadRequest {
		t.Fatalf("expected cycle status %d, got %d with body %s", http.StatusBadRequest, cycle.Code, cycle.Body.String())
	}
	assertSafeError(t, cycle, "invalid_request", "Invalid request.")

	itemParent := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+box.Data.ID, "Bearer dev:owner", map[string]any{
		"parentAssetId": wrench.Data.ID,
	})
	if itemParent.Code != http.StatusBadRequest {
		t.Fatalf("expected item parent status %d, got %d with body %s", http.StatusBadRequest, itemParent.Code, itemParent.Body.String())
	}
	assertSafeError(t, itemParent, "invalid_request", "Invalid request.")

	viewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]string{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if viewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, viewerGrant.Code, viewerGrant.Body.String())
	}
	viewerUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+box.Data.ID, "Bearer dev:viewer-user", map[string]string{"title": "Viewer Rename"})
	if viewerUpdate.Code != http.StatusForbidden {
		t.Fatalf("expected viewer update status %d, got %d with body %s", http.StatusForbidden, viewerUpdate.Code, viewerUpdate.Body.String())
	}
	assertSafeError(t, viewerUpdate, "forbidden", "Forbidden.")

	intruderUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+box.Data.ID, "Bearer dev:intruder", map[string]string{"title": "Intruder Rename"})
	if intruderUpdate.Code != http.StatusForbidden {
		t.Fatalf("expected intruder update status %d, got %d with body %s", http.StatusForbidden, intruderUpdate.Code, intruderUpdate.Body.String())
	}
	assertSafeError(t, intruderUpdate, "forbidden", "Forbidden.")

	crossTenantPrincipalUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+box.Data.ID, "Bearer dev:other-owner", map[string]string{"title": "Other Tenant Rename"})
	if crossTenantPrincipalUpdate.Code != http.StatusForbidden {
		t.Fatalf("expected cross-tenant principal update status %d, got %d with body %s", http.StatusForbidden, crossTenantPrincipalUpdate.Code, crossTenantPrincipalUpdate.Body.String())
	}
	assertSafeError(t, crossTenantPrincipalUpdate, "forbidden", "Forbidden.")

	editorGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]string{
		"principalId":  "editor-user",
		"relationship": "editor",
	})
	if editorGrant.Code != http.StatusCreated {
		t.Fatalf("expected editor grant status %d, got %d with body %s", http.StatusCreated, editorGrant.Code, editorGrant.Body.String())
	}
	editorUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+box.Data.ID, "Bearer dev:editor-user", map[string]string{"title": "Editor Rename"})
	if editorUpdate.Code != http.StatusOK {
		t.Fatalf("expected editor update status %d, got %d with body %s", http.StatusOK, editorUpdate.Code, editorUpdate.Body.String())
	}
	if decodeAsset(t, editorUpdate).Data.Title != "Editor Rename" {
		t.Fatalf("expected editor title update, got %s", editorUpdate.Body.String())
	}
}

func TestAuditRecordEndpointsEnforceScopeAndPagination(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const siblingInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	const otherInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAZ"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
			{id: otherTenantID, name: "Cabin", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
			{id: siblingInventoryID, tenantID: tenantID, name: "Medicine", owner: "owner"},
			{id: otherInventoryID, tenantID: otherTenantID, name: "Cabin Tools", owner: "other-owner"},
		},
		ids: []string{
			"asset-one", "audit-asset-one",
			"asset-two", "audit-asset-two",
			"sibling-asset", "audit-sibling-asset",
			"other-tenant-asset", "audit-other-tenant-asset",
			"audit-viewer-grant", "viewer-grant-event", "viewer-grant-claim",
		},
	}))

	firstAsset := performRequestWithHeaders(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]string{"X-Request-ID": "request-audit-one"}, map[string]any{
		"kind":  "item",
		"title": "Drill",
	})
	if firstAsset.Code != http.StatusCreated {
		t.Fatalf("expected first asset status %d, got %d with body %s", http.StatusCreated, firstAsset.Code, firstAsset.Body.String())
	}
	secondAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Hammer",
	})
	if secondAsset.Code != http.StatusCreated {
		t.Fatalf("expected second asset status %d, got %d with body %s", http.StatusCreated, secondAsset.Code, secondAsset.Body.String())
	}
	siblingAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+siblingInventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Bandages",
	})
	if siblingAsset.Code != http.StatusCreated {
		t.Fatalf("expected sibling asset status %d, got %d with body %s", http.StatusCreated, siblingAsset.Code, siblingAsset.Body.String())
	}
	otherTenantAsset := performRequest(server, http.MethodPost, "/tenants/"+otherTenantID+"/inventories/"+otherInventoryID+"/assets", "Bearer dev:other-owner", map[string]any{
		"kind":  "item",
		"title": "Saw",
	})
	if otherTenantAsset.Code != http.StatusCreated {
		t.Fatalf("expected other tenant asset status %d, got %d with body %s", http.StatusCreated, otherTenantAsset.Code, otherTenantAsset.Body.String())
	}

	viewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]string{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if viewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, viewerGrant.Code, viewerGrant.Body.String())
	}

	firstPageResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records?limit=1", "Bearer dev:viewer-user", nil)
	if firstPageResponse.Code != http.StatusOK {
		t.Fatalf("expected first audit page status %d, got %d with body %s", http.StatusOK, firstPageResponse.Code, firstPageResponse.Body.String())
	}
	firstPage := decodeAuditRecordList(t, firstPageResponse)
	if len(firstPage.Data) != 1 || firstPage.Meta.Pagination == nil || !firstPage.Meta.Pagination.HasMore || firstPage.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected paginated first audit page, got %+v", firstPage)
	}
	if firstPage.Data[0].Action != "asset.created" || firstPage.Data[0].Source != "api" || firstPage.Data[0].TargetType != "asset" {
		t.Fatalf("unexpected first audit record: %+v", firstPage.Data[0])
	}
	if firstPage.Data[0].RequestID != "request-audit-one" {
		t.Fatalf("expected request ID on audit record, got %+v", firstPage.Data[0])
	}

	secondPageResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records?limit=1&cursor="+*firstPage.Meta.Pagination.NextCursor, "Bearer dev:viewer-user", nil)
	if secondPageResponse.Code != http.StatusOK {
		t.Fatalf("expected second audit page status %d, got %d with body %s", http.StatusOK, secondPageResponse.Code, secondPageResponse.Body.String())
	}
	secondPage := decodeAuditRecordList(t, secondPageResponse)
	if len(secondPage.Data) != 1 {
		t.Fatalf("expected one record on second audit page, got %+v", secondPage)
	}

	firstInventoryAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records?limit=50", "Bearer dev:owner", nil)
	if firstInventoryAudit.Code != http.StatusOK {
		t.Fatalf("expected first inventory audit status %d, got %d with body %s", http.StatusOK, firstInventoryAudit.Code, firstInventoryAudit.Body.String())
	}
	firstInventoryAuditBody := decodeAuditRecordList(t, firstInventoryAudit)
	if auditRecordsContainTarget(firstInventoryAuditBody.Data, "sibling-asset") || auditRecordsContainTarget(firstInventoryAuditBody.Data, "other-tenant-asset") {
		t.Fatalf("expected first inventory audit to exclude sibling and other tenant records, got %+v", firstInventoryAuditBody.Data)
	}

	tenantAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/audit-records?limit=50", "Bearer dev:owner", nil)
	if tenantAudit.Code != http.StatusOK {
		t.Fatalf("expected tenant audit status %d, got %d with body %s", http.StatusOK, tenantAudit.Code, tenantAudit.Body.String())
	}
	tenantAuditBody := decodeAuditRecordList(t, tenantAudit)
	if len(tenantAuditBody.Data) < 3 {
		t.Fatalf("expected tenant audit to include state changes, got %+v", tenantAuditBody.Data)
	}
	if auditRecordsContainTarget(tenantAuditBody.Data, "other-tenant-asset") {
		t.Fatalf("expected tenant audit to exclude other tenant records, got %+v", tenantAuditBody.Data)
	}

	viewerTenantAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/audit-records", "Bearer dev:viewer-user", nil)
	if viewerTenantAudit.Code != http.StatusForbidden {
		t.Fatalf("expected viewer tenant audit status %d, got %d with body %s", http.StatusForbidden, viewerTenantAudit.Code, viewerTenantAudit.Body.String())
	}
	assertSafeError(t, viewerTenantAudit, "forbidden", "Forbidden.")

	intruderInventoryAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records", "Bearer dev:intruder", nil)
	if intruderInventoryAudit.Code != http.StatusForbidden {
		t.Fatalf("expected intruder inventory audit status %d, got %d with body %s", http.StatusForbidden, intruderInventoryAudit.Code, intruderInventoryAudit.Body.String())
	}
	assertSafeError(t, intruderInventoryAudit, "forbidden", "Forbidden.")

	crossTenantAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/audit-records", "Bearer dev:other-owner", nil)
	if crossTenantAudit.Code != http.StatusForbidden {
		t.Fatalf("expected cross-tenant audit status %d, got %d with body %s", http.StatusForbidden, crossTenantAudit.Code, crossTenantAudit.Body.String())
	}
	assertSafeError(t, crossTenantAudit, "forbidden", "Forbidden.")

	wrongScopeCursor := paginationCursor(map[string]any{
		"v":          1,
		"collection": "audit_records",
		"scope":      tenantID + ":" + siblingInventoryID,
		"lastId":     firstPage.Data[0].ID,
	})
	wrongScopeAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records?cursor="+wrongScopeCursor, "Bearer dev:owner", nil)
	if wrongScopeAudit.Code != http.StatusBadRequest {
		t.Fatalf("expected wrong-scope cursor status %d, got %d with body %s", http.StatusBadRequest, wrongScopeAudit.Code, wrongScopeAudit.Body.String())
	}
	assertSafeError(t, wrongScopeAudit, "invalid_request", "Invalid request.")
}

func TestCustomFieldDefinitionFlowAndAssetValidation(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FB1"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const hiddenInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
			{id: otherTenantID, name: "Cabin", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "inventory-owner"},
			{id: hiddenInventoryID, tenantID: tenantID, name: "Hidden", owner: "hidden-owner"},
		},
		ids: []string{
			"01ARZ3NDEKTSV4RRFFQ69G5FAY", "audit-tenant-definition",
			"01ARZ3NDEKTSV4RRFFQ69G5FAZ", "audit-inventory-definition",
			"01ARZ3NDEKTSV4RRFFQ69G5FB0", "audit-duplicate-definition",
			"01ARZ3NDEKTSV4RRFFQ69G5FB2", "audit-tenant-conflict-definition",
			"01ARZ3NDEKTSV4RRFFQ69G5FB3", "audit-custom-field-asset",
		},
	}))

	tenantDefinitionResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/custom-field-definitions", "Bearer dev:tenant-owner", map[string]any{
		"key":         "serial",
		"displayName": "Serial",
		"type":        "text",
	})
	if tenantDefinitionResponse.Code != http.StatusCreated {
		t.Fatalf("expected tenant definition status %d, got %d with body %s", http.StatusCreated, tenantDefinitionResponse.Code, tenantDefinitionResponse.Body.String())
	}
	tenantDefinition := decodeCustomFieldDefinition(t, tenantDefinitionResponse)
	assertCustomFieldDefinition(t, tenantDefinition.Data, tenantID, "", "tenant", "serial", "text")

	tenantDefinitionList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/custom-field-definitions?limit=50", "Bearer dev:tenant-owner", nil)
	if tenantDefinitionList.Code != http.StatusOK {
		t.Fatalf("expected tenant definition list status %d, got %d with body %s", http.StatusOK, tenantDefinitionList.Code, tenantDefinitionList.Body.String())
	}

	inventoryDefinitionResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions", "Bearer dev:inventory-owner", map[string]any{
		"key":         "condition",
		"displayName": "Condition",
		"type":        "enum",
		"enumOptions": []string{"new", "used"},
	})
	if inventoryDefinitionResponse.Code != http.StatusCreated {
		t.Fatalf("expected inventory definition status %d, got %d with body %s", http.StatusCreated, inventoryDefinitionResponse.Code, inventoryDefinitionResponse.Body.String())
	}
	inventoryDefinition := decodeCustomFieldDefinition(t, inventoryDefinitionResponse)
	assertCustomFieldDefinition(t, inventoryDefinition.Data, tenantID, inventoryID, "inventory", "condition", "enum")

	duplicateDefinition := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions", "Bearer dev:inventory-owner", map[string]any{
		"key":         "serial",
		"displayName": "Serial Again",
		"type":        "text",
	})
	if duplicateDefinition.Code != http.StatusBadRequest {
		t.Fatalf("expected duplicate definition status %d, got %d with body %s", http.StatusBadRequest, duplicateDefinition.Code, duplicateDefinition.Body.String())
	}
	assertSafeError(t, duplicateDefinition, "invalid_request", "Invalid request.")

	tenantConflictDefinition := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/custom-field-definitions", "Bearer dev:tenant-owner", map[string]any{
		"key":         "condition",
		"displayName": "Condition Again",
		"type":        "enum",
		"enumOptions": []string{"new", "used"},
	})
	if tenantConflictDefinition.Code != http.StatusBadRequest {
		t.Fatalf("expected tenant conflict definition status %d, got %d with body %s", http.StatusBadRequest, tenantConflictDefinition.Code, tenantConflictDefinition.Body.String())
	}
	assertSafeError(t, tenantConflictDefinition, "invalid_request", "Invalid request.")

	firstPageResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions?limit=1", "Bearer dev:inventory-owner", nil)
	if firstPageResponse.Code != http.StatusOK {
		t.Fatalf("expected first definition page status %d, got %d with body %s", http.StatusOK, firstPageResponse.Code, firstPageResponse.Body.String())
	}
	firstPage := decodeCustomFieldDefinitionList(t, firstPageResponse)
	if len(firstPage.Data) != 1 || firstPage.Data[0].Key != "serial" || firstPage.Meta.Pagination == nil || !firstPage.Meta.Pagination.HasMore || firstPage.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected first page with inherited tenant definition, got %+v", firstPage)
	}
	secondPageResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions?limit=1&cursor="+*firstPage.Meta.Pagination.NextCursor, "Bearer dev:inventory-owner", nil)
	if secondPageResponse.Code != http.StatusOK {
		t.Fatalf("expected second definition page status %d, got %d with body %s", http.StatusOK, secondPageResponse.Code, secondPageResponse.Body.String())
	}
	secondPage := decodeCustomFieldDefinitionList(t, secondPageResponse)
	if len(secondPage.Data) != 1 || secondPage.Data[0].Key != "condition" || secondPage.Meta.Pagination == nil || secondPage.Meta.Pagination.HasMore || secondPage.Meta.Pagination.NextCursor != nil {
		t.Fatalf("expected second page with inventory definition, got %+v", secondPage)
	}

	createAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:inventory-owner", map[string]any{
		"kind":  "item",
		"title": "Drill",
		"customFields": map[string]any{
			"serial":    "abc",
			"condition": "used",
		},
	})
	if createAsset.Code != http.StatusCreated {
		t.Fatalf("expected asset create status %d, got %d with body %s", http.StatusCreated, createAsset.Code, createAsset.Body.String())
	}
	assetBody := decodeAsset(t, createAsset)
	if assetBody.Data.CustomFields["serial"] != "abc" || assetBody.Data.CustomFields["condition"] != "used" {
		t.Fatalf("expected custom field values in asset response, got %+v", assetBody.Data.CustomFields)
	}

	badAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:inventory-owner", map[string]any{
		"kind":  "item",
		"title": "Bad Drill",
		"customFields": map[string]any{
			"condition": "broken",
		},
	})
	if badAsset.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid custom field value status %d, got %d with body %s", http.StatusBadRequest, badAsset.Code, badAsset.Body.String())
	}
	assertSafeError(t, badAsset, "invalid_request", "Invalid request.")

	for _, item := range []struct {
		name          string
		method        string
		path          string
		principal     string
		body          any
		expectedCode  int
		expectedError string
	}{
		{
			name:          "inventory owner cannot list tenant definitions",
			method:        http.MethodGet,
			path:          "/tenants/" + tenantID + "/custom-field-definitions",
			principal:     "inventory-owner",
			expectedCode:  http.StatusForbidden,
			expectedError: "forbidden",
		},
		{
			name:      "inventory owner cannot create tenant definitions",
			method:    http.MethodPost,
			path:      "/tenants/" + tenantID + "/custom-field-definitions",
			principal: "inventory-owner",
			body: map[string]any{
				"key":         "inventory-owned",
				"displayName": "Inventory Owned",
				"type":        "text",
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "forbidden",
		},
		{
			name:          "unrelated user cannot list tenant definitions",
			method:        http.MethodGet,
			path:          "/tenants/" + tenantID + "/custom-field-definitions",
			principal:     "intruder",
			expectedCode:  http.StatusForbidden,
			expectedError: "forbidden",
		},
		{
			name:      "unrelated user cannot create tenant definitions",
			method:    http.MethodPost,
			path:      "/tenants/" + tenantID + "/custom-field-definitions",
			principal: "intruder",
			body: map[string]any{
				"key":         "intruder-field",
				"displayName": "Intruder Field",
				"type":        "text",
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "forbidden",
		},
		{
			name:          "unrelated user cannot list inventory definitions",
			method:        http.MethodGet,
			path:          "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-field-definitions",
			principal:     "intruder",
			expectedCode:  http.StatusForbidden,
			expectedError: "forbidden",
		},
		{
			name:      "unrelated user cannot create inventory definitions",
			method:    http.MethodPost,
			path:      "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-field-definitions",
			principal: "intruder",
			body: map[string]any{
				"key":         "intruder-inventory-field",
				"displayName": "Intruder Inventory Field",
				"type":        "text",
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "forbidden",
		},
		{
			name:          "inventory owner cannot list hidden inventory definitions",
			method:        http.MethodGet,
			path:          "/tenants/" + tenantID + "/inventories/" + hiddenInventoryID + "/custom-field-definitions",
			principal:     "inventory-owner",
			expectedCode:  http.StatusForbidden,
			expectedError: "forbidden",
		},
		{
			name:      "inventory owner cannot create hidden inventory definitions",
			method:    http.MethodPost,
			path:      "/tenants/" + tenantID + "/inventories/" + hiddenInventoryID + "/custom-field-definitions",
			principal: "inventory-owner",
			body: map[string]any{
				"key":         "hidden-field",
				"displayName": "Hidden Field",
				"type":        "text",
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "forbidden",
		},
		{
			name:          "wrong tenant inventory list is hidden",
			method:        http.MethodGet,
			path:          "/tenants/" + otherTenantID + "/inventories/" + inventoryID + "/custom-field-definitions",
			principal:     "inventory-owner",
			expectedCode:  http.StatusNotFound,
			expectedError: "resource_not_found",
		},
		{
			name:      "wrong tenant inventory create is hidden",
			method:    http.MethodPost,
			path:      "/tenants/" + otherTenantID + "/inventories/" + inventoryID + "/custom-field-definitions",
			principal: "inventory-owner",
			body: map[string]any{
				"key":         "wrong-tenant-field",
				"displayName": "Wrong Tenant Field",
				"type":        "text",
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "resource_not_found",
		},
		{
			name:          "missing inventory list is hidden",
			method:        http.MethodGet,
			path:          "/tenants/" + tenantID + "/inventories/01ARZ3NDEKTSV4RRFFQ69G5FB2/custom-field-definitions",
			principal:     "inventory-owner",
			expectedCode:  http.StatusNotFound,
			expectedError: "resource_not_found",
		},
		{
			name:      "missing inventory create is hidden",
			method:    http.MethodPost,
			path:      "/tenants/" + tenantID + "/inventories/01ARZ3NDEKTSV4RRFFQ69G5FB2/custom-field-definitions",
			principal: "inventory-owner",
			body: map[string]any{
				"key":         "missing-inventory-field",
				"displayName": "Missing Inventory Field",
				"type":        "text",
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "resource_not_found",
		},
	} {
		t.Run(item.name, func(t *testing.T) {
			response := performRequest(server, item.method, item.path, "Bearer dev:"+item.principal, item.body)
			if response.Code != item.expectedCode {
				t.Fatalf("expected status %d, got %d with body %s", item.expectedCode, response.Code, response.Body.String())
			}
			assertErrorCode(t, response, item.expectedError)
		})
	}

	viewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if viewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, viewerGrant.Code, viewerGrant.Body.String())
	}
	viewerList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions?limit=50", "Bearer dev:viewer-user", nil)
	if viewerList.Code != http.StatusOK {
		t.Fatalf("expected viewer list status %d, got %d with body %s", http.StatusOK, viewerList.Code, viewerList.Body.String())
	}
	viewerCreate := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions", "Bearer dev:viewer-user", map[string]any{
		"key":         "viewer-field",
		"displayName": "Viewer Field",
		"type":        "text",
	})
	if viewerCreate.Code != http.StatusForbidden {
		t.Fatalf("expected viewer create status %d, got %d with body %s", http.StatusForbidden, viewerCreate.Code, viewerCreate.Body.String())
	}
	assertSafeError(t, viewerCreate, "forbidden", "Forbidden.")

	wrongScopeCursor := paginationCursor(map[string]any{
		"v":          1,
		"collection": "custom_field_definitions",
		"scope":      tenantID + ":" + hiddenInventoryID,
		"lastId":     "0:" + tenantDefinition.Data.ID,
	})
	wrongCursorList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions?cursor="+wrongScopeCursor, "Bearer dev:inventory-owner", nil)
	if wrongCursorList.Code != http.StatusBadRequest {
		t.Fatalf("expected wrong-scope cursor status %d, got %d with body %s", http.StatusBadRequest, wrongCursorList.Code, wrongCursorList.Body.String())
	}
	assertSafeError(t, wrongCursorList, "invalid_request", "Invalid request.")
}

func TestInventorySharingEnforcesViewerEditorAndShareBoundaries(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const hiddenInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
			{id: otherTenantID, name: "Cabin", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "inventory-owner"},
			{id: hiddenInventoryID, tenantID: tenantID, name: "Hidden", owner: "other-inventory-owner"},
		},
		ids: []string{
			"audit-viewer-grant", "viewer-grant-event", "viewer-grant-claim",
			"audit-duplicate-viewer-grant", "duplicate-viewer-grant-event",
			"audit-editor-grant", "editor-grant-event", "editor-grant-claim",
		},
	}))

	viewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if viewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, viewerGrant.Code, viewerGrant.Body.String())
	}
	assertInventoryAccessGrant(t, decodeInventoryAccessGrant(t, viewerGrant).Data, tenantID, inventoryID, "viewer-user", "viewer")

	duplicateViewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if duplicateViewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected duplicate viewer grant status %d, got %d with body %s", http.StatusCreated, duplicateViewerGrant.Code, duplicateViewerGrant.Body.String())
	}
	assertInventoryAccessGrant(t, decodeInventoryAccessGrant(t, duplicateViewerGrant).Data, tenantID, inventoryID, "viewer-user", "viewer")

	editorGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "editor-user",
		"relationship": "editor",
	})
	if editorGrant.Code != http.StatusCreated {
		t.Fatalf("expected editor grant status %d, got %d with body %s", http.StatusCreated, editorGrant.Code, editorGrant.Body.String())
	}
	assertInventoryAccessGrant(t, decodeInventoryAccessGrant(t, editorGrant).Data, tenantID, inventoryID, "editor-user", "editor")

	invalidRelationship := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "bad-grant-user",
		"relationship": "owner",
	})
	if invalidRelationship.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected invalid relationship status %d, got %d with body %s", http.StatusUnprocessableEntity, invalidRelationship.Code, invalidRelationship.Body.String())
	}
	assertErrorCode(t, invalidRelationship, "invalid_request")

	wrongTenantGrant := performRequest(server, http.MethodPost, "/tenants/"+otherTenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "wrong-tenant-user",
		"relationship": "viewer",
	})
	if wrongTenantGrant.Code != http.StatusNotFound {
		t.Fatalf("expected wrong tenant grant status %d, got %d with body %s", http.StatusNotFound, wrongTenantGrant.Code, wrongTenantGrant.Body.String())
	}
	assertSafeError(t, wrongTenantGrant, "resource_not_found", "Resource not found.")

	wrongInventoryGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/01ARZ3NDEKTSV4RRFFQ69G5FB0/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "wrong-inventory-user",
		"relationship": "viewer",
	})
	if wrongInventoryGrant.Code != http.StatusNotFound {
		t.Fatalf("expected wrong inventory grant status %d, got %d with body %s", http.StatusNotFound, wrongInventoryGrant.Code, wrongInventoryGrant.Body.String())
	}
	assertSafeError(t, wrongInventoryGrant, "resource_not_found", "Resource not found.")

	firstGrantPage := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants?limit=1", "Bearer dev:inventory-owner", nil)
	if firstGrantPage.Code != http.StatusOK {
		t.Fatalf("expected grant list status %d, got %d with body %s", http.StatusOK, firstGrantPage.Code, firstGrantPage.Body.String())
	}
	firstPage := decodeInventoryAccessGrantList(t, firstGrantPage)
	if len(firstPage.Data) != 1 || firstPage.Meta.Pagination == nil || firstPage.Meta.Pagination.Limit != 1 || !firstPage.Meta.Pagination.HasMore || firstPage.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected first grant page metadata, got %+v", firstPage)
	}
	assertInventoryAccessGrant(t, firstPage.Data[0], tenantID, inventoryID, "editor-user", "editor")

	secondGrantPage := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants?limit=1&cursor="+*firstPage.Meta.Pagination.NextCursor, "Bearer dev:inventory-owner", nil)
	if secondGrantPage.Code != http.StatusOK {
		t.Fatalf("expected second grant list status %d, got %d with body %s", http.StatusOK, secondGrantPage.Code, secondGrantPage.Body.String())
	}
	secondPage := decodeInventoryAccessGrantList(t, secondGrantPage)
	if len(secondPage.Data) != 1 || secondPage.Meta.Pagination == nil || secondPage.Meta.Pagination.HasMore || secondPage.Meta.Pagination.NextCursor != nil {
		t.Fatalf("expected final grant page metadata, got %+v", secondPage)
	}
	assertInventoryAccessGrant(t, secondPage.Data[0], tenantID, inventoryID, "viewer-user", "viewer")

	allGrants := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants?limit=50", "Bearer dev:inventory-owner", nil)
	if allGrants.Code != http.StatusOK {
		t.Fatalf("expected all grant list status %d, got %d with body %s", http.StatusOK, allGrants.Code, allGrants.Body.String())
	}
	allGrantBody := decodeInventoryAccessGrantList(t, allGrants)
	if len(allGrantBody.Data) != 2 {
		t.Fatalf("expected duplicate grant not to add a third grant, got %+v", allGrantBody.Data)
	}

	badCursor := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants?cursor=%25%25%25", "Bearer dev:inventory-owner", nil)
	if badCursor.Code != http.StatusBadRequest {
		t.Fatalf("expected bad cursor status %d, got %d with body %s", http.StatusBadRequest, badCursor.Code, badCursor.Body.String())
	}
	assertSafeError(t, badCursor, "invalid_request", "Invalid request.")

	wrongScopeCursor := paginationCursor(map[string]any{
		"v":          1,
		"collection": "inventory_access_grants",
		"scope":      tenantID + ":" + hiddenInventoryID,
		"lastId":     "viewer-user:viewer",
	})
	wrongCursorList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants?cursor="+wrongScopeCursor, "Bearer dev:inventory-owner", nil)
	if wrongCursorList.Code != http.StatusBadRequest {
		t.Fatalf("expected wrong-scope cursor status %d, got %d with body %s", http.StatusBadRequest, wrongCursorList.Code, wrongCursorList.Body.String())
	}
	assertSafeError(t, wrongCursorList, "invalid_request", "Invalid request.")

	viewerListAssets := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:viewer-user", nil)
	if viewerListAssets.Code != http.StatusOK {
		t.Fatalf("expected viewer list assets status %d, got %d with body %s", http.StatusOK, viewerListAssets.Code, viewerListAssets.Body.String())
	}

	viewerListInventories := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories?limit=50", "Bearer dev:viewer-user", nil)
	if viewerListInventories.Code != http.StatusOK {
		t.Fatalf("expected viewer inventory list status %d, got %d with body %s", http.StatusOK, viewerListInventories.Code, viewerListInventories.Body.String())
	}
	assertInventories(t, decodeInventoryList(t, viewerListInventories),
		expectedInventory{id: inventoryID, tenantID: tenantID, name: "Tools"},
	)

	viewerCreateAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:viewer-user", map[string]string{
		"kind":  "item",
		"title": "Drill",
	})
	if viewerCreateAsset.Code != http.StatusForbidden {
		t.Fatalf("expected viewer create asset status %d, got %d with body %s", http.StatusForbidden, viewerCreateAsset.Code, viewerCreateAsset.Body.String())
	}
	assertSafeError(t, viewerCreateAsset, "forbidden", "Forbidden.")

	editorCreateAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:editor-user", map[string]string{
		"kind":  "item",
		"title": "Drill",
	})
	if editorCreateAsset.Code != http.StatusCreated {
		t.Fatalf("expected editor create asset status %d, got %d with body %s", http.StatusCreated, editorCreateAsset.Code, editorCreateAsset.Body.String())
	}

	editorListInventories := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories?limit=50", "Bearer dev:editor-user", nil)
	if editorListInventories.Code != http.StatusOK {
		t.Fatalf("expected editor inventory list status %d, got %d with body %s", http.StatusOK, editorListInventories.Code, editorListInventories.Body.String())
	}
	assertInventories(t, decodeInventoryList(t, editorListInventories),
		expectedInventory{id: inventoryID, tenantID: tenantID, name: "Tools"},
	)

	for _, item := range []struct {
		name          string
		principal     string
		expectedGrant string
	}{
		{name: "viewer", principal: "viewer-user", expectedGrant: "another-viewer"},
		{name: "editor", principal: "editor-user", expectedGrant: "another-editor"},
		{name: "unrelated", principal: "intruder", expectedGrant: "another-intruder"},
	} {
		t.Run(item.name+" cannot share", func(t *testing.T) {
			response := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:"+item.principal, map[string]string{
				"principalId":  item.expectedGrant,
				"relationship": "viewer",
			})
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected share status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
			}
			assertSafeError(t, response, "forbidden", "Forbidden.")
		})

		t.Run(item.name+" cannot list grants", func(t *testing.T) {
			response := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:"+item.principal, nil)
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected grant list status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
			}
			assertSafeError(t, response, "forbidden", "Forbidden.")
		})
	}
}

func TestTenantOwnerListsAllInventoriesInTenant(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID: tenantID, name: "Tools", owner: "other-user"},
			{id: "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID: tenantID, name: "Supplies", owner: "another-user"},
		},
	}))

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories", "Bearer dev:owner", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, list.Code, list.Body.String())
	}

	assertInventories(t, decodeInventoryList(t, list),
		expectedInventory{id: "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID: tenantID, name: "Tools"},
		expectedInventory{id: "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID: tenantID, name: "Supplies"},
	)
}

func TestInventoryOwnerListsOnlyVisibleInventories(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
		},
		inventories: []seedInventory{
			{id: "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID: tenantID, name: "Visible One", owner: "inventory-owner"},
			{id: "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID: tenantID, name: "Hidden", owner: "other-user"},
			{id: "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID: tenantID, name: "Visible Two", owner: "inventory-owner"},
		},
	}))

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories?limit=1", "Bearer dev:inventory-owner", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusOK, list.Code, list.Body.String())
	}
	firstPage := decodeInventoryListBody(t, list)
	if firstPage.Meta.Pagination == nil || firstPage.Meta.Pagination.Limit != 1 || !firstPage.Meta.Pagination.HasMore || firstPage.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected paginated first page metadata, got %+v", firstPage.Meta)
	}

	assertInventories(t, firstPage.Data,
		expectedInventory{id: "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID: tenantID, name: "Visible One"},
	)

	secondList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories?limit=1&cursor="+*firstPage.Meta.Pagination.NextCursor, "Bearer dev:inventory-owner", nil)
	if secondList.Code != http.StatusOK {
		t.Fatalf("expected second list status %d, got %d with body %s", http.StatusOK, secondList.Code, secondList.Body.String())
	}
	if !bytes.Contains(secondList.Body.Bytes(), []byte(`"nextCursor":null`)) {
		t.Fatalf("expected final inventory page to include null nextCursor, got %s", secondList.Body.String())
	}
	secondPage := decodeInventoryListBody(t, secondList)
	if secondPage.Meta.Pagination == nil || secondPage.Meta.Pagination.Limit != 1 || secondPage.Meta.Pagination.HasMore || secondPage.Meta.Pagination.NextCursor != nil {
		t.Fatalf("expected final page metadata, got %+v", secondPage.Meta)
	}
	assertInventories(t, secondPage.Data,
		expectedInventory{id: "01ARZ3NDEKTSV4RRFFQ69G5FAY", tenantID: tenantID, name: "Visible Two"},
	)
}

func TestUnrelatedUserCannotCreateOrListTenantInventories(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID: tenantID, name: "Tools", owner: "owner"},
		},
	}))

	createInventory := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", "Bearer dev:unrelated", map[string]string{"name": "Intrusion"})
	if createInventory.Code != http.StatusForbidden {
		t.Fatalf("expected create status %d, got %d with body %s", http.StatusForbidden, createInventory.Code, createInventory.Body.String())
	}
	assertSafeError(t, createInventory, "forbidden", "Forbidden.")

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories", "Bearer dev:unrelated", nil)
	if list.Code != http.StatusForbidden {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusForbidden, list.Code, list.Body.String())
	}
	assertSafeError(t, list, "forbidden", "Forbidden.")
}

func TestInventoryListRejectsInvalidCursorSafely(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID: tenantID, name: "Tools", owner: "owner"},
		},
	}))

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories?cursor=%25%25%25", "Bearer dev:owner", nil)
	if list.Code != http.StatusBadRequest {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusBadRequest, list.Code, list.Body.String())
	}
	assertSafeError(t, list, "invalid_request", "Invalid request.")
}

func TestInventoryListRejectsWrongCursorShapeSafely(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: "01ARZ3NDEKTSV4RRFFQ69G5FAW", tenantID: tenantID, name: "Tools", owner: "owner"},
		},
	}))

	cases := []struct {
		name   string
		cursor string
	}{
		{name: "wrong collection", cursor: paginationCursor(map[string]any{"v": 1, "collection": "assets", "scope": tenantID, "lastId": "01ARZ3NDEKTSV4RRFFQ69G5FAW"})},
		{name: "wrong tenant scope", cursor: paginationCursor(map[string]any{"v": 1, "collection": "inventories", "scope": "01ARZ3NDEKTSV4RRFFQ69G5FAX", "lastId": "01ARZ3NDEKTSV4RRFFQ69G5FAW"})},
		{name: "wrong version", cursor: paginationCursor(map[string]any{"v": 2, "collection": "inventories", "scope": tenantID, "lastId": "01ARZ3NDEKTSV4RRFFQ69G5FAW"})},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories?cursor="+item.cursor, "Bearer dev:owner", nil)
			if list.Code != http.StatusBadRequest {
				t.Fatalf("expected list status %d, got %d with body %s", http.StatusBadRequest, list.Code, list.Body.String())
			}
			assertSafeError(t, list, "invalid_request", "Invalid request.")
		})
	}
}

func TestStateCreatedDuringAuthorizationGrantFailureStaysProtected(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	store := memory.NewStore()
	server := NewServer(":0", app.New(app.Dependencies{
		Observer:     &fakeObserver{},
		Auth:         auth.NewLocalDevAuthenticator(),
		Authorizer:   failingGrantAuthorizer{},
		Tenants:      store,
		Inventories:  store,
		CustomFields: store,
		Assets:       store,
		Audit:        store,
		Outbox:       store,
		IDs:          &fakeIDGenerator{ids: []string{tenantID, "tenant-event"}},
	}))

	createTenant := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:owner", map[string]string{"name": "Home"})
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected create tenant status %d, got %d with body %s", http.StatusCreated, createTenant.Code, createTenant.Body.String())
	}

	createInventory := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", "Bearer dev:owner", map[string]string{"name": "Tools"})
	if createInventory.Code != http.StatusForbidden {
		t.Fatalf("expected create inventory status %d, got %d with body %s", http.StatusForbidden, createInventory.Code, createInventory.Body.String())
	}
	assertSafeError(t, createInventory, "forbidden", "Forbidden.")

	list := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories", "Bearer dev:owner", nil)
	if list.Code != http.StatusForbidden {
		t.Fatalf("expected list status %d, got %d with body %s", http.StatusForbidden, list.Code, list.Body.String())
	}
	assertSafeError(t, list, "forbidden", "Forbidden.")
}

func TestCreateInventoryForMissingTenantReturnsSafeNotFound(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{}, "unused-id"))

	response := performRequest(server, http.MethodPost, "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories", "Bearer dev:user-one", map[string]string{"name": "Tools"})
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusNotFound, response.Code, response.Body.String())
	}
	assertSafeError(t, response, "resource_not_found", "Resource not found.")
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

func newTestApp(observer ports.Observer, ids ...string) app.App {
	store := memory.NewStore()
	return app.New(app.Dependencies{
		Observer:     observer,
		Auth:         auth.NewLocalDevAuthenticator(),
		Authorizer:   memory.NewAuthorizer(),
		Tenants:      store,
		Inventories:  store,
		CustomFields: store,
		Assets:       store,
		Audit:        store,
		Outbox:       store,
		IDs:          &fakeIDGenerator{ids: ids},
	})
}

func newSeededTestApp(t *testing.T, state seededState) app.App {
	t.Helper()

	ctx := context.Background()
	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()

	for _, item := range state.tenants {
		tenantID := tenant.ID(item.id)
		name, ok := tenant.NewName(item.name)
		if !ok {
			t.Fatalf("invalid tenant name %q", item.name)
		}
		if err := store.SaveTenant(ctx, tenant.Tenant{ID: tenantID, Name: name}); err != nil {
			t.Fatalf("save tenant: %v", err)
		}
		if item.owner != "" {
			if err := authorizer.GrantTenantOwner(ctx, principal(item.owner), tenantID); err != nil {
				t.Fatalf("grant tenant owner: %v", err)
			}
		}
	}

	for _, item := range state.inventories {
		name, ok := inventory.NewName(item.name)
		if !ok {
			t.Fatalf("invalid inventory name %q", item.name)
		}
		inventoryID := inventory.InventoryID(item.id)
		tenantID := tenant.ID(item.tenantID)
		if err := store.SaveInventory(ctx, inventory.Inventory{
			ID:       inventoryID,
			TenantID: inventory.TenantID(tenantID.String()),
			Name:     name,
		}); err != nil {
			t.Fatalf("save inventory: %v", err)
		}
		if item.owner != "" {
			if err := authorizer.GrantInventoryOwner(ctx, principal(item.owner), tenantID, inventoryID); err != nil {
				t.Fatalf("grant inventory owner: %v", err)
			}
		}
	}

	return app.New(app.Dependencies{
		Observer:     &fakeObserver{},
		Auth:         auth.NewLocalDevAuthenticator(),
		Authorizer:   authorizer,
		Tenants:      store,
		Inventories:  store,
		CustomFields: store,
		Assets:       store,
		Audit:        store,
		Outbox:       store,
		IDs:          &fakeIDGenerator{ids: state.ids},
	})
}

func principal(id string) identity.Principal {
	return identity.Principal{ID: identity.PrincipalID(id)}
}

type failingGrantAuthorizer struct{}

func (failingGrantAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	return ports.ErrForbidden
}

func (failingGrantAuthorizer) CheckInventory(context.Context, identity.Principal, ports.InventoryPermission, inventory.InventoryID) error {
	return ports.ErrForbidden
}

func (failingGrantAuthorizer) GrantTenantOwner(context.Context, identity.Principal, tenant.ID) error {
	return errors.New("spicedb unavailable")
}

func (failingGrantAuthorizer) GrantInventoryOwner(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return errors.New("spicedb unavailable")
}

func (failingGrantAuthorizer) GrantInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return errors.New("spicedb unavailable")
}

func (failingGrantAuthorizer) GrantInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return errors.New("spicedb unavailable")
}

func performRequest(server *http.Server, method string, path string, authorization string, body any) *httptest.ResponseRecorder {
	return performRequestWithHeaders(server, method, path, authorization, nil, body)
}

func performRequestWithHeaders(server *http.Server, method string, path string, authorization string, headers map[string]string, body any) *httptest.ResponseRecorder {
	var requestBody []byte
	if body != nil {
		requestBody, _ = json.Marshal(body)
	}

	request := httptest.NewRequest(method, path, bytes.NewReader(requestBody))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)
	return response
}

func decodeBody(t *testing.T, response *httptest.ResponseRecorder, body any) {
	t.Helper()

	if err := json.NewDecoder(response.Body).Decode(body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}

type errorResponse struct {
	Error struct {
		Code    string        `json:"code"`
		Message string        `json:"message"`
		Details []interface{} `json:"details"`
	} `json:"error"`
	Meta responseMeta `json:"meta"`
}

type inventoryListResponse struct {
	Data []struct {
		ID       string `json:"id"`
		TenantID string `json:"tenantId"`
		Name     string `json:"name"`
	} `json:"data"`
	Meta responseMeta `json:"meta"`
}

type assetBody struct {
	Data assetResponse `json:"data"`
	Meta responseMeta  `json:"meta"`
}

type assetListBody struct {
	Data []assetResponse `json:"data"`
	Meta responseMeta    `json:"meta"`
}

type customFieldDefinitionBody struct {
	Data customFieldDefinitionResponse `json:"data"`
	Meta responseMeta                  `json:"meta"`
}

type customFieldDefinitionListBody struct {
	Data []customFieldDefinitionResponse `json:"data"`
	Meta responseMeta                    `json:"meta"`
}

type inventoryAccessGrantBody struct {
	Data inventoryAccessGrantResponse `json:"data"`
	Meta responseMeta                 `json:"meta"`
}

type inventoryAccessGrantListBody struct {
	Data []inventoryAccessGrantResponse `json:"data"`
	Meta responseMeta                   `json:"meta"`
}

type auditRecordListBody struct {
	Data []auditRecordResponse `json:"data"`
	Meta responseMeta          `json:"meta"`
}

type seededState struct {
	tenants     []seedTenant
	inventories []seedInventory
	ids         []string
}

type seedTenant struct {
	id    string
	name  string
	owner string
}

type seedInventory struct {
	id       string
	tenantID string
	name     string
	owner    string
}

type expectedInventory struct {
	id       string
	tenantID string
	name     string
}

func decodeAsset(t *testing.T, response *httptest.ResponseRecorder) assetBody {
	t.Helper()

	var body assetBody
	decodeBody(t, response, &body)
	return body
}

func decodeAssetList(t *testing.T, response *httptest.ResponseRecorder) assetListBody {
	t.Helper()

	var body assetListBody
	decodeBody(t, response, &body)
	return body
}

func decodeAuditRecordList(t *testing.T, response *httptest.ResponseRecorder) auditRecordListBody {
	t.Helper()

	var body auditRecordListBody
	decodeBody(t, response, &body)
	return body
}

func auditRecordsContainTarget(records []auditRecordResponse, targetID string) bool {
	for _, record := range records {
		if record.TargetID == targetID {
			return true
		}
	}
	return false
}

func decodeInventoryList(t *testing.T, response *httptest.ResponseRecorder) []struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	Name     string `json:"name"`
} {
	t.Helper()

	var body inventoryListResponse
	decodeBody(t, response, &body)
	return body.Data
}

func decodeInventoryListBody(t *testing.T, response *httptest.ResponseRecorder) inventoryListResponse {
	t.Helper()

	var body inventoryListResponse
	decodeBody(t, response, &body)
	return body
}

func decodeCustomFieldDefinition(t *testing.T, response *httptest.ResponseRecorder) customFieldDefinitionBody {
	t.Helper()

	var body customFieldDefinitionBody
	decodeBody(t, response, &body)
	return body
}

func decodeCustomFieldDefinitionList(t *testing.T, response *httptest.ResponseRecorder) customFieldDefinitionListBody {
	t.Helper()

	var body customFieldDefinitionListBody
	decodeBody(t, response, &body)
	return body
}

func decodeInventoryAccessGrant(t *testing.T, response *httptest.ResponseRecorder) inventoryAccessGrantBody {
	t.Helper()

	var body inventoryAccessGrantBody
	decodeBody(t, response, &body)
	return body
}

func decodeInventoryAccessGrantList(t *testing.T, response *httptest.ResponseRecorder) inventoryAccessGrantListBody {
	t.Helper()

	var body inventoryAccessGrantListBody
	decodeBody(t, response, &body)
	return body
}

func assertInventoryAccessGrant(t *testing.T, grant inventoryAccessGrantResponse, tenantID string, inventoryID string, principalID string, relationship string) {
	t.Helper()

	if grant.TenantID != tenantID || grant.InventoryID != inventoryID || grant.PrincipalID != principalID || grant.Relationship != relationship {
		t.Fatalf("expected access grant %s/%s/%s/%s, got %+v", tenantID, inventoryID, principalID, relationship, grant)
	}
}

func assertCustomFieldDefinition(t *testing.T, definition customFieldDefinitionResponse, tenantID string, inventoryID string, scope string, key string, fieldType string) {
	t.Helper()

	if definition.TenantID != tenantID || definition.InventoryID != inventoryID || definition.Scope != scope || definition.Key != key || definition.Type != fieldType {
		t.Fatalf("expected custom field definition %s/%s/%s/%s/%s, got %+v", tenantID, inventoryID, scope, key, fieldType, definition)
	}
}

func assertInventories(t *testing.T, inventories []struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	Name     string `json:"name"`
}, expected ...expectedInventory) {
	t.Helper()

	if len(inventories) != len(expected) {
		t.Fatalf("expected inventories %v, got %+v", expected, inventories)
	}

	seen := map[string]struct {
		tenantID string
		name     string
	}{}
	for _, item := range inventories {
		seen[item.ID] = struct {
			tenantID string
			name     string
		}{tenantID: item.TenantID, name: item.Name}
	}
	for _, item := range expected {
		actual, ok := seen[item.id]
		if !ok {
			t.Fatalf("expected inventory ID %q in %+v", item.id, inventories)
		}
		if actual.tenantID != item.tenantID || actual.name != item.name {
			t.Fatalf("expected inventory %q tenant/name %q/%q, got %q/%q", item.id, item.tenantID, item.name, actual.tenantID, actual.name)
		}
	}
}

func assertSafeError(t *testing.T, response *httptest.ResponseRecorder, expectedCode string, expectedMessage string) {
	t.Helper()

	var body errorResponse
	decodeBody(t, response, &body)
	if body.Error.Code != expectedCode {
		t.Fatalf("expected error code %q, got %q", expectedCode, body.Error.Code)
	}
	if body.Error.Message != expectedMessage {
		t.Fatalf("expected error message %q, got %q", expectedMessage, body.Error.Message)
	}
	if len(body.Error.Details) != 0 {
		t.Fatalf("expected no error details, got %+v", body.Error.Details)
	}
	if body.Meta.TenantID != "" || body.Meta.RequestID != "" {
		t.Fatalf("expected empty error metadata, got %+v", body.Meta)
	}
}

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, expectedCode string) {
	t.Helper()

	var body errorResponse
	decodeBody(t, response, &body)
	if body.Error.Code != expectedCode {
		t.Fatalf("expected error code %q, got %q", expectedCode, body.Error.Code)
	}
}

func paginationCursor(payload map[string]any) string {
	data, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

type fakeIDGenerator struct {
	ids []string
}

func (f *fakeIDGenerator) NewID() string {
	if len(f.ids) == 0 {
		return "fixed-id"
	}
	id := f.ids[0]
	f.ids = f.ids[1:]
	return id
}

type fakeObserver struct {
	events []ports.Event
}

func (f *fakeObserver) Record(_ context.Context, event ports.Event) {
	f.events = append(f.events, event)
}
