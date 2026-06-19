package httpserver

import (
	"net/http"
	"testing"
)

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
			"audit-duplicate-viewer-grant", "duplicate-viewer-grant-event", "duplicate-viewer-grant-claim",
			"audit-editor-grant", "editor-grant-event", "editor-grant-claim",
			"editor-created-asset", "audit-editor-created-asset",
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

func TestStateCreatedDuringAuthorizationGrantFailureStaysProtected(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newTestAppWithAuthorizer(&fakeObserver{}, failingGrantAuthorizer{}, tenantID, "tenant-event", "tenant-claim", "tenant-claim-attempt"))

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
