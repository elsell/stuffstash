package httpserver

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/adapters/printingprofiles"
	printingapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"net/http"
	"testing"
	"time"
)

func TestPrintJobsEnforceRolesScopeImmutableRetriesAndCancellation(t *testing.T) {
	coverPrintJobScenarios(t, newExecutedScenarioCoverage("printing"), false)
}

func coverPrintJobScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	application, store, az := labelTestApplication(t)
	application = application.WithPrinterRegistry(store, printingprofiles.Catalog{}).WithPrintJobs(store, printingapp.JobConfig{MaxCopies: 20, MaxArtifactBytes: 1000000, ArtifactTTL: time.Hour})
	server := NewServer(":0", application)
	if err := az.GrantInventoryViewer(context.Background(), identity.Principal{ID: "viewer"}, labelTenant, labelInventory); err != nil {
		t.Fatal(err)
	}
	printer := performRequestWithHeaders(server, "POST", labelPrefix+"/printers", "Bearer dev:owner", map[string]string{"Idempotency-Key": "printer"}, map[string]any{"name": "Garage", "adapterId": "brother-ql800", "presetId": "brother-ql800-29x90", "presetVersion": 1})
	if printer.Code != 201 {
		t.Fatal(printer.Body.String())
	}
	p := labelResponseData(t, printer.Body.Bytes())
	created := performRequest(server, "POST", labelPrefix+"/assets", "Bearer dev:owner", map[string]any{"title": "Tools", "kind": "container"})
	if created.Code != 201 {
		t.Fatal(created.Body.String())
	}
	assetID := decodeAsset(t, created).Data.ID
	path := labelPrefix + "/assets/" + assetID + "/print-jobs"
	body := map[string]any{"printerId": p["id"], "expectedMediaFingerprint": p["mediaFingerprint"], "templateId": "qr-title", "templateVersion": 1, "templateOptions": map[string]any{"showReference": true}, "copies": 1}
	headers := map[string]string{"Idempotency-Key": "request"}
	for _, c := range []struct {
		token  string
		status int
	}{{"", 401}, {"Bearer malformed", 401}, {"Bearer dev:other", 403}, {"Bearer dev:viewer", 403}} {
		r := performRequestWithHeaders(server, "POST", path, c.token, headers, body)
		if r.Code != c.status {
			t.Fatalf("unauthorized: %d %s", r.Code, r.Body.String())
		}
	}
	const template = "/tenants/{tenantId}/inventories/{inventoryId}"
	coverage.operation["POST "+template+"/assets/{assetId}/print-jobs"] = struct{}{}
	first := performRequestWithHeaders(server, "POST", path, "Bearer dev:owner", headers, body)
	if first.Code != 201 {
		t.Fatalf("enqueue: %d %s", first.Code, first.Body.String())
	}
	job := labelResponseData(t, first.Body.Bytes())
	replay := performRequestWithHeaders(server, "POST", path, "Bearer dev:owner", headers, body)
	if replay.Code != 200 || labelResponseData(t, replay.Body.Bytes())["id"] != job["id"] {
		t.Fatalf("replay: %d %s", replay.Code, replay.Body.String())
	}
	body["copies"] = 2
	conflict := performRequestWithHeaders(server, "POST", path, "Bearer dev:owner", headers, body)
	if conflict.Code != 409 {
		t.Fatalf("changed request: %d", conflict.Code)
	}
	detail := labelPrefix + "/print-jobs/" + job["id"].(string)
	token, readStatus := "Bearer dev:viewer", 200
	if adversarial {
		token, readStatus = "Bearer dev:other", 403
	}
	coverage.request(t, server, "GET", template+"/print-jobs/{jobId}", detail, token, nil, readStatus)
	coverage.request(t, server, "GET", template+"/print-jobs", labelPrefix+"/print-jobs", token, nil, readStatus)

	other := performRequest(server, "GET", "/tenants/"+labelTenant+"/inventories/"+labelOtherInventory+"/print-jobs/"+job["id"].(string), "Bearer dev:other", nil)
	if other.Code != 404 {
		t.Fatalf("real other inventory: %d", other.Code)
	}
	coverage.operation["POST "+template+"/print-jobs/{jobId}/cancellation"] = struct{}{}
	for _, c := range []struct {
		token  string
		status int
	}{{"Bearer dev:viewer", 403}, {"Bearer dev:owner", 200}} {
		r := performRequest(server, http.MethodPost, detail+"/cancellation", c.token, map[string]any{"revision": job["revision"]})
		if r.Code != c.status {
			t.Fatalf("cancel: %d %s", r.Code, r.Body.String())
		}
	}
	testPath := labelPrefix + "/printers/" + p["id"].(string) + "/test-jobs"
	reprintPath := detail + "/reprints"
	body["copies"] = 1
	for _, endpoint := range []struct{ route, operation string }{{testPath, template + "/printers/{printerId}/test-jobs"}, {reprintPath, template + "/print-jobs/{jobId}/reprints"}} {
		for _, denied := range []struct {
			token  string
			status int
		}{{"", 401}, {"Bearer malformed", 401}, {"Bearer dev:viewer", 403}, {"Bearer dev:other", 403}} {
			response := performRequestWithHeaders(server, "POST", endpoint.route, denied.token, map[string]string{"Idempotency-Key": "explicit-new-job"}, body)
			if response.Code != denied.status {
				t.Fatalf("new command denied: %d %s", response.Code, response.Body.String())
			}
		}
		coverage.operation["POST "+endpoint.operation] = struct{}{}
	}
	testHeaders := map[string]string{"Idempotency-Key": "diagnostic"}
	diagnostic := performRequestWithHeaders(server, "POST", testPath, "Bearer dev:owner", testHeaders, body)
	if diagnostic.Code != 201 {
		t.Fatalf("diagnostic: %d %s", diagnostic.Code, diagnostic.Body.String())
	}
	diagnosticJob := labelResponseData(t, diagnostic.Body.Bytes())
	if diagnosticJob["kind"] != "printer_test" || diagnosticJob["assetId"] != nil {
		t.Fatal(diagnosticJob)
	}
	repeatDiagnostic := performRequestWithHeaders(server, "POST", testPath, "Bearer dev:owner", testHeaders, body)
	if repeatDiagnostic.Code != 200 || labelResponseData(t, repeatDiagnostic.Body.Bytes())["id"] != diagnosticJob["id"] {
		t.Fatal(repeatDiagnostic.Body.String())
	}
	body["copies"] = 2
	invalidDiagnostic := performRequestWithHeaders(server, "POST", testPath, "Bearer dev:owner", map[string]string{"Idempotency-Key": "invalid-test"}, body)
	if invalidDiagnostic.Code != 400 {
		t.Fatalf("multiple diagnostic copies: %d", invalidDiagnostic.Code)
	}
	body["copies"] = 1
	reprintHeaders := map[string]string{"Idempotency-Key": "explicit-reprint"}
	reprint := performRequestWithHeaders(server, "POST", reprintPath, "Bearer dev:owner", reprintHeaders, body)
	if reprint.Code != 201 {
		t.Fatalf("reprint: %d %s", reprint.Code, reprint.Body.String())
	}
	reprintJob := labelResponseData(t, reprint.Body.Bytes())
	if reprintJob["predecessor"] != job["id"] || reprintJob["id"] == job["id"] {
		t.Fatal(reprintJob)
	}
	repeatReprint := performRequestWithHeaders(server, "POST", reprintPath, "Bearer dev:owner", reprintHeaders, body)
	if repeatReprint.Code != 200 || labelResponseData(t, repeatReprint.Body.Bytes())["id"] != reprintJob["id"] {
		t.Fatal(repeatReprint.Body.String())
	}
	activeReprint := performRequestWithHeaders(server, "POST", labelPrefix+"/print-jobs/"+reprintJob["id"].(string)+"/reprints", "Bearer dev:owner", map[string]string{"Idempotency-Key": "active-reprint"}, body)
	if activeReprint.Code != 409 {
		t.Fatalf("active reprint: %d", activeReprint.Code)
	}
	crossScope := performRequestWithHeaders(server, "POST", "/tenants/"+labelTenant+"/inventories/"+labelOtherInventory+"/print-jobs/"+job["id"].(string)+"/reprints", "Bearer dev:other", reprintHeaders, body)
	if crossScope.Code != 404 {
		t.Fatalf("cross-scope predecessor: %d", crossScope.Code)
	}

}
