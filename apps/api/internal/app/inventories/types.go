package inventories

import (
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type CreateTenantInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	Name      string
}

type GetTenantInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	TenantID  tenant.ID
}

type UpdateTenantInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	TenantID  tenant.ID
	Name      *string
}

type UpdateTenantLifecycleInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	TenantID  tenant.ID
}

type CreateInventoryInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	TenantID  tenant.ID
	Name      string
}

type GetInventoryInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
}

type UpdateInventoryInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Name        *string
}

type UpdateInventoryLifecycleInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
}

type ListInventoriesInput struct {
	Principal identity.Principal
	Source    audit.Source
	RequestID string
	TenantID  tenant.ID
	Limit     int
	Cursor    string
}

type ListInventoriesResult struct {
	Items      []inventory.Inventory
	Limit      int
	NextCursor *string
	HasMore    bool
}
