package app

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) validateRealtimeVoiceActionPlanTranscriptAlignment(ctx context.Context, session RealtimeVoiceSession, commands []ActionPlanCommandInput, transcript string) error {
	for _, command := range commands {
		record, err := realtimeVoiceActionPlanCommandRecord(command)
		if err != nil {
			return err
		}
		switch command.Kind {
		case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation:
			args, err := parseActionPlanCreateArguments(record)
			if err != nil {
				return ports.ErrInvalidProviderInput
			}
			if strings.TrimSpace(args.ParentAssetID) != "" {
				if err := a.validateRealtimeVoiceMentionedParentAssetID(ctx, session, args.ParentAssetID, transcript); err != nil {
					return err
				}
			}
			if strings.TrimSpace(args.ParentAssetID) == "" && strings.TrimSpace(args.ParentCommandID) == "" {
				if err := validateRealtimeVoiceRootCreate(command.Kind, args.Kind, transcript); err != nil {
					return err
				}
				if err := a.validateRealtimeVoiceNoDuplicateRootCreate(ctx, session, command.Kind, args.Title, args.Kind); err != nil {
					return err
				}
			}
		case actionplan.CommandKindMoveAsset:
			args, err := parseActionPlanMoveArguments(record)
			if err != nil {
				return ports.ErrInvalidProviderInput
			}
			if err := a.validateRealtimeVoiceMentionedAssetID(ctx, session, args.AssetID.String(), transcript); err != nil {
				return err
			}
			if strings.TrimSpace(args.ParentAssetID) != "" {
				if err := a.validateRealtimeVoiceMentionedParentAssetID(ctx, session, args.ParentAssetID, transcript); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateRealtimeVoiceRootCreate(commandKind actionplan.CommandKind, rawKind string, transcript string) error {
	if !realtimeVoiceTranscriptNamesDestination(transcript) || realtimeVoiceTranscriptAllowsRootDestination(transcript) {
		return nil
	}
	if commandKind != actionplan.CommandKindCreateAsset {
		return nil
	}
	kind := strings.TrimSpace(rawKind)
	if kind == "" {
		kind = asset.KindItem.String()
	}
	if kind == asset.KindItem.String() {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func realtimeVoiceNoMatchQueries(results []ports.AgentToolResult) []string {
	queries := []string{}
	for _, result := range results {
		if result.Name != RealtimeVoiceToolSearchAuthorizedAssets {
			continue
		}
		var output realtimeVoiceAssetToolOutput
		if err := json.Unmarshal([]byte(result.Content), &output); err != nil {
			continue
		}
		if output.Count == 0 && strings.TrimSpace(output.Query) != "" {
			queries = append(queries, output.Query)
		}
	}
	return queries
}

func realtimeVoiceQueryLooksLikeDestinationSegment(query string, transcript string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	transcript = strings.ToLower(strings.TrimSpace(transcript))
	if query == "" || !strings.Contains(transcript, query) {
		return false
	}
	for _, word := range realtimeVoiceMeaningfulWords(query) {
		if realtimeVoiceDestinationSegmentWords[word] {
			return true
		}
	}
	return false
}

func realtimeVoiceMeaningfulWordsRepresented(query string, represented string) bool {
	representedWords := map[string]struct{}{}
	for _, word := range realtimeVoiceMeaningfulWords(represented) {
		representedWords[word] = struct{}{}
	}
	for _, word := range realtimeVoiceMeaningfulWords(query) {
		if !realtimeVoiceWordPresent(representedWords, word) {
			return false
		}
	}
	return true
}

func realtimeVoiceTranscriptHasUnrepresentedDestinationSegment(transcript string, represented string) bool {
	representedWords := map[string]struct{}{}
	for _, word := range realtimeVoiceMeaningfulWords(represented) {
		representedWords[word] = struct{}{}
	}
	for _, word := range realtimeVoiceMeaningfulWords(transcript) {
		if !realtimeVoiceDestinationSegmentWords[word] {
			continue
		}
		if _, exists := representedWords[word]; !exists {
			return true
		}
	}
	return false
}

func realtimeVoiceLikelyOuterDestinationQuery(transcript string, toolResults []ports.AgentToolResult) string {
	represented := strings.Builder{}
	for _, query := range realtimeVoiceNoMatchQueries(toolResults) {
		represented.WriteString(" ")
		represented.WriteString(query)
	}
	representedText := represented.String()
	for _, phrase := range []string{"living room", "kitchen", "garage", "office", "bedroom", "bathroom", "basement", "attic", "pantry"} {
		if strings.Contains(strings.ToLower(transcript), phrase) && !realtimeVoiceVisibleParentTitlePresent(toolResults, phrase) {
			return phrase
		}
	}
	for _, word := range realtimeVoiceMeaningfulWords(transcript) {
		if realtimeVoiceDestinationSegmentWords[word] && !realtimeVoiceMeaningfulWordsRepresented(word, representedText) {
			return word
		}
	}
	return ""
}

func realtimeVoiceVisibleParentTitlePresent(toolResults []ports.AgentToolResult, title string) bool {
	title = normalizeRealtimeVoiceSourceText(title)
	if title == "" {
		return false
	}
	for _, parent := range realtimeVoiceVisibleParentsInTranscript(toolResults, title) {
		if normalizeRealtimeVoiceSourceText(parent.Title) == title {
			return true
		}
	}
	return false
}

func realtimeVoiceMeaningfulWords(value string) []string {
	words := []string{}
	for _, word := range strings.Fields(strings.ToLower(value)) {
		word = strings.Trim(word, ".,!?;:'\"()[]{}")
		if (len(word) < 3 && !realtimeVoiceShortMeaningfulWords[word]) || realtimeVoiceTranscriptStopWords[word] {
			continue
		}
		words = append(words, word)
	}
	return words
}

func (a App) validateRealtimeVoiceNoDuplicateRootCreate(ctx context.Context, session RealtimeVoiceSession, commandKind actionplan.CommandKind, title string, rawKind string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ports.ErrInvalidProviderInput
	}
	kind := strings.TrimSpace(rawKind)
	if commandKind == actionplan.CommandKindCreateLocation {
		kind = asset.KindLocation.String()
	}
	if kind != asset.KindLocation.String() && kind != asset.KindContainer.String() {
		return nil
	}
	results, err := a.SearchAssets(ctx, SearchAssetsInput{
		Principal:      session.Principal,
		TenantID:       session.TenantID,
		InventoryIDs:   []inventory.InventoryID{session.InventoryID},
		Query:          title,
		Mode:           "fuzzy",
		LifecycleState: "active",
		Limit:          5,
	})
	if err != nil {
		return err
	}
	for _, result := range results.Items {
		if strings.EqualFold(result.Asset.Title.String(), title) && result.Asset.Kind.String() == kind {
			return ports.ErrInvalidProviderInput
		}
	}
	return nil
}

func (a App) validateRealtimeVoiceMentionedAssetID(ctx context.Context, session RealtimeVoiceSession, rawAssetID string, transcript string) error {
	id, ok := asset.NewID(strings.TrimSpace(rawAssetID))
	if !ok {
		return ports.ErrInvalidProviderInput
	}
	item, found, err := a.assets.AssetByID(ctx, session.TenantID, session.InventoryID, id)
	if err != nil {
		return err
	}
	if !found {
		return ports.ErrInvalidProviderInput
	}
	if !realtimeVoiceTitleMentionedInTranscript(item.Title.String(), transcript) {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func (a App) validateRealtimeVoiceMentionedParentAssetID(ctx context.Context, session RealtimeVoiceSession, rawAssetID string, transcript string) error {
	id, ok := asset.NewID(strings.TrimSpace(rawAssetID))
	if !ok {
		return ports.ErrInvalidProviderInput
	}
	item, found, err := a.assets.AssetByID(ctx, session.TenantID, session.InventoryID, id)
	if err != nil {
		return err
	}
	if !found {
		return ports.ErrInvalidProviderInput
	}
	if !realtimeVoiceParentTitleMentionedInTranscript(item.Title.String(), transcript) {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func realtimeVoiceTitleMentionedInTranscript(title string, transcript string) bool {
	transcriptWords := map[string]struct{}{}
	for _, word := range realtimeVoiceMeaningfulWords(transcript) {
		transcriptWords[word] = struct{}{}
	}
	for _, word := range realtimeVoiceMeaningfulWords(title) {
		if realtimeVoiceWordPresent(transcriptWords, word) {
			return true
		}
	}
	return false
}

func realtimeVoiceParentTitleMentionedInTranscript(title string, transcript string) bool {
	transcriptWords := map[string]struct{}{}
	for _, word := range realtimeVoiceMeaningfulWords(transcript) {
		transcriptWords[word] = struct{}{}
	}
	titleWords := realtimeVoiceMeaningfulWords(title)
	if len(titleWords) == 0 {
		return false
	}
	for _, word := range titleWords {
		if !realtimeVoiceWordPresent(transcriptWords, word) {
			return false
		}
	}
	return true
}

func realtimeVoiceWordPresent(words map[string]struct{}, word string) bool {
	for _, candidate := range realtimeVoiceWordForms(word) {
		if _, ok := words[candidate]; ok {
			return true
		}
	}
	return false
}

func realtimeVoiceWordForms(word string) []string {
	word = strings.TrimSpace(word)
	if word == "" {
		return nil
	}
	formSet := map[string]struct{}{
		word:       {},
		word + "s": {},
	}
	if strings.HasSuffix(word, "x") || strings.HasSuffix(word, "s") || strings.HasSuffix(word, "ch") || strings.HasSuffix(word, "sh") {
		formSet[word+"es"] = struct{}{}
	}
	if strings.HasSuffix(word, "y") && len(word) > 1 {
		formSet[strings.TrimSuffix(word, "y")+"ies"] = struct{}{}
	}
	if strings.HasSuffix(word, "f") && len(word) > 1 {
		formSet[strings.TrimSuffix(word, "f")+"ves"] = struct{}{}
	}
	if strings.HasSuffix(word, "fe") && len(word) > 2 {
		formSet[strings.TrimSuffix(word, "fe")+"ves"] = struct{}{}
	}
	if strings.HasSuffix(word, "ies") && len(word) > 3 {
		formSet[strings.TrimSuffix(word, "ies")+"y"] = struct{}{}
	}
	if strings.HasSuffix(word, "ves") && len(word) > 3 {
		formSet[strings.TrimSuffix(word, "ves")+"f"] = struct{}{}
		formSet[strings.TrimSuffix(word, "ves")+"fe"] = struct{}{}
	}
	if strings.HasSuffix(word, "es") && len(word) > 2 {
		formSet[strings.TrimSuffix(word, "es")] = struct{}{}
	}
	if strings.HasSuffix(word, "s") && len(word) > 1 {
		formSet[strings.TrimSuffix(word, "s")] = struct{}{}
	}
	forms := make([]string, 0, len(formSet))
	for form := range formSet {
		if form != "" {
			forms = append(forms, form)
		}
	}
	return forms
}

var realtimeVoiceTranscriptStopWords = map[string]bool{
	"the":    true,
	"and":    true,
	"for":    true,
	"with":   true,
	"under":  true,
	"inside": true,
}

var realtimeVoiceShortMeaningfulWords = map[string]bool{
	"tv": true,
}

var realtimeVoiceDestinationSegmentWords = map[string]bool{
	"box":      true,
	"cabinet":  true,
	"shelf":    true,
	"drawer":   true,
	"bin":      true,
	"basket":   true,
	"closet":   true,
	"counter":  true,
	"kitchen":  true,
	"living":   true,
	"room":     true,
	"garage":   true,
	"office":   true,
	"bedroom":  true,
	"bathroom": true,
	"basement": true,
	"attic":    true,
	"pantry":   true,
}

var realtimeVoiceContainerSegmentWords = map[string]bool{
	"box":     true,
	"cabinet": true,
	"shelf":   true,
	"drawer":  true,
	"bin":     true,
	"basket":  true,
	"closet":  true,
	"counter": true,
}
