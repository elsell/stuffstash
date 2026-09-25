package httpserver

import (
	"net/http"
	"testing"
)

func TestOnboardingInventoryCreationPreservesTenantAuthorization(t *testing.T) {
	const tenantID = "tenant-home"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "owner"}},
		inventories: []seedInventory{{id: "inventory-shared", tenantID: tenantID, name: "Shared", owner: "inventory-owner"}},
	}))
	grant := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/inventory-shared/access-grants", "Bearer dev:inventory-owner", map[string]string{"principalId": "viewer", "relationship": "viewer"})
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
		{"inventory owner is not household owner", "Bearer dev:inventory-owner", http.StatusForbidden},
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

// The mobile switcher must not infer household creation rights from an inventory role.
func TestMobileSwitcherCreationRejectsForgedHouseholdAuthority(t *testing.T) {
	const home = "tenant-home"
	const elsewhere = "tenant-other"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants:     []seedTenant{{id: home, name: "Home", owner: "owner"}, {id: elsewhere, name: "Other", owner: "other-owner"}},
		inventories: []seedInventory{{id: "shared", tenantID: elsewhere, name: "Shared", owner: "owner"}},
	}))
	for _, token := range []string{"", "Bearer invalid", "Bearer dev:owner"} {
		response := performRequestWithHeaders(server, http.MethodPost, "/tenants/"+elsewhere+"/inventories", token,
			map[string]string{"X-Tenant-ID": home, "X-Role": "owner"}, map[string]string{"name": "Forged"})
		want := http.StatusForbidden
		if token != "Bearer dev:owner" {
			want = http.StatusUnauthorized
		}
		if response.Code != want {
			t.Fatalf("creation: want %d got %d: %s", want, response.Code, response.Body.String())
		}
	}
	list := performRequest(server, http.MethodGet, "/tenants/"+elsewhere+"/inventories", "Bearer dev:other-owner", nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list failed: %d", list.Code)
	}
	assertInventories(t, decodeInventoryList(t, list), expectedInventory{id: "shared", tenantID: elsewhere, name: "Shared"})
}

func TestMobileSwitcherHouseholdCreationBelongsToAuthenticatedPrincipal(t *testing.T) {
	const home = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	server := NewServer(":0", newTestApp(&fakeObserver{}, home, "audit", "event", "claim"))
	for _, token := range []string{"", "Bearer invalid"} {
		response := performRequest(server, http.MethodPost, "/tenants", token, map[string]string{"name": "Home"})
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", response.Code)
		}
	}
	created := performRequest(server, http.MethodPost, "/tenants", "Bearer dev:creator", map[string]string{"name": "Home"})
	if created.Code != http.StatusCreated {
		t.Fatalf("create failed: %d %s", created.Code, created.Body.String())
	}
	denied := performRequest(server, http.MethodPost, "/tenants/"+home+"/inventories", "Bearer dev:outsider", map[string]string{"name": "Main"})
	if denied.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", denied.Code)
	}
	allowed := performRequest(server, http.MethodPost, "/tenants/"+home+"/inventories", "Bearer dev:creator", map[string]string{"name": "Main"})
	if allowed.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", allowed.Code, allowed.Body.String())
	}
}
