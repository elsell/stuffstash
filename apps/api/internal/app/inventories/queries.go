package inventories

import (
	"context"
	"errors"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) ListInventories(ctx context.Context, input ListInventoriesInput) (ListInventoriesResult, error) {
	exists, err := a.tenants.TenantExists(ctx, input.TenantID)
	if err != nil {
		return ListInventoriesResult{}, err
	}
	if !exists {
		return ListInventoriesResult{}, apperrors.ErrNotFound
	}

	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionView, input.TenantID); err != nil {
		a.RecordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return ListInventoriesResult{}, err
	}

	limit := appsupport.PageLimit(a.defaultPageLimit, a.maxPageLimit, input.Limit)
	afterInventoryID, err := decodeInventoryCursor(input.TenantID, input.Cursor)
	if err != nil {
		return ListInventoriesResult{}, apperrors.ErrInvalidInput
	}

	visible := make([]inventory.Inventory, 0, limit+1)
	items, err := a.inventories.ListInventoriesByTenant(ctx, inventory.TenantID(input.TenantID.String()), ports.InventoryListPageRequest{
		AfterInventoryID: afterInventoryID,
		Limit:            inventoryScanLimit(limit),
	})
	if err != nil {
		return ListInventoriesResult{}, err
	}

	lastScannedID := inventory.InventoryID("")
	for _, item := range items {
		lastScannedID = item.ID
		err := a.authorizer.CheckInventory(ctx, input.Principal, ports.InventoryPermissionView, item.ID)
		if err == nil {
			visible = append(visible, item)
			continue
		}
		if !errors.Is(err, ports.ErrForbidden) {
			return ListInventoriesResult{}, err
		}
	}

	hasMore := len(visible) > limit
	var nextCursor *string
	if hasMore {
		visible = visible[:limit]
		nextCursor = encodeInventoryCursor(input.TenantID, visible[len(visible)-1].ID)
	} else if len(items) == inventoryScanLimit(limit) {
		hasMore = true
		nextCursor = encodeInventoryCursor(input.TenantID, lastScannedID)
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoriesListed,
		Message: "inventories listed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"principal_id": input.Principal.ID.String(),
			"limit":        strconv.Itoa(limit),
		},
	})
	if err := appsupport.SaveReadAuditRecord(ctx, a.audit, a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:  input.Principal,
		TenantID:   input.TenantID,
		Source:     input.Source,
		RequestID:  input.RequestID,
		Action:     audit.ActionInventoryListed,
		TargetType: audit.TargetTenant,
		TargetID:   input.TenantID.String(),
		Metadata: map[string]string{
			"limit": strconv.Itoa(limit),
		},
	}); err != nil {
		return ListInventoriesResult{}, err
	}

	return ListInventoriesResult{
		Items:      visible,
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func inventoryScanLimit(limit int) int {
	return limit*2 + 1
}

func encodeInventoryCursor(tenantID tenant.ID, id inventory.InventoryID) *string {
	return appsupport.EncodePageCursor("inventories", tenantID.String(), id.String())
}

func decodeInventoryCursor(tenantID tenant.ID, cursor string) (inventory.InventoryID, error) {
	decoded, err := appsupport.DecodePageCursor("inventories", tenantID.String(), cursor)
	if err != nil {
		return inventory.InventoryID(""), err
	}
	if decoded == "" {
		return inventory.InventoryID(""), nil
	}
	id, ok := inventory.NewID(decoded)
	if !ok {
		return inventory.InventoryID(""), apperrors.ErrInvalidInput
	}
	return id, nil
}
