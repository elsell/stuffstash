package spicedb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

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

	gateway, err := NewGateway(endpoint, "", false)
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

	assertAllowed(t, authorizer.CheckTenant(ctx, ownerOne, ports.TenantPermissionCreateInventory, tenantOneID), "tenant owner can create inventory")
	assertAllowed(t, authorizer.CheckInventory(ctx, ownerOne, ports.InventoryPermissionView, inventoryOneID), "tenant owner can view tenant inventory")
	assertAllowed(t, authorizer.CheckInventory(ctx, ownerOne, ports.InventoryPermissionView, inventorySiblingID), "tenant owner can view sibling tenant inventory")
	assertAllowed(t, authorizer.CheckTenant(ctx, inventoryOwner, ports.TenantPermissionView, tenantOneID), "inventory owner can view containing tenant")
	assertAllowed(t, authorizer.CheckInventory(ctx, inventoryOwner, ports.InventoryPermissionView, inventoryOneID), "inventory owner can view owned inventory")

	assertForbidden(t, authorizer.CheckTenant(ctx, inventoryOwner, ports.TenantPermissionCreateInventory, tenantOneID), "inventory owner cannot create inventory")
	assertForbidden(t, authorizer.CheckInventory(ctx, inventoryOwner, ports.InventoryPermissionView, inventorySiblingID), "inventory owner cannot view sibling inventory")
	assertForbidden(t, authorizer.CheckTenant(ctx, unrelated, ports.TenantPermissionView, tenantOneID), "unrelated user cannot view tenant")
	assertForbidden(t, authorizer.CheckInventory(ctx, unrelated, ports.InventoryPermissionView, inventoryOneID), "unrelated user cannot view inventory")
	assertForbidden(t, authorizer.CheckTenant(ctx, ownerTwo, ports.TenantPermissionCreateInventory, tenantOneID), "other tenant owner cannot create inventory across tenant")
	assertForbidden(t, authorizer.CheckInventory(ctx, ownerTwo, ports.InventoryPermissionView, inventoryOneID), "other tenant owner cannot view inventory across tenant")
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
