package httpserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestCustomizationLifecycleListsPreserveScopeAndAuthorization(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const otherTenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	const peerInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	const otherTenantInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAZ"
	const hiddenInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FB4"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: tenantID, name: "Home", owner: "tenant-owner"},
			{id: otherTenantID, name: "Cabin", owner: "tenant-owner"},
		},
		inventories: []seedInventory{
			{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "inventory-owner"},
			{id: peerInventoryID, tenantID: tenantID, name: "Supplies", owner: "inventory-owner"},
			{id: otherTenantInventoryID, tenantID: otherTenantID, name: "Cabin", owner: "inventory-owner"},
			{id: hiddenInventoryID, tenantID: tenantID, name: "Hidden", owner: "hidden-owner"},
		},
		ids: []string{
			"01ARZ3NDEKTSV4RRFFQ69G5FB0", "audit-create-tenant-type",
			"01ARZ3NDEKTSV4RRFFQ69G5FB1", "audit-create-inventory-type",
			"01ARZ3NDEKTSV4RRFFQ69G5FB2", "audit-create-tenant-field",
			"01ARZ3NDEKTSV4RRFFQ69G5FB3", "audit-create-inventory-field",
			"audit-archive-tenant-type", "audit-archive-inventory-type",
			"audit-archive-tenant-field", "audit-archive-inventory-field",
			"audit-viewer-grant", "viewer-grant-event", "viewer-grant-claim",
		},
	}))

	tenantType := createCustomAssetTypeForTest(t, server, "/tenants/"+tenantID+"/custom-asset-types", "Bearer dev:tenant-owner", "tenant-type")
	inventoryType := createCustomAssetTypeForTest(t, server, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types", "Bearer dev:inventory-owner", "inventory-type")
	tenantField := createCustomFieldDefinitionForLifecycleTest(t, server, "/tenants/"+tenantID+"/custom-field-definitions", "Bearer dev:tenant-owner", "tenant-field")
	inventoryField := createCustomFieldDefinitionForLifecycleTest(t, server, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-field-definitions", "Bearer dev:inventory-owner", "inventory-field")

	for _, item := range []struct {
		path string
		auth string
	}{
		{"/tenants/" + tenantID + "/custom-asset-types/" + tenantType.Data.ID + "/archive", "Bearer dev:tenant-owner"},
		{"/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-asset-types/" + inventoryType.Data.ID + "/archive", "Bearer dev:inventory-owner"},
		{"/tenants/" + tenantID + "/custom-field-definitions/" + tenantField.Data.ID + "/archive", "Bearer dev:tenant-owner"},
		{"/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-field-definitions/" + inventoryField.Data.ID + "/archive", "Bearer dev:inventory-owner"},
	} {
		response := performRequest(server, http.MethodPatch, item.path, item.auth, nil)
		if response.Code != http.StatusOK {
			t.Fatalf("archive %s: status=%d body=%s", item.path, response.Code, response.Body.String())
		}
	}
	grantViewer := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/access-grants", "Bearer dev:inventory-owner", map[string]string{
		"principalId": "viewer-user", "relationship": "viewer",
	})
	if grantViewer.Code != http.StatusCreated {
		t.Fatalf("grant viewer: status=%d body=%s", grantViewer.Code, grantViewer.Body.String())
	}

	resources := []struct {
		name          string
		tenantPath    string
		inventoryPath string
		decodeCount   func(*testing.T, *httptest.ResponseRecorder) ([]string, []string)
	}{
		{
			name: "custom fields", tenantPath: "/tenants/" + tenantID + "/custom-field-definitions", inventoryPath: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-field-definitions",
			decodeCount: func(t *testing.T, response *httptest.ResponseRecorder) ([]string, []string) {
				body := decodeCustomFieldDefinitionList(t, response)
				ids, scopes := make([]string, len(body.Data)), make([]string, len(body.Data))
				for i, item := range body.Data {
					ids[i], scopes[i] = item.ID, item.Scope
				}
				return ids, scopes
			},
		},
		{
			name: "custom asset types", tenantPath: "/tenants/" + tenantID + "/custom-asset-types", inventoryPath: "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-asset-types",
			decodeCount: func(t *testing.T, response *httptest.ResponseRecorder) ([]string, []string) {
				body := decodeCustomAssetTypeList(t, response)
				ids, scopes := make([]string, len(body.Data)), make([]string, len(body.Data))
				for i, item := range body.Data {
					ids[i], scopes[i] = item.ID, item.Scope
				}
				return ids, scopes
			},
		},
	}

	for _, resource := range resources {
		t.Run(resource.name, func(t *testing.T) {
			for _, lifecycle := range []string{"", "active"} {
				response := performRequest(server, http.MethodGet, resource.inventoryPath+customizationLifecycleQuery(lifecycle, 10, ""), "Bearer dev:inventory-owner", nil)
				if response.Code != http.StatusOK {
					t.Fatalf("list %q: status=%d body=%s", lifecycle, response.Code, response.Body.String())
				}
				ids, _ := resource.decodeCount(t, response)
				if len(ids) != 0 {
					t.Fatalf("expected %q view to hide archived records, got %v", lifecycle, ids)
				}
			}

			for _, lifecycle := range []string{"archived", "all"} {
				response := performRequest(server, http.MethodGet, resource.inventoryPath+customizationLifecycleQuery(lifecycle, 10, ""), "Bearer dev:inventory-owner", nil)
				if response.Code != http.StatusOK {
					t.Fatalf("list %q: status=%d body=%s", lifecycle, response.Code, response.Body.String())
				}
				ids, scopes := resource.decodeCount(t, response)
				if len(ids) != 2 || scopes[0] != "tenant" || scopes[1] != "inventory" {
					t.Fatalf("expected inherited then local %q records, ids=%v scopes=%v", lifecycle, ids, scopes)
				}
			}
			viewerResponse := performRequest(server, http.MethodGet, resource.inventoryPath+"?lifecycleState=archived", "Bearer dev:viewer-user", nil)
			if viewerResponse.Code != http.StatusOK {
				t.Fatalf("viewer archived list: status=%d body=%s", viewerResponse.Code, viewerResponse.Body.String())
			}
			invalidLifecycle := performRequest(server, http.MethodGet, resource.inventoryPath+"?lifecycleState=deleted", "Bearer dev:inventory-owner", nil)
			if invalidLifecycle.Code != http.StatusUnprocessableEntity {
				t.Fatalf("invalid lifecycle: expected %d, got %d body=%s", http.StatusUnprocessableEntity, invalidLifecycle.Code, invalidLifecycle.Body.String())
			}

			for _, lifecycle := range []string{"", "active", "archived", "all"} {
				tenantResponse := performRequest(server, http.MethodGet, resource.tenantPath+customizationLifecycleQuery(lifecycle, 10, ""), "Bearer dev:tenant-owner", nil)
				if tenantResponse.Code != http.StatusOK {
					t.Fatalf("tenant %q list: status=%d body=%s", lifecycle, tenantResponse.Code, tenantResponse.Body.String())
				}
				ids, scopes := resource.decodeCount(t, tenantResponse)
				expectedCount := 0
				if lifecycle == "archived" || lifecycle == "all" {
					expectedCount = 1
				}
				if len(ids) != expectedCount || expectedCount == 1 && scopes[0] != "tenant" {
					t.Fatalf("expected %d tenant %q records, ids=%v scopes=%v", expectedCount, lifecycle, ids, scopes)
				}
			}

			for _, denied := range []struct {
				name, path, auth string
				status           int
			}{
				{"unauthenticated", resource.inventoryPath + "?lifecycleState=archived", "", http.StatusUnauthorized},
				{"unrelated", resource.inventoryPath + "?lifecycleState=all", "Bearer dev:intruder", http.StatusForbidden},
				{"inventory owner tenant list", resource.tenantPath + "?lifecycleState=archived", "Bearer dev:inventory-owner", http.StatusForbidden},
				{"hidden inventory", strings.Replace(resource.inventoryPath, inventoryID, hiddenInventoryID, 1) + "?lifecycleState=archived", "Bearer dev:inventory-owner", http.StatusForbidden},
				{"wrong tenant", "/tenants/" + otherTenantID + "/inventories/" + inventoryID + resource.inventoryPath[strings.LastIndex(resource.inventoryPath, "/"):] + "?lifecycleState=all", "Bearer dev:inventory-owner", http.StatusNotFound},
			} {
				t.Run(denied.name, func(t *testing.T) {
					response := performRequest(server, http.MethodGet, denied.path, denied.auth, nil)
					if response.Code != denied.status {
						t.Fatalf("expected %d, got %d body=%s", denied.status, response.Code, response.Body.String())
					}
				})
			}

			first := performRequest(server, http.MethodGet, resource.inventoryPath+"?lifecycleState=archived&limit=1", "Bearer dev:inventory-owner", nil)
			cursor := customizationNextCursor(t, resource.name, first)
			for _, cursorReuse := range []struct {
				path string
				auth string
			}{
				{strings.Replace(resource.inventoryPath, inventoryID, peerInventoryID, 1) + customizationLifecycleQuery("archived", 1, cursor), "Bearer dev:inventory-owner"},
				{strings.Replace(strings.Replace(resource.inventoryPath, tenantID, otherTenantID, 1), inventoryID, otherTenantInventoryID, 1) + customizationLifecycleQuery("archived", 1, cursor), "Bearer dev:inventory-owner"},
				{resource.inventoryPath + customizationLifecycleQuery("all", 1, cursor), "Bearer dev:inventory-owner"},
				{resource.tenantPath + customizationLifecycleQuery("archived", 1, cursor), "Bearer dev:tenant-owner"},
			} {
				response := performRequest(server, http.MethodGet, cursorReuse.path, cursorReuse.auth, nil)
				if response.Code != http.StatusBadRequest {
					t.Fatalf("expected scoped cursor rejection for %s, got %d body=%s", cursorReuse.path, response.Code, response.Body.String())
				}
			}
		})
	}
}

