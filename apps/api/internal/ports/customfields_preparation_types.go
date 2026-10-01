package ports

import (
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type PreparedCustomAssetType struct {
	Item        customfield.AssetType
	AuditRecord audit.Record
}
type PreparedCustomFieldDefinition struct {
	Item        customfield.Definition
	AuditRecord audit.Record
}
type CreateCustomAssetTypeInput struct {
	ExpirationEnabled bool
	Principal         identity.Principal
	Source            audit.Source
	RequestID         string
	TenantID          tenant.ID
	InventoryID       inventory.InventoryID
	Key               string
	DisplayName       string
	Description       string
}
type CreateCustomFieldDefinitionInput struct {
	Principal          identity.Principal
	Source             audit.Source
	RequestID          string
	TenantID           tenant.ID
	InventoryID        inventory.InventoryID
	Key                string
	DisplayName        string
	Type               string
	EnumOptions        []string
	Applicability      string
	CustomAssetTypeIDs []string
}
