package app

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const maxImportCSVBytes = 10 * 1024 * 1024
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

func (a App) CreateImportJobPreview(ctx context.Context, input CreateImportJobPreviewInput) (importjob.Record, error) {
	requestID, err := normalizedImportRequestID(input.RequestID)
	if err != nil {
		return importjob.Record{}, err
	}
	input.RequestID = requestID
	if err := a.ensureImportJobCreateAccess(ctx, input.Principal, input.TenantID, input.InventoryID); err != nil {
		return importjob.Record{}, err
	}
	if a.importJobs == nil {
		return importjob.Record{}, ErrInvalidInput
	}
	jobID := strings.TrimSpace(a.ids.NewID())
	if jobID == "" {
		return importjob.Record{}, ErrInvalidInput
	}
	sourceRequest, err := a.importSourceRequest(input.Source)
	if err != nil {
		return importjob.Record{}, err
	}
	plan, err := a.readImportSourceRequest(ctx, sourceRequest)
	if err != nil {
		return importjob.Record{}, importSourceInputError(err)
	}
	plan = a.normalizedImportPlanForJob(ctx, input.TenantID, input.InventoryID, plan)
	fingerprint, err := sourceFingerprint(plan)
	if err != nil {
		return importjob.Record{}, err
	}
	now := a.clock.Now().UTC()
	job := importjob.NewPreviewedRecord(importjob.ID(jobID), input.TenantID, input.InventoryID, input.Principal.ID, importjob.SourceRefFromPlan(plan, fingerprint), importjob.CountsFromPlan(plan), plan.Messages, now)
	job.Preview = importjob.PreviewSummaryFromPlan(plan, 12)
	if err := a.importJobs.SaveImportJob(ctx, job); err != nil {
		return importjob.Record{}, err
	}
	if err := a.saveImportJobAuditRecord(ctx, input.Principal, input.RequestID, job, audit.ActionImportJobPreviewed, nil); err != nil {
		return importjob.Record{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobPreviewed,
		Message: "Import job preview completed.",
		Fields:  importJobEventFields(job),
	})
	return job, nil
}

func (a App) ListImportJobs(ctx context.Context, input ListImportJobsInput) ([]importjob.Record, error) {
	if err := a.ensureImportJobViewAccess(ctx, input.Principal, input.TenantID, input.InventoryID); err != nil {
		return nil, err
	}
	if a.importJobs == nil {
		return nil, ErrInvalidInput
	}
	jobs, err := a.importJobs.ListImportJobs(ctx, input.TenantID, input.InventoryID, ports.ImportJobPageRequest{Limit: input.Limit})
	if err != nil {
		return nil, err
	}
	return a.withImportJobResources(ctx, input.TenantID, input.InventoryID, jobs)
}

func (a App) GetImportJob(ctx context.Context, input GetImportJobInput) (importjob.Record, error) {
	if err := a.ensureImportJobViewAccess(ctx, input.Principal, input.TenantID, input.InventoryID); err != nil {
		return importjob.Record{}, err
	}
	job, err := a.importJob(ctx, input.TenantID, input.InventoryID, input.JobID)
	if err != nil {
		return importjob.Record{}, err
	}
	return a.withImportJobResource(ctx, job)
}

func (a App) GetImportJobForWorker(ctx context.Context, command ports.ImportJobCommand) (importjob.Record, error) {
	job, err := a.importJob(ctx, command.TenantID, command.InventoryID, command.JobID)
	if err != nil {
		return importjob.Record{}, err
	}
	return a.withImportJobResource(ctx, job)
}

