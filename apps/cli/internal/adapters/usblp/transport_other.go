//go:build !linux

package usblp

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (Access) OpenTransport(context.Context, printing.Device) (ports.PrinterTransport, error) {
	return nil, ports.Failure("unsupported", "USB printing is supported on Linux only")
}
