package brotherql

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type DeviceAccess interface {
	OpenTransport(context.Context, printing.Device) (ports.PrinterTransport, error)
}
type Adapter struct{ Devices DeviceAccess }

func (Adapter) Descriptor() printing.Descriptor { return Descriptor() }
func (a Adapter) Open(ctx context.Context, device printing.Device) (ports.PrinterConnection, error) {
	transport, err := a.Devices.OpenTransport(ctx, device)
	if err != nil {
		return nil, err
	}
	return NewConnection(transport), nil
}
