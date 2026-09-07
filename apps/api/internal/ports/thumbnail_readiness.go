package ports

import "github.com/stuffstash/stuff-stash/internal/domain/media"

// ThumbnailReadiness carries publication hints, never image data or authorization.
// Watch subscribes once; its returned function unsubscribes idempotently.
// Published wakes current subscribers only after durable publication succeeds.
type ThumbnailReadiness interface {
	Watch(media.StorageKey) (<-chan struct{}, func())
	Published(media.StorageKey)
}
