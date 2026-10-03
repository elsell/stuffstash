package gormstore

import (
	"context"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"testing"
	"time"
)

func TestPairingCleanupIsBoundedAndPreservesUnexpiredRequests(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t, ctx)
	now := time.Now().UTC()
	for _, pair := range []p.Pairing{{ID: "old", CodeHash: "old", ExpiresAt: now.Add(-time.Minute), State: p.PairingPending}, {ID: "edge", CodeHash: "edge", ExpiresAt: now, State: p.PairingConsumed}, {ID: "live", CodeHash: "live", ExpiresAt: now.Add(time.Minute), State: p.PairingApproved}} {
		if err := s.CreatePrintPairing(ctx, pair); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CleanupExpiredPrintPairings(ctx, now, 0); err == nil {
		t.Fatal("unbounded cleanup accepted")
	}
	if count, err := s.CleanupExpiredPrintPairings(ctx, now, 1); err != nil || count != 1 {
		t.Fatal("first bounded pass", count, err)
	}
	if _, err := s.GetPrintPairing(ctx, "live"); err != nil {
		t.Fatal("deleted unexpired request", err)
	}
	if count, err := s.CleanupExpiredPrintPairings(ctx, now, 1); err != nil || count != 1 {
		t.Fatal("second bounded pass", count, err)
	}
	for _, id := range []p.PairingID{"old", "edge"} {
		if _, err := s.GetPrintPairing(ctx, id); err == nil {
			t.Fatal("retained expired handshake", id)
		}
	}
	if count, err := s.CleanupExpiredPrintPairings(ctx, now, 1); err != nil || count != 0 {
		t.Fatal("removed fresh request", count, err)
	}
}
