package app

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func emitRealtimeVoiceLanguageFailureDiagnostic(session RealtimeVoiceSession, turn int, finalOnly bool, toolResults []ports.AgentToolResult, safeCode string, err error, emit RealtimeVoiceEventSink) error {
	if !session.DeveloperDiagnostics {
		return nil
	}
	toolNames := make([]string, 0, len(toolResults))
	for _, result := range toolResults {
		name := strings.TrimSpace(result.Name)
		if name == "" {
			continue
		}
		toolNames = append(toolNames, name)
	}
	payload, marshalErr := json.MarshalIndent(map[string]any{
		"stage":           "language_inference",
		"safeCode":        strings.TrimSpace(safeCode),
		"safeError":       safeRealtimeVoiceProviderDiagnosticError(err),
		"turn":            turn,
		"previousTurns":   max(turn-1, 0),
		"finalOnly":       finalOnly,
		"toolResultCount": len(toolResults),
		"toolNames":       toolNames,
	}, "", "  ")
	if marshalErr != nil {
		return emitRealtimeVoiceDiagnostic(session.ID, "Language provider failed", "Language provider failure diagnostic could not be rendered safely.", emit)
	}
	return emitRealtimeVoiceDiagnostic(session.ID, "Language provider failed", string(payload), emit)
}

type realtimeVoiceSafeDiagnosticError interface {
	SafeRealtimeVoiceDiagnostic() string
}

func safeRealtimeVoiceProviderDiagnosticError(err error) string {
	var safeErr realtimeVoiceSafeDiagnosticError
	if errors.As(err, &safeErr) {
		value := strings.TrimSpace(safeErr.SafeRealtimeVoiceDiagnostic())
		if safeRealtimeVoiceProviderDiagnosticCategory(value) {
			return value
		}
	}
	return "provider_request_failed"
}

func safeRealtimeVoiceProviderDiagnosticCategory(value string) bool {
	if value == "provider_request_failed" || value == "provider_timeout" || value == "provider_auth_failed" || value == "provider_rate_limited" {
		return true
	}
	if !strings.HasPrefix(value, "provider_http_status_") {
		return false
	}
	status := strings.TrimPrefix(value, "provider_http_status_")
	if len(status) != 3 {
		return false
	}
	for _, char := range status {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
