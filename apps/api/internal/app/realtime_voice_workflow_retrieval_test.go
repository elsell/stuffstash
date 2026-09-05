package app

import (
	"context"
	"fmt"
	"testing"

	appmodel "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
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
	modes []search.Mode
}

func (r *workflowSearchRecorder) SearchAssets(ctx context.Context, t tenant.ID, ids []inventory.InventoryID, input ports.AssetSearchPageRequest) ([]ports.AssetSearchResult, error) {
	if t != "tenant-home" || len(ids) != 1 || ids[0] != "inventory-home" {
		return nil, ports.ErrForbidden
	}
	r.modes = append(r.modes, input.Mode)
	return r.AssetSearchRepository.SearchAssets(ctx, t, ids, input)
}
