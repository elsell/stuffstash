package httpserver

import (
	"net/http"
	"testing"
)

func TestMyTenantsListShowsOnlyAccessibleTenantsWithEffectiveAccess(t *testing.T) {
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants: []seedTenant{
			{id: "tenant-owned", name: "Home", owner: "owner"},
			{id: "tenant-viewed", name: "Shared", owner: "other-owner"},
			{id: "tenant-hidden", name: "Hidden", owner: "hidden-owner"},
		},
		inventories: []seedInventory{
			{id: "inventory-viewed", tenantID: "tenant-viewed", name: "Tools", owner: "other-owner"},
		},
	}))

	grant := performRequest(server, http.MethodPost, "/tenants/tenant-viewed/inventories/inventory-viewed/access-grants", "Bearer dev:other-owner", map[string]string{
		"principalId":  "owner",
		"relationship": "viewer",
	})
	if grant.Code != http.StatusCreated {
		t.Fatalf("expected grant status %d, got %d with body %s", http.StatusCreated, grant.Code, grant.Body.String())
	}

	first := performRequest(server, http.MethodGet, "/me/tenants?limit=1", "Bearer dev:owner", nil)
	if first.Code != http.StatusOK {
		t.Fatalf("expected first tenant page status %d, got %d with body %s", http.StatusOK, first.Code, first.Body.String())
	}
	firstPage := decodeMyTenantList(t, first)
	if len(firstPage.Data) != 1 || firstPage.Data[0].ID != "tenant-owned" {
		t.Fatalf("expected first page to contain owned tenant, got %+v", firstPage.Data)
	}
	if firstPage.Data[0].Access.Relationship != "owner" || !accessContainsPermission(firstPage.Data[0].Access.Permissions, "configure") || !accessContainsPermission(firstPage.Data[0].Access.Permissions, "create_inventory") {
		t.Fatalf("expected owner tenant access metadata, got %+v", firstPage.Data[0].Access)
	}
	if firstPage.Meta.Pagination == nil || !firstPage.Meta.Pagination.HasMore || firstPage.Meta.Pagination.NextCursor == nil {
		t.Fatalf("expected paginated first page metadata, got %+v", firstPage.Meta)
	}

	second := performRequest(server, http.MethodGet, "/me/tenants?limit=1&cursor="+*firstPage.Meta.Pagination.NextCursor, "Bearer dev:owner", nil)
	if second.Code != http.StatusOK {
		t.Fatalf("expected second tenant page status %d, got %d with body %s", http.StatusOK, second.Code, second.Body.String())
	}
	secondPage := decodeMyTenantList(t, second)
	if len(secondPage.Data) != 1 || secondPage.Data[0].ID != "tenant-viewed" {
		t.Fatalf("expected second page to contain shared tenant only, got %+v", secondPage.Data)
	}
	if secondPage.Data[0].Access.Relationship != "viewer" || !accessContainsPermission(secondPage.Data[0].Access.Permissions, "view") || accessContainsPermission(secondPage.Data[0].Access.Permissions, "configure") {
		t.Fatalf("expected viewer tenant access metadata, got %+v", secondPage.Data[0].Access)
	}
	if secondPage.Meta.Pagination == nil || secondPage.Meta.Pagination.HasMore || secondPage.Meta.Pagination.NextCursor != nil {
		t.Fatalf("expected final page metadata, got %+v", secondPage.Meta)
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
		{name: "list my tenants", method: http.MethodGet, path: "/me/tenants"},
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
		{name: "create inventory access invitation", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/access-invitations", body: map[string]string{"email": "viewer@example.com", "relationship": "viewer"}},
		{name: "list inventory access invitations", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/access-invitations"},
		{name: "accept inventory access invitation", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/access-invitations/01ARZ3NDEKTSV4RRFFQ69G5FAX/accept", body: map[string]string{"acceptanceToken": "token"}},
		{name: "update inventory access invitation expiration", method: http.MethodPatch, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/access-invitations/01ARZ3NDEKTSV4RRFFQ69G5FAX/expiration", body: map[string]string{"expiresAt": "2026-06-20T12:00:00Z"}},
		{name: "revoke inventory access invitation", method: http.MethodDelete, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/access-invitations/01ARZ3NDEKTSV4RRFFQ69G5FAX"},
		{name: "create tenant custom asset type", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/custom-asset-types", body: map[string]string{"key": "medicine", "displayName": "Medicine"}},
		{name: "list tenant custom asset types", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/custom-asset-types"},
		{name: "update tenant custom asset type", method: http.MethodPatch, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/custom-asset-types/01ARZ3NDEKTSV4RRFFQ69G5FAX", body: map[string]string{"displayName": "Medicine"}},
		{name: "archive tenant custom asset type", method: http.MethodPatch, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/custom-asset-types/01ARZ3NDEKTSV4RRFFQ69G5FAX/archive"},
		{name: "create inventory custom asset type", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/custom-asset-types", body: map[string]string{"key": "medicine", "displayName": "Medicine"}},
		{name: "list inventory custom asset types", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/custom-asset-types"},
		{name: "update inventory custom asset type", method: http.MethodPatch, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/custom-asset-types/01ARZ3NDEKTSV4RRFFQ69G5FAX", body: map[string]string{"displayName": "Medicine"}},
		{name: "archive inventory custom asset type", method: http.MethodPatch, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/custom-asset-types/01ARZ3NDEKTSV4RRFFQ69G5FAX/archive"},
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
