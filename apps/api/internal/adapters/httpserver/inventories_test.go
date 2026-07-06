package httpserver

import (
	"bytes"
	"net/http"
	"testing"
)

func TestInventoryEndpointsDenyCrossUserAccess(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newTestApp(&fakeObserver{}, tenantID, "audit-tenant", "tenant-event", "tenant-claim"))

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

func TestInventoryResponsesIncludeEffectiveAccessMetadata(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "inventory-owner"},
		},
	}))

	grantViewer := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "viewer",
		"relationship": "viewer",
	})
	if grantViewer.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, grantViewer.Code, grantViewer.Body.String())
	}
	grantEditor := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "editor",
		"relationship": "editor",
	})
	if grantEditor.Code != http.StatusCreated {
		t.Fatalf("expected editor grant status %d, got %d with body %s", http.StatusCreated, grantEditor.Code, grantEditor.Body.String())
	}

	cases := []struct {
		name                 string
		authorization        string
		expectedRelationship string
		expectedPermission   string
		expectedPermissions  []string
		forbiddenPermission  string
		forbiddenPermissions []string
	}{
		{name: "tenant owner", authorization: "Bearer dev:tenant-owner", expectedRelationship: "owner", expectedPermission: "share", expectedPermissions: []string{"view_import_job", "create_import_job"}},
		{name: "inventory owner", authorization: "Bearer dev:inventory-owner", expectedRelationship: "owner", expectedPermission: "configure", expectedPermissions: []string{"view_import_job", "create_import_job"}},
		{name: "editor", authorization: "Bearer dev:editor", expectedRelationship: "editor", expectedPermission: "edit_asset", expectedPermissions: []string{"view_import_job", "create_import_job"}, forbiddenPermission: "share"},
		{name: "viewer", authorization: "Bearer dev:viewer", expectedRelationship: "viewer", expectedPermission: "view", forbiddenPermission: "edit_asset", forbiddenPermissions: []string{"view_import_job", "create_import_job"}},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			response := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID, item.authorization, nil)
			if response.Code != http.StatusOK {
				t.Fatalf("expected inventory detail status %d, got %d with body %s", http.StatusOK, response.Code, response.Body.String())
			}
			var body inventoryBody
			decodeBody(t, response, &body)
			if body.Data.Access.Relationship != item.expectedRelationship {
				t.Fatalf("expected relationship %q, got %+v", item.expectedRelationship, body.Data.Access)
			}
			if !accessContainsPermission(body.Data.Access.Permissions, item.expectedPermission) {
				t.Fatalf("expected permission %q in %+v", item.expectedPermission, body.Data.Access)
			}
			for _, permission := range item.expectedPermissions {
				if !accessContainsPermission(body.Data.Access.Permissions, permission) {
					t.Fatalf("expected permission %q in %+v", permission, body.Data.Access)
				}
			}
			if item.forbiddenPermission != "" && accessContainsPermission(body.Data.Access.Permissions, item.forbiddenPermission) {
				t.Fatalf("did not expect permission %q in %+v", item.forbiddenPermission, body.Data.Access)
			}
			for _, permission := range item.forbiddenPermissions {
				if accessContainsPermission(body.Data.Access.Permissions, permission) {
					t.Fatalf("did not expect permission %q in %+v", permission, body.Data.Access)
				}
			}
		})
	}
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

func TestCreateInventoryForMissingTenantReturnsSafeNotFound(t *testing.T) {
	server := NewServer(":0", newTestApp(&fakeObserver{}, "unused-id"))

	response := performRequest(server, http.MethodPost, "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories", "Bearer dev:user-one", map[string]string{"name": "Tools"})
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusNotFound, response.Code, response.Body.String())
	}
	assertSafeError(t, response, "resource_not_found", "Resource not found.")
}