func (a App) StartImportJob(ctx context.Context, input StartImportJobInput) (importjob.Record, error) {
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
		return importjob.Record{}, ErrPrecondition
	}
	sourceRequest, err := a.importSourceRequest(input.Source)
	if err != nil {
		return importjob.Record{}, err
	}
	plan, err := a.readImportSourceRequest(ctx, sourceRequest)
	if err != nil {
		return importjob.Record{}, importSourceInputError(err)
	}
	plan = a.normalizedImportPlanForJob(ctx, input.TenantID, input.InventoryID, plan)
	fingerprint, err := sourceFingerprint(plan)
	if err != nil {
		return importjob.Record{}, err
	}
	if fingerprint != job.Source.Fingerprint {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventImportJobSourceFingerprintMismatch,
			Message: "Import source changed after preview.",
			Fields:  importJobEventFields(job),
		})
		return importjob.Record{}, ErrPrecondition
	}
	command, err := a.importJobCommand(input)
	if err != nil {
		return importjob.Record{}, err
	}
	if a.importWorker == nil {
		return importjob.Record{}, ErrInvalidInput
	}
	now := a.clock.Now().UTC()
	job.Status = importjob.StatusRunning
	job.StartedAt = now
	job.UpdatedAt = now
	job.Progress = importjob.Progress{Phase: importjob.PhaseReading, Message: "Reading source", UpdatedAt: now}
	job.ProgressHistory = importjob.AppendProgressHistory(job.ProgressHistory, job.Progress)
	updated, err := a.importJobs.UpdateImportJobIfStatus(ctx, job, importjob.StatusPreviewed)
	if err != nil {
		return importjob.Record{}, err
	}
	if !updated {
		return importjob.Record{}, ErrPrecondition
	}
	applySourceRequest := sourceRequest
	applySourceRequest.FetchAttachmentBytes = true
	if err := a.storeImportJobSource(ctx, job, applySourceRequest); err != nil {
		_ = a.failStartedImportJob(ctx, input.Principal, input.RequestID, job, "Import source credentials could not be stored", now, true, true)
		return importjob.Record{}, err
	}
	if err := a.saveImportJobAuditRecord(ctx, input.Principal, input.RequestID, job, audit.ActionImportJobStarted, nil); err != nil {
		_ = a.failStartedImportJob(ctx, input.Principal, input.RequestID, job, "Import could not be started", now, true, false)
		return importjob.Record{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobStarted,
		Message: "Import job started.",
		Fields:  importJobEventFields(job),
	})
	started, err := a.importWorker.ExecuteImportJob(ctx, command)
	if err != nil {
		_ = a.failStartedImportJob(ctx, input.Principal, input.RequestID, job, "Import could not be started", now, true, true)
		return importjob.Record{}, err
	}
	return started, nil
}

func (a App) failStartedImportJob(ctx context.Context, principal identity.Principal, requestID string, job importjob.Record, message string, now time.Time, cleanupSource bool, auditFailure bool) error {
	failed := job
	failed.Status = importjob.StatusFailed
	failed.CompletedAt = now
	failed.UpdatedAt = now
	failed.Progress = importjob.Progress{Phase: importjob.PhaseTerminal, Done: job.Progress.Done, Total: job.Progress.Total, Message: message, UpdatedAt: now}
	failed.ProgressHistory = importjob.AppendProgressHistory(failed.ProgressHistory, failed.Progress)
	sourceDeleted := false
	if cleanupSource && a.importSourceVault != nil {
		if deleted, err := a.importSourceVault.DeleteImportJobSource(ctx, a.importJobSourceScope(job.TenantID, job.InventoryID, job.ID)); err == nil {
			sourceDeleted = deleted
		}
	}
	if updated, err := a.importJobs.UpdateImportJobIfStatus(ctx, failed, importjob.StatusRunning); err != nil {
		return err
	} else if !updated {
		return ErrPrecondition
	}
	if sourceDeleted {
		_ = a.saveImportJobCredentialCleanedAuditRecord(ctx, a.importJobSourceScope(job.TenantID, job.InventoryID, job.ID))
	}
	if auditFailure {
		return a.saveImportJobAuditRecord(ctx, principal, requestID, failed, audit.ActionImportJobFailed, nil)
	}
	return nil
}

