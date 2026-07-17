package app

import (
	"context"
	"strings"
	"unicode"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) generateRealtimeVoiceResponse(ctx context.Context, session RealtimeVoiceSession, brief agentmodel.GroundedVoiceResponseBrief, bindings []ports.StructuredAgentResponseArtifact) (ports.StructuredAgentResponse, error) {
	if brief.Validate() != nil || session.responseGenerator == nil {
		return ports.StructuredAgentResponse{}, ports.ErrInvalidProviderInput
	}
	result, err := session.responseGenerator.GenerateResponse(ctx, ports.VoiceResponseGenerationInput{
		TenantID: session.TenantID, InventoryID: session.InventoryID, Principal: session.Principal, Brief: brief,
	})
	if err != nil {
		return ports.StructuredAgentResponse{}, realtimeVoiceProviderStageError{code: realtimeVoiceFailureLanguageInference, err: err}
	}
	if err := validateRealtimeVoiceGeneratedResponse(brief, result); err != nil {
		return ports.StructuredAgentResponse{}, realtimeVoiceProviderStageError{code: realtimeVoiceFailureLanguageInference, err: err}
	}
	response := ports.StructuredAgentResponse{
		Kind:            realtimeVoiceStructuredResponseKind(brief.Kind),
		SpokenResponse:  strings.TrimSpace(result.SpokenResponse),
		DisplayResponse: strings.TrimSpace(result.DisplayResponse),
		Artifacts:       realtimeVoiceDisplayedResponseArtifacts(result.DisplayResponse, bindings),
	}
	if err := validateRealtimeVoiceFinalResponse(response); err != nil {
		return ports.StructuredAgentResponse{}, realtimeVoiceProviderStageError{code: realtimeVoiceFailureLanguageInference, err: err}
	}
	return response, nil
}

func validateRealtimeVoiceGeneratedResponse(brief agentmodel.GroundedVoiceResponseBrief, result ports.VoiceResponseGenerationResult) error {
	if brief.Validate() != nil {
		return ports.ErrInvalidProviderInput
	}
	response := ports.StructuredAgentResponse{
		Kind: realtimeVoiceStructuredResponseKind(brief.Kind), SpokenResponse: result.SpokenResponse, DisplayResponse: result.DisplayResponse,
	}
	if validateRealtimeVoiceFinalResponse(response) != nil {
		return ports.ErrInvalidProviderInput
	}
	if validateRealtimeVoiceGeneratedChannel(brief, result.SpokenResponse) != nil || validateRealtimeVoiceGeneratedChannel(brief, result.DisplayResponse) != nil {
		return ports.ErrInvalidProviderInput
	}
	for _, title := range realtimeVoiceRequiredDisplayEntityTitles(brief) {
		if !containsExactRealtimeVoiceEntityTitle(result.DisplayResponse, title) {
			return ports.ErrInvalidProviderInput
		}
	}
	return nil
}

func realtimeVoiceRequiredDisplayEntityTitles(brief agentmodel.GroundedVoiceResponseBrief) []string {
	findings := brief.Findings
	if brief.Mode == agentmodel.ResponseAnswerModeInventory {
		findings = realtimeVoiceInventoryResponseFindings(brief.Findings)
	}
	switch brief.Mode {
	case agentmodel.ResponseAnswerModeNotFound, agentmodel.ResponseAnswerModeUnsupported:
		return nil
	}
	titles := make([]string, 0, len(findings)*2)
	for _, finding := range findings {
		titles = append(titles, finding.Title)
		if brief.Mode == agentmodel.ResponseAnswerModeLocate && len(finding.ContainmentPath) > 1 &&
			(finding.Kind == "item" || brief.Confidence == agentmodel.ResponseConfidenceStrong) {
			titles = append(titles, finding.ContainmentPath[len(finding.ContainmentPath)-2])
		}
	}
	return titles
}

