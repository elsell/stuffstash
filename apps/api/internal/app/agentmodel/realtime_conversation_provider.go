package agentmodel

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// Stage attribution is applied only around the model port, never around tool execution.
type realtimeConversationProvider struct{ model ports.ConversationModel }

func (p realtimeConversationProvider) Converse(ctx context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	turn, err := p.model.Converse(ctx, input)
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, ports.ErrForbidden) || errors.Is(err, ports.ErrUnauthenticated) || errors.Is(err, ErrConversationBudgetExhausted) {
		return turn, err
	}
	return turn, RealtimeVoiceProviderStageError{Code: RealtimeVoiceFailureLanguageInference, Cause: err}
}
