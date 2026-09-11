package httpserver

import (
	"net/http"
	"testing"
)

func coverNotificationPreferenceScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	t.Helper()
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants:     []seedTenant{{id: "home", name: "Home", owner: "owner"}},
		inventories: []seedInventory{{id: "main", tenantID: "home", name: "Main", owner: "owner"}},
	}))
	const path = "/tenants/home/inventories/main/notification-preferences"
	const template = "/tenants/{tenantId}/inventories/{inventoryId}/notification-preferences"
	created := performRequest(server, http.MethodPost, "/tenants/home/inventories/main/custom-asset-types", "Bearer dev:owner", map[string]any{"key": "food", "displayName": "Food", "expirationEnabled": true})
	requireStatus(t, created, http.StatusCreated)
	typeID := decodeCustomAssetType(t, created).Data.ID
	token, status := "Bearer dev:owner", http.StatusOK
	if adversarial {
		token, status = "Bearer dev:outsider", http.StatusForbidden
	}
	coverage.request(t, server, http.MethodPost, template+"/initialize", path+"/initialize", token, map[string]any{"timezone": "UTC"}, status)
	coverage.request(t, server, http.MethodGet, template, path, token, nil, status)
	settings := map[string]any{"enabled": true, "upcoming": true, "expired": true, "advanceDays": 7}
	coverage.request(t, server, http.MethodPut, template, path, token, map[string]any{"revision": 1, "defaults": settings, "timezone": "UTC", "pushEnabled": false}, status)
	coverage.request(t, server, http.MethodPut, template+"/types/{customAssetTypeId}", path+"/types/"+typeID, token, map[string]any{"revision": 2, "settings": settings}, status)
	coverage.request(t, server, http.MethodDelete, template+"/types/{customAssetTypeId}", path+"/types/"+typeID+"?revision=3", token, nil, status)
}
