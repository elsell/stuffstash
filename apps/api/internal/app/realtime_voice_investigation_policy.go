package app

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func canonicalRealtimeVoiceInvestigationStep(canonicalIntent agentmodel.Intent, step agentmodel.InvestigationStep, observations []agentmodel.CandidateObservation) (agentmodel.InvestigationStep, error) {
	if step.Decision != agentmodel.InvestigationDecisionFinish || step.Validate() != nil || !sameRealtimeVoiceInvestigationIntent(canonicalIntent, step.Intent) {
		return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
	}
	byReference := map[agentmodel.SemanticReferenceKey]map[string]agentmodel.CandidateObservation{}
	allCandidates := map[string]agentmodel.CandidateObservation{}
	for _, observation := range observations {
		if observation.Validate() != nil {
			return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
		}
		if byReference[observation.ReferenceKey] == nil {
			byReference[observation.ReferenceKey] = map[string]agentmodel.CandidateObservation{}
		}
		byReference[observation.ReferenceKey][observation.CandidateID] = observation
		allCandidates[observation.CandidateID] = observation
	}
	resolutions := map[agentmodel.SemanticReferenceKey]agentmodel.Resolution{}
	for _, resolution := range step.Resolutions {
		if _, duplicate := resolutions[resolution.ReferenceKey]; duplicate {
			return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
		}
		for _, candidateID := range resolution.CandidateIDs {
			if _, visibleForReference := byReference[resolution.ReferenceKey][candidateID]; !visibleForReference {
				return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
			}
		}
		resolutions[resolution.ReferenceKey] = resolution
	}
	required := realtimeVoiceInvestigationReferenceKeys(canonicalIntent)
	if len(resolutions) != len(required) {
		return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
	}
	ordered := make([]agentmodel.Resolution, 0, len(required))
	for _, reference := range required {
		resolution, exists := resolutions[reference]
		if !exists {
			return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
		}
		mention := realtimeVoiceInvestigationReferenceMention(canonicalIntent, reference)
		resolution = realtimeVoiceExactTitleResolution(mention, resolution, byReference[reference])
		ordered = append(ordered, resolution)
	}
	if canonicalIntent.Operation == agentmodel.OperationMove || canonicalIntent.Operation == agentmodel.OperationCreate {
		ordered = canonicalRealtimeVoiceDestinationChain(ordered, allCandidates, len(canonicalIntent.DestinationPath))
	}
	step.Intent = canonicalIntent
	step.Resolutions = ordered
	if step.Validate() != nil {
		return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
	}
	return step, nil
}

func sameRealtimeVoiceInvestigationIntent(left, right agentmodel.Intent) bool {
	if left.Kind != right.Kind || left.Operation != right.Operation || strings.TrimSpace(left.SubjectMention) != strings.TrimSpace(right.SubjectMention) ||
		strings.TrimSpace(left.NewAssetKind) != strings.TrimSpace(right.NewAssetKind) || strings.TrimSpace(left.Details) != strings.TrimSpace(right.Details) ||
		len(left.DestinationPath) != len(right.DestinationPath) {
		return false
	}
	for index := range left.DestinationPath {
		if strings.TrimSpace(left.DestinationPath[index]) != strings.TrimSpace(right.DestinationPath[index]) {
			return false
		}
	}
	return true
}

func realtimeVoiceInvestigationReferenceKeys(intent agentmodel.Intent) []agentmodel.SemanticReferenceKey {
	keys := []agentmodel.SemanticReferenceKey{agentmodel.SemanticReferenceSubject}
	if intent.Operation != agentmodel.OperationCreate && intent.Operation != agentmodel.OperationMove {
		return keys
	}
	for index := range intent.DestinationPath {
		key, _ := agentmodel.NewSemanticReferenceKey(fmt.Sprintf("destination.%d", index))
		keys = append(keys, key)
	}
	return keys
}

