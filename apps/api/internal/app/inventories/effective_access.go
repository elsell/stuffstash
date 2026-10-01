package inventories

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) TenantAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID) (AccessSummary, error) {
	access, ok, err := a.effectiveTenantAccess(ctx, principal, tenantID)
	if err != nil {
		return AccessSummary{}, err
	}
	if !ok {
		return AccessSummary{}, ports.ErrForbidden
	}
	return access, nil
}

func (a Service) InventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) (AccessSummary, error) {
	if _, found, err := a.inventories.InventoryByID(ctx, tenantID, inventoryID); err != nil {
		return AccessSummary{}, err
	} else if !found {
		return AccessSummary{}, apperrors.ErrNotFound
	}

	access, ok, err := a.effectiveInventoryAccess(ctx, principal, inventoryID)
	if err != nil {
		return AccessSummary{}, err
	}
	if !ok {
		return AccessSummary{}, ports.ErrForbidden
	}
	return access, nil
}

func (a Service) effectiveTenantAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID) (AccessSummary, bool, error) {
	permissions := make([]string, 0, 3)
	for _, permission := range []ports.TenantPermission{
		ports.TenantPermissionView,
		ports.TenantPermissionCreateInventory,
		ports.TenantPermissionConfigure,
	} {
		allowed, err := a.tenantPermissionAllowed(ctx, principal, tenantID, permission)
		if err != nil {
			return AccessSummary{}, false, err
		}
		if allowed {
			permissions = append(permissions, string(permission))
		}
	}
	if len(permissions) == 0 {
		return AccessSummary{}, false, nil
	}
	relationship := AccessRelationshipViewer
	if containsString(permissions, string(ports.TenantPermissionConfigure)) || containsString(permissions, string(ports.TenantPermissionCreateInventory)) {
		relationship = AccessRelationshipOwner
	}
	return AccessSummary{Relationship: relationship, Permissions: permissions}, true, nil
}

func (a Service) effectiveInventoryAccess(ctx context.Context, principal identity.Principal, inventoryID inventory.InventoryID) (AccessSummary, bool, error) {
	permissions := make([]string, 0, 7)
	for _, permission := range []ports.InventoryPermission{
		ports.InventoryPermissionView,
		ports.InventoryPermissionCreateAsset,
		ports.InventoryPermissionEditAsset,
		ports.InventoryPermissionShare,
		ports.InventoryPermissionConfigure,
		ports.InventoryPermissionViewImportJob,
		ports.InventoryPermissionCreateImportJob,
	} {
		allowed, err := a.inventoryPermissionAllowed(ctx, principal, inventoryID, permission)
		if err != nil {
			return AccessSummary{}, false, err
		}
		if allowed {
			permissions = append(permissions, string(permission))
		}
	}
	if len(permissions) == 0 {
		return AccessSummary{}, false, nil
	}
	relationship := AccessRelationshipViewer
	if containsString(permissions, string(ports.InventoryPermissionShare)) || containsString(permissions, string(ports.InventoryPermissionConfigure)) {
		relationship = AccessRelationshipOwner
	} else if containsString(permissions, string(ports.InventoryPermissionCreateAsset)) || containsString(permissions, string(ports.InventoryPermissionEditAsset)) {
		relationship = AccessRelationshipEditor
	}
	return AccessSummary{Relationship: relationship, Permissions: permissions}, true, nil
}

func (a Service) tenantPermissionAllowed(ctx context.Context, principal identity.Principal, tenantID tenant.ID, permission ports.TenantPermission) (bool, error) {
	err := a.authorizer.CheckTenant(ctx, principal, permission, tenantID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ports.ErrForbidden) {
		return false, nil
	}
	return false, err
}

func (a Service) inventoryPermissionAllowed(ctx context.Context, principal identity.Principal, inventoryID inventory.InventoryID, permission ports.InventoryPermission) (bool, error) {
	err := a.authorizer.CheckInventory(ctx, principal, permission, inventoryID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ports.ErrForbidden) {
		return false, nil
	}
	return false, err
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
