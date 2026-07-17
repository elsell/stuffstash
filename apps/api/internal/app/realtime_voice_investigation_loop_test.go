package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type unshapedRealtimeLanguageInference struct{}

func (unshapedRealtimeLanguageInference) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	step := agentmodel.InvestigationStep{
		Decision: agentmodel.InvestigationDecisionSearch,
		Intent:   agentmodel.Intent{Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "drill"},
		SearchRequests: []agentmodel.SearchRequest{{
			ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets,
			Mention: "drill", SearchProbes: []string{"drill"},
		}},
	}
	return ports.LanguageInferenceTurn{Investigation: &step}, nil
}

func TestRealtimeVoiceInvestigationRejectsProviderIntentWithoutExplicitRequestShape(t *testing.T) {
	t.Parallel()

	session := RealtimeVoiceSession{languageInference: unshapedRealtimeLanguageInference{}}
	input := agentmodel.InvestigationInput{
		Phase: agentmodel.InvestigationPhaseInitial, PromptVersion: realtimeVoiceInvestigationVersion,
		SchemaVersion: realtimeVoiceInvestigationVersion, Transcript: "Where is the drill?", MaxEvidenceRounds: agentmodel.MaxEvidenceRounds,
	}
	_, err := (App{}).nextRealtimeVoiceInvestigation(context.Background(), session, input.Transcript, nil, input, nil, func(RealtimeVoiceEvent) error { return nil })
	if !errors.Is(err, ports.ErrInvalidProviderInput) {
		t.Fatalf("missing explicit request shape error = %v, want invalid provider input", err)
	}
}

func TestRealtimeVoiceInvestigationLoopAnswersFromPlausibleApproximateMatch(t *testing.T) {
	t.Parallel()
	initial := agentmodel.InvestigationStep{
		Decision:       agentmodel.InvestigationDecisionSearch,
		Intent:         agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "Sarah winter coat"},
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
	if language.callCount != 2 {
		t.Fatalf("expected initial and evidence investigation calls, got %d", language.callCount)
	}
}

func TestRealtimeVoiceInvestigationLoopReturnsUnsupportedActionAfterInitialClassification(t *testing.T) {
	t.Parallel()
	for _, shape := range []agentmodel.RequestShape{agentmodel.RequestShapeCollectionTarget, agentmodel.RequestShapeCompound} {
		shape := shape
		t.Run(string(shape), func(t *testing.T) {
			t.Parallel()
			intent := agentmodel.Intent{
				RequestShape: shape, Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "requested assets",
				DestinationPath: []string{"destination"}, DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation},
			}
			initial := agentmodel.InvestigationStep{
				Decision: agentmodel.InvestigationDecisionSearch, Intent: intent,
				SearchRequests: []agentmodel.SearchRequest{{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "requested assets", SearchProbes: []string{"requested assets"}}},
			}
			language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{Investigation: &initial}}}
			resolver := successfulRealtimeVoiceResolver()
			resolver.providers.LanguageInference = language
			resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Generated unsupported request"}
			application, _ := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
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
				t.Fatalf("run unsupported investigation: %v", err)
			}
			response := realtimeVoiceInvestigationCompletedResponse(events)
			if response == nil || response.Kind != ports.StructuredAgentResponseKindUnsupportedAction {
				t.Fatalf("expected bounded unsupported response, got %+v", events)
			}
			if language.callCount != 1 {
				t.Fatalf("expected one semantic classification turn, got %d", language.callCount)
			}
			if realtimeVoiceInvestigationHasEvent(events, RealtimeVoiceEventToolCallStarted) || realtimeVoiceInvestigationHasEvent(events, RealtimeVoiceEventToolCallCompleted) {
				t.Fatalf("unsupported request must not execute inventory reads, got %+v", events)
			}
		})
	}
}

