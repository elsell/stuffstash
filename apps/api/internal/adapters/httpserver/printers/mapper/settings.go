package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

func PrintSettings(value printing.InventoryPrintSettings) dto.InventoryPrintSettings {
	var destination *string
	if value.DefaultPrinterID != "" {
		id := string(value.DefaultPrinterID)
		destination = &id
	}
	return dto.InventoryPrintSettings{Revision: value.Revision, DefaultPrinterID: destination, Template: dto.PrintSettingsTemplate{ID: string(value.Template.ID), Version: value.Template.Version, Options: dto.PrintSettingsOptions{ShowReference: value.Template.Options.ShowReference}}, PrintOnCreateDefault: value.PrintOnCreateDefault}
}
func PrintSettingsCommand(value dto.InventoryPrintSettings) printing.InventoryPrintSettings {
	result := printing.InventoryPrintSettings{Revision: value.Revision, Template: printing.TemplateSelection{ID: printing.TemplateID(value.Template.ID), Version: value.Template.Version, Options: printing.TemplateOptions{ShowReference: value.Template.Options.ShowReference}}, PrintOnCreateDefault: value.PrintOnCreateDefault}
	if value.DefaultPrinterID != nil {
		result.DefaultPrinterID = printing.PrinterID(*value.DefaultPrinterID)
	}
	return result
}
