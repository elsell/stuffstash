package app

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestExecuteImportJobRetryDoesNotDoubleCountAlreadyDiscardedResources(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:first")}
	source.plan.Fields = nil
	source.plan.Attachments = nil
	source.plan.Assets[0].CustomFields = nil
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
			"job-one", "audit-preview", "audit-start", "asset-one", "audit-create", "audit-complete", "audit-cancelled",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})
	succeeded := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if succeeded.Status != importjob.StatusSucceeded || succeeded.Counts.AssetsCreated != 1 {
		t.Fatalf("expected setup import to succeed, got %+v", succeeded)
	}
	importedAssets, err := store.ListAssetsByInventory(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AssetListPageRequest{Limit: 10, LifecycleFilter: ports.AssetLifecycleFilterAll})
	if err != nil {
		t.Fatalf("list imported assets: %v", err)
	}
	if len(importedAssets) != 1 {
		t.Fatalf("expected one imported asset before simulating partial cleanup, got %+v", importedAssets)
	}
	if err := store.DeleteAsset(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), importedAssets[0].ID, auditRecord("audit-prior-discard", "tenant-one", "inventory-one", audit.ActionAssetDeleted)); err != nil {
		t.Fatalf("simulate prior discard delete: %v", err)
	}
	succeeded.Status = importjob.StatusDiscardFailed
	succeeded.CancellationMode = importjob.CancellationModeDiscardPartial
	succeeded.Counts.RecordsDiscarded = 1
	succeeded.Counts.SourceLinksDiscarded = 0
	if err := store.UpdateImportJob(ctx, succeeded); err != nil {
		t.Fatalf("mark discard failed after partial cleanup: %v", err)
	}

	result, err := application.ExecuteImportJob(ctx, ports.ImportJobCommand{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       importjob.ID("job-one"),
	})
	if err != nil {
		t.Fatalf("retry discard cleanup: %v", err)
	}
	if result.Status != importjob.StatusCancelledDiscarded || result.Counts.RecordsDiscarded != 1 || result.Counts.SourceLinksDiscarded != 1 {
		t.Fatalf("expected retry cleanup to preserve prior discard count and delete remaining source link, got %+v", result)
	}
	records, err := store.ListInventoryAuditRecords(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AuditRecordPageRequest{Limit: 20})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	if got := countAuditAction(records, audit.ActionAssetDeleted); got != 1 {
		t.Fatalf("expected retry not to emit another asset delete audit, got %d records in %+v", got, records)
	}
	if got := countAuditAction(records, audit.ActionImportJobCancelled); got != 1 {
		t.Fatalf("expected one terminal cancellation audit, got %d records in %+v", got, records)
	}
}

func TestExecuteImportJobRetryDoesNotDoubleCountAlreadyDiscardedAttachments(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:first")}
	source.plan.Fields = nil
	source.plan.Assets[0].CustomFields = nil
	source.plan.Attachments[0].ContentType = "image/png"
	source.plan.Attachments[0].FileName = "drill.png"
	source.plan.Attachments[0].Content = pngAttachmentBytes()
	source.plan.Attachments[0].SizeBytes = len(source.plan.Attachments[0].Content)
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
			"job-one", "audit-preview", "audit-start", "asset-one", "audit-asset", "attachment-one", "audit-attachment", "audit-complete", "audit-credential-cleaned", "audit-retry-asset-delete", "audit-cancelled",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})
	succeeded := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if succeeded.Status != importjob.StatusSucceeded || succeeded.Counts.AssetsCreated != 1 || succeeded.Counts.AttachmentsCreated != 1 {
		t.Fatalf("expected setup import to create asset and attachment, got %+v", succeeded)
	}
	importedAssets, err := store.ListAssetsByInventory(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AssetListPageRequest{Limit: 10, LifecycleFilter: ports.AssetLifecycleFilterAll})
	if err != nil {
		t.Fatalf("list imported assets: %v", err)
	}
	if len(importedAssets) != 1 {
		t.Fatalf("expected one imported asset before simulating partial cleanup, got %+v", importedAssets)
	}
	attachments, err := store.ListAttachmentsByAsset(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), importedAssets[0].ID, ports.AttachmentListPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list imported attachments: %v", err)
	}
	if len(attachments) != 1 {
		t.Fatalf("expected one imported attachment before simulating partial cleanup, got %+v", attachments)
	}
	if _, removed, err := store.DeleteAttachmentAndEnqueueBlobDeletion(ctx, "blob-prior-discard", tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), importedAssets[0].ID, attachments[0].ID, auditRecord("audit-prior-attachment-discard", "tenant-one", "inventory-one", audit.ActionAttachmentDeleted)); err != nil || !removed {
		t.Fatalf("simulate prior attachment discard delete: removed=%t err=%v", removed, err)
	}
	succeeded.Status = importjob.StatusDiscardFailed
	succeeded.CancellationMode = importjob.CancellationModeDiscardPartial
	succeeded.Counts.RecordsDiscarded = 1
	succeeded.Counts.SourceLinksDiscarded = 0
	if err := store.UpdateImportJob(ctx, succeeded); err != nil {
		t.Fatalf("mark discard failed after partial attachment cleanup: %v", err)
	}

	result, err := application.ExecuteImportJob(ctx, ports.ImportJobCommand{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       importjob.ID("job-one"),
	})
	if err != nil {
		t.Fatalf("retry discard cleanup: %v", err)
	}
	if result.Status != importjob.StatusCancelledDiscarded || result.Counts.RecordsDiscarded != 2 || result.Counts.SourceLinksDiscarded != 2 {
		t.Fatalf("expected retry cleanup to preserve prior attachment discard count, delete asset, and remove remaining source links, got %+v", result)
	}
	records, err := store.ListInventoryAuditRecords(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AuditRecordPageRequest{Limit: 30})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	if got := countAuditAction(records, audit.ActionAttachmentDeleted); got != 1 {
		t.Fatalf("expected retry not to emit another attachment delete audit, got %d records in %+v", got, records)
	}
	if got := countAuditAction(records, audit.ActionAssetDeleted); got != 1 {
		t.Fatalf("expected retry to delete remaining asset once, got %d records in %+v", got, records)
	}
	if got := countAuditAction(records, audit.ActionImportJobCancelled); got != 1 {
		t.Fatalf("expected one terminal cancellation audit, got %d records in %+v", got, records)
	}
}

func countAuditAction(records []audit.Record, action audit.Action) int {
	count := 0
	for _, record := range records {
		if record.Action == action {
			count++
		}
	}
	return count
}
