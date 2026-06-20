package httpserver

import (
	"net/http"
	"testing"
)

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
		{name: "create tenant custom field definition", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/custom-field-definitions", body: map[string]string{"key": "serial", "displayName": "Serial", "type": "text"}},
		{name: "list tenant custom field definitions", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/custom-field-definitions"},
		{name: "list tenant audit records", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/audit-records"},
		{name: "create asset", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/assets", body: map[string]string{"kind": "item", "title": "Drill"}},
		{name: "list assets", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/assets"},
		{name: "update asset", method: http.MethodPatch, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/assets/01ARZ3NDEKTSV4RRFFQ69G5FAX", body: map[string]string{"title": "Drill"}},
		{name: "grant inventory access", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/access-grants", body: map[string]string{"principalId": "viewer", "relationship": "viewer"}},
		{name: "list inventory access grants", method: http.MethodGet, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/access-grants"},
		{name: "create inventory access invitation", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/access-invitations", body: map[string]string{"email": "viewer@example.com", "relationship": "viewer"}},
		{name: "accept inventory access invitation", method: http.MethodPost, path: "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAV/inventories/01ARZ3NDEKTSV4RRFFQ69G5FAW/access-invitations/01ARZ3NDEKTSV4RRFFQ69G5FAX/accept", body: map[string]string{"acceptanceToken": "token"}},
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
