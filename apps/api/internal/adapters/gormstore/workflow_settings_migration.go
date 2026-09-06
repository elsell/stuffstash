package gormstore

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

// Historical JSON is a persistence concern. None of its stages survive into
// the executable domain definition or select an alternate runtime.
func (snapshot *workflowDefinitionSnapshot) UnmarshalJSON(raw []byte) error {
	type current workflowDefinitionSnapshot
	var value current
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	var historical struct {
		Definition struct {
			Steps *[]struct {
				Kind              string
				ProviderProfileID string
				Instructions      string
			}
		}
	}
	if err := json.Unmarshal(raw, &historical); err != nil {
		return err
	}
	if historical.Definition.Steps != nil {
		steps := *historical.Definition.Steps
		if len(steps) != 3 || steps[0].Kind != "interpret" || steps[1].Kind != "assess" || steps[2].Kind != "respond" {
			return agentmodel.ErrInvalidWorkflowDefinition
		}
		// Old evidence rounds have different units and cannot become tool calls.
		value.Definition.ProviderProfileID = steps[0].ProviderProfileID
		value.Definition.Instructions = steps[0].Instructions
		value.Definition.Budget.ToolCalls = min(6, value.Definition.Budget.ModelCalls)
		value.Limits.Budget.ToolCalls = min(6, value.Limits.Budget.ModelCalls)
		value.SettingsMigration = agentmodel.WorkflowSettingsMigrationLegacy
	}
	*snapshot = workflowDefinitionSnapshot(value)
	return nil
}
