package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

func realtimeVoiceInvestigationAsset(id string, title string, kind asset.Kind, parentID string) asset.Asset {
	item := assetItem(id, "tenant-home", "inventory-home", kind, parentID)
	itemTitle, ok := asset.NewTitle(title)
	if !ok {
		panic("invalid investigation test asset title")
	}
	item.Title = itemTitle
	return item
}

func realtimeVoiceInvestigationHasEvent(events []RealtimeVoiceEvent, eventType string) bool {
	for _, event := range events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

func seedRealtimeVoiceLoopAsset(t *testing.T, store interface {
	CreateAsset(context.Context, asset.Asset, audit.Record, *ports.UndoableOperation) error
}, item asset.Asset, auditID string) {
	t.Helper()
	if err := store.CreateAsset(context.Background(), item, audit.Record{ID: audit.ID(auditID), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: item.ID.String(), OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed asset %s: %v", item.ID, err)
	}
}

func checkoutToolSession() RealtimeVoiceSession {
	return RealtimeVoiceSession{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		Source:      RealtimeVoiceSourceMobile,
	}
}

func runRealtimeVoiceProductionEntrypoint(t *testing.T, application App) []RealtimeVoiceEvent {
	t.Helper()
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatalf("run production realtime voice entrypoint: %v", err)
	}
	return events
}

func realtimeVoiceInvestigationCompletedResponse(events []RealtimeVoiceEvent) *ports.StructuredAgentResponse {
	for _, event := range events {
		if event.Type == RealtimeVoiceEventAssistantResponseCompleted {
			return event.Response
		}
	}
	return nil
}

func realtimeVoiceInvestigationProposedPlan(events []RealtimeVoiceEvent) *RealtimeVoiceActionPlanProposal {
	for _, event := range events {
		if event.Type == RealtimeVoiceEventActionPlanProposed {
			return event.ActionPlan
		}
	}
	return nil
}
