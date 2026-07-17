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

func canonicalRealtimeVoiceInvestigationStep(canonicalIntent agentmodel.Intent, step agentmodel.InvestigationStep, observations []agentmodel.CandidateObservation, readEvidence []agentmodel.ReadEvidence) (agentmodel.InvestigationStep, error) {
	if step.Decision != agentmodel.InvestigationDecisionFinish || step.Validate() != nil || !sameRealtimeVoiceInvestigationIntent(canonicalIntent, step.Intent) {
		return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
	}
	step.Resolutions = canonicalRealtimeVoiceNoCandidateStatuses(canonicalIntent, step.Resolutions)
	byReference := map[agentmodel.SemanticReferenceKey]map[string]agentmodel.CandidateObservation{}
	allCandidates := map[string]agentmodel.CandidateObservation{}
	coverage := map[agentmodel.SemanticReferenceKey]bool{}
	discoveryCoverage := map[agentmodel.SemanticReferenceKey]bool{}
	discoveryScopes := map[agentmodel.SemanticReferenceKey][]agentmodel.LifecycleScope{}
	for _, evidence := range readEvidence {
		if evidence.Validate() != nil {
			return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
		}
		coverage[evidence.ReferenceKey] = true
		switch evidence.ReadKind {
		case agentmodel.InvestigationReadSearchAssets, agentmodel.InvestigationReadListInventory, agentmodel.InvestigationReadListContents:
			discoveryCoverage[evidence.ReferenceKey] = true
			discoveryScopes[evidence.ReferenceKey] = append(discoveryScopes[evidence.ReferenceKey], evidence.LifecycleScope.Effective())
		}
	}
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
		if resolution.Status != agentmodel.ResolutionUnsupported && !coverage[reference] {
			return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
		}
		if (resolution.Status == agentmodel.ResolutionAbsent || resolution.Status == agentmodel.ResolutionMissing) && !discoveryCoverage[reference] {
			return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
		}
		mention := realtimeVoiceInvestigationReferenceMention(canonicalIntent, reference)
		resolution = realtimeVoiceExactTitleResolution(mention, resolution, byReference[reference])
		resolution = realtimeVoiceCalibratedStrongResolution(mention, resolution, byReference[reference])
		for _, candidateID := range resolution.CandidateIDs {
			candidate := byReference[reference][candidateID]
			if !realtimeVoiceLifecycleScopeIncludes(discoveryScopes[reference], candidate.LifecycleState) {
				return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
			}
		}
		ordered = append(ordered, resolution)
	}
	if canonicalIntent.Operation == agentmodel.OperationMove || canonicalIntent.Operation == agentmodel.OperationCreate {
		ordered = canonicalRealtimeVoiceDestinationChain(ordered, allCandidates, len(canonicalIntent.DestinationPath))
	}
	if canonicalIntent.Operation == agentmodel.OperationListContents && len(ordered) == 1 {
		ordered[0] = canonicalRealtimeVoiceContentsResolution(ordered[0], observations, readEvidence)
	}
	step.Intent = canonicalIntent
	step.Resolutions = ordered
	if step.Validate() != nil {
		return agentmodel.InvestigationStep{}, ports.ErrInvalidProviderInput
	}
	return step, nil
}

func realtimeVoiceCalibratedStrongResolution(mention string, resolution agentmodel.Resolution, candidates map[string]agentmodel.CandidateObservation) agentmodel.Resolution {
	if resolution.Status != agentmodel.ResolutionStrong || len(resolution.CandidateIDs) != 1 {
		return resolution
	}
	candidate, found := candidates[resolution.CandidateIDs[0]]
	if !found || realtimeVoiceInvestigationTitleMatchesMention(candidate.Title, mention) {
		return resolution
	}
	resolution.Status = agentmodel.ResolutionPlausible
	resolution.Evidence = "The application calibrated the sole authorized non-exact title match as plausible."
	return resolution
}

func canonicalRealtimeVoiceContentsResolution(resolution agentmodel.Resolution, observations []agentmodel.CandidateObservation, evidence []agentmodel.ReadEvidence) agentmodel.Resolution {
	targetID := ""
	for _, record := range evidence {
		if record.ReferenceKey == agentmodel.SemanticReferenceSubject && record.ReadKind == agentmodel.InvestigationReadListContents && strings.TrimSpace(record.VisibleAssetID) != "" {
			targetID = record.VisibleAssetID
		}
	}
	if targetID == "" {
		return resolution
	}
	contents := make([]agentmodel.CandidateObservation, 0)
	for _, observation := range observations {
		if observation.ReferenceKey == agentmodel.SemanticReferenceSubject && observation.ParentAssetID == targetID {
			contents = append(contents, observation)
		}
	}
	sort.Slice(contents, func(i, j int) bool {
		if contents[i].Title == contents[j].Title {
			return contents[i].CandidateID < contents[j].CandidateID
		}
		return contents[i].Title < contents[j].Title
	})
	ids := make([]string, 0, len(contents))
	for _, item := range contents {
		ids = append(ids, item.CandidateID)
	}
	return agentmodel.Resolution{
		ReferenceKey: agentmodel.SemanticReferenceSubject, Status: agentmodel.ResolutionCollection,
		CandidateIDs: ids, Evidence: "The application selected the authorized direct contents returned for the resolved subject.",
	}
}

