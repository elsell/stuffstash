package app

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const maxActionPlanEditedTitleLength = 200

func (a App) approveEditedActionPlan(ctx context.Context, input ActionPlanDecisionInput) (ports.ActionPlanRecord, error) {
	if err := a.ensureActionPlanDependencies(); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	record, found, err := a.actionPlans.ActionPlanByID(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID))
	if err != nil {
		return ports.ActionPlanRecord{}, err
	}
	if !found {
		return ports.ActionPlanRecord{}, ErrNotFound
	}
	if record.PrincipalID != input.Principal.ID || record.State != actionplan.StateProposed {
		return ports.ActionPlanRecord{}, ErrConflict
	}
	commands, err := a.applyActionPlanCommandEdits(ctx, input, record.Commands)
	if err != nil {
		return ports.ActionPlanRecord{}, err
	}
	updated, found, err := a.actionPlans.UpdateActionPlanCommandsAndState(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), commands, ports.ActionPlanStateTransition{
		PrincipalID: input.Principal.ID,
		From:        actionplan.StateProposed,
		To:          actionplan.StateApproved,
		At:          a.clock.Now(),
	})
	if err != nil {
		if err == ports.ErrConflict {
			return ports.ActionPlanRecord{}, ErrConflict
		}
		return ports.ActionPlanRecord{}, err
	}
	if !found {
		return ports.ActionPlanRecord{}, ErrNotFound
	}
	return updated, nil
}

func (a App) applyActionPlanCommandEdits(ctx context.Context, input ActionPlanDecisionInput, commands []ports.ActionPlanCommandRecord) ([]ports.ActionPlanCommandRecord, error) {
	if len(input.CommandEdits) > len(commands) || len(input.CommandEdits) > maxActionPlanCommands {
		return nil, ErrValidation
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
			return nil, ErrValidation
		}
		if _, duplicate := seen[commandID]; duplicate {
			return nil, ErrValidation
		}
		seen[commandID] = struct{}{}
		command := result[index]
		if command.Kind != actionplan.CommandKindCreateAsset && command.Kind != actionplan.CommandKindCreateLocation {
			return nil, ErrValidation
		}
		var arguments map[string]any
		if err := json.Unmarshal(command.ArgumentsJSON, &arguments); err != nil {
			return nil, ErrValidation
		}
		if edit.Title != nil {
			title := strings.TrimSpace(*edit.Title)
			if title == "" || len([]rune(title)) > maxActionPlanEditedTitleLength {
				return nil, ErrValidation
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
					return nil, ErrValidation
				}
			case "asset":
				parentID := strings.TrimSpace(edit.ParentSelection.ID)
				if parentID == "" || a.assets == nil {
					return nil, ErrValidation
				}
				parent, found, err := a.assets.AssetByID(ctx, input.TenantID, input.InventoryID, asset.ID(parentID))
				if err != nil {
					return nil, err
				}
				if !found || parent.LifecycleState != asset.LifecycleStateActive {
					return nil, ErrValidation
				}
				arguments["parentAssetId"] = parentID
			case "command":
				parentID := strings.TrimSpace(edit.ParentSelection.ID)
				parentIndex, exists := indexes[parentID]
				if !exists || parentIndex >= index {
					return nil, ErrValidation
				}
				parent := commands[parentIndex]
				if parent.Kind != actionplan.CommandKindCreateAsset && parent.Kind != actionplan.CommandKindCreateLocation {
					return nil, ErrValidation
				}
				arguments["parentCommandId"] = parentID
			default:
				return nil, ErrValidation
			}
		}
		encoded, err := json.Marshal(arguments)
		if err != nil || len(encoded) > maxActionPlanArgumentBytes {
			return nil, ErrValidation
		}
		command.ArgumentsJSON = encoded
		result[index] = command
	}
	return result, nil
}
