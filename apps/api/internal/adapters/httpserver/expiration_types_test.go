package httpserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestExpirationTypeCapabilityIsScopedAndEditable(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const otherInventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "tenant-owner"}},
		inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Home", owner: "owner"}, {id: otherInventoryID, tenantID: tenantID, name: "Other", owner: "other-owner"}},
		ids:         []string{"01ARZ3NDEKTSV4RRFFQ69G5FAY", "audit-create-type"},
	}))
	base := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/custom-asset-types"
	response := performRequest(server, http.MethodPost, base, "Bearer dev:owner", map[string]any{"key": "medicine", "displayName": "Medicine", "expirationEnabled": true})
	if response.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", response.Code, response.Body.String())
	}
	var created struct {
		Data struct {
			ID      string `json:"id"`
			Enabled bool   `json:"expirationEnabled"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if !created.Data.Enabled {
		t.Fatal("expiration capability was discarded")
	}
	endpoint := base + "/" + created.Data.ID
	for _, auth := range []string{"", "Bearer malformed", "Bearer dev:outsider", "Bearer dev:other-owner"} {
		denied := performRequest(server, http.MethodPatch, endpoint, auth, map[string]any{"expirationEnabled": false})
		if denied.Code != http.StatusUnauthorized && denied.Code != http.StatusForbidden {
			t.Fatalf("unauthorized capability change: %d", denied.Code)
		}
	}
	wrongScope := "/tenants/" + tenantID + "/inventories/" + otherInventoryID + "/custom-asset-types/" + created.Data.ID
	denied := performRequest(server, http.MethodPatch, wrongScope, "Bearer dev:other-owner", map[string]any{"expirationEnabled": false})
	if denied.Code != http.StatusNotFound && denied.Code != http.StatusForbidden {
		t.Fatalf("cross-inventory change: %d", denied.Code)
	}
	for _, enabled := range []bool{false, true} {
		result := performRequest(server, http.MethodPatch, endpoint, "Bearer dev:owner", map[string]any{"expirationEnabled": enabled})
		if result.Code != http.StatusOK {
			t.Fatalf("update: %d %s", result.Code, result.Body.String())
		}
		read := performRequest(server, http.MethodGet, endpoint, "Bearer dev:owner", nil)
		if read.Code != http.StatusOK {
			t.Fatalf("read: %d", read.Code)
		}
		if err := json.Unmarshal(read.Body.Bytes(), &created); err != nil {
			t.Fatal(err)
		}
		if created.Data.Enabled != enabled {
			t.Fatal("capability did not persist")
		}
	}
}
