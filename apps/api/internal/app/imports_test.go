package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestImportSourceInputErrorOnlySurfacesTypedUserErrors(t *testing.T) {
	raw := importSourceInputError(errors.New("password=secret token=abc internal=/tmp/source.json"))
	if !errors.Is(raw, ErrInvalidInput) {
		t.Fatalf("expected raw source error to remain invalid input, got %v", raw)
	}
	var rawDetail ImportSourceInvalidInputError
	if errors.As(raw, &rawDetail) {
		t.Fatalf("expected raw source error detail to stay hidden, got %q", rawDetail.Detail)
	}

	safe := importSourceInputError(ports.NewImportSourceUserError("Homebox URL resolves to a blocked address"))
	var safeDetail ImportSourceInvalidInputError
	if !errors.As(safe, &safeDetail) {
		t.Fatalf("expected typed source user error detail, got %v", safe)
	}
	if safeDetail.Detail != "Homebox URL resolves to a blocked address" {
		t.Fatalf("unexpected safe detail %q", safeDetail.Detail)
	}
	if !errors.Is(safe, ErrInvalidInput) {
		t.Fatalf("expected safe source error to remain invalid input, got %v", safe)
	}
}

func TestCreateImportJobStoresPreviewForInventoryHistory(t *testing.T) {
	repository := newFakeImportJobRepository()
	plan := importPlanForDurableJob("Homebox", "source:drill")
	plan.Assets = append([]importplan.Asset{{
		SourceID: "location:garage",
		Kind:     "location",
		Title:    "Garage",
		CustomFields: map[string]any{
			"homebox-source-id": "location:garage",
		},
	}}, plan.Assets...)
	source := &fakeImportSourceReader{plan: plan}
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	application := newDurableImportTestApp(repository, source, &fakeIDGenerator{ids: []string{"job-one"}}, fakeClock{now: now})

	job, err := application.CreateImportJobPreview(context.Background(), CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "request-one",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType:    string(importplan.SourceLegacyHomebox),
			BaseURL:       "https://homebox.example.test",
			Username:      "owner@example.com",
			Password:      "secret",
			IncludeImages: true,
		},
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}
	if job.ID.String() != "job-one" || job.Status != importjob.StatusPreviewed {
		t.Fatalf("unexpected job identity/status: %+v", job)
	}
	if job.Source.Fingerprint == "" {
		t.Fatalf("expected source fingerprint")
	}
	if job.Counts.Locations != 1 || job.Counts.Assets != 1 || job.Counts.Attachments != 1 {
		t.Fatalf("unexpected preview counts: %+v", job.Counts)
	}
	if len(job.Preview.Locations) != 1 || job.Preview.Locations[0].Kind != "location" {
		t.Fatalf("expected location preview sample, got %+v", job.Preview.Locations)
	}
	if len(job.Preview.Assets) != 1 || job.Preview.Assets[0].Kind != "item" {
		t.Fatalf("expected non-location preview asset sample, got %+v", job.Preview.Assets)
	}

	jobs, err := application.ListImportJobs(context.Background(), ListImportJobsInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
	})
	if err != nil {
		t.Fatalf("list import jobs: %v", err)
	}
	if len(jobs) != 1 || jobs[0].ID != job.ID {
		t.Fatalf("expected saved job in history, got %+v", jobs)
	}
	if source.lastRequest.Password != "secret" {
		t.Fatalf("expected source reader to receive request credential for preview")
	}
	if source.lastRequest.FetchAttachmentBytes {
		t.Fatalf("preview must not request attachment bytes")
	}
	if repository.jobs[job.ID].Source.Fingerprint != job.Source.Fingerprint {
		t.Fatalf("expected repository to persist source fingerprint")
	}
	vault := application.importSourceVault.(*fakeImportSourceVault)
	if len(vault.requests) != 0 {
		t.Fatalf("preview must not persist raw source material")
	}
}

func TestRemoveImportJobFromHistoryHidesJobAndRecordsProbe(t *testing.T) {
	repository := newFakeImportJobRepository()
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	observer := &fakeObserver{}
	application := newDurableImportTestApp(repository, source, &fakeIDGenerator{ids: []string{"job-one"}}, fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)})
	application.observer = observer

	job, err := application.CreateImportJobPreview(context.Background(), CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "request-one",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
			Username:   "owner@example.com",
			Password:   "secret",
		},
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}
	if _, err := application.CancelImportJob(context.Background(), CancelImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Mode:        importjob.CancellationModeKeepPartial,
	}); err != nil {
		t.Fatalf("cancel import job: %v", err)
	}
	if err := application.RemoveImportJobFromHistory(context.Background(), RemoveImportJobFromHistoryInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
	}); err != nil {
		t.Fatalf("remove import job from history: %v", err)
	}
	history, err := application.ListImportJobs(context.Background(), ListImportJobsInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
	})
	if err != nil {
		t.Fatalf("list import jobs: %v", err)
	}
	if len(history) != 0 {
		t.Fatalf("expected removed job to be hidden from history, got %+v", history)
	}
	event, ok := observer.eventNamed(ports.EventImportJobHistoryRemoved)
	if !ok {
		t.Fatalf("expected import job history removed observability event, got %+v", observer.events)
	}
	if event.Fields["job_id"] != job.ID.String() || event.Fields["inventory_id"] != "inventory-one" {
		t.Fatalf("unexpected history removed event fields: %+v", event.Fields)
	}
}

func TestRemoveImportJobFromHistoryRejectsDiscardFailedJobs(t *testing.T) {
	repository := newFakeImportJobRepository()
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	application := newDurableImportTestApp(repository, source, &fakeIDGenerator{ids: []string{"job-one"}}, fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)})

	job, err := application.CreateImportJobPreview(context.Background(), CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "request-one",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
			Username:   "owner@example.com",
			Password:   "secret",
		},
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}
	job.Status = importjob.StatusDiscardFailed
	job.UpdatedAt = job.UpdatedAt.Add(time.Minute)
	if err := repository.UpdateImportJob(context.Background(), job); err != nil {
		t.Fatalf("seed discard failed job: %v", err)
	}

	err = application.RemoveImportJobFromHistory(context.Background(), RemoveImportJobFromHistoryInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
	})
	if !errors.Is(err, ErrPrecondition) {
		t.Fatalf("expected discard-failed remove to be rejected, got %v", err)
	}
	history, err := application.ListImportJobs(context.Background(), ListImportJobsInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
	})
	if err != nil {
		t.Fatalf("list import jobs: %v", err)
	}
	if len(history) != 1 || history[0].Status != importjob.StatusDiscardFailed {
		t.Fatalf("expected discard-failed job to remain visible, got %+v", history)
	}
}

