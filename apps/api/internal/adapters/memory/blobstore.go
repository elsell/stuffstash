package memory

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) PutBlob(_ context.Context, key media.StorageKey, _ media.ContentType, data []byte) error {
	s.blobMu.Lock()
	defer s.blobMu.Unlock()

	s.blobs[key] = append([]byte(nil), data...)
	return nil
}

func (s *Store) GetBlob(_ context.Context, key media.StorageKey) ([]byte, error) {
	s.blobMu.RLock()
	defer s.blobMu.RUnlock()

	data, ok := s.blobs[key]
	if !ok {
		return nil, ports.ErrBlobNotFound
	}
	return append([]byte(nil), data...), nil
}

func (s *Store) DeleteBlob(_ context.Context, key media.StorageKey) error {
	s.blobMu.Lock()
	defer s.blobMu.Unlock()

	delete(s.blobs, key)
	return nil
}
