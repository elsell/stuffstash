package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type printerConfigurationStore struct {
	printer        ports.RegisteredPrinter
	presets        []ports.PrinterMediaPreset
	concurrentEdit bool
}

func (f *printerConfigurationStore) RegisteredPrinter(context.Context, ports.Scope, string) (ports.RegisteredPrinter, error) {
	return f.printer, nil
}
func (f *printerConfigurationStore) PrinterMediaPresets(_ context.Context, _ ports.Scope, adapter string) ([]ports.PrinterMediaPreset, error) {
	if adapter != f.printer.AdapterID {
		return nil, nil
	}
	if f.concurrentEdit {
		f.printer.Revision++
		f.printer.Name = "Changed elsewhere"
	}
	return f.presets, nil
}
func (f *printerConfigurationStore) ConfigurePrinterMedia(_ context.Context, _ ports.Scope, id string, revision uint64, preset ports.PrinterMediaPreset) (ports.Result[ports.RegisteredPrinter], error) {
	if id != f.printer.ID || revision != f.printer.Revision {
		return ports.Result[ports.RegisteredPrinter]{}, ports.Failure("conflict", "changed")
	}
	f.printer.MediaPreset = preset.ID
	f.printer.Revision++
	return ports.Result[ports.RegisteredPrinter]{Data: f.printer}, nil
}
func TestPrinterConfigurationRejectsUnsupportedMediaAndNeverOverwritesConcurrentEdits(t *testing.T) {
	options, err := Parse([]string{"printers", "configure", "brother", "--label-size", "brother-ql800-29x90", "--tenant", "home", "--inventory", "garage"}, func(string) string { return "" })
	if err != nil || validateCommand(options) != nil {
		t.Fatalf("command rejected: %+v %v", options, err)
	}
	fake := &printerConfigurationStore{printer: ports.RegisteredPrinter{ID: "brother", AdapterID: "brother-ql", Name: "Garage Brother", Revision: 2, Readiness: "unavailable"}, presets: []ports.PrinterMediaPreset{{ID: options.LabelSize, Version: 1}}}
	result, err := configurePrinter(context.Background(), fake, options)
	if err != nil || result.Data.Revision != 3 || result.Data.MediaPreset != options.LabelSize || result.Data.Name != "Garage Brother" {
		t.Fatalf("offline configuration: %+v %v", result, err)
	}
	options.LabelSize = "unsupported"
	if _, err = configurePrinter(context.Background(), fake, options); err == nil || fake.printer.Revision != 3 {
		t.Fatal("unsupported media changed printer")
	}
	options.LabelSize = fake.presets[0].ID
	fake.concurrentEdit = true
	if _, err = configurePrinter(context.Background(), fake, options); err == nil || fake.printer.Revision != 4 || fake.printer.Name != "Changed elsewhere" {
		t.Fatal("concurrent change overwritten or retried")
	}
	options.LabelSize = ""
	if err = validateCommand(options); err == nil {
		t.Fatal("missing explicit media selection accepted")
	}
}