func TestCreateImportJobPreviewSanitizesUnsafeSourceMessages(t *testing.T) {
	repository := newFakeImportJobRepository()
	plan := importPlanForDurableJob("Homebox", "source:drill")
	plan.Messages = append(plan.Messages, importplan.Message{
		Code:       "bad-secret",
		Severity:   importplan.SeverityWarning,
		Summary:    "Password was rejected",
		Detail:     "Bearer token=abc123",
		SourceID:   "file:///var/lib/homebox/blob",
		SourceName: "Cordless drill",
	})
	source := &fakeImportSourceReader{plan: plan}
	application := newDurableImportTestApp(repository, source, &fakeIDGenerator{ids: []string{"job-one"}}, fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)})

	job, err := application.CreateImportJobPreview(context.Background(), CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "request-one",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
			Username:   "owner@example.com",
			Password:   "secret",
		},
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}
	message := job.Preview.Messages[0]
	if message.Summary != "" || message.Detail != "" || message.SourceID != "" || message.SourceName != "Cordless drill" {
		t.Fatalf("unsafe preview message was not sanitized: %+v", message)
	}
	persisted := repository.jobs[job.ID]
	if persisted.Messages[0].Detail != "" || persisted.Preview.Messages[0].Detail != "" {
		t.Fatalf("unsafe message persisted: messages=%+v preview=%+v", persisted.Messages, persisted.Preview.Messages)
	}
}

func TestStartImportJobRequiresNewPreviewWhenSourceFingerprintChanges(t *testing.T) {
	repository := newFakeImportJobRepository()
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	application := newDurableImportTestApp(repository, source, &fakeIDGenerator{ids: []string{"job-one"}}, fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)})

	job, err := application.CreateImportJobPreview(context.Background(), CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
			Username:   "owner@example.com",
			Password:   "secret",
		},
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}

	source.plan = importPlanForDurableJob("Homebox", "source:changed")
	_, err = application.StartImportJob(context.Background(), StartImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
			Username:   "owner@example.com",
			Password:   "secret",
		},
	})
	if !errors.Is(err, ErrPrecondition) {
		t.Fatalf("expected source fingerprint precondition, got %v", err)
	}
	if repository.jobs[job.ID].Status != importjob.StatusPreviewed {
		t.Fatalf("expected stale job to remain previewed, got %+v", repository.jobs[job.ID])
	}
}

func TestStartImportJobStoresSourceOnlyAfterFingerprintMatch(t *testing.T) {
	repository := newFakeImportJobRepository()
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	application := newDurableImportTestApp(repository, source, &fakeIDGenerator{ids: []string{"job-one"}}, fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)})

	job, err := application.CreateImportJobPreview(context.Background(), CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
			Username:   "owner@example.com",
			Password:   "secret",
		},
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}
	_, err = application.StartImportJob(context.Background(), StartImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
			Username:   "owner@example.com",
			Password:   "secret",
		},
	})
	if err != nil {
		t.Fatalf("start import job: %v", err)
	}
	vault := application.importSourceVault.(*fakeImportSourceVault)
	if vault.requests[job.ID].Password != "secret" {
		t.Fatalf("expected start to store source material for worker")
	}
	if !vault.requests[job.ID].FetchAttachmentBytes {
		t.Fatalf("expected stored worker source request to fetch attachment bytes")
	}
}

func TestStartImportJobDoesNotStoreSourceWhenStartTransitionLosesRace(t *testing.T) {
	repository := newFakeImportJobRepository()
	repository.failNextPreviewedStatusUpdate = true
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	vault := &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}}
	application := New(Dependencies{
		Authorizer:        &fakeAuthorizer{},
		Tenants:           &fakeTenantRepository{exists: true},
		Inventories:       &fakeInventoryRepository{items: []inventory.Inventory{activeInventory("inventory-one")}},
		ImportSources:     source,
		ImportJobs:        repository,
		ImportSourceVault: vault,
		ImportWorker:      &fakeImportWorker{},
		IDs:               &fakeIDGenerator{ids: []string{"job-one"}},
		Clock:             fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	job, err := application.CreateImportJobPreview(context.Background(), CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
			Username:   "owner@example.com",
			Password:   "winner",
		},
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}
	_, err = application.StartImportJob(context.Background(), StartImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
			Username:   "owner@example.com",
			Password:   "loser",
		},
	})
	if !errors.Is(err, ErrPrecondition) {
		t.Fatalf("expected start precondition after losing race, got %v", err)
	}
	if len(vault.requests) != 0 {
		t.Fatalf("losing start wrote source credentials: %+v", vault.requests)
	}
}

func TestStartImportJobRequiresWorkerBeforeRunningTransition(t *testing.T) {
	repository := newFakeImportJobRepository()
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	application := New(Dependencies{
		Authorizer:        &fakeAuthorizer{},
		Tenants:           &fakeTenantRepository{exists: true},
		Inventories:       &fakeInventoryRepository{items: []inventory.Inventory{activeInventory("inventory-one")}},
		ImportSources:     source,
		ImportJobs:        repository,
		ImportSourceVault: &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}},
		IDs:               &fakeIDGenerator{ids: []string{"job-one"}},
		Clock:             fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	job, err := application.CreateImportJobPreview(context.Background(), CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}
	_, err = application.StartImportJob(context.Background(), StartImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected worker configuration error, got %v", err)
	}
	if repository.jobs[job.ID].Status != importjob.StatusPreviewed {
		t.Fatalf("expected missing worker to leave job previewed, got %+v", repository.jobs[job.ID])
	}
}

