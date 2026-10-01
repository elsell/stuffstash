package dataportability

import (
	"context"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const MaxImportCSVBytes = 10 * 1024 * 1024
const maxImportJobResourceSummaries = 50
const maxImportRequestIDLength = 128

type ImportSourceInput struct {
	SourceType          string
	BaseURL             string
	Username            string
	Password            string
	IncludeImages       bool
	AllowInsecureTLS    bool
	AllowPrivateNetwork bool
	FileName            string
	ContentBase64       string
}

type CreateImportJobPreviewInput struct {
	Principal   identity.Principal
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Source      ImportSourceInput
}

type StartImportJobInput struct {
	Principal   identity.Principal
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	JobID       importjob.ID
	Source      ImportSourceInput
}

type GetImportJobInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	JobID       importjob.ID
}

type ListImportJobsInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Limit       int
}

type CancelImportJobInput struct {
	Principal   identity.Principal
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	JobID       importjob.ID
	Mode        importjob.CancellationMode
}

type RemoveImportJobFromHistoryInput struct {
	Principal   identity.Principal
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	JobID       importjob.ID
}

type ImportResult struct {
	Counts   importjob.Counts
	Messages []importplan.Message
}

func (a ImportService) CreateImportJobPreview(ctx context.Context, input CreateImportJobPreviewInput) (importjob.Record, error) {
	requestID, err := normalizedImportRequestID(input.RequestID)
	if err != nil {
		return importjob.Record{}, err
	}
	input.RequestID = requestID
	if err := a.ensureImportJobCreateAccess(ctx, input.Principal, input.TenantID, input.InventoryID); err != nil {
		return importjob.Record{}, err
	}
	if a.deps.ImportJobs == nil {
		return importjob.Record{}, apperrors.ErrInvalidInput
	}
	jobID := strings.TrimSpace(a.deps.IDs.NewID())
	if jobID == "" {
		return importjob.Record{}, apperrors.ErrInvalidInput
	}
	sourceRequest, err := a.ImportSourceRequest(input.Source)
	if err != nil {
		return importjob.Record{}, err
	}
	plan, err := a.readImportSourceRequest(ctx, sourceRequest)
	if err != nil {
		return importjob.Record{}, ImportSourceInputError(err)
	}
	plan, err = a.NormalizedImportPlanForJob(ctx, input.TenantID, input.InventoryID, plan)
	if err != nil {
		return importjob.Record{}, err
	}
	fingerprint, err := SourceFingerprint(plan)
	if err != nil {
		return importjob.Record{}, err
	}
	now := a.deps.Clock.Now().UTC()
	job := importjob.NewPreviewedRecord(
		importjob.ID(jobID),
		importjob.TenantID(input.TenantID.String()),
		importjob.InventoryID(input.InventoryID.String()),
		importjob.PrincipalID(input.Principal.ID.String()),
		importJobSourceRefFromPlan(plan, fingerprint, sourceRequest),
		importJobCountsFromPlan(plan),
		importJobMessagesFromPlanMessages(plan.Messages),
		now,
	)
	job.Preview = importJobPreviewSummaryFromPlan(plan, 12)
	if err := a.deps.ImportJobs.SaveImportJob(ctx, job); err != nil {
		return importjob.Record{}, err
	}
	if err := a.saveImportJobAuditRecord(ctx, input.Principal, input.RequestID, job, audit.ActionImportJobPreviewed, nil); err != nil {
		return importjob.Record{}, err
	}
	a.deps.Observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobPreviewed,
		Message: "Import job preview completed.",
		Fields:  importJobEventFields(job),
	})
	return job, nil
}

func (a ImportService) ListImportJobs(ctx context.Context, input ListImportJobsInput) ([]importjob.Record, error) {
	if err := a.ensureImportJobViewAccess(ctx, input.Principal, input.TenantID, input.InventoryID); err != nil {
		return nil, err
	}
	if a.deps.ImportJobs == nil {
		return nil, apperrors.ErrInvalidInput
	}
	jobs, err := a.deps.ImportJobs.ListImportJobs(ctx, input.TenantID, input.InventoryID, ports.ImportJobPageRequest{Limit: input.Limit})
	if err != nil {
		return nil, err
	}
	return a.withImportJobResources(ctx, input.TenantID, input.InventoryID, jobs)
}

