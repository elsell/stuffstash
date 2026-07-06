package app

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) GetTenant(ctx context.Context, input GetTenantInput) (tenant.Tenant, error) {
	item, found, err := a.tenants.TenantByID(ctx, input.TenantID)
	if err != nil {
		return tenant.Tenant{}, err
	}
	if !found || !item.IsActive() {
		return tenant.Tenant{}, ErrNotFound
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionView, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return tenant.Tenant{}, err
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
		Principal:  input.Principal,
		TenantID:   input.TenantID,
		Source:     input.Source,
		RequestID:  input.RequestID,
		Action:     audit.ActionTenantViewed,
		TargetType: audit.TargetTenant,
		TargetID:   item.ID.String(),
		Metadata:   map[string]string{},
	}); err != nil {
		return tenant.Tenant{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventTenantViewed,
		Message: "tenant viewed",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	return item, nil
}

func (a App) UpdateTenant(ctx context.Context, input UpdateTenantInput) (tenant.Tenant, error) {
	if input.Name == nil {
		return tenant.Tenant{}, ErrInvalidInput
	}
	current, found, err := a.tenants.TenantByID(ctx, input.TenantID)
	if err != nil {
		return tenant.Tenant{}, err
	}
	if !found || !current.IsActive() {
		return tenant.Tenant{}, ErrNotFound
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return tenant.Tenant{}, err
	}
	name, ok := tenant.NewName(*input.Name)
	if !ok {
		return tenant.Tenant{}, ErrInvalidInput
	}
	updated := current
	updated.Name = name
	if updated.Name == current.Name {
		return current, nil
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		Principal:  input.Principal,
		TenantID:   input.TenantID,
		Source:     input.Source,
		RequestID:  input.RequestID,
		Action:     audit.ActionTenantUpdated,
		TargetType: audit.TargetTenant,
		TargetID:   updated.ID.String(),
		Metadata: map[string]string{
			"name": updated.Name.String(),
		},
	})
	if err != nil {
		return tenant.Tenant{}, err
	}
	if err := a.tenantUnitOfWork.UpdateTenant(ctx, updated, auditRecord); err != nil {
		return tenant.Tenant{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventTenantUpdated,
		Message: "tenant updated",
		Fields: map[string]string{
			"tenant_id":    updated.ID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	return updated, nil
}

func (a App) ArchiveTenant(ctx context.Context, input UpdateTenantLifecycleInput) (tenant.Tenant, error) {
	return a.updateTenantLifecycle(ctx, input, tenant.LifecycleStateActive, tenant.LifecycleStateArchived, audit.ActionTenantArchived, ports.EventTenantArchived, "tenant archived")
}

func (a App) RestoreTenant(ctx context.Context, input UpdateTenantLifecycleInput) (tenant.Tenant, error) {
	return a.updateTenantLifecycle(ctx, input, tenant.LifecycleStateArchived, tenant.LifecycleStateActive, audit.ActionTenantRestored, ports.EventTenantRestored, "tenant restored")
}

func (a App) updateTenantLifecycle(ctx context.Context, input UpdateTenantLifecycleInput, from tenant.LifecycleState, to tenant.LifecycleState, action audit.Action, eventName ports.EventName, eventMessage string) (tenant.Tenant, error) {
	current, found, err := a.tenants.TenantByID(ctx, input.TenantID)
	if err != nil {
		return tenant.Tenant{}, err
	}
	if !found {
		return tenant.Tenant{}, ErrNotFound
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return tenant.Tenant{}, err
	}
	if current.LifecycleState != from {
		return tenant.Tenant{}, ErrInvalidInput
	}
	updated := current
	updated.LifecycleState = to
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		Principal:  input.Principal,
		TenantID:   input.TenantID,
		Source:     input.Source,
		RequestID:  input.RequestID,
		Action:     action,
		TargetType: audit.TargetTenant,
		TargetID:   updated.ID.String(),
		Metadata: map[string]string{
			"previous_state":  current.LifecycleState.String(),
			"lifecycle_state": updated.LifecycleState.String(),
		},
	})
	if err != nil {
		return tenant.Tenant{}, err
	}
	if err := a.tenantUnitOfWork.UpdateTenantLifecycle(ctx, updated, auditRecord); err != nil {
		return tenant.Tenant{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    eventName,
		Message: eventMessage,
		Fields: map[string]string{
			"tenant_id":       updated.ID.String(),
			"principal_id":    input.Principal.ID.String(),
			"lifecycle_state": updated.LifecycleState.String(),
		},
	})
	return updated, nil
}

func (a App) DeleteTenant(ctx context.Context, input UpdateTenantLifecycleInput) error {
	current, found, err := a.tenants.TenantByID(ctx, input.TenantID)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFound
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return err
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
		Principal:  input.Principal,
		TenantID:   input.TenantID,
		Source:     input.Source,
		RequestID:  input.RequestID,
		Action:     audit.ActionTenantDeleted,
		TargetType: audit.TargetTenant,
		TargetID:   current.ID.String(),
		Metadata: map[string]string{
			"lifecycle_state": current.LifecycleState.String(),
		},
	})
	if err != nil {
		return err
	}
	if err := a.tenantUnitOfWork.DeleteTenant(ctx, input.TenantID, auditRecord); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return ErrInvalidInput
		}
		return err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventTenantDeleted,
		Message: "tenant deleted",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	return nil
}

