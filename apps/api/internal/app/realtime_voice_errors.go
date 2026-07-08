package app

import (
	"errors"
	"regexp"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/ports"
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
	if strings.TrimSpace(response.DisplayResponse) != "" && !safeRealtimeVoiceFinalText(response.DisplayResponse, 1000) {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func safeRealtimeVoiceFinalText(value string, limit int) bool {
	value = strings.TrimSpace(value)
	return value != "" && len(value) <= limit && !realtimeVoiceFinalTextLooksUnsafe(value)
}

func realtimeVoiceFinalTextLooksUnsafe(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if strings.HasPrefix(normalized, "{") || strings.HasPrefix(normalized, "[") || realtimeVoiceFinalJSONFragmentPattern.MatchString(value) {
		return true
	}
	for _, token := range []string{
		"```",
		"search_authorized_assets",
		"list_authorized_assets",
		"get_asset_detail",
		"list_asset_audit_history",
		"list_asset_checkout_history",
		"list_checked_out_assets",
		"propose_action_plan",
		"chain of thought",
		"reasoning:",
		"raw prompt",
		"provider response",
		"provider session",
		"stack trace",
		"raw transcript",
		"raw audio",
		"tool_call",
		"functioncall",
	} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	collapsed := realtimeVoiceFinalCollapsedText(normalized)
	for _, token := range []string{"assetid", "parentassetid", "inventoryid", "tenantid", "toolcallid"} {
		if strings.Contains(collapsed, token) {
			return true
		}
	}
	return realtimeVoiceFinalSecretPattern.MatchString(value)
}

var (
	realtimeVoiceFinalJSONFragmentPattern = regexp.MustCompile(`["']?[a-zA-Z][a-zA-Z0-9_-]*["']?\s*:\s*["'{\[]`)
	realtimeVoiceFinalSecretPattern       = regexp.MustCompile(`(?i)\b(api[-_ ]?key|authorization|credential|password|secret|token)\s*[:=]\s*["']?(bearer\s+)?[a-z0-9._~+/=-]{16,}|bearer\s+[^"',\s}\]\)]+`)
)

func realtimeVoiceFinalCollapsedText(value string) string {
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
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
