package memory

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a *Authorizer) CheckPrintConnector(_ context.Context, principal printing.ServiceAccountID, id printing.ConnectorID) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.printingUnavailable {
		return errors.New("printing authorization unavailable")
	}
	if principal != "" && id != "" && a.printConnectorAgents[id] == principal {
		return nil
	}
	return ports.ErrForbidden
}
func (a *Authorizer) CheckPrinter(_ context.Context, principal printing.ServiceAccountID, id printing.PrinterID, permission ports.PrinterPermission) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.printingUnavailable {
		return errors.New("printing authorization unavailable")
	}
	switch permission {
	case ports.PrinterPermissionConsume, ports.PrinterPermissionReport, ports.PrinterPermissionViewConsumer:
	default:
		return ports.ErrForbidden
	}
	if principal != "" && id != "" && a.printConsumers[id][principal] {
		return nil
	}
	return ports.ErrForbidden
}
func (a *Authorizer) SyncPrinterInventory(_ context.Context, p printing.Printer) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.printingUnavailable {
		return errors.New("printing authorization unavailable")
	}
	if a.printInventories == nil {
		a.printInventories = map[printing.PrinterID]string{}
	}
	a.printInventories[p.ID] = p.Scope.InventoryID
	return nil
}
func (a *Authorizer) SyncPrintConnector(_ context.Context, c printing.Connector, bindings []printing.PrinterBinding) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.printingUnavailable {
		return errors.New("printing authorization unavailable")
	}
	for _, b := range bindings {
		if b.Scope != c.Scope || b.ConnectorID != c.ID {
			return ports.ErrPrintDenied
		}
	}
	if a.printConnectorAgents == nil {
		a.printConnectorAgents = map[printing.ConnectorID]printing.ServiceAccountID{}
	}
	if a.printConsumers == nil {
		a.printConsumers = map[printing.PrinterID]map[printing.ServiceAccountID]bool{}
	}
	active := c.State != printing.ConnectorRevoked
	if active {
		a.printConnectorAgents[c.ID] = c.ServiceAccountID
	} else {
		delete(a.printConnectorAgents, c.ID)
	}
	for _, b := range bindings {
		if a.printConsumers[b.PrinterID] == nil {
			a.printConsumers[b.PrinterID] = map[printing.ServiceAccountID]bool{}
		}
		if active && !b.Revoked {
			a.printConsumers[b.PrinterID][c.ServiceAccountID] = true
		} else {
			delete(a.printConsumers[b.PrinterID], c.ServiceAccountID)
		}
	}
	return nil
}

var _ ports.PrintingAuthorization = (*Authorizer)(nil)

// SetPrintingAvailable models an authorization service outage while preserving
// its relationships; recovery resumes from the same durable graph.
func (a *Authorizer) SetPrintingAvailable(available bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.printingUnavailable = !available
}