func createCustomFieldDefinitionForLifecycleTest(t *testing.T, server *http.Server, path string, authorization string, key string) customFieldDefinitionBody {
	t.Helper()
	response := performRequest(server, http.MethodPost, path, authorization, map[string]any{
		"key": key, "displayName": key, "type": "text",
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("create custom field %q: status=%d body=%s", key, response.Code, response.Body.String())
	}
	return decodeCustomFieldDefinition(t, response)
}

func customizationLifecycleQuery(lifecycle string, limit int, cursor string) string {
	values := url.Values{}
	if lifecycle != "" {
		values.Set("lifecycleState", lifecycle)
	}
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}
	if cursor != "" {
		values.Set("cursor", cursor)
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}

func customizationNextCursor(t *testing.T, resource string, response *httptest.ResponseRecorder) string {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("list first archived %s page: status=%d body=%s", resource, response.Code, response.Body.String())
	}
	if resource == "custom fields" {
		body := decodeCustomFieldDefinitionList(t, response)
		if body.Meta.Pagination == nil || body.Meta.Pagination.NextCursor == nil {
			t.Fatalf("expected custom field cursor, got %+v", body.Meta)
		}
		return *body.Meta.Pagination.NextCursor
	}
	body := decodeCustomAssetTypeList(t, response)
	if body.Meta.Pagination == nil || body.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected custom asset type cursor, got %+v", body.Meta)
	}
	return *body.Meta.Pagination.NextCursor
}
