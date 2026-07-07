package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestExecuteImportJobCleansBlobWhenImportedAttachmentLinkTransactionFails(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	source.plan.Fields = nil
	source.plan.Attachments[0].ContentType = "image/png"
	source.plan.Attachments[0].FileName = "drill.png"
	source.plan.Attachments[0].Content = pngAttachmentBytes()
	source.plan.Assets[0].CustomFields = map[string]any{}
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
		ImportAttachmentUnitOfWork:    failingImportAttachmentUnitOfWork{err: ports.ErrConflict},
		ImportWorker:                  &fakeImportWorker{},
		BlobDeletionOutboxMaxAttempts: 2,
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "audit-preview", "audit-start", "asset-one", "audit-asset", "attachment-one", "audit-attachment", "audit-complete",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	result := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if result.Status != importjob.StatusSucceeded || result.Counts.AttachmentsCreated != 0 || result.Counts.AttachmentsSkipped != 1 {
		t.Fatalf("expected attachment to be skipped after import link conflict, got %+v", result)
	}
	if _, err := store.GetBlob(ctx, media.StorageKey("tenant-one/inventory-one/asset-one/attachment-one")); err == nil {
		t.Fatalf("expected failed imported attachment transaction to clean up blob")
	}
	attachments, err := store.ListAttachmentsByAsset(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), asset.ID("asset-one"), ports.AttachmentListPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list attachments: %v", err)
	}
	if len(attachments) != 0 {
		t.Fatalf("expected no orphaned imported attachment metadata, got %+v", attachments)
	}
}

func TestExecuteImportJobSkipsAttachmentWithExistingSourceLinkAcrossJobs(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	observer := &fakeObserver{}
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	source.plan.Fields = nil
	source.plan.Assets[0].CustomFields = map[string]any{}
	source.plan.Attachments[0].ContentType = "image/png"
	source.plan.Attachments[0].FileName = "drill.png"
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
			"job-one", "asset-one", "audit-asset-one", "attachment-one", "audit-attachment-one", "audit-complete-one",
			"job-two", "audit-complete-two",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})
	application.observer = observer

	first := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if first.Status != importjob.StatusSucceeded || first.Counts.AssetsCreated != 1 || first.Counts.AttachmentsCreated != 1 {
		t.Fatalf("expected first import to create one asset and attachment, got %+v", first)
	}
	secondPreview, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importjob.SourceTypeLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if err != nil {
		t.Fatalf("create second preview: %v", err)
	}
	second := startAndExecuteImportJob(t, ctx, application, secondPreview.ID)
	if second.Status != importjob.StatusSucceeded || second.Counts.AssetsCreated != 0 || second.Counts.AssetsSkipped != 1 || second.Counts.AttachmentsCreated != 0 || second.Counts.AttachmentsSkipped != 1 {
		t.Fatalf("expected second import to skip linked asset and attachment, got %+v", second)
	}
	resources, err := store.ListImportJobResources(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), importjob.ID("job-one"), ports.ImportJobResourcePageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list import resources: %v", err)
	}
	var attachmentOwnerID asset.ID
	var attachmentResource ports.ImportJobResource
	for _, resource := range resources {
		if resource.ResourceType == ports.ImportResourceAttachment {
			if resource.SourceEntityType != ports.ImportSourceEntityAttachment || resource.SourceEntityID != "attachment:source:drill" {
				t.Fatalf("expected attachment source identity in import resource, got %+v", resource)
			}
			parsed, ok := asset.NewID(resource.ResourceOwnerID)
			if !ok {
				t.Fatalf("expected valid attachment owner ID in import resource, got %+v", resource)
			}
			attachmentOwnerID = parsed
			attachmentResource = resource
			break
		}
	}
	if attachmentOwnerID.String() == "" {
		t.Fatalf("expected first import to record imported attachment resource, got %+v", resources)
	}
	attachments, err := store.ListAttachmentsByAsset(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), attachmentOwnerID, ports.AttachmentListPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list attachments: %v", err)
	}
	if len(attachments) != 1 {
		t.Fatalf("expected exactly one attachment after repeated import, got %+v", attachments)
	}
	if attachmentResource.ResourceID != attachments[0].ID.String() {
		t.Fatalf("expected import resource to point at persisted attachment %q, got %+v", attachments[0].ID, attachmentResource)
	}
	sourceIdentity := importSourceIdentity{sourceType: importplan.SourceLegacyHomebox, sourceInstanceKey: "https://homebox.example.test"}
	link, found, err := store.ImportSourceLinkByKey(ctx, importAttachmentSourceLinkKey(tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), sourceIdentity, source.plan.Attachments[0]))
	if err != nil {
		t.Fatalf("read attachment source link: %v", err)
	}
	if !found || link.ResourceType != ports.ImportResourceAttachment || link.ResourceID != attachments[0].ID.String() || link.JobID != importjob.ID("job-one") {
		t.Fatalf("expected attachment source link to point at first imported attachment, found=%t link=%+v attachment=%+v", found, link, attachments[0])
	}
	var attachmentDuplicateEvent ports.Event
	for _, event := range observer.events {
		if event.Name == ports.EventImportJobSourceLinkDuplicateSkipped && event.Fields["source_entity_type"] == string(ports.ImportSourceEntityAttachment) && event.Fields["job_id"] == secondPreview.ID.String() {
			attachmentDuplicateEvent = event
			break
		}
	}
	if attachmentDuplicateEvent.Name == "" {
		t.Fatalf("expected attachment source-link duplicate skip event, got %+v", observer.events)
	}
	if attachmentDuplicateEvent.Fields["tenant_id"] != "tenant-one" || attachmentDuplicateEvent.Fields["inventory_id"] != "inventory-one" || attachmentDuplicateEvent.Fields["job_id"] != secondPreview.ID.String() {
		t.Fatalf("unexpected attachment duplicate skip event fields: %+v", attachmentDuplicateEvent.Fields)
	}
	for _, value := range attachmentDuplicateEvent.Fields {
		if strings.Contains(value, "attachment:source:drill") || strings.Contains(value, "secret") {
			t.Fatalf("duplicate skip event leaked source internals: %+v", attachmentDuplicateEvent.Fields)
		}
	}
}

