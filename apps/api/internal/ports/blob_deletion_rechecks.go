package ports

import (
	"context"
	"time"
)

type BlobDeletionRechecks interface {
	ClaimBlobDeletionRechecks(ctx context.Context, claimID string, limit int, now, leaseUntil time.Time, interval time.Duration) ([]BlobDeletionEvent, error)
	ResolveBlobDeletionRecheck(ctx context.Context, event BlobDeletionEvent, now time.Time, failed bool) error
}
