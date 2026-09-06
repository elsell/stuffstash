package ports

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
)

type ThumbnailJobStatus string

const (
	ThumbnailJobPending   ThumbnailJobStatus = "pending"
	ThumbnailJobCompleted ThumbnailJobStatus = "completed"
	ThumbnailJobFailed    ThumbnailJobStatus = "failed"
)

type ThumbnailFailure string

const (
	ThumbnailFailureProcessing ThumbnailFailure = "processing"
	ThumbnailFailureStorage    ThumbnailFailure = "storage"
)

type ClaimedThumbnailJob struct {
	Job          media.ThumbnailJob
	ClaimID      string
	ClaimedUntil time.Time
	Attempts     int
}

type ThumbnailJobResolution struct {
	Status        ThumbnailJobStatus
	At            time.Time
	NextAttemptAt time.Time
	Failure       ThumbnailFailure
}

func (r ThumbnailJobResolution) Validate() error {
	if r.At.IsZero() {
		return errors.New("thumbnail job resolution time is required")
	}
	switch r.Status {
	case ThumbnailJobCompleted:
		if r.Failure == "" && r.NextAttemptAt.IsZero() {
			return nil
		}
	case ThumbnailJobPending, ThumbnailJobFailed:
		if r.Failure != ThumbnailFailureProcessing && r.Failure != ThumbnailFailureStorage {
			break
		}
		if r.Status == ThumbnailJobPending && r.NextAttemptAt.After(r.At) {
			return nil
		}
		if r.Status == ThumbnailJobFailed && r.NextAttemptAt.IsZero() {
			return nil
		}
	}
	return errors.New("invalid thumbnail job resolution")
}

// ThumbnailJobQueue is an operational work stream, not a user-facing discovery API.
// Claiming is atomic across processes; resolutions fence expired and replaced claims.
// Supply a fresh token per acquisition and retain the returned UTC-microsecond lease.
// Attempts count acquisitions, including claims lost to crashes.
type ThumbnailJobQueue interface {
	ClaimThumbnailJobs(ctx context.Context, claimID string, limit int, now, leaseUntil time.Time) ([]ClaimedThumbnailJob, error)
	ResolveThumbnailJob(ctx context.Context, claim ClaimedThumbnailJob, resolution ThumbnailJobResolution) error
}