func (a App) CancelImportJob(ctx context.Context, input CancelImportJobInput) (importjob.Record, error) {
	requestID, err := normalizedImportRequestID(input.RequestID)
	if err != nil {
		return importjob.Record{}, err
	}
	input.RequestID = requestID
	if err := a.ensureImportJobCreateAccess(ctx, input.Principal, input.TenantID, input.InventoryID); err != nil {
		return importjob.Record{}, err
	}
	if input.Mode != importjob.CancellationModeKeepPartial && input.Mode != importjob.CancellationModeDiscardPartial {
		return importjob.Record{}, ErrInvalidInput
	}
	job, err := a.importJob(ctx, input.TenantID, input.InventoryID, input.JobID)
	if err != nil {
		return importjob.Record{}, err
	}
	if job.Status != importjob.StatusRunning && job.Status != importjob.StatusPreviewed {
		return importjob.Record{}, ErrPrecondition
	}
	now := a.clock.Now().UTC()
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
	updated, err := a.importJobs.UpdateImportJobIfStatus(ctx, job, expected)
	if err != nil {
		return importjob.Record{}, err
	}
	if !updated {
		return importjob.Record{}, ErrPrecondition
	}
	if err := a.saveImportJobAuditRecord(ctx, input.Principal, input.RequestID, job, audit.ActionImportJobCancellationRequested, nil); err != nil {
		return importjob.Record{}, err
	}
	if job.Status == importjob.StatusCancelledKept || job.Status == importjob.StatusCancelledDiscarded {
		if err := a.saveImportJobAuditRecord(ctx, input.Principal, input.RequestID, job, audit.ActionImportJobCancelled, nil); err != nil {
			return importjob.Record{}, err
		}
	}
	if a.importWorker != nil && job.Status == importjob.StatusCancelRequested {
		if err := a.importWorker.CancelImportJob(ctx, job.ID, input.Mode); err != nil {
			return importjob.Record{}, err
		}
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobCancellationRequested,
		Message: "Import job cancellation requested.",
		Fields:  importJobEventFields(job),
	})
	return job, nil
}

func (a App) RemoveImportJobFromHistory(ctx context.Context, input RemoveImportJobFromHistoryInput) error {
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
		return ErrPrecondition
	}
	now := a.clock.Now().UTC()
	expectedUpdatedAt := job.UpdatedAt
	job.HistoryRemovedAt = now
	job.UpdatedAt = now
	updated, err := a.importJobs.MarkImportJobHistoryRemoved(ctx, job.TenantID, job.InventoryID, job.ID, job.HistoryRemovedAt, expectedUpdatedAt)
	if err != nil {
		return err
	}
	if !updated {
		return ErrPrecondition
	}
	if err := a.saveImportJobAuditRecord(ctx, input.Principal, input.RequestID, job, audit.ActionImportJobHistoryRemoved, nil); err != nil {
		return err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobHistoryRemoved,
		Message: "Import job removed from history.",
		Fields:  importJobEventFields(job),
	})
	return nil
}

