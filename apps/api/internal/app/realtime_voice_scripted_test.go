package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type scriptedRealtimeLanguageInference struct {
	turns                 []ports.LanguageInferenceTurn
	errs                  []error
	callCount             int
	seenTranscripts       []string
	seenConversationTurns [][]ports.AgentConversationTurn
	seenInvestigations    []*agentmodel.InvestigationInput
}

func (s *scriptedRealtimeLanguageInference) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	s.callCount++
	s.seenTranscripts = append(s.seenTranscripts, input.Transcript)
	s.seenConversationTurns = append(s.seenConversationTurns, append([]ports.AgentConversationTurn{}, input.ConversationTurns...))
	s.seenInvestigations = append(s.seenInvestigations, input.Investigation)
	if len(s.errs) > 0 {
		err := s.errs[0]
		s.errs = s.errs[1:]
		if err != nil {
			return ports.LanguageInferenceTurn{}, err
		}
	}
	if len(s.turns) == 0 {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	turn := s.turns[0]
	s.turns = s.turns[1:]
	return turn, nil
}

func seedRealtimeVoiceLoopAsset(t *testing.T, store interface {
	CreateAsset(context.Context, asset.Asset, audit.Record, *ports.UndoableOperation) error
}, item asset.Asset, auditID string) {
	t.Helper()
	if err := store.CreateAsset(context.Background(), item, audit.Record{ID: audit.ID(auditID), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: item.ID.String(), OccurredAt: time.Date(2026, 6, 26, 15, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed asset %s: %v", item.ID, err)
	}
}

func containsAll(text string, terms ...string) bool {
	for _, term := range terms {
		if !strings.Contains(text, term) {
			return false
		}
	}
	return true
}

func realtimeVoiceProgressStatuses(events []RealtimeVoiceEvent) []string {
	statuses := []string{}
	for _, event := range events {
		if event.Type == RealtimeVoiceEventAgentProgress {
			statuses = append(statuses, event.Status)
		}
	}
	return statuses
}

func slicesContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func assertRealtimeVoiceLocalCompletionOrder(t *testing.T, events []RealtimeVoiceEvent) {
	t.Helper()
	transcriptIndex := realtimeVoiceEventIndex(events, func(event RealtimeVoiceEvent) bool {
		return event.Type == RealtimeVoiceEventTranscriptFinal
	})
	understandingIndex := realtimeVoiceEventIndex(events, func(event RealtimeVoiceEvent) bool {
		return event.Type == RealtimeVoiceEventAgentProgress && event.Status == realtimeVoiceProgressUnderstanding
	})
	completedIndex := realtimeVoiceEventIndex(events, func(event RealtimeVoiceEvent) bool {
		return event.Type == RealtimeVoiceEventAssistantResponseCompleted
	})
	if transcriptIndex < 0 || understandingIndex < 0 || completedIndex < 0 || !(transcriptIndex < understandingIndex && understandingIndex < completedIndex) {
		t.Fatalf("expected transcript.final before understanding progress before assistant completion, got %+v", events)
	}
}

func realtimeVoiceEventIndex(events []RealtimeVoiceEvent, match func(RealtimeVoiceEvent) bool) int {
	for index, event := range events {
		if match(event) {
			return index
		}
	}
	return -1
}

func checkoutToolSession() RealtimeVoiceSession {
	return RealtimeVoiceSession{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		Source:      RealtimeVoiceSourceMobile,
	}
}
