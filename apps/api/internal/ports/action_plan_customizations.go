package ports

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

// Each operation commits the customization, audit and plan transition atomically.
type ActionPlanCustomizationRepository interface {
	ExecuteCreateCustomAssetTypeActionPlan(context.Context, tenant.ID, inventory.InventoryID, string, ActionPlanStateTransition, customfield.AssetType, audit.Record) (ActionPlanRecord, bool, error)
	ExecuteCreateCustomFieldActionPlan(context.Context, tenant.ID, inventory.InventoryID, string, ActionPlanStateTransition, customfield.Definition, audit.Record) (ActionPlanRecord, bool, error)
}
