package memory

import (
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

func printingBindingKey(connectorID printing.ConnectorID, printerID printing.PrinterID) string {
	return string(connectorID) + ":" + string(printerID)
}

// Caller holds s.mu through the protected mutation.
func (s *Store) printingConsumerFenceLocked(authority printing.ConsumerAuthority, now time.Time) error {
	connector, ok := s.printingConnectors[authority.ConnectorID]
	if !ok {
		return ports.ErrPrintDenied
	}
	binding, ok := s.printingBindings[printingBindingKey(authority.ConnectorID, authority.PrinterID)]
	if !ok {
		return ports.ErrPrintDenied
	}
	if !printing.AcceptsAuthority(connector, binding, authority, now) {
		return ports.ErrPrintDenied
	}
	return nil
}