func normalizedImportRequestID(requestID string) (string, error) {
	requestID = strings.TrimSpace(requestID)
	if len(requestID) > maxImportRequestIDLength {
		return "", ErrInvalidInput
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

func (a App) ResumeRunningImportJobs(ctx context.Context, limit int) (int, error) {
	if a.importJobs == nil || a.importWorker == nil {
		return 0, ErrInvalidInput
	}
	resumed := 0
	for _, status := range []importjob.Status{importjob.StatusCancelRequested, importjob.StatusRunning, importjob.StatusDiscardFailed} {
		if limit > 0 && resumed >= limit {
			break
		}
		pageLimit := limit
		if pageLimit > 0 {
			pageLimit -= resumed
		}
		jobs, err := a.importJobs.ListImportJobsByStatus(ctx, ports.ImportJobStatusPageRequest{Status: status, Limit: pageLimit})
		if err != nil {
			return resumed, err
		}
		for _, job := range jobs {
			if a.importSourceVault != nil && job.Status == importjob.StatusRunning {
				if _, found, err := a.importSourceVault.ImportJobSourceRequest(ctx, a.importJobSourceScope(job.TenantID, job.InventoryID, job.ID)); err != nil || !found {
					if handled, handleErr := a.failRecoveringImportJobWithMissingSource(ctx, job); handleErr != nil {
						return resumed, handleErr
					} else if handled {
						resumed++
					}
					continue
				}
			}
			claimed, err := a.claimImportJobForRecovery(ctx, job)
			if err != nil {
				return resumed, err
			}
			if !claimed {
				continue
			}
			_, err = a.importWorker.ExecuteImportJob(ctx, ports.ImportJobCommand{
				Principal:   identity.Principal{ID: job.ActorID},
				TenantID:    job.TenantID,
				InventoryID: job.InventoryID,
				JobID:       job.ID,
			})
			if err != nil {
				return resumed, err
			}
			resumed++
		}
	}
	return resumed, nil
}

func (a App) failRecoveringImportJobWithMissingSource(ctx context.Context, job importjob.Record) (bool, error) {
	claimed, err := a.claimImportJobForRecovery(ctx, job)
	if err != nil {
		return false, err
	}
	if !claimed {
		return false, nil
	}
	current, err := a.importJob(ctx, job.TenantID, job.InventoryID, job.ID)
	if err != nil {
		return false, err
	}
	now := a.clock.Now().UTC()
	current.Status = importjob.StatusFailed
	current.CompletedAt = now
	current.UpdatedAt = now
	current.Progress = importjob.Progress{
		Phase:     importjob.PhaseTerminal,
		Done:      current.Progress.Done,
		Total:     current.Progress.Total,
		Message:   "Import source credentials were unavailable",
		UpdatedAt: now,
	}
	current.ProgressHistory = importjob.AppendProgressHistory(current.ProgressHistory, current.Progress)
	updated, err := a.importJobs.UpdateImportJobIfStatus(ctx, current, importjob.StatusRunning)
	if err != nil {
		return false, err
	}
	if !updated {
		return false, ErrPrecondition
	}
	if err := a.saveImportJobAuditRecord(ctx, identity.Principal{ID: current.ActorID}, "", current, audit.ActionImportJobFailed, nil); err != nil {
		return false, err
	}
	return true, nil
}

func (a App) claimImportJobForRecovery(ctx context.Context, job importjob.Record) (bool, error) {
	expectedUpdatedAt := job.UpdatedAt
	now := a.clock.Now().UTC()
	job.UpdatedAt = now
	job.Progress.UpdatedAt = now
	if strings.TrimSpace(job.Progress.Message) == "" {
		job.Progress.Message = "Resuming import"
	}
	claimed, err := a.importJobs.ClaimImportJob(ctx, job, expectedUpdatedAt)
	if err != nil || !claimed {
		return claimed, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobRecoveryClaimed,
		Message: "Import job claimed for recovery.",
		Fields:  importJobEventFields(job),
	})
	return true, nil
}

func (a App) VacuumImportJobCredentials(ctx context.Context) (int, error) {
	if a.importSourceVault == nil {
		return 0, ErrInvalidInput
	}
	scopes, err := a.importSourceVault.VacuumImportJobSources(ctx, a.clock.Now().UTC())
	if err != nil {
		a.observer.Record(ctx, ports.Event{
			Name:    ports.EventImportJobCredentialVacuumFailed,
			Message: "import job credential vacuum failed",
			Fields:  map[string]string{"error_class": "credential_vacuum_failed"},
		})
		return 0, err
	}
	for _, scope := range scopes {
		if err := a.saveImportJobCredentialCleanedAuditRecord(ctx, scope); err != nil {
			return 0, err
		}
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobCredentialVacuumed,
		Message: "import job credential vacuum completed",
		Fields:  map[string]string{"deleted": fmt.Sprintf("%d", len(scopes))},
	})
	return len(scopes), nil
}

func (a App) readImportSource(ctx context.Context, input ImportSourceInput) (importplan.Plan, error) {
	request, err := a.importSourceRequest(input)
	if err != nil {
		return importplan.Plan{}, err
	}
	return a.readImportSourceRequest(ctx, request)
}

func (a App) readImportSourceRequest(ctx context.Context, request ports.ImportSourceRequest) (importplan.Plan, error) {
	if a.importSources == nil {
		return importplan.Plan{}, ErrInvalidInput
	}
	return a.importSources.ReadImportPlan(ctx, request)
}

