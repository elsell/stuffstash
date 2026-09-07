package worklimit

import (
	"sync"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type thumbnailReadiness struct {
	mu       sync.Mutex
	watchers map[media.StorageKey]map[chan struct{}]struct{}
}

func NewThumbnailReadiness() ports.ThumbnailReadiness {
	return &thumbnailReadiness{watchers: make(map[media.StorageKey]map[chan struct{}]struct{})}
}

func (r *thumbnailReadiness) Watch(key media.StorageKey) (<-chan struct{}, func()) {
	ready := make(chan struct{})
	r.mu.Lock()
	if r.watchers[key] == nil {
		r.watchers[key] = make(map[chan struct{}]struct{})
	}
	r.watchers[key][ready] = struct{}{}
	r.mu.Unlock()
	return ready, func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.watchers[key], ready)
		if len(r.watchers[key]) == 0 {
			delete(r.watchers, key)
		}
	}
}

func (r *thumbnailReadiness) Published(key media.StorageKey) {
	r.mu.Lock()
	defer r.mu.Unlock()
	watchers := r.watchers[key]
	delete(r.watchers, key)
	for ready := range watchers {
		close(ready)
	}
}
