package app

import (
	"context"
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
