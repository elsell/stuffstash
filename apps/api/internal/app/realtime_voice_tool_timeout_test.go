package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceToolCallsUseBoundedTimeout(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-water-bottle",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "water bottle"},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindSafeFailure,
				SpokenResponse:  "I could not check that in time. Please try again.",
				DisplayResponse: "I could not check that in time. Please try again.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Where is my water bottle?"}
	resolver.providers.LanguageInference = language
	tts := &resolvedTextToSpeech{}
	resolver.providers.TextToSpeech = tts
	application, _ := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	search := &blockingAssetSearchRepository{ready: make(chan struct{})}
	application.search = search
	application.realtimeVoiceToolCallTimeout = time.Millisecond

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{
		Session:     session,
		AudioChunks: [][]byte{[]byte("audio")},
	}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("run realtime voice query should recover from timed-out tool call: %T %[1]v events=%+v", err, events)
	}
	if !search.cancelled {
		t.Fatalf("expected search tool context to be cancelled by tool-call timeout")
	}
	if !realtimeVoiceToolTimeoutEvent(events, "invalid_tool_request") {
		t.Fatalf("expected safe timed-out tool failure event, got %+v", events)
	}
	if tts.lastText == "" {
		t.Fatalf("expected safe recovery response to be synthesized")
	}
}

func TestRealtimeVoiceToolCallsDoNotRecoverParentDeadline(t *testing.T) {
	t.Parallel()

	application, _ := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
	application.search = &blockingAssetSearchRepository{ready: make(chan struct{})}
	application.realtimeVoiceToolCallTimeout = time.Minute
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	_, _, err = application.executeRealtimeVoiceTool(ctx, session, "Where is my water bottle?", nil, ports.AgentToolCall{
		ID:        "search-water-bottle",
		Name:      RealtimeVoiceToolSearchAuthorizedAssets,
		Arguments: map[string]any{"query": "water bottle"},
	}, map[string]struct{}{})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected parent deadline to propagate, got %T %[1]v", err)
	}
	if recoverableRealtimeVoiceToolError(err) {
		t.Fatalf("expected parent deadline to remain terminal, got recoverable error %T %[1]v", err)
	}
}

type blockingAssetSearchRepository struct {
	ready     chan struct{}
	cancelled bool
}

func (r *blockingAssetSearchRepository) SearchAssets(ctx context.Context, _ tenant.ID, _ []inventory.InventoryID, _ ports.AssetSearchPageRequest) ([]ports.AssetSearchResult, error) {
	close(r.ready)
	<-ctx.Done()
	r.cancelled = true
	return nil, ctx.Err()
}

func realtimeVoiceToolTimeoutEvent(events []RealtimeVoiceEvent, code string) bool {
	for _, event := range events {
		if event.Type == RealtimeVoiceEventToolCallFailed && event.Code == code {
			return true
		}
	}
	return false
}
