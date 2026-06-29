package httpserver

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceActionPlanApprovalCompletesRealtimeSessionRecord(t *testing.T) {
	t.Parallel()

	application, store := newRealtimeVoiceReviewTestAppWithStore(t, actionPlanProposalLanguageModel{})
	ctx, connection, sessionID, planID := openRealtimeVoiceReviewSessionForApplication(t, application)

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.approve",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    planID,
	})
	_ = readRealtimeMessage(t, ctx, connection)
	_ = readRealtimeMessage(t, ctx, connection)

	record := realtimeSessionRecordEventually(t, store, sessionID)
	if record.State != ports.RealtimeSessionStateCompleted || record.EndedAt.IsZero() {
		t.Fatalf("expected approved review session to complete, got %+v", record)
	}
}

func TestRealtimeVoiceActionPlanCancellationCancelsRealtimeSessionRecord(t *testing.T) {
	t.Parallel()

	application, store := newRealtimeVoiceReviewTestAppWithStore(t, actionPlanProposalLanguageModel{})
	ctx, connection, sessionID, planID := openRealtimeVoiceReviewSessionForApplication(t, application)

	writeRealtimeMessage(t, ctx, connection, map[string]any{
		"type":      "action.plan.cancel",
		"seq":       4,
		"sessionId": sessionID,
		"planId":    planID,
	})
	cancelled := readRealtimeMessage(t, ctx, connection)
	if cancelled["type"] != "action.plan.cancelled" || cancelled["status"] != "cancelled" {
		t.Fatalf("expected cancelled action plan event, got %+v", cancelled)
	}

	record := realtimeSessionRecordEventually(t, store, sessionID)
	if record.State != ports.RealtimeSessionStateCancelled || record.EndedAt.IsZero() {
		t.Fatalf("expected cancelled review session to cancel realtime session, got %+v", record)
	}
}

func newRealtimeVoiceReviewTestAppWithStore(t *testing.T, languageInference ports.LanguageInferenceProvider) (app.App, *memory.Store) {
	t.Helper()

	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id", "plan-id", "command-id", "response-id"},
	}, store, authorizer).WithRealtimeVoiceProviders(fakeSpeechToText{transcript: "Add a water bottle."}, languageInference, fakeTextToSpeech{
		chunks: [][]byte{[]byte("spoken-audio")},
	})
	return application, store
}

func realtimeSessionRecordEventually(t *testing.T, store *memory.Store, sessionID string) ports.RealtimeSessionRecord {
	t.Helper()

	var record ports.RealtimeSessionRecord
	var found bool
	var err error
	for attempt := 0; attempt < 20; attempt++ {
		record, found, err = store.RealtimeSessionByID(context.Background(), tenant.ID("tenant-home"), inventory.InventoryID("inventory-home"), sessionID)
		if err != nil {
			t.Fatalf("read realtime session: %v", err)
		}
		if found && record.State != ports.RealtimeSessionStateStarted {
			return record
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !found {
		t.Fatalf("expected realtime session %q to exist", sessionID)
	}
	return record
}
