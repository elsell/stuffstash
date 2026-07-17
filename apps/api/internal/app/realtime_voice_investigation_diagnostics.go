package app

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

func emitRealtimeVoiceInvestigationDiagnostic(session RealtimeVoiceSession, input agentmodel.InvestigationInput, step agentmodel.InvestigationStep, emit RealtimeVoiceEventSink) error {
	if !session.DeveloperDiagnostics {
		return nil
	}
	payload := struct {
		Phase                     string `json:"phase"`
		EvidenceRound             int    `json:"evidenceRound"`
		MaxEvidenceRounds         int    `json:"maxEvidenceRounds"`
		PromptVersion             string `json:"promptVersion"`
		SchemaVersion             string `json:"schemaVersion"`
		PreviousRequestCount      int    `json:"previousRequestCount"`
		ObservationCount          int    `json:"observationCount"`
		ReadEvidenceCount         int    `json:"readEvidenceCount"`
		CustomAssetTypeCount      int    `json:"customAssetTypeCount"`
		CustomFieldCount          int    `json:"customFieldCount"`
		TagCount                  int    `json:"tagCount"`
		VocabularyRequestCount    int    `json:"vocabularyRequestCount"`
		VocabularyDefinitionCount int    `json:"vocabularyDefinitionCount"`
		Decision                  string `json:"decision"`
		IntentKind                string `json:"intentKind"`
		Operation                 string `json:"operation"`
		SearchRequestCount        int    `json:"searchRequestCount"`
		ResolutionCount           int    `json:"resolutionCount"`
	}{
		Phase: string(input.Phase), EvidenceRound: input.EvidenceRound, MaxEvidenceRounds: input.MaxEvidenceRounds,
		PromptVersion: safeRealtimeVoiceDiagnosticVersion(input.PromptVersion), SchemaVersion: safeRealtimeVoiceDiagnosticVersion(input.SchemaVersion),
		PreviousRequestCount: len(input.PreviousRequests), ObservationCount: len(input.Observations), ReadEvidenceCount: len(input.ReadEvidence),
		CustomAssetTypeCount: len(input.Vocabulary.CustomAssetTypes), CustomFieldCount: len(input.Vocabulary.CustomFields), TagCount: len(input.Vocabulary.Tags),
		VocabularyRequestCount: len(input.VocabularyRequests), VocabularyDefinitionCount: len(input.VocabularyDefinitions),
		Decision: string(step.Decision), IntentKind: string(step.Intent.Kind), Operation: string(step.Intent.Operation),
		SearchRequestCount: len(step.SearchRequests), ResolutionCount: len(step.Resolutions),
	}
	detail, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return emitRealtimeVoiceDiagnostic(session.ID, fmt.Sprintf("Language investigation (turn %d)", input.EvidenceRound+1), string(detail), emit)
}

func safeRealtimeVoiceDiagnosticVersion(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 100 {
		return "unknown"
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '.' || char == '-' || char == '_' {
			continue
		}
		return "unknown"
	}
	return value
}
