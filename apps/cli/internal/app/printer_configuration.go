package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type printerConfigurationAPI interface {
	RegisteredPrinter(context.Context, ports.Scope, string) (ports.RegisteredPrinter, error)
	PrinterMediaPresets(context.Context, ports.Scope, string) ([]ports.PrinterMediaPreset, error)
	ConfigurePrinterMedia(context.Context, ports.Scope, string, uint64, ports.PrinterMediaPreset) (ports.Result[ports.RegisteredPrinter], error)
}

func configurePrinter(ctx context.Context, api printerConfigurationAPI, o Options) (ports.Result[ports.RegisteredPrinter], error) {
	current, err := api.RegisteredPrinter(ctx, o.Scope, o.Command[2])
	if err != nil {
		return ports.Result[ports.RegisteredPrinter]{}, err
	}
	presets, err := api.PrinterMediaPresets(ctx, o.Scope, current.AdapterID)
	if err != nil {
		return ports.Result[ports.RegisteredPrinter]{}, err
	}
	var selected ports.PrinterMediaPreset
	for _, preset := range presets {
		if preset.ID != o.LabelSize {
			continue
		}
		if selected.Version != 0 && selected != preset {
			return ports.Result[ports.RegisteredPrinter]{}, ports.Failure("usage", "label size is ambiguous; refresh the supported printer catalog")
		}
		selected = preset
	}
	if selected.Version == 0 {
		return ports.Result[ports.RegisteredPrinter]{}, ports.Failure("usage", "label size is not supported by this printer's adapter")
	}
	// A conflict is returned to the caller; never overwrite a concurrent change.
	return api.ConfigurePrinterMedia(ctx, o.Scope, current.ID, current.Revision, selected)
}
