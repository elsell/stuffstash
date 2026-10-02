package dataportability

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ArchiveWorkerDependencies struct {
	Service                          ArchiveService
	Snapshots                        ports.ArchiveSnapshotRepository
	Metadata                         ports.ArchiveMetadataCodec
	Packages                         ports.ArchivePackageWriter
	Limits                           ports.ArchivePackageLimits
	MaxRecords                       int
	LeaseDuration, HeartbeatInterval time.Duration
}
type ArchiveWorker struct{ deps ArchiveWorkerDependencies }

func NewArchiveWorker(d ArchiveWorkerDependencies) (ArchiveWorker, error) {
	if d.Service.deps.Jobs == nil || d.Snapshots == nil || d.Metadata == nil || d.Packages == nil || d.MaxRecords <= 0 || d.Limits.CompressedBytes <= 0 || d.Limits.CompressedBytes > d.Service.deps.MaxArchiveBytes || d.Limits.ExpandedBytes <= 0 || d.Limits.MetadataBytes <= 0 || d.Limits.MetadataBytes > int64(int(^uint(0)>>1)) || d.Limits.EntryBytes <= 0 || d.Limits.Entries < 2 || d.LeaseDuration <= 0 || d.HeartbeatInterval <= 0 || d.HeartbeatInterval >= d.LeaseDuration/2 {
		return ArchiveWorker{}, apperrors.ErrInvalidInput
	}
	return ArchiveWorker{deps: d}, nil
}

// RunJob claims one export attempt. Restore execution is added separately before
// runtime registration; unsupported jobs must remain queued, never be claimed.
func (w ArchiveWorker) RunJob(ctx context.Context, job archivejob.Record) error {
	if job.Kind != archivejob.Export {
		return archivejob.ErrTransition
	}
	s := w.deps.Service
	now := s.deps.Clock.Now()
	until := now.Add(w.deps.LeaseDuration)
	if until.After(job.ExpiresAt) {
		until = job.ExpiresAt
	}
	claimed, err := job.Claim(s.deps.IDs.NewID(), now, until)
	if err != nil {
		return err
	}
	changed, err := s.deps.Jobs.UpdateArchiveJob(ctx, claimed, job.Revision)
	if err != nil {
		return err
	}
	if !changed {
		return ports.ErrArchiveJobConflict
	}
	job = claimed
	if err = s.authorize(ctx, archiveJobAccess(job)); err != nil {
		return w.fail(ctx, job, archivejob.FailurePermission, err)
	}
	work, cancel := context.WithCancel(ctx)
	defer cancel()
	type result struct {
		artifact string
		err      error
	}
	done := make(chan result, 1)
	go func() { artifact, err := w.export(work, claimed); done <- result{artifact, err} }()
	ticker := time.NewTicker(w.deps.HeartbeatInterval)
	defer ticker.Stop()
	// This loop exclusively owns the changing job revision. Packaging only uses
	// the immutable claim, so heartbeat and final publication cannot race locally.
	for {
		select {
		case <-ctx.Done():
			cancel()
			<-done
			return ctx.Err()
		case <-ticker.C:
			now = s.deps.Clock.Now()
			until = now.Add(w.deps.LeaseDuration)
			if until.After(job.ExpiresAt) {
				until = job.ExpiresAt
			}
			if err = s.authorize(ctx, archiveJobAccess(job)); err != nil {
				cancel()
				<-done
				return w.fail(ctx, job, archivejob.FailurePermission, err)
			}
			// The final lease can already reach retention expiry. It cannot be
			// extended, but still check cancellation/takeover until it ends.
			if !until.After(job.LeaseUntil) && now.Before(job.LeaseUntil) {
				current, found, readErr := s.deps.Jobs.ArchiveJobByID(ctx, archiveJobAccess(job).scope(), job.ID)
				if readErr != nil || !found || current != job {
					cancel()
					<-done
					if readErr != nil {
						return readErr
					}
					return ports.ErrArchiveJobConflict
				}
				continue
			}
			next, renewErr := job.Heartbeat(job.LeaseToken, now, until)
			if renewErr == nil {
				changed, renewErr = s.deps.Jobs.UpdateArchiveJob(ctx, next, job.Revision)
				if renewErr == nil && !changed {
					renewErr = ports.ErrArchiveJobConflict
				}
			}
			if renewErr != nil {
				cancel()
				<-done
				return renewErr
			}
			job = next
		case output := <-done:
			if output.err != nil {
				return w.fail(ctx, job, archivejob.FailureStorage, output.err)
			}
			if err = s.authorize(ctx, archiveJobAccess(job)); err != nil {
				return w.fail(ctx, job, archivejob.FailurePermission, err)
			}
			next, completeErr := job.Complete(job.LeaseToken, s.deps.Clock.Now(), output.artifact)
			if completeErr != nil {
				return completeErr
			}
			_, completeErr = s.transition(ctx, job, next, nil)
			return completeErr
		}
	}
}
func archiveJobAccess(job archivejob.Record) ArchiveAccess {
	return ArchiveAccess{Principal: identity.Principal{ID: identity.PrincipalID(job.PrincipalID)}, TenantID: tenant.ID(job.TenantID), InventoryID: inventory.InventoryID(job.SourceInventoryID)}
}
func (w ArchiveWorker) fail(ctx context.Context, job archivejob.Record, failure archivejob.Failure, cause error) error {
	next, err := job.Fail(job.LeaseToken, w.deps.Service.deps.Clock.Now(), failure)
	if err != nil {
		return errors.Join(cause, err)
	}
	_, err = w.deps.Service.transition(ctx, job, next, nil)
	return errors.Join(cause, err)
}
