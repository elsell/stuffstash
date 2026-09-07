package ports

import "github.com/stuffstash/stuff-stash/internal/domain/media"

// ThumbnailFlights coordinates generation, not authorization or cached data.
// TryStart never blocks. A nonowner receives the current owner's completion signal.
// Owners must release on every exit. Release is idempotent and forgets the key.
type ThumbnailFlights interface {
	TryStart(media.StorageKey) (release func(), finished <-chan struct{}, owned bool)
}
