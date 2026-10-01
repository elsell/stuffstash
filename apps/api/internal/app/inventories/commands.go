package inventories

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) CreateTenant(ctx context.Context, input CreateTenantInput) (tenant.Tenant, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return tenant.Tenant{}, apperrors.ErrInvalidInput
	}

	id := a.ids.NewID()

	tenantName, ok := tenant.NewName(name)
	if !ok {
		return tenant.Tenant{}, apperrors.ErrInvalidInput
	}
	tenantID, ok := tenant.NewID(id)
	if !ok {
		return tenant.Tenant{}, apperrors.ErrInvalidInput
	}

	item := tenant.Tenant{
		ID:   tenantID,
		Name: tenantName,
	}

	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:  input.Principal,
		TenantID:   item.ID,
		Source:     input.Source,
		RequestID:  input.RequestID,
		Action:     audit.ActionTenantCreated,
		TargetType: audit.TargetTenant,
		TargetID:   item.ID.String(),
		Metadata: map[string]string{
			"name": item.Name.String(),
		},
	})
	if err != nil {
		return tenant.Tenant{}, err
	}
	if err := a.outbox.SaveTenantAndEnqueueOwnerGrant(ctx, a.ids.NewID(), item, input.Principal, auditRecord); err != nil {
		return tenant.Tenant{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventTenantCreated,
		Message: "tenant created",
		Fields: map[string]string{
			"tenant_id":    item.ID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	a.DrainAuthorizationOutboxBestEffort(ctx, a.AuthorizationOutboxDrainLimit())

	return item, nil
}

func (a Service) CreateInventory(ctx context.Context, input CreateInventoryInput) (inventory.Inventory, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return inventory.Inventory{}, apperrors.ErrInvalidInput
	}

	exists, err := a.tenants.TenantExists(ctx, input.TenantID)
	if err != nil {
		return inventory.Inventory{}, err
	}
	if !exists {
		return inventory.Inventory{}, apperrors.ErrNotFound
	}

	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionCreateInventory, input.TenantID); err != nil {
		a.RecordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return inventory.Inventory{}, err
	}

	id := a.ids.NewID()
	inventoryID, ok := inventory.NewID(id)
	if !ok {
		return inventory.Inventory{}, apperrors.ErrInvalidInput
	}
	inventoryName, ok := inventory.NewName(name)
	if !ok {
		return inventory.Inventory{}, apperrors.ErrInvalidInput
	}

	item := inventory.Inventory{
		ID:       inventoryID,
		TenantID: inventory.TenantID(input.TenantID.String()),
		Name:     inventoryName,
	}

	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: item.ID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryCreated,
		TargetType:  audit.TargetInventory,
		TargetID:    item.ID.String(),
		Metadata: map[string]string{
			"name": item.Name.String(),
		},
	})
	if err != nil {
		return inventory.Inventory{}, err
	}
	if err := a.outbox.SaveInventoryAndEnqueueOwnerGrant(ctx, a.ids.NewID(), item, input.TenantID, input.Principal, auditRecord); err != nil {
		return inventory.Inventory{}, err
	}

	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryCreated,
		Message: "inventory created",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": item.ID.String(),
			"principal_id": input.Principal.ID.String(),
		},
	})
	a.DrainAuthorizationOutboxBestEffort(ctx, a.AuthorizationOutboxDrainLimit())

	return item, nil
}
