package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func TestCreateAssetTagReusesExistingActiveTagByKey(t *testing.T) {
	assets := &fakeAssetRepository{}
	application := New(Dependencies{
		Observer:           &fakeObserver{},
		Authorizer:         &fakeAuthorizer{},
		Tenants:            &fakeTenantRepository{exists: true},
		Inventories:        &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		AssetTags:          assets,
		AssetTagUnitOfWork: assets,
		Audit:              &fakeAuditRepository{},
		Outbox:             &fakeOutbox{},
		IDs:                &fakeIDGenerator{ids: []string{"tag-one", "audit-tag-one"}},
	})

	first, err := application.CreateAssetTag(context.Background(), CreateAssetTagInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		DisplayName: "Workshop",
		Color:       "#2f80ed",
	})
	if err != nil {
		t.Fatalf("create first tag: %v", err)
	}
	second, err := application.CreateAssetTag(context.Background(), CreateAssetTagInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		DisplayName: "workshop",
		Color:       "#00aa88",
	})
	if err != nil {
		t.Fatalf("reuse existing tag: %v", err)
	}
	if second.ID != first.ID || second.DisplayName != first.DisplayName || second.Color != first.Color {
		t.Fatalf("expected existing active tag, got first=%+v second=%+v", first, second)
	}
	if len(assets.assetTags) != 1 {
		t.Fatalf("expected one stored tag, got %+v", assets.assetTags)
	}
}

func TestCreateAssetTagRejectsArchivedDuplicateKey(t *testing.T) {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	key, _ := assettag.NewKey("workshop")
	displayName, _ := assettag.NewDisplayName("Workshop")
	tag, _ := assettag.NewTag("tag-archived", "tenant-one", "inventory-one", key, displayName, "", now)
	tag.LifecycleState = assettag.LifecycleStateArchived
	assets := &fakeAssetRepository{assetTags: map[assettag.ID]assettag.Tag{tag.ID: tag}}
	application := New(Dependencies{
		Observer:           &fakeObserver{},
		Authorizer:         &fakeAuthorizer{},
		Tenants:            &fakeTenantRepository{exists: true},
		Inventories:        &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		AssetTags:          assets,
		AssetTagUnitOfWork: assets,
		Audit:              &fakeAuditRepository{},
		Outbox:             &fakeOutbox{},
	})

	if _, err := application.CreateAssetTag(context.Background(), CreateAssetTagInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		DisplayName: "Workshop",
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input for archived duplicate key, got %v", err)
	}
}
