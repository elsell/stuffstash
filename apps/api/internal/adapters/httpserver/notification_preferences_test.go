package httpserver

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"net/http"
	"testing"
)

func TestPersonalNotificationPreferencesLifecycleAndIsolation(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const hiddenID = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	authorizer := memory.NewAuthorizer()
	application := newSeededTestAppWithAuthorizer(t, seededState{
		tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "household-owner"}},
		inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Home", owner: "owner"}, {id: hiddenID, tenantID: tenantID, name: "Private", owner: "other"}},
	}, authorizer)
	if err := authorizer.GrantInventoryViewer(context.Background(), identity.Principal{ID: "viewer"}, tenant.ID(tenantID), inventory.InventoryID(inventoryID)); err != nil {
		t.Fatal(err)
	}
	server := NewServer(":0", application)
	base := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/notification-preferences"
	for _, auth := range []string{"", "Bearer malformed", "Bearer dev:outsider"} {
		denied := performRequest(server, http.MethodPost, base+"/initialize", auth, map[string]any{"timezone": "UTC"})
		if denied.Code != http.StatusUnauthorized && denied.Code != http.StatusForbidden {
			t.Fatalf("unauthorized initialize %d", denied.Code)
		}
	}
	initialized := performRequest(server, http.MethodPost, base+"/initialize", "Bearer dev:owner", map[string]any{"timezone": "America/New_York"})
	if initialized.Code != http.StatusOK {
		t.Fatalf("initialize %d %s", initialized.Code, initialized.Body.String())
	}
	var record struct {
		Data struct {
			Revision int    `json:"revision"`
			Timezone string `json:"timezone"`
			Defaults struct {
				Enabled     bool `json:"enabled"`
				AdvanceDays int  `json:"advanceDays"`
			} `json:"defaults"`
		} `json:"data"`
	}
	if err := json.Unmarshal(initialized.Body.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record.Data.Revision != 1 || record.Data.Timezone != "America/New_York" || !record.Data.Defaults.Enabled || record.Data.Defaults.AdvanceDays != 30 {
		t.Fatalf("defaults %+v", record.Data)
	}
	defaults := map[string]any{"enabled": false, "upcoming": true, "expired": false, "advanceDays": 60}
	updated := performRequest(server, http.MethodPut, base, "Bearer dev:owner", map[string]any{"revision": 1, "defaults": defaults, "timezone": "UTC", "pushEnabled": false})
	if updated.Code != http.StatusOK {
		t.Fatalf("update %d %s", updated.Code, updated.Body.String())
	}
	stale := performRequest(server, http.MethodPut, base, "Bearer dev:owner", map[string]any{"revision": 1, "defaults": defaults, "timezone": "UTC", "pushEnabled": true})
	if stale.Code != http.StatusConflict {
		t.Fatalf("stale settings overwrite %d %s", stale.Code, stale.Body.String())
	}
	repeated := performRequest(server, http.MethodPost, base+"/initialize", "Bearer dev:owner", map[string]any{"timezone": "Europe/London"})
	if repeated.Code != http.StatusOK {
		t.Fatal(repeated.Body.String())
	}
	if err := json.Unmarshal(repeated.Body.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record.Data.Revision != 2 || record.Data.Timezone != "UTC" || record.Data.Defaults.Enabled {
		t.Fatal("initialize overwrote personal preferences")
	}
	denied := performRequest(server, http.MethodGet, "/tenants/"+tenantID+"/inventories/"+hiddenID+"/notification-preferences", "Bearer dev:owner", nil)
	if denied.Code != http.StatusForbidden {
		t.Fatalf("cross inventory read %d", denied.Code)
	}
	spoof := performRequest(server, http.MethodPut, base, "Bearer dev:owner", map[string]any{"revision": 2, "defaults": defaults, "timezone": "UTC", "pushEnabled": false, "principalId": "other"})
	if spoof.Code != http.StatusUnprocessableEntity {
		t.Fatalf("recipient spoof accepted %d", spoof.Code)
	}
	createdType := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/custom-asset-types", "Bearer dev:owner", map[string]any{"key": "medicine", "displayName": "Medicine", "expirationEnabled": true})
	if createdType.Code != http.StatusCreated {
		t.Fatal(createdType.Body.String())
	}
	typeID := decodeCustomAssetType(t, createdType).Data.ID
	override := map[string]any{"enabled": true, "upcoming": true, "expired": true, "advanceDays": 7}
	set := performRequest(server, http.MethodPut, base+"/types/"+typeID, "Bearer dev:owner", map[string]any{"revision": 2, "settings": override})
	if set.Code != http.StatusOK {
		t.Fatalf("override %d %s", set.Code, set.Body.String())
	}
	var withOverrides struct {
		Data struct {
			Revision  int `json:"revision"`
			Overrides []struct {
				CustomAssetTypeID string `json:"customAssetTypeId"`
				Settings          struct {
					Enabled bool `json:"enabled"`
				} `json:"settings"`
			} `json:"overrides"`
		} `json:"data"`
	}
	if err := json.Unmarshal(set.Body.Bytes(), &withOverrides); err != nil {
		t.Fatal(err)
	}
	if withOverrides.Data.Revision != 3 || len(withOverrides.Data.Overrides) != 1 || !withOverrides.Data.Overrides[0].Settings.Enabled {
		t.Fatal("type did not override disabled inventory default")
	}
	reset := performRequest(server, http.MethodDelete, base+"/types/"+typeID+"?revision=3", "Bearer dev:owner", nil)
	if reset.Code != http.StatusOK {
		t.Fatalf("reset %d %s", reset.Code, reset.Body.String())
	}
	if err := json.Unmarshal(reset.Body.Bytes(), &withOverrides); err != nil {
		t.Fatal(err)
	}
	if withOverrides.Data.Revision != 4 || len(withOverrides.Data.Overrides) != 0 {
		t.Fatal("reset did not restore inheritance")
	}
	viewer := performRequest(server, http.MethodPost, base+"/initialize", "Bearer dev:viewer", map[string]any{"timezone": "UTC"})
	if viewer.Code != http.StatusOK {
		t.Fatalf("viewer personal settings %d %s", viewer.Code, viewer.Body.String())
	}
	if err := json.Unmarshal(viewer.Body.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record.Data.Revision != 1 || !record.Data.Defaults.Enabled {
		t.Fatal("settings leaked between users")
	}

	crossTenant := performRequest(server, http.MethodGet, "/tenants/other/inventories/"+inventoryID+"/notification-preferences", "Bearer dev:owner", nil)
	if crossTenant.Code != http.StatusNotFound {
		t.Fatalf("cross tenant read %d %s", crossTenant.Code, crossTenant.Body.String())
	}
	missingType := performRequest(server, http.MethodPut, base+"/types/missing", "Bearer dev:owner", map[string]any{"revision": 4, "settings": override})
	if missingType.Code != http.StatusNotFound {
		t.Fatalf("missing type override %d", missingType.Code)
	}
	invalidTimezone := performRequest(server, http.MethodPut, base, "Bearer dev:owner", map[string]any{"revision": 4, "defaults": defaults, "timezone": "Local", "pushEnabled": false})
	if invalidTimezone.Code != http.StatusBadRequest {
		t.Fatalf("invalid timezone %d", invalidTimezone.Code)
	}

}
