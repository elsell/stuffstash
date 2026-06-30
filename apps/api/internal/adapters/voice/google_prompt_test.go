package voice

import (
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiLanguagePromptGuidesNestedMissingDestinationsIntoPlans(t *testing.T) {
	t.Parallel()

	prompt := languagePrompt(ports.LanguageInferenceInput{
		Transcript: "Move my water bottle to the second shelf in the big cabinet in the kitchen.",
	})

	for _, required := range []string{
		"For nested missing destinations, create every missing path segment in order",
		"create Kitchen, create Big cabinet with parentCommandId cmd-kitchen, create Second shelf with parentCommandId cmd-big-cabinet",
		"then move the water bottle with parentCommandId cmd-second-shelf",
	} {
		if !strings.Contains(prompt, required) {
			t.Fatalf("expected prompt to include %q, got %s", required, prompt)
		}
	}
}

func TestGoogleGeminiLanguagePromptGuidesNewThingsIntoCreateCommands(t *testing.T) {
	t.Parallel()

	prompt := languagePrompt(ports.LanguageInferenceInput{
		Transcript: "Add an Apple TV remote to the box under the TV in the living room.",
	})

	for _, required := range []string{
		"For add/create requests for a new item, use create_asset with title or name and kind item",
		"Do not invent an assetId for a new item",
		"Never include assetId in create_asset arguments",
		"When a new item should go inside an existing parent, use one create_asset command with parentAssetId set to the visible parent.",
		"Do not create the item and then move it.",
		"create the container with parentAssetId set to that visible location assetId",
		"Use create_asset with kind container for new containers",
		"Do not create the new item first and do not add a move_asset command for the new item.",
		"use move_asset only for an existing asset returned by a read tool",
	} {
		if !strings.Contains(prompt, required) {
			t.Fatalf("expected prompt to include %q, got %s", required, prompt)
		}
	}
}

func TestGoogleGeminiLanguagePromptUsesCompactReadOnlyPromptForRequiredToolTurns(t *testing.T) {
	t.Parallel()

	prompt := languagePrompt(ports.LanguageInferenceInput{
		Transcript:      "Add an Apple TV remote to the box under the TV in the living room.",
		RequireToolCall: true,
	})

	for _, required := range []string{
		"This turn must gather context with exactly one provided read tool.",
		"For add/create requests into a nested destination, search the outermost room, place, or container separately first",
		"do not search only the item or the whole destination phrase",
		"this first read turn must search the source item first, not the destination",
		"Use short search keywords copied from the transcript.",
		"Do not answer yet and do not propose changes on this turn.",
	} {
		if !strings.Contains(prompt, required) {
			t.Fatalf("expected compact required-tool prompt to include %q, got %s", required, prompt)
		}
	}
	for _, forbidden := range []string{
		"propose_action_plan",
		"create Kitchen",
		"response schema",
	} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("expected compact required-tool prompt to omit %q, got %s", forbidden, prompt)
		}
	}
}