func TestStartImportJobTerminalizesWhenStartAuditFailsAfterRunningTransition(t *testing.T) {
	repository := newFakeImportJobRepository()
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	auditRepository := &failingAfterAuditRepository{failAfter: 1, err: errors.New("audit write failed")}
	vault := &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}}
	worker := &fakeImportWorker{}
	application := New(Dependencies{
		Authorizer:        &fakeAuthorizer{},
		Tenants:           &fakeTenantRepository{exists: true},
		Inventories:       &fakeInventoryRepository{items: []inventory.Inventory{activeInventory("inventory-one")}},
		Audit:             auditRepository,
		ImportSources:     source,
		ImportJobs:        repository,
		ImportSourceVault: vault,
		ImportWorker:      worker,
		IDs:               &fakeIDGenerator{ids: []string{"job-one", "audit-preview"}},
		Clock:             fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	job, err := application.CreateImportJobPreview(context.Background(), CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}
	_, err = application.StartImportJob(context.Background(), StartImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if err == nil {
		t.Fatalf("expected audit failure")
	}
	stored := repository.jobs[job.ID]
	if stored.Status != importjob.StatusFailed || stored.Progress.Message != "Import could not be started" {
		t.Fatalf("expected start failure to terminalize job, got %+v", stored)
	}
	if len(worker.executed) != 0 {
		t.Fatalf("worker should not execute after start audit failure: %+v", worker.executed)
	}
	if _, found := vault.requests[job.ID]; found {
		t.Fatalf("expected stored source material to be cleaned up after start failure")
	}
}

func TestStartImportJobTerminalizesCleansCredentialsAndAuditsWhenWorkerDispatchFails(t *testing.T) {
	repository := newFakeImportJobRepository()
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	auditRepository := &fakeAuditRepository{}
	vault := &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}}
	application := New(Dependencies{
		Authorizer:        &fakeAuthorizer{},
		Tenants:           &fakeTenantRepository{exists: true},
		Inventories:       &fakeInventoryRepository{items: []inventory.Inventory{activeInventory("inventory-one")}},
		Audit:             auditRepository,
		ImportSources:     source,
		ImportJobs:        repository,
		ImportSourceVault: vault,
		ImportWorker:      failingImportWorker{err: errors.New("worker dispatch failed")},
		IDs:               &fakeIDGenerator{ids: []string{"job-one", "audit-preview", "audit-start", "audit-failed"}},
		Clock:             fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	job, err := application.CreateImportJobPreview(context.Background(), CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}
	_, err = application.StartImportJob(context.Background(), StartImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if err == nil {
		t.Fatalf("expected worker dispatch failure")
	}
	stored := repository.jobs[job.ID]
	if stored.Status != importjob.StatusFailed || stored.Progress.Message != "Import could not be started" {
		t.Fatalf("expected worker dispatch failure to terminalize job, got %+v", stored)
	}
	if _, found := vault.requests[job.ID]; found {
		t.Fatalf("expected worker dispatch failure to clean stored source material")
	}
	if !auditRepository.hasAction(audit.ActionImportJobStarted) ||
		!auditRepository.hasAction(audit.ActionImportJobFailed) ||
		!auditRepository.hasAction(audit.ActionImportJobCredentialCleaned) {
		t.Fatalf("expected started, failed, and credential-cleaned audit records, got %+v", auditRepository.items)
	}
}

func TestImportJobLifecycleWritesAuditRecords(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
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
		Authorizer:        &fakeAuthorizer{},
		Tenants:           store,
		Inventories:       store,
		Audit:             store,
		ImportSources:     source,
		ImportJobs:        store,
		ImportSourceVault: &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}},
		ImportLinks:       store,
		ImportWorker:      &fakeImportWorker{},
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "audit-preview", "audit-start", "audit-complete", "audit-credential-cleaned",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	job := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if job.Status != importjob.StatusSucceeded {
		t.Fatalf("expected succeeded job, got %+v", job)
	}
	if len(job.ProgressHistory) < 3 ||
		job.ProgressHistory[0].Phase != importjob.PhaseReady ||
		job.ProgressHistory[1].Phase != importjob.PhaseReading ||
		job.ProgressHistory[len(job.ProgressHistory)-1].Phase != importjob.PhaseTerminal {
		t.Fatalf("expected durable job phase history from preview to terminal, got %+v", job.ProgressHistory)
	}
	records, err := store.ListInventoryAuditRecords(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AuditRecordPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	actions := map[audit.Action]audit.Record{}
	actionCounts := map[audit.Action]int{}
	for _, record := range records {
		if record.TargetType == audit.TargetImportJob {
			actions[record.Action] = record
			actionCounts[record.Action]++
		}
	}
	for _, action := range []audit.Action{audit.ActionImportJobPreviewed, audit.ActionImportJobStarted, audit.ActionImportJobCompleted, audit.ActionImportJobCredentialCleaned} {
		record, ok := actions[action]
		if !ok {
			t.Fatalf("expected import job audit action %s in %+v", action, records)
		}
		if record.TargetID != "job-one" || record.Metadata["source_type"] != string(importplan.SourceLegacyHomebox) {
			t.Fatalf("unexpected audit record for %s: %+v", action, record)
		}
		if strings.Contains(record.Metadata["source_type"], "secret") || strings.Contains(record.Metadata["import_job_status"], "secret") {
			t.Fatalf("audit metadata leaked secret-looking value: %+v", record.Metadata)
		}
		if actionCounts[action] != 1 {
			t.Fatalf("expected one import job audit action %s, got %d in %+v", action, actionCounts[action], records)
		}
	}
}

func TestImportJobAuditRecordsPreserveRequestIDs(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	application := New(Dependencies{
		Authorizer:        &fakeAuthorizer{},
		Tenants:           store,
		Inventories:       store,
		Audit:             store,
		ImportSources:     source,
		ImportJobs:        store,
		ImportSourceVault: &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}},
		ImportLinks:       store,
		ImportWorker:      &fakeImportWorker{},
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "audit-preview", "audit-cancel-requested", "audit-cancelled", "audit-history-removed",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	job, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "preview-request",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
			Username:   "owner@example.com",
			Password:   "secret",
		},
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}
	if _, err := application.CancelImportJob(ctx, CancelImportJobInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "cancel-request",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Mode:        importjob.CancellationModeKeepPartial,
	}); err != nil {
		t.Fatalf("cancel import job: %v", err)
	}
	if err := application.RemoveImportJobFromHistory(ctx, RemoveImportJobFromHistoryInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "remove-request",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
	}); err != nil {
		t.Fatalf("remove import job from history: %v", err)
	}

	records, err := store.ListInventoryAuditRecords(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AuditRecordPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	requestIDs := map[audit.Action]string{}
	for _, record := range records {
		if record.TargetType == audit.TargetImportJob {
			requestIDs[record.Action] = record.RequestID
		}
	}
	expected := map[audit.Action]string{
		audit.ActionImportJobPreviewed:             "preview-request",
		audit.ActionImportJobCancellationRequested: "cancel-request",
		audit.ActionImportJobCancelled:             "cancel-request",
		audit.ActionImportJobHistoryRemoved:        "remove-request",
	}
	for action, requestID := range expected {
		if requestIDs[action] != requestID {
			t.Fatalf("expected request ID %q for %s, got %q in records %+v", requestID, action, requestIDs[action], records)
		}
	}
}

