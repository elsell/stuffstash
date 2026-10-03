package memory

import (
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"testing"
	"time"
)

func TestPrintingFenceRejectsRevokedStaleCrossScopeAndExpiredAuthority(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	scope := printing.Scope{TenantID: "tenant", InventoryID: "inventory"}
	authority := printing.ConsumerAuthority{Scope: scope, ConnectorID: "connector", ServiceAccountID: "service", CredentialVersion: 2, PrinterID: "printer", BindingGeneration: 3}
	store := NewStore()
	store.printingConnectors = map[printing.ConnectorID]printing.Connector{"connector": {ID: "connector", Scope: scope, ServiceAccountID: "service", State: printing.ConnectorActive, CredentialVersion: 2, CredentialExpiresAt: now.Add(time.Hour), Generation: 3, SyncedGeneration: 3}}
	store.printingBindings = map[string]printing.PrinterBinding{"connector:printer": {Scope: scope, ConnectorID: "connector", PrinterID: "printer", Generation: 3, SyncedGeneration: 3}}
	if err := store.printingConsumerFenceLocked(authority, now); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*printing.ConsumerAuthority){"cross tenant": func(a *printing.ConsumerAuthority) { a.Scope.TenantID = "other" }, "cross inventory": func(a *printing.ConsumerAuthority) { a.Scope.InventoryID = "other" }, "old credential": func(a *printing.ConsumerAuthority) { a.CredentialVersion-- }, "old binding": func(a *printing.ConsumerAuthority) { a.BindingGeneration-- }, "different principal": func(a *printing.ConsumerAuthority) { a.ServiceAccountID = "other" }} {
		t.Run(name, func(t *testing.T) {
			a := authority
			mutate(&a)
			if store.printingConsumerFenceLocked(a, now) == nil {
				t.Fatal("unexpected authorization")
			}
		})
	}
	c := store.printingConnectors["connector"]
	for _, state := range []printing.ConnectorState{printing.ConnectorPending, printing.ConnectorRevoked} {
		c.State = state
		store.printingConnectors[c.ID] = c
		if store.printingConsumerFenceLocked(authority, now) == nil {
			t.Fatal("inactive connector admitted")
		}
	}
	c.State = printing.ConnectorActive
	c.CredentialExpiresAt = now
	store.printingConnectors[c.ID] = c
	if store.printingConsumerFenceLocked(authority, now) == nil {
		t.Fatal("expired credential admitted")
	}
	c.CredentialExpiresAt = now.Add(time.Hour)
	store.printingConnectors[c.ID] = c
	b := store.printingBindings["connector:printer"]
	b.Revoked = true
	store.printingBindings["connector:printer"] = b
	if store.printingConsumerFenceLocked(authority, now) == nil {
		t.Fatal("revoked binding admitted")
	}
}
