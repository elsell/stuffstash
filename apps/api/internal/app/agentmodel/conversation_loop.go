package agentmodel

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ConversationLimits struct{ ModelCalls, ToolCalls int }
type ConversationResult struct {
	Answer         *ports.ConversationAnswer
	Messages       []ports.ConversationMessage
	ApprovalPlanID string
}

// The red-test implementation is deliberately not wired into production.
func RunConversation(ctx context.Context, model ports.ConversationModel, executor ports.ConversationToolExecutor, input ports.ConversationModelInput, limits ConversationLimits) (ConversationResult, error) {
	return ConversationResult{}, errors.New("model-led conversation loop not implemented")
}
