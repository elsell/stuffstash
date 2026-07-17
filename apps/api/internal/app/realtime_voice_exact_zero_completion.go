package app

import (
	"fmt"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

func realtimeVoiceExactOrZeroCompletion(intent agentmodel.Intent, step agentmodel.InvestigationStep, observations []agentmodel.CandidateObservation, evidence []agentmodel.ReadEvidence) (agentmodel.InvestigationStep, bool) {
	if (step.Decision != agentmodel.InvestigationDecisionSearchAgain && step.Decision != agentmodel.InvestigationDecisionFinish) ||
		(intent.Operation != agentmodel.OperationCreate && intent.Operation != agentmodel.OperationMove) {
		return step, false
	}
	covered := map[agentmodel.SemanticReferenceKey]bool{}
	for _, record := range evidence {
		switch record.ReadKind {
		case agentmodel.InvestigationReadSearchAssets, agentmodel.InvestigationReadListInventory, agentmodel.InvestigationReadListContents:
			covered[record.ReferenceKey] = true
		}
	}
	byReference := map[agentmodel.SemanticReferenceKey][]agentmodel.CandidateObservation{}
	for _, observation := range observations {
		byReference[observation.ReferenceKey] = append(byReference[observation.ReferenceKey], observation)
	}
	resolutions := make([]agentmodel.Resolution, 0, len(realtimeVoiceInvestigationReferenceKeys(intent)))
	missingSuffix := false
	for _, reference := range realtimeVoiceInvestigationReferenceKeys(intent) {
		if !covered[reference] {
			return step, false
		}
		if reference != agentmodel.SemanticReferenceSubject && missingSuffix {
			resolutions = append(resolutions, missingRealtimeVoiceDestinationResolution(reference))
			continue
		}
		candidates := byReference[reference]
		if reference != agentmodel.SemanticReferenceSubject {
			var destinationIndex int
			if _, err := fmt.Sscanf(reference.String(), "destination.%d", &destinationIndex); err != nil || destinationIndex < 0 || destinationIndex >= len(intent.DestinationKinds) {
				return step, false
			}
			compatible := make([]agentmodel.CandidateObservation, 0, len(candidates))
			for _, candidate := range candidates {
				if candidate.Kind == string(intent.DestinationKinds[destinationIndex]) {
					compatible = append(compatible, candidate)
				}
			}
			candidates = compatible
		}
		if len(candidates) == 0 {
			status := agentmodel.ResolutionAbsent
			if intent.Operation == agentmodel.OperationCreate || reference != agentmodel.SemanticReferenceSubject {
				status = agentmodel.ResolutionMissing
			}
			resolutions = append(resolutions, agentmodel.Resolution{ReferenceKey: reference, Status: status, Evidence: "Executed authorized discovery returned no candidates for this reference."})
			if reference != agentmodel.SemanticReferenceSubject {
				missingSuffix = true
			}
			continue
		}
		mention := realtimeVoiceInvestigationReferenceMention(intent, reference)
		exactID := ""
		for _, candidate := range candidates {
			if !realtimeVoiceInvestigationTitleMatchesMention(candidate.Title, mention) {
				continue
			}
			if exactID != "" {
				return step, false
			}
			exactID = candidate.CandidateID
		}
		if exactID == "" {
			return step, false
		}
		resolutions = append(resolutions, agentmodel.Resolution{
			ReferenceKey: reference, Status: agentmodel.ResolutionStrong, CandidateIDs: []string{exactID},
			Evidence: "The application selected the sole authorized exact normalized title match for this reference.",
		})
	}
	completed := agentmodel.InvestigationStep{
		Decision: agentmodel.InvestigationDecisionFinish, Intent: intent, Resolutions: resolutions,
		Rationale: "Complete exact-or-zero references from executed authorized discovery.",
	}
	if completed.Validate() != nil {
		return agentmodel.InvestigationStep{}, false
	}
	return completed, true
}