func canonicalRealtimeVoiceNoCandidateStatuses(intent agentmodel.Intent, resolutions []agentmodel.Resolution) []agentmodel.Resolution {
	canonical := append([]agentmodel.Resolution{}, resolutions...)
	for index := range canonical {
		resolution := &canonical[index]
		if resolution.Status != agentmodel.ResolutionAbsent && resolution.Status != agentmodel.ResolutionMissing {
			continue
		}
		if resolution.ReferenceKey == agentmodel.SemanticReferenceSubject {
			if intent.Operation == agentmodel.OperationCreate {
				resolution.Status = agentmodel.ResolutionMissing
			} else {
				resolution.Status = agentmodel.ResolutionAbsent
			}
			continue
		}
		if intent.Operation == agentmodel.OperationCreate || intent.Operation == agentmodel.OperationMove {
			resolution.Status = agentmodel.ResolutionMissing
		}
	}
	return canonical
}

func realtimeVoiceLifecycleScopeIncludes(scopes []agentmodel.LifecycleScope, lifecycle string) bool {
	lifecycle = strings.TrimSpace(lifecycle)
	if lifecycle == "" {
		lifecycle = string(agentmodel.LifecycleScopeActive)
	}
	for _, scope := range scopes {
		effective := scope.Effective()
		if effective == agentmodel.LifecycleScopeAll || string(effective) == lifecycle {
			return true
		}
	}
	return false
}

