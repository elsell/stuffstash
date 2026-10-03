package printing

import (
	"testing"
	"time"
)

func TestPairingCannotApproveOrConsumeAfterExpiryOrReplay(t *testing.T) {
	now := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	p := Pairing{ID: "pair", State: PairingPending, ExpiresAt: now.Add(time.Minute)}
	if p.Approve(Scope{TenantID: "tenant", InventoryID: "inventory"}, "connector", now.Add(time.Minute)) {
		t.Fatal("expired pairing approved")
	}
	if !p.Approve(Scope{TenantID: "tenant", InventoryID: "inventory"}, "connector", now) {
		t.Fatal("valid pairing rejected")
	}
	if p.Approve(Scope{TenantID: "other", InventoryID: "inventory"}, "other", now) {
		t.Fatal("approved pairing rebound")
	}
	if p.Consume(now.Add(time.Minute)) {
		t.Fatal("expired pairing consumed")
	}
	if !p.Consume(now) {
		t.Fatal("approved pairing not consumed")
	}
	if p.Consume(now) {
		t.Fatal("credential exchange replay accepted")
	}
}
