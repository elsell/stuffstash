package memory

import (
	"context"
	"sync"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type Authorizer struct {
	mu               sync.RWMutex
	tenantOwners     map[tenant.ID]map[identity.PrincipalID]struct{}
	inventoryOwners  map[inventory.InventoryID]map[identity.PrincipalID]struct{}
	inventoryEditors map[inventory.InventoryID]map[identity.PrincipalID]struct{}
	inventoryViewers map[inventory.InventoryID]map[identity.PrincipalID]struct{}
	inventoryTenants map[inventory.InventoryID]tenant.ID
}

func NewAuthorizer() *Authorizer {
	return &Authorizer{
		tenantOwners:     map[tenant.ID]map[identity.PrincipalID]struct{}{},
		inventoryOwners:  map[inventory.InventoryID]map[identity.PrincipalID]struct{}{},
		inventoryEditors: map[inventory.InventoryID]map[identity.PrincipalID]struct{}{},
		inventoryViewers: map[inventory.InventoryID]map[identity.PrincipalID]struct{}{},
		inventoryTenants: map[inventory.InventoryID]tenant.ID{},
	}
}

func (a *Authorizer) CheckTenant(_ context.Context, principal identity.Principal, permission ports.TenantPermission, tenantID tenant.ID) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	switch permission {
	case ports.TenantPermissionView, ports.TenantPermissionCreateInventory, ports.TenantPermissionConfigure:
		if hasPrincipal(a.tenantOwners[tenantID], principal.ID) {
			return nil
		}
		if permission == ports.TenantPermissionView {
			for inventoryID, inventoryTenantID := range a.inventoryTenants {
				if inventoryTenantID == tenantID &&
					(hasPrincipal(a.inventoryOwners[inventoryID], principal.ID) ||
						hasPrincipal(a.inventoryEditors[inventoryID], principal.ID) ||
						hasPrincipal(a.inventoryViewers[inventoryID], principal.ID)) {
					return nil
				}
			}
		}
	}
	return ports.ErrForbidden
}

func (a *Authorizer) CheckInventory(_ context.Context, principal identity.Principal, permission ports.InventoryPermission, inventoryID inventory.InventoryID) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	tenantID := a.inventoryTenants[inventoryID]
	switch permission {
	case ports.InventoryPermissionView:
		if hasPrincipal(a.tenantOwners[tenantID], principal.ID) || hasPrincipal(a.inventoryOwners[inventoryID], principal.ID) || hasPrincipal(a.inventoryEditors[inventoryID], principal.ID) || hasPrincipal(a.inventoryViewers[inventoryID], principal.ID) {
			return nil
		}
	case ports.InventoryPermissionCreateAsset, ports.InventoryPermissionEditAsset:
		if hasPrincipal(a.tenantOwners[tenantID], principal.ID) || hasPrincipal(a.inventoryOwners[inventoryID], principal.ID) || hasPrincipal(a.inventoryEditors[inventoryID], principal.ID) {
			return nil
		}
	case ports.InventoryPermissionShare, ports.InventoryPermissionConfigure:
		if hasPrincipal(a.tenantOwners[tenantID], principal.ID) || hasPrincipal(a.inventoryOwners[inventoryID], principal.ID) {
			return nil
		}
	}
	return ports.ErrForbidden
}

func (a *Authorizer) GrantTenantOwner(_ context.Context, principal identity.Principal, tenantID tenant.ID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.tenantOwners[tenantID] == nil {
		a.tenantOwners[tenantID] = map[identity.PrincipalID]struct{}{}
	}
	a.tenantOwners[tenantID][principal.ID] = struct{}{}
	return nil
}

func (a *Authorizer) GrantInventoryOwner(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.inventoryOwners[inventoryID] == nil {
		a.inventoryOwners[inventoryID] = map[identity.PrincipalID]struct{}{}
	}
	a.inventoryOwners[inventoryID][principal.ID] = struct{}{}
	a.inventoryTenants[inventoryID] = tenantID
	return nil
}

func (a *Authorizer) GrantInventoryViewer(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.inventoryViewers[inventoryID] == nil {
		a.inventoryViewers[inventoryID] = map[identity.PrincipalID]struct{}{}
	}
	a.inventoryViewers[inventoryID][principal.ID] = struct{}{}
	a.inventoryTenants[inventoryID] = tenantID
	return nil
}

func (a *Authorizer) GrantInventoryEditor(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.inventoryEditors[inventoryID] == nil {
		a.inventoryEditors[inventoryID] = map[identity.PrincipalID]struct{}{}
	}
	a.inventoryEditors[inventoryID][principal.ID] = struct{}{}
	a.inventoryTenants[inventoryID] = tenantID
	return nil
}

func (a *Authorizer) RevokeInventoryViewer(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.inventoryViewers[inventoryID], principal.ID)
	a.inventoryTenants[inventoryID] = tenantID
	return nil
}

func (a *Authorizer) RevokeInventoryEditor(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.inventoryEditors[inventoryID], principal.ID)
	a.inventoryTenants[inventoryID] = tenantID
	return nil
}

func hasPrincipal(principals map[identity.PrincipalID]struct{}, principalID identity.PrincipalID) bool {
	if principals == nil {
		return false
	}
	_, ok := principals[principalID]
	return ok
}
