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