func TestExecuteImportJobKeepsFetchedAttachmentBytesAfterFingerprintNormalization(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	source.plan.Fields = nil
	source.plan.Attachments[0].ContentType = "image/png"
	source.plan.Attachments[0].FileName = "drill.png"
	source.plan.Attachments[0].Content = pngAttachmentBytes()
	source.plan.Attachments[0].SizeBytes = len(source.plan.Attachments[0].Content)
	source.plan.Assets[0].CustomFields = map[string]any{}
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
			"job-one", "audit-preview", "audit-start", "asset-one", "audit-asset", "attachment-one", "audit-attachment", "audit-complete",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	result := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if result.Status != importjob.StatusSucceeded || result.Counts.AttachmentsCreated != 1 || result.Counts.AttachmentsSkipped != 0 {
		t.Fatalf("expected imported attachment to be created, got %+v", result)
	}
	attachments, err := store.ListAttachmentsByAsset(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), asset.ID("asset-one"), ports.AttachmentListPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list attachments: %v", err)
	}
	if len(attachments) != 1 || attachments[0].FileName.String() != "drill.png" {
		t.Fatalf("expected persisted imported attachment, got %+v", attachments)
	}
	if _, err := store.GetBlob(ctx, attachments[0].StorageKey); err != nil {
		t.Fatalf("expected imported attachment blob: %v", err)
	}
}