func (a ImportService) GetImportJob(ctx context.Context, input GetImportJobInput) (importjob.Record, error) {
	if err := a.ensureImportJobViewAccess(ctx, input.Principal, input.TenantID, input.InventoryID); err != nil {
		return importjob.Record{}, err
	}
	job, err := a.importJob(ctx, input.TenantID, input.InventoryID, input.JobID)
	if err != nil {
		return importjob.Record{}, err
	}
	return a.withImportJobResource(ctx, job)
}

func (a ImportService) GetImportJobForWorker(ctx context.Context, command ports.ImportJobCommand) (importjob.Record, error) {
	job, err := a.importJob(ctx, command.TenantID, command.InventoryID, command.JobID)
	if err != nil {
		return importjob.Record{}, err
	}
	return a.withImportJobResource(ctx, job)
}

func (a ImportService) StartImportJob(ctx context.Context, input StartImportJobInput) (importjob.Record, error) {
	requestID, err := normalizedImportRequestID(input.RequestID)
	if err != nil {
		return importjob.Record{}, err
	}
	input.RequestID = requestID
	if err := a.ensureImportJobCreateAccess(ctx, input.Principal, input.TenantID, input.InventoryID); err != nil {
		return importjob.Record{}, err
	}
	job, err := a.importJob(ctx, input.TenantID, input.InventoryID, input.JobID)
	if err != nil {
		return importjob.Record{}, err
	}
	if job.Status != importjob.StatusPreviewed {
		return importjob.Record{}, apperrors.ErrPrecondition
	}
	sourceRequest, err := a.ImportSourceRequest(input.Source)
	if err != nil {
		return importjob.Record{}, err
	}
	if !importSourceOptionsMatchPreview(job.Source, sourceRequest) {
		a.deps.Observer.Record(ctx, ports.Event{
			Name:    ports.EventImportJobSourceOptionsMismatch,
			Message: "Import source options changed after preview.",
			Fields:  importJobEventFields(job),
		})
		return importjob.Record{}, ImportSourceChangedAfterPreviewError{}
	}
	plan, err := a.readImportSourceRequest(ctx, sourceRequest)
	if err != nil {
		return importjob.Record{}, ImportSourceInputError(err)
	}
	plan, err = a.NormalizedImportPlanForJob(ctx, input.TenantID, input.InventoryID, plan)
	if err != nil {
		return importjob.Record{}, err
	}
	fingerprint, err := SourceFingerprint(plan)
	if err != nil {
		return importjob.Record{}, err
	}
	if fingerprint != job.Source.Fingerprint {
		a.deps.Observer.Record(ctx, ports.Event{
			Name:    ports.EventImportJobSourceFingerprintMismatch,
			Message: "Import source changed after preview.",
			Fields:  importJobEventFields(job),
		})
		return importjob.Record{}, ImportSourceChangedAfterPreviewError{}
	}
	command, err := a.importJobCommand(input)
	if err != nil {
		return importjob.Record{}, err
	}
	if a.deps.ImportWorker == nil {
		return importjob.Record{}, apperrors.ErrInvalidInput
	}
	now := a.deps.Clock.Now().UTC()
	job.Status = importjob.StatusRunning
	job.StartedAt = now
	job.UpdatedAt = now
	job.Progress = importjob.Progress{Phase: importjob.PhaseReading, Message: "Reading source", UpdatedAt: now}
	job.ProgressHistory = importjob.AppendProgressHistory(job.ProgressHistory, job.Progress)
	updated, err := a.deps.ImportJobs.UpdateImportJobIfStatus(ctx, job, importjob.StatusPreviewed)
	if err != nil {
		return importjob.Record{}, err
	}
	if !updated {
		return importjob.Record{}, apperrors.ErrPrecondition
	}
	if err := a.storeImportJobSource(ctx, job, sourceRequest); err != nil {
		_ = a.failStartedImportJob(ctx, input.Principal, input.RequestID, job, "Import source credentials could not be stored", now, true, true)
		return importjob.Record{}, err
	}
	if err := a.saveImportJobAuditRecord(ctx, input.Principal, input.RequestID, job, audit.ActionImportJobStarted, nil); err != nil {
		_ = a.failStartedImportJob(ctx, input.Principal, input.RequestID, job, "Import could not be started", now, true, false)
		return importjob.Record{}, err
	}
	a.deps.Observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobStarted,
		Message: "Import job started.",
		Fields:  importJobEventFields(job),
	})
	started, err := a.deps.ImportWorker.ExecuteImportJob(ctx, command)
	if err != nil {
		_ = a.failStartedImportJob(ctx, input.Principal, input.RequestID, job, "Import could not be started", now, true, true)
		return importjob.Record{}, err
	}
	return started, nil
}

