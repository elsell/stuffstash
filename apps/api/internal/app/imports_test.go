package app

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
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

func TestImportSourceRequestAllowsCSVExactlyAtDecodedLimit(t *testing.T) {
	content := strings.Repeat("a", MaxImportCSVBytes)
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	application := New(Dependencies{})

	request, err := application.importSourceRequest(ImportSourceInput{
		SourceType:    string(importplan.SourceLegacyHomeboxCSV),
		FileName:      "homebox.csv",
		ContentBase64: encoded,
	})
	if err != nil {
		t.Fatalf("expected CSV at decoded limit to be accepted, got %v", err)
	}
	if len(request.Content) != MaxImportCSVBytes {
		t.Fatalf("expected decoded CSV content at limit, got %d bytes", len(request.Content))
	}
}

func TestImportSourceRequestRejectsCSVOverDecodedLimit(t *testing.T) {
	content := strings.Repeat("a", MaxImportCSVBytes+1)
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	application := New(Dependencies{})

	_, err := application.importSourceRequest(ImportSourceInput{
		SourceType:    string(importplan.SourceLegacyHomeboxCSV),
		FileName:      "homebox.csv",
		ContentBase64: encoded,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected over-limit CSV to be rejected, got %v", err)
	}
	var sourceDetail ImportSourceInvalidInputError
	if !errors.As(err, &sourceDetail) {
		t.Fatalf("expected safe CSV size detail, got %v", err)
	}
	if sourceDetail.Detail != "CSV import file is too large. Choose a CSV up to 10 MB." {
		t.Fatalf("unexpected CSV size detail %q", sourceDetail.Detail)
	}
}

func TestImportSourceRequestRejectsInvalidCSVBase64WithSafeDetail(t *testing.T) {
	application := New(Dependencies{})

	_, err := application.importSourceRequest(ImportSourceInput{
		SourceType:    string(importplan.SourceLegacyHomeboxCSV),
		FileName:      "homebox.csv",
		ContentBase64: "not base64",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid CSV payload to be rejected, got %v", err)
	}
	var sourceDetail ImportSourceInvalidInputError
	if !errors.As(err, &sourceDetail) {
		t.Fatalf("expected safe CSV decode detail, got %v", err)
	}
	if sourceDetail.Detail != "CSV import file could not be decoded. Choose a valid exported CSV file and try again." {
		t.Fatalf("unexpected CSV decode detail %q", sourceDetail.Detail)
	}
}

func TestImportSourceRequestRejectsCSVContentForLiveSource(t *testing.T) {
	application := New(Dependencies{})

	_, err := application.importSourceRequest(ImportSourceInput{
		SourceType:    string(importplan.SourceLegacyHomebox),
		BaseURL:       "https://homebox.example.test",
		Username:      "owner@example.com",
		Password:      "secret",
		ContentBase64: base64.StdEncoding.EncodeToString([]byte("HB.location,HB.asset_id,HB.name\nGarage,HB-1,Drill\n")),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected mixed live source and CSV content to be rejected, got %v", err)
	}
	var sourceDetail ImportSourceInvalidInputError
	if !errors.As(err, &sourceDetail) {
		t.Fatalf("expected safe mixed-source detail, got %v", err)
	}
	if sourceDetail.Detail != "Uploaded CSV content is only valid for CSV imports. Choose CSV upload or remove the file content." {
		t.Fatalf("unexpected mixed-source detail %q", sourceDetail.Detail)
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
			SourceType:          string(importplan.SourceLegacyHomebox),
			BaseURL:             "https://homebox.example.test",
			Username:            "owner@example.com",
			Password:            "secret",
			IncludeImages:       true,
			AllowPrivateNetwork: true,
			AllowInsecureTLS:    true,
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
	if repository.jobs[job.ID].Source.Fingerprint != job.Source.Fingerprint {
		t.Fatalf("expected repository to persist source fingerprint")
	}
	if !job.Source.AllowPrivateNetwork || !job.Source.AllowInsecureTLS {
		t.Fatalf("expected preview to persist safe connection options, got %+v", job.Source)
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
	var sourceChangedErr ImportSourceChangedAfterPreviewError
	if !errors.As(err, &sourceChangedErr) {
		t.Fatalf("expected typed source changed error, got %T %[1]v", err)
	}
	if !strings.Contains(err.Error(), "Preview the source again") {
		t.Fatalf("expected actionable source changed error, got %q", err.Error())
	}
	if repository.jobs[job.ID].Status != importjob.StatusPreviewed {
		t.Fatalf("expected stale job to remain previewed, got %+v", repository.jobs[job.ID])
	}
}

func TestStartImportJobRequiresNewPreviewWhenSecurityOptionsChange(t *testing.T) {
	repository := newFakeImportJobRepository()
	source := &fakeImportSourceReader{plan: importPlanForDurableJob("Homebox", "source:drill")}
	vault := &fakeImportSourceVault{requests: map[importjob.ID]ports.ImportSourceRequest{}}
	worker := &fakeImportWorker{}
	observer := &fakeObserver{}
	application := newDurableImportTestApp(repository, source, &fakeIDGenerator{ids: []string{"job-one"}}, fakeClock{now: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)})
	application.importSourceVault = vault
	application.importWorker = worker
	application.observer = observer
	importSource := ImportSourceInput{
		SourceType:    string(importplan.SourceLegacyHomebox),
		BaseURL:       "https://homebox.example.test",
		Username:      "owner@example.com",
		Password:      "secret",
		IncludeImages: true,
	}
	job, err := application.CreateImportJobPreview(context.Background(), CreateImportJobPreviewInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Source:      importSource,
	})
	if err != nil {
		t.Fatalf("create import job preview: %v", err)
	}

	changed := importSource
	changed.AllowPrivateNetwork = true
	_, err = application.StartImportJob(context.Background(), StartImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Source:      changed,
	})
	var sourceChangedErr ImportSourceChangedAfterPreviewError
	if !errors.As(err, &sourceChangedErr) {
		t.Fatalf("expected source changed error for private-network option change, got %T %[1]v", err)
	}
	if repository.jobs[job.ID].Status != importjob.StatusPreviewed {
		t.Fatalf("expected security-option mismatch to leave job previewed, got %+v", repository.jobs[job.ID])
	}
	if len(vault.requests) != 0 {
		t.Fatalf("expected security-option mismatch not to store source credentials, got %+v", vault.requests)
	}
	if len(worker.executed) != 0 {
		t.Fatalf("expected security-option mismatch not to dispatch worker, got %+v", worker.executed)
	}
	if !observer.hasEvent(ports.EventImportJobSourceOptionsMismatch) {
		t.Fatalf("expected source-option mismatch observability event, got %+v", observer.events)
	}

	changed = importSource
	changed.AllowInsecureTLS = true
	_, err = application.StartImportJob(context.Background(), StartImportJobInput{
		Principal:   durableImportPrincipal(),
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       job.ID,
		Source:      changed,
	})
	if !errors.As(err, &sourceChangedErr) {
		t.Fatalf("expected source changed error for TLS option change, got %T %[1]v", err)
	}
	if repository.jobs[job.ID].Status != importjob.StatusPreviewed || len(vault.requests) != 0 || len(worker.executed) != 0 {
		t.Fatalf("expected TLS-option mismatch not to mutate job/store credentials/dispatch worker, job=%+v vault=%+v worker=%+v", repository.jobs[job.ID], vault.requests, worker.executed)
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
