package app

import (
	"strings"
	"testing"
)

func TestRealtimeVoiceActionPlanToolGuidesMissingDestinationsToCreate(t *testing.T) {
	t.Parallel()

	var proposalDescription string
	for _, tool := range realtimeVoiceToolDescriptors() {
		if tool.Name == RealtimeVoiceToolProposeActionPlan {
			proposalDescription = tool.Description
		}
	}
	for _, required := range []string{
		"missing but clearly named location or container",
		"assume they want it created",
		"moving the existing asset using parentCommandId",
	} {
		if !strings.Contains(proposalDescription, required) {
			t.Fatalf("expected proposal tool guidance to include %q, got %q", required, proposalDescription)
		}
	}
}
