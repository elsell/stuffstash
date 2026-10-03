package dto

import "github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"

type PrinterScope struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-ID"`
	TenantID      string `path:"tenantId"`
	InventoryID   string `path:"inventoryId"`
}
type CreatePrinterInput struct {
	PrinterScope
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"200"`
	Body           RegisterPrinterBody
}
type RegisterPrinterBody struct {
	Name          string `json:"name" minLength:"1" maxLength:"100"`
	AdapterID     string `json:"adapterId"`
	PresetID      string `json:"presetId"`
	PresetVersion uint32 `json:"presetVersion" minimum:"1"`
}
type PrinterInput struct {
	PrinterScope
	PrinterID string `path:"printerId"`
}
type UpdatePrinterInput struct {
	PrinterInput
	Body UpdatePrinterBody
}
type UpdatePrinterBody struct {
	Revision      uint64  `json:"revision" minimum:"1"`
	Name          *string `json:"name,omitempty" minLength:"1" maxLength:"100"`
	PresetID      *string `json:"presetId,omitempty"`
	PresetVersion *uint32 `json:"presetVersion,omitempty"`
	Retired       *bool   `json:"retired,omitempty"`
}
type ListPrintersInput struct {
	PrinterScope
	Limit  int    `query:"limit" default:"50" minimum:"1" maximum:"100"`
	Cursor string `query:"cursor"`
}
type PrinterOutput struct {
	Status int
	Body   shared.SuccessEnvelope[Printer]
}
type PrintersOutput struct {
	Body shared.SuccessEnvelope[[]Printer]
}
type Printer struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	AdapterID        string       `json:"adapterId"`
	Revision         uint64       `json:"revision"`
	Retired          bool         `json:"retired"`
	Media            MediaProfile `json:"media"`
	MediaFingerprint string       `json:"mediaFingerprint"`
	Readiness        string       `json:"readiness"`
}