func realtimeVoiceInvestigationReferenceMention(intent agentmodel.Intent, reference agentmodel.SemanticReferenceKey) string {
	if reference == agentmodel.SemanticReferenceSubject {
		return intent.SubjectMention
	}
	var index int
	if _, err := fmt.Sscanf(reference.String(), "destination.%d", &index); err == nil && index >= 0 && index < len(intent.DestinationPath) {
		return intent.DestinationPath[index]
	}
	return ""
}

func realtimeVoiceExactTitleResolution(mention string, resolution agentmodel.Resolution, candidates map[string]agentmodel.CandidateObservation) agentmodel.Resolution {
	normalizedMention := normalizeRealtimeVoiceInvestigationTitle(mention)
	if normalizedMention == "" {
		return resolution
	}
	exact := []string{}
	for id, candidate := range candidates {
		if normalizeRealtimeVoiceInvestigationTitle(candidate.Title) == normalizedMention {
			exact = append(exact, id)
		}
	}
	if len(exact) != 1 {
		return resolution
	}
	return agentmodel.Resolution{
		ReferenceKey: resolution.ReferenceKey,
		Status:       agentmodel.ResolutionStrong,
		CandidateIDs: exact,
		Evidence:     "The application selected the sole authorized exact normalized title match for this reference.",
	}
}

func normalizeRealtimeVoiceInvestigationTitle(value string) string {
	words := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })
	return strings.Join(words, " ")
}

func canonicalRealtimeVoiceDestinationChain(resolutions []agentmodel.Resolution, candidates map[string]agentmodel.CandidateObservation, destinationCount int) []agentmodel.Resolution {
	missing := false
	parentID := ""
	for index := 0; index < destinationCount; index++ {
		resolutionIndex := index + 1
		if resolutionIndex >= len(resolutions) {
			break
		}
		resolution := resolutions[resolutionIndex]
		if missing || resolution.Status == agentmodel.ResolutionMissing {
			missing = true
			resolutions[resolutionIndex] = missingRealtimeVoiceDestinationResolution(resolution.ReferenceKey)
			continue
		}
		if (resolution.Status != agentmodel.ResolutionStrong && resolution.Status != agentmodel.ResolutionPlausible) || len(resolution.CandidateIDs) != 1 {
			continue
		}
		candidate, exists := candidates[resolution.CandidateIDs[0]]
		validKind := candidate.Kind == "location" || candidate.Kind == "container"
		validParent := index == 0 || candidate.ParentAssetID == parentID
		if !exists || !validKind || !validParent {
			missing = true
			resolutions[resolutionIndex] = missingRealtimeVoiceDestinationResolution(resolution.ReferenceKey)
			continue
		}
		parentID = candidate.CandidateID
	}
	return resolutions
}

func missingRealtimeVoiceDestinationResolution(reference agentmodel.SemanticReferenceKey) agentmodel.Resolution {
	return agentmodel.Resolution{ReferenceKey: reference, Status: agentmodel.ResolutionMissing, Evidence: "No authorized candidate forms the requested destination chain."}
}