func validateRealtimeVoiceGeneratedChannel(brief agentmodel.GroundedVoiceResponseBrief, value string) error {
	text := strings.ToLower(strings.TrimSpace(value))
	for _, forbidden := range []string{"visible match", "candidate", "resolution", "tool result", "tool call", "asset id", "inventory id", "tenant id"} {
		if strings.Contains(text, forbidden) {
			return ports.ErrInvalidProviderInput
		}
	}
	if brief.Kind == agentmodel.ResponseBriefKindClarification {
		if !strings.Contains(text, "?") {
			return ports.ErrInvalidProviderInput
		}
	} else if strings.Contains(text, "?") {
		return ports.ErrInvalidProviderInput
	}
	if brief.Confidence == agentmodel.ResponseConfidencePlausible && !containsRealtimeVoiceUncertainty(text) {
		return ports.ErrInvalidProviderInput
	}
	if (brief.Truncated || realtimeVoiceFactsTruncated(brief.Findings)) && !containsRealtimeVoiceTruncationDisclosure(text) {
		return ports.ErrInvalidProviderInput
	}
	switch brief.Mode {
	case agentmodel.ResponseAnswerModeLocate:
		if containsRealtimeVoiceNoLocation(text) && containsRealtimeVoiceFindingTitle(text, brief.Findings) {
			break
		}
		if !containsRealtimeVoiceLocationRelation(text) || !containsRealtimeVoiceLocationAnchors(text, brief.Findings, brief.Confidence) ||
			(!containsRealtimeVoiceFindingTitle(text, brief.Findings) && !containsNormalizedRealtimeVoiceText(text, brief.Subject)) {
			return ports.ErrInvalidProviderInput
		}
		if len(brief.Findings) > 1 {
			for _, finding := range brief.Findings {
				if !containsRealtimeVoiceFindingReference(text, finding.Title) {
					return ports.ErrInvalidProviderInput
				}
			}
		}
		if brief.Confidence == agentmodel.ResponseConfidencePlausible && len(brief.Findings) == 1 &&
			(brief.Findings[0].Kind == "container" || brief.Findings[0].Kind == "location") &&
			!containsAllRealtimeVoiceSubjectWords(text, brief.Subject) {
			return ports.ErrInvalidProviderInput
		}
	case agentmodel.ResponseAnswerModeInventory:
		required := realtimeVoiceInventoryResponseFindings(brief.Findings)
		for _, finding := range required {
			if !containsRealtimeVoiceFindingReference(text, finding.Title) {
				return ports.ErrInvalidProviderInput
			}
		}
	case agentmodel.ResponseAnswerModeContents:
		for _, finding := range brief.Findings {
			if !containsRealtimeVoiceFindingReference(text, finding.Title) {
				return ports.ErrInvalidProviderInput
			}
		}
	case agentmodel.ResponseAnswerModeExists:
		if !containsRealtimeVoiceEveryFindingTitle(text, brief.Findings) || !containsRealtimeVoiceExistence(text) || containsRealtimeVoiceNegatedClaim(text) {
			return ports.ErrInvalidProviderInput
		}
	case agentmodel.ResponseAnswerModeDetail, agentmodel.ResponseAnswerModeHistory, agentmodel.ResponseAnswerModeCheckout:
		if !containsRealtimeVoiceEveryFindingTitle(text, brief.Findings) || !containsRealtimeVoiceFindingFacts(text, brief.Findings) ||
			containsRealtimeVoiceContradictionForFindings(text, brief.Mode, brief.Findings) {
			return ports.ErrInvalidProviderInput
		}
	case agentmodel.ResponseAnswerModeNotFound:
		if !containsAllRealtimeVoiceSubjectWords(text, brief.Subject) || !containsRealtimeVoiceAbsence(text) || containsRealtimeVoiceInventedLocation(text) {
			return ports.ErrInvalidProviderInput
		}
	case agentmodel.ResponseAnswerModeClarify:
		if brief.Confidence == agentmodel.ResponseConfidenceAbsent {
			if !containsAllRealtimeVoiceSubjectWords(text, brief.Subject) || containsRealtimeVoiceInventedLocation(text) {
				return ports.ErrInvalidProviderInput
			}
			break
		}
		if !containsRealtimeVoiceEveryFindingTitle(text, brief.Findings) {
			return ports.ErrInvalidProviderInput
		}
	case agentmodel.ResponseAnswerModeUnsupported:
		if !containsRealtimeVoiceCapabilityLimit(text) {
			return ports.ErrInvalidProviderInput
		}
	}
	return nil
}

func realtimeVoiceFactsTruncated(findings []agentmodel.ResponseFinding) bool {
	for _, finding := range findings {
		if finding.FactsTruncated {
			return true
		}
	}
	return false
}

func containsRealtimeVoiceTruncationDisclosure(value string) bool {
	for _, term := range []string{" other ", " more ", " additional ", " including ", " among "} {
		if strings.Contains(" "+value+" ", term) {
			return true
		}
	}
	return false
}

