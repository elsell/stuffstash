package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceInvestigationLoopAnswersFromPlausibleApproximateMatch(t *testing.T) {
	t.Parallel()
	initial := agentmodel.InvestigationStep{
		Decision:       agentmodel.InvestigationDecisionSearch,
		Intent:         agentmodel.Intent{Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "Sarah winter coat"},
		SearchRequests: []agentmodel.SearchRequest{{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Sarah winter coat", SearchProbes: []string{"Sarah", "winter clothes"}}},
	}
	final := agentmodel.InvestigationStep{
		Decision: agentmodel.InvestigationDecisionFinish, Intent: initial.Intent,
		Resolutions: []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionPlausible, CandidateIDs: []string{"winter-clothes-1"}}},
	}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{Investigation: &initial}, {Investigation: &final}}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.LanguageInference = language
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Where are Sarah's winter coat?"}
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	closet := realtimeVoiceInvestigationAsset("closet-1", "Hall closet", asset.KindContainer, "")
	clothes := realtimeVoiceInvestigationAsset("winter-clothes-1", "Sarah Winter Clothes and Shoes", asset.KindItem, closet.ID.String())
	seedRealtimeVoiceLoopAsset(t, store, closet, "audit-closet")
	seedRealtimeVoiceLoopAsset(t, store, clothes, "audit-clothes")
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("run investigation loop: %v", err)
	}
	response := realtimeVoiceInvestigationCompletedResponse(events)
	if response == nil || response.Kind != ports.StructuredAgentResponseKindAnswer || response.SpokenResponse == "" {
		t.Fatalf("expected grounded answer, got %+v", events)
	}
	if got := resolver.providers.TextToSpeech.(*resolvedTextToSpeech).lastText; got != response.SpokenResponse {
		t.Fatalf("expected only grounded response to be spoken, got %q", got)
	}
	if len(language.seenTools) != 2 || len(language.seenTools[0]) != 0 || len(language.seenTools[1]) != 0 {
		t.Fatalf("investigation turns must not expose tools: %+v", language.seenTools)
	}
}

func TestRealtimeVoiceInvestigationLoopCompilesNestedMissingDestinationAndStopsAtReview(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{
		Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "Drill",
		DestinationPath:  []string{"Garage", "Blue cabinet", "Upper shelf"},
		DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation, agentmodel.DestinationKindContainer, agentmodel.DestinationKindContainer},
	}
	initial := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionSearch, Intent: intent, SearchRequests: []agentmodel.SearchRequest{
		{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Drill", SearchProbes: []string{"drill"}},
		{ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Garage", SearchProbes: []string{"garage"}},
		{ReferenceKey: "destination.1", ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Blue cabinet", SearchProbes: []string{"blue cabinet"}},
		{ReferenceKey: "destination.2", ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Upper shelf", SearchProbes: []string{"upper shelf"}},
	}}
	final := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: []agentmodel.Resolution{
		{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"drill-1"}},
		{ReferenceKey: "destination.0", Status: agentmodel.ResolutionMissing},
		{ReferenceKey: "destination.1", Status: agentmodel.ResolutionMissing},
		{ReferenceKey: "destination.2", Status: agentmodel.ResolutionMissing},
	}}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{Investigation: &initial}, {Investigation: &final}}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.LanguageInference = language
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move the drill to the upper shelf in the blue cabinet in the garage"}
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	drill := realtimeVoiceInvestigationAsset("drill-1", "Drill", asset.KindItem, "")
	seedRealtimeVoiceLoopAsset(t, store, drill, "audit-drill")
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("run investigation loop: %v", err)
	}
	proposal := realtimeVoiceInvestigationProposedPlan(events)
	if proposal == nil || len(proposal.Commands) != 4 {
		t.Fatalf("expected create/create/create/move review plan, got %+v", events)
	}
	if proposal.Commands[0].Kind != "create_location" || proposal.Commands[3].Kind != "move_asset" {
		t.Fatalf("unexpected compiled command graph: %+v", proposal.Commands)
	}
	if len(language.seenInvestigations) != 2 || language.seenInvestigations[1] == nil {
		t.Fatalf("expected evidence assessment input, got %+v", language.seenInvestigations)
	}
	readEvidence := language.seenInvestigations[1].ReadEvidence
	if len(readEvidence) != 4 {
		t.Fatalf("expected one accumulated safe-read record for every semantic reference, got %+v", readEvidence)
	}
	for _, reference := range []agentmodel.SemanticReferenceKey{agentmodel.SemanticReferenceSubject, "destination.0", "destination.1", "destination.2"} {
		found := false
		for _, evidence := range readEvidence {
			if evidence.ReferenceKey == reference {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected executed read evidence for %s, got %+v", reference, readEvidence)
		}
	}
	for _, event := range events {
		if event.Type == RealtimeVoiceEventAssistantResponseCompleted || event.Type == RealtimeVoiceEventTextToSpeechAudioStarted || event.Type == RealtimeVoiceEventSessionCompleted {
			t.Fatalf("loop must pause silently at review, got %+v", events)
		}
	}
	if got := resolver.providers.TextToSpeech.(*resolvedTextToSpeech).lastText; got != "" {
		t.Fatalf("action plan must not be spoken before approval, got %q", got)
	}
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
