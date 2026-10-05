package ports

import "context"

type PrintSettings struct {
	Schema               *string               `json:"$schema,omitempty"`
	DefaultPrinterID     *string               `json:"defaultPrinterId"`
	PrintOnCreateDefault bool                  `json:"printOnCreateDefault"`
	Revision             uint64                `json:"revision"`
	Template             PrintSettingsTemplate `json:"template"`
}
type PrintSettingsTemplate struct {
	ID      string               `json:"id"`
	Version uint32               `json:"version"`
	Options PrintSettingsOptions `json:"options"`
}
type PrintSettingsOptions struct {
	ShowReference bool `json:"showReference"`
}
type PrintSettingsAPI interface {
	PrintSettings(context.Context, Scope) (Result[PrintSettings], error)
}