func TestRunningImportCancellationCompletionAuditsCancelRequestID(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	application := New(Dependencies{
		Authorizer:        &fakeAuthorizer{},
		Tenants:           store,
		Inventories:       store,
		Audit:             store,
		ImportSources:     source,
		ImportJobs:        store,
		ImportSourceVault: &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}},
		ImportLinks:       store,
		ImportWorker:      &fakeImportWorker{},
		IDs: &fakeIDGenerator{ids: []string{
			"job-one", "audit-preview", "audit-start", "audit-cancel-requested", "audit-cancelled", "audit-credential-cleaned",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})
	importSource := ImportSourceInput{
		SourceType: string(importplan.SourceLegacyHomebox),
		BaseURL:    "https://homebox.example.test",
		Username:   "owner@example.com",
		Password:   "secret",
	}

	job, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "preview-request",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source:      importSource,
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}
	started, err := application.StartImportJob(ctx, StartImportJobInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "start-request",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Source:      importSource,
	})
	if err != nil {
		t.Fatalf("start import job: %v", err)
	}
	if _, err := application.CancelImportJob(ctx, CancelImportJobInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "cancel-request",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Mode:        importjob.CancellationModeKeepPartial,
	}); err != nil {
		t.Fatalf("cancel import job: %v", err)
	}
	if _, err := application.ExecuteImportJob(ctx, ports.ImportJobCommand{
		Principal:   durableImportPrincipal(),
		RequestID:   "start-request",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       started.ID,
	}); err != nil {
		t.Fatalf("execute cancelled import job: %v", err)
	}

	records, err := store.ListInventoryAuditRecords(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AuditRecordPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	requestIDs := map[audit.Action]string{}
	for _, record := range records {
		if record.TargetType == audit.TargetImportJob {
			requestIDs[record.Action] = record.RequestID
		}
	}
	if requestIDs[audit.ActionImportJobStarted] != "start-request" {
		t.Fatalf("expected start audit to keep start request ID, got %q in %+v", requestIDs[audit.ActionImportJobStarted], records)
	}
	if requestIDs[audit.ActionImportJobCancellationRequested] != "cancel-request" {
		t.Fatalf("expected cancellation requested audit to keep cancel request ID, got %q in %+v", requestIDs[audit.ActionImportJobCancellationRequested], records)
	}
	if requestIDs[audit.ActionImportJobCancelled] != "cancel-request" {
		t.Fatalf("expected terminal cancellation audit to use cancel request ID, got %q in %+v", requestIDs[audit.ActionImportJobCancelled], records)
	}
}

func TestCancelImportJobRejectsOverlongRequestIDBeforeMutation(t *testing.T) {
	repository := newFakeImportJobRepository()
	application := newDurableImportTestApp(
		repository,
		&fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")},
		&fakeIDGenerator{ids: []string{"job-one", "audit-preview"}},
		fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	)

	job, err := application.CreateImportJobPreview(context.Background(), CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "preview-request",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
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

	_, err = application.CancelImportJob(context.Background(), CancelImportJobInput{
		Principal:   durableImportPrincipal(),
		RequestID:   strings.Repeat("x", maxImportRequestIDLength+1),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Mode:        importjob.CancellationModeKeepPartial,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input for overlong request ID, got %v", err)
	}
	unchanged, found, err := repository.ImportJobByID(context.Background(), job.TenantID, job.InventoryID, job.ID)
	if err != nil || !found {
		t.Fatalf("read import job after rejected cancel found=%t err=%v", found, err)
	}
	if unchanged.Status != importjob.StatusRunning || unchanged.CancellationRequestID != "" || unchanged.CancellationMode != "" {
		t.Fatalf("expected rejected cancel not to mutate running job, got %+v", unchanged)
	}
}

func TestImportJobLifecycleDoesNotAuditCredentialCleanupWhenVaultReportsNoDeletion(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	source := &fakeImportSourceReader{plan: importplan.Plan{
		Source: importplan.SourceSummary{
			Type:        importplan.SourceLegacyHomebox,
			Name:        "Homebox",
			BaseURL:     "https://homebox.example.test",
			ImageImport: "disabled",
		},
	}}
	application := New(Dependencies{
		Authorizer:        &fakeAuthorizer{},
		Tenants:           store,
		Inventories:       store,
		Audit:             store,
		ImportSources:     source,
		ImportJobs:        store,
		ImportSourceVault: &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}, deleteReportsMissing: true},
		ImportLinks:       store,
		ImportWorker:      &fakeImportWorker{},
		IDs:               &fakeIDGenerator{ids: []string{"job-one", "audit-preview", "audit-start", "audit-complete"}},
		Clock:             fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	job := createStartAndExecuteImportJob(t, ctx, application, importjob.ID("job-one"))
	if job.Status != importjob.StatusSucceeded {
		t.Fatalf("expected succeeded job, got %+v", job)
	}
	records, err := store.ListInventoryAuditRecords(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AuditRecordPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	for _, record := range records {
		if record.Action == audit.ActionImportJobCredentialCleaned {
			t.Fatalf("did not expect credential cleanup audit without a deleted credential, got %+v", records)
		}
	}
}

func TestVacuumImportJobCredentialsWritesSafeAuditRecord(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	seedDurableImportMemoryInventory(t, ctx, store)
	job := importjob.NewPreviewedRecord(importjob.ID("job-one"), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), identity.PrincipalID("owner"), importjob.SourceRef{
		Type:        importplan.SourceLegacyHomebox,
		Name:        "Homebox",
		Fingerprint: "sha256:test",
	}, importjob.Counts{}, nil, time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))
	job.Status = importjob.StatusSucceeded
	if err := store.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save import job: %v", err)
	}
	application := New(Dependencies{
		Authorizer:  &fakeAuthorizer{},
		Tenants:     store,
		Inventories: store,
		Audit:       store,
		ImportJobs:  store,
		ImportSourceVault: &fakeImportSourceVault{
			requests: map[importjob.ID]ports.ImportSourceRequest{},
			vacuumScopes: []ports.ImportJobSourceScope{{
				TenantID:    tenant.ID("tenant-one"),
				InventoryID: inventory.InventoryID("inventory-one"),
				JobID:       importjob.ID("job-one"),
			}},
		},
		IDs:   &fakeIDGenerator{ids: []string{"audit-credential-cleaned"}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 5, 0, 0, time.UTC)},
	})

	deleted, err := application.VacuumImportJobCredentials(ctx)
	if err != nil {
		t.Fatalf("vacuum import job credentials: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected one cleaned credential, got %d", deleted)
	}
	records, err := store.ListInventoryAuditRecords(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AuditRecordPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	var found audit.Record
	for _, record := range records {
		if record.Action == audit.ActionImportJobCredentialCleaned {
			found = record
		}
	}
	if found.ID.String() == "" || found.TargetID != "job-one" || found.Metadata["source_type"] != string(importplan.SourceLegacyHomebox) {
		t.Fatalf("expected safe credential cleanup audit record, got %+v", records)
	}
	for _, value := range found.Metadata {
		if strings.Contains(value, "secret") || strings.Contains(value, "password") {
			t.Fatalf("credential cleanup audit metadata leaked unsafe value: %+v", found.Metadata)
		}
	}
}

