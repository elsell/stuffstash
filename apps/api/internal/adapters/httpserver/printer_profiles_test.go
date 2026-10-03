package httpserver

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/adapters/printingprofiles"
	"net/http"
	"testing"
)

func TestPrinterProfilesEnforceInventoryScopeWithoutRegisteredHardware(t *testing.T) {
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	application := newSeededTestApp(t, seededState{tenants: []seedTenant{{id: tenantID, name: "Home", owner: "tenant-owner"}}, inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"}}}).WithPrinterCatalog(printingprofiles.Catalog{})
	server := NewServer(":0", application)
	path := "/tenants/" + tenantID + "/inventories/" + inventoryID + "/printer-profiles"
	for _, test := range []struct {
		name, auth, path string
		status           int
	}{
		{"unauthenticated", "", path, http.StatusUnauthorized},
		{"malformed token", "Bearer malformed", path, http.StatusUnauthorized},
		{"unrelated user", "Bearer dev:other", path, http.StatusForbidden},
		{"cross tenant", "Bearer dev:owner", "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAX/inventories/" + inventoryID + "/printer-profiles", http.StatusNotFound},
		{"owner without printer", "Bearer dev:owner", path, http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := performRequest(server, http.MethodGet, test.path, test.auth, nil)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if response.Code == http.StatusOK {
				var envelope struct {
					Data []struct {
						AdapterID string `json:"adapterId"`
						Media     []struct {
							PresetID    string `json:"presetId"`
							RasterWidth int    `json:"rasterWidth"`
						}
					}
				}
				if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
					t.Fatal(err)
				}
				if len(envelope.Data) != 1 || len(envelope.Data[0].Media) != 1 || envelope.Data[0].Media[0].RasterWidth != 306 {
					t.Fatalf("missing usable printer media: %s", response.Body.String())
				}
			}
		})
	}
}

func coverPrinterProfileScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	t.Helper()
	const tenantID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	const inventoryID = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	application := newSeededTestApp(t, seededState{tenants: []seedTenant{{id: tenantID, name: "Home", owner: "owner"}}, inventories: []seedInventory{{id: inventoryID, tenantID: tenantID, name: "Tools", owner: "owner"}}}).WithPrinterCatalog(printingprofiles.Catalog{})
	token := "Bearer dev:owner"
	status := http.StatusOK
	if adversarial {
		token = "Bearer dev:other"
		status = http.StatusForbidden
	}
	coverage.request(t, NewServer(":0", application), http.MethodGet, "/tenants/{tenantId}/inventories/{inventoryId}/printer-profiles", "/tenants/"+tenantID+"/inventories/"+inventoryID+"/printer-profiles", token, nil, status)
}
