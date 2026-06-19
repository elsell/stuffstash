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
		{name: "create asset", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/assets", body: map[string]string{"kind": "item", "title": "Drill"}},
		{name: "list assets", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/assets"},
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

	server := NewServer(":0", newTestApp(&fakeObserver{}, tenantID, "tenant-event", "tenant-claim", inventoryID, "inventory-event", "inventory-claim", "location-one", "asset-one"))

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
		Observer:    &fakeObserver{},
		Auth:        auth.NewLocalDevAuthenticator(),
		Authorizer:  failingGrantAuthorizer{},
		Tenants:     store,
		Inventories: store,
		Assets:      store,
		Outbox:      store,
		IDs:         &fakeIDGenerator{ids: []string{tenantID, "tenant-event"}},
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
		Observer:    observer,
		Auth:        auth.NewLocalDevAuthenticator(),
		Authorizer:  memory.NewAuthorizer(),
		Tenants:     store,
		Inventories: store,
		Assets:      store,
		Outbox:      store,
		IDs:         &fakeIDGenerator{ids: ids},
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
		Observer:    &fakeObserver{},
		Auth:        auth.NewLocalDevAuthenticator(),
		Authorizer:  authorizer,
		Tenants:     store,
		Inventories: store,
		Assets:      store,
		Outbox:      store,
		IDs:         &fakeIDGenerator{},
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

func performRequest(server *http.Server, method string, path string, authorization string, body any) *httptest.ResponseRecorder {
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

type seededState struct {
	tenants     []seedTenant
	inventories []seedInventory
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
