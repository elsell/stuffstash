package agentmodel

import (
	"context"
	"errors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

type workflowExecutionClock struct{ now time.Time }

func (c *workflowExecutionClock) Now() time.Time { return c.now }

type workflowExecutionProvider struct {
	calls    int
	failures int
	prompts  []string
}

func (p *workflowExecutionProvider) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	p.calls++
	p.prompts = append(p.prompts, input.PromptTemplate)
	if p.calls <= p.failures {
		return ports.LanguageInferenceTurn{}, errors.New("provider unavailable")
	}
	return ports.LanguageInferenceTurn{}, nil
}
func (p *workflowExecutionProvider) GenerateResponse(context.Context, ports.VoiceResponseGenerationInput) (ports.VoiceResponseGenerationResult, error) {
	p.calls++
	return ports.VoiceResponseGenerationResult{SpokenResponse: "Found clothes.", DisplayResponse: "Found clothes."}, nil
}
func workflowExecutionFixture(t *testing.T, calls int) (*WorkflowModelExecution, *workflowExecutionProvider, *workflowExecutionClock) {
	t.Helper()
	clock := &workflowExecutionClock{now: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)}
	provider := &workflowExecutionProvider{failures: 1}
	limits := domain.WorkflowLimits{Budget: domain.WorkflowBudget{ModelCalls: calls, EvidenceRounds: 2, ElapsedSeconds: 30, FollowUpTurns: 2}, MaxStepAttempts: 2, MaxNameRunes: 100, MaxInstructionRunes: 1000}
	definition, err := domain.NewWorkflowDefinition(domain.WorkflowDefinitionInput{Name: "Test", Retrieval: domain.WorkflowRetrievalExpanded, Response: domain.WorkflowResponseGroundedFallback, Budget: limits.Budget, Steps: []domain.WorkflowStep{{Kind: domain.WorkflowStepInterpret, Attempts: 2, Instructions: "Use tags."}, {Kind: domain.WorkflowStepAssess, Attempts: 2}, {Kind: domain.WorkflowStepRespond, Attempts: 1}}}, limits)
	if err != nil {
		t.Fatal(err)
	}
	binding := WorkflowModelBinding{Language: provider, Response: provider}
	execution, err := NewWorkflowModelExecution(definition, limits, clock, map[domain.WorkflowStepKind]WorkflowModelBinding{domain.WorkflowStepInterpret: binding, domain.WorkflowStepAssess: binding, domain.WorkflowStepRespond: binding})
	if err != nil {
		t.Fatal(err)
	}
	return execution, provider, clock
}
func TestWorkflowModelAttemptsConsumeSharedBudget(t *testing.T) {
	execution, provider, _ := workflowExecutionFixture(t, 2)
	input := ports.LanguageInferenceInput{Investigation: &domain.InvestigationInput{Phase: domain.InvestigationPhaseInitial}, PromptTemplate: "Household guidance."}
	if _, err := execution.NextTurn(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if provider.calls != 2 || provider.prompts[0] != "Household guidance.\nUse tags." {
		t.Fatalf("step policy not applied: %+v", provider)
	}
	if _, err := execution.NextTurn(context.Background(), input); !errors.Is(err, ErrWorkflowBudgetExhausted) {
		t.Fatalf("budget not enforced: %v", err)
	}
	if provider.calls != 2 {
		t.Fatal("exhausted call reached provider")
	}
}
func TestWorkflowModelHonorsCancellationAndElapsedBudget(t *testing.T) {
	execution, provider, clock := workflowExecutionFixture(t, 4)
	input := ports.LanguageInferenceInput{Investigation: &domain.InvestigationInput{Phase: domain.InvestigationPhaseInitial}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := execution.NextTurn(ctx, input); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
	clock.now = clock.now.Add(31 * time.Second)
	if _, err := execution.NextTurn(context.Background(), input); !errors.Is(err, ErrWorkflowBudgetExhausted) {
		t.Fatalf("elapsed budget: %v", err)
	}
	if provider.calls != 0 {
		t.Fatal("expired execution called provider")
	}
}
