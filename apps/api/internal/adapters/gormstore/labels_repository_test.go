package gormstore

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"testing"
	"time"
)

func TestLabelIdentityPersistsAndTombstonesWithAsset(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	tid := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	iid := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FAW")
	saveTenant(t, ctx, store, tid, "Home")
	saveInventory(t, ctx, store, iid.String(), tid, "Tools")
	item := assetItem("01ARZ3NDEKTSV4RRFFQ69G5FAX", tid.String(), iid.String(), asset.KindItem, "")
	if err := createAsset(t, ctx, store, item); err != nil {
		t.Fatal(err)
	}
	instance, err := store.BootstrapLabelInstance(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAY")
	if err != nil {
		t.Fatal(err)
	}
	restarted := NewStore(store.db)
	again, err := restarted.BootstrapLabelInstance(ctx, "01ARZ3NDEKTSV4RRFFQ69G5FAZ")
	if err != nil || again != instance {
		t.Fatal("instance changed on restart")
	}
	value := printing.Label{InstanceID: instance, ID: "01ARZ3NDEKTSV4RRFFQ69G5FB0", TenantID: tid.String(), InventoryID: iid.String(), AssetID: item.ID.String(), CreatedAt: time.Now()}
	record := auditRecord(t, "label-audit", tid, iid, audit.ActionLabelProvisioned)
	record.TargetID = item.ID.String()
	saved, created, err := store.ProvisionLabel(ctx, value, record)
	if err != nil || !created {
		t.Fatalf("provision: %v", err)
	}
	value.ID = "01ARZ3NDEKTSV4RRFFQ69G5FB1"
	repeated, created, err := restarted.ProvisionLabel(ctx, value, record)
	if err != nil || created || saved.ID != repeated.ID {
		t.Fatalf("idempotency: %v", err)
	}
	if _, found, err := store.LabelForAsset(ctx, tid, "other", item.ID); err != nil || found {
		t.Fatal("cross inventory discovery")
	}
	deletion := auditRecord(t, "delete-label-asset", tid, iid, audit.ActionAssetDeleted)
	if err := store.DeleteAsset(ctx, tid, iid, item.ID, deletion); err != nil {
		t.Fatal(err)
	}
	tombstone, found, err := store.LookupLabel(ctx, instance, saved.ID)
	if err != nil || !found || !tombstone.Tombstoned {
		t.Fatal("deleted label lost tombstone")
	}
}
