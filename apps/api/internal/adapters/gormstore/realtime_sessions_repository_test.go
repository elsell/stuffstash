package gormstore

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeSessionRepositorySavesAndUpdatesSafeMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	startedAt := time.Date(2026, 6, 26, 16, 0, 0, 0, time.UTC)
	record := realtimeSessionRecord("voice-session-one", startedAt)

	if err := store.SaveRealtimeSession(ctx, record); err != nil {
		t.Fatalf("save realtime session: %v", err)
	}
	endedAt := startedAt.Add(3 * time.Second)
	if err := store.UpdateRealtimeSessionOutcome(ctx, tenant.ID("tenant-home"), inventory.InventoryID("inventory-home"), record.ID, ports.RealtimeSessionOutcome{State: ports.RealtimeSessionStateCompleted, At: endedAt}); err != nil {
		t.Fatalf("update realtime session outcome: %v", err)
	}

	got, found, err := store.RealtimeSessionByID(ctx, tenant.ID("tenant-home"), inventory.InventoryID("inventory-home"), record.ID)
	if err != nil {
		t.Fatalf("read realtime session: %v", err)
	}
	if !found {
		t.Fatalf("expected realtime session to be found")
	}
	if got.State != ports.RealtimeSessionStateCompleted || !got.EndedAt.Equal(endedAt) || !got.LastActivityAt.Equal(endedAt) {
		t.Fatalf("unexpected completed metadata: %+v", got)
	}
	if got.TenantID != record.TenantID || got.InventoryID != record.InventoryID || got.PrincipalID != record.PrincipalID {
		t.Fatalf("unexpected session scope metadata: %+v", got)
	}
}

func TestRealtimeSessionRepositoryScopesReadsByTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveTenant(t, ctx, store, tenant.ID("tenant-other"), "Other")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	if err := store.SaveRealtimeSession(ctx, realtimeSessionRecord("voice-session-one", time.Now().UTC())); err != nil {
		t.Fatalf("save realtime session: %v", err)
	}

	if _, found, err := store.RealtimeSessionByID(ctx, tenant.ID("tenant-other"), inventory.InventoryID("inventory-home"), "voice-session-one"); err != nil || found {
		t.Fatalf("expected wrong tenant read to miss, found=%t err=%v", found, err)
	}
	if _, found, err := store.RealtimeSessionByID(ctx, tenant.ID("tenant-home"), inventory.InventoryID("inventory-other"), "voice-session-one"); err != nil || found {
		t.Fatalf("expected wrong inventory read to miss, found=%t err=%v", found, err)
	}
}

func TestRealtimeSessionRepositoryScopesAndFreezesOutcomeUpdates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	saveTenant(t, ctx, store, tenant.ID("tenant-home"), "Home")
	saveTenant(t, ctx, store, tenant.ID("tenant-other"), "Other")
	saveInventory(t, ctx, store, "inventory-home", tenant.ID("tenant-home"), "Home")
	record := realtimeSessionRecord("voice-session-one", time.Date(2026, 6, 26, 16, 0, 0, 0, time.UTC))
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

func TestRealtimeSessionRepositoryStoresOnlySafeColumns(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t, ctx)
	disallowed := []string{
		"audio",
		"audio_chunks",
		"transcript",
		"prompt",
		"provider_prompt",
		"provider_response",
		"model_response",
		"generated_speech",
		"credential",
		"bearer_token",
		"provider_session_id",
	}
	for _, column := range disallowed {
		if store.db.WithContext(ctx).Migrator().HasColumn(&realtimeSessionModel{}, column) {
			t.Fatalf("realtime session model must not persist unsafe column %q", column)
		}
	}
}

func realtimeSessionRecord(id string, startedAt time.Time) ports.RealtimeSessionRecord {
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
