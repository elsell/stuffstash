package worklimit

import (
	"sync"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type thumbnailFlights struct {
	mu     sync.Mutex
	active map[media.StorageKey]chan struct{}
}

func NewThumbnailFlights() ports.ThumbnailFlights {
	return &thumbnailFlights{active: make(map[media.StorageKey]chan struct{})}
}

func (f *thumbnailFlights) TryStart(key media.StorageKey) (func(), <-chan struct{}, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if done, found := f.active[key]; found {
		return nil, done, false
	}
	done := make(chan struct{})
	f.active[key] = done
	var once sync.Once
	release := func() {
		once.Do(func() {
			f.mu.Lock()
			defer f.mu.Unlock()
			delete(f.active, key)
			close(done)
		})
	}
	return release, done, true
}
