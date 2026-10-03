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
}
