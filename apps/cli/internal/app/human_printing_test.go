package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestPrintSelectionRequiresExplicitIntentAndKeepsConfiguredOfflineDestination(t *testing.T) {
	fake := &printingSelectionStore{settings: ports.InventoryPrintDefaults{PrinterID: "garage", TemplateID: "qr-title", TemplateVersion: 1, ShowReference: true}, printer: ports.RegisteredPrinter{ID: "garage", MediaFingerprint: "roll-v1", Readiness: "unavailable"}}
	o, err := Parse([]string{"assets", "create", "--kind", "item", "--title", "Tools"}, func(string) string { return "" })
	if err != nil || o.PrintLabel {
		t.Fatal("inventory scripts opted in implicitly", err)
	}
	o.PrintLabel = true
	selection, err := selectPrinter(context.Background(), fake, o)
	if err != nil || selection.PrinterID != "garage" || selection.ExpectedMediaFingerprint != "roll-v1" || !selection.ShowReference {
		t.Fatalf("configured offline printer lost: %+v %v", selection, err)
	}
	fake.settings.PrinterID = ""
	if _, err = selectPrinter(context.Background(), fake, o); err == nil {
		t.Fatal("missing destination silently selected")
	}
	o.PrinterID = "garage"
	if _, err = selectPrinter(context.Background(), fake, o); err != nil {
		t.Fatal(err)
	}
}

type printingSelectionStore struct {
	settings ports.InventoryPrintDefaults
	printer  ports.RegisteredPrinter
}

func (f *printingSelectionStore) PrintDefaults(context.Context, ports.Scope) (ports.InventoryPrintDefaults, error) {
	return f.settings, nil
}
func (f *printingSelectionStore) RegisteredPrinter(_ context.Context, _ ports.Scope, id string) (ports.RegisteredPrinter, error) {
	if f.printer.ID != id {
		return ports.RegisteredPrinter{}, ports.Failure("not_found", "printer missing")
	}
	return f.printer, nil
}
