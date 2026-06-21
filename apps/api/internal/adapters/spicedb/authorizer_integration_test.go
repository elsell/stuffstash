package spicedb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const integrationEndpointEnv = "STUFF_STASH_SPICEDB_INTEGRATION_ENDPOINT"

func TestSpiceDBIntegrationEnforcesTenantAndInventoryRelationships(t *testing.T) {
	endpoint := os.Getenv(integrationEndpointEnv)
	if endpoint == "" {
		t.Skipf("%s is not set", integrationEndpointEnv)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	gateway, err := NewGateway(endpoint, "", false, "")
	if err != nil {
		t.Fatalf("create spicedb gateway: %v", err)
	}
	t.Cleanup(func() {
		if err := gateway.Close(); err != nil {
			t.Fatalf("close spicedb gateway: %v", err)
		}
	})

	authorizer := NewAuthorizer(gateway)
	if err := bootstrapIntegrationSchema(ctx, authorizer); err != nil {
		t.Fatalf("bootstrap schema: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	tenantOneID := tenant.ID("tenant-one-" + suffix)
	tenantTwoID := tenant.ID("tenant-two-" + suffix)
	inventoryOneID := inventory.InventoryID("inventory-one-" + suffix)
	inventorySiblingID := inventory.InventoryID("inventory-sibling-" + suffix)
	inventoryTwoID := inventory.InventoryID("inventory-two-" + suffix)
	ownerOne := principal("owner-one-" + suffix)
	ownerTwo := principal("owner-two-" + suffix)
	inventoryOwner := principal("inventory-owner-" + suffix)
	siblingOwner := principal("sibling-owner-" + suffix)
	inventoryViewer := principal("inventory-viewer-" + suffix)
	inventoryEditor := principal("inventory-editor-" + suffix)
	unrelated := principal("unrelated-" + suffix)

	if err := authorizer.GrantTenantOwner(ctx, ownerOne, tenantOneID); err != nil {
		t.Fatalf("grant tenant one owner: %v", err)
	}
	if err := authorizer.GrantTenantOwner(ctx, ownerTwo, tenantTwoID); err != nil {
		t.Fatalf("grant tenant two owner: %v", err)
	}
	if err := authorizer.GrantInventoryOwner(ctx, inventoryOwner, tenantOneID, inventoryOneID); err != nil {
		t.Fatalf("grant inventory owner: %v", err)
	}
	if err := authorizer.GrantInventoryOwner(ctx, siblingOwner, tenantOneID, inventorySiblingID); err != nil {
		t.Fatalf("grant sibling inventory owner: %v", err)
	}
	if err := authorizer.GrantInventoryOwner(ctx, ownerTwo, tenantTwoID, inventoryTwoID); err != nil {
		t.Fatalf("grant second inventory owner: %v", err)
	}
	if err := authorizer.GrantInventoryViewer(ctx, inventoryViewer, tenantOneID, inventoryOneID); err != nil {
		t.Fatalf("grant inventory viewer: %v", err)
	}
	if err := authorizer.GrantInventoryEditor(ctx, inventoryEditor, tenantOneID, inventoryOneID); err != nil {
		t.Fatalf("grant inventory editor: %v", err)
	}

	assertAllowed(t, authorizer.CheckTenant(ctx, ownerOne, ports.TenantPermissionCreateInventory, tenantOneID), "tenant owner can create inventory")
	assertAllowed(t, authorizer.CheckInventory(ctx, ownerOne, ports.InventoryPermissionView, inventoryOneID), "tenant owner can view tenant inventory")
	assertAllowed(t, authorizer.CheckInventory(ctx, ownerOne, ports.InventoryPermissionCreateAsset, inventoryOneID), "tenant owner can create assets in tenant inventory")
	assertAllowed(t, authorizer.CheckInventory(ctx, ownerOne, ports.InventoryPermissionEditAsset, inventoryOneID), "tenant owner can edit assets in tenant inventory")
	assertAllowed(t, authorizer.CheckInventory(ctx, ownerOne, ports.InventoryPermissionShare, inventoryOneID), "tenant owner can share tenant inventory")
	assertAllowed(t, authorizer.CheckInventory(ctx, ownerOne, ports.InventoryPermissionView, inventorySiblingID), "tenant owner can view sibling tenant inventory")
	assertAllowed(t, authorizer.CheckTenant(ctx, inventoryOwner, ports.TenantPermissionView, tenantOneID), "inventory owner can view containing tenant")
	assertAllowed(t, authorizer.CheckInventory(ctx, inventoryOwner, ports.InventoryPermissionView, inventoryOneID), "inventory owner can view owned inventory")
	assertAllowed(t, authorizer.CheckInventory(ctx, inventoryOwner, ports.InventoryPermissionCreateAsset, inventoryOneID), "inventory owner can create assets in owned inventory")
	assertAllowed(t, authorizer.CheckInventory(ctx, inventoryOwner, ports.InventoryPermissionEditAsset, inventoryOneID), "inventory owner can edit assets in owned inventory")
	assertAllowed(t, authorizer.CheckInventory(ctx, inventoryOwner, ports.InventoryPermissionShare, inventoryOneID), "inventory owner can share owned inventory")
	assertAllowed(t, authorizer.CheckTenant(ctx, inventoryViewer, ports.TenantPermissionView, tenantOneID), "inventory viewer can view containing tenant")
	assertAllowed(t, authorizer.CheckInventory(ctx, inventoryViewer, ports.InventoryPermissionView, inventoryOneID), "inventory viewer can view inventory")
	assertAllowed(t, authorizer.CheckTenant(ctx, inventoryEditor, ports.TenantPermissionView, tenantOneID), "inventory editor can view containing tenant")
	assertAllowed(t, authorizer.CheckInventory(ctx, inventoryEditor, ports.InventoryPermissionView, inventoryOneID), "inventory editor can view inventory")
	assertAllowed(t, authorizer.CheckInventory(ctx, inventoryEditor, ports.InventoryPermissionCreateAsset, inventoryOneID), "inventory editor can create assets")
	assertAllowed(t, authorizer.CheckInventory(ctx, inventoryEditor, ports.InventoryPermissionEditAsset, inventoryOneID), "inventory editor can edit assets")

	assertForbidden(t, authorizer.CheckTenant(ctx, inventoryOwner, ports.TenantPermissionCreateInventory, tenantOneID), "inventory owner cannot create inventory")
	assertForbidden(t, authorizer.CheckInventory(ctx, inventoryOwner, ports.InventoryPermissionView, inventorySiblingID), "inventory owner cannot view sibling inventory")
	assertForbidden(t, authorizer.CheckInventory(ctx, inventoryOwner, ports.InventoryPermissionCreateAsset, inventorySiblingID), "inventory owner cannot create assets in sibling inventory")
	assertForbidden(t, authorizer.CheckInventory(ctx, inventoryOwner, ports.InventoryPermissionEditAsset, inventorySiblingID), "inventory owner cannot edit assets in sibling inventory")
	assertForbidden(t, authorizer.CheckTenant(ctx, inventoryViewer, ports.TenantPermissionCreateInventory, tenantOneID), "inventory viewer cannot create inventory")
	assertForbidden(t, authorizer.CheckInventory(ctx, inventoryViewer, ports.InventoryPermissionCreateAsset, inventoryOneID), "inventory viewer cannot create assets")
	assertForbidden(t, authorizer.CheckInventory(ctx, inventoryViewer, ports.InventoryPermissionEditAsset, inventoryOneID), "inventory viewer cannot edit assets")
	assertForbidden(t, authorizer.CheckInventory(ctx, inventoryViewer, ports.InventoryPermissionShare, inventoryOneID), "inventory viewer cannot share inventory")
	assertForbidden(t, authorizer.CheckTenant(ctx, inventoryEditor, ports.TenantPermissionCreateInventory, tenantOneID), "inventory editor cannot create inventory")
	assertForbidden(t, authorizer.CheckInventory(ctx, inventoryEditor, ports.InventoryPermissionShare, inventoryOneID), "inventory editor cannot share inventory")
	assertForbidden(t, authorizer.CheckTenant(ctx, unrelated, ports.TenantPermissionView, tenantOneID), "unrelated user cannot view tenant")
	assertForbidden(t, authorizer.CheckInventory(ctx, unrelated, ports.InventoryPermissionView, inventoryOneID), "unrelated user cannot view inventory")
	assertForbidden(t, authorizer.CheckInventory(ctx, unrelated, ports.InventoryPermissionCreateAsset, inventoryOneID), "unrelated user cannot create assets")
	assertForbidden(t, authorizer.CheckInventory(ctx, unrelated, ports.InventoryPermissionEditAsset, inventoryOneID), "unrelated user cannot edit assets")
	assertForbidden(t, authorizer.CheckInventory(ctx, unrelated, ports.InventoryPermissionShare, inventoryOneID), "unrelated user cannot share inventory")
	assertForbidden(t, authorizer.CheckTenant(ctx, ownerTwo, ports.TenantPermissionCreateInventory, tenantOneID), "other tenant owner cannot create inventory across tenant")
	assertForbidden(t, authorizer.CheckInventory(ctx, ownerTwo, ports.InventoryPermissionView, inventoryOneID), "other tenant owner cannot view inventory across tenant")
	assertForbidden(t, authorizer.CheckInventory(ctx, ownerTwo, ports.InventoryPermissionCreateAsset, inventoryOneID), "other tenant owner cannot create assets across tenant")
	assertForbidden(t, authorizer.CheckInventory(ctx, ownerTwo, ports.InventoryPermissionEditAsset, inventoryOneID), "other tenant owner cannot edit assets across tenant")

	assertViewableInventoryIDs(t, authorizer, ctx, ownerOne, tenantOneID, []inventory.InventoryID{inventoryOneID, inventoryTwoID, inventorySiblingID}, []inventory.InventoryID{inventoryOneID, inventorySiblingID}, "tenant owner can list tenant-visible candidate inventories")
	assertViewableInventoryIDs(t, authorizer, ctx, inventoryOwner, tenantOneID, []inventory.InventoryID{inventorySiblingID, inventoryOneID, inventoryTwoID}, []inventory.InventoryID{inventoryOneID}, "inventory owner can list owned candidate inventory")
	assertViewableInventoryIDs(t, authorizer, ctx, unrelated, tenantOneID, []inventory.InventoryID{inventoryOneID, inventorySiblingID}, nil, "unrelated user cannot list candidate inventories")

	if err := authorizer.RevokeInventoryViewer(ctx, inventoryViewer, tenantOneID, inventoryOneID); err != nil {
		t.Fatalf("revoke inventory viewer: %v", err)
	}
	if err := authorizer.RevokeInventoryEditor(ctx, inventoryEditor, tenantOneID, inventoryOneID); err != nil {
		t.Fatalf("revoke inventory editor: %v", err)
	}
	assertForbidden(t, authorizer.CheckInventory(ctx, inventoryViewer, ports.InventoryPermissionView, inventoryOneID), "revoked inventory viewer cannot view inventory")
	assertAllowed(t, authorizer.CheckTenant(ctx, inventoryViewer, ports.TenantPermissionView, tenantOneID), "revoked inventory viewer keeps tenant viewer relationship")
	assertForbidden(t, authorizer.CheckInventory(ctx, inventoryEditor, ports.InventoryPermissionView, inventoryOneID), "revoked inventory editor cannot view inventory")
	assertForbidden(t, authorizer.CheckInventory(ctx, inventoryEditor, ports.InventoryPermissionCreateAsset, inventoryOneID), "revoked inventory editor cannot create assets")
}

func bootstrapIntegrationSchema(ctx context.Context, authorizer Authorizer) error {
	schema, err := os.ReadFile(integrationSchemaPath())
	if err != nil {
		return err
	}

	deadline := time.Now().Add(20 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := authorizer.BootstrapSchema(ctx, string(schema)); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(250 * time.Millisecond)
	}

	return lastErr
}

func integrationSchemaPath() string {
	if path := os.Getenv("STUFF_STASH_SPICEDB_INTEGRATION_SCHEMA_PATH"); path != "" {
		return path
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return "deploy/spicedb/schema.zed"
	}
	for directory := workingDirectory; ; directory = filepath.Dir(directory) {
		candidate := filepath.Join(directory, "deploy", "spicedb", "schema.zed")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "deploy/spicedb/schema.zed"
		}
	}
}

func assertAllowed(t *testing.T, err error, behavior string) {
	t.Helper()

	if err != nil {
		t.Fatalf("%s: expected allowed, got %v", behavior, err)
	}
}

func assertForbidden(t *testing.T, err error, behavior string) {
	t.Helper()

	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("%s: expected forbidden, got %v", behavior, err)
	}
}

func assertViewableInventoryIDs(t *testing.T, authorizer Authorizer, ctx context.Context, principal identity.Principal, tenantID tenant.ID, candidates []inventory.InventoryID, expected []inventory.InventoryID, behavior string) {
	t.Helper()

	visible, err := authorizer.ListViewableInventoryIDs(ctx, principal, tenantID, candidates)
	if err != nil {
		t.Fatalf("%s: list viewable inventory ids: %v", behavior, err)
	}
	if len(visible) != len(expected) {
		t.Fatalf("%s: expected %d visible inventories, got %d: %#v", behavior, len(expected), len(visible), visible)
	}
	for index := range expected {
		if visible[index] != expected[index] {
			t.Fatalf("%s: expected visible[%d] %q, got %q", behavior, index, expected[index], visible[index])
		}
	}
}
