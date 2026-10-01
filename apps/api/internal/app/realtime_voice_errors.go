package app

import (
	"context"
	"errors"

	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

var errRealtimeVoiceToolCallTimedOut = errors.New("realtime voice tool call timed out")

type realtimeVoiceProviderStageError = agentmodelapp.RealtimeVoiceProviderStageError

func validateRealtimeVoiceFinalResponse(response ports.StructuredAgentResponse) error {
	return agentmodelapp.ValidateRealtimeVoiceFinalResponse(response)
}
func safeRealtimeVoiceFinalText(value string, limit int) bool {
	return agentmodelapp.SafeRealtimeVoiceFinalText(value, limit)
}
func realtimeVoiceErrorCode(err error) string { return agentmodelapp.RealtimeVoiceErrorCode(err) }
func safeRealtimeVoiceErrorDetail(err error) string {
	return agentmodelapp.SafeRealtimeVoiceErrorDetail(err)
}

// Stage attribution is applied only around the model port, never around tool execution.
type realtimeConversationProvider struct{ model ports.ConversationModel }

func (p realtimeConversationProvider) Converse(ctx context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	turn, err := p.model.Converse(ctx, input)
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, ports.ErrForbidden) || errors.Is(err, ports.ErrUnauthenticated) || errors.Is(err, agentmodelapp.ErrConversationBudgetExhausted) {
		return turn, err
	}
	return turn, realtimeVoiceProviderStageError{Code: realtimeVoiceFailureLanguageInference, Cause: err}
}
