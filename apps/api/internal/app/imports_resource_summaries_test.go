package app

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestImportJobResourceSummariesUseExistingRecordDisplayNames(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	source.plan.Fields = nil
	source.plan.Assets[0].Title = "Cordless drill"
	source.plan.Assets[0].CustomFields = map[string]any{}
	source.plan.Attachments[0].ContentType = "image/png"
	source.plan.Attachments[0].FileName = "drill-photo.png"
	source.plan.Attachments[0].Content = pngAttachmentBytes()
	application := New(Dependencies{
		Authorizer:                &fakeAuthorizer{},
		Tenants:                   store,
		Inventories:               store,
		CustomAssetTypes:          store,
		CustomAssetTypeUnitOfWork: store,
		CustomFields:              store,
		CustomFieldUnitOfWork:     store,
		Assets:                    store,
		AssetUnitOfWork:           store,
		Undoables:                 store,
		Audit:                     store,
		AttachmentUnitOfWork:      store,
		Attachments:               store,
		Blobs:                     store,
		BlobDeletionOutbox:        store,
		ImportSources:             source,
		ImportJobs:                store,
		ImportSourceVault: &fakeImportSourceVault{
			requests: map[importjob.ID]ports.ImportSourceRequest{},
		},
		ImportLinks:                   store,
		ImportAssetUnitOfWork:         store,
		ImportAttachmentUnitOfWork:    store,
		ImportWorker:                  &fakeImportWorker{},
		BlobDeletionOutboxMaxAttempts: 2,
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "asset-one", "audit-asset", "attachment-one", "audit-attachment", "audit-complete",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	result := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if result.Status != importjob.StatusSucceeded {
		t.Fatalf("expected import to succeed, got %+v", result)
	}
	detail, err := application.GetImportJob(ctx, GetImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       importjob.ID("job-one"),
	})
	if err != nil {
		t.Fatalf("get import job: %v", err)
	}

	displayNames := map[string]string{}
	for _, resource := range detail.Resources {
		displayNames[resource.ResourceType] = resource.DisplayName
	}
	if displayNames[string(ports.ImportResourceAsset)] != "Cordless drill" {
		t.Fatalf("expected imported asset display name, got resources %+v", detail.Resources)
	}
	if displayNames[string(ports.ImportResourceAttachment)] != "drill-photo.png" {
		t.Fatalf("expected imported attachment display name, got resources %+v", detail.Resources)
	}
}
