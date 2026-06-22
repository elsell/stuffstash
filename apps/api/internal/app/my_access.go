package app

import (
	"context"
	"errors"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type AccessRelationship string

const (
	AccessRelationshipOwner  AccessRelationship = "owner"
	AccessRelationshipEditor AccessRelationship = "editor"
	AccessRelationshipViewer AccessRelationship = "viewer"
)

type AccessSummary struct {
	Relationship AccessRelationship
	Permissions  []string
}

type MyTenantAccess struct {
	Tenant tenant.Tenant
	Access AccessSummary
}

type ListMyTenantsInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	Limit     int
	Cursor    string
}

type ListMyTenantsResult struct {
	Items      []MyTenantAccess
	Limit      int
	NextCursor *string
	HasMore    bool
}

func (a App) ListMyTenants(ctx context.Context, input ListMyTenantsInput) (ListMyTenantsResult, error) {
	limit := pageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterTenantID, err := decodeTenantCursor(input.Cursor)
	if err != nil {
		return ListMyTenantsResult{}, ErrInvalidInput
	}

	visible := make([]MyTenantAccess, 0, limit+1)
	items, err := a.tenants.ListTenants(ctx, ports.TenantListPageRequest{
		AfterTenantID: afterTenantID,
		Limit:         tenantScanLimit(limit),
	})
	if err != nil {
		return ListMyTenantsResult{}, err
	}

	lastScannedID := tenant.ID("")
	for _, item := range items {
		lastScannedID = item.ID
		access, ok, err := a.effectiveTenantAccess(ctx, input.Principal, item.ID)
		if err != nil {
			return ListMyTenantsResult{}, err
		}
		if !ok {
			continue
		}
		visible = append(visible, MyTenantAccess{Tenant: item, Access: access})
	}

	hasMore := len(visible) > limit
	var nextCursor *string
	if hasMore {
		visible = visible[:limit]
		nextCursor = encodeTenantCursor(visible[len(visible)-1].Tenant.ID)
	} else if len(items) == tenantScanLimit(limit) {
		hasMore = true
		nextCursor = encodeTenantCursor(lastScannedID)
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventTenantsListed,
		Message: "tenants listed",
		Fields: map[string]string{
			"principal_id": input.Principal.ID.String(),
			"limit":        strconv.Itoa(limit),
		},
	})
	for _, item := range visible {
		if err := a.saveReadAuditRecord(ctx, auditRecordInput{
			PrincipalID: input.Principal.ID,
			TenantID:    item.Tenant.ID,
			Source:      input.Source,
			RequestID:   input.RequestID,
			Action:      audit.ActionTenantListed,
			TargetType:  audit.TargetTenant,
			TargetID:    item.Tenant.ID.String(),
			Metadata: map[string]string{
				"limit": strconv.Itoa(limit),
			},
		}); err != nil {
			return ListMyTenantsResult{}, err
		}
	}

	return ListMyTenantsResult{
		Items:      visible,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (a App) TenantAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID) (AccessSummary, error) {
	access, ok, err := a.effectiveTenantAccess(ctx, principal, tenantID)
	if err != nil {
		return AccessSummary{}, err
	}
	if !ok {
		return AccessSummary{}, ports.ErrForbidden
	}
	return access, nil
}

func (a App) InventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) (AccessSummary, error) {
	if _, found, err := a.inventories.InventoryByID(ctx, tenantID, inventoryID); err != nil {
		return AccessSummary{}, err
	} else if !found {
		return AccessSummary{}, ErrNotFound
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

func (a App) effectiveTenantAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID) (AccessSummary, bool, error) {
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

func (a App) effectiveInventoryAccess(ctx context.Context, principal identity.Principal, inventoryID inventory.InventoryID) (AccessSummary, bool, error) {
	permissions := make([]string, 0, 5)
	for _, permission := range []ports.InventoryPermission{
		ports.InventoryPermissionView,
		ports.InventoryPermissionCreateAsset,
		ports.InventoryPermissionEditAsset,
		ports.InventoryPermissionShare,
		ports.InventoryPermissionConfigure,
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

func (a App) tenantPermissionAllowed(ctx context.Context, principal identity.Principal, tenantID tenant.ID, permission ports.TenantPermission) (bool, error) {
	err := a.authorizer.CheckTenant(ctx, principal, permission, tenantID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ports.ErrForbidden) {
		return false, nil
	}
	return false, err
}

func (a App) inventoryPermissionAllowed(ctx context.Context, principal identity.Principal, inventoryID inventory.InventoryID, permission ports.InventoryPermission) (bool, error) {
	err := a.authorizer.CheckInventory(ctx, principal, permission, inventoryID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ports.ErrForbidden) {
		return false, nil
	}
	return false, err
}

func tenantScanLimit(limit int) int {
	return limit*2 + 1
}

func encodeTenantCursor(id tenant.ID) *string {
	return encodePageCursor("my_tenants", "me", id.String())
}

func decodeTenantCursor(cursor string) (tenant.ID, error) {
	decoded, err := decodePageCursor("my_tenants", "me", cursor)
	if err != nil {
		return tenant.ID(""), err
	}
	if decoded == "" {
		return tenant.ID(""), nil
	}
	id, ok := tenant.NewID(decoded)
	if !ok {
		return tenant.ID(""), ErrInvalidInput
	}
	return id, nil
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
