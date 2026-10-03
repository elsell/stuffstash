package gormstore

import (
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"time"
)

type printSettingsModel struct {
	TenantID             string    `gorm:"primaryKey"`
	InventoryID          string    `gorm:"primaryKey"`
	Revision             uint64    `gorm:"not null"`
	DefaultPrinterID     string    `gorm:"not null"`
	TemplateID           string    `gorm:"not null"`
	TemplateVersion      uint32    `gorm:"not null"`
	ShowReference        bool      `gorm:"not null"`
	PrintOnCreateDefault bool      `gorm:"not null"`
	UpdatedAt            time.Time `gorm:"not null"`
}

func (printSettingsModel) TableName() string { return "inventory_print_settings" }
func settingsModel(value printing.InventoryPrintSettings) printSettingsModel {
	return printSettingsModel{TenantID: value.Scope.TenantID, InventoryID: value.Scope.InventoryID, Revision: value.Revision, DefaultPrinterID: string(value.DefaultPrinterID), TemplateID: string(value.Template.ID), TemplateVersion: value.Template.Version, ShowReference: value.Template.Options.ShowReference, PrintOnCreateDefault: value.PrintOnCreateDefault, UpdatedAt: value.UpdatedAt}
}
func (value printSettingsModel) domain() printing.InventoryPrintSettings {
	return printing.InventoryPrintSettings{Scope: printing.Scope{TenantID: value.TenantID, InventoryID: value.InventoryID}, Revision: value.Revision, DefaultPrinterID: printing.PrinterID(value.DefaultPrinterID), Template: printing.TemplateSelection{ID: printing.TemplateID(value.TemplateID), Version: value.TemplateVersion, Options: printing.TemplateOptions{ShowReference: value.ShowReference}}, PrintOnCreateDefault: value.PrintOnCreateDefault, UpdatedAt: value.UpdatedAt}
}
