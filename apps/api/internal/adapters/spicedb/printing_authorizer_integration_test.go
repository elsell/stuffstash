package spicedb

import (
	"context"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"os"
	"testing"
	"time"
)

func TestSpiceDBIntegrationRestrictsPrintServicePrincipals(t *testing.T) {
	endpoint := os.Getenv(integrationEndpointEnv)
	if endpoint == "" {
		t.Skipf("%s is not set", integrationEndpointEnv)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	gateway, err := NewGateway(endpoint, "", false, "")
	if err != nil {
		t.Fatal(err)
	}
	defer gateway.Close()
	a := NewAuthorizer(gateway)
	if err := bootstrapIntegrationSchema(ctx, a); err != nil {
		t.Fatal(err)
	}
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	scope := printing.Scope{TenantID: "print-tenant-" + suffix, InventoryID: "print-inventory-" + suffix}
	c := printing.Connector{ID: printing.ConnectorID("connector-" + suffix), Scope: scope, ServiceAccountID: printing.ServiceAccountID("worker-" + suffix), State: printing.ConnectorActive}
	p := printing.Printer{ID: printing.PrinterID("printer-" + suffix), Scope: scope}
	other := printing.Printer{ID: printing.PrinterID("other-printer-" + suffix), Scope: printing.Scope{TenantID: "other-tenant-" + suffix, InventoryID: "other-inventory-" + suffix}}
	binding := printing.PrinterBinding{ConnectorID: c.ID, PrinterID: p.ID, Scope: scope}
	for _, printer := range []printing.Printer{p, other} {
		if err := a.SyncPrinterInventory(ctx, printer); err != nil {
			t.Fatal(err)
		}
	}
	if err := a.SyncPrintConnector(ctx, c, []printing.PrinterBinding{binding}); err != nil {
		t.Fatal(err)
	}
	assertAllowed(t, a.CheckPrintConnector(ctx, c.ServiceAccountID, c.ID), "registered connector can heartbeat")
	for _, permission := range []ports.PrinterPermission{ports.PrinterPermissionViewConsumer, ports.PrinterPermissionReport, ports.PrinterPermissionConsume} {
		assertAllowed(t, a.CheckPrinter(ctx, c.ServiceAccountID, p.ID, permission), "consumer assigned printer")
		assertForbidden(t, a.CheckPrinter(ctx, c.ServiceAccountID, other.ID, permission), "cross inventory printer")
		assertForbidden(t, a.CheckPrinter(ctx, "unrelated-service", p.ID, permission), "unrelated service")
	}
	assertForbidden(t, a.check(ctx, objectRef("printer", string(p.ID)), "consume", userSubject(principal(string(c.ServiceAccountID)))), "human with matching subject ID is not service")
	assertForbidden(t, a.check(ctx, objectRef("inventory", scope.InventoryID), "view", objectSubject("service_account", string(c.ServiceAccountID))), "worker cannot read inventory")
	binding.Revoked = true
	if err := a.SyncPrintConnector(ctx, c, []printing.PrinterBinding{binding}); err != nil {
		t.Fatal(err)
	}
	assertForbidden(t, a.CheckPrinter(ctx, c.ServiceAccountID, p.ID, ports.PrinterPermissionConsume), "revoked printer grant")
	c.State = printing.ConnectorRevoked
	if err := a.SyncPrintConnector(ctx, c, []printing.PrinterBinding{binding}); err != nil {
		t.Fatal(err)
	}
	assertForbidden(t, a.CheckPrintConnector(ctx, c.ServiceAccountID, c.ID), "revoked connector grant")
}
