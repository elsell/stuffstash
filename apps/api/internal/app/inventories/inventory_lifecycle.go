package inventories

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) GetInventory(ctx context.Context, input GetInventoryInput) (inventory.Inventory, error) {
	if err := a.EnsureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return inventory.Inventory{}, err
	}
	item, found, err := a.inventories.InventoryByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !found || !item.IsActive() {
		return inventory.Inventory{}, apperrors.ErrNotFound
	}
	if err := appsupport.SaveReadAuditRecord(ctx, a.audit, a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryViewed,
		TargetType:  audit.TargetInventory,
		TargetID:    item.ID.String(),
		Metadata:    map[string]string{},
	}); err != nil {
		return inventory.Inventory{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryViewed,
		Message: "inventory viewed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": item.ID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	return item, nil
}

func (a Service) UpdateInventory(ctx context.Context, input UpdateInventoryInput) (inventory.Inventory, error) {
	if input.Name == nil {
		return inventory.Inventory{}, apperrors.ErrInvalidInput
	}
	if err := a.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return inventory.Inventory{}, err
	}
	current, found, err := a.inventories.InventoryByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !found || !current.IsActive() {
		return inventory.Inventory{}, apperrors.ErrNotFound
	}
	name, ok := inventory.NewName(*input.Name)
	if !ok {
		return inventory.Inventory{}, apperrors.ErrInvalidInput
	}
	updated := current
	updated.Name = name
	if updated.Name == current.Name {
		return current, nil
	}
	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryUpdated,
		TargetType:  audit.TargetInventory,
		TargetID:    updated.ID.String(),
		Metadata: map[string]string{
			"name": updated.Name.String(),
		},
	})
	if err != nil {
		return inventory.Inventory{}, err
	}
	if err := a.inventoryUnitOfWork.UpdateInventory(ctx, updated, auditRecord); err != nil {
		return inventory.Inventory{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryUpdated,
		Message: "inventory updated",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": updated.ID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	return updated, nil
}

func (a Service) ArchiveInventory(ctx context.Context, input UpdateInventoryLifecycleInput) (inventory.Inventory, error) {
	return a.updateInventoryLifecycle(ctx, input, inventory.LifecycleStateActive, inventory.LifecycleStateArchived, audit.ActionInventoryArchived, ports.EventInventoryArchived, "inventory archived")
}

func (a Service) RestoreInventory(ctx context.Context, input UpdateInventoryLifecycleInput) (inventory.Inventory, error) {
	return a.updateInventoryLifecycle(ctx, input, inventory.LifecycleStateArchived, inventory.LifecycleStateActive, audit.ActionInventoryRestored, ports.EventInventoryRestored, "inventory restored")
}

func (a Service) updateInventoryLifecycle(ctx context.Context, input UpdateInventoryLifecycleInput, from inventory.LifecycleState, to inventory.LifecycleState, action audit.Action, eventName ports.EventName, eventMessage string) (inventory.Inventory, error) {
	if err := a.EnsureTenantExists(ctx, input.TenantID); err != nil {
		return inventory.Inventory{}, err
	}
	item, found, err := a.inventories.InventoryByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !found {
		return inventory.Inventory{}, apperrors.ErrNotFound
	}
	if err := a.authorizer.CheckInventory(ctx, input.Principal, ports.InventoryPermissionConfigure, input.InventoryID); err != nil {
		a.RecordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return inventory.Inventory{}, err
	}
	if item.LifecycleState != from {
		return inventory.Inventory{}, apperrors.ErrInvalidInput
	}
	updated := item
	updated.LifecycleState = to
	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      action,
		TargetType:  audit.TargetInventory,
		TargetID:    updated.ID.String(),
		Metadata: map[string]string{
			"previous_state":  item.LifecycleState.String(),
			"lifecycle_state": updated.LifecycleState.String(),
		},
	})
	if err != nil {
		return inventory.Inventory{}, err
	}
	if err := a.inventoryUnitOfWork.UpdateInventoryLifecycle(ctx, updated, auditRecord); err != nil {
		return inventory.Inventory{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    eventName,
		Message: eventMessage,
		Fields: map[string]string{
			"tenant_id":       input.TenantID.String(),
			"inventory_id":    updated.ID.String(),
			"principal_id":    input.Principal.ID.String(),
			"lifecycle_state": updated.LifecycleState.String(),
		},
	})
	return updated, nil
}

func (a Service) DeleteInventory(ctx context.Context, input UpdateInventoryLifecycleInput) error {
	if err := a.EnsureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return err
	}
	item, found, err := a.inventories.InventoryByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return err
	}
	if !found {
		return apperrors.ErrNotFound
	}
	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryDeleted,
		TargetType:  audit.TargetInventory,
		TargetID:    item.ID.String(),
		Metadata: map[string]string{
			"lifecycle_state": item.LifecycleState.String(),
		},
	})
	if err != nil {
		return err
	}
	if err := a.inventoryUnitOfWork.DeleteInventory(ctx, input.TenantID, input.InventoryID, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return apperrors.ErrInvalidInput
		}
		return err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryDeleted,
		Message: "inventory deleted",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	return nil
}
