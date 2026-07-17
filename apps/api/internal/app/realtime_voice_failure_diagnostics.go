package app

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func emitRealtimeVoiceLanguageFailureDiagnostic(session RealtimeVoiceSession, input agentmodel.InvestigationInput, toolResults []ports.AgentToolResult, safeCode string, err error, emit RealtimeVoiceEventSink) error {
	if !session.DeveloperDiagnostics {
		return nil
	}
	payload, marshalErr := json.MarshalIndent(map[string]any{
		"stage":                     "language_inference",
		"safeCode":                  strings.TrimSpace(safeCode),
		"safeError":                 safeRealtimeVoiceProviderDiagnosticError(err),
		"phase":                     string(input.Phase),
		"evidenceRound":             input.EvidenceRound,
		"maxEvidenceRounds":         input.MaxEvidenceRounds,
		"promptVersion":             safeRealtimeVoiceDiagnosticVersion(input.PromptVersion),
		"schemaVersion":             safeRealtimeVoiceDiagnosticVersion(input.SchemaVersion),
		"previousRequestCount":      len(input.PreviousRequests),
		"observationCount":          len(input.Observations),
		"readEvidenceCount":         len(input.ReadEvidence),
		"customAssetTypeCount":      len(input.Vocabulary.CustomAssetTypes),
		"customFieldCount":          len(input.Vocabulary.CustomFields),
		"tagCount":                  len(input.Vocabulary.Tags),
		"vocabularyRequestCount":    len(input.VocabularyRequests),
		"vocabularyDefinitionCount": len(input.VocabularyDefinitions),
		"toolResultCount":           len(toolResults),
		"toolNames":                 realtimeVoiceToolResultNames(toolResults),
	}, "", "  ")
	if marshalErr != nil {
		return emitRealtimeVoiceDiagnostic(session.ID, "Language provider failed", "Language provider failure diagnostic could not be rendered safely.", emit)
	}
	return emitRealtimeVoiceDiagnostic(session.ID, "Language provider failed", string(payload), emit)
}

func emitRealtimeVoiceTextToSpeechFailureDiagnostic(session RealtimeVoiceSession, toolResults []ports.AgentToolResult, safeCode string, err error, emit RealtimeVoiceEventSink) error {
	if !session.DeveloperDiagnostics {
		return nil
	}
	payload, marshalErr := json.MarshalIndent(map[string]any{
		"stage":           "text_to_speech",
		"safeCode":        strings.TrimSpace(safeCode),
		"safeError":       safeRealtimeVoiceProviderDiagnosticError(err),
		"toolResultCount": len(toolResults),
		"toolNames":       realtimeVoiceToolResultNames(toolResults),
	}, "", "  ")
	if marshalErr != nil {
		return emitRealtimeVoiceDiagnostic(session.ID, "Text-to-speech provider failed", "Text-to-speech provider failure diagnostic could not be rendered safely.", emit)
	}
	return emitRealtimeVoiceDiagnostic(session.ID, "Text-to-speech provider failed", string(payload), emit)
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

func safeRealtimeVoiceProviderDiagnosticError(err error) string {
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
	if value == "provider_request_failed" || value == "provider_timeout" || value == "provider_auth_failed" || value == "provider_rate_limited" || value == "invalid_provider_output" {
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
