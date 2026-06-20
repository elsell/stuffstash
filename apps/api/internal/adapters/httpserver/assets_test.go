package httpserver

import (
	"net/http"
	"testing"
)

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
		ids: []string{"garage-location", "audit-garage-location", "cross-inventory-asset"},
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

func TestAssetLifecycleArchiveRestoreFlowAndListing(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{
			"garage", "audit-garage",
			"wrench", "audit-wrench",
			"audit-wrench-archive",
			"audit-garage-archive",
			"audit-garage-restore",
			"audit-wrench-restore",
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
	wrenchResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":          "item",
		"title":         "Wrench",
		"parentAssetId": garage.Data.ID,
	})
	if wrenchResponse.Code != http.StatusCreated {
		t.Fatalf("expected wrench create status %d, got %d with body %s", http.StatusCreated, wrenchResponse.Code, wrenchResponse.Body.String())
	}
	wrench := decodeAsset(t, wrenchResponse)

	archiveWithChild := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+garage.Data.ID+"/archive", "Bearer dev:owner", nil)
	if archiveWithChild.Code != http.StatusBadRequest {
		t.Fatalf("expected archive with child status %d, got %d with body %s", http.StatusBadRequest, archiveWithChild.Code, archiveWithChild.Body.String())
	}
	assertSafeError(t, archiveWithChild, "invalid_request", "Invalid request.")

	archiveChild := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+wrench.Data.ID+"/archive", "Bearer dev:owner", nil)
	if archiveChild.Code != http.StatusOK {
		t.Fatalf("expected child archive status %d, got %d with body %s", http.StatusOK, archiveChild.Code, archiveChild.Body.String())
	}
	archivedChild := decodeAsset(t, archiveChild)
	if archivedChild.Data.LifecycleState != "archived" || archivedChild.Data.ParentAssetID != garage.Data.ID {
		t.Fatalf("expected archived child with same parent, got %+v", archivedChild.Data)
	}
	updateArchivedChild := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+wrench.Data.ID, "Bearer dev:owner", map[string]any{"title": "Renamed Wrench"})
	if updateArchivedChild.Code != http.StatusBadRequest {
		t.Fatalf("expected archived child update status %d, got %d with body %s", http.StatusBadRequest, updateArchivedChild.Code, updateArchivedChild.Body.String())
	}
	assertSafeError(t, updateArchivedChild, "invalid_request", "Invalid request.")

	defaultList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?limit=50", "Bearer dev:owner", nil)
	if defaultList.Code != http.StatusOK {
		t.Fatalf("expected default list status %d, got %d with body %s", http.StatusOK, defaultList.Code, defaultList.Body.String())
	}
	defaultListBody := decodeAssetList(t, defaultList)
	if assetListContainsID(defaultListBody.Data, wrench.Data.ID) || !assetListContainsID(defaultListBody.Data, garage.Data.ID) {
		t.Fatalf("expected default list to include active garage and exclude archived wrench, got %+v", defaultListBody.Data)
	}

	archivedList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?lifecycleState=archived&limit=50", "Bearer dev:owner", nil)
	if archivedList.Code != http.StatusOK {
		t.Fatalf("expected archived list status %d, got %d with body %s", http.StatusOK, archivedList.Code, archivedList.Body.String())
	}
	if !assetListContainsID(decodeAssetList(t, archivedList).Data, wrench.Data.ID) {
		t.Fatalf("expected archived list to include wrench, got %s", archivedList.Body.String())
	}

	archiveParent := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+garage.Data.ID+"/archive", "Bearer dev:owner", nil)
	if archiveParent.Code != http.StatusOK {
		t.Fatalf("expected parent archive status %d, got %d with body %s", http.StatusOK, archiveParent.Code, archiveParent.Body.String())
	}

	restoreChildWithArchivedParent := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+wrench.Data.ID+"/restore", "Bearer dev:owner", nil)
	if restoreChildWithArchivedParent.Code != http.StatusBadRequest {
		t.Fatalf("expected child restore with archived parent status %d, got %d with body %s", http.StatusBadRequest, restoreChildWithArchivedParent.Code, restoreChildWithArchivedParent.Body.String())
	}
	assertSafeError(t, restoreChildWithArchivedParent, "invalid_request", "Invalid request.")

	restoreParent := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+garage.Data.ID+"/restore", "Bearer dev:owner", nil)
	if restoreParent.Code != http.StatusOK {
		t.Fatalf("expected parent restore status %d, got %d with body %s", http.StatusOK, restoreParent.Code, restoreParent.Body.String())
	}
	if decodeAsset(t, restoreParent).Data.LifecycleState != "active" {
		t.Fatalf("expected restored parent active, got %s", restoreParent.Body.String())
	}
	restoreChild := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+wrench.Data.ID+"/restore", "Bearer dev:owner", nil)
	if restoreChild.Code != http.StatusOK {
		t.Fatalf("expected child restore status %d, got %d with body %s", http.StatusOK, restoreChild.Code, restoreChild.Body.String())
	}
	if decodeAsset(t, restoreChild).Data.LifecycleState != "active" {
		t.Fatalf("expected restored child active, got %s", restoreChild.Body.String())
	}

	auditResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records?limit=50", "Bearer dev:owner", nil)
	if auditResponse.Code != http.StatusOK {
		t.Fatalf("expected audit status %d, got %d with body %s", http.StatusOK, auditResponse.Code, auditResponse.Body.String())
	}
	auditRecords := decodeAuditRecordList(t, auditResponse).Data
	if !auditRecordsContainAction(auditRecords, "asset.archived") || !auditRecordsContainAction(auditRecords, "asset.restored") {
		t.Fatalf("expected archive and restore audit actions, got %+v", auditRecords)
	}
}

