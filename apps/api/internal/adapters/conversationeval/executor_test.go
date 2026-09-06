package conversationeval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
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
	entered  chan struct{}
}

func (model *taggedClothesModel) Converse(ctx context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	model.calls++
	if model.cancel {
		if model.entered != nil {
			close(model.entered)
		}
		<-ctx.Done()
		return ports.ConversationModelTurn{}, ctx.Err()
	}
	last := input.Messages[len(input.Messages)-1]
	if last.Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "search-clothes", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "baby clothes"}}}}, nil
	}
	var result struct {
		Items []struct {
			AssetID     string   `json:"assetId"`
			TagNames    []string `json:"tagNames"`
			ParentTitle string   `json:"parentTitle"`
		} `json:"items"`
	}
	if len(last.ToolResults) != 1 || json.Unmarshal([]byte(last.ToolResults[0].Content), &result) != nil {
		return ports.ConversationModelTurn{}, ports.ErrInvalidProviderInput
	}
	ids := []string{}
	location := ""
	for _, item := range result.Items {
		if slices.Contains(item.TagNames, "baby") && slices.Contains(item.TagNames, "clothes") {
			model.seenTags = true
			ids = append(ids, item.AssetID)
			location = item.ParentTitle
		}
	}
	return ports.ConversationModelTurn{Answer: &ports.ConversationAnswer{Spoken: "They are in " + location + ".", Display: "Matching clothes are shown below.", AssetIDs: ids}}, nil
}

// Compatibility methods are never invoked by the evaluation path.
func (*taggedClothesModel) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
}
func (*taggedClothesModel) GenerateResponse(context.Context, ports.VoiceResponseGenerationInput) (ports.VoiceResponseGenerationResult, error) {
	return ports.VoiceResponseGenerationResult{}, ports.ErrInvalidProviderInput
}
func testInput(t *testing.T, model ports.ConversationModel) ports.ConversationEvaluationInput {
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
	return ports.ConversationEvaluationInput{Case: caseDefinition, Revision: revision, Limits: limits, Principal: identity.Principal{ID: "owner"}, Providers: ports.RealtimeVoiceProviderSet{LanguageInferenceProfileID: "configured", ConversationModel: model}}
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
	model := &taggedClothesModel{cancel: true, entered: make(chan struct{})}
	executor := New(Dependencies{Clock: testClock{}, IDs: &testIDs{}})
	input := testInput(t, model)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finished := make(chan error, 1)
	go func() { _, err := executor.Execute(ctx, input); finished <- err }()
	select {
	case <-model.entered:
		cancel()
	case err := <-finished:
		t.Fatalf("execution stopped before provider: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("provider never started")
	}
	select {
	case err := <-finished:
		if err != context.Canceled {
			t.Fatalf("cancellation not propagated: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("cancelled provider did not stop")
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

type additionalClothesModel struct{ taggedClothesModel }

func (*additionalClothesModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	if input.Messages[len(input.Messages)-1].Role == ports.ConversationRoleUser {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "find-existing", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "3 to 6 months"}}}}, nil
	}
	return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "proposal", Name: "propose_inventory_change", Arguments: map[string]any{"summary": "Record the additional clothes.", "commands": []any{map[string]any{"id": "create-clothes", "kind": "create_asset", "summary": "Record additional clothes.", "arguments": map[string]any{"kind": "item", "title": "3 to 6 months"}}}}}}}, nil
}
func TestTextEvaluationCapturesUnapprovedAdditionalCreation(t *testing.T) {
	input := testInput(t, &additionalClothesModel{})
	settings := input.Case.Settings()
	settings.Utterance = "I bought another pack of clothes"
	settings.Expectations = domain.EvaluationExpectations{Kind: domain.EvaluationOutcomeProposal, Proposals: []domain.EvaluationProposal{{Operation: domain.OperationCreate, NewTitle: "3 to 6 months", NewKind: domain.EvaluationFixtureItem}}}
	var err error
	input.Case, err = domain.NewEvaluationCaseDefinition(settings)
	if err != nil {
		t.Fatal(err)
	}
	result, err := New(Dependencies{Clock: testClock{}, IDs: &testIDs{}}).Execute(context.Background(), input)
	if err != nil || !input.Case.Evaluate(result.Outcome).Passed {
		t.Fatalf("proposal not captured: %+v %v", result, err)
	}
	if len(result.Outcome.ExecutedOperations) > 0 {
		t.Fatal("evaluation approved proposal")
	}
}
func TestPinnedEvaluationRevisionRejectsOtherScopesAndWrites(t *testing.T) {
	input := testInput(t, &taggedClothesModel{})
	repository := pinnedWorkflow{revision: input.Revision}
	ctx := context.Background()
	if _, found, err := repository.SelectedWorkflowRevision(ctx, "other-tenant"); err != nil || found {
		t.Fatal("selection leaked across tenant")
	}
	if _, found, err := repository.WorkflowRevision(ctx, "tenant-home", "other-workflow", "revision-seven"); err != nil || found {
		t.Fatal("revision leaked across workflow")
	}
	if _, found, err := repository.WorkflowRevision(ctx, "tenant-home", "workflow-one", "revision-other"); err != nil || found {
		t.Fatal("wrong revision returned")
	}
	if _, found, err := repository.WorkflowRevision(ctx, "tenant-home", "workflow-one", "revision-seven"); err != nil || !found {
		t.Fatal("candidate revision inaccessible")
	}
}