func TestRealtimeVoiceInvestigationLoopCompletesRequiredContentsEvidenceAfterTargetResolution(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationListContents, SubjectMention: "Office"}
	initial := agentmodel.InvestigationStep{
		Decision: agentmodel.InvestigationDecisionSearch, Intent: intent,
		SearchRequests: []agentmodel.SearchRequest{{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Office", SearchProbes: []string{"office"}}},
	}
	prematureFinish := agentmodel.InvestigationStep{
		Decision: agentmodel.InvestigationDecisionFinish, Intent: intent,
		Resolutions: []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{"office-1"}}},
	}
	final := agentmodel.InvestigationStep{
		Decision: agentmodel.InvestigationDecisionFinish, Intent: intent,
		Resolutions: []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionCollection, CandidateIDs: []string{"bottle-1"}}},
	}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{Investigation: &initial}, {Investigation: &prematureFinish}, {Investigation: &final}}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.LanguageInference = language
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "What is in the office?"}
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	office := realtimeVoiceInvestigationAsset("office-1", "Office", asset.KindLocation, "")
	bottle := realtimeVoiceInvestigationAsset("bottle-1", "Water bottle", asset.KindItem, office.ID.String())
	seedRealtimeVoiceLoopAsset(t, store, office, "audit-office")
	seedRealtimeVoiceLoopAsset(t, store, bottle, "audit-bottle")
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
		t.Fatalf("run contents investigation: %v", err)
	}
	response := realtimeVoiceInvestigationCompletedResponse(events)
	if response == nil || response.SpokenResponse != "I found 1 visible matches: Water bottle." {
		t.Fatalf("expected contents grounded after required read, got %+v", events)
	}
	if len(language.seenInvestigations) != 3 {
		t.Fatalf("expected initial, target assessment, and contents assessment turns, got %d", len(language.seenInvestigations))
	}
	last := language.seenInvestigations[2]
	foundRequiredRead := false
	for _, evidence := range last.ReadEvidence {
		if evidence.ReferenceKey == agentmodel.SemanticReferenceSubject && evidence.ReadKind == agentmodel.InvestigationReadListContents && evidence.VisibleAssetID == office.ID.String() {
			foundRequiredRead = true
		}
	}
	if !foundRequiredRead {
		t.Fatalf("expected application-scheduled contents evidence, got %+v", last.ReadEvidence)
	}
}

func TestRealtimeVoiceCurrentCheckoutStatusRequiresCheckoutEvidence(t *testing.T) {
	t.Parallel()
	readKind, required := realtimeVoiceOperationRequiredRead(agentmodel.OperationCheckoutStatus)
	if !required || readKind != agentmodel.InvestigationReadCheckoutHistory {
		t.Fatalf("checkout status must require checkout history evidence, got %q required=%t", readKind, required)
	}
}

func TestRealtimeVoiceInvestigationLoopCompilesNestedMissingDestinationAndStopsAtReview(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{
		RequestShape: agentmodel.RequestShapeSingleTarget,
		Kind:         agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "Drill",
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

func TestRealtimeVoiceInvestigationLoopFinishesExactOrZeroPathWithoutRedundantSearchRound(t *testing.T) {
	t.Parallel()
	intent := agentmodel.Intent{
		RequestShape: agentmodel.RequestShapeSingleTarget,
		Kind:         agentmodel.IntentKindChange, Operation: agentmodel.OperationMove, SubjectMention: "my Drill",
		DestinationPath:  []string{"Kitchen", "Big cabinet", "Second shelf"},
		DestinationKinds: []agentmodel.DestinationKind{agentmodel.DestinationKindLocation, agentmodel.DestinationKindContainer, agentmodel.DestinationKindContainer},
	}
	requests := []agentmodel.SearchRequest{
		{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "my Drill", SearchProbes: []string{"drill"}},
		{ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Kitchen", SearchProbes: []string{"kitchen"}},
		{ReferenceKey: "destination.1", ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Big cabinet", SearchProbes: []string{"big cabinet"}},
		{ReferenceKey: "destination.2", ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Second shelf", SearchProbes: []string{"second shelf"}},
	}
	initial := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionSearch, Intent: intent, SearchRequests: requests}
	redundant := agentmodel.InvestigationStep{Decision: agentmodel.InvestigationDecisionSearchAgain, Intent: intent, SearchRequests: []agentmodel.SearchRequest{
		{ReferenceKey: "destination.0", ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Kitchen", SearchProbes: []string{"kitchen room"}},
		{ReferenceKey: "destination.1", ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Big cabinet", SearchProbes: []string{"large cabinet"}},
		{ReferenceKey: "destination.2", ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Second shelf", SearchProbes: []string{"shelf two"}},
	}}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{{Investigation: &initial}, {Investigation: &redundant}}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.LanguageInference = language
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move the drill into the second shelf in the big cabinet in the kitchen"}
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	seedRealtimeVoiceLoopAsset(t, store, realtimeVoiceInvestigationAsset("drill-1", "Drill", asset.KindItem, ""), "audit-drill-exact-zero")
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
		t.Fatalf("run exact-or-zero investigation: %v", err)
	}
	proposal := realtimeVoiceInvestigationProposedPlan(events)
	if proposal == nil || len(proposal.Commands) != 4 || proposal.Commands[3].ParentCommandID != proposal.Commands[2].ID {
		t.Fatalf("expected nested create/create/create/move plan, got %+v", events)
	}
	if language.callCount != 2 {
		t.Fatalf("expected no redundant third provider call, got %d", language.callCount)
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
