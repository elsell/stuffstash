package bootstrap

import (
	"context"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/brotherql"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/usblp"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"slices"
	"strings"
)

type PrinterRuntime struct {
	Printer   ports.Printer
	Discovery ports.PrinterDiscovery
}

// BuiltinPrinters is the one runtime registration source used for workers and
// offline public catalog export. Constructing adapters never probes hardware.
func BuiltinPrinters(getenv func(string) string) []PrinterRuntime {
	sysfs := getenv("STUFF_STASH_CLI_SYSFS_USB_ROOT")
	if sysfs == "" {
		sysfs = "/sys/bus/usb/devices"
	}
	devices := getenv("STUFF_STASH_CLI_USB_DEVICE_ROOT")
	if devices == "" {
		devices = "/dev/usb"
	}
	access := usblp.Access{Roots: usblp.Roots{SysfsUSB: sysfs, Devices: devices}, VendorID: brotherql.VendorID, ProductID: brotherql.ProductID, Model: brotherql.Model}
	return []PrinterRuntime{{Printer: brotherql.Adapter{Devices: access}, Discovery: access}}
}
func printerCommand(ctx context.Context, action string, getenv func(string) string, output ports.Output) error {
	runtimes := BuiltinPrinters(getenv)
	if action == "catalog" {
		descriptors := make([]printing.Descriptor, 0, len(runtimes))
		for _, runtime := range runtimes {
			descriptors = append(descriptors, runtime.Printer.Descriptor())
		}
		return output.Result(descriptors)
	}
	devices := make([]printing.Device, 0)
	for _, runtime := range runtimes {
		found, err := runtime.Discovery.Discover(ctx)
		if err != nil {
			return err
		}
		devices = append(devices, found...)
	}
	return output.Result(devices)
}

func supportsUSB(printers []PrinterRuntime, platform string) bool {
	for _, printer := range printers {
		descriptor := printer.Printer.Descriptor()
		if strings.HasPrefix(descriptor.Transport, "usb") && slices.Contains(descriptor.Platforms, platform) {
			return true
		}
	}
	return false
}