func TestPinnedEvaluationRevisionCannotActivateOrAppend(t *testing.T) {
	input := testInput(t, &taggedClothesModel{})
	repository := pinnedWorkflow{revision: input.Revision}
	if err := repository.AppendWorkflowRevision(context.Background(), input.Revision, 6, audit.Record{}); err == nil {
		t.Fatal("evaluation repository accepted revision write")
	}
	if err := repository.ActivateWorkflowRevision(context.Background(), "tenant-home", "workflow-one", "revision-seven", ports.WorkflowSelectionReference{}, testClock{}.Now(), audit.Record{}); err == nil {
		t.Fatal("evaluation repository accepted activation")
	}
}

func TestFixtureGuardDetectsAssetAndCheckoutMutation(t *testing.T) {
	for _, checkout := range []bool{false, true} {
		t.Run(fmt.Sprint(checkout), func(t *testing.T) {
			input := testInput(t, &taggedClothesModel{})
			executor := New(Dependencies{Clock: testClock{}, IDs: &testIDs{}})
			runtime, err := executor.prepare(context.Background(), input, &atomic.Int64{})
			if err != nil {
				t.Fatal(err)
			}
			before, err := runtime.snapshot(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			var target asset.ID
			for id, fixture := range runtime.runtimeIDs {
				if fixture == "clothes" {
					target = asset.ID(id)
				}
			}
			if checkout {
				_, err = runtime.application.CheckoutAsset(context.Background(), app.CheckoutAssetInput{Principal: input.Principal, TenantID: runtime.tenantID, InventoryID: runtime.inventoryID, AssetID: target, Source: audit.SourceSystem, Details: "Taken"})
			} else {
				_, err = runtime.application.ArchiveAssetWithOperation(context.Background(), app.UpdateAssetLifecycleInput{Principal: input.Principal, TenantID: runtime.tenantID, InventoryID: runtime.inventoryID, AssetID: target, Source: audit.SourceSystem})
			}
			if err != nil {
				t.Fatal(err)
			}
			if err = runtime.assertUnchanged(context.Background(), before); !errors.Is(err, ErrFixtureMutation) {
				t.Fatalf("mutation undetected: %v", err)
			}
		})
	}
}

type failingModel struct {
	taggedClothesModel
	failure error
}

func (model failingModel) Converse(context.Context, ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	return ports.ConversationModelTurn{}, model.failure
}
func TestEvaluationDistinguishesProviderFailureFromInvalidEvidence(t *testing.T) {
	for _, scenario := range []struct {
		name     string
		failure  error
		observed bool
	}{
		{"provider unavailable", errors.New("provider unavailable"), true},
		{"adapter invalid output", ports.ErrInvalidProviderInput, false},
		{"empty result", nil, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			input := testInput(t, &failingModel{failure: scenario.failure})
			settings := input.Case.Settings()
			settings.Expectations = domain.EvaluationExpectations{Kind: domain.EvaluationOutcomeFailure}
			var err error
			input.Case, err = domain.NewEvaluationCaseDefinition(settings)
			if err != nil {
				t.Fatal(err)
			}
			result, err := New(Dependencies{Clock: testClock{}, IDs: &testIDs{}}).Execute(context.Background(), input)
			if scenario.observed {
				if err != nil || !input.Case.Evaluate(result.Outcome).Passed || result.ModelCalls != 1 {
					t.Fatalf("provider failure lost: %+v %v", result, err)
				}
			} else if err == nil {
				t.Fatalf("invalid evidence passed failure case: %+v", result)
			}
		})
	}
}
