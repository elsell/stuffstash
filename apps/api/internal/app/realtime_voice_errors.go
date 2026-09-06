package app

import (
	"context"
	"errors"
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strings"
	"unicode/utf8"
)

var errRealtimeVoiceToolCallTimedOut = errors.New("realtime voice tool call timed out")

func validateRealtimeVoiceFinalResponse(response ports.StructuredAgentResponse) error {
	kind := response.Kind
	if kind == "" {
		kind = ports.StructuredAgentResponseKindAnswer
	}
	switch kind {
	case ports.StructuredAgentResponseKindAnswer,
		ports.StructuredAgentResponseKindClarification,
		ports.StructuredAgentResponseKindUnsupportedAction,
		ports.StructuredAgentResponseKindSafeFailure:
	default:
		return ports.ErrInvalidProviderInput
	}
	if !safeRealtimeVoiceFinalText(response.SpokenResponse, 500) {
		return ports.ErrInvalidProviderInput
	}
	if response.DisplayResponse != "" && !safeRealtimeVoiceFinalText(response.DisplayResponse, 1000) {
		return ports.ErrInvalidProviderInput
	}
	if err := validateRealtimeVoiceResponseArtifacts(response.DisplayResponse, response.Artifacts); err != nil {
		return err
	}
	return nil
}

func safeRealtimeVoiceFinalText(value string, limit int) bool {
	return strings.TrimSpace(value) != "" && len(value) <= limit && utf8.ValidString(value)
}

func realtimeVoiceErrorCode(err error) string {
	var providerErr realtimeVoiceProviderStageError
	if errors.As(err, &providerErr) {
		return providerErr.code
	}
	switch {
	case errors.Is(err, ports.ErrUnauthenticated):
		return "unauthenticated"
	case errors.Is(err, ports.ErrForbidden), errors.Is(err, apperrors.ErrNotFound):
		return "forbidden"
	case errors.Is(err, ports.ErrInvalidProviderInput), errors.Is(err, apperrors.ErrInvalidInput):
		return "invalid_request"
	default:
		return "voice_session_failed"
	}
}

func safeRealtimeVoiceErrorDetail(err error) string {
	if err == nil {
		return ""
	}
	var providerErr realtimeVoiceProviderStageError
	if errors.As(err, &providerErr) {
		return providerErr.code
	}
	switch {
	case errors.Is(err, ports.ErrInvalidProviderInput):
		return "invalid_provider_input"
	case errors.Is(err, apperrors.ErrInvalidInput):
		return "invalid_input"
	case errors.Is(err, ports.ErrForbidden):
		return "forbidden"
	case errors.Is(err, ports.ErrUnauthenticated):
		return "unauthenticated"
	default:
		return "unexpected_error"
	}
}

type realtimeVoiceProviderStageError struct {
	code string
	err  error
}

func (e realtimeVoiceProviderStageError) Error() string {
	return e.code
}

func (e realtimeVoiceProviderStageError) Unwrap() error {
	return e.err
}

// Stage attribution is applied only around the model port, never around tool execution.
type realtimeConversationProvider struct{ model ports.ConversationModel }

func (p realtimeConversationProvider) Converse(ctx context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	turn, err := p.model.Converse(ctx, input)
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, ports.ErrForbidden) || errors.Is(err, ports.ErrUnauthenticated) || errors.Is(err, agentmodelapp.ErrWorkflowBudgetExhausted) || errors.Is(err, agentmodelapp.ErrConversationBudgetExhausted) {
		return turn, err
	}
	return turn, realtimeVoiceProviderStageError{code: realtimeVoiceFailureLanguageInference, err: err}
}
