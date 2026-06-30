package app

import (
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
)

func canonicalRealtimeVoiceTranscriptCreateHierarchy(commands []ActionPlanCommandInput, transcript string) []ActionPlanCommandInput {
	if len(commands) < 2 || strings.TrimSpace(transcript) == "" {
		return commands
	}
	normalized := append([]ActionPlanCommandInput{}, commands...)
	type createCommand struct {
		index int
		id    string
		word  string
		pos   int
	}
	creates := []createCommand{}
	for index, command := range normalized {
		if command.ID == "" || (command.Kind != actionplan.CommandKindCreateAsset && command.Kind != actionplan.CommandKindCreateLocation) {
			continue
		}
		if strings.TrimSpace(stringArg(command.Arguments["parentAssetId"])) != "" || strings.TrimSpace(stringArg(command.Arguments["parentCommandId"])) != "" {
			continue
		}
		word, pos := realtimeVoiceFirstDestinationWordPosition(firstStringArg(command.Arguments["title"], command.Arguments["name"]), transcript)
		if word == "" || pos < 0 {
			continue
		}
		creates = append(creates, createCommand{index: index, id: command.ID, word: word, pos: pos})
	}
	for childIndex, child := range creates {
		parent := createCommand{pos: len(transcript) + 1}
		for parentIndex, candidate := range creates {
			if parentIndex == childIndex || candidate.pos <= child.pos || candidate.pos >= parent.pos {
				continue
			}
			parent = candidate
		}
		if parent.id == "" {
			continue
		}
		normalized[child.index].Arguments = realtimeVoiceArgumentsWithParentReference(normalized[child.index].Arguments, "", parent.id)
	}
	return normalized
}

func realtimeVoiceFirstDestinationWordPosition(title string, transcript string) (string, int) {
	transcript = strings.ToLower(transcript)
	bestWord := ""
	bestIndex := len(transcript) + 1
	for _, word := range realtimeVoiceMeaningfulWords(title) {
		if !realtimeVoiceDestinationSegmentWords[word] {
			continue
		}
		index := strings.Index(transcript, word)
		if index >= 0 && index < bestIndex {
			bestWord = word
			bestIndex = index
		}
	}
	if bestWord == "" {
		return "", -1
	}
	return bestWord, bestIndex
}

func appendStableRealtimeVoiceCommandIndex(indexes []int, index int) []int {
	insertAt := len(indexes)
	for readyIndex, existing := range indexes {
		if index < existing {
			insertAt = readyIndex
			break
		}
	}
	indexes = append(indexes, 0)
	copy(indexes[insertAt+1:], indexes[insertAt:])
	indexes[insertAt] = index
	return indexes
}
