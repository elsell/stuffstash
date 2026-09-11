package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestExpirationProposalRejectsUnavailableTypeBeforeReview(t *testing.T) {
	for _, mode := range []string{"enabled", "disabled", "archived", "foreign", "missing"} {
		t.Run(mode, func(t *testing.T) {
			application, store := newRealtimeVoiceResolutionTestAppWithStore(t, successfulRealtimeVoiceResolver())
			application.customAssetTypes = store
			kind, _ := customfield.NewAssetType("medicine", "tenant-home", "inventory-home", customfield.ScopeInventory, "medicine", "Medicine", "")
			kind.ExpirationEnabled = mode != "disabled"
			if mode == "archived" {
				kind.LifecycleState = customfield.AssetTypeLifecycleArchived
			}
			if mode == "foreign" {
				kind.InventoryID = "other-inventory"
				name, _ := inventory.NewName("Other")
				if err := store.SaveInventory(context.Background(), inventory.Inventory{ID: "other-inventory", TenantID: "tenant-home", Name: name}); err != nil {
					t.Fatal(err)
				}
			}
			if mode != "missing" {
				if err := store.SaveCustomAssetType(context.Background(), kind, audit.Record{ID: "seed-type"}); err != nil {
					t.Fatal(err)
				}
			}
			executor := realtimeConversationTools{application: application, session: checkoutToolSession(), visible: map[string]struct{}{}}
			result, err := executor.propose(context.Background(), ports.AgentToolCall{ID: "propose", Name: realtimeConversationProposeTool, Arguments: map[string]any{"summary": "Add bottle expiring February 2028", "commands": []any{map[string]any{"id": "bottle", "kind": "create_asset", "summary": "Add bottle", "arguments": map[string]any{"title": "Bottle", "customAssetTypeId": "medicine", "expiration": map[string]any{"date": "2028-02", "precision": "month"}}}}}})
			if mode != "enabled" {
				if err == nil || result.ApprovalPlanID != "" || executor.proposal != nil {
					t.Fatal("unavailable type became dated review")
				}
				return
			}
			if err != nil || executor.proposal == nil || executor.proposal.Commands[0].Expiration == nil {
				t.Fatalf("valid dated review failed: %v", err)
			}
			saved, found, err := store.ActionPlanByID(context.Background(), "tenant-home", "inventory-home", result.ApprovalPlanID)
			if err != nil || !found {
				t.Fatal("proposal not persisted")
			}
			args, err := parseActionPlanCreateArguments(saved.Commands[0])
			if err != nil || args.Expiration == nil || args.Expiration.Date != "2028-02" {
				t.Fatal("saved draft lost date")
			}
		})
	}
}