func TestVacuumImportJobCredentialsFailureRecordsSafeEvent(t *testing.T) {
	ctx := context.Background()
	observer := &fakeObserver{}
	application := New(Dependencies{
		Authorizer: &fakeAuthorizer{},
		ImportSourceVault: &fakeImportSourceVault{
			requests:  map[importjob.ID]ports.ImportSourceRequest{},
			vacuumErr: errors.New("password=secret /tmp/provider-key ciphertext=abc"),
		},
		Observer: observer,
		Clock:    fakeClock{now: time.Date(2026, 7, 6, 12, 5, 0, 0, time.UTC)},
	})

	if _, err := application.VacuumImportJobCredentials(ctx); err == nil {
		t.Fatalf("expected vacuum failure")
	}
	event, ok := observer.eventNamed(ports.EventImportJobCredentialVacuumFailed)
	if !ok {
		t.Fatalf("expected credential vacuum failed event, got %+v", observer.events)
	}
	if event.Fields["error_class"] != "credential_vacuum_failed" {
		t.Fatalf("expected safe error class, got %+v", event.Fields)
	}
	for _, value := range event.Fields {
		if strings.Contains(value, "secret") || strings.Contains(value, "/tmp") || strings.Contains(value, "ciphertext") {
			t.Fatalf("credential vacuum event leaked unsafe value: %+v", event.Fields)
		}
	}
}

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
			SourceType: string(importplan.SourceLegacyHomebox),
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
			SourceType: string(importplan.SourceLegacyHomebox),
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
			SourceType: string(importplan.SourceLegacyHomebox),
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
			SourceType: string(importplan.SourceLegacyHomebox),
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
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	job := importjob.NewPreviewedRecord(importjob.ID("job-one"), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), identity.PrincipalID("owner"), importjob.SourceRef{
		Type:        importplan.SourceLegacyHomebox,
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
}

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

func TestNormalizedImportPlanForJobDoesNotMutateAttachmentContent(t *testing.T) {
	application := New(Dependencies{})
	plan := importPlanForDurableJob("Homebox", "source:drill")
	plan.Attachments[0].Content = pngAttachmentBytes()

	normalized := application.normalizedImportPlanForJob(context.Background(), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), plan)
	if len(normalized.Attachments[0].Content) != 0 {
		t.Fatalf("expected normalized plan to strip attachment bytes")
	}
	if len(plan.Attachments[0].Content) == 0 {
		t.Fatalf("expected original plan attachment bytes to remain available for apply")
	}
}

