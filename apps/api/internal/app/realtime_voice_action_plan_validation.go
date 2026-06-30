package app

import (
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func validateRealtimeVoiceMissingDestinationSegmentsAccountedFor(commands []ActionPlanCommandInput, transcript string, priorResults []ports.AgentToolResult) error {
	represented := strings.Builder{}
	for _, command := range commands {
		if command.Kind != actionplan.CommandKindCreateAsset && command.Kind != actionplan.CommandKindCreateLocation {
			continue
		}
		represented.WriteString(" ")
		represented.WriteString(firstStringArg(command.Arguments["title"], command.Arguments["name"]))
	}
	representedText := represented.String()
	for _, query := range realtimeVoiceNoMatchQueries(priorResults) {
		if !realtimeVoiceQueryLooksLikeDestinationSegment(query, transcript) {
			continue
		}
		if !realtimeVoiceMeaningfulWordsRepresented(query, representedText) {
			return ports.ErrInvalidProviderInput
		}
	}
	return nil
}

func validateRealtimeVoiceMoveRequestUsesVisibleSource(commands []ActionPlanCommandInput, transcript string, priorResults []ports.AgentToolResult) error {
	source := realtimeVoiceRequestedMoveSource(transcript)
	if source == "" {
		return nil
	}
	sourceID := realtimeVoiceVisibleAssetIDForTitle(source, priorResults)
	if sourceID == "" {
		return nil
	}
	for _, command := range commands {
		if command.Kind != actionplan.CommandKindMoveAsset {
			continue
		}
		if strings.TrimSpace(stringArg(command.Arguments["assetId"])) == sourceID {
			return nil
		}
	}
	return ports.ErrInvalidProviderInput
}

func realtimeVoiceVisibleAssetIDForTitle(source string, toolResults []ports.AgentToolResult) string {
	source = normalizeRealtimeVoiceSourceText(source)
	if source == "" {
		return ""
	}
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
			if title == source || strings.Contains(source, title) || strings.Contains(title, source) {
				return item.AssetID
			}
		}
	}
	return ""
}

func validateRealtimeVoiceMissingDestinationHierarchy(commands []ActionPlanCommandInput, transcript string, priorResults []ports.AgentToolResult) error {
	pathWords := realtimeVoiceMissingDestinationPathWords(transcript, priorResults)
	if len(pathWords) < 2 {
		return nil
	}
	commandIDByWord := map[string]string{}
	parentCommandIDByWord := map[string]string{}
	for _, command := range commands {
		if command.Kind != actionplan.CommandKindCreateAsset && command.Kind != actionplan.CommandKindCreateLocation {
			continue
		}
		title := strings.ToLower(firstStringArg(command.Arguments["title"], command.Arguments["name"]))
		for _, word := range realtimeVoiceMeaningfulWords(title) {
			if _, exists := commandIDByWord[word]; !exists {
				commandIDByWord[word] = command.ID
				parentCommandIDByWord[word] = strings.TrimSpace(stringArg(command.Arguments["parentCommandId"]))
			}
		}
	}
	for index := 0; index < len(pathWords)-1; index++ {
		innerID := commandIDByWord[pathWords[index]]
		outerID := commandIDByWord[pathWords[index+1]]
		if innerID == "" || outerID == "" {
			continue
		}
		if parentCommandIDByWord[pathWords[index]] != outerID {
			return ports.ErrInvalidProviderInput
		}
	}
	return nil
}

func realtimeVoiceMissingDestinationPathWords(transcript string, priorResults []ports.AgentToolResult) []string {
	noMatchWords := map[string]struct{}{}
	for _, query := range realtimeVoiceNoMatchQueries(priorResults) {
		if !realtimeVoiceQueryLooksLikeDestinationSegment(query, transcript) {
			continue
		}
		for _, word := range realtimeVoiceMeaningfulWords(query) {
			if realtimeVoiceDestinationSegmentWords[word] {
				noMatchWords[word] = struct{}{}
			}
		}
	}
	transcript = strings.ToLower(transcript)
	type positionedWord struct {
		word  string
		index int
	}
	positioned := []positionedWord{}
	for word := range noMatchWords {
		index := strings.Index(transcript, word)
		if index >= 0 {
			positioned = append(positioned, positionedWord{word: word, index: index})
		}
	}
	for i := 1; i < len(positioned); i++ {
		current := positioned[i]
		j := i - 1
		for ; j >= 0 && positioned[j].index > current.index; j-- {
			positioned[j+1] = positioned[j]
		}
		positioned[j+1] = current
	}
	words := make([]string, 0, len(positioned))
	for _, item := range positioned {
		words = append(words, item.word)
	}
	return words
}
