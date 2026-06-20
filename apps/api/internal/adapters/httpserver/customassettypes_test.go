package httpserver

import (
	"net/http"
	"testing"
)

func TestCustomAssetTypeFlowAndTargetedFieldValidation(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const hiddenInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Medicine", owner: "inventory-owner"},
			{id: hiddenInventoryID, tenantID: tenantID, name: "Hidden", owner: "hidden-owner"},
		},
		ids: []string{
			"01ARZ3NDEKTSV4RRFFQ69G5FAY", "audit-medicine-type",
			"01ARZ3NDEKTSV4RRFFQ69G5FAZ", "audit-supply-type",
			"01ARZ3NDEKTSV4RRFFQ69G5FB0", "audit-expiration-field",
			"01ARZ3NDEKTSV4RRFFQ69G5FB1", "audit-untyped-asset",
			"01ARZ3NDEKTSV4RRFFQ69G5FB2", "audit-typed-asset",
		},
	}))

	createType := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types", "Bearer dev:inventory-owner", map[string]any{
		"key":         "medicine",
		"displayName": "Medicine",
		"description": "Items with medication-specific fields.",
	})
	if createType.Code != http.StatusCreated {
		t.Fatalf("expected custom asset type status %d, got %d with body %s", http.StatusCreated, createType.Code, createType.Body.String())
	}
	medicineType := decodeCustomAssetType(t, createType)
	if medicineType.Data.TenantID != tenantID || medicineType.Data.InventoryID != inventoryID || medicineType.Data.Scope != "inventory" || medicineType.Data.Key != "medicine" {
		t.Fatalf("expected inventory custom asset type, got %+v", medicineType.Data)
	}
	createSecondType := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types", "Bearer dev:inventory-owner", map[string]any{
		"key":         "supply",
		"displayName": "Supply",
	})
	if createSecondType.Code != http.StatusCreated {
		t.Fatalf("expected second custom asset type status %d, got %d with body %s", http.StatusCreated, createSecondType.Code, createSecondType.Body.String())
	}

	listTypes := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types?limit=10", "Bearer dev:inventory-owner", nil)
	if listTypes.Code != http.StatusOK {
		t.Fatalf("expected custom asset type list status %d, got %d with body %s", http.StatusOK, listTypes.Code, listTypes.Body.String())
	}
	typeList := decodeCustomAssetTypeList(t, listTypes)
	if len(typeList.Data) != 2 || typeList.Data[0].ID != medicineType.Data.ID {
		t.Fatalf("expected listed medicine type, got %+v", typeList)
	}
	firstPageResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types?limit=1", "Bearer dev:inventory-owner", nil)
	if firstPageResponse.Code != http.StatusOK {
		t.Fatalf("expected first custom asset type page status %d, got %d with body %s", http.StatusOK, firstPageResponse.Code, firstPageResponse.Body.String())
	}
	firstPage := decodeCustomAssetTypeList(t, firstPageResponse)
	if len(firstPage.Data) != 1 || firstPage.Meta.Pagination == nil || !firstPage.Meta.Pagination.HasMore || firstPage.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected first paginated custom asset type page, got %+v", firstPage)
	}
	secondPageResponse := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types?limit=1&cursor="+*firstPage.Meta.Pagination.NextCursor, "Bearer dev:inventory-owner", nil)
	if secondPageResponse.Code != http.StatusOK {
		t.Fatalf("expected second custom asset type page status %d, got %d with body %s", http.StatusOK, secondPageResponse.Code, secondPageResponse.Body.String())
	}
	secondPage := decodeCustomAssetTypeList(t, secondPageResponse)
	if len(secondPage.Data) != 1 || secondPage.Meta.Pagination == nil || secondPage.Meta.Pagination.HasMore || secondPage.Meta.Pagination.NextCursor != nil {
		t.Fatalf("expected final paginated custom asset type page, got %+v", secondPage)
	}

	wrongScopeCursor := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/custom-asset-types?cursor="+*firstPage.Meta.Pagination.NextCursor, "Bearer dev:tenant-owner", nil)
	if wrongScopeCursor.Code != http.StatusBadRequest {
		t.Fatalf("expected wrong-scope cursor status %d, got %d with body %s", http.StatusBadRequest, wrongScopeCursor.Code, wrongScopeCursor.Body.String())
	}
	assertSafeError(t, wrongScopeCursor, "invalid_request", "Invalid request.")

	createField := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions", "Bearer dev:inventory-owner", map[string]any{
		"key":                "expires-on",
		"displayName":        "Expires On",
		"type":               "date",
		"applicability":      "custom_asset_types",
		"customAssetTypeIds": []string{medicineType.Data.ID},
	})
	if createField.Code != http.StatusCreated {
		t.Fatalf("expected targeted field status %d, got %d with body %s", http.StatusCreated, createField.Code, createField.Body.String())
	}
	field := decodeCustomFieldDefinition(t, createField)
	if field.Data.Applicability != "custom_asset_types" || len(field.Data.CustomAssetTypeIDs) != 1 || field.Data.CustomAssetTypeIDs[0] != medicineType.Data.ID {
		t.Fatalf("expected field to target medicine type, got %+v", field.Data)
	}

	untypedAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:inventory-owner", map[string]any{
		"kind":  "item",
		"title": "Aspirin",
		"customFields": map[string]any{
			"expires-on": "2027-01-01",
		},
	})
	if untypedAsset.Code != http.StatusBadRequest {
		t.Fatalf("expected untyped asset status %d, got %d with body %s", http.StatusBadRequest, untypedAsset.Code, untypedAsset.Body.String())
	}
	assertSafeError(t, untypedAsset, "invalid_request", "Invalid request.")

	typedAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:inventory-owner", map[string]any{
		"kind":              "item",
		"title":             "Aspirin",
		"customAssetTypeId": medicineType.Data.ID,
		"customFields": map[string]any{
			"expires-on": "2027-01-01",
		},
	})
	if typedAsset.Code != http.StatusCreated {
		t.Fatalf("expected typed asset status %d, got %d with body %s", http.StatusCreated, typedAsset.Code, typedAsset.Body.String())
	}
	assetBody := decodeAsset(t, typedAsset)
	if assetBody.Data.CustomAssetTypeID != medicineType.Data.ID || assetBody.Data.CustomFields["expires-on"] != "2027-01-01" {
		t.Fatalf("expected typed asset with expiration field, got %+v", assetBody.Data)
	}
}

