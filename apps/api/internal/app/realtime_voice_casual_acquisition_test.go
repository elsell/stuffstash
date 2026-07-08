package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceReadsDestinationBeforePlanningCasualAcquisition(t *testing.T) {
	t.Parallel()

	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:   "plan-phone-charger",
				Name: RealtimeVoiceToolProposeActionPlan,
				Arguments: map[string]any{
					"intentSummary":              "Create a phone charger in the office.",
					"modelInterpretationSummary": "The user got a phone charger and wants it stored in the visible Office location.",
					"confirmationSummary":        "Add a phone charger to the office?",
					"commands": []any{
						map[string]any{
							"id":      "cmd-phone-charger",
							"kind":    "create_asset",
							"summary": "Create phone charger in Office",
							"arguments": map[string]any{
								"title":         "phone charger",
								"kind":          "item",
								"parentAssetId": "office-1",
							},
						},
					},
				},
			}},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "I got a phone charger and put it inside the office."}
	resolver.providers.LanguageInference = language
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	office := assetItem("office-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	officeTitle, _ := asset.NewTitle("Office")
	office.Title = officeTitle
	seedRealtimeVoiceLoopAsset(t, store, office, "audit-office")

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	var proposed *RealtimeVoiceActionPlanProposal
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		if event.Type == RealtimeVoiceEventActionPlanProposed {
			proposed = event.ActionPlan
		}
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %T %[1]v", err)
	}
	if len(language.seenToolResults) != 1 || len(language.seenToolResults[0]) != 1 || language.seenToolResults[0][0].Name != RealtimeVoiceToolSearchAuthorizedAssets || !containsAll(language.seenToolResults[0][0].Content, `"query":"office"`, "office-1") {
		t.Fatalf("expected server-selected Office destination read for casual acquisition, got %+v", language.seenToolResults)
	}
	if proposed == nil || len(proposed.Commands) != 1 || proposed.Commands[0].ParentAssetID != "office-1" {
		t.Fatalf("expected casual acquisition proposal inside Office, got %+v", proposed)
	}
}
