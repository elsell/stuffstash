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
	if got := plan.Counts().Tags; got != 3 {
		t.Fatalf("tags = %d, tags = %#v", got, plan.Tags)
	}
	if _, ok := plan.Assets[1].CustomFields["homebox-tags"]; ok {
		t.Fatalf("legacy tags custom field should not be created: %#v", plan.Assets[1].CustomFields)
	}
	if got := plan.Assets[1].TagKeys; len(got) != 2 || got[0] != "clothing" || got[1] != "storage" {
		t.Fatalf("first item tag keys = %#v", got)
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

func TestLegacyHomeboxCSVPrefersLabelsOverTags(t *testing.T) {
	csv := `HB.location,HB.labels,HB.tags,HB.asset_id,HB.name
Shelf,"Favorite, Fragile",ignored,000-001,Vase`

	plan, err := NewLegacyImporter(nil).ReadImportPlan(context.Background(), ports.ImportSourceRequest{
		SourceType: importplan.SourceLegacyHomeboxCSV,
		FileName:   "export.csv",
		Content:    []byte(csv),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := plan.Assets[1].TagKeys; len(got) != 2 || got[0] != "favorite" || got[1] != "fragile" {
		t.Fatalf("tag keys = %#v", got)
	}
	for _, tag := range plan.Tags {
		if tag.Key == "ignored" {
			t.Fatalf("HB.tags should not be used when HB.labels is present: %#v", plan.Tags)
		}
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

func TestNormalizeBaseURLPreservesExplicitHTTPSchemeCaseInsensitively(t *testing.T) {
	baseURL, err := normalizeBaseURL("HTTP://homebox.local:7744")
	if err != nil {
		t.Fatalf("normalize explicit http URL: %v", err)
	}
	if baseURL != "http://homebox.local:7744/api/v1" {
		t.Fatalf("base URL = %q", baseURL)
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

func TestLegacyHomeboxPreviewPlansImagesWithoutDownloadingAttachmentBytes(t *testing.T) {
	var attachmentDownloads int
	server := newLegacyHomeboxTestServerWithAttachmentCounter(t, http.StatusOK, &attachmentDownloads)

	plan, err := NewLegacyImporter(nil).ReadImportPlan(context.Background(), ports.ImportSourceRequest{
		SourceType:          importplan.SourceLegacyHomebox,
		BaseURL:             server.URL,
		Username:            "user@example.com",
		Password:            "secret",
		IncludeImages:       true,
		AllowPrivateNetwork: true,
	})
	if err != nil {
		t.Fatalf("read private-network Homebox source: %v", err)
	}
	tagsByKey := map[string]importplan.TagDefinition{}
	for _, tag := range plan.Tags {
		tagsByKey[tag.Key] = tag
	}
	if tagsByKey["workshop"].Color != "#2F80ED" {
		t.Fatalf("expected live Homebox tag color preservation, got %+v", tagsByKey)
	}
	if tagsByKey["damaged"].Color != "" {
		t.Fatalf("expected invalid live Homebox tag color to be cleared, got %+v", tagsByKey["damaged"])
	}
	if attachmentDownloads != 0 {
		t.Fatalf("preview downloaded attachment bytes %d times", attachmentDownloads)
	}
	if len(plan.Attachments) != 1 {
		t.Fatalf("expected planned attachment metadata, got %+v", plan.Attachments)
	}
	if len(plan.Attachments[0].Content) != 0 {
		t.Fatalf("preview attachment included content bytes")
	}
}

func TestLegacyHomeboxLiveImportEnrichesItemTagsFromTagList(t *testing.T) {
	server := newLegacyHomeboxTestServerWithTags(t, `[{"id":"tag-workshop","name":"Workshop","color":"2f80ed"},{"id":"tag-damaged","name":"Damaged","color":"not-a-color"}]`, `[{"id":"tag-workshop","name":"Stale Workshop","color":"not-a-color"},{"id":"tag-damaged"}]`)

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
	tagsByKey := map[string]importplan.TagDefinition{}
	for _, tag := range plan.Tags {
		tagsByKey[tag.Key] = tag
	}
	if tagsByKey["workshop"].DisplayName != "Workshop" || tagsByKey["workshop"].Color != "#2F80ED" {
		t.Fatalf("expected tag list color enrichment, got %+v", tagsByKey)
	}
	if tagsByKey["damaged"].DisplayName != "Damaged" || tagsByKey["damaged"].Color != "" {
		t.Fatalf("expected invalid catalog color to be cleared, got %+v", tagsByKey["damaged"])
	}
	if got := plan.Assets[1].TagKeys; len(got) != 2 || got[0] != "damaged" || got[1] != "workshop" {
		t.Fatalf("expected item tags from tag list, got %#v", got)
	}
}

func TestLegacyHomeboxLiveImportContinuesWhenTagListIsUnavailable(t *testing.T) {
	server := newLegacyHomeboxTestServerWithConfig(t, legacyHomeboxServerConfig{
		itemDetailStatus: http.StatusOK,
		attachmentStatus: http.StatusOK,
		tagsStatus:       http.StatusNotFound,
		itemTagsJSON:     `[{"name":"Workshop","color":"2f80ed"}]`,
	})

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
	if got := plan.Assets[1].TagKeys; len(got) != 1 || got[0] != "workshop" {
		t.Fatalf("expected fallback item detail tags, got %#v", got)
	}
	if plan.Tags[0].Color != "#2F80ED" {
		t.Fatalf("expected item detail tag color fallback, got %+v", plan.Tags)
	}
	if len(plan.Messages) != 1 || plan.Messages[0].Code != "tag-list-unavailable" {
		t.Fatalf("expected tag list warning, got %+v", plan.Messages)
	}
}

func TestLegacyHomeboxApplyDownloadsAttachmentBytesWhenRequested(t *testing.T) {
	var attachmentDownloads int
	server := newLegacyHomeboxTestServerWithAttachmentCounter(t, http.StatusOK, &attachmentDownloads)

	plan, err := NewLegacyImporter(nil).ReadImportPlan(context.Background(), ports.ImportSourceRequest{
		SourceType:           importplan.SourceLegacyHomebox,
		BaseURL:              server.URL,
		Username:             "user@example.com",
		Password:             "secret",
		IncludeImages:        true,
		FetchAttachmentBytes: true,
		AllowPrivateNetwork:  true,
	})
	if err != nil {
		t.Fatalf("read private-network Homebox source: %v", err)
	}
	if attachmentDownloads != 1 {
		t.Fatalf("apply downloaded attachment bytes %d times", attachmentDownloads)
	}
	if len(plan.Attachments) != 1 || string(plan.Attachments[0].Content) != "image-bytes" {
		t.Fatalf("expected downloaded attachment content, got %+v", plan.Attachments)
	}
}

func TestLegacyHomeboxApplyPreservesAttachmentIdentityWhenDownloadFails(t *testing.T) {
	var attachmentDownloads int
	server := newLegacyHomeboxTestServerWithAttachmentResponse(t, http.StatusOK, http.StatusNotFound, &attachmentDownloads)

	plan, err := NewLegacyImporter(nil).ReadImportPlan(context.Background(), ports.ImportSourceRequest{
		SourceType:           importplan.SourceLegacyHomebox,
		BaseURL:              server.URL,
		Username:             "user@example.com",
		Password:             "secret",
		IncludeImages:        true,
		FetchAttachmentBytes: true,
		AllowPrivateNetwork:  true,
	})
	if err != nil {
		t.Fatalf("read private-network Homebox source: %v", err)
	}
	if attachmentDownloads != 1 {
		t.Fatalf("apply downloaded attachment bytes %d times", attachmentDownloads)
	}
	if len(plan.Attachments) != 1 || plan.Attachments[0].SourceID != "attachment-one" || len(plan.Attachments[0].Content) != 0 {
		t.Fatalf("expected attachment identity without bytes, got %+v", plan.Attachments)
	}
	if plan.Attachments[0].UnavailableReason != "attachment could not be downloaded" {
		t.Fatalf("expected unavailable attachment reason, got %+v", plan.Attachments[0])
	}
	if len(plan.Messages) != 0 {
		t.Fatalf("expected application layer to report unavailable attachment warning, got %+v", plan.Messages)
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
	return newLegacyHomeboxTestServerWithAttachmentCounter(t, itemDetailStatus, nil)
}

func newLegacyHomeboxTestServerWithAttachmentCounter(t *testing.T, itemDetailStatus int, attachmentDownloads *int) *httptest.Server {
	return newLegacyHomeboxTestServerWithAttachmentResponse(t, itemDetailStatus, http.StatusOK, attachmentDownloads)
}

func newLegacyHomeboxTestServerWithAttachmentResponse(t *testing.T, itemDetailStatus int, attachmentStatus int, attachmentDownloads *int) *httptest.Server {
	t.Helper()
	return newLegacyHomeboxTestServerWithConfig(t, legacyHomeboxServerConfig{
		itemDetailStatus:    itemDetailStatus,
		attachmentStatus:    attachmentStatus,
		attachmentDownloads: attachmentDownloads,
		itemTagsJSON:        `[{"name":"Workshop","color":"2f80ed"},{"name":"Damaged","color":"not-a-color"}]`,
		tagsJSON:            `[{"id":"tag-workshop","name":"Workshop","color":"2f80ed"},{"id":"tag-damaged","name":"Damaged","color":"not-a-color"}]`,
	})
}

func newLegacyHomeboxTestServerWithTags(t *testing.T, tagsJSON string, itemTagsJSON string) *httptest.Server {
	t.Helper()
	return newLegacyHomeboxTestServerWithConfig(t, legacyHomeboxServerConfig{
		itemDetailStatus: http.StatusOK,
		attachmentStatus: http.StatusOK,
		tagsJSON:         tagsJSON,
		itemTagsJSON:     itemTagsJSON,
	})
}

type legacyHomeboxServerConfig struct {
	itemDetailStatus    int
	attachmentStatus    int
	attachmentDownloads *int
	tagsStatus          int
	tagsJSON            string
	itemTagsJSON        string
}

func newLegacyHomeboxTestServerWithConfig(t *testing.T, config legacyHomeboxServerConfig) *httptest.Server {
	t.Helper()
	if config.itemDetailStatus == 0 {
		config.itemDetailStatus = http.StatusOK
	}
	if config.attachmentStatus == 0 {
		config.attachmentStatus = http.StatusOK
	}
	if config.tagsStatus == 0 {
		config.tagsStatus = http.StatusOK
	}
	if config.tagsJSON == "" {
		config.tagsJSON = `[]`
	}
	if config.itemTagsJSON == "" {
		config.itemTagsJSON = `[]`
	}

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
		case "/api/v1/tags":
			if config.tagsStatus != http.StatusOK {
				http.Error(w, "tag list failed", config.tagsStatus)
				return
			}
			_, _ = w.Write([]byte(config.tagsJSON))
		case "/api/v1/items/item-one":
			if config.itemDetailStatus != http.StatusOK {
				http.Error(w, "database password leaked", config.itemDetailStatus)
				return
			}
			_, _ = w.Write([]byte(`{"id":"item-one","assetId":"HB-1","name":"Drill","description":"Cordless","quantity":1,"location":{"id":"location-one","name":"Garage"},"tags":` + config.itemTagsJSON + `,"attachments":[{"id":"attachment-one","type":"photo","primary":true,"title":"drill.jpg","mimeType":"image/jpeg"}]}`))
		case "/api/v1/items/item-one/attachments/attachment-one":
			if config.attachmentDownloads != nil {
				(*config.attachmentDownloads)++
			}
			if config.attachmentStatus != http.StatusOK {
				http.Error(w, "not found", config.attachmentStatus)
				return
			}
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("image-bytes"))
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
