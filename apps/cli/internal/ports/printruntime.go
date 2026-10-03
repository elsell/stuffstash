package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"time"
)

type PrintRegistry interface {
	Heartbeat(context.Context, string, *printing.ConnectorReport) error
	Printers(context.Context) ([]printing.RegisteredPrinter, error)
	Report(context.Context, string, printing.Readiness) error
}

// DevicePrinters opens the locally registered adapter and discovered stable
// physical identity. A successful connection already holds the device OS lock.
type DevicePrinters interface {
	Open(context.Context, printing.RegisteredPrinter) (PrinterConnection, error)
}

type PrintBackoff interface {
	Delay(failures int) time.Duration
}
