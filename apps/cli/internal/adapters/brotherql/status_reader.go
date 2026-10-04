package brotherql

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

const statusQueueCapacity = 32

// statusReader drains printer IN even while the foreground sends raster OUT.
// Its lifetime belongs to the connection, never an individual API request.
type statusReader struct {
	cancel   context.CancelFunc
	done     chan struct{}
	frames   chan status
	mu       sync.Mutex
	terminal error
}

func newStatusReader(transport ports.PrinterTransport) *statusReader {
	ctx, cancel := context.WithCancel(context.Background())
	r := &statusReader{cancel: cancel, done: make(chan struct{}), frames: make(chan status, statusQueueCapacity)}
	go r.run(ctx, transport)
	return r
}
func (r *statusReader) run(ctx context.Context, transport ports.PrinterTransport) {
	defer close(r.done)
	for {
		s, err := readDeviceStatus(ctx, transport)
		if err != nil {
			r.fail(err)
			return
		}
		select {
		case <-ctx.Done():
			r.fail(ctx.Err())
			return
		case r.frames <- s:
		default:
			r.fail(errors.New("printer status queue overflow"))
			return
		}
	}
}
func (r *statusReader) fail(err error) { r.mu.Lock(); defer r.mu.Unlock(); r.terminal = err }
func (r *statusReader) err() error     { r.mu.Lock(); defer r.mu.Unlock(); return r.terminal }
func (r *statusReader) next(ctx context.Context) (status, error) {
	if err := ctx.Err(); err != nil {
		return status{}, err
	}
	if err := r.err(); err != nil {
		return status{}, err
	}
	select {
	case <-ctx.Done():
		return status{}, ctx.Err()
	case <-r.done:
		return status{}, r.err()
	case s := <-r.frames:
		if err := ctx.Err(); err != nil {
			return status{}, err
		}
		if err := r.err(); err != nil {
			return status{}, err
		}
		return s, nil
	}
}
func (r *statusReader) close() { r.cancel(); <-r.done }
func readDeviceStatus(ctx context.Context, transport ports.PrinterTransport) (status, error) {
	var packet [32]byte
	read := 0
	for read < len(packet) {
		n, err := transport.Read(ctx, packet[read:])
		read += n
		if err != nil {
			return status{}, err
		}
		if n == 0 {
			return status{}, io.ErrNoProgress
		}
	}
	return parseStatus(packet[:])
}