func (a App) importSourceRequest(input ImportSourceInput) (ports.ImportSourceRequest, error) {
	var content []byte
	if strings.TrimSpace(input.ContentBase64) != "" {
		if base64.StdEncoding.DecodedLen(len(strings.TrimSpace(input.ContentBase64))) > maxImportCSVBytes {
			return ports.ImportSourceRequest{}, ErrInvalidInput
		}
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(input.ContentBase64))
		if err != nil {
			return ports.ImportSourceRequest{}, ErrInvalidInput
		}
		if len(decoded) > maxImportCSVBytes {
			return ports.ImportSourceRequest{}, ErrInvalidInput
		}
		content = decoded
	}
	return ports.ImportSourceRequest{
		SourceType:          importplan.SourceType(input.SourceType),
		BaseURL:             input.BaseURL,
		Username:            input.Username,
		Password:            input.Password,
		IncludeImages:       input.IncludeImages,
		AllowInsecureTLS:    input.AllowInsecureTLS,
		AllowPrivateNetwork: input.AllowPrivateNetwork,
		MaxAttachmentBytes:  int64(a.maxAttachmentBytes),
		FileName:            input.FileName,
		Content:             content,
	}, nil
}

func (a App) importJobCommand(input StartImportJobInput) (ports.ImportJobCommand, error) {
	return ports.ImportJobCommand{
		Principal:   input.Principal,
		RequestID:   input.RequestID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		JobID:       input.JobID,
	}, nil
}

func (a App) importJobSourceScope(tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) ports.ImportJobSourceScope {
	return ports.ImportJobSourceScope{TenantID: tenantID, InventoryID: inventoryID, JobID: jobID}
}

func (a App) storeImportJobSource(ctx context.Context, job importjob.Record, request ports.ImportSourceRequest) error {
	if a.importSourceVault == nil {
		return ErrInvalidInput
	}
	now := a.clock.Now().UTC()
	return a.importSourceVault.StoreImportJobSource(ctx, a.importJobSourceScope(job.TenantID, job.InventoryID, job.ID), request, now.Add(a.importJobTimeout), now)
}

func (a App) importJobSourceRequest(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (ports.ImportSourceRequest, error) {
	if a.importSourceVault == nil {
		return ports.ImportSourceRequest{}, ErrInvalidInput
	}
	request, found, err := a.importSourceVault.ImportJobSourceRequest(ctx, a.importJobSourceScope(tenantID, inventoryID, jobID))
	if err != nil {
		return ports.ImportSourceRequest{}, err
	}
	if !found {
		return ports.ImportSourceRequest{}, ErrPrecondition
	}
	return request, nil
}

func importSourceInputError(err error) error {
	var userError ports.ImportSourceUserError
	if errors.As(err, &userError) {
		return NewImportSourceInvalidInputError(strings.TrimSpace(userError.Detail))
	}
	return ErrInvalidInput
}

func (a App) normalizedImportPlanForJob(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, plan importplan.Plan) importplan.Plan {
	plan = cloneImportPlan(plan)
	plan.Messages = append(plan.Messages, a.sourceLinkDuplicateWarnings(ctx, tenantID, inventoryID, plan)...)
	plan.Messages = append(plan.Messages, a.duplicateWarnings(ctx, tenantID, inventoryID, plan)...)
	plan.Messages = append(plan.Messages, archivedWarnings(plan)...)
	plan.Messages = safeImportMessages(plan.Messages)
	stripAttachmentContent(&plan)
	return plan
}

func cloneImportPlan(plan importplan.Plan) importplan.Plan {
	clone := plan
	clone.Fields = append([]importplan.FieldDefinition(nil), plan.Fields...)
	clone.Assets = make([]importplan.Asset, len(plan.Assets))
	for index, planned := range plan.Assets {
		clone.Assets[index] = planned
		if planned.CustomFields != nil {
			clone.Assets[index].CustomFields = make(map[string]any, len(planned.CustomFields))
			for key, value := range planned.CustomFields {
				clone.Assets[index].CustomFields[key] = value
			}
		}
	}
	clone.Attachments = make([]importplan.Attachment, len(plan.Attachments))
	for index, attachment := range plan.Attachments {
		clone.Attachments[index] = attachment
		clone.Attachments[index].Content = append([]byte(nil), attachment.Content...)
	}
	clone.Messages = append([]importplan.Message(nil), plan.Messages...)
	return clone
}

func safeImportMessages(messages []importplan.Message) []importplan.Message {
	safe := make([]importplan.Message, 0, len(messages))
	for _, message := range messages {
		message.Code = safeImportMessageText(message.Code)
		message.Summary = safeImportMessageText(message.Summary)
		message.Detail = safeImportMessageText(message.Detail)
		message.SourceID = safeImportMessageText(message.SourceID)
		message.SourceName = safeImportMessageText(message.SourceName)
		safe = append(safe, message)
	}
	return safe
}

func safeImportMessageText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	unsafeFragments := []string{
		"password",
		"passwd",
		"bearer ",
		"authorization:",
		"token=",
		"access_token",
		"refresh_token",
		"secret",
		"ciphertext",
		"nonce",
		"storage key",
		"s3://",
		"file://",
	}
	for _, fragment := range unsafeFragments {
		if strings.Contains(lower, fragment) {
			return ""
		}
	}
	const maxImportMessageTextLength = 240
	if len(value) > maxImportMessageTextLength {
		return value[:maxImportMessageTextLength]
	}
	return value
}

