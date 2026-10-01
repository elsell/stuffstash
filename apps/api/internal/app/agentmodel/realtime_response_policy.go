package agentmodel

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/stuffstash/stuff-stash/internal/app/agentmodel/tools"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func ValidateRealtimeVoiceFinalResponse(response ports.StructuredAgentResponse) error {
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
	if !SafeRealtimeVoiceFinalText(response.SpokenResponse, 500) {
		return ports.ErrInvalidProviderInput
	}
	if response.DisplayResponse != "" && !SafeRealtimeVoiceFinalText(response.DisplayResponse, 1000) {
		return ports.ErrInvalidProviderInput
	}
	if err := tools.ValidateRealtimeVoiceResponseArtifacts(response.DisplayResponse, response.Artifacts); err != nil {
		return err
	}
	return nil
}

func SafeRealtimeVoiceFinalText(value string, limit int) bool {
	return strings.TrimSpace(value) != "" && len(value) <= limit && utf8.ValidString(value)
}

func RealtimeVoiceErrorCode(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "request_timeout"
	case errors.Is(err, ErrConversationBudgetExhausted):
		return "conversation_budget_exhausted"
	case errors.Is(err, ErrConversationContextExhausted):
		return "conversation_context_exhausted"
	}

	var providerErr RealtimeVoiceProviderStageError
	if errors.As(err, &providerErr) {
		if errors.Is(providerErr.Cause, ports.ErrInvalidProviderInput) {
			return "invalid_provider_output"
		}
		if SafeRealtimeVoiceProviderDiagnosticError(providerErr.Cause) == "provider_billing_disabled" {
			return "provider_billing_disabled"
		}
		return providerErr.Code
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

func SafeRealtimeVoiceErrorDetail(err error) string {
	code := RealtimeVoiceErrorCode(err)
	if code == "request_timeout" || code == "conversation_budget_exhausted" || code == "conversation_context_exhausted" {
		return code
	}

	if err == nil {
		return ""
	}
	var providerErr RealtimeVoiceProviderStageError
	if errors.As(err, &providerErr) {
		if errors.Is(providerErr.Cause, ports.ErrInvalidProviderInput) {
			return "invalid_provider_output"
		}
		if SafeRealtimeVoiceProviderDiagnosticError(providerErr.Cause) == "provider_billing_disabled" {
			return "provider_billing_disabled"
		}
		return providerErr.Code
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

type RealtimeVoiceProviderStageError struct {
	Code  string
	Cause error
}

func (e RealtimeVoiceProviderStageError) Error() string {
	return e.Code
}

func (e RealtimeVoiceProviderStageError) Unwrap() error {
	return e.Cause
}
