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

func TestRealtimeVoiceAssetDetailToolRequiresVisibleAsset(t *testing.T) {
	t.Parallel()

	var detailDescription string
	var detailReadOnly bool
	var detailProviderCallable bool
	for _, tool := range realtimeVoiceToolDescriptors() {
		if tool.Name == RealtimeVoiceToolGetAssetDetail {
			detailDescription = tool.Description
			detailReadOnly = tool.ReadOnly
			detailProviderCallable = tool.ProviderCallable
		}
	}
	if !detailReadOnly || !detailProviderCallable {
		t.Fatalf("expected asset detail tool to be provider-callable read-only")
	}
	for _, required := range []string{
		"already returned by an authorized read tool",
		"precise detail",
		"Do not speak or display asset IDs",
	} {
		if !strings.Contains(detailDescription, required) {
			t.Fatalf("expected asset detail guidance to include %q, got %q", required, detailDescription)
		}
	}
}

func TestRealtimeVoiceListToolDescribesRootLevelFilter(t *testing.T) {
	t.Parallel()

	var listDescription string
	var parentScopeEnum []string
	for _, tool := range realtimeVoiceToolDescriptors() {
		if tool.Name != RealtimeVoiceToolListAuthorizedAssets {
			continue
		}
		listDescription = tool.Description
		parentScopeEnum = tool.Parameters.Properties["parentScope"].Enum
	}
	for _, required := range []string{
		"root-level assets",
		"parentScope any|root",
		"do not combine it with parentTitle or locationTitle",
	} {
		if !strings.Contains(listDescription, required) {
			t.Fatalf("expected list guidance to include %q, got %q", required, listDescription)
		}
	}
	if len(parentScopeEnum) != 2 || parentScopeEnum[0] != "any" || parentScopeEnum[1] != "root" {
		t.Fatalf("expected parentScope enum any/root, got %+v", parentScopeEnum)
	}
}

func TestRealtimeVoiceCheckoutReadToolsAreProviderCallableReadOnly(t *testing.T) {
	t.Parallel()

	tools := map[string]string{}
	readOnly := map[string]bool{}
	providerCallable := map[string]bool{}
	for _, tool := range realtimeVoiceToolDescriptors() {
		tools[tool.Name] = tool.Description
		readOnly[tool.Name] = tool.ReadOnly
		providerCallable[tool.Name] = tool.ProviderCallable
	}
	for _, name := range []string{
		RealtimeVoiceToolListCheckedOutAssets,
		RealtimeVoiceToolListAssetCheckoutHistory,
	} {
		if !readOnly[name] || !providerCallable[name] {
			t.Fatalf("expected %s to be provider-callable read-only", name)
		}
	}
	for _, required := range []string{
		"currently checked out",
		"checkout state",
	} {
		if !strings.Contains(tools[RealtimeVoiceToolListCheckedOutAssets], required) {
			t.Fatalf("expected checked-out list guidance to include %q, got %q", required, tools[RealtimeVoiceToolListCheckedOutAssets])
		}
	}
	for _, required := range []string{
		"checkout history",
		"already returned by an authorized read tool",
		"checkout and return details",
	} {
		if !strings.Contains(tools[RealtimeVoiceToolListAssetCheckoutHistory], required) {
			t.Fatalf("expected checkout history guidance to include %q, got %q", required, tools[RealtimeVoiceToolListAssetCheckoutHistory])
		}
	}
}
