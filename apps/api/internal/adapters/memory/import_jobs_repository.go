package memory

import (
	"context"
	"sort"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/importplan"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) SaveImportJob(_ context.Context, job importjob.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.importJobs[job.ID.String()] = job
	return nil
}

func (s *Store) ImportJobByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID) (importjob.Record, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.importJobs[jobID.String()]
	if !ok || job.TenantID != tenantID || job.InventoryID != inventoryID || !job.HistoryRemovedAt.IsZero() {
		return importjob.Record{}, false, nil
	}
	return cloneImportJob(job), true, nil
}

func (s *Store) ListImportJobs(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.ImportJobPageRequest) ([]importjob.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]importjob.Record, 0, len(s.importJobs))
	for _, job := range s.importJobs {
		if job.TenantID != tenantID || job.InventoryID != inventoryID || !job.HistoryRemovedAt.IsZero() {
			continue
		}
		jobs = append(jobs, cloneImportJob(job))
	}
	sort.SliceStable(jobs, func(left, right int) bool {
		return jobs[left].CreatedAt.After(jobs[right].CreatedAt)
	})
	if page.Limit > 0 && len(jobs) > page.Limit {
		jobs = jobs[:page.Limit]
	}
	return jobs, nil
}

func (s *Store) ListImportJobsByStatus(_ context.Context, page ports.ImportJobStatusPageRequest) ([]importjob.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]importjob.Record, 0, len(s.importJobs))
	for _, job := range s.importJobs {
		if job.Status != page.Status || !job.HistoryRemovedAt.IsZero() {
			continue
		}
		jobs = append(jobs, cloneImportJob(job))
	}
	sort.SliceStable(jobs, func(left, right int) bool {
		return jobs[left].UpdatedAt.Before(jobs[right].UpdatedAt)
	})
	if page.Limit > 0 && len(jobs) > page.Limit {
		jobs = jobs[:page.Limit]
	}
	return jobs, nil
}

func (s *Store) UpdateImportJob(_ context.Context, job importjob.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.importJobs[job.ID.String()]
	if !ok || current.TenantID != job.TenantID || current.InventoryID != job.InventoryID || !current.HistoryRemovedAt.IsZero() {
		return ports.ErrConflict
	}
	job.HistoryRemovedAt = current.HistoryRemovedAt
	s.importJobs[job.ID.String()] = job
	return nil
}

func (s *Store) MarkImportJobHistoryRemoved(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, removedAt time.Time, expectedUpdatedAt time.Time) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.importJobs[jobID.String()]
	if !ok || current.TenantID != tenantID || current.InventoryID != inventoryID || !current.UpdatedAt.Equal(expectedUpdatedAt) || !current.HistoryRemovedAt.IsZero() {
		return false, nil
	}
	current.HistoryRemovedAt = removedAt
	current.UpdatedAt = removedAt
	s.importJobs[jobID.String()] = current
	return true, nil
}

func (s *Store) UpdateImportJobIfStatus(_ context.Context, job importjob.Record, expected importjob.Status) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.importJobs[job.ID.String()]
	if !ok || current.TenantID != job.TenantID || current.InventoryID != job.InventoryID || current.Status != expected || !current.HistoryRemovedAt.IsZero() {
		return false, nil
	}
	s.importJobs[job.ID.String()] = job
	return true, nil
}

func (s *Store) UpdateImportJobProgress(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, jobID importjob.ID, progress importjob.Progress, expectedUpdatedAt time.Time) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.importJobs[jobID.String()]
	if !ok || current.TenantID != tenantID || current.InventoryID != inventoryID || !current.UpdatedAt.Equal(expectedUpdatedAt) || !current.HistoryRemovedAt.IsZero() {
		return false, nil
	}
	current.Progress = progress
	current.ProgressHistory = importjob.AppendProgressHistory(current.ProgressHistory, progress)
	current.UpdatedAt = progress.UpdatedAt
	s.importJobs[jobID.String()] = current
	return true, nil
}

func (s *Store) ClaimImportJob(_ context.Context, job importjob.Record, expectedUpdatedAt time.Time) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.importJobs[job.ID.String()]
	if !ok || current.TenantID != job.TenantID || current.InventoryID != job.InventoryID || !current.UpdatedAt.Equal(expectedUpdatedAt) || !current.HistoryRemovedAt.IsZero() {
		return false, nil
	}
	s.importJobs[job.ID.String()] = job
	return true, nil
}

func cloneImportJob(job importjob.Record) importjob.Record {
	job.Messages = append([]importplan.Message{}, job.Messages...)
	job.ProgressHistory = append([]importjob.Progress{}, job.ProgressHistory...)
	return job
}
