package presentation

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"sync"
)

// ReceiptDelivery has a bounded best-effort shutdown, separate from Record.
type ReceiptDelivery interface {
	ports.ProtocolReceipts
	Close(context.Context) bool
}

const receiptQueueCapacity = 64

type protocolReceipts struct {
	mu               sync.Mutex
	output           ports.Output
	queue            chan ports.ProtocolReceipt
	done             chan struct{}
	disabled, closed bool
}

func NewProtocolReceipts(output ports.Output) ReceiptDelivery {
	sink := &protocolReceipts{output: output, queue: make(chan ports.ProtocolReceipt, receiptQueueCapacity), done: make(chan struct{})}
	go sink.write()
	return sink
}
func (s *protocolReceipts) Record(_ context.Context, receipt ports.ProtocolReceipt) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.disabled || s.closed {
		return
	}
	select {
	case s.queue <- receipt:
	default:
		s.disabled = true
	}
}
func (s *protocolReceipts) Close(ctx context.Context) bool {
	s.mu.Lock()
	if !s.closed {
		s.closed = true
		close(s.queue)
	}
	s.mu.Unlock()
	select {
	case <-s.done:
		return true
	case <-ctx.Done():
		return false
	}
}
func (s *protocolReceipts) write() {
	defer close(s.done)
	if s.output == nil {
		return
	}
	for receipt := range s.queue {
		s.mu.Lock()
		disabled := s.disabled
		s.mu.Unlock()
		if disabled {
			s.warn()
			return
		}
		if err := s.output.Result(receipt); err != nil {
			s.mu.Lock()
			s.disabled = true
			s.mu.Unlock()
			s.warn()
			return
		}
	}
}
func (s *protocolReceipts) warn() {
	_ = s.output.Notice("Protocol receipt output failed or fell behind. Receipt output is disabled; do not repeat operations based on missing output.")
}
