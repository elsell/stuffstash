package homebox

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestLegacyHomeboxCSVBuildsNormalizedPlan(t *testing.T) {
	csv := `HB.import_ref,HB.location,HB.tags,HB.asset_id,HB.archived,HB.url,HB.name,HB.quantity,HB.description,HB.insured,HB.notes,HB.purchase_price,HB.purchase_from,HB.purchase_time,HB.manufacturer,HB.model_number,HB.serial_number,HB.lifetime_warranty,HB.warranty_expires,HB.warranty_details,HB.sold_to,HB.sold_price,HB.sold_time,HB.sold_notes
,Bin 8,"Storage; Clothing",000-001,false,/item/one,Plastic Bags,30,,false,kept folded,0,,,,,,false,,,,0,,
,Bin 8,tools,000-002,false,/item/two,Bike Tool,2,Blue handle,true,,12.5,Store,0001-11-08,Park,MT-1,SN-1,true,0001-12-01,details,,0,,`

	plan, err := NewLegacyImporter(nil).ReadImportPlan(context.Background(), ports.ImportSourceRequest{
		SourceType: importplan.SourceLegacyHomeboxCSV,
		FileName:   "export.csv",
		Content:    []byte(csv),
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Source.Type != importplan.SourceLegacyHomeboxCSV {
		t.Fatalf("source type = %q", plan.Source.Type)
	}
	if got := plan.Counts().Locations; got != 1 {
		t.Fatalf("locations = %d", got)
	}
	if got := plan.Counts().Assets; got != 2 {
		t.Fatalf("assets = %d", got)
	}
	if got := plan.Counts().Fields; got == 0 {
		t.Fatal("expected field definitions")
	}
	item := plan.Assets[2]
	if item.Title != "Bike Tool" {
		t.Fatalf("item title = %q", item.Title)
	}
	if item.ParentSourceID != "location:Bin 8" {
		t.Fatalf("parent source id = %q", item.ParentSourceID)
	}
	if item.CustomFields["homebox-quantity"] != float64(2) {
		t.Fatalf("quantity field = %#v", item.CustomFields["homebox-quantity"])
	}
	if item.CustomFields["homebox-insured"] != true {
		t.Fatalf("insured field = %#v", item.CustomFields["homebox-insured"])
	}
	if plan.Assets[0].CustomFields["homebox-source-id"] != "location:Bin 8" {
		t.Fatalf("location source field = %#v", plan.Assets[0].CustomFields["homebox-source-id"])
	}
	if item.CustomFields["homebox-source-id"] != "000-002" {
		t.Fatalf("item source field = %#v", item.CustomFields["homebox-source-id"])
	}
	if plan.Counts().Warnings != 2 {
		t.Fatalf("warnings = %d, messages = %#v", plan.Counts().Warnings, plan.Messages)
	}
}

func TestLegacyHomeboxRejectsBlockedOutboundAddress(t *testing.T) {
	_, err := NewLegacyImporter(nil).ReadImportPlan(context.Background(), ports.ImportSourceRequest{
		SourceType: importplan.SourceLegacyHomebox,
		BaseURL:    "http://127.0.0.1:7744",
		Username:   "user@example.com",
		Password:   "secret",
	})
	if err == nil {
		t.Fatal("expected blocked outbound address error")
	}
}

func TestLegacyHomeboxAllowsPrivateNetworkWhenExplicit(t *testing.T) {
	server := newLegacyHomeboxTestServer(t)

	plan, err := NewLegacyImporter(nil).ReadImportPlan(context.Background(), ports.ImportSourceRequest{
		SourceType:          importplan.SourceLegacyHomebox,
		BaseURL:             server.URL,
		Username:            "user@example.com",
		Password:            "secret",
		AllowPrivateNetwork: true,
	})
	if err != nil {
		t.Fatalf("read private-network Homebox source: %v", err)
	}
	if plan.Source.Version != "v0.test" || plan.Counts().Assets != 1 || plan.Counts().Locations != 1 {
		t.Fatalf("unexpected plan: source=%+v counts=%+v", plan.Source, plan.Counts())
	}
	if plan.Assets[1].CustomFields["homebox-source-id"] != "item-one" {
		t.Fatalf("live item source id = %#v", plan.Assets[1].CustomFields["homebox-source-id"])
	}
}

func TestLegacyHomeboxWarningDetailsAreSanitized(t *testing.T) {
	server := newLegacyHomeboxTestServerWithItemDetailStatus(t, http.StatusInternalServerError)

	plan, err := NewLegacyImporter(nil).ReadImportPlan(context.Background(), ports.ImportSourceRequest{
		SourceType:          importplan.SourceLegacyHomebox,
		BaseURL:             server.URL,
		Username:            "user@example.com",
		Password:            "secret",
		AllowPrivateNetwork: true,
	})
	if err != nil {
		t.Fatalf("read Homebox source: %v", err)
	}
	if len(plan.Messages) != 1 {
		t.Fatalf("messages = %#v", plan.Messages)
	}
	if plan.Messages[0].Detail != "item detail could not be read" {
		t.Fatalf("warning detail leaked raw error: %#v", plan.Messages[0])
	}
}

func TestLegacyHomeboxRejectsPrivateRedirectWithoutOptIn(t *testing.T) {
	target := newLegacyHomeboxTestServer(t)
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+r.URL.Path, http.StatusFound)
	}))
	t.Cleanup(redirect.Close)

	_, err := NewLegacyImporter(nil).ReadImportPlan(context.Background(), ports.ImportSourceRequest{
		SourceType: importplan.SourceLegacyHomebox,
		BaseURL:    redirect.URL,
		Username:   "user@example.com",
		Password:   "secret",
	})
	if err == nil {
		t.Fatal("expected private redirect to be blocked")
	}
}

func newLegacyHomeboxTestServer(t *testing.T) *httptest.Server {
	return newLegacyHomeboxTestServerWithItemDetailStatus(t, http.StatusOK)
}

func newLegacyHomeboxTestServerWithItemDetailStatus(t *testing.T, itemDetailStatus int) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/users/login":
			_, _ = w.Write([]byte(`{"token":"token-one"}`))
		case "/api/v1/status":
			_, _ = w.Write([]byte(`{"build":{"version":"v0.test"}}`))
		case "/api/v1/locations":
			_, _ = w.Write([]byte(`[{"id":"location-one","name":"Garage","description":"Main garage"}]`))
		case "/api/v1/locations/tree":
			_, _ = w.Write([]byte(`[{"id":"location-one","name":"Garage","type":"location","children":[]}]`))
		case "/api/v1/items":
			_, _ = w.Write([]byte(`{"items":[{"id":"item-one","assetId":"HB-1","name":"Drill"}]}`))
		case "/api/v1/items/item-one":
			if itemDetailStatus != http.StatusOK {
				http.Error(w, "database password leaked", itemDetailStatus)
				return
			}
			_, _ = w.Write([]byte(`{"id":"item-one","assetId":"HB-1","name":"Drill","description":"Cordless","quantity":1,"location":{"id":"location-one","name":"Garage"},"attachments":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func TestDecodeCSVBase64(t *testing.T) {
	decoded, err := DecodeCSVBase64(base64.StdEncoding.EncodeToString([]byte("hello")))
	if err != nil {
		t.Fatal(err)
	}
	if string(decoded) != "hello" {
		t.Fatalf("decoded = %q", string(decoded))
	}
}