func TestExecuteImportJobAllowsApplyOnlyAttachmentMetadataToDifferFromPreview(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	preview := importPlanForDurableJob("Homebox", "source:drill")
	preview.Fields = nil
	preview.Assets[0].CustomFields = map[string]any{}
	preview.Attachments[0].FileName = "drill.jpg"
	preview.Attachments[0].ContentType = "image/jpeg"
	preview.Attachments[0].Content = nil
	preview.Attachments[0].SizeBytes = 0
	apply := preview
	apply.Attachments = append([]importplan.Attachment(nil), preview.Attachments...)
	apply.Attachments[0].FileName = "drill.png"
	apply.Attachments[0].ContentType = "image/png"
	apply.Attachments[0].Content = pngAttachmentBytes()
	apply.Attachments[0].SizeBytes = len(apply.Attachments[0].Content)
	source := &phaseAwareImportSourceReader{preview: preview, apply: apply}
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
			"job-one", "audit-preview", "audit-start", "asset-one", "audit-asset", "attachment-one", "audit-attachment", "audit-complete",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	result := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if result.Status != importjob.StatusSucceeded || result.Counts.AttachmentsCreated != 1 {
		t.Fatalf("expected apply-only attachment metadata to import, got %+v", result)
	}
	attachments, err := store.ListAttachmentsByAsset(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), asset.ID("asset-one"), ports.AttachmentListPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list attachments: %v", err)
	}
	if len(attachments) != 1 || attachments[0].ContentType != media.ContentTypePNG || attachments[0].FileName.String() != "drill.png" {
		t.Fatalf("expected sniffed apply attachment metadata to persist, got %+v", attachments)
	}
}

func TestExecuteImportJobSkipsUnavailableSourceAttachmentBytesWithoutUploadValidationWarning(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	preview := importPlanForDurableJob("Homebox", "source:drill")
	preview.Fields = nil
	preview.Assets[0].CustomFields = map[string]any{}
	preview.Attachments[0].Content = nil
	preview.Attachments[0].SizeBytes = 0
	apply := preview
	apply.Attachments = append([]importplan.Attachment(nil), preview.Attachments...)
	apply.Attachments[0].UnavailableReason = "attachment could not be downloaded"
	source := &phaseAwareImportSourceReader{preview: preview, apply: apply}
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
			"job-one", "audit-preview", "audit-start", "asset-one", "audit-asset", "audit-complete",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	result := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if result.Status != importjob.StatusSucceeded || result.Counts.AttachmentsCreated != 0 || result.Counts.AttachmentsSkipped != 1 {
		t.Fatalf("expected unavailable source attachment to be skipped, got %+v", result)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("expected one source attachment warning, got %+v", result.Messages)
	}
	if message := result.Messages[0]; message.Code != "attachment-unavailable" || message.Summary != "Attachment could not be downloaded" || message.Detail != "attachment could not be downloaded" {
		t.Fatalf("expected source download warning, got %+v", message)
	}
	attachments, err := store.ListAttachmentsByAsset(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), asset.ID("asset-one"), ports.AttachmentListPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list attachments: %v", err)
	}
	if len(attachments) != 0 {
		t.Fatalf("expected no persisted attachment for unavailable source bytes, got %+v", attachments)
	}
}

