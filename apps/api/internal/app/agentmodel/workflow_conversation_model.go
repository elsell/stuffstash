package agentmodel

import (
	"context"
	"math"
	"strings"
	"time"

	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// workflowConversationModel adds pinned tenant guidance to each invocation.
// The conversation loop and its caller own per-turn call and time limits.
type workflowConversationModel struct {
	model        ports.ConversationModel
	instructions string
}

func newWorkflowConversationModel(model ports.ConversationModel, settings domain.WorkflowDefinitionInput, prompt string) (*workflowConversationModel, error) {
	if model == nil || settings.Budget.ModelCalls <= 0 || settings.Budget.ToolCalls <= 0 || settings.Budget.ElapsedSeconds <= 0 || int64(settings.Budget.ElapsedSeconds) > math.MaxInt64/int64(time.Second) {
		return nil, ports.ErrInvalidProviderInput
	}
	return &workflowConversationModel{model: model, instructions: strings.TrimSpace(prompt + "\n" + settings.Instructions)}, nil
}
func (m *workflowConversationModel) Converse(ctx context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	if err := ctx.Err(); err != nil {
		return ports.ConversationModelTurn{}, err
	}
	input.Instructions = strings.TrimSpace(input.Instructions + "\nTenant conversation guidance:\n" + m.instructions)
	return m.model.Converse(ctx, input)
}
func (p *PreparedWorkflow) ConversationModel() ports.ConversationModel {
	if p.conversation == nil {
		return nil
	}
	return p.conversation
}
func (p *PreparedWorkflow) ConversationProfileID() string { return p.conversationProfileID }
