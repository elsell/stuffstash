package app

import (
	"context"
	"errors"
	"reflect"
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
		SourceType: string(importjob.SourceTypeLegacyHomebox),
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
	unchanged, found, err := repository.ImportJobByID(context.Background(), tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID)
	if err != nil || !found {
		t.Fatalf("read import job after rejected cancel found=%t err=%v", found, err)
	}
	if unchanged.Status != importjob.StatusRunning || unchanged.CancellationRequestID != "" || unchanged.CancellationMode != "" {
		t.Fatalf("expected rejected cancel not to mutate running job, got %+v", unchanged)
	}
}

func TestImportJobMutationsTrimRequestIDsBeforeAudit(t *testing.T) {
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
			"job-for-start", "audit-preview-for-start", "audit-start",
			"job-one", "audit-preview", "audit-cancel-requested", "audit-cancelled", "audit-history-removed",
		}},
		Clock: fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
	})

	startJob, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "preview-for-start",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importjob.SourceTypeLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
			Username:   "owner@example.com",
			Password:   "secret",
		},
	})
	if err != nil {
		t.Fatalf("create start import job preview: %v", err)
	}
	if _, err := application.StartImportJob(ctx, StartImportJobInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "  start-request  ",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       startJob.ID,
		Source: ImportSourceInput{
			SourceType: string(importjob.SourceTypeLegacyHomebox),
			BaseURL:    "https://homebox.example.test",
			Username:   "owner@example.com",
			Password:   "secret",
		},
	}); err != nil {
		t.Fatalf("start import job: %v", err)
	}
	job, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "  preview-request  ",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source: ImportSourceInput{
			SourceType: string(importjob.SourceTypeLegacyHomebox),
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
		RequestID:   "  cancel-request  ",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Mode:        importjob.CancellationModeKeepPartial,
	}); err != nil {
		t.Fatalf("cancel import job: %v", err)
	}
	if err := application.RemoveImportJobFromHistory(ctx, RemoveImportJobFromHistoryInput{
		Principal:   durableImportPrincipal(),
		RequestID:   "  remove-request  ",
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
	requestIDs := map[string]string{}
	for _, record := range records {
		if record.TargetType == audit.TargetImportJob {
			requestIDs[record.TargetID+":"+string(record.Action)] = record.RequestID
		}
	}
	expected := map[string]string{
		"job-one:" + string(audit.ActionImportJobPreviewed):             "preview-request",
		"job-for-start:" + string(audit.ActionImportJobStarted):         "start-request",
		"job-one:" + string(audit.ActionImportJobCancellationRequested): "cancel-request",
		"job-one:" + string(audit.ActionImportJobCancelled):             "cancel-request",
		"job-one:" + string(audit.ActionImportJobHistoryRemoved):        "remove-request",
	}
	for key, requestID := range expected {
		if requestIDs[key] != requestID {
			t.Fatalf("expected trimmed request ID %q for %s, got %q in records %+v", requestID, key, requestIDs[key], records)
		}
	}
}

func TestPreviewStartAndRemoveImportJobRejectOverlongRequestIDBeforeMutation(t *testing.T) {
	ctx := context.Background()
	sourceInput := ImportSourceInput{
		SourceType: string(importjob.SourceTypeLegacyHomebox),
		BaseURL:    "https://homebox.example.test",
	}
	overlongRequestID := strings.Repeat("x", maxImportRequestIDLength+1)
	t.Run("preview", func(t *testing.T) {
		store := memory.NewStore()
		seedDurableImportMemoryInventory(t, ctx, store)
		application := New(Dependencies{
			Authorizer:    &fakeAuthorizer{},
			Tenants:       store,
			Inventories:   store,
			Audit:         store,
			ImportSources: &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")},
			ImportJobs:    store,
			IDs:           &fakeIDGenerator{ids: []string{"job-one"}},
			Clock:         fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
		})
		beforeAuditCount := importJobAuditRecordCount(t, ctx, store)

		_, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
			Principal:   durableImportPrincipal(),
			RequestID:   overlongRequestID,
			TenantID:    tenant.ID("tenant-one"),
			InventoryID: inventory.InventoryID("inventory-one"),
			Source:      sourceInput,
		})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected invalid input for overlong preview request ID, got %v", err)
		}
		jobs, err := store.ListImportJobs(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.ImportJobPageRequest{Limit: 10})
		if err != nil {
			t.Fatalf("list import jobs after rejected preview: %v", err)
		}
		if len(jobs) != 0 {
			t.Fatalf("expected rejected preview not to save a job, got %+v", jobs)
		}
		if got := importJobAuditRecordCount(t, ctx, store); got != beforeAuditCount {
			t.Fatalf("expected rejected preview not to write audit records, before=%d after=%d", beforeAuditCount, got)
		}
	})
	t.Run("start", func(t *testing.T) {
		store := memory.NewStore()
		seedDurableImportMemoryInventory(t, ctx, store)
		vault := &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}}
		worker := &fakeImportWorker{}
		application := New(Dependencies{
			Authorizer:        &fakeAuthorizer{},
			Tenants:           store,
			Inventories:       store,
			Audit:             store,
			ImportSources:     &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")},
			ImportJobs:        store,
			ImportSourceVault: vault,
			ImportWorker:      worker,
			IDs:               &fakeIDGenerator{ids: []string{"job-one", "audit-preview"}},
			Clock:             fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
		})
		job, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
			Principal:   durableImportPrincipal(),
			RequestID:   "preview-request",
			TenantID:    tenant.ID("tenant-one"),
			InventoryID: inventory.InventoryID("inventory-one"),
			Source:      sourceInput,
		})
		if err != nil {
			t.Fatalf("create import job preview: %v", err)
		}
		before, found, err := store.ImportJobByID(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), job.ID)
		if err != nil || !found {
			t.Fatalf("read import job before rejected start found=%t err=%v", found, err)
		}
		beforeAuditCount := importJobAuditRecordCount(t, ctx, store)

		_, err = application.StartImportJob(ctx, StartImportJobInput{
			Principal:   durableImportPrincipal(),
			RequestID:   overlongRequestID,
			TenantID:    tenant.ID("tenant-one"),
			InventoryID: inventory.InventoryID("inventory-one"),
			JobID:       job.ID,
			Source:      sourceInput,
		})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected invalid input for overlong start request ID, got %v", err)
		}
		unchanged, found, err := store.ImportJobByID(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), job.ID)
		if err != nil || !found {
			t.Fatalf("read import job after rejected start found=%t err=%v", found, err)
		}
		if !reflect.DeepEqual(unchanged, before) {
			t.Fatalf("expected rejected start not to mutate previewed job, before=%+v after=%+v", before, unchanged)
		}
		if len(vault.requests) != 0 {
			t.Fatalf("expected rejected start not to store source credentials, got %+v", vault.requests)
		}
		if len(worker.executed) != 0 {
			t.Fatalf("expected rejected start not to dispatch worker, got %+v", worker.executed)
		}
		if got := importJobAuditRecordCount(t, ctx, store); got != beforeAuditCount {
			t.Fatalf("expected rejected start not to write audit records, before=%d after=%d", beforeAuditCount, got)
		}
	})
	t.Run("remove", func(t *testing.T) {
		store := memory.NewStore()
		seedDurableImportMemoryInventory(t, ctx, store)
		application := New(Dependencies{
			Authorizer:    &fakeAuthorizer{},
			Tenants:       store,
			Inventories:   store,
			Audit:         store,
			ImportSources: &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")},
			ImportJobs:    store,
			IDs:           &fakeIDGenerator{ids: []string{"job-one", "audit-preview"}},
			Clock:         fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)},
		})
		job, err := application.CreateImportJobPreview(ctx, CreateImportJobPreviewInput{
			Principal:   durableImportPrincipal(),
			RequestID:   "preview-request",
			TenantID:    tenant.ID("tenant-one"),
			InventoryID: inventory.InventoryID("inventory-one"),
			Source:      sourceInput,
		})
		if err != nil {
			t.Fatalf("create import job preview: %v", err)
		}
		job.Status = importjob.StatusSucceeded
		if err := store.UpdateImportJob(ctx, job); err != nil {
			t.Fatalf("mark succeeded: %v", err)
		}
		before, found, err := store.ImportJobByID(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), job.ID)
		if err != nil || !found {
			t.Fatalf("read import job before rejected remove found=%t err=%v", found, err)
		}
		beforeAuditCount := importJobAuditRecordCount(t, ctx, store)

		err = application.RemoveImportJobFromHistory(ctx, RemoveImportJobFromHistoryInput{
			Principal:   durableImportPrincipal(),
			RequestID:   overlongRequestID,
			TenantID:    tenant.ID("tenant-one"),
			InventoryID: inventory.InventoryID("inventory-one"),
			JobID:       job.ID,
		})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected invalid input for overlong remove request ID, got %v", err)
		}
		unchanged, found, err := store.ImportJobByID(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), job.ID)
		if err != nil || !found {
			t.Fatalf("read import job after rejected remove found=%t err=%v", found, err)
		}
		if !reflect.DeepEqual(unchanged, before) {
			t.Fatalf("expected rejected remove not to mutate job, before=%+v after=%+v", before, unchanged)
		}
		if got := importJobAuditRecordCount(t, ctx, store); got != beforeAuditCount {
			t.Fatalf("expected rejected remove not to write audit records, before=%d after=%d", beforeAuditCount, got)
		}
	})
}

func importJobAuditRecordCount(t *testing.T, ctx context.Context, store *memory.Store) int {
	t.Helper()
	records, err := store.ListInventoryAuditRecords(ctx, tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"), ports.AuditRecordPageRequest{Limit: 100})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	count := 0
	for _, record := range records {
		if record.TargetType == audit.TargetImportJob {
			count++
		}
	}
	return count
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
	job := importjob.NewPreviewedRecord(importjob.ID("job-one"), importjob.TenantID("tenant-one"), importjob.InventoryID("inventory-one"), importjob.PrincipalID("owner"), importjob.SourceRef{
		Type:        importjob.SourceTypeLegacyHomebox,
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
