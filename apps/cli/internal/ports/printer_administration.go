package ports

import "context"

type PrinterAdministrationAPI interface {
	CreatePrinter(context.Context, Scope, string, []byte) (Result[RegisteredPrinter], error)
	UpdatePrinter(context.Context, Scope, string, []byte) (Result[RegisteredPrinter], error)
	UpdatePrintConnector(context.Context, Scope, string, []byte) (Result[PrintConnector], error)
	UpdatePrintSettings(context.Context, Scope, []byte) (Result[PrintSettings], error)
}
