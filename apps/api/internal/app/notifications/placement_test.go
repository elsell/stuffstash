package notifications

import (
	"context"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

type placementAssets struct {
	ports.AssetRepository
	items map[asset.ID]asset.Asset
}

func (f placementAssets) AssetByID(ctx context.Context, tid tenant.ID, iid inventory.InventoryID, id asset.ID) (asset.Asset, bool, error) {
	if item, ok := f.items[id]; ok {
		return item, true, nil
	}
	return f.AssetRepository.AssetByID(ctx, tid, iid, id)
}
func TestNotificationPlacementBoundsAndRejectsUnsafeAncestors(t *testing.T) {
	for _, mode := range []string{"foreign", "archived", "missing", "cycle", "depth"} {
		t.Run(mode, func(t *testing.T) {
			s, store, input := deliveryFixture(t, &pushSenderFake{})
			ctx := context.Background()
			entries, err := store.ListNotifications(ctx, input.Scope(), "", 10)
			if err != nil || len(entries) != 1 {
				t.Fatal("fixture notifications", err)
			}
			item, _, _ := store.AssetByID(ctx, input.TenantID, input.InventoryID, "bottle")
			item.ParentAssetID = "parent"
			title, _ := asset.NewTitle("Parent")
			parent := asset.Asset{ID: "parent", TenantID: "home", InventoryID: "main", Title: title, Kind: asset.KindContainer, LifecycleState: asset.LifecycleStateActive}
			items := map[asset.ID]asset.Asset{"bottle": item, "parent": parent}
			want := 0
			switch mode {
			case "foreign":
				parent.InventoryID = "other"
				items["parent"] = parent
			case "archived":
				parent.LifecycleState = asset.LifecycleStateArchived
				items["parent"] = parent
			case "missing":
				delete(items, "parent")
			case "cycle":
				parent.ParentAssetID = "parent"
				items["parent"] = parent
				want = 1
			case "depth":
				want = maxNotificationAncestors
				for i := 0; i <= maxNotificationAncestors; i++ {
					node := parent
					node.ID = asset.ID(fmt.Sprintf("p%d", i))
					node.ParentAssetID = asset.ID(fmt.Sprintf("p%d", i+1))
					items[node.ID] = node
				}
				item.ParentAssetID = "p0"
				items["bottle"] = item
			}
			s.deps.Assets = placementAssets{AssetRepository: store, items: items}
			view, err := s.GetNotification(ctx, input, entries[0].ID)
			if err != nil {
				t.Fatal(err)
			}
			if !view.ParentTrailIncomplete || len(view.ParentTrail) != want {
				t.Fatalf("unsafe or unbounded trail: %d %v", len(view.ParentTrail), view.ParentTrailIncomplete)
			}
		})
	}
}

type unavailableParent struct{ ports.AssetRepository }

func (f unavailableParent) AssetByID(ctx context.Context, tid tenant.ID, iid inventory.InventoryID, id asset.ID) (asset.Asset, bool, error) {
	if id == "unavailable" {
		return asset.Asset{}, false, fmt.Errorf("parent unavailable")
	}
	item, found, err := f.AssetRepository.AssetByID(ctx, tid, iid, id)
	if id == "bottle" {
		item.ParentAssetID = "unavailable"
	}
	return item, found, err
}
func TestNotificationBatchOperationsDoNotReadAncestors(t *testing.T) {
	s, store, input := deliveryFixture(t, &pushSenderFake{})
	s.deps.Assets = unavailableParent{store}
	ctx := context.Background()
	if _, err := s.ListInbox(ctx, input, "", 10, false); err == nil {
		t.Fatal("fixture must fail parent lookup for displayed inbox")
	}
	count, err := s.CountUnreadPage(ctx, input, "")
	if err != nil || count.Count != 1 {
		t.Fatal("count read parent", count, err)
	}
	if _, err := s.MarkInboxPageRead(ctx, input, ""); err != nil {
		t.Fatal("mark all read parent", err)
	}
	count, err = s.CountUnreadPage(ctx, input, "")
	if err != nil || count.Count != 0 {
		t.Fatal("mark all did not update read state", count, err)
	}
}
