package agentmodel

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type workflowExecutionClock struct{ now time.Time }

func (c *workflowExecutionClock) Now() time.Time { return c.now }

type workflowExecutionProvider struct {
	calls        int
	failures     int
	prompts      []string
	instructions []string
}

func (p *workflowExecutionProvider) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	p.calls++
	p.prompts = append(p.prompts, input.PromptTemplate)
	p.instructions = append(p.instructions, input.WorkflowInstructions)
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
	if provider.calls != 2 || provider.prompts[0] != "Household guidance." || provider.instructions[0] != "Use tags." {
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

type lateWorkflowProvider struct{}

func (lateWorkflowProvider) NextTurn(ctx context.Context, _ ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	<-ctx.Done()
	return ports.LanguageInferenceTurn{}, nil
}
func (lateWorkflowProvider) GenerateResponse(ctx context.Context, _ ports.VoiceResponseGenerationInput) (ports.VoiceResponseGenerationResult, error) {
	<-ctx.Done()
	return ports.VoiceResponseGenerationResult{SpokenResponse: "Late success", DisplayResponse: "Late success"}, nil
}
func TestWorkflowRejectsLateSuccessInBothModelStages(t *testing.T) {
	for _, kind := range []domain.WorkflowStepKind{domain.WorkflowStepInterpret, domain.WorkflowStepRespond} {
		t.Run(string(kind), func(t *testing.T) {
			execution, _, clock := workflowExecutionFixture(t, 4)
			clock.now = clock.now.Add(30*time.Second - time.Millisecond)
			execution.providers[kind] = WorkflowModelBinding{Language: lateWorkflowProvider{}, Response: lateWorkflowProvider{}}
			var err error
			if kind == domain.WorkflowStepInterpret {
				_, err = execution.NextTurn(context.Background(), ports.LanguageInferenceInput{Investigation: &domain.InvestigationInput{Phase: domain.InvestigationPhaseInitial}})
			} else {
				_, err = execution.GenerateResponse(context.Background(), ports.VoiceResponseGenerationInput{Brief: domain.GroundedVoiceResponseBrief{Kind: domain.ResponseBriefKindAnswer, Mode: domain.ResponseAnswerModeNotFound, Operation: domain.OperationLocate, Subject: "clothes", Confidence: domain.ResponseConfidenceAbsent}})
			}
			if !errors.Is(err, ErrWorkflowBudgetExhausted) || execution.ModelCalls() != 1 {
				t.Fatalf("late response escaped deadline: calls=%d err=%v", execution.ModelCalls(), err)
			}
		})
	}
}

func TestWorkflowSharesBudgetBetweenInferenceAndWording(t *testing.T) {
	execution, provider, _ := workflowExecutionFixture(t, 2)
	provider.failures = 0
	if _, err := execution.NextTurn(context.Background(), ports.LanguageInferenceInput{Investigation: &domain.InvestigationInput{Phase: domain.InvestigationPhaseInitial}}); err != nil {
		t.Fatal(err)
	}
	brief := domain.GroundedVoiceResponseBrief{Kind: domain.ResponseBriefKindAnswer, Mode: domain.ResponseAnswerModeNotFound, Operation: domain.OperationLocate, Subject: "clothes", Confidence: domain.ResponseConfidenceAbsent}
	if _, err := execution.GenerateResponse(context.Background(), ports.VoiceResponseGenerationInput{Brief: brief}); err != nil {
		t.Fatal(err)
	}
	if _, err := execution.NextTurn(context.Background(), ports.LanguageInferenceInput{Investigation: &domain.InvestigationInput{Phase: domain.InvestigationPhaseEvidenceAssessment}}); !errors.Is(err, ErrWorkflowBudgetExhausted) {
		t.Fatalf("assessment exceeded shared budget: %v", err)
	}
	if provider.calls != 2 || execution.ModelCalls() != 2 {
		t.Fatal("model stages did not share their call budget")
	}
}

func TestWorkflowGroundedResponseNeedsNoWordingProvider(t *testing.T) {
	original, _, clock := workflowExecutionFixture(t, 2)
	settings := original.definition.Settings()
	settings.Response = domain.WorkflowResponseGrounded
	limits := domain.WorkflowLimits{Budget: settings.Budget, MaxStepAttempts: 2, MaxNameRunes: 100, MaxInstructionRunes: 1000}
	definition, err := domain.NewWorkflowDefinition(settings, limits)
	if err != nil {
		t.Fatal(err)
	}
	bindings := map[domain.WorkflowStepKind]WorkflowModelBinding{
		domain.WorkflowStepInterpret: original.providers[domain.WorkflowStepInterpret],
		domain.WorkflowStepAssess:    original.providers[domain.WorkflowStepAssess],
	}
	execution, err := NewWorkflowModelExecution(definition, limits, clock, bindings)
	if err != nil {
		t.Fatal(err)
	}
	result, err := execution.GenerateResponse(context.Background(), ports.VoiceResponseGenerationInput{Brief: domain.GroundedVoiceResponseBrief{Kind: domain.ResponseBriefKindAnswer, Mode: domain.ResponseAnswerModeNotFound, Operation: domain.OperationLocate, Subject: "clothes", Confidence: domain.ResponseConfidenceAbsent}})
	if err != nil {
		t.Fatal(err)
	}
	if result.SpokenResponse == "" || execution.ModelCalls() != 0 {
		t.Fatalf("grounded mode used provider or produced no response: %+v", result)
	}
}

func TestWorkflowContinuationAvailabilityRespectsSharedBudgets(t *testing.T) {
	execution, provider, clock := workflowExecutionFixture(t, 2)
	if !execution.CanContinue() {
		t.Fatal("fresh workflow unavailable")
	}
	_, err := execution.NextTurn(context.Background(), ports.LanguageInferenceInput{Investigation: &domain.InvestigationInput{Phase: domain.InvestigationPhaseInitial}})
	if err != nil {
		t.Fatal(err)
	}
	if execution.CanContinue() || provider.calls != 2 {
		t.Fatal("exhausted call budget advertised follow-up")
	}
	execution, _, clock = workflowExecutionFixture(t, 2)
	clock.now = clock.now.Add(time.Minute)
	if execution.CanContinue() {
		t.Fatal("elapsed budget advertised follow-up")
	}
}
