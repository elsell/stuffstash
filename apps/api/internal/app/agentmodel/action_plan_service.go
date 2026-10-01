package agentmodel

import (
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ActionPlanDependencies struct {
	AssetPreparation         ports.ActionPlanAssetPreparation
	CustomizationPreparation ports.ActionPlanCustomizationPreparation
	InventoryAccess          ports.ActiveInventoryAccess
	ActionPlanCustomizations ports.ActionPlanCustomizationRepository
	ActionPlans              ports.ActionPlanRepository
	Assets                   ports.AssetRepository
	Authorizer               ports.Authorizer
	Clock                    ports.Clock
	CustomAssetTypes         ports.CustomAssetTypeRepository
	CustomFields             ports.CustomFieldDefinitionRepository
	IDs                      ports.IDGenerator
	Inventories              ports.InventoryRepository
	MaxAttachmentBytes       int
	Tenants                  ports.TenantRepository
}
type ActionPlanService struct{ deps ActionPlanDependencies }

func NewActionPlanService(deps ActionPlanDependencies) ActionPlanService {
	return ActionPlanService{deps: deps}
}
