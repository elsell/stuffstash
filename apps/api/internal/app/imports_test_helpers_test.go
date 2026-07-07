package app

import (
	"context"
	"errors"
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
	if !ok || tenant.ID(job.TenantID.String()) != tenantID || inventory.InventoryID(job.InventoryID.String()) != inventoryID || !job.HistoryRemovedAt.IsZero() {
		return importjob.Record{}, false, nil
	}
	return job, true, nil
}

func (f *fakeImportJobRepository) ListImportJobs(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, _ ports.ImportJobPageRequest) ([]importjob.Record, error) {
	var jobs []importjob.Record
	for _, job := range f.jobs {
		if tenant.ID(job.TenantID.String()) == tenantID && inventory.InventoryID(job.InventoryID.String()) == inventoryID && job.HistoryRemovedAt.IsZero() {
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
	if !ok || tenant.ID(current.TenantID.String()) != tenantID || inventory.InventoryID(current.InventoryID.String()) != inventoryID || !current.UpdatedAt.Equal(expectedUpdatedAt) || !current.HistoryRemovedAt.IsZero() {
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
	if !ok || tenant.ID(current.TenantID.String()) != tenantID || inventory.InventoryID(current.InventoryID.String()) != inventoryID || !current.HistoryRemovedAt.IsZero() {
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
	current.ProgressHistory = importjob.AppendProgressHistory(current.ProgressHistory, progress)
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
		current, found, err := r.delegate.ImportJobByID(ctx, tenant.ID(job.TenantID.String()), inventory.InventoryID(job.InventoryID.String()), job.ID)
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
	requestErr           error
	vacuumErr            error
}

func (f *fakeImportSourceVault) StoreImportJobSource(_ context.Context, scope ports.ImportJobSourceScope, request ports.ImportSourceRequest, _ time.Time, _ time.Time) error {
	f.requests[scope.JobID] = request
	return nil
}

func (f *fakeImportSourceVault) ImportJobSourceRequest(_ context.Context, scope ports.ImportJobSourceScope) (ports.ImportSourceRequest, bool, error) {
	if f.requestErr != nil {
		return ports.ImportSourceRequest{}, false, f.requestErr
	}
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
