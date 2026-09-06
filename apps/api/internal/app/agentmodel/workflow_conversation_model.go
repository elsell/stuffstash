package agentmodel

import (
	"context"
	"errors"
	"math"
	"strings"
	"sync"
	"time"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

var ErrWorkflowBudgetExhausted = errors.New("conversation workflow budget exhausted")

// workflowConversationModel supplies tenant guidance and one shared, lazy
// session budget. Each call invokes the selected native model exactly once.
type workflowConversationModel struct {
	model        ports.ConversationModel
	clock        ports.Clock
	budget       domain.WorkflowBudget
	instructions string
	mu           sync.Mutex
	started      bool
	startedAt    time.Time
	calls        int
	expired      bool
}

func newWorkflowConversationModel(model ports.ConversationModel, clock ports.Clock, settings domain.WorkflowDefinitionInput, prompt string) (*workflowConversationModel, error) {
	if model == nil || clock == nil || settings.Budget.ModelCalls <= 0 || settings.Budget.ElapsedSeconds <= 0 || int64(settings.Budget.ElapsedSeconds) > math.MaxInt64/int64(time.Second) {
		return nil, ports.ErrInvalidProviderInput
	}
	return &workflowConversationModel{model: model, clock: clock, budget: settings.Budget, instructions: strings.TrimSpace(prompt + "\n" + settings.Instructions)}, nil
}
func (m *workflowConversationModel) Converse(ctx context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	if err := ctx.Err(); err != nil {
		return ports.ConversationModelTurn{}, err
	}
	m.mu.Lock()
	if !m.started {
		m.started = true
		m.startedAt = m.clock.Now()
	}
	elapsed := m.clock.Now().Sub(m.startedAt)
	if elapsed < 0 {
		elapsed = 0
	}
	remaining := time.Duration(m.budget.ElapsedSeconds)*time.Second - elapsed
	if m.expired || m.calls >= m.budget.ModelCalls || remaining <= 0 {
		m.mu.Unlock()
		return ports.ConversationModelTurn{}, ErrWorkflowBudgetExhausted
	}
	m.calls++
	m.mu.Unlock()
	callCtx, cancel := context.WithTimeout(ctx, remaining)
	defer cancel()
	input.Instructions = strings.TrimSpace(input.Instructions + "\nTenant conversation guidance:\n" + m.instructions)
	turn, err := m.model.Converse(callCtx, input)
	if callCtx.Err() == context.DeadlineExceeded {
		m.mu.Lock()
		m.expired = true
		m.mu.Unlock()
	}
	if callCtx.Err() != nil {
		return ports.ConversationModelTurn{}, callCtx.Err()
	}
	return turn, err
}
func (m *workflowConversationModel) CanContinue() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return !m.expired && m.calls < m.budget.ModelCalls && (!m.started || m.clock.Now().Sub(m.startedAt) < time.Duration(m.budget.ElapsedSeconds)*time.Second)
}
func (p *PreparedWorkflow) ConversationModel() ports.ConversationModel {
	if p.conversation == nil {
		return nil
	}
	return p.conversation
}
func (p *PreparedWorkflow) ConversationProfileID() string { return p.conversationProfileID }
