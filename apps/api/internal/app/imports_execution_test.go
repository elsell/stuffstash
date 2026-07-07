package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestExecuteImportJobSkipsAssetWithExistingSourceLinkAcrossJobs(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	observer := &fakeObserver{}
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	source.plan.Fields = nil
	source.plan.Attachments = nil
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
		ImportSources:             source,
		ImportJobs:                store,
		ImportSourceVault: &fakeImportSourceVault{
			requests: map[importjob.ID]ports.ImportSourceRequest{},
		},
		ImportLinks:           store,
		ImportAssetUnitOfWork: store,
		ImportWorker:          &fakeImportWorker{},
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "asset-one", "audit-one",
			"job-two",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})
	application.observer = observer

	first := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if first.Status != importjob.StatusSucceeded || first.Counts.AssetsCreated != 1 {
		t.Fatalf("expected first import to create one asset, got %+v", first)
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
	if len(secondPreview.Messages) == 0 || secondPreview.Messages[0].Code != "duplicate-source-asset" {
		t.Fatalf("expected source-link duplicate warning, got %+v", secondPreview.Messages)
	}
	second := startAndExecuteImportJob(t, ctx, application, secondPreview.ID)
	if second.Status != importjob.StatusSucceeded || second.Counts.AssetsCreated != 0 || second.Counts.AssetsSkipped != 1 {
		t.Fatalf("expected second import to skip linked asset, got %+v", second)
	}
	event, ok := observer.eventNamed(ports.EventImportJobSourceLinkDuplicateSkipped)
	if !ok {
		t.Fatalf("expected source-link duplicate skip observability event, got %+v", observer.events)
	}
	if event.Fields["source_entity_type"] != string(ports.ImportSourceEntityAsset) || event.Fields["tenant_id"] != "tenant-one" || event.Fields["inventory_id"] != "inventory-one" || event.Fields["job_id"] == "" {
		t.Fatalf("unexpected duplicate skip event fields: %+v", event.Fields)
	}
	for _, value := range event.Fields {
		if strings.Contains(value, "source:drill") || strings.Contains(value, "secret") {
			t.Fatalf("duplicate skip event leaked source internals: %+v", event.Fields)
		}
	}
	assets, err := store.ListAssetsByInventory(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AssetListPageRequest{Limit: 10, LifecycleFilter: ports.AssetLifecycleFilterAll})
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) != 1 {
		t.Fatalf("expected one asset after repeated import, got %+v", assets)
	}
}

func TestExecuteImportJobCancelsAndDiscardsPartialProgressAtCheckpoint(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:first")}
	source.plan.Fields = nil
	source.plan.Attachments = nil
	source.plan.Assets[0].CustomFields = map[string]any{}
	source.plan.Assets = append(source.plan.Assets, importplan.Asset{
		SourceID: "source:second",
		Kind:     "item",
		Title:    "Second item",
	})
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
		ImportAssetUnitOfWork:         &cancellingImportAssetUnitOfWork{delegate: store, jobs: store},
		ImportWorker:                  &fakeImportWorker{},
		BlobDeletionOutboxMaxAttempts: 2,
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "asset-one", "undo-one", "audit-create-one", "audit-delete-one",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	result := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if result.Status != importjob.StatusCancelledDiscarded {
		t.Fatalf("expected cancelled discarded, got %+v", result)
	}
	if result.Counts.AssetsCreated != 1 || result.Counts.RecordsDiscarded != 1 || result.Counts.SourceLinksDiscarded != 1 {
		t.Fatalf("unexpected cancellation counts: %+v", result.Counts)
	}
	if result.Progress.Phase != importjob.PhaseTerminal || result.Progress.Done != 1 || result.Progress.Total != 2 {
		t.Fatalf("expected cancellation terminal progress to preserve last truthful checkpoint, got %+v", result.Progress)
	}
	assets, err := store.ListAssetsByInventory(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AssetListPageRequest{Limit: 10, LifecycleFilter: ports.AssetLifecycleFilterAll})
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) != 0 {
		t.Fatalf("expected imported asset discarded, got %+v", assets)
	}
	detail, err := application.GetImportJob(ctx, GetImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       importjob.ID("job-one"),
	})
	if err != nil {
		t.Fatalf("get discarded import job: %v", err)
	}
	if len(detail.Resources) != 0 {
		t.Fatalf("discarded job must not expose deleted resource links, got %+v", detail.Resources)
	}
	history, err := application.ListImportJobs(ctx, ListImportJobsInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
	})
	if err != nil {
		t.Fatalf("list discarded import job history: %v", err)
	}
	if len(history) != 1 || len(history[0].Resources) != 0 {
		t.Fatalf("discarded job history must not expose deleted resource links, got %+v", history)
	}
}