func TestCustomAssetTypeEndpointsRejectUnauthorizedUsers(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "inventory-owner"},
		},
		ids: []string{"01ARZ3NDEKTSV4RRFFQ69G5FAY", "audit-type", "audit-viewer-grant", "viewer-grant-event", "viewer-grant-claim"},
	}))

	createInventoryType := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types", "Bearer dev:intruder", map[string]any{
		"key":         "medicine",
		"displayName": "Medicine",
	})
	if createInventoryType.Code != http.StatusForbidden {
		t.Fatalf("expected create inventory type status %d, got %d with body %s", http.StatusForbidden, createInventoryType.Code, createInventoryType.Body.String())
	}
	assertSafeError(t, createInventoryType, "forbidden", "Forbidden.")

	listInventoryTypes := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types", "Bearer dev:intruder", nil)
	if listInventoryTypes.Code != http.StatusForbidden {
		t.Fatalf("expected list inventory type status %d, got %d with body %s", http.StatusForbidden, listInventoryTypes.Code, listInventoryTypes.Body.String())
	}
	assertSafeError(t, listInventoryTypes, "forbidden", "Forbidden.")

	createTenantType := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/custom-asset-types", "Bearer dev:inventory-owner", map[string]any{
		"key":         "medicine",
		"displayName": "Medicine",
	})
	if createTenantType.Code != http.StatusForbidden {
		t.Fatalf("expected create tenant type status %d, got %d with body %s", http.StatusForbidden, createTenantType.Code, createTenantType.Body.String())
	}
	assertSafeError(t, createTenantType, "forbidden", "Forbidden.")

	listTenantTypes := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/custom-asset-types", "Bearer dev:inventory-owner", nil)
	if listTenantTypes.Code != http.StatusForbidden {
		t.Fatalf("expected list tenant type status %d, got %d with body %s", http.StatusForbidden, listTenantTypes.Code, listTenantTypes.Body.String())
	}
	assertSafeError(t, listTenantTypes, "forbidden", "Forbidden.")

	grantViewer := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]any{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if grantViewer.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, grantViewer.Code, grantViewer.Body.String())
	}
	viewerCreateType := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types", "Bearer dev:viewer-user", map[string]any{
		"key":         "viewer-type",
		"displayName": "Viewer Type",
	})
	if viewerCreateType.Code != http.StatusForbidden {
		t.Fatalf("expected viewer create type status %d, got %d with body %s", http.StatusForbidden, viewerCreateType.Code, viewerCreateType.Body.String())
	}
	assertSafeError(t, viewerCreateType, "forbidden", "Forbidden.")

	viewerListTypes := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types", "Bearer dev:viewer-user", nil)
	if viewerListTypes.Code != http.StatusOK {
		t.Fatalf("expected viewer list type status %d, got %d with body %s", http.StatusOK, viewerListTypes.Code, viewerListTypes.Body.String())
	}
}

