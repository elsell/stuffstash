package conversationeval

import (
	"context"
	"fmt"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	modelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type testIDs struct{ next atomic.Int64 }

func (ids *testIDs) NewID() string { return fmt.Sprintf("generated-%d", ids.next.Add(1)) }

type testClock struct{}

func (testClock) Now() time.Time { return time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC) }

type taggedClothesModel struct {
	seenTags bool
	calls    int
	cancel   bool
}

func (model *taggedClothesModel) NextTurn(ctx context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	model.calls++
	if model.cancel {
		<-ctx.Done()
		return ports.LanguageInferenceTurn{}, ctx.Err()
	}
	intent := domain.Intent{RequestShape: domain.RequestShapeSingleTarget, Kind: domain.IntentKindRead, Operation: domain.OperationLocate, SubjectMention: "baby clothes"}
	step := domain.InvestigationStep{Intent: intent, Decision: domain.InvestigationDecisionSearch, SearchRequests: []domain.SearchRequest{{ReferenceKey: domain.SemanticReferenceSubject, ReadKind: domain.InvestigationReadSearchAssets, Mention: "baby clothes", SearchProbes: []string{"baby"}}}}
	if input.Investigation.Phase != domain.InvestigationPhaseInitial {
		ids := []string{}
		for _, candidate := range input.Investigation.Observations {
			if slices.Contains(candidate.TagNames, "baby") && slices.Contains(candidate.TagNames, "clothes") {
				model.seenTags = true
				ids = append(ids, candidate.CandidateID)
			}
		}
		step.Decision = domain.InvestigationDecisionFinish
		step.SearchRequests = nil
		step.Resolutions = []domain.Resolution{{ReferenceKey: domain.SemanticReferenceSubject, Status: domain.ResolutionCollection, CandidateIDs: ids}}
	}
	return ports.LanguageInferenceTurn{Investigation: &step}, nil
}
func (*taggedClothesModel) GenerateResponse(_ context.Context, input ports.VoiceResponseGenerationInput) (ports.VoiceResponseGenerationResult, error) {
	spoken, err := modelapp.RenderGroundedVoiceResponse(input.Brief, 2000)
	return ports.VoiceResponseGenerationResult{SpokenResponse: spoken, DisplayResponse: spoken}, err
}
func testInput(t *testing.T, model ports.RealtimeLanguageProvider) ports.ConversationEvaluationInput {
	t.Helper()
	limits := domain.WorkflowLimits{Budget: domain.WorkflowBudget{EvidenceRounds: 2, ModelCalls: 4, ElapsedSeconds: 60, FollowUpTurns: 2}, MaxStepAttempts: 1, MaxNameRunes: 100, MaxInstructionRunes: 1000}
	definition, err := domain.NewWorkflowDefinition(domain.WorkflowDefinitionInput{Name: "Test", Retrieval: domain.WorkflowRetrievalExpanded, Response: domain.WorkflowResponseGrounded, Budget: limits.Budget, Steps: []domain.WorkflowStep{{Kind: domain.WorkflowStepInterpret, Attempts: 1}, {Kind: domain.WorkflowStepAssess, Attempts: 1}, {Kind: domain.WorkflowStepRespond, Attempts: 1}}}, limits)
	if err != nil {
		t.Fatal(err)
	}
	revision, err := domain.NewWorkflowRevision(domain.WorkflowRevisionInput{ID: "revision-seven", WorkflowID: "workflow-one", TenantID: "tenant-home", AuthorID: "owner", Number: 7, Definition: definition, Limits: limits, CreatedAt: testClock{}.Now()})
	if err != nil {
		t.Fatal(err)
	}
	caseDefinition, err := domain.NewEvaluationCaseDefinition(domain.EvaluationCaseDefinitionInput{Title: "Baby clothes", Utterance: "Where are my baby clothes?", Assets: []domain.EvaluationFixtureAsset{{ID: "clothes", Title: "3 to 6 months", Kind: domain.EvaluationFixtureItem, ParentID: "box", TagNames: []string{"baby", "clothes"}}, {ID: "box", Title: "Attic box", Kind: domain.EvaluationFixtureContainer}}, Expectations: domain.EvaluationExpectations{Kind: domain.EvaluationOutcomeAnswer, ReferencedAssets: []string{"clothes"}, Locations: []domain.EvaluationLocationExpectation{{AssetID: "clothes", AncestorID: "box"}}}})
	if err != nil {
		t.Fatal(err)
	}
	return ports.ConversationEvaluationInput{Case: caseDefinition, Revision: revision, Limits: limits, Principal: identity.Principal{ID: "owner"}, Providers: ports.RealtimeVoiceProviderSet{LanguageInferenceProfileID: "configured", LanguageInference: model, ResponseGenerator: model}}
}
func TestTextEvaluationUsesProductionLoopWithIsolatedTaggedFixtures(t *testing.T) {
	model := &taggedClothesModel{}
	executor := New(Dependencies{Clock: testClock{}, IDs: &testIDs{}})
	input := testInput(t, model)
	for range 2 {
		result, err := executor.Execute(context.Background(), input)
		if err != nil {
			t.Fatal(err)
		}
		if !input.Case.Evaluate(result.Outcome).Passed || !model.seenTags || result.ModelCalls != 2 || result.Coverage != ports.EvaluationCoverageText {
			t.Fatalf("real loop not evaluated: %+v", result)
		}
		if len(result.Outcome.ExecutedOperations) != 0 {
			t.Fatal("evaluation approved a write")
		}
	}
}
func TestTextEvaluationPropagatesCancellation(t *testing.T) {
	model := &taggedClothesModel{cancel: true}
	executor := New(Dependencies{Clock: testClock{}, IDs: &testIDs{}})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := executor.Execute(ctx, testInput(t, model)); err == nil {
		t.Fatal("cancelled execution reported an outcome")
	}
}
func TestInvalidEvaluationDoesNotCallProvider(t *testing.T) {
	model := &taggedClothesModel{}
	executor := New(Dependencies{Clock: testClock{}, IDs: &testIDs{}})
	input := testInput(t, model)
	input.Revision = domain.WorkflowRevision{}
	if _, err := executor.Execute(context.Background(), input); err == nil || model.calls != 0 {
		t.Fatal("invalid revision reached provider")
	}
}
