package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

// PrintSettingsRepository atomically compares revision, fences the optional
// active destination snapshot, and persists settings together with audit.
type PrintSettingsRepository interface {
	GetPrintSettings(context.Context, printing.Scope) (printing.InventoryPrintSettings, error)
	SavePrintSettings(context.Context, printing.InventoryPrintSettings, uint64, *printing.SettingsDestination, audit.Record) (printing.InventoryPrintSettings, error)
}
type LabelSelectionValidator interface {
	ValidateDefaultLabelSelection(context.Context, printing.TemplateSelection, *printing.MediaSnapshot) error
}
