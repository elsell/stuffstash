package audithistory

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/audit"
)

func (a Service) ProjectAssetActivityEntry(ctx context.Context, input ListAssetActivityInput, record audit.Record, canUndo bool) audit.AssetActivityEntry {
	entry := audit.AssetActivityEntry{
		ID: record.ID, PrincipalID: record.PrincipalID, Action: record.Action, Category: record.Action.AssetActivityCategory(), Source: record.Source,
		OccurredAt: record.OccurredAt, RequestID: record.RequestID, Changes: projectAssetActivityChanges(record), TechnicalMetadata: projectAssetActivityTechnicalMetadata(record),
	}
	operationID := strings.TrimSpace(record.Metadata["operation_id"])
	if canUndo && operationID != "" && a.deps.Undoables != nil {
		operation, found, err := a.deps.Undoables.UndoableOperationByID(ctx, input.TenantID, input.InventoryID, operationID)
		if err == nil && found && operation.TargetType == audit.TargetAsset && operation.TargetID == input.AssetID.String() && operation.OriginalAction == record.Action {
			entry.Undo = &audit.AssetActivityUndo{OperationID: operation.ID, Status: string(operation.Status)}
		}
	}
	return entry
}

func projectAssetActivityTechnicalMetadata(record audit.Record) map[string]string {
	technical := map[string]string{}
	if action, ok := audit.NewAction(record.Metadata["original_action"]); ok {
		technical["original_action"] = action.String()
	}
	if targetType, ok := audit.NewTargetType(record.Metadata["target_type"]); ok {
		technical["target_type"] = targetType.String()
	}
	return technical
}

func projectAssetActivityChanges(record audit.Record) []audit.AssetActivityChange {
	metadata := record.Metadata
	changes := make([]audit.AssetActivityChange, 0, 4)
	appendValues := func(field audit.AssetActivityField, previousKey, currentKey string) {
		previous, previousOK := metadata[previousKey]
		current, currentOK := metadata[currentKey]
		if previousOK || currentOK {
			changes = append(changes, audit.AssetActivityChange{Field: field, PreviousValue: previous, CurrentValue: current})
		}
	}
	appendValues(audit.AssetActivityFieldTitle, "previous_title", "updated_title")
	if metadata["description_changed"] == "true" {
		changes = append(changes, audit.AssetActivityChange{Field: audit.AssetActivityFieldDescription})
	}
	appendValues(audit.AssetActivityFieldTags, "previous_tag_count", "updated_tag_count")
	appendValues(audit.AssetActivityFieldParent, "previous_parent", "new_parent")
	appendValues(audit.AssetActivityFieldLifecycleState, "previous_lifecycle_state", "new_lifecycle_state")
	appendValues(audit.AssetActivityFieldCheckoutState, "previous_checkout_state", "new_checkout_state")
	return changes
}
