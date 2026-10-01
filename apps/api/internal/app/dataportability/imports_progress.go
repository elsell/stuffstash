package dataportability

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a ImportService) UpdateImportProgress(ctx context.Context, command ports.ImportJobCommand, phase importjob.Phase, done int, total int, message string) error {
	for {
		job, err := a.importJob(ctx, command.TenantID, command.InventoryID, command.JobID)
		if err != nil {
			return err
		}
		updatedAt := a.deps.Clock.Now().UTC().Truncate(time.Microsecond)
		if !updatedAt.After(job.UpdatedAt) {
			updatedAt = job.UpdatedAt.UTC().Truncate(time.Microsecond).Add(time.Microsecond)
		}
		progress := importjob.Progress{Phase: phase, Done: done, Total: total, Message: message, UpdatedAt: updatedAt}
		if job.Status == importjob.StatusCancelRequested {
			mode := job.CancellationMode
			if mode == "" {
				mode = importjob.CancellationModeKeepPartial
			}
			if shouldPersistProgressAfterCancellation(job.Progress, progress) {
				updated, err := a.deps.ImportJobs.UpdateImportJobProgress(ctx, command.TenantID, command.InventoryID, command.JobID, progress, job.UpdatedAt)
				if err != nil {
					return err
				}
				if !updated {
					continue
				}
				a.recordImportProgressUpdated(ctx, job, progress)
			}
			return ImportCancelledError{Mode: mode}
		}
		if job.Status != importjob.StatusRunning {
			return apperrors.ErrPrecondition
		}
		updated, err := a.deps.ImportJobs.UpdateImportJobProgress(ctx, command.TenantID, command.InventoryID, command.JobID, progress, job.UpdatedAt)
		if err != nil {
			return err
		}
		if updated {
			a.recordImportProgressUpdated(ctx, job, progress)
			return nil
		}
	}
}

func shouldPersistProgressAfterCancellation(current importjob.Progress, next importjob.Progress) bool {
	if next.Done <= 0 {
		return false
	}
	if current.Phase != next.Phase {
		return true
	}
	return next.Done > current.Done
}
