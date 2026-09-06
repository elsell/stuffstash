package memory

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type thumbnailJobKey struct {
	AttachmentID media.ID
	Revision     media.ThumbnailRevision
}

type thumbnailJobRecord struct {
	claim         ports.ClaimedThumbnailJob
	status        ports.ThumbnailJobStatus
	nextAttemptAt time.Time
	failure       ports.ThumbnailFailure
}

func thumbnailKey(job media.ThumbnailJob) thumbnailJobKey {
	return thumbnailJobKey{AttachmentID: job.AttachmentID, Revision: job.Revision}
}

// Called under the attachment transaction's mutex after all validation succeeds.
func (s *Store) enqueueThumbnailJob(job *media.ThumbnailJob) {
	if job == nil {
		return
	}
	s.thumbnailJobs[thumbnailKey(*job)] = thumbnailJobRecord{
		claim: ports.ClaimedThumbnailJob{Job: *job}, status: ports.ThumbnailJobPending, nextAttemptAt: job.CreatedAt,
	}
}

func (s *Store) ClaimThumbnailJobs(ctx context.Context, claimID string, limit int, now, leaseUntil time.Time) ([]ports.ClaimedThumbnailJob, error) {
	now = now.UTC().Truncate(time.Microsecond)
	leaseUntil = leaseUntil.UTC().Truncate(time.Microsecond)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if claimID == "" || limit <= 0 || now.IsZero() || !leaseUntil.After(now) {
		return nil, errors.New("invalid thumbnail claim settings")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	candidates := []thumbnailJobRecord{}
	for key, record := range s.thumbnailJobs {
		attachment, exists := s.attachments[key.AttachmentID]
		if !exists || !record.claim.Job.Matches(attachment) {
			delete(s.thumbnailJobs, key)
			continue
		}
		if record.status != ports.ThumbnailJobPending || record.nextAttemptAt.After(now) || (record.claim.ClaimID != "" && record.claim.ClaimedUntil.After(now)) {
			continue
		}
		candidates = append(candidates, record)
	}
	sort.Slice(candidates, func(i, j int) bool {
		left, right := candidates[i].claim.Job, candidates[j].claim.Job
		if left.Priority != right.Priority {
			return left.Priority == media.ThumbnailJobNewImage
		}
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.Before(right.CreatedAt)
		}
		return left.AttachmentID.String() < right.AttachmentID.String()
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	result := make([]ports.ClaimedThumbnailJob, 0, len(candidates))
	for _, record := range candidates {
		record.claim.Attempts++
		record.claim.ClaimID = claimID
		record.claim.ClaimedUntil = leaseUntil
		s.thumbnailJobs[thumbnailKey(record.claim.Job)] = record
		result = append(result, record.claim)
	}
	return result, nil
}

func (s *Store) ResolveThumbnailJob(ctx context.Context, claim ports.ClaimedThumbnailJob, resolution ports.ThumbnailJobResolution) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := resolution.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := thumbnailKey(claim.Job)
	record, exists := s.thumbnailJobs[key]
	attachment, attachmentExists := s.attachments[claim.Job.AttachmentID]
	if !exists || !attachmentExists || claim.ClaimID == "" || record.claim.ClaimID != claim.ClaimID || !record.claim.ClaimedUntil.Equal(claim.ClaimedUntil) || !record.claim.ClaimedUntil.After(resolution.At) || !claim.Job.Matches(attachment) {
		return ports.ErrOutboxClaimLost
	}
	record.status = resolution.Status
	record.failure = resolution.Failure
	record.nextAttemptAt = resolution.NextAttemptAt
	record.claim.ClaimID = ""
	record.claim.ClaimedUntil = time.Time{}
	s.thumbnailJobs[key] = record
	return nil
}
