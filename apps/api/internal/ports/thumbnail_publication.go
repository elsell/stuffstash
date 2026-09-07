package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
)

// ThumbnailPublicationGuard serializes final derivative publication with deletion.
// A nil claim denotes already-authorized foreground work; background work must
// supply its current claim. The callback must publish blobs only, without calling
// repositories or reentering this guard, and honor its bounded context.
type ThumbnailPublicationGuard interface {
	Publish(ctx context.Context, attachment media.Attachment, claim *ClaimedThumbnailJob, publish func(context.Context) error) error
}
