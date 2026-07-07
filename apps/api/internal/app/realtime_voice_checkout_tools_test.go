package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceListCheckedOutAssetsToolReturnsCurrentCheckoutState(t *testing.T) {
	t.Parallel()

	drill := assetItem("drill-1", "tenant-home", "inventory-home", asset.KindItem, "")
	now := time.Date(2026, 6, 26, 17, 30, 0, 0, time.UTC)
	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{
			drill.ID: drill,
		},
		checkouts: map[asset.CheckoutID]asset.Checkout{
			asset.CheckoutID("checkout-drill"): {
				ID:                    asset.CheckoutID("checkout-drill"),
				TenantID:              asset.TenantID("tenant-home"),
				InventoryID:           asset.InventoryID("inventory-home"),
				AssetID:               drill.ID,
				State:                 asset.CheckoutStateOpen,
				CheckedOutAt:          now,
				CheckedOutByPrincipal: "user-1",
			},
		},
	}, &fakeIDGenerator{})

	result, _, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), "", nil, ports.AgentToolCall{
		ID:        "tool-call-checked-out",
		Name:      RealtimeVoiceToolListCheckedOutAssets,
		Arguments: map[string]any{"limit": 5},
	}, map[string]struct{}{})
	if err != nil {
		t.Fatalf("execute checked-out list tool: %v", err)
	}
	for _, required := range []string{
		`"tool":"list_checked_out_assets"`,
		`"assetId":"drill-1"`,
		`"title":"Asset drill-1"`,
		`"currentCheckout":{"id":"checkout-drill"`,
		`"checkedOutByPrincipalId":"user-1"`,
	} {
		if !strings.Contains(result.Content, required) {
			t.Fatalf("expected checked-out tool result to include %q, got %s", required, result.Content)
		}
	}
}

func TestRealtimeVoiceListCheckedOutAssetsToolMakesAssetsVisibleForCheckoutHistory(t *testing.T) {
	t.Parallel()

	drill := assetItem("drill-1", "tenant-home", "inventory-home", asset.KindItem, "")
	now := time.Date(2026, 6, 26, 17, 30, 0, 0, time.UTC)
	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{
			drill.ID: drill,
		},
		checkouts: map[asset.CheckoutID]asset.Checkout{
			asset.CheckoutID("checkout-drill"): {
				ID:                    asset.CheckoutID("checkout-drill"),
				TenantID:              asset.TenantID("tenant-home"),
				InventoryID:           asset.InventoryID("inventory-home"),
				AssetID:               drill.ID,
				State:                 asset.CheckoutStateOpen,
				CheckedOutAt:          now,
				CheckedOutByPrincipal: "user-1",
			},
		},
	}, &fakeIDGenerator{})

	visibleAssetIDs := map[string]struct{}{}
	result, _, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), "", nil, ports.AgentToolCall{
		ID:        "tool-call-checked-out",
		Name:      RealtimeVoiceToolListCheckedOutAssets,
		Arguments: map[string]any{"limit": 5},
	}, visibleAssetIDs)
	if err != nil {
		t.Fatalf("execute checked-out list tool: %v", err)
	}
	if err := collectRealtimeVoiceVisibleAssetIDs(result, visibleAssetIDs); err != nil {
		t.Fatalf("collect visible asset IDs: %v", err)
	}
	if _, ok := visibleAssetIDs["drill-1"]; !ok {
		t.Fatalf("expected checked-out list result to mark drill visible, got %+v", visibleAssetIDs)
	}

	_, _, err = application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), "", nil, ports.AgentToolCall{
		ID:        "tool-call-history",
		Name:      RealtimeVoiceToolListAssetCheckoutHistory,
		Arguments: map[string]any{"assetId": "drill-1", "limit": 5},
	}, visibleAssetIDs)
	if err != nil {
		t.Fatalf("expected checkout history to accept asset made visible by checked-out list: %v", err)
	}
}