func TestExecuteImportJobDoesNotRecordDiscardCleanupEventWhenTerminalUpdateFails(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	importJobs := &failingTerminalImportJobRepository{delegate: store}
	observer := &fakeObserver{}
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:first")}
	source.plan.Fields = nil
	source.plan.Attachments = nil
	source.plan.Assets[0].CustomFields = map[string]any{}
	source.plan.Assets = append(source.plan.Assets, importplan.Asset{
		SourceID: "source:second",
		Kind:     "item",
		Title:    "Second item",
	})
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
		ImportAssetUnitOfWork:         &cancellingImportAssetUnitOfWork{delegate: store, jobs: store},
		ImportWorker:                  &fakeImportWorker{},
		BlobDeletionOutboxMaxAttempts: 2,
		Observer:                      observer,
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "asset-one", "undo-one", "audit-create-one", "audit-delete-one",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	job, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
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
	if _, err := application.StartImportJob(ctx, StartImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Source: ImportSourceInput{
			SourceType: string(importjob.SourceTypeLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	}); err != nil {
		t.Fatalf("start import job: %v", err)
	}

	_, err = application.ExecuteImportJob(ctx, ports.ImportJobCommand{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
	})
	if !errors.Is(err, errTerminalImportUpdateFailed) {
		t.Fatalf("expected terminal update failure, got %v", err)
	}
	if observer.hasEvent(ports.EventImportJobDiscardCleanupCompleted) || observer.hasEvent(ports.EventImportJobDiscardCleanupFailed) {
		t.Fatalf("discard cleanup event must wait for durable terminal update, got %+v", observer.events)
	}
}

func TestExecuteImportJobReturnsTypedSourceChangedErrorAfterTerminalizingJob(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:first")}
	source.plan.Fields = nil
	source.plan.Attachments = nil
	source.plan.Assets[0].CustomFields = nil
	source.plan.Messages = []importplan.Message{{
		Code:     "original-preview-warning",
		Severity: importplan.SeverityWarning,
		Summary:  "Original preview warning",
	}}
	vault := &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}}
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
		ImportSourceVault:         vault,
		ImportLinks:               store,
		ImportAssetUnitOfWork:     store,
		ImportWorker:              &fakeImportWorker{},
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "audit-preview", "audit-start", "audit-failed", "audit-credential-cleaned",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})
	job, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importjob.SourceTypeLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if err != nil {
		t.Fatalf("create preview: %v", err)
	}
	if _, err := application.StartImportJob(ctx, StartImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Source: ImportSourceInput{
			SourceType: string(importjob.SourceTypeLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	}); err != nil {
		t.Fatalf("start import job: %v", err)
	}
	source.plan = importPlanForDurableJob("Homebox", "source:changed")
	source.plan.Fields = nil
	source.plan.Attachments = nil
	source.plan.Assets[0].CustomFields = nil
	source.plan.Assets[0].Archived = true
	source.plan.Messages = []importplan.Message{{
		Code:     "changed-source-warning",
		Severity: importplan.SeverityWarning,
		Summary:  "Changed source warning",
	}}

	_, err = application.ExecuteImportJob(ctx, ports.ImportJobCommand{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
	})
	var sourceChanged ImportSourceChangedAfterPreviewError
	if !errors.As(err, &sourceChanged) || !errors.Is(err, ErrPrecondition) {
		t.Fatalf("expected typed source changed precondition, got %T %[1]v", err)
	}
	failed, found, err := store.ImportJobByID(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), job.ID)
	if err != nil || !found {
		t.Fatalf("get import job after source changed: found=%t err=%v", found, err)
	}
	if failed.Status != importjob.StatusFailed || failed.Progress.Message != "Import source changed after preview" {
		t.Fatalf("expected source changed job failure, got %+v", failed)
	}
	if len(failed.Messages) != 1 {
		t.Fatalf("expected only source changed message without warnings from unreviewed changed source, got %+v", failed.Messages)
	}
	if failed.Messages[0].Code != "import-source-changed" || !strings.Contains(failed.Messages[0].Detail, "Preview the source again") {
		t.Fatalf("expected actionable source changed job message, got %+v", failed.Messages)
	}
	if _, found := vault.requests[job.ID]; found {
		t.Fatalf("expected source material cleaned after source changed failure")
	}
}

