package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestStoreUpdatesCustomAssetTypeMetadata(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tenantID, "Home")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Medicine")

	assetType := customAssetType(t, "01ARZ3NDEKTSV4RRFFQ69G5FAX", tenantID.String(), inventoryID.String(), customfield.ScopeInventory, "medicine")
	if err := saveCustomAssetType(t, ctx, store, assetType); err != nil {
		t.Fatalf("save custom asset type: %v", err)
	}
	displayName, ok := customfield.NewDisplayName("Medicine and Vitamins")
	if !ok {
		t.Fatalf("expected valid display name")
	}
	description, ok := customfield.NewDescription("Medication and supplement supplies")
	if !ok {
		t.Fatalf("expected valid description")
	}
	assetType.DisplayName = displayName
	assetType.Description = description
	if err := store.UpdateCustomAssetType(ctx, assetType, auditRecord(t, auditIDWithSuffix(assetType.ID.String(), "T"), tenantID, inventoryID, audit.ActionCustomAssetTypeUpdated)); err != nil {
		t.Fatalf("update custom asset type: %v", err)
	}

	found, ok, err := store.CustomAssetTypeByID(ctx, tenantID, inventoryID, assetType.ID)
	if err != nil {
		t.Fatalf("find custom asset type: %v", err)
	}
	if !ok || found.DisplayName != displayName || found.Description != description || found.Key != assetType.Key {
		t.Fatalf("expected updated custom asset type metadata, got %+v", found)
	}

	mutatedKey := assetType
	mutatedKey.Key = customfield.Key("changed")
	if err := store.UpdateCustomAssetType(ctx, mutatedKey, auditRecord(t, auditIDWithSuffix(assetType.ID.String(), "T"), tenantID, inventoryID, audit.ActionCustomAssetTypeUpdated)); !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected immutable key rejection, got %v", err)
	}
}
