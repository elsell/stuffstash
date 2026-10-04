package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
)

// PrinterDiscovery must not open a printer or send any bytes.
type PrinterDiscovery interface {
	Discover(context.Context) ([]printing.Device, error)
}

// PrinterConnection belongs to one worker holding a physical-device lock.
// Calls are serialized; Observe never writes. Each Submit represents one copy.
type PrinterConnection interface {
	Readiness(context.Context) (printing.Readiness, error)
	Submit(context.Context, printing.Label) (printing.Submission, error)
	Observe(context.Context, printing.Submission) (printing.Observation, error)
	Close() error
}
type Printer interface {
	Descriptor() printing.Descriptor
	Open(context.Context, printing.Device) (PrinterConnection, error)
}

// PrinterTransport is the bidirectional byte channel; it performs no rendering.
// It permits one concurrent reader and writer. Read and Write honor cancellation;
// the caller joins them before Close releases the underlying device.
type PrinterTransport interface {
	Read(context.Context, []byte) (int, error)
	Write(context.Context, []byte) (int, error)
	Close() error
}

// PrinterIdleConfirmation is optional recovery evidence. Implementations must
// verify current device state without submitting output; an uncertain or active
// local connection must fail. The caller holds its physical and journal locks.
type PrinterIdleConfirmation interface {
	ConfirmIdle(context.Context) error
}