func TestExecuteImportJobRetriesDiscardFailedCleanup(t *testing.T) {
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
			"job-one", "audit-preview", "audit-start", "asset-one", "audit-create", "audit-complete", "audit-delete", "audit-cancelled",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})
	succeeded := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if succeeded.Status != importjob.StatusSucceeded || succeeded.Counts.AssetsCreated != 1 {
		t.Fatalf("expected setup import to succeed, got %+v", succeeded)
	}
	succeeded.Status = importjob.StatusDiscardFailed
	succeeded.CancellationMode = importjob.CancellationModeDiscardPartial
	if err := store.UpdateImportJob(ctx, succeeded); err != nil {
		t.Fatalf("mark discard failed: %v", err)
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
		t.Fatalf("expected discard retry to finish cleanup, got %+v", result)
	}
	assets, err := store.ListAssetsByInventory(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AssetListPageRequest{Limit: 10, LifecycleFilter: ports.AssetLifecycleFilterAll})
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) != 0 {
		t.Fatalf("expected retry cleanup to remove imported asset, got %+v", assets)
	}
}

func TestResumeRunningImportJobsTerminalizesCancelRequestedJob(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	observer := &fakeObserver{}
	source := &fakeImportSourceReader{plan: importplan.Plan{
		Source: importplan.SourceSummary{
			Type:        importplan.SourceLegacyHomebox,
			Name:        "Homebox",
			BaseURL:     "https://homebox.example.test",
			Version:     "0.24.0",
			ImageImport: "disabled",
		},
	}}
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
		BlobDeletionOutboxMaxAttempts: 2,
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "audit-preview", "audit-cancelled",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})
	application.observer = observer
	application = application.WithImportWorker(syncImportWorker{application: application})
	job, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importjob.SourceTypeLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if err != nil {
		t.Fatalf("create preview: %v", err)
	}
	job.Status = importjob.StatusCancelRequested
	job.CancellationMode = importjob.CancellationModeDiscardPartial
	job.UpdatedAt = time.Date(2026, 7, 6, 12, 1, 0, 0, time.UTC)
	if err := store.UpdateImportJob(ctx, job); err != nil {
		t.Fatalf("seed cancel requested job: %v", err)
	}

	resumed, err := application.ResumeRunningImportJobs(ctx, 10)
	if err != nil {
		t.Fatalf("resume import jobs: %v", err)
	}
	if resumed != 1 {
		t.Fatalf("expected one resumed job, got %d", resumed)
	}
	got, err := application.GetImportJob(ctx, GetImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
	})
	if err != nil {
		t.Fatalf("get import job: %v", err)
	}
	if got.Status != importjob.StatusCancelledDiscarded || got.CompletedAt.IsZero() {
		t.Fatalf("expected resumed cancellation to terminalize, got %+v", got)
	}
	recoveryEvent, ok := observer.eventNamed(ports.EventImportJobRecoveryClaimed)
	if !ok {
		t.Fatalf("expected recovery claim observability event, got %+v", observer.events)
	}
	if recoveryEvent.Fields["tenant_id"] != "tenant-one" || recoveryEvent.Fields["inventory_id"] != "inventory-one" || recoveryEvent.Fields["job_id"] != job.ID.String() {
		t.Fatalf("unexpected recovery event fields: %+v", recoveryEvent.Fields)
	}
	discardEvent, ok := observer.eventNamed(ports.EventImportJobDiscardCleanupCompleted)
	if !ok {
		t.Fatalf("expected discard cleanup observability event, got %+v", observer.events)
	}
	if discardEvent.Fields["records_discarded"] != "0" || discardEvent.Fields["source_links_discarded"] != "0" {
		t.Fatalf("unexpected discard cleanup event fields: %+v", discardEvent.Fields)
	}
	for _, event := range []ports.Event{recoveryEvent, discardEvent} {
		for _, value := range event.Fields {
			if strings.Contains(value, "secret") || strings.Contains(value, "password") || strings.Contains(value, "source:") {
				t.Fatalf("import recovery/discard event leaked unsafe value: %+v", event.Fields)
			}
		}
	}
}

