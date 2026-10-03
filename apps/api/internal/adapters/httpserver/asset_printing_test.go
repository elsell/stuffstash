package httpserver

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/adapters/printingprofiles"
	printingapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"strings"
	"testing"
	"time"
)

func TestAssetCreationWithPrintIsExplicitScopedAndIdempotent(t *testing.T) {
	application, store, az := labelTestApplication(t)
	application = application.WithPrinterRegistry(store, printingprofiles.Catalog{}).WithPrintJobs(store, printingapp.JobConfig{MaxCopies: 20, MaxArtifactBytes: 1000000, ArtifactTTL: time.Hour}).WithAssetPrintUnitOfWork(store)
	server := NewServer(":0", application)
	printer := performRequestWithHeaders(server, "POST", labelPrefix+"/printers", "Bearer dev:owner", map[string]string{"Idempotency-Key": "printer"}, map[string]any{"name": "Garage", "adapterId": "brother-ql800", "presetId": "brother-ql800-29x90", "presetVersion": 1})
	requireStatus(t, printer, 201)
	p := labelResponseData(t, printer.Body.Bytes())
	if err := az.GrantInventoryViewer(context.Background(), identity.Principal{ID: "viewer"}, labelTenant, labelInventory); err != nil {
		t.Fatal(err)
	}
	body := map[string]any{"title": "New tools", "kind": "item", "printLabel": map[string]any{"printerId": p["id"], "expectedMediaFingerprint": p["mediaFingerprint"], "templateId": "qr-title", "templateVersion": 1, "templateOptions": map[string]any{"showReference": true}, "copies": 1}}
	headers := map[string]string{"Idempotency-Key": "create-and-print"}
	for _, c := range []struct {
		token  string
		status int
	}{{"", 401}, {"Bearer dev:viewer", 403}, {"Bearer dev:other", 403}} {
		requireStatus(t, performRequestWithHeaders(server, "POST", labelPrefix+"/assets", c.token, headers, body), c.status)
	}
	requireStatus(t, performRequest(server, "POST", labelPrefix+"/assets", "Bearer dev:owner", body), 400)
	selection := body["printLabel"].(map[string]any)
	selection["templateId"] = "unknown-template"
	requireStatus(t, performRequestWithHeaders(server, "POST", labelPrefix+"/assets", "Bearer dev:owner", headers, body), 400)
	selection["templateId"] = "qr-title"
	listing := performRequest(server, "GET", labelPrefix+"/assets", "Bearer dev:owner", nil)
	requireStatus(t, listing, 200)
	if !strings.Contains(listing.Body.String(), `"data":[]`) {
		t.Fatal("failed rendering saved asset", listing.Body.String())
	}

	created := performRequestWithHeaders(server, "POST", labelPrefix+"/assets", "Bearer dev:owner", headers, body)
	requireStatus(t, created, 201)
	first := labelResponseData(t, created.Body.Bytes())
	if first["printJobId"] == nil || first["printJobId"] == "" {
		t.Fatal("no committed print job", created.Body.String())
	}
	repeated := performRequestWithHeaders(server, "POST", labelPrefix+"/assets", "Bearer dev:owner", headers, body)
	requireStatus(t, repeated, 200)
	again := labelResponseData(t, repeated.Body.Bytes())
	if again["id"] != first["id"] || again["printJobId"] != first["printJobId"] {
		t.Fatal("duplicate create")
	}
	body["title"] = "Changed payload"
	requireStatus(t, performRequestWithHeaders(server, "POST", labelPrefix+"/assets", "Bearer dev:owner", headers, body), 409)
}