func TestRealtimeVoiceListAssetCheckoutHistoryToolRequiresVisibleAsset(t *testing.T) {
	t.Parallel()

	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{}, &fakeIDGenerator{})

	_, _, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), "", nil, ports.AgentToolCall{
		ID:        "tool-call-history",
		Name:      RealtimeVoiceToolListAssetCheckoutHistory,
		Arguments: map[string]any{"assetId": "hidden-asset", "limit": 5},
	}, map[string]struct{}{})
	if err == nil {
		t.Fatalf("expected invisible checkout history asset to be rejected")
	}
}

func TestRealtimeVoiceListAssetCheckoutHistoryToolReturnsBoundedCheckoutRecords(t *testing.T) {
	t.Parallel()

	drill := assetItem("drill-1", "tenant-home", "inventory-home", asset.KindItem, "")
	now := time.Date(2026, 6, 26, 17, 30, 0, 0, time.UTC)
	details, _ := asset.NewCheckoutDetails("Loaned to Sam")
	returnDetails, _ := asset.NewCheckoutDetails("Back in the tool bin")
	application := newActionPlanExecutionTestApp(&fakeActionPlanRepository{}, &fakeAssetRepository{
		items: map[asset.ID]asset.Asset{
			drill.ID: drill,
		},
		checkouts: map[asset.CheckoutID]asset.Checkout{
			asset.CheckoutID("checkout-drill"): {
				ID:                    asset.CheckoutID("checkout-drill"),
				TenantID:              asset.TenantID("tenant-home"),
				InventoryID:           asset.InventoryID("inventory-home"),
				AssetID:               drill.ID,
				State:                 asset.CheckoutStateReturned,
				CheckedOutAt:          now.Add(-2 * time.Hour),
				CheckedOutByPrincipal: "user-1",
				CheckoutDetails:       details,
				ReturnedAt:            now,
				ReturnedByPrincipal:   "user-2",
				ReturnDetails:         returnDetails,
			},
		},
	}, &fakeIDGenerator{})

	result, _, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), "", nil, ports.AgentToolCall{
		ID:        "tool-call-history",
		Name:      RealtimeVoiceToolListAssetCheckoutHistory,
		Arguments: map[string]any{"assetId": "drill-1", "limit": 5},
	}, map[string]struct{}{"drill-1": {}})
	if err != nil {
		t.Fatalf("execute checkout history tool: %v", err)
	}
	for _, required := range []string{
		`"tool":"list_asset_checkout_history"`,
		`"assetId":"drill-1"`,
		`"state":"returned"`,
		`"checkoutDetails":"Loaned to Sam"`,
		`"returnDetails":"Back in the tool bin"`,
		`"returnedByPrincipalId":"user-2"`,
	} {
		if !strings.Contains(result.Content, required) {
			t.Fatalf("expected checkout history tool result to include %q, got %s", required, result.Content)
		}
	}
}