func sameRealtimeVoiceInvestigationIntent(left, right agentmodel.Intent) bool {
	if left.RequestShape != right.RequestShape || left.Kind != right.Kind || left.Operation != right.Operation || strings.TrimSpace(left.SubjectMention) != strings.TrimSpace(right.SubjectMention) ||
		strings.TrimSpace(left.NewAssetKind) != strings.TrimSpace(right.NewAssetKind) ||
		len(left.DestinationPath) != len(right.DestinationPath) || len(left.DestinationKinds) != len(right.DestinationKinds) {
		return false
	}
	for index := range left.DestinationPath {
		if strings.TrimSpace(left.DestinationPath[index]) != strings.TrimSpace(right.DestinationPath[index]) ||
			left.DestinationKinds[index] != right.DestinationKinds[index] {
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
	if resolution.Status == agentmodel.ResolutionCollection {
		return resolution
	}
	if normalizeRealtimeVoiceInvestigationTitle(mention) == "" {
		return resolution
	}
	exact := []string{}
	for id, candidate := range candidates {
		if realtimeVoiceInvestigationTitleMatchesMention(candidate.Title, mention) {
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

func realtimeVoiceInvestigationTitleMatchesMention(title, mention string) bool {
	normalizedTitle := normalizeRealtimeVoiceInvestigationTitle(title)
	normalizedMention := normalizeRealtimeVoiceInvestigationTitle(mention)
	if normalizedTitle == "" || normalizedMention == "" {
		return false
	}
	if normalizedTitle == normalizedMention {
		return true
	}
	words := strings.Fields(normalizedMention)
	if len(words) < 2 {
		return false
	}
	switch words[0] {
	case "my", "the", "a", "an":
		return normalizedTitle == strings.Join(words[1:], " ")
	default:
		return false
	}
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

func realtimeVoiceInvestigationResponseBrief(intent agentmodel.Intent, resolutions []agentmodel.Resolution, candidates map[string]agentmodel.CandidateObservation) (agentmodel.GroundedVoiceResponseBrief, error) {
	if intent.Kind == agentmodel.IntentKindUnsupported || hasRealtimeVoiceInvestigationStatus(resolutions, agentmodel.ResolutionUnsupported) {
		return validatedRealtimeVoiceResponseBrief(agentmodel.GroundedVoiceResponseBrief{
			Kind: agentmodel.ResponseBriefKindUnsupported, Mode: agentmodel.ResponseAnswerModeUnsupported,
			Operation: agentmodel.OperationUnsupported, Subject: intent.SubjectMention, Confidence: agentmodel.ResponseConfidenceAbsent,
		})
	}
	for _, resolution := range resolutions {
		if resolution.ReferenceKey == agentmodel.SemanticReferenceSubject || resolution.Status != agentmodel.ResolutionPlausible || len(resolution.CandidateIDs) != 1 {
			continue
		}
		candidate, exists := candidates[resolution.CandidateIDs[0]]
		if !exists {
			return agentmodel.GroundedVoiceResponseBrief{}, ports.ErrInvalidProviderInput
		}
		requested := realtimeVoiceInvestigationReferenceMention(intent, resolution.ReferenceKey)
		return validatedRealtimeVoiceResponseBrief(agentmodel.GroundedVoiceResponseBrief{
			Kind: agentmodel.ResponseBriefKindClarification, Mode: agentmodel.ResponseAnswerModeClarify, Operation: intent.Operation,
			Subject: requested, Confidence: agentmodel.ResponseConfidencePlausible,
			Findings: []agentmodel.ResponseFinding{realtimeVoiceResponseFinding("finding.0", candidate)},
		})
	}
	for _, resolution := range resolutions {
		if resolution.Status == agentmodel.ResolutionAmbiguous {
			findings, truncated, err := realtimeVoiceResponseFindings(agentmodel.ResponseAnswerModeClarify, resolution.CandidateIDs, candidates)
			if err != nil {
				return agentmodel.GroundedVoiceResponseBrief{}, err
			}
			return validatedRealtimeVoiceResponseBrief(agentmodel.GroundedVoiceResponseBrief{
				Kind: agentmodel.ResponseBriefKindClarification, Mode: agentmodel.ResponseAnswerModeClarify, Operation: intent.Operation,
				Subject: realtimeVoiceInvestigationReferenceMention(intent, resolution.ReferenceKey), Confidence: agentmodel.ResponseConfidenceAmbiguous, Findings: findings, Truncated: truncated,
			})
		}
	}
	subject, exists := realtimeVoiceInvestigationResolution(resolutions, agentmodel.SemanticReferenceSubject)
	if !exists {
		return agentmodel.GroundedVoiceResponseBrief{}, ports.ErrInvalidProviderInput
	}
	if intent.Kind == agentmodel.IntentKindChange {
		if subject.Status == agentmodel.ResolutionAbsent {
			return validatedRealtimeVoiceResponseBrief(agentmodel.GroundedVoiceResponseBrief{
				Kind: agentmodel.ResponseBriefKindClarification, Mode: agentmodel.ResponseAnswerModeClarify, Operation: intent.Operation,
				Subject: intent.SubjectMention, Confidence: agentmodel.ResponseConfidenceAbsent,
			})
		}
		return agentmodel.GroundedVoiceResponseBrief{}, errors.New("change intent requires action-plan compilation")
	}
	if subject.Status == agentmodel.ResolutionAbsent {
		return validatedRealtimeVoiceResponseBrief(agentmodel.GroundedVoiceResponseBrief{
			Kind: agentmodel.ResponseBriefKindAnswer, Mode: agentmodel.ResponseAnswerModeNotFound, Operation: intent.Operation,
			Subject: intent.SubjectMention, Confidence: agentmodel.ResponseConfidenceAbsent,
		})
	}
	mode := realtimeVoiceResponseAnswerMode(intent.Operation)
	findings, truncated, err := realtimeVoiceResponseFindings(mode, subject.CandidateIDs, candidates)
	if err != nil {
		return agentmodel.GroundedVoiceResponseBrief{}, err
	}
	confidence := agentmodel.ResponseConfidenceStrong
	if subject.Status == agentmodel.ResolutionPlausible {
		confidence = agentmodel.ResponseConfidencePlausible
	}
	if subject.Status == agentmodel.ResolutionCollection && intent.Operation == agentmodel.OperationLocate && len(findings) == 1 && (findings[0].Kind == "container" || findings[0].Kind == "location") {
		confidence = agentmodel.ResponseConfidencePlausible
	}
	return validatedRealtimeVoiceResponseBrief(agentmodel.GroundedVoiceResponseBrief{
		Kind: agentmodel.ResponseBriefKindAnswer, Mode: mode, Operation: intent.Operation,
		Subject: intent.SubjectMention, Confidence: confidence, Findings: findings, Truncated: truncated,
	})
}

func realtimeVoiceResponseAnswerMode(operation agentmodel.Operation) agentmodel.ResponseAnswerMode {
	switch operation {
	case agentmodel.OperationLocate:
		return agentmodel.ResponseAnswerModeLocate
	case agentmodel.OperationListInventory:
		return agentmodel.ResponseAnswerModeInventory
	case agentmodel.OperationListContents:
		return agentmodel.ResponseAnswerModeContents
	case agentmodel.OperationExists:
		return agentmodel.ResponseAnswerModeExists
	case agentmodel.OperationDetail:
		return agentmodel.ResponseAnswerModeDetail
	case agentmodel.OperationAssetHistory, agentmodel.OperationCheckoutHistory:
		return agentmodel.ResponseAnswerModeHistory
	case agentmodel.OperationCheckoutStatus:
		return agentmodel.ResponseAnswerModeCheckout
	default:
		return agentmodel.ResponseAnswerModeDetail
	}
}

func realtimeVoiceResponseFindings(mode agentmodel.ResponseAnswerMode, ids []string, candidates map[string]agentmodel.CandidateObservation) ([]agentmodel.ResponseFinding, bool, error) {
	items := make([]agentmodel.CandidateObservation, 0, len(ids))
	for _, id := range ids {
		candidate, exists := candidates[id]
		if !exists {
			return nil, false, ports.ErrInvalidProviderInput
		}
		items = append(items, candidate)
	}
	sort.Slice(items, func(i, j int) bool {
		if mode == agentmodel.ResponseAnswerModeInventory && (items[i].Kind == "item") != (items[j].Kind == "item") {
			return items[i].Kind == "item"
		}
		if items[i].Title == items[j].Title {
			return items[i].CandidateID < items[j].CandidateID
		}
		return items[i].Title < items[j].Title
	})
	findings := make([]agentmodel.ResponseFinding, 0, min(len(items), 8))
	presentationCost := 0
	for index, candidate := range items {
		finding := realtimeVoiceResponseFinding(fmt.Sprintf("finding.%d", index), candidate)
		cost := len(finding.Title)
		if len(finding.ContainmentPath) > 1 {
			parent := finding.ContainmentPath[len(finding.ContainmentPath)-2]
			if !strings.EqualFold(parent, finding.Title) {
				cost += len(parent)
			}
		}
		if len(findings) > 0 && (len(findings) >= 8 || presentationCost+cost > 320) {
			break
		}
		findings = append(findings, finding)
		presentationCost += cost
	}
	return findings, len(findings) < len(items), nil
}

func realtimeVoiceResponseFinding(key string, candidate agentmodel.CandidateObservation) agentmodel.ResponseFinding {
	path := candidate.ContainmentPath
	if len(path) > 2 {
		path = path[len(path)-2:]
	}
	facts := make([]string, 0, min(len(candidate.Facts), 3))
	factCost := 0
	factsTruncated := false
	firstNonemptyFact := ""
	for _, fact := range candidate.Facts {
		fact = strings.TrimSpace(fact)
		if fact == "" {
			continue
		}
		if firstNonemptyFact == "" {
			firstNonemptyFact = fact
		}
		if len(facts) >= 3 {
			factsTruncated = true
			break
		}
		if factCost+len(fact) > 180 {
			factsTruncated = true
			continue
		}
		facts = append(facts, fact)
		factCost += len(fact)
	}
	if len(facts) == 0 && firstNonemptyFact != "" {
		facts = append(facts, truncateRealtimeVoiceResponseFact(firstNonemptyFact, 180))
		factsTruncated = true
	}
	return agentmodel.ResponseFinding{
		FactKey: key, Title: candidate.Title, Kind: candidate.Kind, LifecycleState: candidate.LifecycleState, CheckoutState: candidate.CheckoutState,
		ContainmentPath: append([]string{}, path...), Facts: facts, FactsTruncated: factsTruncated || len(facts) < len(candidate.Facts),
	}
}

func truncateRealtimeVoiceResponseFact(value string, maxBytes int) string {
	value = strings.TrimSpace(value)
	if len(value) <= maxBytes {
		return value
	}
	end := 0
	for index := range value {
		if index > maxBytes {
			break
		}
		end = index
	}
	if end == 0 {
		return ""
	}
	value = strings.TrimSpace(value[:end])
	if wordBoundary := strings.LastIndexByte(value, ' '); wordBoundary >= maxBytes/2 {
		value = strings.TrimSpace(value[:wordBoundary])
	}
	return value
}

func validatedRealtimeVoiceResponseBrief(brief agentmodel.GroundedVoiceResponseBrief) (agentmodel.GroundedVoiceResponseBrief, error) {
	if brief.Validate() != nil {
		return agentmodel.GroundedVoiceResponseBrief{}, ports.ErrInvalidProviderInput
	}
	return brief, nil
}

func hasRealtimeVoicePlausibleDestination(resolutions []agentmodel.Resolution) bool {
	for _, resolution := range resolutions {
		if resolution.ReferenceKey != agentmodel.SemanticReferenceSubject && resolution.Status == agentmodel.ResolutionPlausible {
			return true
		}
	}
	return false
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