func (a App) ensureImportJobViewAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return a.ensureActiveInventoryAccess(ctx, principal, tenantID, inventoryID, ports.InventoryPermissionViewImportJob)
}

func (a App) ensureImportJobCreateAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return a.ensureActiveInventoryAccess(ctx, principal, tenantID, inventoryID, ports.InventoryPermissionCreateImportJob)
}

func (a App) importJob(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (importjob.Record, error) {
	if a.importJobs == nil || jobID.String() == "" {
		return importjob.Record{}, ErrInvalidInput
	}
	job, ok, err := a.importJobs.ImportJobByID(ctx, tenantID, inventoryID, jobID)
	if err != nil {
		return importjob.Record{}, err
	}
	if !ok {
		return importjob.Record{}, ErrNotFound
	}
	return job, nil
}

func (a App) withImportJobResources(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobs []importjob.Record) ([]importjob.Record, error) {
	if a.importLinks == nil {
		return jobs, nil
	}
	out := make([]importjob.Record, 0, len(jobs))
	for _, job := range jobs {
		enriched, err := a.withImportJobResource(ctx, job)
		if err != nil {
			return nil, err
		}
		if enriched.TenantID == tenantID && enriched.InventoryID == inventoryID {
			out = append(out, enriched)
		}
	}
	return out, nil
}

func (a App) withImportJobResource(ctx context.Context, job importjob.Record) (importjob.Record, error) {
	if a.importLinks == nil {
		return job, nil
	}
	if job.Status == importjob.StatusCancelledDiscarded {
		job.Resources = nil
		return job, nil
	}
	records, err := a.importLinks.ListImportJobResources(ctx, job.TenantID, job.InventoryID, job.ID, ports.ImportJobResourcePageRequest{Limit: maxImportJobResourceSummaries})
	if err != nil {
		return importjob.Record{}, err
	}
	job.Resources = make([]importjob.ResourceSummary, 0, len(records))
	for _, record := range records {
		job.Resources = append(job.Resources, importjob.ResourceSummary{
			ResourceType:     string(record.ResourceType),
			ResourceID:       strings.TrimSpace(record.ResourceID),
			ResourceOwnerID:  strings.TrimSpace(record.ResourceOwnerID),
			SourceEntityType: string(record.SourceEntityType),
			SourceEntityID:   strings.TrimSpace(record.SourceEntityID),
			CreatedAt:        record.CreatedAt.UTC(),
		})
	}
	return job, nil
}

func sourceFingerprint(plan importplan.Plan) (string, error) {
	safe := struct {
		Source      importplan.SourceSummary
		Fields      []importplan.FieldDefinition
		Assets      []importplan.Asset
		Attachments []sourceFingerprintAttachment
	}{
		Source: plan.Source,
		Fields: append([]importplan.FieldDefinition{}, plan.Fields...),
		Assets: append([]importplan.Asset{}, plan.Assets...),
	}
	safe.Attachments = make([]sourceFingerprintAttachment, 0, len(plan.Attachments))
	for _, attachment := range plan.Attachments {
		safe.Attachments = append(safe.Attachments, sourceFingerprintAttachment{
			SourceID:      attachment.SourceID,
			AssetSourceID: attachment.AssetSourceID,
			Primary:       attachment.Primary,
		})
	}
	sort.SliceStable(safe.Attachments, func(left, right int) bool {
		if safe.Attachments[left].AssetSourceID != safe.Attachments[right].AssetSourceID {
			return safe.Attachments[left].AssetSourceID < safe.Attachments[right].AssetSourceID
		}
		if safe.Attachments[left].SourceID != safe.Attachments[right].SourceID {
			return safe.Attachments[left].SourceID < safe.Attachments[right].SourceID
		}
		return !safe.Attachments[left].Primary && safe.Attachments[right].Primary
	})
	payload, err := json.Marshal(safe)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return fmt.Sprintf("sha256:%x", sum), nil
}

type sourceFingerprintAttachment struct {
	SourceID      string
	AssetSourceID string
	Primary       bool
}

func importJobEventFields(job importjob.Record) map[string]string {
	return map[string]string{
		"tenant_id":    job.TenantID.String(),
		"inventory_id": job.InventoryID.String(),
		"job_id":       job.ID.String(),
		"source_type":  string(job.Source.Type),
		"status":       string(job.Status),
	}
}

func (a App) recordImportProgressUpdated(ctx context.Context, job importjob.Record, progress importjob.Progress) {
	fields := importJobEventFields(job)
	fields["phase"] = string(progress.Phase)
	fields["done"] = fmt.Sprintf("%d", progress.Done)
	fields["total"] = fmt.Sprintf("%d", progress.Total)
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobProgressUpdated,
		Message: "Import job progress updated.",
		Fields:  fields,
	})
}