func TestRealtimeVoiceRunsRequiredCheckoutHistoryBeforeAcceptingFinalAnswer(t *testing.T) {
	t.Parallel()

	args, ok := realtimeVoiceCheckoutHistoryArgs("Who checked out the loaner flashlight?", []ports.AgentToolResult{{
		Name:    RealtimeVoiceToolSearchAuthorizedAssets,
		Content: `{"tool":"search_authorized_assets","query":"loaner flashlight","count":1,"items":[{"assetId":"loaner-flashlight-1","title":"Loaner flashlight","kind":"item"}]}`,
	}})
	if !ok || args["assetId"] != "loaner-flashlight-1" {
		t.Fatalf("expected checkout history args to use visible asset, ok=%v args=%+v", ok, args)
	}

	tts := &resolvedTextToSpeech{}
	language := &scriptedRealtimeLanguageInference{turns: []ports.LanguageInferenceTurn{
		{
			ToolCalls: []ports.AgentToolCall{{
				ID:        "search-loaner-flashlight",
				Name:      RealtimeVoiceToolSearchAuthorizedAssets,
				Arguments: map[string]any{"query": "loaner flashlight"},
			}},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "Sam has the loaner flashlight.",
				DisplayResponse: "Sam has the loaner flashlight.",
			},
		},
		{
			Final: &ports.StructuredAgentResponse{
				Kind:            ports.StructuredAgentResponseKindAnswer,
				SpokenResponse:  "Sam has the loaner flashlight.",
				DisplayResponse: "Sam has the loaner flashlight.",
			},
		},
	}}
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Who checked out the loaner flashlight?"}
	resolver.providers.LanguageInference = language
	resolver.providers.TextToSpeech = tts
	application, store := newRealtimeVoiceResolutionTestAppWithStore(t, resolver)
	loaner := assetItem("loaner-flashlight-1", "tenant-home", "inventory-home", asset.KindItem, "")
	loanerTitle, _ := asset.NewTitle("Loaner flashlight")
	loaner.Title = loanerTitle
	seedRealtimeVoiceLoopAsset(t, store, loaner, "audit-loaner-flashlight")
	details, _ := asset.NewCheckoutDetails("Loaned to Sam")
	if err := store.CheckOutAsset(context.Background(), asset.Checkout{
		ID:                    asset.CheckoutID("checkout-loaner-flashlight"),
		TenantID:              asset.TenantID("tenant-home"),
		InventoryID:           asset.InventoryID("inventory-home"),
		AssetID:               loaner.ID,
		State:                 asset.CheckoutStateOpen,
		CheckedOutAt:          time.Date(2026, 6, 29, 13, 8, 30, 0, time.UTC),
		CheckedOutByPrincipal: "principal-home",
		CheckoutDetails:       details,
		CreatedAt:             time.Date(2026, 6, 29, 13, 8, 30, 0, time.UTC),
		UpdatedAt:             time.Date(2026, 6, 29, 13, 8, 30, 0, time.UTC),
	}, audit.Record{
		ID:          audit.ID("audit-checkout-loaner-flashlight"),
		TenantID:    audit.TenantID("tenant-home"),
		InventoryID: audit.InventoryID("inventory-home"),
		Action:      audit.ActionAssetCheckedOut,
		TargetType:  audit.TargetAsset,
		TargetID:    "loaner-flashlight-1",
		OccurredAt:  time.Date(2026, 6, 29, 13, 8, 30, 0, time.UTC),
	}, nil); err != nil {
		t.Fatalf("seed checkout: %v", err)
	}
	directHistory, _, err := application.executeRealtimeVoiceTool(context.Background(), checkoutToolSession(), "", nil, ports.AgentToolCall{
		ID:        "direct-history",
		Name:      RealtimeVoiceToolListAssetCheckoutHistory,
		Arguments: map[string]any{"assetId": "loaner-flashlight-1"},
	}, map[string]struct{}{"loaner-flashlight-1": {}})
	if err != nil || !strings.Contains(directHistory.Content, "Loaned to Sam") {
		t.Fatalf("expected direct checkout history to work, result=%+v err=%v", directHistory, err)
	}

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error {
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %T %[1]v toolResults=%+v", err, language.seenToolResults)
	}

	if len(language.seenToolResults) < 3 {
		t.Fatalf("expected search, required checkout history, and final turn, got %+v", language.seenToolResults)
	}
	finalTurnResults := language.seenToolResults[len(language.seenToolResults)-1]
	if len(finalTurnResults) < 2 || finalTurnResults[1].Name != RealtimeVoiceToolListAssetCheckoutHistory || !strings.Contains(finalTurnResults[1].Content, "Loaned to Sam") {
		t.Fatalf("expected checkout history before final answer, got %+v", finalTurnResults)
	}
	if tts.lastText != "Sam has the loaner flashlight." {
		t.Fatalf("expected final response after required checkout history, got %q", tts.lastText)
	}
}

func checkoutToolSession() RealtimeVoiceSession {
	return RealtimeVoiceSession{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		Source:      RealtimeVoiceSourceMobile,
	}
}
