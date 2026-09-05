package search

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type pathRepository struct {
	*memory.Store
	items   map[asset.ID]asset.Asset
	reads   int
	corrupt bool
	cancel  context.CancelFunc
}

func (r *pathRepository) AssetByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, id asset.ID) (asset.Asset, bool, error) {
	if err := ctx.Err(); err != nil {
		return asset.Asset{}, false, err
	}
	r.reads++
	if r.cancel != nil {
		r.cancel()
	}
	item, ok := r.items[id]
	if !r.corrupt && (item.TenantID.String() != tenantID.String() || item.InventoryID.String() != inventoryID.String()) {
		ok = false
	}
	return item, ok, nil
}

func TestAncestorPathsShareReadsAndPreserveRootOrder(t *testing.T) {
	root := asset.Asset{ID: "root", TenantID: "tenant", InventoryID: "inventory", Title: "Root"}
	parent := asset.Asset{ID: "parent", TenantID: "tenant", InventoryID: "inventory", Title: "Parent", ParentAssetID: root.ID}
	repository := &pathRepository{items: map[asset.ID]asset.Asset{root.ID: root, parent.ID: parent}}
	results := []ports.AssetSearchResult{pathResult("one", parent.ID), pathResult("two", parent.ID)}
	got, err := WithAncestorPaths(context.Background(), repository, results)
	if err != nil {
		t.Fatal(err)
	}
	if repository.reads != 2 {
		t.Fatalf("expected two shared ancestor reads, got %d", repository.reads)
	}
	for _, item := range got {
		if len(item.AncestorPath) != 2 || item.AncestorPath[0].ID != root.ID || item.AncestorPath[1].ID != parent.ID {
			t.Fatalf("invalid path: %+v", item.AncestorPath)
		}
	}
}

func TestAncestorPathsRejectCorruptionAndCancellation(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		parent asset.Asset
	}{
		{"cross tenant", asset.Asset{ID: "parent", TenantID: "other", InventoryID: "inventory"}},
		{"cross inventory", asset.Asset{ID: "parent", TenantID: "tenant", InventoryID: "other"}},
		{"cycle", asset.Asset{ID: "parent", TenantID: "tenant", InventoryID: "inventory", ParentAssetID: "child"}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			repository := &pathRepository{items: map[asset.ID]asset.Asset{"parent": scenario.parent}, corrupt: true}
			_, err := WithAncestorPaths(context.Background(), repository, []ports.AssetSearchResult{pathResult("child", "parent")})
			if !errors.Is(err, apperrors.ErrInvalidInput) {
				t.Fatalf("expected safe rejection, got %v", err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repository := &pathRepository{}
	_, err := WithAncestorPaths(ctx, repository, []ports.AssetSearchResult{pathResult("child", "parent")})
	if !errors.Is(err, context.Canceled) || repository.reads != 0 {
		t.Fatalf("cancelled read reached repository: %v, %d", err, repository.reads)
	}
}

func TestAncestorPathsCheckCancellationForRootAndAfterRepositoryRead(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := WithAncestorPaths(ctx, &pathRepository{}, []ports.AssetSearchResult{pathResult("root", "")}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled root page succeeded: %v", err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	repository := &pathRepository{cancel: cancel, items: map[asset.ID]asset.Asset{"parent": {ID: "parent", TenantID: "tenant", InventoryID: "inventory"}}}
	if _, err := WithAncestorPaths(ctx, repository, []ports.AssetSearchResult{pathResult("child", "parent")}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled final read succeeded: %v", err)
	}
}

func pathResult(id, parent asset.ID) ports.AssetSearchResult {
	return ports.AssetSearchResult{TenantID: "tenant", Inventory: inventory.Inventory{ID: "inventory"}, Asset: asset.Asset{ID: id, TenantID: "tenant", InventoryID: "inventory", ParentAssetID: parent}}
}
