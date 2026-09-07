package ports

import (
	"context"
	"time"
)

type ThumbnailQueueStatus struct {
	Pending, Leased, Failed, Completed int64
	OldestPendingAt                    time.Time
	BackfillComplete                   bool
}
type ThumbnailJobOperations interface {
	ThumbnailQueueStatus(ctx context.Context, now time.Time) (ThumbnailQueueStatus, error)
	RetryFailedThumbnailJobs(ctx context.Context, limit int, now time.Time) (int, error)
}