func TestAssetLifecycleAuthorizationBoundaries(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	const otherInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
			{id: otherTenantID, name: "Cabin", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
			{id: otherInventoryID, tenantID: otherTenantID, name: "Cabin Tools", owner: "other-owner"},
		},
		ids: []string{
			"wrench", "audit-wrench",
			"viewer-grant-event", "audit-viewer-grant", "viewer-grant-claim",
		},
	}))

	wrenchResponse := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
		"kind":  "item",
		"title": "Wrench",
	})
	if wrenchResponse.Code != http.StatusCreated {
		t.Fatalf("expected wrench create status %d, got %d with body %s", http.StatusCreated, wrenchResponse.Code, wrenchResponse.Body.String())
	}
	wrench := decodeAsset(t, wrenchResponse)

	viewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:owner", map[string]string{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if viewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, viewerGrant.Code, viewerGrant.Body.String())
	}
	viewerArchive := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+wrench.Data.ID+"/archive", "Bearer dev:viewer-user", nil)
	if viewerArchive.Code != http.StatusForbidden {
		t.Fatalf("expected viewer archive status %d, got %d with body %s", http.StatusForbidden, viewerArchive.Code, viewerArchive.Body.String())
	}
	assertSafeError(t, viewerArchive, "forbidden", "Forbidden.")

	intruderRestore := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+wrench.Data.ID+"/restore", "Bearer dev:intruder", nil)
	if intruderRestore.Code != http.StatusForbidden {
		t.Fatalf("expected intruder restore status %d, got %d with body %s", http.StatusForbidden, intruderRestore.Code, intruderRestore.Body.String())
	}
	assertSafeError(t, intruderRestore, "forbidden", "Forbidden.")

	crossTenantArchive := performRequest(server, http.MethodPatch, "/tenants/"+otherTenantID+"/inventories/"+otherInventoryID+"/assets/"+wrench.Data.ID+"/archive", "Bearer dev:other-owner", nil)
	if crossTenantArchive.Code != http.StatusNotFound {
		t.Fatalf("expected cross-tenant archive status %d, got %d with body %s", http.StatusNotFound, crossTenantArchive.Code, crossTenantArchive.Body.String())
	}
	assertSafeError(t, crossTenantArchive, "resource_not_found", "Resource not found.")

	missingAuth := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/"+wrench.Data.ID+"/archive", "", nil)
	if missingAuth.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing auth status %d, got %d with body %s", http.StatusUnauthorized, missingAuth.Code, missingAuth.Body.String())
	}
	assertSafeError(t, missingAuth, "authentication_required", "Authentication required.")
}

