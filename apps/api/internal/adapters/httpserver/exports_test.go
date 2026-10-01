package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestInventoryExportBoundary(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	const otherTenant = "01ARZ3NDEKTSV4RRFFQ69G5FAX"
	const otherInventory = "01ARZ3NDEKTSV4RRFFQ69G5FAY"
	server := NewServer(":0", newSeededTestApp(t, seededState{
		tenants:     []seedTenant{{id: tenantID, name: "Home", owner: "owner"}, {id: otherTenant, name: "Hidden", owner: "other"}},
		inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"}, {id: otherInventory, tenantID: otherTenant, name: "Hidden", owner: "other"}},
	}))
	path := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/export"
	for _, tc := range []struct {
		name, path, token string
		status            int
	}{
		{"missing token", path, "", http.StatusUnauthorized},
		{"malformed token", path, "Bearer invalid", http.StatusUnauthorized},
		{"unauthorized principal", path, "Bearer dev:stranger", http.StatusForbidden},
		{"cross tenant", "/tenants/" + tenantID + "/inventories/" + otherInventory + "/export", "Bearer dev:owner", http.StatusNotFound},
		{"other inventory", "/tenants/" + otherTenant + "/inventories/" + otherInventory + "/export", "Bearer dev:owner", http.StatusForbidden},
		{"invalid format", path + "?format=xml", "Bearer dev:owner", http.StatusUnprocessableEntity},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := performRequest(server, http.MethodGet, tc.path, tc.token, nil)
			if r.Code != tc.status {
				t.Fatalf("status %d wanted %d: %s", r.Code, tc.status, r.Body.String())
			}
			if strings.Contains(r.Body.String(), "Hidden") {
				t.Fatal("hidden inventory leaked")
			}
		})
	}
	created := performRequest(server, http.MethodPost, "/tenants/"+tenantID+"/inventories/"+inventoryID+"/assets", "Bearer dev:owner", map[string]any{"kind": "item", "title": "Camping tent"})
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %s", created.Body.String())
	}
	r := performRequest(server, http.MethodGet, path, "Bearer dev:owner", nil)
	if r.Code != http.StatusOK {
		t.Fatalf("export: %d %s", r.Code, r.Body.String())
	}
	var document struct {
		SchemaVersion int `json:"schemaVersion"`
		Assets        []struct {
			Title string `json:"title"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(r.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document.SchemaVersion != 1 || len(document.Assets) != 1 || document.Assets[0].Title != "Camping tent" {
		t.Fatalf("wrong export: %s", r.Body.String())
	}
	if r.Header().Get("Cache-Control") != "private, no-store" || !strings.Contains(r.Header().Get("Content-Disposition"), "attachment;") {
		t.Fatalf("unsafe download headers: %v", r.Header())
	}
	csv := performRequest(server, http.MethodGet, path+"?format=csv", "Bearer dev:owner", nil)
	if csv.Code != http.StatusOK || !strings.Contains(csv.Body.String(), "Camping tent") {
		t.Fatalf("csv: %d %s", csv.Code, csv.Body.String())
	}
}