func TestExecuteImportJobPersistsPerPhaseProgressCheckpoints(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	importJobs := &recordingImportJobRepository{delegate: store}
	observer := &fakeObserver{}
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:item")}
	source.plan.Fields = []importplan.FieldDefinition{{Key: "homebox-source-id", DisplayName: "Homebox Source ID", Type: "text"}}
	source.plan.Assets = []importplan.Asset{
		{
			SourceID: "source:location",
			Kind:     "location",
			Title:    "Garage",
			CustomFields: map[string]any{
				"homebox-source-id": "source:location",
			},
		},
		{
			SourceID:       "source:item",
			ParentSourceID: "source:location",
			Kind:           "item",
			Title:          "Cordless drill",
			CustomFields: map[string]any{
				"homebox-source-id": "source:item",
			},
		},
	}
	source.plan.Attachments[0].AssetSourceID = "source:item"
	source.plan.Attachments[0].ContentType = "image/png"
	source.plan.Attachments[0].FileName = "drill.png"
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
		ImportJobs:                importJobs,
		ImportSourceVault: &fakeImportSourceVault{
			requests: map[importjob.ID]ports.ImportSourceRequest{},
		},
		ImportLinks:                   store,
		ImportAssetUnitOfWork:         store,
		ImportAttachmentUnitOfWork:    store,
		ImportWorker:                  &fakeImportWorker{},
		BlobDeletionOutboxMaxAttempts: 2,
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "field-one", "audit-field", "asset-location", "audit-location", "asset-item", "audit-item", "attachment-one", "audit-attachment", "audit-complete",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})
	application.observer = observer

	result := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if result.Status != importjob.StatusSucceeded {
		t.Fatalf("expected succeeded job, got %+v", result)
	}
	for _, expected := range []importjob.Progress{
		{Phase: importjob.PhaseFields, Done: 1, Total: 1},
		{Phase: importjob.PhaseLocations, Done: 1, Total: 1},
		{Phase: importjob.PhaseAssets, Done: 1, Total: 1},
		{Phase: importjob.PhaseAttachments, Done: 1, Total: 1},
	} {
		if !importJobs.sawProgress(expected.Phase, expected.Done, expected.Total) {
			t.Fatalf("expected progress checkpoint %+v in %+v", expected, importJobs.progresses)
		}
	}
	event, ok := observer.eventNamed(ports.EventImportJobProgressUpdated)
	if !ok {
		t.Fatalf("expected progress observability event, got %+v", observer.events)
	}
	if event.Fields["job_id"] != "job-one" || event.Fields["phase"] == "" || event.Fields["done"] == "" || event.Fields["total"] == "" {
		t.Fatalf("unexpected progress event fields: %+v", event.Fields)
	}
	for _, value := range event.Fields {
		if strings.Contains(value, "source:item") || strings.Contains(value, "secret") {
			t.Fatalf("progress event leaked source internals: %+v", event.Fields)
		}
	}
}

func TestExecuteImportJobImportsContainersAsAssets(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:container")}
	source.plan.Fields = nil
	source.plan.Attachments = nil
	source.plan.Assets[0].Kind = "container"
	source.plan.Assets[0].Title = "Tool bin"
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
		IDs:                           &fakeIDGenerator{ids: []string{"job-one", "asset-container", "audit-container", "audit-complete"}},
		Clock:                         fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	result := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if result.Status != importjob.StatusSucceeded || result.Counts.AssetsCreated != 1 {
		t.Fatalf("expected container to import as an asset, got %+v", result)
	}
	assets, err := store.ListAssetsByInventory(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AssetListPageRequest{Limit: 10, LifecycleFilter: ports.AssetLifecycleFilterAll})
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) != 1 || assets[0].Kind != "container" {
		t.Fatalf("expected imported container asset, got %+v", assets)
	}
}

func TestExecuteImportJobAdvancesProgressForSkippedAssetRecords(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	importJobs := &recordingImportJobRepository{delegate: store}
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:first")}
	source.plan.Fields = nil
	source.plan.Attachments = nil
	source.plan.Assets = []importplan.Asset{
		{
			SourceID: "source:archived",
			Kind:     "item",
			Title:    "Archived source item",
			Archived: true,
		},
		{
			SourceID: "source:active",
			Kind:     "item",
			Title:    "Active source item",
		},
	}
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
		ImportJobs:                importJobs,
		ImportSourceVault: &fakeImportSourceVault{
			requests: map[importjob.ID]ports.ImportSourceRequest{},
		},
		ImportLinks:                   store,
		ImportAssetUnitOfWork:         store,
		ImportAttachmentUnitOfWork:    store,
		ImportWorker:                  &fakeImportWorker{},
		BlobDeletionOutboxMaxAttempts: 2,
		IDs:                           &fakeIDGenerator{ids: []string{"job-one", "asset-active", "audit-active", "audit-complete"}},
		Clock:                         fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	result := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if result.Status != importjob.StatusSucceeded || result.Counts.AssetsCreated != 1 || result.Counts.AssetsSkipped != 1 {
		t.Fatalf("expected one created and one skipped asset, got %+v", result)
	}
	if !importJobs.sawProgress(importjob.PhaseAssets, 1, 2) || !importJobs.sawProgress(importjob.PhaseAssets, 2, 2) {
		t.Fatalf("expected skipped and completed asset checkpoints, got %+v", importJobs.progresses)
	}
}

