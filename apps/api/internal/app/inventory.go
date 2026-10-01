package app

import (
	"context"
	"time"

	inventoryapp "github.com/stuffstash/stuff-stash/internal/app/inventories"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateTenantInput = inventoryapp.CreateTenantInput
type GetTenantInput = inventoryapp.GetTenantInput
type UpdateTenantInput = inventoryapp.UpdateTenantInput
type UpdateTenantLifecycleInput = inventoryapp.UpdateTenantLifecycleInput
type CreateInventoryInput = inventoryapp.CreateInventoryInput
type GetInventoryInput = inventoryapp.GetInventoryInput
type UpdateInventoryInput = inventoryapp.UpdateInventoryInput
type UpdateInventoryLifecycleInput = inventoryapp.UpdateInventoryLifecycleInput
type ListInventoriesInput = inventoryapp.ListInventoriesInput
type ListInventoriesResult = inventoryapp.ListInventoriesResult

func (a App) inventoryService() inventoryapp.Service {
	return inventoryapp.New(inventoryapp.Dependencies{
		InvitationAllowInsecureHTTP: a.invitationAllowInsecureHTTP,
		InvitationPublicBaseURL:     a.invitationPublicBaseURL,
		InvitationTTL:               a.invitationTTL,
		InventoryAccessUnitOfWork:   a.inventoryAccessUnitOfWork,
		InventoryAccess:             a.inventoryAccess,
		Observer:                    a.observer,
		Authorizer:                  a.authorizer,
		Tenants:                     a.tenants,
		TenantUnitOfWork:            a.tenantUnitOfWork,
		Inventories:                 a.inventories,
		InventoryUnitOfWork:         a.inventoryUnitOfWork,
		Audit:                       a.audit,
		Outbox:                      a.outbox,
		IDs:                         a.ids,
		Clock:                       a.clock,
		OutboxDrainLimit:            a.outboxDrainLimit,
		OutboxClaimLease:            a.outboxClaimLease,
		DefaultPageLimit:            a.defaultPageLimit,
		MaxPageLimit:                a.maxPageLimit,
	})
}

func (a App) CurrentPrincipal(principal identity.Principal) identity.Principal {
	return principal
}

func (a App) CreateTenant(ctx context.Context, input CreateTenantInput) (tenant.Tenant, error) {
	return a.inventoryService().CreateTenant(ctx, input)
}

func (a App) CreateInventory(ctx context.Context, input CreateInventoryInput) (inventory.Inventory, error) {
	return a.inventoryService().CreateInventory(ctx, input)
}

func (a App) drainAuthorizationOutboxBestEffort(ctx context.Context, limit int) {
	a.inventoryService().DrainAuthorizationOutboxBestEffort(ctx, limit)
}

func (a App) DrainAuthorizationOutbox(ctx context.Context, limit int) error {
	return a.inventoryService().DrainAuthorizationOutbox(ctx, limit)
}

func (a App) DrainAuthorizationOutboxEvent(ctx context.Context, eventID string) error {
	return a.inventoryService().DrainAuthorizationOutboxEvent(ctx, eventID)
}

func (a App) processClaimedAuthorizationOutboxEvent(ctx context.Context, event ports.AuthorizationOutboxEvent, claimID string) error {
	return a.inventoryService().ProcessClaimedAuthorizationOutboxEvent(ctx, event, claimID)
}

func ApplyAuthorizationOutboxEvent(ctx context.Context, authorizer ports.Authorizer, event ports.AuthorizationOutboxEvent) error {
	return inventoryapp.ApplyAuthorizationOutboxEvent(ctx, authorizer, event)
}

func (a App) authorizationOutboxDrainLimit() int {
	return a.inventoryService().AuthorizationOutboxDrainLimit()
}

func (a App) authorizationOutboxClaimLease() time.Duration {
	return a.inventoryService().AuthorizationOutboxClaimLease()
}

func (a App) ListInventories(ctx context.Context, input ListInventoriesInput) (ListInventoriesResult, error) {
	return a.inventoryService().ListInventories(ctx, input)
}

func (a App) recordAuthorizationDenied(ctx context.Context, principal identity.Principal, tenantID tenant.ID) {
	a.inventoryService().RecordAuthorizationDenied(ctx, principal, tenantID)
}

func (a App) GetTenant(ctx context.Context, input GetTenantInput) (tenant.Tenant, error) {
	return a.inventoryService().GetTenant(ctx, input)
}

func (a App) UpdateTenant(ctx context.Context, input UpdateTenantInput) (tenant.Tenant, error) {
	return a.inventoryService().UpdateTenant(ctx, input)
}

func (a App) ArchiveTenant(ctx context.Context, input UpdateTenantLifecycleInput) (tenant.Tenant, error) {
	return a.inventoryService().ArchiveTenant(ctx, input)
}

func (a App) RestoreTenant(ctx context.Context, input UpdateTenantLifecycleInput) (tenant.Tenant, error) {
	return a.inventoryService().RestoreTenant(ctx, input)
}

func (a App) DeleteTenant(ctx context.Context, input UpdateTenantLifecycleInput) error {
	return a.inventoryService().DeleteTenant(ctx, input)
}

func (a App) GetInventory(ctx context.Context, input GetInventoryInput) (inventory.Inventory, error) {
	return a.inventoryService().GetInventory(ctx, input)
}

func (a App) UpdateInventory(ctx context.Context, input UpdateInventoryInput) (inventory.Inventory, error) {
	return a.inventoryService().UpdateInventory(ctx, input)
}

func (a App) ArchiveInventory(ctx context.Context, input UpdateInventoryLifecycleInput) (inventory.Inventory, error) {
	return a.inventoryService().ArchiveInventory(ctx, input)
}

func (a App) RestoreInventory(ctx context.Context, input UpdateInventoryLifecycleInput) (inventory.Inventory, error) {
	return a.inventoryService().RestoreInventory(ctx, input)
}

func (a App) DeleteInventory(ctx context.Context, input UpdateInventoryLifecycleInput) error {
	return a.inventoryService().DeleteInventory(ctx, input)
}

func (a App) ensureActiveInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	return a.inventoryService().EnsureActiveInventoryAccess(ctx, principal, tenantID, inventoryID, permission)
}

func (a App) ensureInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	return a.inventoryService().EnsureInventoryAccess(ctx, principal, tenantID, inventoryID, permission)
}

func (a App) ensureTenantExists(ctx context.Context, tenantID tenant.ID) error {
	return a.inventoryService().EnsureTenantExists(ctx, tenantID)
}

func (a App) ensureInventoryAccessItem(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) (inventory.Inventory, error) {
	return a.inventoryService().EnsureInventoryAccessItem(ctx, principal, tenantID, inventoryID, permission)
}
