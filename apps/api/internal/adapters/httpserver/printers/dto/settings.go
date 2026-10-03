package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type PrintSettingsOptions struct {
	ShowReference bool `json:"showReference"`
}
type PrintSettingsTemplate struct {
	ID      string               `json:"id"`
	Version uint32               `json:"version" minimum:"1"`
	Options PrintSettingsOptions `json:"options"`
}
type InventoryPrintSettings struct {
	Revision             uint64                `json:"revision" minimum:"0"`
	DefaultPrinterID     *string               `json:"defaultPrinterId"`
	Template             PrintSettingsTemplate `json:"template"`
	PrintOnCreateDefault bool                  `json:"printOnCreateDefault"`
}
type PrintSettingsInput struct{ PrinterScope }
type ReplacePrintSettingsInput struct {
	PrinterScope
	Body InventoryPrintSettings
}
type PrintSettingsOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         shared.SuccessEnvelope[InventoryPrintSettings]
}