func (a ImportService) failStartedImportJob(ctx context.Context, principal identity.Principal, requestID string, job importjob.Record, message string, now time.Time, cleanupSource bool, auditFailure bool) error {
	failed := job
	failed.Status = importjob.StatusFailed
	failed.CompletedAt = now
	failed.UpdatedAt = now
	failed.Progress = importjob.Progress{Phase: importjob.PhaseTerminal, Done: job.Progress.Done, Total: job.Progress.Total, Message: message, UpdatedAt: now}
	failed.ProgressHistory = importjob.AppendProgressHistory(failed.ProgressHistory, failed.Progress)
	sourceDeleted := false
	if cleanupSource && a.deps.ImportSourceVault != nil {
		if deleted, err := a.deps.ImportSourceVault.DeleteImportJobSource(ctx, a.importJobSourceScope(importJobTenantID(job.TenantID), importJobInventoryID(job.InventoryID), job.ID)); err == nil {
			sourceDeleted = deleted
		}
	}
	if updated, err := a.deps.ImportJobs.UpdateImportJobIfStatus(ctx, failed, importjob.StatusRunning); err != nil {
		return err
	} else if !updated {
		return apperrors.ErrPrecondition
	}
	if sourceDeleted {
		_ = a.saveImportJobCredentialCleanedAuditRecord(ctx, a.importJobSourceScope(importJobTenantID(job.TenantID), importJobInventoryID(job.InventoryID), job.ID))
	}
	if auditFailure {
		return a.saveImportJobAuditRecord(ctx, principal, requestID, failed, audit.ActionImportJobFailed, nil)
	}
	return nil
}

func (a ImportService) CancelImportJob(ctx context.Context, input CancelImportJobInput) (importjob.Record, error) {
	requestID, err := normalizedImportRequestID(input.RequestID)
	if err != nil {
		return importjob.Record{}, err
	}
	input.RequestID = requestID
	if err := a.ensureImportJobCreateAccess(ctx, input.Principal, input.TenantID, input.InventoryID); err != nil {
		return importjob.Record{}, err
	}
	if input.Mode != importjob.CancellationModeKeepPartial && input.Mode != importjob.CancellationModeDiscardPartial {
		return importjob.Record{}, apperrors.ErrInvalidInput
	}
	job, err := a.importJob(ctx, input.TenantID, input.InventoryID, input.JobID)
	if err != nil {
		return importjob.Record{}, err
	}
	if job.Status != importjob.StatusRunning && job.Status != importjob.StatusPreviewed {
		return importjob.Record{}, apperrors.ErrPrecondition
	}
	now := a.deps.Clock.Now().UTC()
	expected := job.Status
	if job.Status == importjob.StatusPreviewed {
		if input.Mode == importjob.CancellationModeDiscardPartial {
			job.Status = importjob.StatusCancelledDiscarded
			job.Progress = importjob.Progress{Phase: importjob.PhaseTerminal, Done: job.Progress.Total, Total: job.Progress.Total, Message: "Import cancelled before it started", UpdatedAt: now}
		} else {
			job.Status = importjob.StatusCancelledKept
			job.Progress = importjob.Progress{Phase: importjob.PhaseTerminal, Done: job.Progress.Total, Total: job.Progress.Total, Message: "Import cancelled before it started", UpdatedAt: now}
		}
		job.ProgressHistory = importjob.AppendProgressHistory(job.ProgressHistory, job.Progress)
		job.CompletedAt = now
	} else {
		job.Status = importjob.StatusCancelRequested
		job.Progress = importjob.Progress{Phase: job.Progress.Phase, Done: job.Progress.Done, Total: job.Progress.Total, Message: "Cancellation requested", UpdatedAt: now}
		job.ProgressHistory = importjob.AppendProgressHistory(job.ProgressHistory, job.Progress)
	}
	job.CancellationMode = input.Mode
	job.CancellationRequestID = input.RequestID
	job.UpdatedAt = now
	updated, err := a.deps.ImportJobs.UpdateImportJobIfStatus(ctx, job, expected)
	if err != nil {
		return importjob.Record{}, err
	}
	if !updated {
		return importjob.Record{}, apperrors.ErrPrecondition
	}
	if err := a.saveImportJobAuditRecord(ctx, input.Principal, input.RequestID, job, audit.ActionImportJobCancellationRequested, nil); err != nil {
		return importjob.Record{}, err
	}
	if job.Status == importjob.StatusCancelledKept || job.Status == importjob.StatusCancelledDiscarded {
		if err := a.saveImportJobAuditRecord(ctx, input.Principal, input.RequestID, job, audit.ActionImportJobCancelled, nil); err != nil {
			return importjob.Record{}, err
		}
	}
	if a.deps.ImportWorker != nil && job.Status == importjob.StatusCancelRequested {
		if err := a.deps.ImportWorker.CancelImportJob(ctx, job.ID, input.Mode); err != nil {
			return importjob.Record{}, err
		}
	}
	a.deps.Observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobCancellationRequested,
		Message: "Import job cancellation requested.",
		Fields:  importJobEventFields(job),
	})
	return job, nil
}

