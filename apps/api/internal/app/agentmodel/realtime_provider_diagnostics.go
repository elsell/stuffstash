package agentmodel

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func emitRealtimeVoiceTextToSpeechFailureDiagnostic(session RealtimeResponseSession, toolResults []ports.AgentToolResult, safeCode string, err error, emit RealtimeVoiceEventSink) error {
	if !session.DeveloperDiagnostics {
		return nil
	}
	payload, marshalErr := json.MarshalIndent(map[string]any{
		"stage":           "text_to_speech",
		"safeCode":        strings.TrimSpace(safeCode),
		"safeError":       SafeRealtimeVoiceProviderDiagnosticError(err),
		"toolResultCount": len(toolResults),
		"toolNames":       realtimeVoiceToolResultNames(toolResults),
	}, "", "  ")
	if marshalErr != nil {
		return EmitRealtimeVoiceDiagnostic(session.ID, "Text-to-speech provider failed", "Text-to-speech provider failure diagnostic could not be rendered safely.", emit)
	}
	return EmitRealtimeVoiceDiagnostic(session.ID, "Text-to-speech provider failed", string(payload), emit)
}

func realtimeVoiceToolResultNames(toolResults []ports.AgentToolResult) []string {
	toolNames := make([]string, 0, len(toolResults))
	for _, result := range toolResults {
		name := strings.TrimSpace(result.Name)
		if name == "" {
			continue
		}
		toolNames = append(toolNames, name)
	}
	return toolNames
}

type realtimeVoiceSafeDiagnosticError interface {
	SafeRealtimeVoiceDiagnostic() string
}

func SafeRealtimeVoiceProviderDiagnosticError(err error) string {
	if errors.Is(err, ports.ErrInvalidProviderInput) {
		return "invalid_provider_output"
	}
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
	if value == "provider_billing_disabled" || value == "provider_request_failed" || value == "provider_timeout" || value == "provider_auth_failed" || value == "provider_rate_limited" || value == "invalid_provider_output" {
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