func (a App) recordImportSourceLinkDuplicateSkipped(ctx context.Context, command ports.ImportJobCommand, entityType ports.ImportSourceEntityType, jobID importjob.ID) {
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobSourceLinkDuplicateSkipped,
		Message: "Import source link duplicate skipped.",
		Fields: map[string]string{
			"tenant_id":          command.TenantID.String(),
			"inventory_id":       command.InventoryID.String(),
			"job_id":             jobID.String(),
			"source_entity_type": string(entityType),
		},
	})
}

func (a App) recordImportDiscardCleanupEvent(ctx context.Context, job importjob.Record, name ports.EventName, recordsDiscarded int, sourceLinksDiscarded int) {
	fields := importJobEventFields(job)
	fields["records_discarded"] = fmt.Sprintf("%d", recordsDiscarded)
	fields["source_links_discarded"] = fmt.Sprintf("%d", sourceLinksDiscarded)
	a.observer.Record(ctx, ports.Event{
		Name:    name,
		Message: "Import job discard cleanup updated.",
		Fields:  fields,
	})
}

func (a App) existingFieldKeys(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (map[string]struct{}, error) {
	keys := map[string]struct{}{}
	if a.customFields == nil {
		return keys, nil
	}
	fields, err := a.customFields.ListEffectiveCustomFieldDefinitions(ctx, tenantID, inventoryID)
	if err != nil {
		return nil, err
	}
	for _, field := range fields {
		keys[field.Key.String()] = struct{}{}
	}
	return keys, nil
}

func (a App) existingHomeboxReferences(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (map[string]struct{}, error) {
	ids := map[string]struct{}{}
	if a.assets == nil {
		return ids, nil
	}
	items, err := a.assets.ListAssetsByInventory(ctx, tenantID, inventoryID, ports.AssetListPageRequest{
		Limit:           10000,
		LifecycleFilter: ports.AssetLifecycleFilterAll,
		Sort:            ports.AssetListSortIDAsc,
	})
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		for _, key := range []string{"homebox-source-id", "homebox-asset-id"} {
			if value, ok := item.CustomFields.Values()[key].(string); ok && strings.TrimSpace(value) != "" {
				ids[key+"="+strings.TrimSpace(value)] = struct{}{}
			}
		}
	}
	return ids, nil
}

func (a App) duplicateWarnings(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, plan importplan.Plan) []importplan.Message {
	duplicates, err := a.existingHomeboxReferences(ctx, tenantID, inventoryID)
	if err != nil {
		return nil
	}
	var messages []importplan.Message
	for _, planned := range plan.Assets {
		for _, key := range []string{"homebox-source-id", "homebox-asset-id"} {
			value, ok := planned.CustomFields[key].(string)
			if !ok || strings.TrimSpace(value) == "" {
				continue
			}
			if _, duplicate := duplicates[key+"="+strings.TrimSpace(value)]; duplicate {
				messages = append(messages, importplan.Message{
					Code:       "duplicate-source-asset",
					Severity:   importplan.SeverityWarning,
					Summary:    "Asset appears to have already been imported",
					Detail:     key + "=" + strings.TrimSpace(value),
					SourceID:   planned.SourceID,
					SourceName: planned.Title,
				})
				break
			}
		}
	}
	return messages
}

func archivedWarnings(plan importplan.Plan) []importplan.Message {
	var messages []importplan.Message
	for _, planned := range plan.Assets {
		if !planned.Archived {
			continue
		}
		messages = append(messages, importplan.Message{
			Code:       "archived-source-asset-skipped",
			Severity:   importplan.SeverityWarning,
			Summary:    "Archived Homebox asset will be skipped",
			Detail:     "archived source assets are not imported in this version",
			SourceID:   planned.SourceID,
			SourceName: planned.Title,
		})
	}
	return messages
}

func sortedImportAssets(items []importplan.Asset, kind string) []importplan.Asset {
	return sortedImportAssetsByPredicate(items, func(item importplan.Asset) bool {
		return item.Kind == kind
	})
}

func sortedNonLocationImportAssets(items []importplan.Asset) []importplan.Asset {
	return sortedImportAssetsByPredicate(items, func(item importplan.Asset) bool {
		return item.Kind != "location"
	})
}

func sortedImportAssetsByPredicate(items []importplan.Asset, include func(importplan.Asset) bool) []importplan.Asset {
	bySourceID := map[string]importplan.Asset{}
	children := map[string][]importplan.Asset{}
	for _, item := range items {
		if include(item) {
			bySourceID[item.SourceID] = item
			children[item.ParentSourceID] = append(children[item.ParentSourceID], item)
		}
	}
	for parentID := range children {
		sort.SliceStable(children[parentID], func(left, right int) bool {
			return children[parentID][left].Title < children[parentID][right].Title
		})
	}
	var sorted []importplan.Asset
	visited := map[string]struct{}{}
	var visit func(importplan.Asset)
	visit = func(item importplan.Asset) {
		if _, ok := visited[item.SourceID]; ok {
			return
		}
		if parent, ok := bySourceID[item.ParentSourceID]; ok {
			visit(parent)
		}
		visited[item.SourceID] = struct{}{}
		sorted = append(sorted, item)
		for _, child := range children[item.SourceID] {
			visit(child)
		}
	}
	for _, root := range children[""] {
		visit(root)
	}
	var remaining []importplan.Asset
	for sourceID, item := range bySourceID {
		if _, ok := visited[sourceID]; !ok {
			remaining = append(remaining, item)
		}
	}
	sort.SliceStable(remaining, func(left, right int) bool {
		return remaining[left].Title < remaining[right].Title
	})
	for _, item := range remaining {
		visit(item)
	}
	return sorted
}

func stripAttachmentContent(plan *importplan.Plan) {
	for index := range plan.Attachments {
		plan.Attachments[index].Content = nil
	}
}

func safeImportError(err error) string {
	switch {
	case errors.Is(err, ErrAttachmentTooLarge):
		return "attachment is too large"
	case errors.Is(err, ErrAttachmentFileNameInvalid):
		return "attachment file name is invalid"
	case errors.Is(err, ErrAttachmentContentTypeUnsupported):
		return "attachment file type is unsupported"
	case errors.Is(err, ErrAttachmentContentMismatch):
		return "attachment content did not match its file type"
	case errors.Is(err, ErrAttachmentContentEmpty):
		return "attachment content was empty"
	default:
		return "import validation failed"
	}
}

func safeImportAttachmentUnavailableReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "attachment could not be downloaded"
	}
	return reason
}