func containsRealtimeVoiceNegatedClaim(value string) bool {
	normalized := " " + normalizeRealtimeVoiceInvestigationTitle(value) + " "
	for _, term := range []string{" no ", " not ", " never ", " dont ", " doesnt ", " didnt ", " isnt ", " arent ", " wasnt ", " werent ", " cannot ", " cant ", " couldnt "} {
		if strings.Contains(normalized, term) {
			return true
		}
	}
	return false
}

func containsRealtimeVoiceContradictionForFindings(value string, mode agentmodel.ResponseAnswerMode, findings []agentmodel.ResponseFinding) bool {
	for _, finding := range findings {
		if mode == agentmodel.ResponseAnswerModeCheckout && finding.CheckoutState == "checked_out" {
			for _, term := range []string{"not checked out", "isn't checked out", "isnt checked out", "wasn't checked out", "wasnt checked out", "never checked out", "available"} {
				if strings.Contains(value, term) {
					return true
				}
			}
		}
		if mode == agentmodel.ResponseAnswerModeCheckout && finding.CheckoutState == "available" && strings.Contains(value, "checked out") && !strings.Contains(value, "not checked out") && !strings.Contains(value, "isn't checked out") && !strings.Contains(value, "isnt checked out") {
			return true
		}
		if mode != agentmodel.ResponseAnswerModeHistory && finding.LifecycleState == "archived" && (strings.Contains(value, "not archived") || strings.Contains(value, "isn't archived") || strings.Contains(value, "isnt archived") || strings.Contains(value, " active")) {
			return true
		}
		if mode != agentmodel.ResponseAnswerModeHistory && finding.LifecycleState == "active" && strings.Contains(value, " archived") && !strings.Contains(value, "not archived") && !strings.Contains(value, "isn't archived") && !strings.Contains(value, "isnt archived") {
			return true
		}
		if len(finding.Facts) > 0 && containsRealtimeVoiceNegatedClaim(value) {
			return true
		}
	}
	return false
}

func realtimeVoiceInventoryResponseFindings(findings []agentmodel.ResponseFinding) []agentmodel.ResponseFinding {
	items := make([]agentmodel.ResponseFinding, 0, len(findings))
	for _, finding := range findings {
		if finding.Kind == "item" {
			items = append(items, finding)
		}
	}
	if len(items) > 0 {
		return items
	}
	return findings
}

func containsAllRealtimeVoiceSubjectWords(value string, subject string) bool {
	normalizedValue := normalizeRealtimeVoiceGeneratedText(value)
	words := strings.Fields(normalizeRealtimeVoiceGeneratedText(subject))
	if len(words) == 0 {
		return false
	}
	if len(words) > 1 && (words[0] == "my" || words[0] == "the" || words[0] == "a" || words[0] == "an") {
		words = words[1:]
	}
	for _, word := range words {
		if !strings.Contains(normalizedValue, " "+word+" ") {
			return false
		}
	}
	return true
}

func containsRealtimeVoiceEveryFindingTitle(value string, findings []agentmodel.ResponseFinding) bool {
	for _, finding := range findings {
		if !containsRealtimeVoiceFindingReference(value, finding.Title) {
			return false
		}
	}
	return len(findings) > 0
}

func containsRealtimeVoiceFindingFacts(value string, findings []agentmodel.ResponseFinding) bool {
	for _, finding := range findings {
		for _, fact := range finding.Facts {
			if !containsRealtimeVoiceFactAnchors(value, fact) {
				return false
			}
		}
	}
	return true
}

func containsRealtimeVoiceFactAnchors(value string, fact string) bool {
	words := strings.Fields(normalizeRealtimeVoiceGeneratedText(fact))
	stop := map[string]struct{}{"the": {}, "and": {}, "that": {}, "this": {}, "with": {}, "from": {}, "into": {}, "onto": {}, "was": {}, "were": {}, "has": {}, "had": {}, "for": {}, "are": {}}
	anchors := make([]string, 0, len(words))
	for _, word := range words {
		if _, ignored := stop[word]; ignored || len([]rune(word)) < 3 || strings.IndexFunc(word, unicode.IsDigit) >= 0 {
			continue
		}
		anchors = append(anchors, word)
	}
	if len(anchors) == 0 {
		return false
	}
	normalized := normalizeRealtimeVoiceGeneratedText(value)
	matches := 0
	for _, anchor := range anchors {
		if strings.Contains(normalized, " "+anchor+" ") {
			matches++
		}
	}
	return matches >= (len(anchors)+1)/2
}

func containsRealtimeVoiceExistence(value string) bool {
	for _, term := range []string{" you have ", " you've got ", " there is ", " there are ", " i found ", " is in ", " is at ", " exists "} {
		if strings.Contains(" "+value+" ", term) {
			return true
		}
	}
	return false
}