func TestImportProgressCheckpointsDoNotOverwriteCancellationRequest(t *testing.T) {
	ctx := context.Background()
	repository := newFakeImportJobRepository()
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	job := importjob.NewPreviewedRecord(importjob.ID("job-one"), importjob.TenantID("tenant-one"), importjob.InventoryID("inventory-one"), importjob.PrincipalID("owner"), importjob.SourceRef{
		Type:        importjob.SourceTypeLegacyHomebox,
		Name:        "Homebox",
		Fingerprint: "sha256:test",
	}, importjob.Counts{}, nil, now)
	job.Status = importjob.StatusRunning
	if err := repository.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save import job: %v", err)
	}
	application := newDurableImportTestApp(repository, &fakeImportSourceReader{}, &fakeIDGenerator{}, fakeClock{now: now})
	job.Status = importjob.StatusCancelRequested
	job.CancellationMode = importjob.CancellationModeDiscardPartial
	if err := repository.UpdateImportJob(ctx, job); err != nil {
		t.Fatalf("seed cancellation: %v", err)
	}

	err := application.updateImportProgress(ctx, ports.ImportJobCommand{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
	}, importjob.PhaseAssets, 1, 2, "Creating assets")
	var cancelled importCancelledError
	if !errors.As(err, &cancelled) || cancelled.mode != importjob.CancellationModeDiscardPartial {
		t.Fatalf("expected progress checkpoint to observe cancellation request, got %v", err)
	}
	got := repository.jobs[job.ID]
	if got.Status != importjob.StatusCancelRequested {
		t.Fatalf("progress checkpoint overwrote cancellation request: %+v", got)
	}
}

func TestImportProgressCASPreservesCancellationWrittenDuringCheckpoint(t *testing.T) {
	ctx := context.Background()
	repository := newFakeImportJobRepository()
	repository.cancelDuringNextProgressUpdate = true
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	job := importjob.NewPreviewedRecord(importjob.ID("job-one"), importjob.TenantID("tenant-one"), importjob.InventoryID("inventory-one"), importjob.PrincipalID("owner"), importjob.SourceRef{
		Type:        importjob.SourceTypeLegacyHomebox,
		Name:        "Homebox",
		Fingerprint: "sha256:test",
	}, importjob.Counts{}, nil, now)
	job.Status = importjob.StatusRunning
	if err := repository.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save import job: %v", err)
	}
	application := newDurableImportTestApp(repository, &fakeImportSourceReader{}, &fakeIDGenerator{}, fakeClock{now: now})

	err := application.updateImportProgress(ctx, ports.ImportJobCommand{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
	}, importjob.PhaseAssets, 1, 2, "Creating assets")
	var cancelled importCancelledError
	if !errors.As(err, &cancelled) || cancelled.mode != importjob.CancellationModeDiscardPartial {
		t.Fatalf("expected cancellation from raced progress update, got %v", err)
	}
	got := repository.jobs[job.ID]
	if got.Status != importjob.StatusCancelRequested {
		t.Fatalf("progress update erased raced cancellation: %+v", got)
	}
	if got.Progress.Phase != importjob.PhaseAssets || got.Progress.Done != 1 || got.Progress.Total != 2 {
		t.Fatalf("expected raced progress checkpoint to persist without status overwrite, got %+v", got.Progress)
	}
}

