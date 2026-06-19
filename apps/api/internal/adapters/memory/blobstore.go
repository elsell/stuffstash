package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) PutBlob(_ context.Context, key media.StorageKey, _ media.ContentType, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.blobs[key] = append([]byte(nil), data...)
	return nil
}

func (s *Store) GetBlob(_ context.Context, key media.StorageKey) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.blobs[key]
	if !ok {
		return nil, ports.ErrBlobNotFound
	}
	return append([]byte(nil), data...), nil
}

func (s *Store) DeleteBlob(_ context.Context, key media.StorageKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.blobs, key)
	return nil
}