func realtimeVoiceInvestigationResponse(intent agentmodel.Intent, resolutions []agentmodel.Resolution, candidates map[string]agentmodel.CandidateObservation) (ports.StructuredAgentResponse, error) {
	if intent.Kind == agentmodel.IntentKindUnsupported || hasRealtimeVoiceInvestigationStatus(resolutions, agentmodel.ResolutionUnsupported) {
		return investigationResponse(ports.StructuredAgentResponseKindUnsupportedAction, "I can't safely handle that request with inventory voice actions."), nil
	}
	for _, resolution := range resolutions {
		if resolution.Status == agentmodel.ResolutionAmbiguous {
			choices := []string{}
			for _, id := range resolution.CandidateIDs {
				candidate, exists := candidates[id]
				if !exists {
					return ports.StructuredAgentResponse{}, ports.ErrInvalidProviderInput
				}
				choice := candidate.Title
				if len(candidate.ContainmentPath) > 1 {
					choice += " at " + strings.Join(candidate.ContainmentPath[:len(candidate.ContainmentPath)-1], " / ")
				}
				choices = append(choices, choice)
			}
			return investigationResponse(ports.StructuredAgentResponseKindClarification, "I found multiple plausible matches: "+strings.Join(choices, "; ")+". Which one did you mean?"), nil
		}
	}
	subject, exists := realtimeVoiceInvestigationResolution(resolutions, agentmodel.SemanticReferenceSubject)
	if !exists {
		return ports.StructuredAgentResponse{}, ports.ErrInvalidProviderInput
	}
	if intent.Kind == agentmodel.IntentKindChange {
		if subject.Status == agentmodel.ResolutionAbsent {
			return investigationResponse(ports.StructuredAgentResponseKindClarification, "I couldn't find the existing item you want to change. Please describe it another way."), nil
		}
		return ports.StructuredAgentResponse{}, errors.New("change intent requires action-plan compilation")
	}
	if subject.Status == agentmodel.ResolutionAbsent {
		return investigationResponse(ports.StructuredAgentResponseKindAnswer, "I couldn't find a visible match in this inventory."), nil
	}
	if subject.Status == agentmodel.ResolutionCollection {
		titles := realtimeVoiceInvestigationCandidateTitles(subject.CandidateIDs, candidates)
		if len(titles) == 0 {
			return investigationResponse(ports.StructuredAgentResponseKindAnswer, "I didn't find any visible items matching that request."), nil
		}
		return investigationResponse(ports.StructuredAgentResponseKindAnswer, fmt.Sprintf("I found %d visible matches: %s.", len(titles), strings.Join(titles, ", "))), nil
	}
	if len(subject.CandidateIDs) != 1 {
		return ports.StructuredAgentResponse{}, ports.ErrInvalidProviderInput
	}
	candidate, exists := candidates[subject.CandidateIDs[0]]
	if !exists {
		return ports.StructuredAgentResponse{}, ports.ErrInvalidProviderInput
	}
	var message string
	switch intent.Operation {
	case agentmodel.OperationExists:
		message = "Yes, I found " + candidate.Title + "."
	case agentmodel.OperationCheckoutStatus:
		if candidate.CheckoutState == "checked_out" {
			message = candidate.Title + " is currently checked out."
		} else {
			message = candidate.Title + " is currently available."
		}
	case agentmodel.OperationDetail:
		message = candidate.Title + " is recorded as a " + candidate.Kind + "."
	case agentmodel.OperationAssetHistory, agentmodel.OperationCheckoutHistory:
		if len(candidate.Facts) == 0 {
			message = "I found no recorded history for " + candidate.Title + "."
		} else {
			message = candidate.Title + ": " + candidate.Facts[0] + "."
		}
	default:
		prefix := ""
		if subject.Status == agentmodel.ResolutionPlausible {
			prefix = "I think you mean "
		}
		path := strings.Join(candidate.ContainmentPath, " / ")
		if path == "" {
			path = candidate.Title
		}
		message = prefix + candidate.Title + ". Its recorded path is " + path + "."
	}
	return investigationResponse(ports.StructuredAgentResponseKindAnswer, message), nil
}

func investigationResponse(kind ports.StructuredAgentResponseKind, message string) ports.StructuredAgentResponse {
	return ports.StructuredAgentResponse{Kind: kind, SpokenResponse: message, DisplayResponse: message}
}

func realtimeVoiceInvestigationResolution(resolutions []agentmodel.Resolution, reference agentmodel.SemanticReferenceKey) (agentmodel.Resolution, bool) {
	for _, resolution := range resolutions {
		if resolution.ReferenceKey == reference {
			return resolution, true
		}
	}
	return agentmodel.Resolution{}, false
}

func hasRealtimeVoiceInvestigationStatus(resolutions []agentmodel.Resolution, status agentmodel.ResolutionStatus) bool {
	for _, resolution := range resolutions {
		if resolution.Status == status {
			return true
		}
	}
	return false
}

func realtimeVoiceInvestigationCandidateTitles(ids []string, candidates map[string]agentmodel.CandidateObservation) []string {
	titles := []string{}
	for _, id := range ids {
		if candidate, exists := candidates[id]; exists {
			titles = append(titles, candidate.Title)
		}
	}
	sort.Strings(titles)
	return titles
}
