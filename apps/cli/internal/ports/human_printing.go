package ports

import (
	"context"
	"time"
)

type InventoryPrintDefaults struct {
	PrinterID, TemplateID string
	TemplateVersion       uint32
	ShowReference         bool
}
type LabelPrintSelection struct {
	// RequestBody retains explicit input without replacing its concurrency checks.
	RequestBody                                     []byte
	PreviewFingerprint                              string
	PrinterID, ExpectedMediaFingerprint, TemplateID string
	TemplateVersion                                 uint32
	ShowReference                                   bool
	Copies                                          int
}
type PrinterMediaPreset struct {
	ID      string
	Version uint32
}
type RegisteredPrinter struct {
	Media            PrinterMedia `json:"media"`
	ReadinessReason  *string      `json:"readinessReason,omitempty"`
	ReportedAt       *time.Time   `json:"reportedAt,omitempty"`
	AdapterID        string       `json:"adapterId"`
	Revision         uint64       `json:"revision"`
	MediaName        string       `json:"mediaName"`
	MediaPreset      string       `json:"mediaPreset"`
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	Readiness        string       `json:"readiness"`
	MediaFingerprint string       `json:"mediaFingerprint"`
	Retired          bool         `json:"retired"`
}
type PrintJobResolution struct {
	ReportedOutcome string    `json:"reportedOutcome"`
	ResolvedAt      time.Time `json:"resolvedAt"`
	ResolvedBy      string    `json:"resolvedBy"`
}
type PrintJobSummary struct {
	Kind             string                `json:"kind"`
	MediaFingerprint string                `json:"mediaFingerprint"`
	RequestedBy      string                `json:"requestedBy"`
	UpdatedAt        time.Time             `json:"updatedAt"`
	Resolution       *PrintJobResolution   `json:"resolution,omitempty"`
	ID               string                `json:"id"`
	PrinterID        string                `json:"printerId"`
	AssetID          string                `json:"assetId,omitempty"`
	Predecessor      string                `json:"predecessor,omitempty"`
	Status           string                `json:"status"`
	Revision         uint64                `json:"revision"`
	Copies           int                   `json:"copies"`
	CreatedAt        time.Time             `json:"createdAt"`
	Attempts         []PrintAttemptSummary `json:"attempts"`
}
type PrintAttemptSummary struct {
	ClaimedAt       time.Time  `json:"claimedAt"`
	LeaseExpiresAt  time.Time  `json:"leaseExpiresAt"`
	IdleConfirmedAt *time.Time `json:"idleConfirmedAt,omitempty"`
	SettledAt       *time.Time `json:"settledAt,omitempty"`
	StartedAt       *time.Time `json:"startedAt,omitempty"`
	ID              string     `json:"id"`
	ConnectorID     string     `json:"connectorId"`
	Outcome         string     `json:"outcome"`
	Reason          string     `json:"reason"`
	CompletedCopies int        `json:"completedCopies"`
}
type PrintSelectionSource interface {
	PrintDefaults(context.Context, Scope) (InventoryPrintDefaults, error)
	RegisteredPrinter(context.Context, Scope, string) (RegisteredPrinter, error)
}
type HumanPrintingAPI interface {
	ResolvePrint(context.Context, Scope, string, []byte) (Result[PrintJobSummary], error)
	PrinterProfiles(context.Context, Scope) (Result[[]PrinterProfile], error)
	Printer(context.Context, Scope, string) (Result[RegisteredPrinter], error)
	PrinterMediaPresets(context.Context, Scope, string) ([]PrinterMediaPreset, error)
	ConfigurePrinterMedia(context.Context, Scope, string, uint64, PrinterMediaPreset) (Result[RegisteredPrinter], error)
	PrintSelectionSource
	RegisteredPrinters(context.Context, Scope, Page) (Result[[]RegisteredPrinter], error)
	PrintJobs(context.Context, Scope, Page, string) (Result[[]PrintJobSummary], error)
	PrintJob(context.Context, Scope, string) (Result[PrintJobSummary], error)
	QueueLabel(context.Context, Scope, string, LabelPrintSelection, string) (Result[PrintJobSummary], error)
	TestPrinter(context.Context, Scope, LabelPrintSelection, string) (Result[PrintJobSummary], error)
	Reprint(context.Context, Scope, string, LabelPrintSelection, string) (Result[PrintJobSummary], error)
	CancelPrint(context.Context, Scope, string, uint64) (Result[PrintJobSummary], error)
}
