package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
)

// ThumbnailReader prepares content for an attachment already authorized by the
// calling application boundary. cached identifies an existing valid derivative.
type ThumbnailReader interface {
	ReadThumbnail(ctx context.Context, attachment media.Attachment, variant media.ThumbnailVariant) (derivative ImageDerivative, cached bool, err error)
}
