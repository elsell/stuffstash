package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

type PrinterMutation func(*printing.Printer) error

// PrinterAudit is evaluated under the mutation lock. History and mutation commit
// atomically; generating or storing history can fail the whole command.
type PrinterAudit func(printing.Printer) (audit.Record, error)
type PrinterRepository interface {
	CreatePrinter(context.Context, printing.Printer, audit.Record) (printing.Printer, bool, error)
	GetPrinter(context.Context, printing.Scope, printing.PrinterID) (printing.Printer, error)
	ListPrinters(context.Context, printing.Scope, int, string) ([]printing.Printer, error)
	UpdatePrinter(context.Context, printing.Scope, printing.PrinterID, uint64, PrinterMutation, PrinterAudit) (printing.Printer, error)
}
