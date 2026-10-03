package httpserver

import (
	"context"
	"net/http"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/printingprofiles"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
)

func TestInventoryPrintSettingsEnforceScopeRolesAndRevision(t *testing.T) {
	coverPrintSettingsScenarios(t, newExecutedScenarioCoverage("printing settings"), false)
}
func coverPrintSettingsScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	application, store, az := labelTestApplication(t)
	application = application.WithPrinterRegistry(store, printingprofiles.Catalog{}).WithPrintSettings(store)
	server := NewServer(":0", application)
	ctx := context.Background()
	if err := az.GrantInventoryViewer(ctx, identity.Principal{ID: "viewer"}, labelTenant, labelInventory); err != nil {
		t.Fatal(err)
	}
	if err := az.GrantInventoryEditor(ctx, identity.Principal{ID: "editor"}, labelTenant, labelInventory); err != nil {
		t.Fatal(err)
	}
	path := labelPrefix + "/print-settings"
	route := "/tenants/{tenantId}/inventories/{inventoryId}/print-settings"
	body := map[string]any{"revision": 0, "defaultPrinterId": nil, "template": map[string]any{"id": "qr-only", "version": 1, "options": map[string]any{"showReference": false}}, "printOnCreateDefault": false}
	if adversarial {
		coverage.request(t, server, "GET", route, path, "", nil, 401)
		coverage.request(t, server, "PUT", route, path, "Bearer dev:viewer", body, 403)
		return
	}
	defaults := coverage.request(t, server, "GET", route, path, "Bearer dev:viewer", nil, 200)
	value := labelResponseData(t, defaults.Body.Bytes())
	if value["revision"] != float64(0) || value["printOnCreateDefault"] != false || value["defaultPrinterId"] != nil {
		t.Fatal("incorrect virtual defaults")
	}
	for _, actor := range []struct {
		token       string
		read, write int
	}{{"", 401, 401}, {"Bearer malformed", 401, 401}, {"Bearer dev:other", 403, 403}, {"Bearer dev:viewer", 200, 403}, {"Bearer dev:editor", 200, 403}} {
		for _, request := range []struct {
			method string
			want   int
			body   any
		}{{"GET", actor.read, nil}, {"PUT", actor.write, body}} {
			response := performRequest(server, request.method, path, actor.token, request.body)
			if response.Code != request.want {
				t.Fatalf("%s %s: %d %s", request.method, actor.token, response.Code, response.Body.String())
			}
		}
	}
	body["printOnCreateDefault"] = true
	if response := performRequest(server, "PUT", path, "Bearer dev:owner", body); response.Code != 400 {
		t.Fatalf("auto without printer: %d %s", response.Code, response.Body.String())
	}
	body["printOnCreateDefault"] = false
	saved := coverage.request(t, server, "PUT", route, path, "Bearer dev:owner", body, 200)
	if labelResponseData(t, saved.Body.Bytes())["revision"] != float64(1) {
		t.Fatal("first replacement revision")
	}
	if response := performRequest(server, "PUT", path, "Bearer dev:owner", body); response.Code != 409 {
		t.Fatal("stale setting write accepted")
	}
	body["revision"] = 1
	printer := performRequestWithHeaders(server, "POST", labelPrefix+"/printers", "Bearer dev:owner", map[string]string{"Idempotency-Key": "settings-printer"}, map[string]any{"name": "Offline Brother", "adapterId": "brother-ql800", "presetId": "brother-ql800-29x90", "presetVersion": 1})
	if printer.Code != 201 {
		t.Fatal(printer.Body.String())
	}
	destination := labelResponseData(t, printer.Body.Bytes())
	body["defaultPrinterId"] = destination["id"]
	body["printOnCreateDefault"] = true
	enabled := performRequest(server, "PUT", path, "Bearer dev:owner", body)
	if enabled.Code != 200 {
		t.Fatalf("offline configured printer rejected: %d %s", enabled.Code, enabled.Body.String())
	}
	body["revision"] = 2
	body["defaultPrinterId"] = nil
	if response := performRequest(server, "PUT", path, "Bearer dev:owner", body); response.Code != 400 {
		t.Fatal("cleared active default destination")
	}
	body["printOnCreateDefault"] = false
	cleared := performRequest(server, "PUT", path, "Bearer dev:owner", body)
	if cleared.Code != 200 {
		t.Fatal(cleared.Body.String())
	}
	selected := labelResponseData(t, cleared.Body.Bytes())["template"].(map[string]any)
	if selected["id"] != "qr-only" || selected["options"].(map[string]any)["showReference"] != false {
		t.Fatal("clearing printer erased independent template")
	}
	body["revision"] = 3
	body["defaultPrinterId"] = destination["id"]
	wrongInventory := "/tenants/" + labelTenant + "/inventories/" + labelOtherInventory + "/print-settings"
	body["revision"] = 0
	if response := performRequest(server, "PUT", wrongInventory, "Bearer dev:other", body); response.Code != 404 {
		t.Fatalf("cross inventory printer: %d %s", response.Code, response.Body.String())
	}
	if response := performRequest(server, "GET", "/tenants/01ARZ3NDEKTSV4RRFFQ69G5FAZ/inventories/"+labelInventory+"/print-settings", "Bearer dev:owner", nil); response.Code != 404 {
		t.Fatal("cross tenant scope accepted")
	}
	body["revision"] = 3
	retired := performRequest(server, "PATCH", labelPrefix+"/printers/"+destination["id"].(string), "Bearer dev:owner", map[string]any{"revision": 1, "retired": true})
	if retired.Code != 200 {
		t.Fatal(retired.Body.String())
	}
	if response := performRequest(server, "PUT", path, "Bearer dev:owner", body); response.Code != 400 {
		t.Fatalf("retired destination accepted: %d", response.Code)
	}
	body["defaultPrinterId"] = nil
	body["template"] = map[string]any{"id": "uploaded-code", "version": 1, "options": map[string]any{"showReference": false}}
	if response := performRequest(server, http.MethodPut, path, "Bearer dev:owner", body); response.Code != 400 {
		t.Fatal("unregistered template accepted")
	}
}