func TestAssetLifecycleStateAndCursorContracts(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"},
		},
		ids: []string{
			"drill", "audit-drill",
			"hammer", "audit-hammer",
			"wrench", "audit-wrench",
			"audit-wrench-archive",
		},
	}))

	for _, item := range []struct {
		id    string
		title string
	}{
		{id: "drill", title: "Drill"},
		{id: "hammer", title: "Hammer"},
		{id: "wrench", title: "Wrench"},
	} {
		response := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{
			"kind":  "item",
			"title": item.title,
		})
		if response.Code != http.StatusCreated {
			t.Fatalf("expected %s create status %d, got %d with body %s", item.id, http.StatusCreated, response.Code, response.Body.String())
		}
	}

	restoreActive := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/wrench/restore", "Bearer dev:owner", nil)
	if restoreActive.Code != http.StatusBadRequest {
		t.Fatalf("expected active restore status %d, got %d with body %s", http.StatusBadRequest, restoreActive.Code, restoreActive.Body.String())
	}
	assertSafeError(t, restoreActive, "invalid_request", "Invalid request.")

	archiveWrench := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/wrench/archive", "Bearer dev:owner", nil)
	if archiveWrench.Code != http.StatusOK {
		t.Fatalf("expected archive status %d, got %d with body %s", http.StatusOK, archiveWrench.Code, archiveWrench.Body.String())
	}
	archiveAgain := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets/wrench/archive", "Bearer dev:owner", nil)
	if archiveAgain.Code != http.StatusBadRequest {
		t.Fatalf("expected archived archive status %d, got %d with body %s", http.StatusBadRequest, archiveAgain.Code, archiveAgain.Body.String())
	}
	assertSafeError(t, archiveAgain, "invalid_request", "Invalid request.")

	allList := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?lifecycleState=all&limit=50", "Bearer dev:owner", nil)
	if allList.Code != http.StatusOK {
		t.Fatalf("expected all lifecycle list status %d, got %d with body %s", http.StatusOK, allList.Code, allList.Body.String())
	}
	allListBody := decodeAssetList(t, allList)
	if !assetListContainsID(allListBody.Data, "drill") || !assetListContainsID(allListBody.Data, "hammer") || !assetListContainsID(allListBody.Data, "wrench") {
		t.Fatalf("expected all lifecycle list to include active and archived assets, got %+v", allListBody.Data)
	}

	activePage := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?lifecycleState=active&limit=1", "Bearer dev:owner", nil)
	if activePage.Code != http.StatusOK {
		t.Fatalf("expected active page status %d, got %d with body %s", http.StatusOK, activePage.Code, activePage.Body.String())
	}
	activePageBody := decodeAssetList(t, activePage)
	if activePageBody.Meta.Pagination == nil || activePageBody.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected active page to include next cursor, got %+v", activePageBody.Meta)
	}
	wrongScopeCursor := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?lifecycleState=archived&cursor="+*activePageBody.Meta.Pagination.NextCursor, "Bearer dev:owner", nil)
	if wrongScopeCursor.Code != http.StatusBadRequest {
		t.Fatalf("expected wrong-scope cursor status %d, got %d with body %s", http.StatusBadRequest, wrongScopeCursor.Code, wrongScopeCursor.Body.String())
	}
	assertSafeError(t, wrongScopeCursor, "invalid_request", "Invalid request.")

	badLifecycleFilter := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?lifecycleState=deleted", "Bearer dev:owner", nil)
	if badLifecycleFilter.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected bad lifecycle filter status %d, got %d with body %s", http.StatusUnprocessableEntity, badLifecycleFilter.Code, badLifecycleFilter.Body.String())
	}
	assertErrorCode(t, badLifecycleFilter, "invalid_request")
}
