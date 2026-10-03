package spicedb

import (
	"context"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Authorizer) CheckPrintConnector(ctx context.Context, principal printing.ServiceAccountID, id printing.ConnectorID) error {
	if principal == "" || id == "" {
		return ports.ErrForbidden
	}
	return a.check(ctx, objectRef("print_connector", string(id)), "report", objectSubject("service_account", string(principal)))
}
func (a Authorizer) CheckPrinter(ctx context.Context, principal printing.ServiceAccountID, id printing.PrinterID, permission ports.PrinterPermission) error {
	if principal == "" || id == "" {
		return ports.ErrForbidden
	}
	switch permission {
	case ports.PrinterPermissionViewConsumer, ports.PrinterPermissionReport, ports.PrinterPermissionConsume:
	default:
		return ports.ErrForbidden
	}
	return a.check(ctx, objectRef("printer", string(id)), string(permission), objectSubject("service_account", string(principal)))
}
func (a Authorizer) SyncPrinterInventory(ctx context.Context, p printing.Printer) error {
	return a.touchRelationships(ctx, relationship(objectRef("printer", string(p.ID)), "inventory", objectSubject(objectTypeInventory, p.Scope.InventoryID)))
}

// The caller serializes all generations of this connector and includes revoked
// bindings, so retrying the complete desired state also removes stale grants.
func (a Authorizer) SyncPrintConnector(ctx context.Context, c printing.Connector, bindings []printing.PrinterBinding) error {
	active := c.State != printing.ConnectorRevoked
	updates := []*v1.RelationshipUpdate{{Operation: v1.RelationshipUpdate_OPERATION_TOUCH, Relationship: relationship(objectRef("print_connector", string(c.ID)), "inventory", objectSubject(objectTypeInventory, c.Scope.InventoryID))}}
	change := func(resource *v1.ObjectReference, relation string, enabled bool) {
		op := v1.RelationshipUpdate_OPERATION_DELETE
		if enabled {
			op = v1.RelationshipUpdate_OPERATION_TOUCH
		}
		updates = append(updates, &v1.RelationshipUpdate{Operation: op, Relationship: relationship(resource, relation, objectSubject("service_account", string(c.ServiceAccountID)))})
	}
	change(objectRef("print_connector", string(c.ID)), "agent", active)
	for _, binding := range bindings {
		if binding.Scope != c.Scope || binding.ConnectorID != c.ID {
			return ports.ErrPrintDenied
		}
		change(objectRef("printer", string(binding.PrinterID)), "consumer", active && !binding.Revoked)
	}
	_, err := a.gateway.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{Updates: updates})
	return err
}

var _ ports.PrintingAuthorization = Authorizer{}
