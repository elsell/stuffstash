package dataportability

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// CleanupArchives fences expired jobs before handing unused blobs to the normal
// deletion outbox. Outbox rechecks also cover writes that finish after expiry.
func (s ArchiveService) CleanupArchives(ctx context.Context, limit int) error {
	now := s.deps.Clock.Now()
	jobs, err := s.deps.Artifacts.ListExpiredArchiveJobs(ctx, now, limit)
	if err != nil {
		return err
	}
	var failures []error
	for _, job := range jobs {
		if err = s.expireArchive(ctx, job); err != nil {
			failures = append(failures, err)
		}
	}
	artifacts, err := s.deps.Artifacts.ListDueArchiveArtifacts(ctx, now, limit)
	if err != nil {
		return errors.Join(append(failures, err)...)
	}
	for _, artifact := range artifacts {
		job, found, err := s.deps.Jobs.ArchiveJobByID(ctx, ports.ArchiveJobScope{TenantID: artifact.TenantID, SourceInventoryID: artifact.SourceInventoryID}, artifact.JobID)
		if err == nil && found && job.State != archivejob.Expired {
			err = s.expireArchive(ctx, job)
		}
		if err == nil {
			err = s.deps.Artifacts.RetireArchiveArtifact(ctx, artifact, now, s.deps.IDs.NewID())
		}
		if err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
func (s ArchiveService) expireArchive(ctx context.Context, job archivejob.Record) error {
	next, err := job.Expire(s.deps.Clock.Now())
	if err != nil {
		return err
	}
	record, err := s.jobAudit(next, audit.ActionArchiveJobUpdated)
	if err != nil {
		return err
	}
	record.Source = audit.SourceBackgroundJob
	changed, err := s.deps.Commands.UpdateArchiveJobAudited(ctx, next, job.Revision, record)
	if err != nil {
		return err
	}
	if !changed {
		return ports.ErrArchiveJobConflict
	}
	s.observe(ctx, ports.EventArchiveJobUpdated, next)
	return nil
}