func TestResumeRunningImportJobsFailsRunningJobWhenSourceMaterialIsMissing(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	observer := &fakeObserver{}
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	job := importjob.NewPreviewedRecord(importjob.ID("job-one"), importjob.TenantID("tenant-one"), importjob.InventoryID("inventory-one"), importjob.PrincipalID("owner"), importjob.SourceRef{
		Type:        importjob.SourceTypeLegacyHomebox,
		Name:        "Homebox",
		Fingerprint: "sha256:test",
	}, importjob.Counts{}, nil, now)
	job.Status = importjob.StatusRunning
	job.StartedAt = now
	job.Progress = importjob.Progress{Phase: importjob.PhaseReading, Message: "Reading source", UpdatedAt: now}
	if err := store.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save running import job: %v", err)
	}
	application := New(Dependencies{
		Authorizer:        &fakeAuthorizer{},
		Tenants:           store,
		Inventories:       store,
		Audit:             store,
		ImportJobs:        store,
		ImportSourceVault: &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}},
		ImportWorker:      &fakeImportWorker{},
		IDs:               &fakeIDGenerator{ids: []string{"audit-failed"}},
		Observer:          observer,
		Clock:             fakeClock{now: now.Add(5 * time.Minute)},
	})

	resumed, err := application.ResumeRunningImportJobs(ctx, 10)
	if err != nil {
		t.Fatalf("resume import jobs: %v", err)
	}
	if resumed != 1 {
		t.Fatalf("expected missing source job to be handled as one resumed job, got %d", resumed)
	}
	got, err := application.GetImportJob(ctx, GetImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
	})
	if err != nil {
		t.Fatalf("get import job: %v", err)
	}
	if got.Status != importjob.StatusFailed || got.Progress.Message != "Import source credentials were unavailable" || got.CompletedAt.IsZero() {
		t.Fatalf("expected missing source material to fail running job, got %+v", got)
	}
	records, err := store.ListInventoryAuditRecords(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AuditRecordPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	if len(records) != 1 || records[0].Action != audit.ActionImportJobFailed {
		t.Fatalf("expected failed audit record for missing source material, got %+v", records)
	}
	event, ok := observer.eventNamed(ports.EventImportJobWorkerFailed)
	if !ok {
		t.Fatalf("expected safe worker failure event for missing source material, got %+v", observer.events)
	}
	if event.Fields["job_id"] != job.ID.String() || event.Fields["error_class"] != "missing_import_source" {
		t.Fatalf("unexpected missing source recovery failure event fields: %+v", event.Fields)
	}
	for _, value := range event.Fields {
		if strings.Contains(value, "secret") || strings.Contains(value, "password") || strings.Contains(value, "source:") {
			t.Fatalf("missing source recovery event leaked unsafe value: %+v", event.Fields)
		}
	}
}

