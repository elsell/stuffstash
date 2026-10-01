package inventories

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) ListInventoryAccessGrants(ctx context.Context, input ListInventoryAccessGrantsInput) (ListInventoryAccessGrantsResult, error) {
	if err := a.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return ListInventoryAccessGrantsResult{}, err
	}

	limit := appsupport.PageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterGrantKey, err := decodeInventoryAccessGrantCursor(input.TenantID, input.InventoryID, input.Cursor)
	if err != nil {
		return ListInventoryAccessGrantsResult{}, apperrors.ErrInvalidInput
	}

	items, err := a.inventoryAccess.ListInventoryAccessGrants(ctx, input.TenantID, input.InventoryID, ports.InventoryAccessGrantPageRequest{
		AfterGrantKey: afterGrantKey,
		Limit:         limit + 1,
	})
	if err != nil {
		return ListInventoryAccessGrantsResult{}, err
	}

	hasMore := len(items) > limit
	var nextCursor *string
	if hasMore {
		items = items[:limit]
		nextCursor = encodeInventoryAccessGrantCursor(input.TenantID, input.InventoryID, items[len(items)-1].CursorKey())
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryAccessListed,
		Message: "inventory access grants listed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"limit":        strconv.Itoa(limit),
		},
	})
	if err := appsupport.SaveReadAuditRecord(ctx, a.audit, a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryAccessGrantListed,
		TargetType:  audit.TargetInventory,
		TargetID:    input.InventoryID.String(),
		Metadata: map[string]string{
			"limit": strconv.Itoa(limit),
		},
	}); err != nil {
		return ListInventoryAccessGrantsResult{}, err
	}

	return ListInventoryAccessGrantsResult{
		Items:      items,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (a Service) GetInventoryAccessGrant(ctx context.Context, input GetInventoryAccessGrantInput) (ports.InventoryAccessGrant, error) {
	if err := a.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return ports.InventoryAccessGrant{}, err
	}
	targetPrincipalID, ok := identity.NewPrincipalID(input.TargetUserID)
	if !ok {
		return ports.InventoryAccessGrant{}, apperrors.ErrInvalidInput
	}
	relationship, ok := inventoryAccessRelationship(input.Relationship)
	if !ok {
		return ports.InventoryAccessGrant{}, apperrors.ErrInvalidInput
	}
	grant, found, err := a.inventoryAccess.InventoryAccessGrantByID(ctx, input.TenantID, input.InventoryID, targetPrincipalID, relationship)
	if err != nil {
		return ports.InventoryAccessGrant{}, err
	}
	if !found {
		return ports.InventoryAccessGrant{}, apperrors.ErrNotFound
	}
	if err := appsupport.SaveReadAuditRecord(ctx, a.audit, a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryAccessGrantViewed,
		TargetType:  audit.TargetInventoryAccessGrant,
		TargetID:    grant.CursorKey(),
		Metadata: map[string]string{
			"target_principal_id": targetPrincipalID.String(),
			"relationship":        string(relationship),
		},
	}); err != nil {
		return ports.InventoryAccessGrant{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryAccessViewed,
		Message: "inventory access grant viewed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"target_id":    targetPrincipalID.String(),
			"relationship": string(relationship),
		},
	})
	return grant, nil
}

func encodeInventoryAccessGrantCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, key string) *string {
	return appsupport.EncodePageCursor("inventory_access_grants", tenantID.String()+":"+inventoryID.String(), key)
}

func decodeInventoryAccessGrantCursor(tenantID tenant.ID, inventoryID inventory.InventoryID, cursor string) (string, error) {
	return appsupport.DecodePageCursor("inventory_access_grants", tenantID.String()+":"+inventoryID.String(), cursor)
}
