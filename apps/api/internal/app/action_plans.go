package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type preparedActionPlanCustomization = agentmodel.PreparedActionPlanCustomization
type ActionPlanPhotoAttachmentMetadataInput = agentmodel.ActionPlanPhotoAttachmentMetadataInput
type ActionPlanPhotoAttachmentMetadata = agentmodel.ActionPlanPhotoAttachmentMetadata
type CreateActionPlanInput = agentmodel.CreateActionPlanInput
type ActionPlanCommandInput = agentmodel.ActionPlanCommandInput
type ActionPlanDecisionInput = agentmodel.ActionPlanDecisionInput
type ActionPlanCommandEditInput = agentmodel.ActionPlanCommandEditInput
type ActionPlanParentSelectionInput = agentmodel.ActionPlanParentSelectionInput
type ActionPlanExecutionResult = agentmodel.ActionPlanExecutionResult
type ActionPlanCommandExecutionResult = agentmodel.ActionPlanCommandExecutionResult
type actionPlanMoveArguments = agentmodel.ActionPlanMoveArguments
type actionPlanCheckoutArguments = agentmodel.ActionPlanCheckoutArguments
type actionPlanCreateArguments = agentmodel.ActionPlanCreateArguments
type actionPlanUpdateArguments = agentmodel.ActionPlanUpdateArguments

const maxActionPlanSummaryLength = agentmodel.MaxActionPlanSummaryLength

func (a App) actionPlanService() agentmodel.ActionPlanService {
	return agentmodel.NewActionPlanService(agentmodel.ActionPlanDependencies{
		AssetPreparation:         a.assetService,
		CustomizationPreparation: a.customFieldService,
		InventoryAccess:          actionPlanInventoryAccess{a},
		ActionPlanCustomizations: a.actionPlanCustomizations,
		ActionPlans:              a.actionPlans,
		Assets:                   a.assets,
		Authorizer:               a.authorizer,
		Clock:                    a.clock,
		CustomAssetTypes:         a.customAssetTypes,
		CustomFields:             a.customFields,
		IDs:                      a.ids,
		Inventories:              a.inventories,
		MaxAttachmentBytes:       a.maxAttachmentBytes,
		Tenants:                  a.tenants,
	})
}

type actionPlanInventoryAccess struct{ app App }

func (a actionPlanInventoryAccess) EnsureActiveInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	return a.app.ensureActiveInventoryAccess(ctx, principal, tenantID, inventoryID, permission)
}
func isCustomizationCommand(kind actionplan.CommandKind) bool {
	return agentmodel.IsCustomizationCommand(kind)
}
func (a App) prepareActionPlanCustomization(ctx context.Context, input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (preparedActionPlanCustomization, error) {
	return a.actionPlanService().PrepareActionPlanCustomization(ctx, input, command)
}
func (a App) ValidateActionPlanPhotoAttachmentMetadata(ctx context.Context, input ActionPlanPhotoAttachmentMetadataInput) error {
	return a.actionPlanService().ValidateActionPlanPhotoAttachmentMetadata(ctx, input)
}
func (a App) CreateActionPlan(ctx context.Context, input CreateActionPlanInput) (ports.ActionPlanRecord, error) {
	return a.actionPlanService().CreateActionPlan(ctx, input)
}
func (a App) ApproveActionPlan(ctx context.Context, input ActionPlanDecisionInput) (ports.ActionPlanRecord, error) {
	return a.actionPlanService().ApproveActionPlan(ctx, input)
}
func (a App) CancelActionPlan(ctx context.Context, input ActionPlanDecisionInput) (ports.ActionPlanRecord, error) {
	return a.actionPlanService().CancelActionPlan(ctx, input)
}
func (a App) ExecuteActionPlan(ctx context.Context, input ActionPlanDecisionInput) (ports.ActionPlanRecord, error) {
	return a.actionPlanService().ExecuteActionPlan(ctx, input)
}
func (a App) ExecuteActionPlanDetailed(ctx context.Context, input ActionPlanDecisionInput) (ActionPlanExecutionResult, error) {
	return a.actionPlanService().ExecuteActionPlanDetailed(ctx, input)
}
func parseActionPlanMoveArguments(command ports.ActionPlanCommandRecord) (actionPlanMoveArguments, error) {
	return agentmodel.ParseActionPlanMoveArguments(command)
}
func parseActionPlanAssetIDOnlyArguments(command ports.ActionPlanCommandRecord) (asset.ID, error) {
	return agentmodel.ParseActionPlanAssetIDOnlyArguments(command)
}
func parseActionPlanCheckoutArguments(command ports.ActionPlanCommandRecord) (actionPlanCheckoutArguments, error) {
	return agentmodel.ParseActionPlanCheckoutArguments(command)
}
func (a App) actionPlanCommands(inputs []ActionPlanCommandInput) ([]ports.ActionPlanCommandRecord, error) {
	return a.actionPlanService().ActionPlanCommands(inputs)
}
func parseActionPlanCreateArguments(command ports.ActionPlanCommandRecord) (actionPlanCreateArguments, error) {
	return agentmodel.ParseActionPlanCreateArguments(command)
}
func parseActionPlanUpdateArguments(command ports.ActionPlanCommandRecord) (actionPlanUpdateArguments, error) {
	return agentmodel.ParseActionPlanUpdateArguments(command)
}
func validActionPlanCommandID(value string) bool { return agentmodel.ValidActionPlanCommandID(value) }
func validateExecutableActionPlanArguments(kind actionplan.CommandKind, arguments []byte) error {
	return agentmodel.ValidateExecutableActionPlanArguments(kind, arguments)
}
