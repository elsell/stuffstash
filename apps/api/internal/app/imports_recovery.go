package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

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
				if _, found, err := a.importSourceVault.ImportJobSourceRequest(ctx, a.importJobSourceScope(importJobTenantID(job.TenantID), importJobInventoryID(job.InventoryID), job.ID)); err != nil {
					if errors.Is(err, ports.ErrImportJobSourceUnreadable) {
						if handled, handleErr := a.failRecoveringImportJobWithUnavailableSource(ctx, job, "unreadable_import_source"); handleErr != nil {
							return resumed, handleErr
						} else if handled {
							resumed++
						}
						continue
					}
					return resumed, err
				} else if !found {
					if handled, handleErr := a.failRecoveringImportJobWithUnavailableSource(ctx, job, "missing_import_source"); handleErr != nil {
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
				Principal:   identity.Principal{ID: importJobPrincipalID(job.ActorID)},
				TenantID:    importJobTenantID(job.TenantID),
				InventoryID: importJobInventoryID(job.InventoryID),
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

func (a App) failRecoveringImportJobWithUnavailableSource(ctx context.Context, job importjob.Record, errorClass string) (bool, error) {
	claimed, err := a.claimImportJobForRecovery(ctx, job)
	if err != nil {
		return false, err
	}
	if !claimed {
		return false, nil
	}
	current, err := a.importJob(ctx, importJobTenantID(job.TenantID), importJobInventoryID(job.InventoryID), job.ID)
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
	if err := a.saveImportJobAuditRecord(ctx, identity.Principal{ID: importJobPrincipalID(current.ActorID)}, "", current, audit.ActionImportJobFailed, nil); err != nil {
		return false, err
	}
	fields := importJobEventFields(current)
	fields["error_class"] = errorClass
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventImportJobWorkerFailed,
		Message: "import job recovery failed because source material was unavailable",
		Fields:  fields,
	})
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
