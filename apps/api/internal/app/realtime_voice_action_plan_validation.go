package app

import (
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
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

func validateRealtimeVoiceMoveRequestDoesNotCreateMissingSource(commands []ActionPlanCommandInput, transcript string, priorResults []ports.AgentToolResult) error {
	source := realtimeVoiceRequestedMoveSource(transcript)
	if source == "" || !realtimeVoiceSourceWasSearchedWithNoMatch(source, priorResults) {
		return nil
	}
	for _, command := range commands {
		if command.Kind != actionplan.CommandKindCreateAsset {
			continue
		}
		kind := strings.TrimSpace(stringArg(command.Arguments["kind"]))
		if kind == "" {
			kind = asset.KindItem.String()
		}
		if kind != asset.KindItem.String() {
			continue
		}
		title := firstStringArg(command.Arguments["title"], command.Arguments["name"])
		if normalizeRealtimeVoiceSourceText(title) == normalizeRealtimeVoiceSourceText(source) {
			return ports.ErrInvalidProviderInput
		}
	}
	return nil
}

func validateRealtimeVoiceRootCreatesUseVisibleParents(commands []ActionPlanCommandInput, transcript string, priorResults []ports.AgentToolResult) error {
	if !realtimeVoiceTranscriptNamesDestination(transcript) {
		return nil
	}
	parents := realtimeVoiceVisibleParentTitlesInTranscript(priorResults, transcript)
	if len(parents) == 0 {
		return nil
	}
	for _, command := range commands {
		if command.Kind != actionplan.CommandKindCreateAsset && command.Kind != actionplan.CommandKindCreateLocation {
			continue
		}
		if strings.TrimSpace(stringArg(command.Arguments["parentAssetId"])) != "" || strings.TrimSpace(stringArg(command.Arguments["parentCommandId"])) != "" {
			continue
		}
		kind := strings.TrimSpace(stringArg(command.Arguments["kind"]))
		if command.Kind == actionplan.CommandKindCreateLocation {
			kind = asset.KindLocation.String()
		}
		if kind == "" {
			kind = asset.KindItem.String()
		}
		if kind != asset.KindContainer.String() {
			continue
		}
		title := normalizeRealtimeVoiceSourceText(firstStringArg(command.Arguments["title"], command.Arguments["name"]))
		for _, parentTitle := range parents {
			if title != normalizeRealtimeVoiceSourceText(parentTitle) {
				return ports.ErrInvalidProviderInput
			}
		}
	}
	return nil
}

func realtimeVoiceSourceWasSearchedWithNoMatch(source string, toolResults []ports.AgentToolResult) bool {
	source = normalizeRealtimeVoiceSourceText(source)
	if source == "" {
		return false
	}
	for _, query := range realtimeVoiceNoMatchQueries(toolResults) {
		query = normalizeRealtimeVoiceSourceText(query)
		if query == source || strings.Contains(source, query) || strings.Contains(query, source) {
			return true
		}
	}
	return false
}

func realtimeVoiceVisibleParentTitlesInTranscript(toolResults []ports.AgentToolResult, transcript string) []string {
	parents := realtimeVoiceVisibleParentsInTranscript(toolResults, transcript)
	titles := []string{}
	for _, parent := range parents {
		titles = append(titles, parent.Title)
	}
	return titles
}

type realtimeVoiceVisibleParent struct {
	AssetID string
	Title   string
}

func realtimeVoiceVisibleParentsInTranscript(toolResults []ports.AgentToolResult, transcript string) []realtimeVoiceVisibleParent {
	parents := []realtimeVoiceVisibleParent{}
	for _, result := range toolResults {
		if result.Name != RealtimeVoiceToolSearchAuthorizedAssets && result.Name != RealtimeVoiceToolListAuthorizedAssets {
			continue
		}
		var output realtimeVoiceAssetToolOutput
		if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
			continue
		}
		for _, item := range output.Items {
			if item.Kind != asset.KindLocation.String() && item.Kind != asset.KindContainer.String() {
				continue
			}
			if realtimeVoiceTitleMentionedInTranscript(item.Title, transcript) {
				parents = append(parents, realtimeVoiceVisibleParent{AssetID: item.AssetID, Title: item.Title})
			}
		}
	}
	return parents
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
