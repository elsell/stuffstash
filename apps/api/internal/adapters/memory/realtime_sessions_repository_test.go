package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeSessionRepositoryScopesAndFreezesOutcomeUpdates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewStore()
	record := memoryRealtimeSessionRecord("voice-session-one", time.Date(2026, 6, 26, 16, 0, 0, 0, time.UTC))
	if err := store.SaveRealtimeSession(ctx, record); err != nil {
		t.Fatalf("save realtime session: %v", err)
	}

	outcome := ports.RealtimeSessionOutcome{State: ports.RealtimeSessionStateCompleted, At: record.StartedAt.Add(time.Second)}
	if err := store.UpdateRealtimeSessionOutcome(ctx, tenant.ID("tenant-other"), record.InventoryID, record.ID, outcome); err == nil {
		t.Fatalf("expected wrong tenant outcome update to fail")
	}
	if err := store.UpdateRealtimeSessionOutcome(ctx, record.TenantID, inventory.InventoryID("inventory-other"), record.ID, outcome); err == nil {
		t.Fatalf("expected wrong inventory outcome update to fail")
	}
	if err := store.UpdateRealtimeSessionOutcome(ctx, record.TenantID, record.InventoryID, record.ID, ports.RealtimeSessionOutcome{State: ports.RealtimeSessionStateCompleted, At: record.StartedAt.Add(-time.Second)}); err == nil {
		t.Fatalf("expected regressive timestamp outcome update to fail")
	}
	if err := store.UpdateRealtimeSessionOutcome(ctx, record.TenantID, record.InventoryID, record.ID, outcome); err != nil {
		t.Fatalf("update outcome: %v", err)
	}
	if err := store.UpdateRealtimeSessionOutcome(ctx, record.TenantID, record.InventoryID, record.ID, ports.RealtimeSessionOutcome{State: ports.RealtimeSessionStateFailed, At: record.StartedAt.Add(2 * time.Second), SafeFailureCode: "voice_session_failed"}); err == nil {
		t.Fatalf("expected final session outcome to be immutable")
	}
}

func memoryRealtimeSessionRecord(id string, startedAt time.Time) ports.RealtimeSessionRecord {
	return ports.RealtimeSessionRecord{
		ID:                         id,
		TenantID:                   tenant.ID("tenant-home"),
		InventoryID:                inventory.InventoryID("inventory-home"),
		PrincipalID:                identity.PrincipalID("user-1"),
		Source:                     "mobile_voice",
		State:                      ports.RealtimeSessionStateStarted,
		SpeechToTextProfileID:      "stt-profile",
		LanguageInferenceProfileID: "lm-profile",
		TextToSpeechProfileID:      "tts-profile",
		StartedAt:                  startedAt,
		LastActivityAt:             startedAt,
	}
}
