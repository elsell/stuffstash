package app

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestExecuteImportJobCreatesReusesAndAssignsTags(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	source.plan.Fields = nil
	source.plan.Attachments = nil
	source.plan.Tags = []importplan.TagDefinition{
		{Key: "existing", DisplayName: "Existing", Color: "#111111"},
		{Key: "workshop", DisplayName: "Workshop", Color: "#2F80ED"},
	}
	source.plan.Assets[0].CustomFields = map[string]any{}
	source.plan.Assets[0].TagKeys = []string{"existing", "workshop"}
	application := New(Dependencies{
		Authorizer:                &fakeAuthorizer{},
		Tenants:                   store,
		Inventories:               store,
		CustomAssetTypes:          store,
		CustomAssetTypeUnitOfWork: store,
		CustomFields:              store,
		CustomFieldUnitOfWork:     store,
		Assets:                    store,
		AssetTags:                 store,
		AssetUnitOfWork:           store,
		AssetTagUnitOfWork:        store,
		Undoables:                 store,
		Audit:                     store,
		ImportSources:             source,
		ImportJobs:                store,
		ImportSourceVault: &fakeImportSourceVault{
			requests: map[importjob.ID]ports.ImportSourceRequest{},
		},
		ImportLinks:           store,
		ImportAssetUnitOfWork: store,
		ImportWorker:          &fakeImportWorker{},
		IDs: &fakeIDGenerator{ids: []string{
			"existing-tag", "audit-existing-tag",
			"job-one", "audit-preview", "audit-start",
			"workshop-tag", "audit-workshop-tag",
			"asset-one", "op-asset-one", "audit-asset-one", "audit-asset-tags",
			"audit-terminal", "audit-read",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	if _, err := application.CreateAssetTag(ctx, CreateAssetTagInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Key:         "existing",
		DisplayName: "Existing",
	}); err != nil {
		t.Fatalf("seed existing tag: %v", err)
	}
	preview, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importjob.SourceTypeLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if err != nil {
		t.Fatalf("create import preview: %v", err)
	}
	if preview.Counts.Tags != 2 || len(preview.Preview.Tags) != 2 {
		t.Fatalf("expected tag counts and preview tags, got counts=%+v preview=%+v", preview.Counts, preview.Preview.Tags)
	}

	result := startAndExecuteImportJob(t, ctx, application, preview.ID)
	if result.Status != importjob.StatusSucceeded || result.Counts.TagsCreated != 1 || result.Counts.TagsExisting != 1 || result.Counts.AssetsCreated != 1 {
		t.Fatalf("expected tag create/reuse counts and imported asset, got %+v", result)
	}
	assets, err := store.ListAssetsByInventory(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AssetListPageRequest{Limit: 10, LifecycleFilter: ports.AssetLifecycleFilterAll})
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) != 1 {
		t.Fatalf("expected one imported asset, got %+v", assets)
	}
	detail, err := application.GetAssetDetail(ctx, GetAssetInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     assets[0].ID,
	})
	if err != nil {
		t.Fatalf("get imported asset detail: %v", err)
	}
	if len(detail.Tags) != 2 || detail.Tags[0].Key.String() != "existing" || detail.Tags[1].Key.String() != "workshop" {
		t.Fatalf("expected imported asset tags, got %+v", detail.Tags)
	}
}