func containsRealtimeVoiceAbsence(value string) bool {
	for _, term := range []string{"can't find", "cannot find", "couldn't find", "could not find", "didn't find", "did not find", "not found", "no match", "don't see", "do not see"} {
		if strings.Contains(value, term) {
			return true
		}
	}
	return false
}

func containsRealtimeVoiceCapabilityLimit(value string) bool {
	for _, term := range []string{"can't", "cannot", "unable", "not supported", "don't support", "do not support"} {
		if strings.Contains(value, term) {
			return true
		}
	}
	return false
}

func containsRealtimeVoiceInventedLocation(value string) bool {
	for _, scope := range []string{" in this inventory", " in your inventory", " in the inventory"} {
		value = strings.ReplaceAll(value, scope, "")
	}
	return containsRealtimeVoiceLocationRelation(value)
}

func containsRealtimeVoiceNoLocation(value string) bool {
	for _, term := range []string{"not assigned to a location", "isn't assigned to a location", "doesn't have a location", "has no location", "without a location"} {
		if strings.Contains(value, term) {
			return true
		}
	}
	return false
}

func containsRealtimeVoiceFindingTitle(value string, findings []agentmodel.ResponseFinding) bool {
	for _, finding := range findings {
		if containsRealtimeVoiceFindingReference(value, finding.Title) {
			return true
		}
	}
	return false
}

func containsRealtimeVoiceFindingReference(value string, title string) bool {
	if containsNormalizedRealtimeVoiceText(value, title) {
		return true
	}
	if len(title) <= 64 {
		return false
	}
	stop := map[string]struct{}{"the": {}, "and": {}, "for": {}, "from": {}, "with": {}, "into": {}, "onto": {}, "your": {}, "this": {}, "that": {}}
	anchors := make([]string, 0)
	for _, word := range strings.Fields(normalizeRealtimeVoiceGeneratedText(title)) {
		if _, ignored := stop[word]; ignored || len([]rune(word)) < 3 {
			continue
		}
		anchors = append(anchors, word)
	}
	if len(anchors) == 0 {
		return false
	}
	normalized := normalizeRealtimeVoiceGeneratedText(value)
	if !strings.Contains(normalized, " "+anchors[0]+" ") {
		return false
	}
	required := min(3, len(anchors))
	matches := 0
	for _, anchor := range anchors {
		if strings.Contains(normalized, " "+anchor+" ") {
			matches++
		}
	}
	return matches >= required
}

func realtimeVoiceStructuredResponseKind(kind agentmodel.ResponseBriefKind) ports.StructuredAgentResponseKind {
	switch kind {
	case agentmodel.ResponseBriefKindClarification:
		return ports.StructuredAgentResponseKindClarification
	case agentmodel.ResponseBriefKindUnsupported:
		return ports.StructuredAgentResponseKindUnsupportedAction
	default:
		return ports.StructuredAgentResponseKindAnswer
	}
}

func containsRealtimeVoiceUncertainty(value string) bool {
	for _, term := range []string{"probably", "likely", "i think", "may ", "might ", "could ", "appears", "seems", "possible"} {
		if strings.Contains(value, term) {
			return true
		}
	}
	return false
}

func containsRealtimeVoiceLocationRelation(value string) bool {
	for _, term := range []string{" in ", " at ", " inside ", " within ", " under ", " on ", " near ", " by "} {
		if strings.Contains(value, term) {
			return true
		}
	}
	return false
}

func containsRealtimeVoiceLocationAnchors(value string, findings []agentmodel.ResponseFinding, confidence agentmodel.ResponseConfidence) bool {
	for _, finding := range findings {
		path := finding.ContainmentPath
		anchor := finding.Title
		if finding.Kind == "item" || (confidence == agentmodel.ResponseConfidenceStrong && len(path) > 1) {
			if len(path) < 2 {
				continue
			}
			anchor = path[len(path)-2]
		}
		if !containsNormalizedRealtimeVoiceText(value, anchor) {
			return false
		}
	}
	return len(findings) > 0
}

func containsNormalizedRealtimeVoiceText(haystack string, needle string) bool {
	return strings.Contains(normalizeRealtimeVoiceGeneratedText(haystack), normalizeRealtimeVoiceGeneratedText(needle))
}

func normalizeRealtimeVoiceGeneratedText(value string) string {
	words := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	return " " + strings.Join(words, " ") + " "
}
