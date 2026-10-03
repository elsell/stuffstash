package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

type PrinterPermission string

const (
	PrinterPermissionViewConsumer PrinterPermission = "view_consumer"
	PrinterPermissionReport       PrinterPermission = "report"
	PrinterPermissionConsume      PrinterPermission = "consume"
)

type PrintingAuthorization interface {
	CheckPrintConnector(context.Context, printing.ServiceAccountID, printing.ConnectorID) error
	CheckPrinter(context.Context, printing.ServiceAccountID, printing.PrinterID, PrinterPermission) error
	SyncPrintConnector(context.Context, printing.Connector, []printing.PrinterBinding) error
	SyncPrinterInventory(context.Context, printing.Printer) error
}
type PrinterCatalog interface {
	ListPrinterProfiles() []printing.PrinterProfile
	ResolvePrinterMedia(adapterID, presetID string, version uint32) (printing.MediaSnapshot, error)
}
