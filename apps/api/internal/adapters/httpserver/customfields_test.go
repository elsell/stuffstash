package httpserver

import (
	"net/http"
	"testing"
)

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
			"audit-custom-field-viewer-grant", "custom-field-viewer-grant-event", "custom-field-viewer-grant-claim",
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

func TestCustomFieldDefinitionUpdateFlowAndAuthorization(t *testing.T) {
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
			"audit-tenant-definition-update",
			"01ARZ3NDEKTSV4RRFFQ69G5FAZ", "audit-inventory-definition",
			"audit-inventory-definition-update",
			"audit-field-viewer-grant", "field-viewer-grant-event", "field-viewer-grant-claim",
		},
	}))

	createTenantDefinition := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/custom-field-definitions", "Bearer dev:tenant-owner", map[string]any{
		"key":         "serial",
		"displayName": "Serial",
		"type":        "text",
	})
	if createTenantDefinition.Code != http.StatusCreated {
		t.Fatalf("expected tenant definition create status %d, got %d with body %s", http.StatusCreated, createTenantDefinition.Code, createTenantDefinition.Body.String())
	}
	tenantDefinition := decodeCustomFieldDefinition(t, createTenantDefinition)
	updateTenantDefinition := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/custom-field-definitions/"+tenantDefinition.Data.ID, "Bearer dev:tenant-owner", map[string]any{
		"displayName": "Serial Number",
	})
	if updateTenantDefinition.Code != http.StatusOK {
		t.Fatalf("expected tenant definition update status %d, got %d with body %s", http.StatusOK, updateTenantDefinition.Code, updateTenantDefinition.Body.String())
	}
	updatedTenantDefinition := decodeCustomFieldDefinition(t, updateTenantDefinition)
	if updatedTenantDefinition.Data.ID != tenantDefinition.Data.ID || updatedTenantDefinition.Data.Key != "serial" || updatedTenantDefinition.Data.Type != "text" || updatedTenantDefinition.Data.DisplayName != "Serial Number" {
		t.Fatalf("expected updated tenant definition metadata, got %+v", updatedTenantDefinition.Data)
	}
	tenantAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/audit-records?limit=50", "Bearer dev:tenant-owner", nil)
	if tenantAudit.Code != http.StatusOK {
		t.Fatalf("expected tenant audit status %d, got %d with body %s", http.StatusOK, tenantAudit.Code, tenantAudit.Body.String())
	}
	if !auditRecordsContainAction(decodeAuditRecordList(t, tenantAudit).Data, "custom_field_definition.updated") {
		t.Fatalf("expected tenant audit to include custom field update action, got %s", tenantAudit.Body.String())
	}

	crossTenantDefinitionUpdate := performRequest(server, http.MethodPatch, "/tenants/"+otherTenantID+"/custom-field-definitions/"+tenantDefinition.Data.ID, "Bearer dev:other-owner", map[string]any{
		"displayName": "Cabin Serial",
	})
	if crossTenantDefinitionUpdate.Code != http.StatusNotFound {
		t.Fatalf("expected cross-tenant definition update status %d, got %d with body %s", http.StatusNotFound, crossTenantDefinitionUpdate.Code, crossTenantDefinitionUpdate.Body.String())
	}
	assertSafeError(t, crossTenantDefinitionUpdate, "resource_not_found", "Resource not found.")

	createInventoryDefinition := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions", "Bearer dev:inventory-owner", map[string]any{
		"key":         "condition",
		"displayName": "Condition",
		"type":        "enum",
		"enumOptions": []string{"new", "used"},
	})
	if createInventoryDefinition.Code != http.StatusCreated {
		t.Fatalf("expected inventory definition create status %d, got %d with body %s", http.StatusCreated, createInventoryDefinition.Code, createInventoryDefinition.Body.String())
	}
	inventoryDefinition := decodeCustomFieldDefinition(t, createInventoryDefinition)
	updateInventoryDefinition := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions/"+inventoryDefinition.Data.ID, "Bearer dev:inventory-owner", map[string]any{
		"displayName": "Item Condition",
	})
	if updateInventoryDefinition.Code != http.StatusOK {
		t.Fatalf("expected inventory definition update status %d, got %d with body %s", http.StatusOK, updateInventoryDefinition.Code, updateInventoryDefinition.Body.String())
	}
	updatedInventoryDefinition := decodeCustomFieldDefinition(t, updateInventoryDefinition)
	if updatedInventoryDefinition.Data.ID != inventoryDefinition.Data.ID || updatedInventoryDefinition.Data.Key != "condition" || updatedInventoryDefinition.Data.Type != "enum" || updatedInventoryDefinition.Data.DisplayName != "Item Condition" {
		t.Fatalf("expected updated inventory definition metadata, got %+v", updatedInventoryDefinition.Data)
	}
	inventoryAudit := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/audit-records?limit=50", "Bearer dev:inventory-owner", nil)
	if inventoryAudit.Code != http.StatusOK {
		t.Fatalf("expected inventory audit status %d, got %d with body %s", http.StatusOK, inventoryAudit.Code, inventoryAudit.Body.String())
	}
	if !auditRecordsContainAction(decodeAuditRecordList(t, inventoryAudit).Data, "custom_field_definition.updated") {
		t.Fatalf("expected inventory audit to include custom field update action, got %s", inventoryAudit.Body.String())
	}

	updateInventoryDefinitionThroughTenantRoute := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/custom-field-definitions/"+inventoryDefinition.Data.ID, "Bearer dev:tenant-owner", map[string]any{
		"displayName": "Tenant Route Rename",
	})
	if updateInventoryDefinitionThroughTenantRoute.Code != http.StatusNotFound {
		t.Fatalf("expected inventory definition through tenant route status %d, got %d with body %s", http.StatusNotFound, updateInventoryDefinitionThroughTenantRoute.Code, updateInventoryDefinitionThroughTenantRoute.Body.String())
	}
	assertSafeError(t, updateInventoryDefinitionThroughTenantRoute, "resource_not_found", "Resource not found.")

	updateInheritedDefinition := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions/"+tenantDefinition.Data.ID, "Bearer dev:inventory-owner", map[string]any{
		"displayName": "Inventory Rename",
	})
	if updateInheritedDefinition.Code != http.StatusNotFound {
		t.Fatalf("expected inherited definition update status %d, got %d with body %s", http.StatusNotFound, updateInheritedDefinition.Code, updateInheritedDefinition.Body.String())
	}
	assertSafeError(t, updateInheritedDefinition, "resource_not_found", "Resource not found.")

	intruderUpdateDefinition := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions/"+inventoryDefinition.Data.ID, "Bearer dev:intruder", map[string]any{
		"displayName": "Intruder Rename",
	})
	if intruderUpdateDefinition.Code != http.StatusForbidden {
		t.Fatalf("expected intruder update definition status %d, got %d with body %s", http.StatusForbidden, intruderUpdateDefinition.Code, intruderUpdateDefinition.Body.String())
	}
	assertSafeError(t, intruderUpdateDefinition, "forbidden", "Forbidden.")

	hiddenInventoryUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+hiddenInventoryID+"/custom-field-definitions/"+inventoryDefinition.Data.ID, "Bearer dev:inventory-owner", map[string]any{
		"displayName": "Hidden Rename",
	})
	if hiddenInventoryUpdate.Code != http.StatusForbidden {
		t.Fatalf("expected hidden inventory update status %d, got %d with body %s", http.StatusForbidden, hiddenInventoryUpdate.Code, hiddenInventoryUpdate.Body.String())
	}
	assertSafeError(t, hiddenInventoryUpdate, "forbidden", "Forbidden.")

	hiddenOwnerUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+hiddenInventoryID+"/custom-field-definitions/"+inventoryDefinition.Data.ID, "Bearer dev:hidden-owner", map[string]any{
		"displayName": "Hidden Owner Rename",
	})
	if hiddenOwnerUpdate.Code != http.StatusNotFound {
		t.Fatalf("expected hidden owner update status %d, got %d with body %s", http.StatusNotFound, hiddenOwnerUpdate.Code, hiddenOwnerUpdate.Body.String())
	}
	assertSafeError(t, hiddenOwnerUpdate, "resource_not_found", "Resource not found.")

	wrongTenantUpdate := performRequest(server, http.MethodPatch, "/tenants/"+otherTenantID+"/inventories/"+inventoryID+"/custom-field-definitions/"+inventoryDefinition.Data.ID, "Bearer dev:inventory-owner", map[string]any{
		"displayName": "Wrong Tenant Rename",
	})
	if wrongTenantUpdate.Code != http.StatusNotFound {
		t.Fatalf("expected wrong tenant update status %d, got %d with body %s", http.StatusNotFound, wrongTenantUpdate.Code, wrongTenantUpdate.Body.String())
	}
	assertSafeError(t, wrongTenantUpdate, "resource_not_found", "Resource not found.")

	missingInventoryUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/01ARZ3NDEKTSV4RRFFQ69G5FB2/custom-field-definitions/"+inventoryDefinition.Data.ID, "Bearer dev:inventory-owner", map[string]any{
		"displayName": "Missing Inventory Rename",
	})
	if missingInventoryUpdate.Code != http.StatusNotFound {
		t.Fatalf("expected missing inventory update status %d, got %d with body %s", http.StatusNotFound, missingInventoryUpdate.Code, missingInventoryUpdate.Body.String())
	}
	assertSafeError(t, missingInventoryUpdate, "resource_not_found", "Resource not found.")

	viewerGrant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId":  "viewer-user",
		"relationship": "viewer",
	})
	if viewerGrant.Code != http.StatusCreated {
		t.Fatalf("expected viewer grant status %d, got %d with body %s", http.StatusCreated, viewerGrant.Code, viewerGrant.Body.String())
	}
	viewerUpdateDefinition := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions/"+inventoryDefinition.Data.ID, "Bearer dev:viewer-user", map[string]any{
		"displayName": "Viewer Rename",
	})
	if viewerUpdateDefinition.Code != http.StatusForbidden {
		t.Fatalf("expected viewer update definition status %d, got %d with body %s", http.StatusForbidden, viewerUpdateDefinition.Code, viewerUpdateDefinition.Body.String())
	}
	assertSafeError(t, viewerUpdateDefinition, "forbidden", "Forbidden.")

	immutableFieldUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/custom-field-definitions/"+tenantDefinition.Data.ID, "Bearer dev:tenant-owner", map[string]any{
		"displayName":        "Serial Label",
		"key":                "changed",
		"type":               "enum",
		"enumOptions":        []string{"new"},
		"applicability":      "custom_asset_types",
		"customAssetTypeIds": []string{"01ARZ3NDEKTSV4RRFFQ69G5FB3"},
	})
	if immutableFieldUpdate.Code != http.StatusBadRequest {
		t.Fatalf("expected immutable field update status %d, got %d with body %s", http.StatusBadRequest, immutableFieldUpdate.Code, immutableFieldUpdate.Body.String())
	}
	assertSafeError(t, immutableFieldUpdate, "invalid_request", "Invalid request.")

	emptyUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/custom-field-definitions/"+tenantDefinition.Data.ID, "Bearer dev:tenant-owner", map[string]any{})
	if emptyUpdate.Code != http.StatusBadRequest {
		t.Fatalf("expected empty update status %d, got %d with body %s", http.StatusBadRequest, emptyUpdate.Code, emptyUpdate.Body.String())
	}
	assertSafeError(t, emptyUpdate, "invalid_request", "Invalid request.")

	missingAuthenticationUpdate := performRequest(server, http.MethodPatch, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions/"+inventoryDefinition.Data.ID, "", map[string]any{
		"displayName": "No Auth Rename",
	})
	if missingAuthenticationUpdate.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing authentication update status %d, got %d with body %s", http.StatusUnauthorized, missingAuthenticationUpdate.Code, missingAuthenticationUpdate.Body.String())
	}
	assertSafeError(t, missingAuthenticationUpdate, "authentication_required", "Authentication required.")
}
