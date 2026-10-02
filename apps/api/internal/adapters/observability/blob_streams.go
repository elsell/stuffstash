package observability

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"io"
	"sync"
)

type observedBlobStreams struct {
	delegate  ports.StreamingBlobStorage
	telemetry ports.Telemetry
}

func ObserveBlobStreams(delegate ports.StreamingBlobStorage, telemetry ports.Telemetry) ports.StreamingBlobStorage {
	if delegate == nil {
		return nil
	}
	if telemetry == nil {
		telemetry = ports.NoopTelemetry{}
	}
	return observedBlobStreams{delegate, telemetry}
}
func (b observedBlobStreams) PutBlobStream(ctx context.Context, input ports.BlobStreamWrite) (err error) {
	ctx, finish := b.telemetry.Start(ctx, ports.OperationBlobWrite)
	defer func() { finish(err) }()
	return b.delegate.PutBlobStream(ctx, input)
}
func (b observedBlobStreams) DeleteBlob(ctx context.Context, key media.StorageKey) (err error) {
	ctx, finish := b.telemetry.Start(ctx, ports.OperationBlobDelete)
	defer func() { finish(err) }()
	return b.delegate.DeleteBlob(ctx, key)
}
func (b observedBlobStreams) OpenBlobStream(ctx context.Context, key media.StorageKey) (ports.BlobReadStream, int64, error) {
	ctx, finish := b.telemetry.Start(ctx, ports.OperationBlobRead)
	stream, size, err := b.delegate.OpenBlobStream(ctx, key)
	if err != nil {
		finish(err)
		return nil, 0, err
	}
	return &observedBlobReadStream{BlobReadStream: stream, finish: finish}, size, nil
}

type observedBlobReadStream struct {
	ports.BlobReadStream
	finish            func(error)
	once              sync.Once
	mu                sync.Mutex
	readErr, closeErr error
}

func (s *observedBlobReadStream) Read(p []byte) (int, error) {
	n, err := s.BlobReadStream.Read(p)
	s.noteRead(err)
	return n, err
}
func (s *observedBlobReadStream) ReadAt(p []byte, offset int64) (int, error) {
	n, err := s.BlobReadStream.ReadAt(p, offset)
	s.noteRead(err)
	return n, err
}
func (s *observedBlobReadStream) noteRead(err error) {
	if err == nil || errors.Is(err, io.EOF) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.readErr == nil {
		s.readErr = err
	}
}
func (s *observedBlobReadStream) Close() error {
	s.once.Do(func() {
		s.closeErr = s.BlobReadStream.Close()
		s.mu.Lock()
		readErr := s.readErr
		s.mu.Unlock()
		s.finish(errors.Join(readErr, s.closeErr))
	})
	return s.closeErr
}
