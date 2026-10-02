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
	Readers                          ports.ArchivePackageReader
	Plans                            ports.ArchivePlanCodec
	Fields                           ports.CustomFieldDefinitionRepository
	Types                            ports.CustomAssetTypeRepository
	Publisher                        ports.ArchiveRestoreUnitOfWork
	Limits                           ports.ArchivePackageLimits
	MaxRecords                       int
	LeaseDuration, HeartbeatInterval time.Duration
}
type ArchiveWorker struct{ deps ArchiveWorkerDependencies }

func NewArchiveWorker(d ArchiveWorkerDependencies) (ArchiveWorker, error) {
	if d.Service.deps.Jobs == nil || d.Snapshots == nil || d.Metadata == nil || d.Packages == nil || d.Readers == nil || d.Plans == nil || d.Fields == nil || d.Types == nil || d.Publisher == nil || d.MaxRecords <= 0 || d.Limits.CompressedBytes <= 0 || d.Limits.CompressedBytes > d.Service.deps.MaxArchiveBytes || d.Limits.ExpandedBytes <= 0 || d.Limits.MetadataBytes <= 0 || d.Limits.MetadataBytes > int64(int(^uint(0)>>1)) || d.Limits.EntryBytes <= 0 || d.Limits.EntryBytes > ports.MaxSinglePutBytes || d.Limits.MetadataBytes > d.Service.deps.MaxArchiveBytes || d.Limits.Entries < 2 || d.LeaseDuration <= 0 || d.HeartbeatInterval <= 0 || d.HeartbeatInterval >= d.LeaseDuration/2 {
		return ArchiveWorker{}, apperrors.ErrInvalidInput
	}
	return ArchiveWorker{deps: d}, nil
}

// RunJob claims one attempt and owns its lease through final publication.
func (w ArchiveWorker) RunJob(ctx context.Context, job archivejob.Record) error {
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
	done := make(chan archiveWorkResult, 1)
	go func() { done <- w.execute(work, claimed) }()
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
				return w.fail(ctx, job, archiveExecutionFailure(output.err), output.err)
			}
			if err = s.authorize(ctx, archiveJobAccess(job)); err != nil {
				return w.fail(ctx, job, archivejob.FailurePermission, err)
			}
			return w.publish(ctx, job, output)
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

func (w ArchiveWorker) publish(ctx context.Context, job archivejob.Record, output archiveWorkResult) error {
	s := w.deps.Service
	if job.Phase == archivejob.Finalization {
		var next archivejob.Record
		var err error
		if output.awaitingGrant {
			now := s.deps.Clock.Now()
			nextAttempt := now.Add(w.deps.HeartbeatInterval)
			if nextAttempt.After(job.ExpiresAt) {
				nextAttempt = job.ExpiresAt
			}
			next, err = job.DeferFinalization(job.LeaseToken, now, nextAttempt)
			if err != nil {
				return err
			}
			changed, err := s.deps.Jobs.UpdateArchiveJob(ctx, next, job.Revision)
			if err == nil && !changed {
				return ports.ErrArchiveJobConflict
			}
			return err
		} else {
			next, err = job.Complete(job.LeaseToken, s.deps.Clock.Now(), "")
		}
		_, err = s.transition(ctx, job, next, err)
		return err
	}
	if job.Kind == archivejob.Restore && job.Phase == archivejob.Execution {
		if output.plan == nil {
			return w.fail(ctx, job, archivejob.FailureInternal, ErrArchiveMetadata)
		}
		records, err := BuildArchiveRestoreAudits(*output.plan, job, s.deps.IDs, s.deps.Clock)
		if err != nil {
			return w.fail(ctx, job, archivejob.FailureInternal, err)
		}
		completed, err := w.deps.Publisher.PublishArchiveRestore(ctx, ports.ArchiveRestorePublication{Job: job, Plan: *output.plan, OwnerGrantEventID: s.deps.IDs.NewID(), AuditRecords: records})
		// A commit error may be ambiguous. CAS failure after committed publication
		// cannot turn the completed job into a failed one.
		if err != nil {
			return w.fail(ctx, job, archivejob.FailureConflict, err)
		}
		s.observe(ctx, ports.EventArchiveJobUpdated, completed)
		return nil
	}
	var next archivejob.Record
	var err error
	if job.Kind == archivejob.Restore {
		next, err = job.PreviewReady(job.LeaseToken, s.deps.Clock.Now(), output.artifact, output.planHash, output.destination)
	} else {
		next, err = job.Complete(job.LeaseToken, s.deps.Clock.Now(), output.artifact)
	}
	_, err = s.transition(ctx, job, next, err)
	return err
}

func archiveExecutionFailure(err error) archivejob.Failure {
	switch {
	case errors.Is(err, ports.ErrForbidden):
		return archivejob.FailurePermission
	case errors.Is(err, ErrArchiveMetadata), errors.Is(err, ports.ErrArchivePackageInvalid):
		return archivejob.FailureInvalidArchive
	case errors.Is(err, ports.ErrArchivePackageLimit), errors.Is(err, ports.ErrInventoryExportLimit), errors.Is(err, ports.ErrBlobStreamSize):
		return archivejob.FailureLimit
	default:
		return archivejob.FailureStorage
	}
}

// Drain recovers an expired claim or starts the next queued job. Competing
// workers use the same atomic claim operation; only one can execute each job.
func (w ArchiveWorker) Drain(ctx context.Context) (bool, error) {
	jobs, err := w.deps.Service.deps.Jobs.ListRunnableArchiveJobs(ctx, w.deps.Service.deps.Clock.Now(), 1)
	if err != nil {
		return false, err
	}
	if len(jobs) == 0 {
		return false, nil
	}
	err = w.RunJob(ctx, jobs[0])
	if errors.Is(err, ports.ErrArchiveJobConflict) {
		return false, nil
	}
	return true, err
}
