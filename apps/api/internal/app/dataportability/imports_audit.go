package dataportability

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a ImportService) saveImportJobAuditRecord(ctx context.Context, principal identity.Principal, requestID string, job importjob.Record, action audit.Action, metadata map[string]string) error {
	if a.deps.Audit == nil {
		return nil
	}
	if metadata == nil {
		metadata = map[string]string{}
	}
	metadata["import_job_status"] = string(job.Status)
	metadata["source_type"] = string(job.Source.Type)
	if job.CancellationMode != "" {
		metadata["cancellation_mode"] = string(job.CancellationMode)
	}
	record, err := appsupport.NewAuditRecord(a.deps.IDs, a.deps.Clock, appsupport.AuditRecordInput{
		Principal:   principal,
		TenantID:    importJobTenantID(job.TenantID),
		InventoryID: importJobInventoryID(job.InventoryID),
		RequestID:   requestID,
		Source:      audit.SourceImport,
		Action:      action,
		TargetType:  audit.TargetImportJob,
		TargetID:    job.ID.String(),
		Metadata:    metadata,
	})
	if err != nil {
		return err
	}
	return a.deps.Audit.SaveAuditRecord(ctx, record)
}

func (a ImportService) saveImportJobCredentialCleanedAuditRecord(ctx context.Context, scope ports.ImportJobSourceScope) error {
	if a.deps.Audit == nil {
		return nil
	}
	metadata := map[string]string{"import_job_status": "unknown"}
	if job, found, err := a.deps.ImportJobs.ImportJobByID(ctx, scope.TenantID, scope.InventoryID, scope.JobID); err == nil && found {
		metadata["import_job_status"] = string(job.Status)
		metadata["source_type"] = string(job.Source.Type)
	}
	record, err := appsupport.NewAuditRecord(a.deps.IDs, a.deps.Clock, appsupport.AuditRecordInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("system")},
		TenantID:    scope.TenantID,
		InventoryID: scope.InventoryID,
		Source:      audit.SourceSystem,
		Action:      audit.ActionImportJobCredentialCleaned,
		TargetType:  audit.TargetImportJob,
		TargetID:    scope.JobID.String(),
		Metadata:    metadata,
	})
	if err != nil {
		return err
	}
	return a.deps.Audit.SaveAuditRecord(ctx, record)
}

func importJobTerminalAuditAction(job importjob.Record) audit.Action {
	switch job.Status {
	case importjob.StatusSucceeded:
		return audit.ActionImportJobCompleted
	case importjob.StatusCancelledKept, importjob.StatusCancelledDiscarded:
		return audit.ActionImportJobCancelled
	case importjob.StatusFailed, importjob.StatusDiscardFailed:
		return audit.ActionImportJobFailed
	default:
		return audit.ActionImportJobCompleted
	}
}
