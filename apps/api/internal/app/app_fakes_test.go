package app

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type fakeAuthorizer struct {
	checkInventoryErr      error
	checkTenantErr         error
	grantTenantOwnerErr    error
	tenantOwnerGrants      []string
	inventoryOwnerGrants   []string
	inventoryViewerGrants  []string
	inventoryEditorGrants  []string
	inventoryViewerRevokes []string
	inventoryEditorRevokes []string
}

func (f *fakeAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	if f.checkTenantErr != nil {
		return f.checkTenantErr
	}
	return nil
}

func (f *fakeAuthorizer) CheckInventory(context.Context, identity.Principal, ports.InventoryPermission, inventory.InventoryID) error {
	return f.checkInventoryErr
}

func (f *fakeAuthorizer) ListViewableInventoryIDs(_ context.Context, _ identity.Principal, _ tenant.ID, candidates []inventory.InventoryID) ([]inventory.InventoryID, error) {
	if f.checkInventoryErr != nil {
		if errors.Is(f.checkInventoryErr, ports.ErrForbidden) {
			return []inventory.InventoryID{}, nil
		}
		return nil, f.checkInventoryErr
	}
	return append([]inventory.InventoryID{}, candidates...), nil
}

func (f *fakeAuthorizer) GrantTenantOwner(_ context.Context, principal identity.Principal, tenantID tenant.ID) error {
	if f.grantTenantOwnerErr != nil {
		return f.grantTenantOwnerErr
	}
	f.tenantOwnerGrants = append(f.tenantOwnerGrants, principal.ID.String()+":"+tenantID.String())
	return nil
}

func (f *fakeAuthorizer) GrantInventoryOwner(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryOwnerGrants = append(f.inventoryOwnerGrants, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

func (f *fakeAuthorizer) GrantInventoryViewer(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryViewerGrants = append(f.inventoryViewerGrants, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

func (f *fakeAuthorizer) GrantInventoryEditor(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryEditorGrants = append(f.inventoryEditorGrants, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

func (f *fakeAuthorizer) RevokeInventoryViewer(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryViewerRevokes = append(f.inventoryViewerRevokes, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

func (f *fakeAuthorizer) RevokeInventoryEditor(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryEditorRevokes = append(f.inventoryEditorRevokes, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

type fakeTenantRepository struct {
	exists bool
}

func (f *fakeTenantRepository) SaveTenant(context.Context, tenant.Tenant) error {
	return nil
}

func (f *fakeTenantRepository) TenantExists(context.Context, tenant.ID) (bool, error) {
	return f.exists, nil
}

func (f *fakeTenantRepository) TenantByID(_ context.Context, tenantID tenant.ID) (tenant.Tenant, bool, error) {
	if !f.exists {
		return tenant.Tenant{}, false, nil
	}
	name, _ := tenant.NewName("Tenant")
	return tenant.Tenant{ID: tenantID, Name: name, LifecycleState: tenant.LifecycleStateActive}, true, nil
}

func (f *fakeTenantRepository) UpdateTenant(context.Context, tenant.Tenant, audit.Record) error {
	return nil
}

func (f *fakeTenantRepository) UpdateTenantLifecycle(context.Context, tenant.Tenant, audit.Record) error {
	return nil
}

func (f *fakeTenantRepository) DeleteTenant(context.Context, tenant.ID, audit.Record) error {
	return nil
}

type fakeInventoryRepository struct {
	items        []inventory.Inventory
	accessGrants []ports.InventoryAccessGrant
	invitations  []ports.InventoryAccessInvitation
	auditRecords []audit.Record
	outbox       *fakeOutbox
	calls        int
	limits       []int
}

func (f *fakeInventoryRepository) InventoryByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error) {
	for _, item := range f.items {
		if item.ID == inventoryID && item.TenantID == inventory.TenantID(tenantID.String()) {
			return item, true, nil
		}
	}
	return inventory.Inventory{}, false, nil
}