func TestSourceFingerprintIgnoresAttachmentOrder(t *testing.T) {
	plan := importPlanForDurableJob("Homebox", "source:drill")
	plan.Attachments = []importplan.Attachment{
		{SourceID: "attachment:two", AssetSourceID: "source:drill", FileName: "two.jpg", ContentType: "image/jpeg", Primary: false},
		{SourceID: "attachment:one", AssetSourceID: "source:drill", FileName: "one.jpg", ContentType: "image/jpeg", Primary: true},
	}
	swapped := plan
	swapped.Attachments = []importplan.Attachment{plan.Attachments[1], plan.Attachments[0]}

	first, err := sourceFingerprint(plan)
	if err != nil {
		t.Fatalf("first fingerprint: %v", err)
	}
	second, err := sourceFingerprint(swapped)
	if err != nil {
		t.Fatalf("second fingerprint: %v", err)
	}
	if first != second {
		t.Fatalf("expected attachment order-insensitive fingerprint, got %q and %q", first, second)
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
	job := importjob.NewPreviewedRecord(importjob.ID("job-one"), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), identity.PrincipalID("owner"), importjob.SourceRef{
		Type:        importplan.SourceLegacyHomebox,
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
	job := importjob.NewPreviewedRecord(importjob.ID("job-one"), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), identity.PrincipalID("owner"), importjob.SourceRef{
		Type:        importplan.SourceLegacyHomebox,
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
			SourceType: string(importplan.SourceLegacyHomebox),
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
	job := importjob.NewPreviewedRecord(importjob.ID("job-one"), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), identity.PrincipalID("owner"), importjob.SourceRef{
		Type:        importplan.SourceLegacyHomebox,
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

func createStartAndExecuteImportJob(t *testing.T, ctx context.Context, application App, expectedJobID importjob.ID) importjob.Record {
	t.Helper()
	job, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	})
	if err != nil {
		t.Fatalf("create preview: %v", err)
	}
	if job.ID != expectedJobID {
		t.Fatalf("expected job %q, got %q", expectedJobID, job.ID)
	}
	return startAndExecuteImportJob(t, ctx, application, job.ID)
}

func startAndExecuteImportJob(t *testing.T, ctx context.Context, application App, jobID importjob.ID) importjob.Record {
	t.Helper()
	if _, err := application.StartImportJob(ctx, StartImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       jobID,
		Source: ImportSourceInput{
			SourceType: string(importplan.SourceLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
		},
	}); err != nil {
		t.Fatalf("start import job: %v", err)
	}
	job, err := application.ExecuteImportJob(ctx, ports.ImportJobCommand{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       jobID,
	})
	if err != nil {
		t.Fatalf("execute import job: %v", err)
	}
	return job
}

func seedDurableImportMemoryInventory(t *testing.T, ctx context.Context, store *memory.Store) {
	t.Helper()
	tenantName, ok := tenant.NewName("Home")
	if !ok {
		t.Fatalf("invalid tenant name")
	}
	tenantRecord, ok := tenant.NewTenant(tenant.ID("tenant-one"), tenantName, tenant.LifecycleStateActive)
	if !ok {
		t.Fatalf("invalid tenant")
	}
	if err := store.SaveTenant(ctx, tenantRecord); err != nil {
		t.Fatalf("save tenant: %v", err)
	}
	if err := store.SaveInventory(ctx, activeInventory("inventory-one")); err != nil {
		t.Fatalf("save inventory: %v", err)
	}
}

func durableImportPrincipal() identity.Principal {
	return identity.Principal{ID: identity.PrincipalID("owner"), Email: "owner@example.com"}
}

func importPlanForDurableJob(name string, sourceID string) importplan.Plan {
	return importplan.Plan{
		Source: importplan.SourceSummary{
			Type:        importplan.SourceLegacyHomebox,
			Name:        name,
			BaseURL:     "https://homebox.example.test",
			Version:     "0.24.0",
			ImageImport: "enabled",
		},
		Fields: []importplan.FieldDefinition{{Key: "homebox-source-id", DisplayName: "Homebox Source ID", Type: "text"}},
		Assets: []importplan.Asset{{
			SourceID: sourceID,
			Kind:     "item",
			Title:    "Cordless drill",
			CustomFields: map[string]any{
				"homebox-source-id": sourceID,
			},
		}},
		Attachments: []importplan.Attachment{{
			SourceID:      "attachment:" + sourceID,
			AssetSourceID: sourceID,
			FileName:      "drill.jpg",
			ContentType:   "image/jpeg",
			SizeBytes:     42,
			Primary:       true,
			Content:       []byte("jpeg bytes"),
		}},
	}
}

func newDurableImportTestApp(repository *fakeImportJobRepository, source ports.ImportSourceReader, ids ports.IDGenerator, clock ports.Clock) App {
	return New(Dependencies{
		Authorizer:    &fakeAuthorizer{},
		Tenants:       &fakeTenantRepository{exists: true},
		Inventories:   &fakeInventoryRepository{items: []inventory.Inventory{activeInventory("inventory-one")}},
		ImportSources: source,
		ImportJobs:    repository,
		ImportSourceVault: &fakeImportSourceVault{
			requests: map[importjob.ID]ports.ImportSourceRequest{},
		},
		ImportWorker: &fakeImportWorker{},
		IDs:          ids,
		Clock:        clock,
	})
}

func activeInventory(id string) inventory.Inventory {
	name, ok := inventory.NewName("Inventory")
	if !ok {
		panic("invalid test inventory name")
	}
	item, ok := inventory.NewInventory(inventory.InventoryID(id), inventory.TenantID("tenant-one"), name, inventory.LifecycleStateActive)
	if !ok {
		panic("invalid test inventory")
	}
	return item
}

type fakeImportSourceReader struct {
	plan        importplan.Plan
	lastRequest ports.ImportSourceRequest
}

func (f *fakeImportSourceReader) ReadImportPlan(_ context.Context, request ports.ImportSourceRequest) (importplan.Plan, error) {
	f.lastRequest = request
	return f.plan, nil
}

type phaseAwareImportSourceReader struct {
	preview         importplan.Plan
	apply           importplan.Plan
	previewRequests int
	applyRequests   int
}

func (f *phaseAwareImportSourceReader) ReadImportPlan(_ context.Context, request ports.ImportSourceRequest) (importplan.Plan, error) {
	if request.FetchAttachmentBytes {
		f.applyRequests++
		return f.apply, nil
	}
	f.previewRequests++
	return f.preview, nil
}

type fakeImportJobRepository struct {
	jobs                           map[importjob.ID]importjob.Record
	cancelDuringNextProgressUpdate bool
	failNextPreviewedStatusUpdate  bool
}

func newFakeImportJobRepository() *fakeImportJobRepository {
	return &fakeImportJobRepository{jobs: map[importjob.ID]importjob.Record{}}
}

func (f *fakeImportJobRepository) SaveImportJob(_ context.Context, job importjob.Record) error {
	f.jobs[job.ID] = job
	return nil
}

func (f *fakeImportJobRepository) ImportJobByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (importjob.Record, bool, error) {
	job, ok := f.jobs[jobID]
	if !ok || job.TenantID != tenantID || job.InventoryID != inventoryID || !job.HistoryRemovedAt.IsZero() {
		return importjob.Record{}, false, nil
	}
	return job, true, nil
}

func (f *fakeImportJobRepository) ListImportJobs(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, _ ports.ImportJobPageRequest) ([]importjob.Record, error) {
	var jobs []importjob.Record
	for _, job := range f.jobs {
		if job.TenantID == tenantID && job.InventoryID == inventoryID && job.HistoryRemovedAt.IsZero() {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

func (f *fakeImportJobRepository) ListImportJobsByStatus(_ context.Context, page ports.ImportJobStatusPageRequest) ([]importjob.Record, error) {
	var jobs []importjob.Record
	for _, job := range f.jobs {
		if job.Status == page.Status && job.HistoryRemovedAt.IsZero() {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

func (f *fakeImportJobRepository) UpdateImportJob(_ context.Context, job importjob.Record) error {
	current, ok := f.jobs[job.ID]
	if !ok || current.TenantID != job.TenantID || current.InventoryID != job.InventoryID || !current.HistoryRemovedAt.IsZero() {
		return ports.ErrConflict
	}
	job.HistoryRemovedAt = current.HistoryRemovedAt
	f.jobs[job.ID] = job
	return nil
}

func (f *fakeImportJobRepository) MarkImportJobHistoryRemoved(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, removedAt time.Time, expectedUpdatedAt time.Time) (bool, error) {
	current, ok := f.jobs[jobID]
	if !ok || current.TenantID != tenantID || current.InventoryID != inventoryID || !current.UpdatedAt.Equal(expectedUpdatedAt) || !current.HistoryRemovedAt.IsZero() {
		return false, nil
	}
	current.HistoryRemovedAt = removedAt
	current.UpdatedAt = removedAt
	f.jobs[jobID] = current
	return true, nil
}

func (f *fakeImportJobRepository) UpdateImportJobIfStatus(_ context.Context, job importjob.Record, expected importjob.Status) (bool, error) {
	current, ok := f.jobs[job.ID]
	if !ok || current.Status != expected || current.TenantID != job.TenantID || current.InventoryID != job.InventoryID || !current.HistoryRemovedAt.IsZero() {
		return false, nil
	}
	if f.failNextPreviewedStatusUpdate && expected == importjob.StatusPreviewed {
		f.failNextPreviewedStatusUpdate = false
		return false, nil
	}
	f.jobs[job.ID] = job
	return true, nil
}

func (f *fakeImportJobRepository) UpdateImportJobProgress(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, progress importjob.Progress, expectedUpdatedAt time.Time) (bool, error) {
	current, ok := f.jobs[jobID]
	if !ok || current.TenantID != tenantID || current.InventoryID != inventoryID || !current.HistoryRemovedAt.IsZero() {
		return false, nil
	}
	if f.cancelDuringNextProgressUpdate {
		current.Status = importjob.StatusCancelRequested
		current.CancellationMode = importjob.CancellationModeDiscardPartial
		current.UpdatedAt = expectedUpdatedAt.Add(time.Nanosecond)
		f.jobs[jobID] = current
		f.cancelDuringNextProgressUpdate = false
		return false, nil
	}
	if !current.UpdatedAt.Equal(expectedUpdatedAt) {
		return false, nil
	}
	current.Progress = progress
	current.UpdatedAt = progress.UpdatedAt
	f.jobs[jobID] = current
	return true, nil
}

func (f *fakeImportJobRepository) ClaimImportJob(_ context.Context, job importjob.Record, expectedUpdatedAt time.Time) (bool, error) {
	current, ok := f.jobs[job.ID]
	if !ok || current.TenantID != job.TenantID || current.InventoryID != job.InventoryID || !current.UpdatedAt.Equal(expectedUpdatedAt) || !current.HistoryRemovedAt.IsZero() {
		return false, nil
	}
	f.jobs[job.ID] = job
	return true, nil
}

type recordingImportJobRepository struct {
	delegate   ports.ImportJobRepository
	progresses []importjob.Progress
}

func (r *recordingImportJobRepository) SaveImportJob(ctx context.Context, job importjob.Record) error {
	return r.delegate.SaveImportJob(ctx, job)
}

func (r *recordingImportJobRepository) ImportJobByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (importjob.Record, bool, error) {
	return r.delegate.ImportJobByID(ctx, tenantID, inventoryID, jobID)
}

func (r *recordingImportJobRepository) ListImportJobs(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.ImportJobPageRequest) ([]importjob.Record, error) {
	return r.delegate.ListImportJobs(ctx, tenantID, inventoryID, page)
}

func (r *recordingImportJobRepository) ListImportJobsByStatus(ctx context.Context, page ports.ImportJobStatusPageRequest) ([]importjob.Record, error) {
	return r.delegate.ListImportJobsByStatus(ctx, page)
}

func (r *recordingImportJobRepository) UpdateImportJobIfStatus(ctx context.Context, job importjob.Record, expected importjob.Status) (bool, error) {
	r.progresses = append(r.progresses, job.Progress)
	return r.delegate.UpdateImportJobIfStatus(ctx, job, expected)
}

func (r *recordingImportJobRepository) UpdateImportJobProgress(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, progress importjob.Progress, expectedUpdatedAt time.Time) (bool, error) {
	r.progresses = append(r.progresses, progress)
	return r.delegate.UpdateImportJobProgress(ctx, tenantID, inventoryID, jobID, progress, expectedUpdatedAt)
}

func (r *recordingImportJobRepository) ClaimImportJob(ctx context.Context, job importjob.Record, expectedUpdatedAt time.Time) (bool, error) {
	r.progresses = append(r.progresses, job.Progress)
	return r.delegate.ClaimImportJob(ctx, job, expectedUpdatedAt)
}

func (r *recordingImportJobRepository) MarkImportJobHistoryRemoved(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, removedAt time.Time, expectedUpdatedAt time.Time) (bool, error) {
	return r.delegate.MarkImportJobHistoryRemoved(ctx, tenantID, inventoryID, jobID, removedAt, expectedUpdatedAt)
}

func (r *recordingImportJobRepository) UpdateImportJob(ctx context.Context, job importjob.Record) error {
	r.progresses = append(r.progresses, job.Progress)
	return r.delegate.UpdateImportJob(ctx, job)
}

func (r *recordingImportJobRepository) sawProgress(phase importjob.Phase, done int, total int) bool {
	for _, progress := range r.progresses {
		if progress.Phase == phase && progress.Done == done && progress.Total == total {
			return true
		}
	}
	return false
}

type terminalRaceImportJobRepository struct {
	delegate ports.ImportJobRepository
	raced    bool
}

func (r *terminalRaceImportJobRepository) SaveImportJob(ctx context.Context, job importjob.Record) error {
	return r.delegate.SaveImportJob(ctx, job)
}

func (r *terminalRaceImportJobRepository) ImportJobByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (importjob.Record, bool, error) {
	return r.delegate.ImportJobByID(ctx, tenantID, inventoryID, jobID)
}

func (r *terminalRaceImportJobRepository) ListImportJobs(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.ImportJobPageRequest) ([]importjob.Record, error) {
	return r.delegate.ListImportJobs(ctx, tenantID, inventoryID, page)
}

func (r *terminalRaceImportJobRepository) ListImportJobsByStatus(ctx context.Context, page ports.ImportJobStatusPageRequest) ([]importjob.Record, error) {
	return r.delegate.ListImportJobsByStatus(ctx, page)
}

func (r *terminalRaceImportJobRepository) UpdateImportJobIfStatus(ctx context.Context, job importjob.Record, expected importjob.Status) (bool, error) {
	if !r.raced && expected == importjob.StatusRunning && job.Status == importjob.StatusSucceeded {
		current, found, err := r.delegate.ImportJobByID(ctx, job.TenantID, job.InventoryID, job.ID)
		if err != nil || !found {
			return false, err
		}
		current.Status = importjob.StatusCancelRequested
		current.CancellationMode = importjob.CancellationModeKeepPartial
		current.Progress.Message = "Cancellation requested"
		current.UpdatedAt = current.UpdatedAt.Add(time.Nanosecond)
		current.Progress.UpdatedAt = current.UpdatedAt
		if err := r.delegate.UpdateImportJob(ctx, current); err != nil {
			return false, err
		}
		r.raced = true
		return false, nil
	}
	return r.delegate.UpdateImportJobIfStatus(ctx, job, expected)
}

func (r *terminalRaceImportJobRepository) UpdateImportJobProgress(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, progress importjob.Progress, expectedUpdatedAt time.Time) (bool, error) {
	return r.delegate.UpdateImportJobProgress(ctx, tenantID, inventoryID, jobID, progress, expectedUpdatedAt)
}

func (r *terminalRaceImportJobRepository) ClaimImportJob(ctx context.Context, job importjob.Record, expectedUpdatedAt time.Time) (bool, error) {
	return r.delegate.ClaimImportJob(ctx, job, expectedUpdatedAt)
}

func (r *terminalRaceImportJobRepository) MarkImportJobHistoryRemoved(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, removedAt time.Time, expectedUpdatedAt time.Time) (bool, error) {
	return r.delegate.MarkImportJobHistoryRemoved(ctx, tenantID, inventoryID, jobID, removedAt, expectedUpdatedAt)
}

func (r *terminalRaceImportJobRepository) UpdateImportJob(ctx context.Context, job importjob.Record) error {
	return r.delegate.UpdateImportJob(ctx, job)
}

var errTerminalImportUpdateFailed = errors.New("terminal import update failed")

type failingTerminalImportJobRepository struct {
	delegate ports.ImportJobRepository
}

func (r *failingTerminalImportJobRepository) SaveImportJob(ctx context.Context, job importjob.Record) error {
	return r.delegate.SaveImportJob(ctx, job)
}

func (r *failingTerminalImportJobRepository) ImportJobByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (importjob.Record, bool, error) {
	return r.delegate.ImportJobByID(ctx, tenantID, inventoryID, jobID)
}

func (r *failingTerminalImportJobRepository) ListImportJobs(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.ImportJobPageRequest) ([]importjob.Record, error) {
	return r.delegate.ListImportJobs(ctx, tenantID, inventoryID, page)
}

func (r *failingTerminalImportJobRepository) ListImportJobsByStatus(ctx context.Context, page ports.ImportJobStatusPageRequest) ([]importjob.Record, error) {
	return r.delegate.ListImportJobsByStatus(ctx, page)
}

func (r *failingTerminalImportJobRepository) UpdateImportJobIfStatus(ctx context.Context, job importjob.Record, expected importjob.Status) (bool, error) {
	if expected == importjob.StatusCancelRequested && (job.Status == importjob.StatusCancelledDiscarded || job.Status == importjob.StatusDiscardFailed) {
		return false, errTerminalImportUpdateFailed
	}
	return r.delegate.UpdateImportJobIfStatus(ctx, job, expected)
}

func (r *failingTerminalImportJobRepository) UpdateImportJobProgress(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, progress importjob.Progress, expectedUpdatedAt time.Time) (bool, error) {
	return r.delegate.UpdateImportJobProgress(ctx, tenantID, inventoryID, jobID, progress, expectedUpdatedAt)
}

func (r *failingTerminalImportJobRepository) ClaimImportJob(ctx context.Context, job importjob.Record, expectedUpdatedAt time.Time) (bool, error) {
	return r.delegate.ClaimImportJob(ctx, job, expectedUpdatedAt)
}

func (r *failingTerminalImportJobRepository) MarkImportJobHistoryRemoved(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, removedAt time.Time, expectedUpdatedAt time.Time) (bool, error) {
	return r.delegate.MarkImportJobHistoryRemoved(ctx, tenantID, inventoryID, jobID, removedAt, expectedUpdatedAt)
}

func (r *failingTerminalImportJobRepository) UpdateImportJob(ctx context.Context, job importjob.Record) error {
	return r.delegate.UpdateImportJob(ctx, job)
}

type fakeImportSourceVault struct {
	requests             map[importjob.ID]ports.ImportSourceRequest
	vacuumScopes         []ports.ImportJobSourceScope
	deleteReportsMissing bool
	vacuumErr            error
}

func (f *fakeImportSourceVault) StoreImportJobSource(_ context.Context, scope ports.ImportJobSourceScope, request ports.ImportSourceRequest, _ time.Time, _ time.Time) error {
	f.requests[scope.JobID] = request
	return nil
}

func (f *fakeImportSourceVault) ImportJobSourceRequest(_ context.Context, scope ports.ImportJobSourceScope) (ports.ImportSourceRequest, bool, error) {
	request, ok := f.requests[scope.JobID]
	return request, ok, nil
}

func (f *fakeImportSourceVault) DeleteImportJobSource(_ context.Context, scope ports.ImportJobSourceScope) (bool, error) {
	_, found := f.requests[scope.JobID]
	delete(f.requests, scope.JobID)
	if f.deleteReportsMissing {
		return false, nil
	}
	return found, nil
}

func (f *fakeImportSourceVault) VacuumImportJobSources(_ context.Context, _ time.Time) ([]ports.ImportJobSourceScope, error) {
	if f.vacuumErr != nil {
		return nil, f.vacuumErr
	}
	return append([]ports.ImportJobSourceScope{}, f.vacuumScopes...), nil
}

type fakeImportWorker struct {
	executed  []importjob.ID
	cancelled []importjob.ID
	modes     []importjob.CancellationMode
}

type failingAfterAuditRepository struct {
	fakeAuditRepository
	failAfter int
	err       error
}

func (r *failingAfterAuditRepository) SaveAuditRecord(ctx context.Context, record audit.Record) error {
	if len(r.items) >= r.failAfter {
		return r.err
	}
	return r.fakeAuditRepository.SaveAuditRecord(ctx, record)
}

type cancellingImportAssetUnitOfWork struct {
	delegate  ports.ImportAssetUnitOfWork
	jobs      ports.ImportJobRepository
	cancelled bool
}

func (u *cancellingImportAssetUnitOfWork) CreateImportedAsset(ctx context.Context, item asset.Asset, auditRecord audit.Record, undoableOperation *ports.UndoableOperation, promotedParent *asset.Asset, parentAuditRecord *audit.Record, link ports.ImportSourceLink, record ports.ImportJobResource) error {
	if err := u.delegate.CreateImportedAsset(ctx, item, auditRecord, undoableOperation, promotedParent, parentAuditRecord, link, record); err != nil {
		return err
	}
	if !u.cancelled {
		job, found, err := u.jobs.ImportJobByID(ctx, record.TenantID, record.InventoryID, record.JobID)
		if err != nil {
			return err
		}
		if found {
			job.Status = importjob.StatusCancelRequested
			job.CancellationMode = importjob.CancellationModeDiscardPartial
			if err := u.jobs.UpdateImportJob(ctx, job); err != nil {
				return err
			}
		}
		u.cancelled = true
	}
	return nil
}

type failingImportAttachmentUnitOfWork struct {
	err error
}

func (u failingImportAttachmentUnitOfWork) CreateImportedAttachment(context.Context, media.Attachment, audit.Record, ports.ImportSourceLink, ports.ImportJobResource) error {
	return u.err
}

type syncImportWorker struct {
	application App
}

type failingImportWorker struct {
	err error
}

func (w syncImportWorker) ExecuteImportJob(ctx context.Context, command ports.ImportJobCommand) (importjob.Record, error) {
	return w.application.ExecuteImportJob(ctx, command)
}

func (w syncImportWorker) CancelImportJob(context.Context, importjob.ID, importjob.CancellationMode) error {
	return nil
}

func (w failingImportWorker) ExecuteImportJob(context.Context, ports.ImportJobCommand) (importjob.Record, error) {
	return importjob.Record{}, w.err
}

func (w failingImportWorker) CancelImportJob(context.Context, importjob.ID, importjob.CancellationMode) error {
	return nil
}

func (f *fakeImportWorker) ExecuteImportJob(_ context.Context, command ports.ImportJobCommand) (importjob.Record, error) {
	f.executed = append(f.executed, command.JobID)
	return importjob.Record{ID: command.JobID}, nil
}

func (f *fakeImportWorker) CancelImportJob(_ context.Context, jobID importjob.ID, mode importjob.CancellationMode) error {
	f.cancelled = append(f.cancelled, jobID)
	f.modes = append(f.modes, mode)
	return nil
}
