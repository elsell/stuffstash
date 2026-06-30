package app

import (
	"context"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceStreamsVerboseAgentDiagnostics(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			Diagnostics: []ports.LanguageInferenceDiagnostic{{
				Title:  "Language prompt",
				Detail: "Transcript: move my water bottle to the kitchen\napiKey: should-not-leak\nBearer abc123",
			}},
			ToolCalls: []ports.AgentToolCall{{
				ID:   "plan-kitchen",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"commandKind":                "create_asset",
					"intentSummary":              "Move the water bottle to a new Kitchen location.",
					"modelInterpretationSummary": "The user wants the visible water bottle moved to Kitchen, which should be created.",
					"confirmationSummary":        "Create Kitchen and move the water bottle there?",
					"commandSummary":             "Create Kitchen",
					"arguments": map[string]any{
						"title": "Kitchen",
						"kind":  "location",
					},
				},
			}},
		},
		{
			Diagnostics: []ports.LanguageInferenceDiagnostic{{
				Title:  "Language model turn",
				Detail: `{"final":{"kind":"answer","spokenResponse":"I checked your inventory.","displayResponse":"I checked your inventory."}}`,
			}},
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I checked your inventory.",
				DisplayResponse: "I checked your inventory.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my water bottle to the kitchen."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	seedRealtimeVoiceLoopAsset(t, store, waterBottle, "audit-water-bottle")

	sessionInput := defaultRealtimeVoiceSessionInput()
	sessionInput.DeveloperDiagnostics = true
	session, err := application.StartRealtimeVoiceSession(context.Background(), sessionInput)
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}

	var promptDiagnostic, toolStarted, toolCompleted, toolResultDiagnostic *RealtimeVoiceEvent
	for index := range events {
		event := &events[index]
		switch {
		case event.Type == RealtimeVoiceEventAgentDiagnostic && event.Message == "Language prompt":
			promptDiagnostic = event
		case event.Type == RealtimeVoiceEventToolCallStarted:
			toolStarted = event
		case event.Type == RealtimeVoiceEventToolCallCompleted:
			toolCompleted = event
		case event.Type == RealtimeVoiceEventAgentDiagnostic && event.Message == "Tool result received":
			toolResultDiagnostic = event
		}
	}
	if promptDiagnostic == nil || !strings.Contains(promptDiagnostic.Detail, "move my water bottle to the kitchen") {
		t.Fatalf("expected prompt diagnostic, got %+v", events)
	}
	if strings.Contains(promptDiagnostic.Detail, "should-not-leak") || strings.Contains(promptDiagnostic.Detail, "abc123") || !strings.Contains(promptDiagnostic.Detail, "[redacted-key]") {
		t.Fatalf("expected redacted prompt diagnostic, got %q", promptDiagnostic.Detail)
	}
	if toolStarted == nil || toolStarted.Detail != "" {
		t.Fatalf("expected bland tool start event, got %+v", toolStarted)
	}
	if toolCompleted == nil || toolCompleted.Detail != "" {
		t.Fatalf("expected bland tool completed event, got %+v", toolCompleted)
	}
	if toolResultDiagnostic == nil || !strings.Contains(toolResultDiagnostic.Detail, `"content"`) || !strings.Contains(toolResultDiagnostic.Detail, "propose_action_plan") {
		t.Fatalf("expected tool result diagnostic, got %+v", toolResultDiagnostic)
	}
}

func TestRealtimeVoiceOmitsVerboseDiagnosticsByDefault(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			Diagnostics: []ports.LanguageInferenceDiagnostic{{
				Title:  "Language prompt",
				Detail: "Transcript: move my water bottle to the kitchen\nBearer should-not-leak",
			}},
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "I checked your inventory.",
				DisplayResponse: "I checked your inventory.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my water bottle to the kitchen."}
	resolver.providers.LanguageInference = language
	application := newRealtimeVoiceResolutionTestApp(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	for _, event := range events {
		if event.Type == RealtimeVoiceEventAgentDiagnostic || event.Detail != "" {
			t.Fatalf("expected default session to omit verbose diagnostics, got %+v from %+v", event, events)
		}
	}
}