func TestCustomAssetTypeUpdateFlowAndAuthorization(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const hiddenInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAZ"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FB0"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
			{id: otherTenantID, name: "Cabin", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Medicine", owner: "inventory-owner"},
			{id: hiddenInventoryID, tenantID: tenantID, name: "Hidden", owner: "hidden-owner"},
		},
		ids: []string{
			"01ARZ3NDEKTSV4RRFFQ69G5FAX", "audit-tenant-type",
			"audit-tenant-type-update",
			"01ARZ3NDEKTSV4RRFFQ69G5FAY", "audit-inventory-type",
			"audit-inventory-type-update",
			"audit-viewer-grant", "viewer-grant-event", "viewer-grant-claim",
		},
	}))

	createTenantType := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/custom-asset-types", "Bearer dev:tenant-owner", map[string]any{
		"key":         "medicine",
		"displayName": "Medicine",
		"description": "Old description",
	})
	if createTenantType.Code != http.StatusCreated {
		t.Fatalf("expected tenant type create status %d, got %d with body %s", http.StatusCreated, createTenantType.Code, createTenantType.Body.String())
	}
	tenantType := decodeCustomAssetType(t, createTenantType)
	updateTenantType := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/custom-asset-types/"+tenantType.Data.ID, "Bearer dev:tenant-owner", map[string]any{
		"displayName": "Medicine and Vitamins",
		"description": "Medication and supplement supplies",
	})
	if updateTenantType.Code != http.StatusOK {
		t.Fatalf("expected tenant type update status %d, got %d with body %s", http.StatusOK, updateTenantType.Code, updateTenantType.Body.String())
	}
	updatedTenantType := decodeCustomAssetType(t, updateTenantType)
	if updatedTenantType.Data.ID != tenantType.Data.ID || updatedTenantType.Data.Key != "medicine" || updatedTenantType.Data.DisplayName != "Medicine and Vitamins" || updatedTenantType.Data.Description != "Medication and supplement supplies" {
		t.Fatalf("expected updated tenant type metadata, got %+v", updatedTenantType.Data)
	}

	crossTenantTypeUpdate := performRequest(server, http.MethodPatch, "/tenants/"+otherTenantID+"/custom-asset-types/"+tenantType.Data.ID, "Bearer dev:other-owner", map[string]any{
		"displayName": "Cabin Medicine",
	})
	if crossTenantTypeUpdate.Code != http.StatusNotFound {
		t.Fatalf("expected cross-tenant type update status %d, got %d with body %s", http.StatusNotFound, crossTenantTypeUpdate.Code, crossTenantTypeUpdate.Body.String())
	}
	assertSafeError(t, crossTenantTypeUpdate, "resource_not_found", "Resource not found.")

	createInventoryType := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types", "Bearer dev:inventory-owner", map[string]any{
		"key":         "supplement",
		"displayName": "Supplement",
	})
	if createInventoryType.Code != http.StatusCreated {
		t.Fatalf("expected inventory type create status %d, got %d with body %s", http.StatusCreated, createInventoryType.Code, createInventoryType.Body.String())
	}
	inventoryType := decodeCustomAssetType(t, createInventoryType)
	updateInventoryType := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types/"+inventoryType.Data.ID, "Bearer dev:inventory-owner", map[string]any{
		"displayName": "Supplements",
		"description": "",
	})
	if updateInventoryType.Code != http.StatusOK {
		t.Fatalf("expected inventory type update status %d, got %d with body %s", http.StatusOK, updateInventoryType.Code, updateInventoryType.Body.String())
	}
	updatedInventoryType := decodeCustomAssetType(t, updateInventoryType)
	if updatedInventoryType.Data.ID != inventoryType.Data.ID || updatedInventoryType.Data.Key != "supplement" || updatedInventoryType.Data.DisplayName != "Supplements" || updatedInventoryType.Data.Description != "" {
		t.Fatalf("expected updated inventory type metadata, got %+v", updatedInventoryType.Data)
	}

	updateInventoryTypeThroughTenantRoute := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/custom-asset-types/"+inventoryType.Data.ID, "Bearer dev:tenant-owner", map[string]any{
		"displayName": "Tenant Route Rename",
	})
	if updateInventoryTypeThroughTenantRoute.Code != http.StatusNotFound {
		t.Fatalf("expected inventory type through tenant route status %d, got %d with body %s", http.StatusNotFound, updateInventoryTypeThroughTenantRoute.Code, updateInventoryTypeThroughTenantRoute.Body.String())
	}
	assertSafeError(t, updateInventoryTypeThroughTenantRoute, "resource_not_found", "Resource not found.")

	updateInheritedType := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types/"+tenantType.Data.ID, "Bearer dev:inventory-owner", map[string]any{
		"displayName": "Inventory Rename",
	})
	if updateInheritedType.Code != http.StatusNotFound {
		t.Fatalf("expected inherited type update status %d, got %d with body %s", http.StatusNotFound, updateInheritedType.Code, updateInheritedType.Body.String())
	}
	assertSafeError(t, updateInheritedType, "resource_not_found", "Resource not found.")

	intruderUpdateType := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types/"+inventoryType.Data.ID, "Bearer dev:intruder", map[string]any{
		"displayName": "Intruder Rename",
	})
	if intruderUpdateType.Code != http.StatusForbidden {
		t.Fatalf("expected intruder update type status %d, got %d with body %s", http.StatusForbidden, intruderUpdateType.Code, intruderUpdateType.Body.String())
	}
	assertSafeError(t, intruderUpdateType, "forbidden", "Forbidden.")

	hiddenInventoryUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+hiddenInventoryID+"/custom-asset-types/"+inventoryType.Data.ID, "Bearer dev:inventory-owner", map[string]any{
		"displayName": "Hidden Rename",
	})
	if hiddenInventoryUpdate.Code != http.StatusForbidden {
		t.Fatalf("expected hidden inventory update status %d, got %d with body %s", http.StatusForbidden, hiddenInventoryUpdate.Code, hiddenInventoryUpdate.Body.String())
	}
	assertSafeError(t, hiddenInventoryUpdate, "forbidden", "Forbidden.")

	hiddenOwnerUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+hiddenInventoryID+"/custom-asset-types/"+inventoryType.Data.ID, "Bearer dev:hidden-owner", map[string]any{
		"displayName": "Hidden Owner Rename",
	})
	if hiddenOwnerUpdate.Code != http.StatusNotFound {
		t.Fatalf("expected hidden owner update status %d, got %d with body %s", http.StatusNotFound, hiddenOwnerUpdate.Code, hiddenOwnerUpdate.Body.String())
	}
	assertSafeError(t, hiddenOwnerUpdate, "resource_not_found", "Resource not found.")

	wrongTenantUpdate := performRequest(server, http.MethodPatch, "/tenants/"+otherTenantID+"/inventories/"+inventoryID+"/custom-asset-types/"+inventoryType.Data.ID, "Bearer dev:inventory-owner", map[string]any{
		"displayName": "Wrong Tenant Rename",
	})
	if wrongTenantUpdate.Code != http.StatusNotFound {
		t.Fatalf("expected wrong tenant update status %d, got %d with body %s", http.StatusNotFound, wrongTenantUpdate.Code, wrongTenantUpdate.Body.String())
	}
	assertSafeError(t, wrongTenantUpdate, "resource_not_found", "Resource not found.")

	missingInventoryUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/01ARZ3NDEKTSV4RRFFQ69G5FB1/custom-asset-types/"+inventoryType.Data.ID, "Bearer dev:inventory-owner", map[string]any{
		"displayName": "Missing Inventory Rename",
	})
	if missingInventoryUpdate.Code != http.StatusNotFound {
		t.Fatalf("expected missing inventory update status %d, got %d with body %s", http.StatusNotFound, missingInventoryUpdate.Code, missingInventoryUpdate.Body.String())
	}
	assertSafeError(t, missingInventoryUpdate, "resource_not_found", "Resource not found.")

	grantViewer := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]any{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if grantViewer.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, grantViewer.Code, grantViewer.Body.String())
	}
	viewerUpdateType := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types/"+inventoryType.Data.ID, "Bearer dev:viewer-user", map[string]any{
		"displayName": "Viewer Rename",
	})
	if viewerUpdateType.Code != http.StatusForbidden {
		t.Fatalf("expected viewer update type status %d, got %d with body %s", http.StatusForbidden, viewerUpdateType.Code, viewerUpdateType.Body.String())
	}
	assertSafeError(t, viewerUpdateType, "forbidden", "Forbidden.")

	emptyUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/custom-asset-types/"+tenantType.Data.ID, "Bearer dev:tenant-owner", map[string]any{})
	if emptyUpdate.Code != http.StatusBadRequest {
		t.Fatalf("expected empty update status %d, got %d with body %s", http.StatusBadRequest, emptyUpdate.Code, emptyUpdate.Body.String())
	}
	assertSafeError(t, emptyUpdate, "invalid_request", "Invalid request.")

	missingAuthenticationUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types/"+inventoryType.Data.ID, "", map[string]any{
		"displayName": "No Auth Rename",
	})
	if missingAuthenticationUpdate.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing authentication update status %d, got %d with body %s", http.StatusUnauthorized, missingAuthenticationUpdate.Code, missingAuthenticationUpdate.Body.String())
	}
	assertSafeError(t, missingAuthenticationUpdate, "authentication_required", "Authentication required.")
}

