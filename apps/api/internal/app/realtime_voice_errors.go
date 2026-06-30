package app

import (
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

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
	if !boundedRealtimeVoiceText(response.SpokenResponse, 500) {
		return ports.ErrInvalidProviderInput
	}
	if strings.TrimSpace(response.DisplayResponse) != "" && !boundedRealtimeVoiceText(response.DisplayResponse, 1000) {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func boundedRealtimeVoiceText(value string, limit int) bool {
	value = strings.TrimSpace(value)
	return value != "" && len(value) <= limit
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
