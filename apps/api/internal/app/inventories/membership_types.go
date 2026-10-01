package inventories

import (
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type GrantInventoryAccessInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	TargetUserID string
	Relationship string
}

type ListInventoryAccessGrantsInput struct {
	Principal   identity.Principal
	Source      audit.Source
	RequestID   string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Limit       int
	Cursor      string
}

type GetInventoryAccessGrantInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	TargetUserID string
	Relationship string
}

type ListInventoryAccessGrantsResult struct {
	Items      []ports.InventoryAccessGrant
	Limit      int
	NextCursor *string
	HasMore    bool
}

type RevokeInventoryAccessInput struct {
	Principal    identity.Principal
	Source       audit.Source
	RequestID    string
	TenantID     tenant.ID
	InventoryID  inventory.InventoryID
	TargetUserID string
	Relationship string
}
