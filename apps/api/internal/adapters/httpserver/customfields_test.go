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
