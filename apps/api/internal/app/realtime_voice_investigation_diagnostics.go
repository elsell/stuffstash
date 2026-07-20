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
		Phase                     string   `json:"phase"`
		EvidenceRound             int      `json:"evidenceRound"`
		MaxEvidenceRounds         int      `json:"maxEvidenceRounds"`
		PromptVersion             string   `json:"promptVersion"`
		SchemaVersion             string   `json:"schemaVersion"`
		PreviousRequestCount      int      `json:"previousRequestCount"`
		ObservationCount          int      `json:"observationCount"`
		ReadEvidenceCount         int      `json:"readEvidenceCount"`
		CustomAssetTypeCount      int      `json:"customAssetTypeCount"`
		CustomFieldCount          int      `json:"customFieldCount"`
		TagCount                  int      `json:"tagCount"`
		VocabularyRequestCount    int      `json:"vocabularyRequestCount"`
		VocabularyDefinitionCount int      `json:"vocabularyDefinitionCount"`
		Decision                  string   `json:"decision"`
		RequestShape              string   `json:"requestShape"`
		IntentKind                string   `json:"intentKind"`
		Operation                 string   `json:"operation"`
		DestinationCount          int      `json:"destinationCount"`
		SearchRequestCount        int      `json:"searchRequestCount"`
		ResolutionCount           int      `json:"resolutionCount"`
		ResolutionStatuses        []string `json:"resolutionStatuses"`
		ResolutionCandidateCounts []int    `json:"resolutionCandidateCounts"`
	}{
		Phase: string(input.Phase), EvidenceRound: input.EvidenceRound, MaxEvidenceRounds: input.MaxEvidenceRounds,
		PromptVersion: safeRealtimeVoiceDiagnosticVersion(input.PromptVersion), SchemaVersion: safeRealtimeVoiceDiagnosticVersion(input.SchemaVersion),
		PreviousRequestCount: len(input.PreviousRequests), ObservationCount: len(input.Observations), ReadEvidenceCount: len(input.ReadEvidence),
		CustomAssetTypeCount: len(input.Vocabulary.CustomAssetTypes), CustomFieldCount: len(input.Vocabulary.CustomFields), TagCount: len(input.Vocabulary.Tags),
		VocabularyRequestCount: len(input.VocabularyRequests), VocabularyDefinitionCount: len(input.VocabularyDefinitions),
		Decision: string(step.Decision), RequestShape: string(step.Intent.RequestShape), IntentKind: string(step.Intent.Kind), Operation: string(step.Intent.Operation), DestinationCount: len(step.Intent.DestinationPath),
		SearchRequestCount: len(step.SearchRequests), ResolutionCount: len(step.Resolutions),
		ResolutionStatuses: realtimeVoiceDiagnosticResolutionStatuses(step.Resolutions), ResolutionCandidateCounts: realtimeVoiceDiagnosticResolutionCandidateCounts(step.Resolutions),
	}
	detail, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return emitRealtimeVoiceDiagnostic(session.ID, fmt.Sprintf("Language investigation (turn %d)", input.EvidenceRound+1), string(detail), emit)
}

func realtimeVoiceDiagnosticResolutionStatuses(resolutions []agentmodel.Resolution) []string {
	statuses := make([]string, 0, len(resolutions))
	for _, resolution := range resolutions {
		statuses = append(statuses, string(resolution.Status))
	}
	return statuses
}

func realtimeVoiceDiagnosticResolutionCandidateCounts(resolutions []agentmodel.Resolution) []int {
	counts := make([]int, 0, len(resolutions))
	for _, resolution := range resolutions {
		counts = append(counts, len(resolution.CandidateIDs))
	}
	return counts
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
