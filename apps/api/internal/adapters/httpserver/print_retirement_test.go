package httpserver

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/adapters/printingprofiles"
	printingapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"testing"
	"time"
)

func TestPrinterRetirementCancelsUnstartedJobsAndClearsDefaults(t *testing.T) {
	clock := &labelTestClock{now: time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)}
	application, store, az := labelTestApplication(t, clock)
	application = application.WithPrinterRegistry(store, printingprofiles.Catalog{}).WithPrintSettings(store).WithPrintJobs(store, printingapp.JobConfig{MaxCopies: 20, MaxArtifactBytes: 1000000, ArtifactTTL: 24 * time.Hour, TerminalTTL: time.Hour})
	server := NewServer(":0", application)
	if err := az.GrantInventoryViewer(context.Background(), identity.Principal{ID: "viewer"}, labelTenant, labelInventory); err != nil {
		t.Fatal(err)
	}
	pResponse := performRequestWithHeaders(server, "POST", labelPrefix+"/printers", "Bearer dev:owner", map[string]string{"Idempotency-Key": "retire"}, map[string]any{"name": "Garage", "adapterId": "brother-ql800", "presetId": "brother-ql800-29x90", "presetVersion": 1})
	if pResponse.Code != 201 {
		t.Fatal(pResponse.Body.String())
	}
	p := labelResponseData(t, pResponse.Body.Bytes())
	settings := performRequest(server, "PUT", labelPrefix+"/print-settings", "Bearer dev:owner", map[string]any{"revision": 0, "defaultPrinterId": p["id"], "printOnCreateDefault": true, "template": map[string]any{"id": "qr-title", "version": 1, "options": map[string]any{"showReference": true}}})
	if settings.Code != 200 {
		t.Fatal(settings.Body.String())
	}
	jobResponse := performRequestWithHeaders(server, "POST", labelPrefix+"/printers/"+p["id"].(string)+"/test-jobs", "Bearer dev:owner", map[string]string{"Idempotency-Key": "test"}, map[string]any{"printerId": p["id"], "expectedMediaFingerprint": p["mediaFingerprint"], "templateId": "qr-title", "templateVersion": 1, "templateOptions": map[string]any{"showReference": true}, "copies": 1})
	if jobResponse.Code != 201 {
		t.Fatal(jobResponse.Body.String())
	}
	job := labelResponseData(t, jobResponse.Body.Bytes())
	path := labelPrefix + "/printers/" + p["id"].(string)
	body := map[string]any{"revision": 1, "retired": true}
	for _, c := range []struct {
		token string
		code  int
	}{{"", 401}, {"Bearer dev:viewer", 403}, {"Bearer dev:other", 403}} {
		response := performRequest(server, "PATCH", path, c.token, body)
		if response.Code != c.code {
			t.Fatalf("retire authorization: %d %s", response.Code, response.Body.String())
		}
	}
	retired := performRequest(server, "PATCH", path, "Bearer dev:owner", body)
	if retired.Code != 200 {
		t.Fatal(retired.Body.String())
	}
	current := performRequest(server, "GET", labelPrefix+"/print-jobs/"+job["id"].(string), "Bearer dev:owner", nil)
	if current.Code != 200 || labelResponseData(t, current.Body.Bytes())["status"] != "canceled" {
		t.Fatalf("retirement left unstarted work: %s", current.Body.String())
	}
	settings = performRequest(server, "GET", labelPrefix+"/print-settings", "Bearer dev:owner", nil)
	state := labelResponseData(t, settings.Body.Bytes())
	if state["revision"] != float64(2) || state["defaultPrinterId"] != nil || state["printOnCreateDefault"] != false {
		t.Fatalf("retirement retained default: %s", settings.Body.String())
	}
}
