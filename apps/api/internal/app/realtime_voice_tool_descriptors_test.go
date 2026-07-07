package app

import (
	"strings"
	"testing"
)

func TestRealtimeVoiceActionPlanToolGuidesMissingDestinationsToCreate(t *testing.T) {
	t.Parallel()

	var proposalDescription string
	var proposalReadOnly bool
	var proposalProviderCallable bool
	var proposalRequiresApproval bool
	for _, tool := range realtimeVoiceToolDescriptors() {
		if tool.Name == RealtimeVoiceToolProposeActionPlan {
			proposalDescription = tool.Description
			proposalReadOnly = tool.ReadOnly
			proposalProviderCallable = tool.ProviderCallable
			proposalRequiresApproval = tool.RequiresApproval
		}
	}
	if proposalReadOnly || !proposalProviderCallable || !proposalRequiresApproval {
		t.Fatalf("expected proposal tool to be provider-callable and approval-gated but not read-only")
	}
	for _, required := range []string{
		"missing but clearly named location or container",
		"assume they want it created",
		"Use create_asset for new items and containers",
		"Use checkout_asset",
		"Use return_asset",
		"optional details",
		"Never put assetId in create_asset arguments",
		"create the missing container first with parentAssetId set to the existing location",
		"moving the existing asset using parentCommandId",
		"Do not ask a final yes/no clarification",
	} {
		if !strings.Contains(proposalDescription, required) {
			t.Fatalf("expected proposal tool guidance to include %q, got %q", required, proposalDescription)
		}
	}
}

func TestRealtimeVoiceAuditHistoryToolRequiresVisibleAsset(t *testing.T) {
	t.Parallel()

	var historyDescription string
	var historyReadOnly bool
	var historyProviderCallable bool
	for _, tool := range realtimeVoiceToolDescriptors() {
		if tool.Name == RealtimeVoiceToolListAssetAuditHistory {
			historyDescription = tool.Description
			historyReadOnly = tool.ReadOnly
			historyProviderCallable = tool.ProviderCallable
		}
	}
	if !historyReadOnly || !historyProviderCallable {
		t.Fatalf("expected audit history tool to be provider-callable read-only")
	}
	for _, required := range []string{
		"already returned by an authorized read tool",
		"movement",
		"when",
		"Do not guess",
	} {
		if !strings.Contains(historyDescription, required) {
			t.Fatalf("expected audit history guidance to include %q, got %q", required, historyDescription)
		}
	}
}
