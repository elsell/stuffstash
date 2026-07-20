package httpserver

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceQueryStreamsSanitizedLanguageFailureDiagnosticBeforeFailure(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"water-bottle-id", "voice-session-id", "tool-call-id"},
	}, fakeSpeechToText{transcript: "Move my water bottle to the kitchen."}, lateFailingLanguageModel{}, fakeTextToSpeech{chunks: [][]byte{[]byte("spoken-audio")}})
	seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Water bottle", "")
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	start := realtimeVoiceStartMessage("tenant-home", "inventory-home")
	start["developerDiagnostics"] = true
	events := runRealtimeVoiceQuestionUntilWithStart(t, server.URL, start, "user-1", "session.failed")

	diagnosticIndex := -1
	failedIndex := -1
	for index, event := range events {
		switch event["type"] {
		case "agent.diagnostic":
			if event["message"] == "Language provider failed" {
				diagnosticIndex = index
			}
		case "session.failed":
			failedIndex = index
		}
	}
	if diagnosticIndex < 0 || failedIndex < 0 || diagnosticIndex > failedIndex {
		t.Fatalf("expected language failure diagnostic before session.failed, got %+v", events)
	}
	detail, _ := events[diagnosticIndex]["detail"].(string)
	for _, required := range []string{`"phase": "evidence_assessment"`, `"evidenceRound": 1`, `"maxEvidenceRounds": 2`, `"previousRequestCount": 2`, `"toolResultCount": 2`, "search_authorized_assets", "provider_http_status_429"} {
		if !strings.Contains(detail, required) {
			t.Fatalf("expected diagnostic detail to include %q, got %s", required, detail)
		}
	}
	if strings.Contains(detail, "provider.invalid") || strings.Contains(detail, "should-not-leak") || strings.Contains(strings.ToLower(detail), "bearer ") || strings.Contains(detail, "finalOnly") || strings.Contains(detail, "previousTurns") {
		t.Fatalf("diagnostic leaked unsafe provider detail: %s", detail)
	}
}

type lateFailingLanguageModel struct{}

func (lateFailingLanguageModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if input.Investigation != nil && input.Investigation.Phase == agentmodel.InvestigationPhaseInitial {
		intent := agentmodel.Intent{
			RequestShape: agentmodel.RequestShapeSingleTarget,
			Kind:         agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "water bottle",
			DestinationPath: []string{"Kitchen"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation},
		}
		return typedVoiceInvestigationTurn(input, intent, nil)
	}
	return ports.LanguageInferenceTurn{}, safeHTTPStatusLanguageError{}
}

type safeHTTPStatusLanguageError struct{}

func (e safeHTTPStatusLanguageError) Error() string {
	return "raw provider response from https://provider.invalid bearer should-not-leak"
}

func (e safeHTTPStatusLanguageError) SafeRealtimeVoiceDiagnostic() string {
	return "provider_http_status_429"
}
