package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"time"
)

type ThumbnailBackfillProgress struct {
	Cursor            media.ID
	Scanned, Enqueued int
	Complete          bool
}

type ThumbnailBackfill interface {
	BackfillThumbnailJobs(ctx context.Context, limit int, now time.Time) (ThumbnailBackfillProgress, error)
}
