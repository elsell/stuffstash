package printing

import "time"

// InventoryPrintSettings controls new drafts and requests. It never changes a
// rendered job or encodes current hardware readiness.
type InventoryPrintSettings struct {
	Scope                Scope
	Revision             uint64
	DefaultPrinterID     PrinterID
	Template             TemplateSelection
	PrintOnCreateDefault bool
	UpdatedAt            time.Time
}

func DefaultInventoryPrintSettings(scope Scope) InventoryPrintSettings {
	return InventoryPrintSettings{Scope: scope, Template: TemplateSelection{ID: TemplateQRTitle, Version: 1, Options: TemplateOptions{ShowReference: true}}}
}

type SettingsDestination struct {
	ID               PrinterID
	Revision         uint64
	MediaFingerprint string
}

func (s InventoryPrintSettings) ValidReplacement(expected uint64) bool {
	return s.Scope.TenantID != "" && s.Scope.InventoryID != "" && s.Revision > 0 && s.Revision == expected+1 && s.Template.ID != "" && s.Template.Version > 0 && (!s.PrintOnCreateDefault || s.DefaultPrinterID != "")
}