func TestResumeRunningImportJobsFailsRunningJobWhenSourceMaterialIsUnreadable(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	observer := &fakeObserver{}
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	job := importjob.NewPreviewedRecord(importjob.ID("job-one"), importjob.TenantID("tenant-one"), importjob.InventoryID("inventory-one"), importjob.PrincipalID("owner"), importjob.SourceRef{
		Type:        importjob.SourceTypeLegacyHomebox,
		Name:        "Homebox",
		Fingerprint: "sha256:test",
	}, importjob.Counts{}, nil, now)
	job.Status = importjob.StatusRunning
	job.StartedAt = now
	job.Progress = importjob.Progress{Phase: importjob.PhaseReading, Message: "Reading source", UpdatedAt: now}
	if err := store.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save running import job: %v", err)
	}
	application := New(Dependencies{
		Authorizer:        &fakeAuthorizer{},
		Tenants:           store,
		Inventories:       store,
		Audit:             store,
		ImportJobs:        store,
		ImportSourceVault: &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}, requestErr: ports.ErrImportJobSourceUnreadable},
		ImportWorker:      &fakeImportWorker{},
		IDs:               &fakeIDGenerator{ids: []string{"audit-failed"}},
		Observer:          observer,
		Clock:             fakeClock{now: now.Add(5 * time.Minute)},
	})

	resumed, err := application.ResumeRunningImportJobs(ctx, 10)
	if err != nil {
		t.Fatalf("resume import jobs: %v", err)
	}
	if resumed != 1 {
		t.Fatalf("expected unreadable source job to be handled as one resumed job, got %d", resumed)
	}
	got, found, err := store.ImportJobByID(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), job.ID)
	if err != nil || !found {
		t.Fatalf("get import job: found=%v err=%v", found, err)
	}
	if got.Status != importjob.StatusFailed || got.Progress.Message != "Import source credentials were unavailable" || got.CompletedAt.IsZero() {
		t.Fatalf("expected unreadable source material to fail running job, got %+v", got)
	}
	event, ok := observer.eventNamed(ports.EventImportJobWorkerFailed)
	if !ok {
		t.Fatalf("expected safe worker failure event for unreadable source material, got %+v", observer.events)
	}
	if event.Fields["job_id"] != job.ID.String() || event.Fields["error_class"] != "unreadable_import_source" {
		t.Fatalf("unexpected unreadable source recovery failure event fields: %+v", event.Fields)
	}
}

func TestResumeRunningImportJobsLeavesRunningJobWhenSourceVaultLookupFails(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	observer := &fakeObserver{}
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	job := importjob.NewPreviewedRecord(importjob.ID("job-one"), importjob.TenantID("tenant-one"), importjob.InventoryID("inventory-one"), importjob.PrincipalID("owner"), importjob.SourceRef{
		Type:        importjob.SourceTypeLegacyHomebox,
		Name:        "Homebox",
		Fingerprint: "sha256:test",
	}, importjob.Counts{}, nil, now)
	job.Status = importjob.StatusRunning
	job.StartedAt = now
	job.Progress = importjob.Progress{Phase: importjob.PhaseReading, Message: "Reading source", UpdatedAt: now}
	if err := store.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save running import job: %v", err)
	}
	vaultErr := errors.New("temporary credential repository unavailable password=secret")
	application := New(Dependencies{
		Authorizer:        &fakeAuthorizer{},
		Tenants:           store,
		Inventories:       store,
		Audit:             store,
		ImportJobs:        store,
		ImportSourceVault: &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}, requestErr: vaultErr},
		ImportWorker:      &fakeImportWorker{},
		IDs:               &fakeIDGenerator{ids: []string{"audit-failed"}},
		Observer:          observer,
		Clock:             fakeClock{now: now.Add(5 * time.Minute)},
	})

	resumed, err := application.ResumeRunningImportJobs(ctx, 10)
	if !errors.Is(err, vaultErr) {
		t.Fatalf("expected source vault lookup error, got %v", err)
	}
	if resumed != 0 {
		t.Fatalf("expected no resumed jobs after source vault lookup error, got %d", resumed)
	}
	got, found, err := store.ImportJobByID(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), job.ID)
	if err != nil {
		t.Fatalf("get import job: %v", err)
	}
	if !found {
		t.Fatalf("expected running import job to remain present")
	}
	if got.Status != importjob.StatusRunning || !got.CompletedAt.IsZero() || got.Progress.Message != "Reading source" {
		t.Fatalf("expected transient source vault lookup failure to leave running job unchanged, got %+v", got)
	}
	if !got.UpdatedAt.Equal(job.UpdatedAt) || !got.Progress.UpdatedAt.Equal(job.Progress.UpdatedAt) || len(got.ProgressHistory) != len(job.ProgressHistory) {
		t.Fatalf("expected transient source vault lookup failure not to claim or mutate job, before=%+v after=%+v", job, got)
	}
	if observer.hasEvent(ports.EventImportJobWorkerFailed) {
		t.Fatalf("did not expect safe worker failure event for transient source vault lookup failure, got %+v", observer.events)
	}
	if observer.hasEvent(ports.EventImportJobRecoveryClaimed) {
		t.Fatalf("did not expect recovery claim event for transient source vault lookup failure, got %+v", observer.events)
	}
	records, err := store.ListInventoryAuditRecords(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AuditRecordPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("did not expect audit records for transient source vault lookup failure, got %+v", records)
	}
}