func TestCustomAssetTypeArchiveFlowAndAuthorization(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	const otherInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
			{id: otherTenantID, name: "Cabin", owner: "other-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Medicine", owner: "inventory-owner"},
			{id: otherInventoryID, tenantID: otherTenantID, name: "Other", owner: "other-owner"},
		},
		ids: []string{
			"01ARZ3NDEKTSV4RRFFQ69G5FAZ", "audit-tenant-type",
			"01ARZ3NDEKTSV4RRFFQ69G5FB0", "audit-inventory-type",
			"01ARZ3NDEKTSV4RRFFQ69G5FB1", "audit-expiration-field",
			"01ARZ3NDEKTSV4RRFFQ69G5FB2", "audit-typed-asset",
			"audit-viewer-grant", "viewer-grant-event", "viewer-grant-claim",
			"audit-editor-grant", "editor-grant-event", "editor-grant-claim",
			"audit-tenant-type-archive",
			"audit-inventory-type-archive",
			"01ARZ3NDEKTSV4RRFFQ69G5FB3",
		},
	}))

	createTenantType := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/custom-asset-types", "Bearer dev:tenant-owner", map[string]any{
		"key":         "medicine",
		"displayName": "Medicine",
	})
	if createTenantType.Code != http.StatusCreated {
		t.Fatalf("expected tenant type create status %d, got %d with body %s", http.StatusCreated, createTenantType.Code, createTenantType.Body.String())
	}
	tenantType := decodeCustomAssetType(t, createTenantType)
	if tenantType.Data.LifecycleState != "active" {
		t.Fatalf("expected active tenant type, got %+v", tenantType.Data)
	}

	createInventoryType := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types", "Bearer dev:inventory-owner", map[string]any{
		"key":         "supply",
		"displayName": "Supply",
	})
	if createInventoryType.Code != http.StatusCreated {
		t.Fatalf("expected inventory type create status %d, got %d with body %s", http.StatusCreated, createInventoryType.Code, createInventoryType.Body.String())
	}
	inventoryType := decodeCustomAssetType(t, createInventoryType)

	createField := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions", "Bearer dev:inventory-owner", map[string]any{
		"key":                "expires-on",
		"displayName":        "Expires On",
		"type":               "date",
		"applicability":      "custom_asset_types",
		"customAssetTypeIds": []string{inventoryType.Data.ID},
	})
	if createField.Code != http.StatusCreated {
		t.Fatalf("expected field create status %d, got %d with body %s", http.StatusCreated, createField.Code, createField.Body.String())
	}

	createAsset := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:inventory-owner", map[string]any{
		"kind":              "item",
		"title":             "Aspirin",
		"customAssetTypeId": inventoryType.Data.ID,
		"customFields": map[string]any{
			"expires-on": "2028-01-01",
		},
	})
	if createAsset.Code != http.StatusCreated {
		t.Fatalf("expected typed asset create status %d, got %d with body %s", http.StatusCreated, createAsset.Code, createAsset.Body.String())
	}
	createdAsset := decodeAsset(t, createAsset)

	grantViewer := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]any{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if grantViewer.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, grantViewer.Code, grantViewer.Body.String())
	}
	grantEditor := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]any{
		"principalId":  "editor-user",
		"relationship": "editor",
	})
	if grantEditor.Code != http.StatusCreated {
		t.Fatalf("expected editor grant status %d, got %d with body %s", http.StatusCreated, grantEditor.Code, grantEditor.Body.String())
	}

	for _, item := range []struct {
		name          string
		authorization string
	}{
		{name: "inventory owner", authorization: "Bearer dev:inventory-owner"},
		{name: "viewer", authorization: "Bearer dev:viewer-user"},
		{name: "editor", authorization: "Bearer dev:editor-user"},
		{name: "intruder", authorization: "Bearer dev:intruder"},
	} {
		t.Run(item.name+" cannot archive tenant custom asset type", func(t *testing.T) {
			response := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/custom-asset-types/"+tenantType.Data.ID+"/archive", item.authorization, nil)
			if response.Code != http.StatusForbidden {
				t.Fatalf("expected tenant archive status %d, got %d with body %s", http.StatusForbidden, response.Code, response.Body.String())
			}
			assertSafeError(t, response, "forbidden", "Forbidden.")
		})
	}

	for _, item := range []struct {
		name          string
		authorization string
		status        int
		code          string
	}{
		{name: "missing auth", status: http.StatusUnauthorized, code: "authentication_required"},
		{name: "malformed auth", authorization: "Bearer nope", status: http.StatusUnauthorized, code: "authentication_required"},
		{name: "viewer", authorization: "Bearer dev:viewer-user", status: http.StatusForbidden, code: "forbidden"},
		{name: "editor", authorization: "Bearer dev:editor-user", status: http.StatusForbidden, code: "forbidden"},
		{name: "intruder", authorization: "Bearer dev:intruder", status: http.StatusForbidden, code: "forbidden"},
	} {
		t.Run(item.name+" cannot archive inventory custom asset type", func(t *testing.T) {
			response := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types/"+inventoryType.Data.ID+"/archive", item.authorization, nil)
			if response.Code != item.status {
				t.Fatalf("expected archive status %d, got %d with body %s", item.status, response.Code, response.Body.String())
			}
			assertSafeError(t, response, item.code, map[int]string{http.StatusUnauthorized: "Authentication required.", http.StatusForbidden: "Forbidden."}[item.status])
		})
	}

	wrongTenantArchive := performRequest(server, http.MethodPatch, "/tenants/"+otherTenantID+"/inventories/"+otherInventoryID+"/custom-asset-types/"+inventoryType.Data.ID+"/archive", "Bearer dev:other-owner", nil)
	if wrongTenantArchive.Code != http.StatusNotFound {
		t.Fatalf("expected wrong tenant archive status %d, got %d with body %s", http.StatusNotFound, wrongTenantArchive.Code, wrongTenantArchive.Body.String())
	}
	assertSafeError(t, wrongTenantArchive, "resource_not_found", "Resource not found.")

	inventoryTypeThroughTenantRoute := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/custom-asset-types/"+inventoryType.Data.ID+"/archive", "Bearer dev:tenant-owner", nil)
	if inventoryTypeThroughTenantRoute.Code != http.StatusNotFound {
		t.Fatalf("expected inventory type through tenant route status %d, got %d with body %s", http.StatusNotFound, inventoryTypeThroughTenantRoute.Code, inventoryTypeThroughTenantRoute.Body.String())
	}
	assertSafeError(t, inventoryTypeThroughTenantRoute, "resource_not_found", "Resource not found.")

	tenantTypeThroughInventoryRoute := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types/"+tenantType.Data.ID+"/archive", "Bearer dev:inventory-owner", nil)
	if tenantTypeThroughInventoryRoute.Code != http.StatusNotFound {
		t.Fatalf("expected tenant type through inventory route status %d, got %d with body %s", http.StatusNotFound, tenantTypeThroughInventoryRoute.Code, tenantTypeThroughInventoryRoute.Body.String())
	}
	assertSafeError(t, tenantTypeThroughInventoryRoute, "resource_not_found", "Resource not found.")

	archiveTenantType := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/custom-asset-types/"+tenantType.Data.ID+"/archive", "Bearer dev:tenant-owner", nil)
	if archiveTenantType.Code != http.StatusOK {
		t.Fatalf("expected tenant type archive status %d, got %d with body %s", http.StatusOK, archiveTenantType.Code, archiveTenantType.Body.String())
	}
	archivedTenantType := decodeCustomAssetType(t, archiveTenantType)
	if archivedTenantType.Data.LifecycleState != "archived" {
		t.Fatalf("expected archived tenant type, got %+v", archivedTenantType.Data)
	}

	archiveInventoryType := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types/"+inventoryType.Data.ID+"/archive", "Bearer dev:inventory-owner", nil)
	if archiveInventoryType.Code != http.StatusOK {
		t.Fatalf("expected inventory type archive status %d, got %d with body %s", http.StatusOK, archiveInventoryType.Code, archiveInventoryType.Body.String())
	}
	archivedInventoryType := decodeCustomAssetType(t, archiveInventoryType)
	if archivedInventoryType.Data.LifecycleState != "archived" {
		t.Fatalf("expected archived inventory type, got %+v", archivedInventoryType.Data)
	}

	listTypes := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types?limit=10", "Bearer dev:inventory-owner", nil)
	if listTypes.Code != http.StatusOK {
		t.Fatalf("expected type list status %d, got %d with body %s", http.StatusOK, listTypes.Code, listTypes.Body.String())
	}
	if listed := decodeCustomAssetTypeList(t, listTypes); len(listed.Data) != 0 {
		t.Fatalf("expected archived custom asset types hidden from list, got %+v", listed.Data)
	}

	listAssets := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets?limit=10", "Bearer dev:inventory-owner", nil)
	if listAssets.Code != http.StatusOK {
		t.Fatalf("expected asset list status %d, got %d with body %s", http.StatusOK, listAssets.Code, listAssets.Body.String())
	}
	assetList := decodeAssetList(t, listAssets)
	if len(assetList.Data) != 1 || assetList.Data[0].ID != createdAsset.Data.ID || assetList.Data[0].CustomAssetTypeID != inventoryType.Data.ID {
		t.Fatalf("expected existing asset custom asset type reference to remain visible, got %+v", assetList.Data)
	}

	reuseArchivedType := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:inventory-owner", map[string]any{
		"kind":              "item",
		"title":             "Ibuprofen",
		"customAssetTypeId": inventoryType.Data.ID,
	})
	if reuseArchivedType.Code != http.StatusNotFound {
		t.Fatalf("expected archived type asset create status %d, got %d with body %s", http.StatusNotFound, reuseArchivedType.Code, reuseArchivedType.Body.String())
	}
	assertSafeError(t, reuseArchivedType, "resource_not_found", "Resource not found.")

	targetArchivedType := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions", "Bearer dev:inventory-owner", map[string]any{
		"key":                "dose",
		"displayName":        "Dose",
		"type":               "text",
		"applicability":      "custom_asset_types",
		"customAssetTypeIds": []string{inventoryType.Data.ID},
	})
	if targetArchivedType.Code != http.StatusNotFound {
		t.Fatalf("expected archived type field target status %d, got %d with body %s", http.StatusNotFound, targetArchivedType.Code, targetArchivedType.Body.String())
	}
	assertSafeError(t, targetArchivedType, "resource_not_found", "Resource not found.")
}
