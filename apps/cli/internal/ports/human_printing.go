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
	PrinterID, ExpectedMediaFingerprint, TemplateID string
	TemplateVersion                                 uint32
	ShowReference                                   bool
	Copies                                          int
}
type RegisteredPrinter struct {
	MediaName        string `json:"mediaName"`
	MediaPreset      string `json:"mediaPreset"`
	ID               string `json:"id"`
	Name             string `json:"name"`
	Readiness        string `json:"readiness"`
	MediaFingerprint string `json:"mediaFingerprint"`
	Retired          bool   `json:"retired"`
}
type PrintJobSummary struct {
	ID          string                `json:"id"`
	PrinterID   string                `json:"printerId"`
	AssetID     string                `json:"assetId,omitempty"`
	Predecessor string                `json:"predecessor,omitempty"`
	Status      string                `json:"status"`
	Revision    uint64                `json:"revision"`
	Copies      int                   `json:"copies"`
	CreatedAt   time.Time             `json:"createdAt"`
	Attempts    []PrintAttemptSummary `json:"attempts"`
}
type PrintAttemptSummary struct {
	ID              string `json:"id"`
	ConnectorID     string `json:"connectorId"`
	Outcome         string `json:"outcome"`
	Reason          string `json:"reason,omitempty"`
	CompletedCopies int    `json:"completedCopies"`
}
type PrintSelectionSource interface {
	PrintDefaults(context.Context, Scope) (InventoryPrintDefaults, error)
	RegisteredPrinter(context.Context, Scope, string) (RegisteredPrinter, error)
}
type HumanPrintingAPI interface {
	PrintSelectionSource
	RegisteredPrinters(context.Context, Scope, Page) (Result[[]RegisteredPrinter], error)
	PrintJobs(context.Context, Scope, Page, string) (Result[[]PrintJobSummary], error)
	PrintJob(context.Context, Scope, string) (Result[PrintJobSummary], error)
	QueueLabel(context.Context, Scope, string, LabelPrintSelection, string) (Result[PrintJobSummary], error)
	TestPrinter(context.Context, Scope, LabelPrintSelection, string) (Result[PrintJobSummary], error)
	Reprint(context.Context, Scope, string, LabelPrintSelection, string) (Result[PrintJobSummary], error)
	CancelPrint(context.Context, Scope, string, uint64) (Result[PrintJobSummary], error)
}
