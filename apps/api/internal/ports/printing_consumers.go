package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
)

// PrintConsumerAccess authenticates the service principal and checks current,
// fully consistent printer grants. Repository fences remain mandatory on writes.
type PrintConsumerAccess interface {
	AuthenticateConsumer(context.Context, string) (printing.Connector, error)
	AuthorizePrinter(context.Context, printing.Connector, printing.PrinterID, PrinterPermission) (printing.ConsumerAuthority, error)
}
