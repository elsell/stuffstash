package dataportability

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ImportCancelledError struct {
	Mode importjob.CancellationMode
}

func (e ImportCancelledError) Error() string {
	return "import cancelled"
}

func (a ImportService) ExecuteImportJob(ctx context.Context, command ports.ImportJobCommand) (importjob.Record, error) {
	job, err := a.importJob(ctx, command.TenantID, command.InventoryID, command.JobID)
	if err != nil {
		return importjob.Record{}, err
	}
	if job.Status != importjob.StatusRunning && job.Status != importjob.StatusCancelRequested && job.Status != importjob.StatusDiscardFailed {
		return importjob.Record{}, apperrors.ErrPrecondition
	}
	expectedTerminalStatus := job.Status
	result := ImportResult{Counts: job.Counts}
	var applyErr error
	if job.Status == importjob.StatusCancelRequested {
		mode := job.CancellationMode
		if mode == "" {
			mode = importjob.CancellationModeKeepPartial
		}
		applyErr = ImportCancelledError{Mode: mode}
	} else if job.Status == importjob.StatusDiscardFailed {
		applyErr = ImportCancelledError{Mode: importjob.CancellationModeDiscardPartial}
	} else {
		result, applyErr = a.applyImportPlan(ctx, command, &job)
	}
	if latest, found, latestErr := a.deps.ImportJobs.ImportJobByID(ctx, command.TenantID, command.InventoryID, command.JobID); latestErr != nil {
		return importjob.Record{}, latestErr
	} else if found {
		job.Progress = latest.Progress
		job.ProgressHistory = latest.ProgressHistory
		if latest.Status == importjob.StatusCancelRequested {
			job.CancellationMode = latest.CancellationMode
			job.CancellationRequestID = latest.CancellationRequestID
			expectedTerminalStatus = importjob.StatusCancelRequested
			if applyErr == nil {
				mode := latest.CancellationMode
				if mode == "" {
					mode = importjob.CancellationModeKeepPartial
				}
				applyErr = ImportCancelledError{Mode: mode}
			}
		} else if latest.Status == importjob.StatusDiscardFailed {
			expectedTerminalStatus = importjob.StatusDiscardFailed
		}
	}
	now := a.deps.Clock.Now().UTC().Truncate(time.Microsecond)
	if !job.Progress.UpdatedAt.IsZero() && !now.After(job.Progress.UpdatedAt) {
		now = job.Progress.UpdatedAt.UTC().Truncate(time.Microsecond).Add(time.Microsecond)
	}
	job.Counts.FieldsCreated = result.Counts.FieldsCreated
	job.Counts.FieldsExisting = result.Counts.FieldsExisting
	job.Counts.TagsCreated = result.Counts.TagsCreated
	job.Counts.TagsExisting = result.Counts.TagsExisting
	job.Counts.LocationsCreated = result.Counts.LocationsCreated
	job.Counts.AssetsCreated = result.Counts.AssetsCreated
	job.Counts.AssetsSkipped = result.Counts.AssetsSkipped
	job.Counts.AttachmentsCreated = result.Counts.AttachmentsCreated
	job.Counts.AttachmentsSkipped = result.Counts.AttachmentsSkipped
	var sourceChangedErr ImportSourceChangedAfterPreviewError
	sourceChangedAfterPreview := errors.As(applyErr, &sourceChangedErr)
	if !sourceChangedAfterPreview {
		job.Messages = append(job.Messages, importJobMessagesFromPlanMessages(result.Messages)...)
	}
	job.CompletedAt = now
	job.UpdatedAt = now
	terminalDone := job.Progress.Done
	terminalTotal := job.Progress.Total
	if applyErr == nil {
		terminalDone = terminalTotal
	}
	job.Progress = importjob.Progress{Phase: importjob.PhaseTerminal, Done: terminalDone, Total: terminalTotal, UpdatedAt: now}
	var cancelled ImportCancelledError
	if errors.As(applyErr, &cancelled) {
		if cancelled.Mode == importjob.CancellationModeDiscardPartial {
			discarded, links, discardErr := a.discardImportedJobResources(ctx, command, job.ID)
			job.Counts.RecordsDiscarded += discarded
			job.Counts.SourceLinksDiscarded += links
			if discardErr != nil {
				job.Status = importjob.StatusDiscardFailed
				job.Progress.Message = "Import cancellation cleanup failed"
				job.Messages = append(job.Messages, importJobMessageFromPlanMessage(importplan.Message{
					Code:     "import-discard-failed",
					Severity: importplan.SeverityError,
					Summary:  "Import cancellation cleanup failed",
					Detail:   safeImportError(discardErr),
				}))
			} else {
				job.Status = importjob.StatusCancelledDiscarded
				job.Progress.Message = "Import cancelled and partial progress discarded"
			}
		} else {
			job.Status = importjob.StatusCancelledKept
			job.Progress.Message = "Import cancelled and partial progress kept"
		}
	} else if sourceChangedAfterPreview {
		job.Status = importjob.StatusFailed
		job.Progress.Message = "Import source changed after preview"
		job.Messages = []importjob.Message{importJobMessageFromPlanMessage(importplan.Message{
			Code:     "import-source-changed",
			Severity: importplan.SeverityError,
			Summary:  "Import source changed after preview",
			Detail:   sourceChangedErr.Error(),
		})}
	} else if applyErr != nil {
		job.Status = importjob.StatusFailed
		job.Progress.Message = "Import failed"
		var attachmentSessionErr importAttachmentSessionStartError
		var attachmentStorageErr importAttachmentStorageError
		if !errors.As(applyErr, &attachmentSessionErr) && !errors.As(applyErr, &attachmentStorageErr) {
			job.Messages = append(job.Messages, importJobMessageFromPlanMessage(importplan.Message{
				Code:     "import-failed",
				Severity: importplan.SeverityError,
				Summary:  "Import failed",
				Detail:   safeImportError(applyErr),
			}))
		}
	} else {
		job.Status = importjob.StatusSucceeded
		job.Progress.Message = "Import completed"
	}
	job.ProgressHistory = importjob.AppendProgressHistory(job.ProgressHistory, job.Progress)
	updated, err := a.deps.ImportJobs.UpdateImportJobIfStatus(ctx, job, expectedTerminalStatus)
	if err != nil {
		return importjob.Record{}, err
	}
	if !updated {
		latest, found, latestErr := a.deps.ImportJobs.ImportJobByID(ctx, command.TenantID, command.InventoryID, command.JobID)
		if latestErr != nil {
			return importjob.Record{}, latestErr
		}
		if found && latest.Status == importjob.StatusCancelRequested {
			return a.ExecuteImportJob(ctx, command)
		}
		if found && isTerminalImportJobStatus(latest.Status) {
			return latest, nil
		}
		return importjob.Record{}, apperrors.ErrPrecondition
	}
	if job.Status == importjob.StatusCancelledDiscarded {
		a.recordImportDiscardCleanupEvent(ctx, job, ports.EventImportJobDiscardCleanupCompleted, job.Counts.RecordsDiscarded, job.Counts.SourceLinksDiscarded)
	} else if job.Status == importjob.StatusDiscardFailed {
		a.recordImportDiscardCleanupEvent(ctx, job, ports.EventImportJobDiscardCleanupFailed, job.Counts.RecordsDiscarded, job.Counts.SourceLinksDiscarded)
	}
	requestID := command.RequestID
	if job.CancellationRequestID != "" && (job.Status == importjob.StatusCancelledKept || job.Status == importjob.StatusCancelledDiscarded || job.Status == importjob.StatusDiscardFailed) {
		requestID = job.CancellationRequestID
	}
	if err := a.saveImportJobAuditRecord(ctx, command.Principal, requestID, job, importJobTerminalAuditAction(job), map[string]string{
		"records_discarded":      fmt.Sprintf("%d", job.Counts.RecordsDiscarded),
		"source_links_discarded": fmt.Sprintf("%d", job.Counts.SourceLinksDiscarded),
	}); err != nil {
		return importjob.Record{}, err
	}
	if a.deps.ImportSourceVault != nil {
		scope := a.importJobSourceScope(command.TenantID, command.InventoryID, command.JobID)
		deleted, err := a.deps.ImportSourceVault.DeleteImportJobSource(ctx, scope)
		if err != nil {
			job.Messages = append(job.Messages, importJobMessageFromPlanMessage(importplan.Message{
				Code:     "import-source-cleanup-failed",
				Severity: importplan.SeverityWarning,
				Summary:  "Temporary import credentials could not be cleaned up automatically",
				Detail:   "credential cleanup will be retried by the import credential vacuum",
			}))
			job.UpdatedAt = a.deps.Clock.Now().UTC()
			if updateErr := a.deps.ImportJobs.UpdateImportJob(ctx, job); updateErr != nil {
				return importjob.Record{}, updateErr
			}
		} else if deleted {
			if err := a.saveImportJobCredentialCleanedAuditRecord(ctx, scope); err != nil {
				return importjob.Record{}, err
			}
		}
	}
	if sourceChangedAfterPreview {
		return job, sourceChangedErr
	}
	return job, nil
}