func TestExecuteImportJobDoesNotOverwriteCancellationWrittenBeforeTerminalUpdate(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	importJobs := &terminalRaceImportJobRepository{delegate: store}
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:item")}
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
		ImportJobs:                importJobs,
		ImportSourceVault: &fakeImportSourceVault{
			requests: map[importjob.ID]ports.ImportSourceRequest{},
		},
		ImportLinks:                   store,
		ImportAssetUnitOfWork:         store,
		ImportAttachmentUnitOfWork:    store,
		ImportWorker:                  &fakeImportWorker{},
		BlobDeletionOutboxMaxAttempts: 2,
		IDs:                           &fakeIDGenerator{ids: []string{"job-one", "asset-one", "audit-create", "audit-cancelled"}},
		Clock:                         fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	result := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if result.Status != importjob.StatusCancelledKept {
		t.Fatalf("expected terminal race to honor cancellation, got %+v", result)
	}
	if stored, found, err := store.ImportJobByID(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), importjob.ID("job-one")); err != nil || !found || stored.Status != importjob.StatusCancelledKept {
		t.Fatalf("expected stored job to be cancelled, found=%v job=%+v err=%v", found, stored, err)
	}
}

func TestCancelImportJobRecordsKeepOrDiscardChoice(t *testing.T) {
	repository := newFakeImportJobRepository()
	application := newDurableImportTestApp(repository, &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}, &fakeIDGenerator{ids: []string{"job-one"}}, fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)})
	job, err := application.CreateImportJobPreview(context.Background(), CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importjob.SourceTypeLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}
	job.Status = importjob.StatusRunning
	if err := repository.UpdateImportJob(context.Background(), job); err != nil {
		t.Fatalf("mark running: %v", err)
	}

	cancelled, err := application.CancelImportJob(context.Background(), CancelImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Mode:        importjob.CancellationModeDiscardPartial,
	})
	if err != nil {
		t.Fatalf("cancel import job: %v", err)
	}
	if cancelled.Status != importjob.StatusCancelRequested || cancelled.CancellationMode != importjob.CancellationModeDiscardPartial {
		t.Fatalf("unexpected cancellation state: %+v", cancelled)
	}
	worker := application.importWorker.(*fakeImportWorker)
	if len(worker.cancelled) != 1 || worker.modes[0] != importjob.CancellationModeDiscardPartial {
		t.Fatalf("expected worker cancellation request")
	}
}

func TestCancelRunningImportJobPreservesProgressCounts(t *testing.T) {
	repository := newFakeImportJobRepository()
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	application := newDurableImportTestApp(repository, &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}, &fakeIDGenerator{ids: []string{"job-one"}}, fakeClock{now: now})
	job := importjob.NewPreviewedRecord(importjob.ID("job-one"), importjob.TenantID("tenant-one"), importjob.InventoryID("inventory-one"), importjob.PrincipalID("owner"), importjob.SourceRef{
		Type:        importjob.SourceTypeLegacyHomebox,
		Name:        "Homebox",
		Fingerprint: "sha256:test",
	}, importjob.Counts{}, nil, now)
	job.Status = importjob.StatusRunning
	job.Progress = importjob.Progress{Phase: importjob.PhaseAssets, Done: 1, Total: 2, Message: "Creating assets", UpdatedAt: now}
	if err := repository.SaveImportJob(context.Background(), job); err != nil {
		t.Fatalf("save running job: %v", err)
	}

	cancelled, err := application.CancelImportJob(context.Background(), CancelImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Mode:        importjob.CancellationModeKeepPartial,
	})
	if err != nil {
		t.Fatalf("cancel import job: %v", err)
	}
	if cancelled.Progress.Phase != importjob.PhaseAssets || cancelled.Progress.Done != 1 || cancelled.Progress.Total != 2 {
		t.Fatalf("expected cancellation to preserve progress counts, got %+v", cancelled.Progress)
	}
}
