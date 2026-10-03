package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func TestPairingRepositoryKeepsPendingGrantsDeniedAndExchangeConsumed(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t, ctx)
	now := time.Now().UTC()
	saveTenant(t, ctx, s, "tenant", "Home")
	saveInventory(t, ctx, s, "inventory", "tenant", "Garage")
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	p := printing.Pairing{ID: "pair", State: printing.PairingPending, CodeHash: "code", ExpiresAt: now.Add(time.Minute)}
	if err := s.CreatePrintPairing(ctx, p); err != nil {
		t.Fatal(err)
	}
	c := printing.Connector{ID: "connector", Scope: scope, State: printing.ConnectorPending, Generation: 1, CredentialVersion: 1}
	record := auditRecord(t, "approve", tenant.ID("tenant"), inventory.InventoryID("inventory"), audit.ActionPrinterRegistered)
	if _, err := s.ApprovePrintPairing(ctx, ports.PairingApproval{PairingID: p.ID, CodeHash: p.CodeHash, Now: now, Registration: ports.ConnectorRegistration{Connector: c}, Audit: record}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ConsumePrintPairing(ctx, ports.PairingExchange{PairingID: p.ID, Now: now, CredentialHash: "hash", CredentialExpiresAt: now.Add(time.Hour), ActivationDeadline: now.Add(time.Minute), Audit: auditRecord(t, "issued", tenant.ID("tenant"), inventory.InventoryID("inventory"), audit.ActionPrintConnectorCredentialIssued)}); err == nil {
		t.Fatal("pending grants issued credential")
	}
	pending, err := s.PendingPrintConnectorScopes(ctx, 10)
	if err != nil || len(pending) != 1 {
		t.Fatalf("pending reconciliation lost: %d %v", len(pending), err)
	}
	if err := s.SynchronizePrintConnector(ctx, scope, c.ID, func(context.Context, ports.ConnectorRegistration, []printing.Printer) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ConsumePrintPairing(ctx, ports.PairingExchange{PairingID: p.ID, Now: now, CredentialHash: "hash", CredentialExpiresAt: now.Add(time.Hour), ActivationDeadline: now.Add(time.Minute), Audit: auditRecord(t, "issued", tenant.ID("tenant"), inventory.InventoryID("inventory"), audit.ActionPrintConnectorCredentialIssued)}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ConsumePrintPairing(ctx, ports.PairingExchange{PairingID: p.ID, Now: now, CredentialHash: "hash", CredentialExpiresAt: now.Add(time.Hour), ActivationDeadline: now.Add(time.Minute), Audit: auditRecord(t, "issued", tenant.ID("tenant"), inventory.InventoryID("inventory"), audit.ActionPrintConnectorCredentialIssued)}); err == nil {
		t.Fatal("replayed credential issuance")
	}
	current, err := s.FindPrintConnectorCredential(ctx, "hash")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.HeartbeatPrintConnector(ctx, current, now.Add(2*time.Minute), nil, nil); err == nil {
		t.Fatal("expired activation accepted")
	}
}
