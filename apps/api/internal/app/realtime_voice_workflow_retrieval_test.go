package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	appmodel "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestSelectedWorkflowControlsEvidenceRoundsAndRetrieval(t *testing.T) {
	for _, rounds := range []int{1, 3} {
		t.Run(fmt.Sprint(rounds), func(t *testing.T) {
			language := &scriptedRealtimeLanguageInference{}
			intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "baby clothes"}
			for round := 0; round <= rounds; round++ {
				decision := agentmodel.InvestigationDecisionSearchAgain
				if round == 0 {
					decision = agentmodel.InvestigationDecisionSearch
				}
				language.turns = append(language.turns, ports.LanguageInferenceTurn{Investigation: &agentmodel.InvestigationStep{Decision: decision, Intent: intent, SearchRequests: []agentmodel.SearchRequest{{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "baby clothes", SearchProbes: []string{fmt.Sprintf("baby clothing %d", round)}}}}})
			}
			resolver := &fakeRealtimeVoiceProviderResolver{providers: ports.RealtimeVoiceProviderSet{SpeechToText: resolvedSpeechToText{transcript: "Where are my baby clothes?"}, LanguageInference: language, ResponseGenerator: &resolvedLanguageInference{}, TextToSpeech: &resolvedTextToSpeech{}}}
			application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
			recorder := &workflowSearchRecorder{AssetSearchRepository: store}
			application.search = recorder
			limits := agentmodel.WorkflowLimits{Budget: agentmodel.WorkflowBudget{EvidenceRounds: rounds, ModelCalls: rounds + 2, ElapsedSeconds: 60, FollowUpTurns: 4}, MaxStepAttempts: 2, MaxNameRunes: 100, MaxInstructionRunes: 1000}
			application.conversationWorkflowService = appmodel.NewConversationWorkflowService(appmodel.ConversationWorkflowDependencies{Authorizer: application.authorizer, Repository: store, Profiles: store, IDs: application.ids, Clock: application.clock, Limits: limits})
			seedSessionWorkflow(t, application, store, limits)
			session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
			if err != nil {
				t.Fatal(err)
			}
			err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil })
			if err != nil {
				t.Fatal(err)
			}
			if language.callCount != rounds+1 {
				t.Fatalf("configured %d rounds, got %d model calls", rounds, language.callCount)
			}
			for _, input := range language.seenInvestigations {
				if input.MaxEvidenceRounds != rounds {
					t.Fatalf("wrong advertised budget: %d", input.MaxEvidenceRounds)
				}
			}
			if len(recorder.modes) != rounds*2 {
				t.Fatalf("expected exact/fuzzy per zero-result round: %v", recorder.modes)
			}
			for index, mode := range recorder.modes {
				expected := search.ModeExact
				if index%2 == 1 {
					expected = search.ModeFuzzy
				}
				if mode != expected {
					t.Fatalf("wrong strategy: %v", recorder.modes)
				}
			}
		})
	}
}

type workflowSearchRecorder struct {
	ports.AssetSearchRepository
	failure error
	modes   []search.Mode
}

func (r *workflowSearchRecorder) SearchAssets(ctx context.Context, t tenant.ID, ids []inventory.InventoryID, input ports.AssetSearchPageRequest) ([]ports.AssetSearchResult, error) {
	if t != "tenant-home" || len(ids) != 1 || ids[0] != "inventory-home" {
		return nil, ports.ErrForbidden
	}
	r.modes = append(r.modes, input.Mode)
	if r.failure != nil {
		return nil, r.failure
	}
	return r.AssetSearchRepository.SearchAssets(ctx, t, ids, input)
}