func (a ImportService) RemoveImportJobFromHistory(ctx context.Context, input RemoveImportJobFromHistoryInput) error {
	requestID, err := normalizedImportRequestID(input.RequestID)
	if err != nil {
		return err
	}
	input.RequestID = requestID
	if err := a.ensureImportJobCreateAccess(ctx, input.Principal, input.TenantID, input.InventoryID); err != nil {
		return err
	}
	job, err := a.importJob(ctx, input.TenantID, input.InventoryID, input.JobID)
	if err != nil {
		return err
	}
	if !isRemovableImportJobStatus(job.Status) {
		return apperrors.ErrPrecondition
	}
	now := a.deps.Clock.Now().UTC()
	expectedUpdatedAt := job.UpdatedAt
	job.HistoryRemovedAt = now
	job.UpdatedAt = now
	updated, err := a.deps.ImportJobs.MarkImportJobHistoryRemoved(ctx, importJobTenantID(job.TenantID), importJobInventoryID(job.InventoryID), job.ID, job.HistoryRemovedAt, expectedUpdatedAt)
	if err != nil {
		return err
	}
	if !updated {
		return apperrors.ErrPrecondition
	}
	if err := a.saveImportJobAuditRecord(ctx, input.Principal, input.RequestID, job, audit.ActionImportJobHistoryRemoved, nil); err != nil {
		return err
	}
	a.deps.Observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobHistoryRemoved,
		Message: "Import job removed from history.",
		Fields:  importJobEventFields(job),
	})
	return nil
}

func normalizedImportRequestID(requestID string) (string, error) {
	requestID = strings.TrimSpace(requestID)
	if len(requestID) > maxImportRequestIDLength {
		return "", apperrors.ErrInvalidInput
	}
	return requestID, nil
}

func isTerminalImportJobStatus(status importjob.Status) bool {
	switch status {
	case importjob.StatusSucceeded, importjob.StatusFailed, importjob.StatusCancelledKept, importjob.StatusCancelledDiscarded, importjob.StatusDiscardFailed:
		return true
	default:
		return false
	}
}

func isRemovableImportJobStatus(status importjob.Status) bool {
	switch status {
	case importjob.StatusSucceeded, importjob.StatusFailed, importjob.StatusCancelledKept, importjob.StatusCancelledDiscarded:
		return true
	default:
		return false
	}
}

func importSourceOptionsMatchPreview(source importjob.SourceRef, request ports.ImportSourceRequest) bool {
	return source.AllowPrivateNetwork == request.AllowPrivateNetwork &&
		source.AllowInsecureTLS == request.AllowInsecureTLS
}
