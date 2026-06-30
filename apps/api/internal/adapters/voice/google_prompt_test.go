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
