package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/inventoryexport"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const exportTestPath = "/tenants/home/inventories/main/export"

func exportFixture(t *testing.T, count int) (app.Dependencies, *memory.Store, *memory.Authorizer) {
	t.Helper()
	ctx := context.Background()
	store := memory.NewStore()
	access := memory.NewAuthorizer()
	seedMemoryStore(t, ctx, store, access, seededState{tenants: []seedTenant{{id: "home", name: "Home", owner: "owner"}}, inventories: []seedInventory{{id: "main", tenantID: "home", name: "Main", owner: "owner"}}})
	if err := access.GrantInventoryViewer(ctx, principal("viewer"), "home", "main"); err != nil {
		t.Fatal(err)
	}
	if err := access.GrantInventoryEditor(ctx, principal("editor"), "home", "main"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("asset-%04d", i)
		state := asset.LifecycleStateActive
		if i%2 == 0 {
			state = asset.LifecycleStateArchived
		}
		item := asset.Asset{ID: asset.ID(id), TenantID: "home", InventoryID: "main", Title: asset.Title(id), Kind: asset.KindItem, LifecycleState: state, CustomFields: asset.NewEmptyCustomFields(), CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if err := store.CreateAsset(ctx, item, audit.Record{ID: audit.ID("seed-" + id)}, nil); err != nil {
			t.Fatal(err)
		}
	}
	return app.Dependencies{Auth: auth.NewLocalDevAuthenticator(), Authorizer: access, Tenants: store, Inventories: store, Assets: store, AssetTags: store, Checkouts: store, Attachments: store, CustomAssetTypes: store, CustomFields: store, Audit: store, IDs: &fakeIDGenerator{}, ExportEncoder: inventoryexport.Encoder{}, ExportMaxRecords: 10000, ExportMaxBytes: 64 * 1024 * 1024}, store, access
}

func TestExportTraversesPagesIncludesArchivedAndAuditsViewers(t *testing.T) {
	deps, store, _ := exportFixture(t, 201)
	server := NewServer(":0", app.New(deps))
	for _, role := range []string{"viewer", "editor", "owner"} {
		r := performRequest(server, http.MethodGet, exportTestPath, "Bearer dev:"+role, nil)
		if r.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", role, r.Code, r.Body.String())
		}
		var doc struct {
			Assets []struct{ ID, LifecycleState string }
		}
		if err := json.Unmarshal(r.Body.Bytes(), &doc); err != nil {
			t.Fatal(err)
		}
		if len(doc.Assets) != 201 || doc.Assets[0].LifecycleState != "archived" || doc.Assets[200].ID != "asset-0200" {
			t.Fatalf("truncated export: %d", len(doc.Assets))
		}
	}
	records, err := store.ListInventoryAuditRecords(context.Background(), "home", "main", ports.AuditRecordPageRequest{})
	if err != nil || len(records) != 3 {
		t.Fatalf("export audit: %d %v", len(records), err)
	}
	for _, record := range records {
		if record.Action != audit.ActionInventoryExported || record.Metadata["asset_count"] != "201" {
			t.Fatalf("unexpected audit: %#v", record)
		}
	}
}

type unavailableExportAudit struct{ *memory.Store }

func (unavailableExportAudit) SaveAuditRecord(context.Context, audit.Record) error {
	return errors.New("audit unavailable")
}
func TestExportFailureNeverPublishesInventory(t *testing.T) {
	for _, mode := range []string{"record limit", "byte limit", "audit unavailable", "wrong scope"} {
		t.Run(mode, func(t *testing.T) {
			deps, store, _ := exportFixture(t, 2)
			expected := http.StatusUnprocessableEntity
			switch mode {
			case "record limit":
				deps.ExportMaxRecords = 1
			case "byte limit":
				deps.ExportMaxBytes = 50
			case "audit unavailable":
				deps.Audit = unavailableExportAudit{store}
				expected = http.StatusInternalServerError
			case "wrong scope":
				deps.Assets = wrongScopeExportAssets{store}
				expected = http.StatusForbidden
			}
			r := performRequest(NewServer(":0", app.New(deps)), http.MethodGet, exportTestPath, "Bearer dev:viewer", nil)
			if r.Code != expected || strings.Contains(r.Body.String(), "asset-0000") || r.Header().Get("Content-Disposition") != "" {
				t.Fatalf("published failed export: %d %s", r.Code, r.Body.String())
			}
		})
	}
}