func (a App) GetInventory(ctx context.Context, input GetInventoryInput) (inventory.Inventory, error) {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return inventory.Inventory{}, err
	}
	item, found, err := a.inventories.InventoryByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !found || !item.IsActive() {
		return inventory.Inventory{}, ErrNotFound
	}
	if err := a.saveReadAuditRecord(ctx, auditRecordInput{
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

func (a App) UpdateInventory(ctx context.Context, input UpdateInventoryInput) (inventory.Inventory, error) {
	if input.Name == nil {
		return inventory.Inventory{}, ErrInvalidInput
	}
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return inventory.Inventory{}, err
	}
	current, found, err := a.inventories.InventoryByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !found || !current.IsActive() {
		return inventory.Inventory{}, ErrNotFound
	}
	name, ok := inventory.NewName(*input.Name)
	if !ok {
		return inventory.Inventory{}, ErrInvalidInput
	}
	updated := current
	updated.Name = name
	if updated.Name == current.Name {
		return current, nil
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
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

func (a App) ArchiveInventory(ctx context.Context, input UpdateInventoryLifecycleInput) (inventory.Inventory, error) {
	return a.updateInventoryLifecycle(ctx, input, inventory.LifecycleStateActive, inventory.LifecycleStateArchived, audit.ActionInventoryArchived, ports.EventInventoryArchived, "inventory archived")
}

func (a App) RestoreInventory(ctx context.Context, input UpdateInventoryLifecycleInput) (inventory.Inventory, error) {
	return a.updateInventoryLifecycle(ctx, input, inventory.LifecycleStateArchived, inventory.LifecycleStateActive, audit.ActionInventoryRestored, ports.EventInventoryRestored, "inventory restored")
}

func (a App) updateInventoryLifecycle(ctx context.Context, input UpdateInventoryLifecycleInput, from inventory.LifecycleState, to inventory.LifecycleState, action audit.Action, eventName ports.EventName, eventMessage string) (inventory.Inventory, error) {
	if err := a.ensureTenantExists(ctx, input.TenantID); err != nil {
		return inventory.Inventory{}, err
	}
	item, found, err := a.inventories.InventoryByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !found {
		return inventory.Inventory{}, ErrNotFound
	}
	if err := a.authorizer.CheckInventory(ctx, input.Principal, ports.InventoryPermissionConfigure, input.InventoryID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return inventory.Inventory{}, err
	}
	if item.LifecycleState != from {
		return inventory.Inventory{}, ErrInvalidInput
	}
	updated := item
	updated.LifecycleState = to
	auditRecord, err := a.newAuditRecord(auditRecordInput{
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

func (a App) DeleteInventory(ctx context.Context, input UpdateInventoryLifecycleInput) error {
	if err := a.ensureInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionConfigure); err != nil {
		return err
	}
	item, found, err := a.inventories.InventoryByID(ctx, input.TenantID, input.InventoryID)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFound
	}
	auditRecord, err := a.newAuditRecord(auditRecordInput{
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
			return ErrInvalidInput
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
