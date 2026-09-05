package httpserver

import (
	"net/http"
	"testing"
)

func TestOnboardingInventoryCreationPreservesTenantAuthorization(t *testing.T) {
	const tenantID = "tenant-home"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "owner"}},
		inventories: []seedInventory{{id: "inventory-shared", tenantID: tenantID, name: "Shared", owner: "owner"}},
	}))
	grant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/inventory-shared/access-grants", "Bearer dev:owner", map[string]string{"principalId": "viewer", "relationship": "viewer"})
	if grant.Code != http.StatusCreated {
		t.Fatalf("grant failed: %d %s", grant.Code, grant.Body.String())
	}
	for _, item := range []struct {
		name, token string
		status      int
	}{
		{"anonymous", "", http.StatusUnauthorized},
		{"malformed", "Bearer invalid", http.StatusUnauthorized},
		{"other tenant principal", "Bearer dev:outsider", http.StatusForbidden},
		{"inventory viewer", "Bearer dev:viewer", http.StatusForbidden},
		{"tenant owner", "Bearer dev:owner", http.StatusCreated},
	} {
		t.Run(item.name, func(t *testing.T) {
			response := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories", item.token, map[string]string{"name": "Home Inventory"})
			if response.Code != item.status {
				t.Fatalf("want %d, got %d: %s", item.status, response.Code, response.Body.String())
			}
		})
	}
}