type wrongScopeExportAssets struct{ *memory.Store }

func (s wrongScopeExportAssets) ListAssetsByInventory(ctx context.Context, t tenant.ID, i inventory.InventoryID, p ports.AssetListPageRequest) ([]asset.Asset, error) {
	rows, err := s.Store.ListAssetsByInventory(ctx, t, i, p)
	if len(rows) > 0 {
		rows[0].InventoryID = "other"
	}
	return rows, err
}

type pausedExportEncoder struct{ entered, release chan struct{} }

func (p pausedExportEncoder) Encode(ctx context.Context, d ports.InventoryExportDocument, f ports.InventoryExportFormat, max int) ([]byte, error) {
	close(p.entered)
	select {
	case <-p.release:
		return (inventoryexport.Encoder{}).Encode(ctx, d, f, max)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func TestExportRechecksRevokedAccessBeforeDownload(t *testing.T) {
	deps, _, access := exportFixture(t, 1)
	encoder := pausedExportEncoder{make(chan struct{}), make(chan struct{})}
	deps.ExportEncoder = encoder
	server := NewServer(":0", app.New(deps))
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { done <- performRequest(server, http.MethodGet, exportTestPath, "Bearer dev:viewer", nil) }()
	select {
	case <-encoder.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("export did not reach encoding")
	}
	if err := access.RevokeInventoryViewer(context.Background(), principal("viewer"), "home", "main"); err != nil {
		close(encoder.release)
		t.Fatal(err)
	}
	close(encoder.release)
	select {
	case r := <-done:
		if r.Code != http.StatusForbidden || strings.Contains(r.Body.String(), "asset-0000") {
			t.Fatalf("revoked export published: %d %s", r.Code, r.Body.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("export did not finish")
	}
}

func TestExportPreservesInheritedAndLocalSchemaAcrossPages(t *testing.T) {
	deps, store, _ := exportFixture(t, 0)
	for i := 0; i < 201; i++ {
		scope := customfield.ScopeTenant
		inventoryID := customfield.InventoryID("")
		if i%2 == 1 {
			scope = customfield.ScopeInventory
			inventoryID = "main"
		}
		key := customfield.Key(fmt.Sprintf("field-%04d", i))
		definition := customfield.Definition{ID: customfield.ID(fmt.Sprintf("definition-%04d", i)), TenantID: "home", InventoryID: inventoryID, Scope: scope, Key: key, DisplayName: customfield.DisplayName(key), Type: customfield.FieldTypeText, Applicability: customfield.ApplicabilityAllAssets, LifecycleState: customfield.DefinitionLifecycleArchived}
		if err := store.SaveCustomFieldDefinition(context.Background(), definition, audit.Record{ID: audit.ID(fmt.Sprintf("definition-audit-%04d", i))}); err != nil {
			t.Fatal(err)
		}
		assetType := customfield.AssetType{ID: customfield.AssetTypeID(fmt.Sprintf("type-%04d", i)), TenantID: "home", InventoryID: inventoryID, Scope: scope, Key: key, DisplayName: customfield.DisplayName(key), LifecycleState: customfield.AssetTypeLifecycleArchived}
		if err := store.SaveCustomAssetType(context.Background(), assetType, audit.Record{ID: audit.ID(fmt.Sprintf("type-audit-%04d", i))}); err != nil {
			t.Fatal(err)
		}
	}
	r := performRequest(NewServer(":0", app.New(deps)), http.MethodGet, exportTestPath, "Bearer dev:viewer", nil)
	if r.Code != http.StatusOK {
		t.Fatalf("schema export: %d %s", r.Code, r.Body.String())
	}
	var doc struct {
		CustomAssetTypes       []struct{ ID, Scope, LifecycleState string }
		CustomFieldDefinitions []struct{ ID, Scope, LifecycleState string }
	}
	if err := json.Unmarshal(r.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.CustomAssetTypes) != 201 || len(doc.CustomFieldDefinitions) != 201 {
		t.Fatalf("lost schema: %d types, %d fields", len(doc.CustomAssetTypes), len(doc.CustomFieldDefinitions))
	}
	seen := map[string]bool{}
	for _, field := range doc.CustomFieldDefinitions {
		if seen[field.ID] || field.LifecycleState != "archived" {
			t.Fatalf("invalid field: %#v", field)
		}
		seen[field.ID] = true
	}
}

func TestExportIncludesArchivedTagsAssignmentsAndAttachmentMetadata(t *testing.T) {
	deps, store, _ := exportFixture(t, 2)
	ctx := context.Background()
	tag := assettag.Tag{ID: "tag", TenantID: "home", InventoryID: "main", Key: "kept", DisplayName: "Kept", LifecycleState: assettag.LifecycleStateActive, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := store.CreateAssetTag(ctx, tag, audit.Record{ID: "tag-create"}); err != nil {
		t.Fatal(err)
	}
	if err := store.SetAssetTags(ctx, "home", "main", "asset-0001", []assettag.ID{"tag"}, audit.Record{ID: "tag-assign"}); err != nil {
		t.Fatal(err)
	}
	tag.LifecycleState = assettag.LifecycleStateArchived
	if err := store.UpdateAssetTagLifecycle(ctx, tag, audit.Record{ID: "tag-archive"}); err != nil {
		t.Fatal(err)
	}
	attachment := media.Attachment{ID: "photo", TenantID: "home", InventoryID: "main", AssetID: "asset-0001", StorageKey: "never-export", FileName: "manual.pdf", ContentType: "application/pdf", SizeBytes: 10, SHA256: strings.Repeat("a", 64), CreatedAt: time.Now(), LifecycleState: media.LifecycleStateArchived}
	if err := store.SaveAttachment(ctx, attachment, audit.Record{ID: "photo-create"}, nil); err != nil {
		t.Fatal(err)
	}
	r := performRequest(NewServer(":0", app.New(deps)), http.MethodGet, exportTestPath, "Bearer dev:viewer", nil)
	if r.Code != http.StatusOK {
		t.Fatalf("metadata export: %d %s", r.Code, r.Body.String())
	}
	var doc struct {
		Tags   []struct{ ID, LifecycleState string }
		Assets []struct {
			TagIDs      []string
			Attachments []struct{ ID, LifecycleState string }
		}
	}
	if err := json.Unmarshal(r.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Tags) != 1 || doc.Tags[0].LifecycleState != "archived" || len(doc.Assets) != 2 || len(doc.Assets[1].TagIDs) != 1 || len(doc.Assets[1].Attachments) != 1 || doc.Assets[1].Attachments[0].LifecycleState != "archived" || strings.Contains(r.Body.String(), "never-export") {
		t.Fatalf("lost or leaked metadata: %s", r.Body.String())
	}
	activeTags, err := store.AssetTagsByAsset(ctx, "home", "main", "asset-0001")
	if err != nil || len(activeTags) != 0 {
		t.Fatal("ordinary tag reads must remain active-only")
	}
	activeMedia, err := store.ListAttachmentsByAsset(ctx, "home", "main", "asset-0001", ports.AttachmentListPageRequest{})
	if err != nil || len(activeMedia) != 0 {
		t.Fatal("ordinary attachment reads must remain active-only")
	}
}
