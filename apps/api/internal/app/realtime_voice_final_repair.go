package app

import (
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeVoiceShouldRepairCreateClarification(transcript string, response ports.StructuredAgentResponse, toolResults []ports.AgentToolResult) bool {
	if len(toolResults) == 0 {
		return false
	}
	if !realtimeVoiceToolResultsContainRequestedSource(transcript, toolResults) {
		return false
	}
	kind := response.Kind
	if kind == "" {
		kind = ports.StructuredAgentResponseKindAnswer
	}
	if kind != ports.StructuredAgentResponseKindClarification && kind != ports.StructuredAgentResponseKindAnswer {
		return false
	}
	if realtimeVoiceAlreadyRejectedCreateClarification(toolResults) {
		return false
	}
	if !realtimeVoiceLooksLikeWriteRequest(transcript) {
		return false
	}
	text := strings.ToLower(response.SpokenResponse + " " + response.DisplayResponse)
	if !strings.Contains(text, "create") && !strings.Contains(text, "add") {
		return false
	}
	if !strings.Contains(text, "do you want") && !strings.Contains(text, "would you like") && !strings.Contains(text, "should i") && !strings.Contains(text, "shall i") {
		return false
	}
	return strings.Contains(text, "can't find") ||
		strings.Contains(text, "cannot find") ||
		strings.Contains(text, "couldn't find") ||
		strings.Contains(text, "could not find") ||
		strings.Contains(text, "not find")
}

func realtimeVoiceShouldRepairWriteClaimAfterFailedProposal(transcript string, response ports.StructuredAgentResponse, toolResults []ports.AgentToolResult) bool {
	if !realtimeVoiceLooksLikeWriteRequest(transcript) || !realtimeVoiceHasRejectedActionPlanProposal(toolResults) || realtimeVoiceAlreadyRejectedWriteClaim(toolResults) {
		return false
	}
	kind := response.Kind
	if kind == "" {
		kind = ports.StructuredAgentResponseKindAnswer
	}
	if kind != ports.StructuredAgentResponseKindAnswer {
		return false
	}
	return true
}

func realtimeVoiceMissingMoveSourceResponse(transcript string, toolResults []ports.AgentToolResult) (ports.StructuredAgentResponse, bool) {
	if !realtimeVoiceLooksLikeMoveRequest(transcript) {
		return ports.StructuredAgentResponse{}, false
	}
	source := realtimeVoiceRequestedMoveSource(transcript)
	if source == "" || source == "it" || source == "them" {
		return ports.StructuredAgentResponse{}, false
	}
	if !realtimeVoiceSourceWasSearchedWithNoMatch(source, toolResults) || realtimeVoiceToolResultsContainVisibleAssetTitle(source, toolResults) {
		return ports.StructuredAgentResponse{}, false
	}
	message := "I could not find " + realtimeVoiceArticleForSpokenSource(source) + source + " to move. Add it first or tell me a different item."
	return ports.StructuredAgentResponse{
		Kind:            ports.StructuredAgentResponseKindClarification,
		SpokenResponse:  message,
		DisplayResponse: message,
	}, true
}

func realtimeVoiceArticleForSpokenSource(source string) string {
	first := ""
	for _, char := range strings.TrimSpace(source) {
		first = strings.ToLower(string(char))
		break
	}
	switch first {
	case "a", "e", "i", "o", "u":
		return "an "
	default:
		return "a "
	}
}

func realtimeVoiceHasRejectedActionPlanProposal(toolResults []ports.AgentToolResult) bool {
	for _, result := range toolResults {
		if result.Name == RealtimeVoiceToolProposeActionPlan && strings.Contains(result.Content, `"status":"error"`) {
			return true
		}
	}
	return false
}

func realtimeVoiceToolResultsContainRequestedSource(transcript string, toolResults []ports.AgentToolResult) bool {
	source := realtimeVoiceRequestedMoveSource(transcript)
	if source == "" {
		return realtimeVoiceToolResultsContainVisibleAssetTitle("", toolResults)
	}
	return realtimeVoiceToolResultsContainVisibleAssetTitle(source, toolResults)
}

func realtimeVoiceToolResultsContainVisibleAssetTitle(source string, toolResults []ports.AgentToolResult) bool {
	source = normalizeRealtimeVoiceSourceText(source)
	for _, result := range toolResults {
		if result.Name != RealtimeVoiceToolSearchAuthorizedAssets && result.Name != RealtimeVoiceToolListAuthorizedAssets {
			continue
		}
		var output realtimeVoiceAssetToolOutput
		if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
			continue
		}
		for _, item := range output.Items {
			title := normalizeRealtimeVoiceSourceText(item.Title)
			if title == "" {
				continue
			}
			if source == "" || title == source || strings.Contains(source, title) || strings.Contains(title, source) {
				return true
			}
		}
	}
	return false
}