func (a ImportService) applyImportPlan(ctx context.Context, command ports.ImportJobCommand, job *importjob.Record) (ImportResult, error) {
	if a.deps.ImportSources == nil || a.deps.ImportLinks == nil {
		return ImportResult{}, apperrors.ErrInvalidInput
	}
	sourceRequest, err := a.importJobSourceRequest(ctx, command.TenantID, command.InventoryID, command.JobID)
	if err != nil {
		return ImportResult{}, err
	}
	plan, err := a.deps.ImportSources.ReadImportPlan(ctx, sourceRequest)
	if err != nil {
		return ImportResult{}, ImportSourceInputError(err)
	}
	plan = cloneImportPlan(plan)
	result := ImportResult{}
	checkPlan, err := a.NormalizedImportPlanForJob(ctx, command.TenantID, command.InventoryID, plan)
	if err != nil {
		return ImportResult{}, err
	}
	result.Messages = append(result.Messages, checkPlan.Messages...)
	fingerprint, err := SourceFingerprint(checkPlan)
	if err != nil {
		return result, err
	}
	if fingerprint != job.Source.Fingerprint {
		a.deps.Observer.Record(ctx, ports.Event{
			Name:    ports.EventImportJobSourceFingerprintMismatch,
			Message: "Import source changed after preview.",
			Fields:  importJobEventFields(*job),
		})
		return result, ImportSourceChangedAfterPreviewError{}
	}
	if plan.Counts().Errors > 0 {
		return result, apperrors.ErrInvalidInput
	}
	sourceIdentity, err := importSourceIdentityForJob(job.Source)
	if err != nil {
		return result, err
	}
	if err := a.applyImportFields(ctx, command, plan, &result); err != nil {
		return result, err
	}
	tagIDsByKey, err := a.applyImportTags(ctx, command, plan, &result)
	if err != nil {
		return result, err
	}
	duplicates, err := a.existingHomeboxReferences(ctx, command.TenantID, command.InventoryID)
	if err != nil {
		return result, err
	}
	sourceToAssetID, err := a.applyImportAssets(ctx, command, job.ID, sourceIdentity, plan, duplicates, tagIDsByKey, &result)
	if err != nil {
		return result, err
	}
	if err := a.applyImportAttachments(ctx, command, job.ID, sourceIdentity, sourceRequest, plan, sourceToAssetID, &result); err != nil {
		return result, err
	}
	return result, nil
}
