package agentmodel

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

var ErrWorkflowBudgetExhausted = errors.New("conversation workflow budget exhausted")

type WorkflowModelBinding struct {
	ProfileID      string
	PromptTemplate string
	Language       ports.LanguageInferenceProvider
	Response       ports.VoiceResponseGenerator
}
type WorkflowModelExecution struct {
	definition domain.WorkflowDefinition
	clock      ports.Clock
	started    time.Time
	steps      map[domain.WorkflowStepKind]domain.WorkflowStep
	providers  map[domain.WorkflowStepKind]WorkflowModelBinding
	mu         sync.Mutex
	calls      int
	expired    bool
}

func NewWorkflowModelExecution(definition domain.WorkflowDefinition, limits domain.WorkflowLimits, clock ports.Clock, bindings map[domain.WorkflowStepKind]WorkflowModelBinding) (*WorkflowModelExecution, error) {
	validated, err := domain.NewWorkflowDefinition(definition.Settings(), limits)
	if err != nil || clock == nil {
		return nil, domain.ErrInvalidWorkflowDefinition
	}
	settings := validated.Settings()
	if int64(settings.Budget.ElapsedSeconds) > math.MaxInt64/int64(time.Second) {
		return nil, domain.ErrInvalidWorkflowDefinition
	}
	execution := &WorkflowModelExecution{definition: validated, clock: clock, started: clock.Now(), steps: map[domain.WorkflowStepKind]domain.WorkflowStep{}, providers: map[domain.WorkflowStepKind]WorkflowModelBinding{}}
	for _, step := range settings.Steps {
		binding := bindings[step.Kind]
		if step.ProviderProfileID != "" && binding.ProfileID != step.ProviderProfileID {
			return nil, ports.ErrInvalidProviderInput
		}
		if step.Kind == domain.WorkflowStepRespond {
			if settings.Response != domain.WorkflowResponseGrounded && binding.Response == nil {
				return nil, ports.ErrInvalidProviderInput
			}
		} else if binding.Language == nil {
			return nil, ports.ErrInvalidProviderInput
		}
		execution.steps[step.Kind] = step
		execution.providers[step.Kind] = binding
	}
	return execution, nil
}
func (e *WorkflowModelExecution) reserve(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	budget := e.definition.Settings().Budget
	elapsed := e.clock.Now().Sub(e.started)
	if elapsed < 0 {
		elapsed = 0
	}
	remaining := time.Duration(budget.ElapsedSeconds)*time.Second - elapsed
	if e.expired || e.calls >= budget.ModelCalls || remaining <= 0 {
		return nil, nil, ErrWorkflowBudgetExhausted
	}
	e.calls++
	attempt, cancel := context.WithTimeout(ctx, remaining)
	return attempt, cancel, nil
}
func (e *WorkflowModelExecution) ModelCalls() int { e.mu.Lock(); defer e.mu.Unlock(); return e.calls }
func (e *WorkflowModelExecution) NextTurn(ctx context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	if input.Investigation == nil {
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	var kind domain.WorkflowStepKind
	switch input.Investigation.Phase {
	case domain.InvestigationPhaseInitial:
		kind = domain.WorkflowStepInterpret
	case domain.InvestigationPhaseEvidenceAssessment:
		kind = domain.WorkflowStepAssess
	default:
		return ports.LanguageInferenceTurn{}, ports.ErrInvalidProviderInput
	}
	step := e.steps[kind]
	input.WorkflowInstructions = step.Instructions
	if e.providers[kind].PromptTemplate != "" {
		input.PromptTemplate = e.providers[kind].PromptTemplate
	}
	return executeWorkflowModelStep(ctx, e, kind, func(callCtx context.Context) (ports.LanguageInferenceTurn, error) {
		return e.providers[kind].Language.NextTurn(callCtx, input)
	})
}

func executeWorkflowModelStep[T any](ctx context.Context, e *WorkflowModelExecution, kind domain.WorkflowStepKind, invoke func(context.Context) (T, error)) (T, error) {
	var zero T
	var last error
	for attempt := 0; attempt < e.steps[kind].Attempts; attempt++ {
		callCtx, cancel, err := e.reserve(ctx)
		if err != nil {
			return zero, err
		}
		result, err := invoke(callCtx)
		deadlineErr := callCtx.Err()
		cancel()
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}
		if deadlineErr != nil {
			e.mu.Lock()
			e.expired = true
			e.mu.Unlock()
			return zero, ErrWorkflowBudgetExhausted
		}
		if err == nil {
			return result, nil
		}
		if errors.Is(err, context.Canceled) {
			return zero, err
		}
		last = err
	}
	return zero, last
}