func realtimeVoiceRequestedMoveSource(transcript string) string {
	text := normalizedRealtimeVoiceVerbText(transcript)
	verbStart := -1
	for _, verb := range []string{" move ", " put ", " place ", " store ", " stash "} {
		if index := strings.Index(text, verb); index >= 0 && (verbStart == -1 || index < verbStart) {
			verbStart = index + len(verb)
		}
	}
	if verbStart == -1 || verbStart >= len(text) {
		return ""
	}
	rest := text[verbStart:]
	end := len(rest)
	for _, marker := range []string{" to ", " into ", " inside ", " in ", " on ", " onto "} {
		if index := strings.Index(rest, marker); index >= 0 && index < end {
			end = index
		}
	}
	return normalizeRealtimeVoiceSourceText(rest[:end])
}

func normalizeRealtimeVoiceSourceText(text string) string {
	words := strings.Fields(strings.ToLower(text))
	kept := make([]string, 0, len(words))
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:\"'()[]{}")
		switch word {
		case "", "my", "the", "a", "an", "this", "that", "these", "those", "item", "items":
			continue
		default:
			kept = append(kept, word)
		}
	}
	return strings.Join(kept, " ")
}

func realtimeVoiceAlreadyRejectedCreateClarification(toolResults []ports.AgentToolResult) bool {
	for _, result := range toolResults {
		if strings.Contains(result.Content, "final_clarification_rejected") {
			return true
		}
	}
	return false
}

func realtimeVoiceAlreadyRejectedWriteClaim(toolResults []ports.AgentToolResult) bool {
	for _, result := range toolResults {
		if strings.Contains(result.Content, "final_write_claim_rejected") {
			return true
		}
	}
	return false
}

func realtimeVoiceLooksLikeWriteRequest(transcript string) bool {
	if realtimeVoiceLooksLikeReadQuestion(transcript) {
		return false
	}
	text := normalizedRealtimeVoiceVerbText(transcript)
	if realtimeVoiceLooksLikeCasualAcquisitionCreateRequest(text) {
		return true
	}
	for _, token := range []string{
		" move ", " put ", " place ", " add ", " create ", " store ", " stash ", " restore ", " archive ", " rename ", " update ",
		" return ", " returned ", " check out ", " checked out ", " check in ", " checked in ", " borrow ", " borrowed ", " loan ", " loaned ",
		" bring back ", " brought back ", " put back ", " back in ", " is back ",
	} {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func realtimeVoiceLooksLikeReadQuestion(transcript string) bool {
	text := normalizedRealtimeVoiceVerbText(transcript)
	for _, prefix := range []string{" where ", " what ", " when ", " who ", " which ", " do i ", " do we ", " did i ", " did we ", " can you find ", " find my ", " find the "} {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return strings.Contains(text, " where did ") ||
		strings.Contains(text, " where is ") ||
		strings.Contains(text, " where are ") ||
		strings.Contains(text, " what is ") ||
		strings.Contains(text, " what are ")
}

func realtimeVoiceLooksLikeMoveRequest(transcript string) bool {
	if realtimeVoiceLooksLikeReadQuestion(transcript) {
		return false
	}
	text := normalizedRealtimeVoiceVerbText(transcript)
	for _, token := range []string{" move ", " put ", " place ", " store ", " stash "} {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func normalizedRealtimeVoiceVerbText(transcript string) string {
	text := strings.ToLower(transcript)
	for _, replacer := range []string{".", ",", "!", "?", ";", ":", "\"", "'", "(", ")", "[", "]", "{", "}"} {
		text = strings.ReplaceAll(text, replacer, " ")
	}
	return " " + strings.Join(strings.Fields(text), " ") + " "
}

func realtimeVoiceFinalClarificationRepairResult(id string) (ports.AgentToolResult, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		id = "final-clarification-repair"
	}
	return realtimeVoiceToolErrorResult(ports.AgentToolCall{
		ID:   id,
		Name: RealtimeVoiceToolProposeActionPlan,
	}, "final_clarification_rejected", "A clear write request with a missing named destination must be turned into a reviewable action plan instead of a final yes/no creation question. Call propose_action_plan with ordered create commands for every missing parent location or container, then the requested move or create command. Use parentCommandId for dependencies such as Kitchen -> Big cabinet -> Second shelf -> move item.", true)
}

func realtimeVoiceFinalWriteClaimRepairResult(id string) (ports.AgentToolResult, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		id = "final-write-claim-repair"
	}
	return realtimeVoiceToolErrorResult(ports.AgentToolCall{
		ID:   id,
		Name: RealtimeVoiceToolProposeActionPlan,
	}, "final_write_claim_rejected", "No inventory change has been applied. A write request is complete only after propose_action_plan succeeds and the user approves the plan. Retry propose_action_plan with valid structured commands, or produce a safe clarification that tells the user what is needed next. For a new item inside a missing container under an existing parent, create the container first with parentAssetId, then create the item with parentCommandId; do not move a newly-created item by invented assetId.", true)
}
