package agentmodel

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const maxActionPlanEditedTitleLength = 200

func (a ActionPlanService) approveEditedActionPlan(ctx context.Context, input ActionPlanDecisionInput) (ports.ActionPlanRecord, error) {
	if err := a.ensureActionPlanDependencies(); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	if err := a.deps.InventoryAccess.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	record, found, err := a.deps.ActionPlans.ActionPlanByID(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID))
	if err != nil {
		return ports.ActionPlanRecord{}, err
	}
	if !found {
		return ports.ActionPlanRecord{}, apperrors.ErrNotFound
	}
	if record.PrincipalID != input.Principal.ID || record.State != actionplan.StateProposed {
		return ports.ActionPlanRecord{}, apperrors.ErrConflict
	}
	commands, err := a.applyActionPlanCommandEdits(ctx, input, record.Commands)
	if err != nil {
		return ports.ActionPlanRecord{}, err
	}
	updated, found, err := a.deps.ActionPlans.UpdateActionPlanCommandsAndState(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), commands, ports.ActionPlanStateTransition{
		PrincipalID: input.Principal.ID,
		From:        actionplan.StateProposed,
		To:          actionplan.StateApproved,
		At:          a.deps.Clock.Now(),
	})
	if err != nil {
		if err == ports.ErrConflict {
			return ports.ActionPlanRecord{}, apperrors.ErrConflict
		}
		return ports.ActionPlanRecord{}, err
	}
	if !found {
		return ports.ActionPlanRecord{}, apperrors.ErrNotFound
	}
	return updated, nil
}

func (a ActionPlanService) applyActionPlanCommandEdits(ctx context.Context, input ActionPlanDecisionInput, commands []ports.ActionPlanCommandRecord) ([]ports.ActionPlanCommandRecord, error) {
	if len(input.CommandEdits) > len(commands) || len(input.CommandEdits) > maxActionPlanCommands {
		return nil, apperrors.ErrValidation
	}
	indexes := make(map[string]int, len(commands))
	for index, command := range commands {
		indexes[command.ID] = index
	}
	result := append([]ports.ActionPlanCommandRecord(nil), commands...)
	seen := map[string]struct{}{}
	for _, edit := range input.CommandEdits {
		commandID := strings.TrimSpace(edit.CommandID)
		index, ok := indexes[commandID]
		if !ok {
			return nil, apperrors.ErrValidation
		}
		if _, duplicate := seen[commandID]; duplicate {
			return nil, apperrors.ErrValidation
		}
		seen[commandID] = struct{}{}
		command := result[index]
		if command.Kind != actionplan.CommandKindCreateAsset && command.Kind != actionplan.CommandKindCreateLocation {
			return nil, apperrors.ErrValidation
		}
		var arguments map[string]any
		if err := json.Unmarshal(command.ArgumentsJSON, &arguments); err != nil {
			return nil, apperrors.ErrValidation
		}
		if edit.Title != nil {
			title := strings.TrimSpace(*edit.Title)
			if title == "" || len([]rune(title)) > maxActionPlanEditedTitleLength {
				return nil, apperrors.ErrValidation
			}
			arguments["title"] = title
			delete(arguments, "name")
		}
		if edit.ParentSelection != nil {
			delete(arguments, "parentAssetId")
			delete(arguments, "parentCommandId")
			switch edit.ParentSelection.Kind {
			case "root":
				if strings.TrimSpace(edit.ParentSelection.ID) != "" {
					return nil, apperrors.ErrValidation
				}
			case "asset":
				parentID := strings.TrimSpace(edit.ParentSelection.ID)
				if parentID == "" || a.deps.Assets == nil {
					return nil, apperrors.ErrValidation
				}
				parent, found, err := a.deps.Assets.AssetByID(ctx, input.TenantID, input.InventoryID, asset.ID(parentID))
				if err != nil {
					return nil, err
				}
				if !found || parent.LifecycleState != asset.LifecycleStateActive {
					return nil, apperrors.ErrValidation
				}
				arguments["parentAssetId"] = parentID
			case "command":
				parentID := strings.TrimSpace(edit.ParentSelection.ID)
				parentIndex, exists := indexes[parentID]
				if !exists || parentIndex >= index {
					return nil, apperrors.ErrValidation
				}
				parent := commands[parentIndex]
				if parent.Kind != actionplan.CommandKindCreateAsset && parent.Kind != actionplan.CommandKindCreateLocation {
					return nil, apperrors.ErrValidation
				}
				arguments["parentCommandId"] = parentID
			default:
				return nil, apperrors.ErrValidation
			}
		}
		encoded, err := json.Marshal(arguments)
		if err != nil || len(encoded) > maxActionPlanArgumentBytes {
			return nil, apperrors.ErrValidation
		}
		command.ArgumentsJSON = encoded
		result[index] = command
	}
	return result, nil
}