func TestRealtimeVoiceOmitsRecoverableToolErrorDiagnosticsByDefault(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-1",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"commandKind":                "create_asset",
					"intentSummary":              "Create an item.",
					"modelInterpretationSummary": "The user wants to add an item.",
					"confirmationSummary":        "Create item?",
					"commandSummary":             "Create item",
					"arguments": map[string]any{
						"apiKey": "secret",
					},
				},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I could not prepare that change safely.",
				DisplayResponse: "I could not prepare that change safely.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Add an item."}
	resolver.providers.LanguageInference = language
	application := newRealtimeVoiceResolutionTestApp(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	for _, event := range events {
		if event.Type == RealtimeVoiceEventAgentDiagnostic || event.Detail != "" {
			t.Fatalf("expected default recoverable tool failure to omit verbose diagnostics, got %+v from %+v", event, events)
		}
	}
}

func TestRealtimeVoiceReturnsRecoverableToolErrorsToModel(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "tool-plan-1",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"commandKind":                "create_asset",
					"intentSummary":              "Create an item.",
					"modelInterpretationSummary": "The user wants to add an item.",
					"confirmationSummary":        "Create item?",
					"commandSummary":             "Create item",
					"arguments": map[string]any{
						"apiKey": "secret",
					},
				},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindClarification,
				SpokenResponse:  "I could not prepare that change safely. Please try again with the item name and location.",
				DisplayResponse: "I could not prepare that change safely. Please try again with the item name and location.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Add an item."}
	resolver.providers.LanguageInference = language
	tts := &resolvedTextToSpeech{}
	resolver.providers.TextToSpeech = tts
	application := newRealtimeVoiceResolutionTestApp(t, resolver)

	sessionInput := defaultRealtimeVoiceSessionInput()
	sessionInput.DeveloperDiagnostics = true
	session, err := application.StartRealtimeVoiceSession(context.Background(), sessionInput)
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if len(language.seenToolResults) < 2 || len(language.seenToolResults[1]) != 1 {
		t.Fatalf("expected recoverable tool error to be returned to model, got %+v", language.seenToolResults)
	}
	if !strings.Contains(language.seenToolResults[1][0].Content, `"status":"error"`) || !strings.Contains(language.seenToolResults[1][0].Content, `"retryable":true`) {
		t.Fatalf("expected safe retryable tool error result, got %s", language.seenToolResults[1][0].Content)
	}
	if !strings.Contains(language.seenToolResults[1][0].Content, "parentCommandId") || !strings.Contains(language.seenToolResults[1][0].Content, "missing location/container") {
		t.Fatalf("expected action-plan repair guidance in tool error result, got %s", language.seenToolResults[1][0].Content)
	}
	if _, leaked := language.seenToolResults[1][0].Call.Arguments["apiKey"]; leaked {
		t.Fatalf("rejected tool arguments leaked into provider-bound call history: %+v", language.seenToolResults[1][0].Call.Arguments)
	}
	seenFailureEvent := false
	seenFailureDiagnostic := false
	failureDiagnosticDetail := ""
	for _, event := range events {
		if event.Type == RealtimeVoiceEventToolCallFailed && event.Code == "invalid_tool_request" {
			seenFailureEvent = true
		}
		if event.Type == RealtimeVoiceEventAgentDiagnostic && event.Message == "Tool result received" && strings.Contains(event.Detail, `\"status\":\"error\"`) && strings.Contains(event.Detail, `\"retryable\":true`) {
			seenFailureDiagnostic = true
			failureDiagnosticDetail = event.Detail
		}
	}
	if !seenFailureEvent {
		t.Fatalf("expected safe tool failure event, got %+v", events)
	}
	if !seenFailureDiagnostic {
		t.Fatalf("expected safe tool failure diagnostic, got %+v", events)
	}
	if strings.Contains(failureDiagnosticDetail, "apiKey") || strings.Contains(failureDiagnosticDetail, "secret") {
		t.Fatalf("expected failed tool diagnostic to omit rejected sensitive arguments, got %q", failureDiagnosticDetail)
	}
	if tts.lastText != "I could not prepare that change safely. Please try again with the item name and location." {
		t.Fatalf("expected recovered final response to be spoken, got %q", tts.lastText)
	}
}