func TestWorkflowRetrievalAvoidsUnneededExpansion(t *testing.T) {
	for _, scenario := range []struct {
		name     string
		strategy agentmodel.WorkflowRetrievalStrategy
		match    bool
		fail     bool
		mode     search.Mode
	}{
		{"exact hit", agentmodel.WorkflowRetrievalPreciseFirst, true, false, search.ModeExact},
		{"expanded", agentmodel.WorkflowRetrievalExpanded, false, false, search.ModeFuzzy},
		{"failed exact", agentmodel.WorkflowRetrievalPreciseFirst, false, true, search.ModeExact},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			resolver := &fakeRealtimeVoiceProviderResolver{providers: ports.RealtimeVoiceProviderSet{SpeechToText: resolvedSpeechToText{}, LanguageInference: &resolvedLanguageInference{}, ResponseGenerator: &resolvedLanguageInference{}, TextToSpeech: &resolvedTextToSpeech{}}}
			application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
			recorder := &workflowSearchRecorder{AssetSearchRepository: store}
			expectedFailure := errors.New("search unavailable")
			if scenario.fail {
				recorder.failure = expectedFailure
			}
			application.search = recorder
			limits := agentmodel.WorkflowLimits{Budget: agentmodel.WorkflowBudget{EvidenceRounds: 2, ModelCalls: 4, ElapsedSeconds: 60, FollowUpTurns: 4}, MaxStepAttempts: 2, MaxNameRunes: 100, MaxInstructionRunes: 1000}
			application.conversationWorkflowService = appmodel.NewConversationWorkflowService(appmodel.ConversationWorkflowDependencies{Authorizer: application.authorizer, Repository: store, Profiles: store, IDs: application.ids, Clock: application.clock, Limits: limits})
			revision := seedSessionWorkflow(t, application, store, limits)
			if scenario.strategy != agentmodel.WorkflowRetrievalPreciseFirst {
				settings := revision.Snapshot().Definition.Settings()
				settings.Retrieval = scenario.strategy
				updated, err := application.SaveConversationWorkflowRevision(context.Background(), SaveConversationWorkflowInput{Principal: defaultRealtimeVoiceSessionInput().Principal, TenantID: "tenant-home", Source: audit.SourceAPI, WorkflowID: revision.Snapshot().WorkflowID, ExpectedRevision: 1, Definition: settings})
				if err != nil {
					t.Fatal(err)
				}
				activateSessionWorkflow(t, application, store, updated, ports.WorkflowSelectionReference{WorkflowID: revision.Snapshot().WorkflowID, RevisionID: revision.Snapshot().ID})
			}
			if scenario.match {
				seedRealtimeVoiceLoopAsset(t, store, asset.Asset{ID: "clothes", TenantID: "tenant-home", InventoryID: "inventory-home", Kind: asset.KindItem, Title: "Baby clothes", LifecycleState: asset.LifecycleStateActive}, "seed-clothes")
			}
			session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
			if err != nil {
				t.Fatal(err)
			}
			_, err = application.executeRealtimeVoiceSearchTool(context.Background(), session, ports.AgentToolCall{ID: "search", Name: RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "Baby clothes", "limit": 10}})
			if scenario.fail {
				if !errors.Is(err, expectedFailure) {
					t.Fatalf("search error changed: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if len(recorder.modes) != 1 || recorder.modes[0] != scenario.mode {
				t.Fatalf("unnecessary or wrong search: %v", recorder.modes)
			}
		})
	}
}

func TestWorkflowCandidateCapacityAsksForNarrowerSearch(t *testing.T) {
	language := &scriptedRealtimeLanguageInference{}
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindRead, Operation: agentmodel.OperationLocate, SubjectMention: "baby clothes"}
	for round := 0; round < 3; round++ {
		decision := agentmodel.InvestigationDecisionSearchAgain
		if round == 0 {
			decision = agentmodel.InvestigationDecisionSearch
		}
		probes := []string{}
		for probe := 0; probe < 3; probe++ {
			probes = append(probes, fmt.Sprintf("batch%dprobe%d", round, probe))
		}
		language.turns = append(language.turns, ports.LanguageInferenceTurn{Investigation: &agentmodel.InvestigationStep{Decision: decision, Intent: intent, SearchRequests: []agentmodel.SearchRequest{{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "baby clothes", SearchProbes: probes}}}})
	}
	resolver := &fakeRealtimeVoiceProviderResolver{providers: ports.RealtimeVoiceProviderSet{SpeechToText: resolvedSpeechToText{transcript: "Where are my baby clothes?"}, LanguageInference: language, ResponseGenerator: &resolvedLanguageInference{}, TextToSpeech: &resolvedTextToSpeech{}}}
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	recorder := &workflowSearchRecorder{AssetSearchRepository: store}
	application.search = recorder
	limits := agentmodel.WorkflowLimits{Budget: agentmodel.WorkflowBudget{EvidenceRounds: 3, ModelCalls: 5, ElapsedSeconds: 60, FollowUpTurns: 4}, MaxStepAttempts: 2, MaxNameRunes: 100, MaxInstructionRunes: 1000}
	application.conversationWorkflowService = appmodel.NewConversationWorkflowService(appmodel.ConversationWorkflowDependencies{Authorizer: application.authorizer, Repository: store, Profiles: store, IDs: application.ids, Clock: application.clock, Limits: limits})
	seedSessionWorkflow(t, application, store, limits)
	for round := 0; round < 3; round++ {
		for probe := 0; probe < 3; probe++ {
			for index := 0; index < 20; index++ {
				id := fmt.Sprintf("candidate-%d-%d-%d", round, probe, index)
				seedRealtimeVoiceLoopAsset(t, store, asset.Asset{ID: asset.ID(id), TenantID: "tenant-home", InventoryID: "inventory-home", Kind: asset.KindItem, Title: asset.Title(fmt.Sprintf("batch%dprobe%d clothes %d", round, probe, index)), LifecycleState: asset.LifecycleStateActive}, "seed-"+id)
			}
		}
	}
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatal(err)
	}
	var response *ports.StructuredAgentResponse
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}, ContinueAfterClarification: true}, func(event RealtimeVoiceEvent) error {
		if event.Type == RealtimeVoiceEventAssistantResponseCompleted {
			response = event.Response
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if response == nil || response.Kind != ports.StructuredAgentResponseKindClarification {
		t.Fatalf("too many candidates failed instead of clarifying: %+v", response)
	}
	if language.callCount != 3 || len(recorder.modes) != 7 {
		t.Fatalf("exhausted context dispatched further work: models=%d reads=%d", language.callCount, len(recorder.modes))
	}
	for _, input := range language.seenInvestigations {
		if len(input.Observations) > agentmodel.MaxCandidateObservations {
			t.Fatal("invalid provider context dispatched")
		}
	}
}
