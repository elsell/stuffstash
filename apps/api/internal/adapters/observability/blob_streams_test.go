package observability

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"io"
	"testing"
)

func TestStreamTelemetryIncludesLazyReadFailures(t *testing.T) {
	failure := errors.New("object changed during read")
	for _, at := range []bool{false, true} {
		for _, readErr := range []error{failure, io.EOF} {
			telemetry := &recordingOperations{}
			store := ObserveBlobStreams(failingStreamStorage{stream: failedReadStream{err: readErr}}, telemetry)
			stream, _, err := store.OpenBlobStream(context.Background(), "archive")
			if err != nil {
				t.Fatal(err)
			}
			if at {
				_, err = stream.ReadAt(make([]byte, 1), 0)
			} else {
				_, err = stream.Read(make([]byte, 1))
			}
			if !errors.Is(err, readErr) {
				t.Fatal("read error changed")
			}
			if len(telemetry.results) != 0 {
				t.Fatal("span ended before stream closed")
			}
			stream.Close()
			stream.Close()
			if len(telemetry.results) != 1 {
				t.Fatal("stream finished more than once")
			}
			if readErr == io.EOF {
				if telemetry.results[0] != nil {
					t.Fatal("EOF recorded as failure")
				}
			} else if !errors.Is(telemetry.results[0], failure) {
				t.Fatal("lazy read failure reported success")
			}
		}
	}
}

type failingStreamStorage struct {
	ports.StreamingBlobStorage
	stream ports.BlobReadStream
}

func (s failingStreamStorage) OpenBlobStream(context.Context, media.StorageKey) (ports.BlobReadStream, int64, error) {
	return s.stream, 1, nil
}

type failedReadStream struct{ err error }

func (s failedReadStream) Read([]byte) (int, error)          { return 0, s.err }
func (s failedReadStream) ReadAt([]byte, int64) (int, error) { return 0, s.err }
func (failedReadStream) Close() error                        { return nil }