func TestRealtimeVoiceEmitsDiagnosticWhenLanguageContinuationFailsAfterTools(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{
		turns: []ports.LanguageInferenceTurn{{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-1",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "water bottle"},
			}},
		}},
		errs: []error{nil, safeRealtimeVoiceDiagnosticFailure{safe: "provider_http_status_429"}},
	}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my water bottle to the kitchen."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	seedRealtimeVoiceLoopAsset(t, store, waterBottle, "audit-water-bottle")

	sessionInput := defaultRealtimeVoiceSessionInput()
	sessionInput.DeveloperDiagnostics = true
	session, err := application.StartRealtimeVoiceSession(context.Background(), sessionInput)
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if err == nil || RealtimeVoiceSafeErrorCode(err) != realtimeVoiceFailureLanguageInference {
		t.Fatalf("expected language inference failure, got %v", err)
	}

	var diagnostic *RealtimeVoiceEvent
	for index := range events {
		event := &events[index]
		if event.Type == RealtimeVoiceEventAgentDiagnostic && event.Message == "Language provider failed" {
			diagnostic = event
			break
		}
	}
	if diagnostic == nil {
		t.Fatalf("expected failed continuation diagnostic, got %+v", events)
	}
	for _, required := range []string{`"turn": 2`, `"previousTurns": 1`, `"finalOnly": false`, `"toolResultCount": 1`, RealtimeVoiceToolSearchAuthorizedAssets, "language_inference_failed", "provider_http_status_429"} {
		if !strings.Contains(diagnostic.Detail, required) {
			t.Fatalf("expected diagnostic to contain %q, got %s", required, diagnostic.Detail)
		}
	}
	if strings.Contains(diagnostic.Detail, "should-not-leak") || strings.Contains(strings.ToLower(diagnostic.Detail), "bearer ") || strings.Contains(diagnostic.Detail, "provider.invalid") {
		t.Fatalf("expected provider failure diagnostic to be redacted, got %s", diagnostic.Detail)
	}
}

func TestRealtimeVoiceDowngradesUnsafeProviderDiagnosticCategories(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{
		turns: []ports.LanguageInferenceTurn{{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-1",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "water bottle"},
			}},
		}},
		errs: []error{nil, safeRealtimeVoiceDiagnosticFailure{safe: "provider_http_status_429 https://provider.invalid project-id should-not-leak"}},
	}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my water bottle to the kitchen."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	seedRealtimeVoiceLoopAsset(t, store, waterBottle, "audit-water-bottle")

	sessionInput := defaultRealtimeVoiceSessionInput()
	sessionInput.DeveloperDiagnostics = true
	session, err := application.StartRealtimeVoiceSession(context.Background(), sessionInput)
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if err == nil {
		t.Fatalf("expected language inference failure")
	}
	diagnostic := findRealtimeVoiceDiagnosticEvent(t, events, "Language provider failed")
	if !strings.Contains(diagnostic.Detail, "provider_request_failed") {
		t.Fatalf("expected unsafe provider diagnostic to downgrade, got %s", diagnostic.Detail)
	}
	if strings.Contains(diagnostic.Detail, "provider.invalid") || strings.Contains(diagnostic.Detail, "project-id") || strings.Contains(diagnostic.Detail, "should-not-leak") {
		t.Fatalf("unsafe provider diagnostic leaked, got %s", diagnostic.Detail)
	}
}

func findRealtimeVoiceDiagnosticEvent(t *testing.T, events []RealtimeVoiceEvent, message string) RealtimeVoiceEvent {
	t.Helper()

	for _, event := range events {
		if event.Type == RealtimeVoiceEventAgentDiagnostic && event.Message == message {
			return event
		}
	}
	t.Fatalf("expected diagnostic %q, got %+v", message, events)
	return RealtimeVoiceEvent{}
}

type safeRealtimeVoiceDiagnosticFailure struct {
	safe string
}

func (e safeRealtimeVoiceDiagnosticFailure) Error() string {
	return "provider failure with raw endpoint https://provider.invalid bearer should-not-leak"
}

func (e safeRealtimeVoiceDiagnosticFailure) SafeRealtimeVoiceDiagnostic() string {
	return e.safe
}
