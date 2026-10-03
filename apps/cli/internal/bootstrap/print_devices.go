package bootstrap

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type registeredDevices struct{ runtimes []PrinterRuntime }

func (d registeredDevices) Open(ctx context.Context, registration printing.RegisteredPrinter) (ports.PrinterConnection, error) {
	for _, runtime := range d.runtimes {
		descriptor := runtime.Printer.Descriptor()
		if descriptor.ID != registration.AdapterID {
			continue
		}
		compatible := false
		for _, media := range descriptor.Media {
			if media == registration.Media {
				compatible = true
				break
			}
		}
		if !compatible {
			return nil, ports.Failure("configuration", "registered media is not supported by this printer adapter")
		}
		devices, err := runtime.Discovery.Discover(ctx)
		if err != nil {
			return nil, err
		}
		for _, device := range devices {
			if device.ID == registration.DeviceID {
				return runtime.Printer.Open(ctx, device)
			}
		}
		return nil, ports.Failure("unavailable", "registered printer is not connected")
	}
	return nil, ports.Failure("configuration", "registered printer adapter is not installed")
}
