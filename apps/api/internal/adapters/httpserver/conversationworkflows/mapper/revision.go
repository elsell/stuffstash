package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/conversationworkflows/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

func DefinitionToDomain(value dto.Definition) agentmodel.WorkflowDefinitionInput {
	return agentmodel.WorkflowDefinitionInput{Name: value.Name, ProviderProfileID: value.ProviderProfileID, Instructions: value.Instructions, Budget: agentmodel.WorkflowBudget{ToolCalls: value.Budget.ToolCalls, ModelCalls: value.Budget.ModelCalls, ElapsedSeconds: value.Budget.ElapsedSeconds, FollowUpTurns: value.Budget.FollowUpTurns}}
}
func RevisionToResponse(revision agentmodel.WorkflowRevision) dto.Revision {
	value := revision.Snapshot()
	definition := value.Definition.Settings()
	return dto.Revision{ID: string(value.ID), WorkflowID: string(value.WorkflowID), Number: value.Number, AuthorID: string(value.AuthorID), CreatedAt: value.CreatedAt, SettingsMigration: string(value.SettingsMigration), Definition: dto.Definition{Name: definition.Name, ProviderProfileID: definition.ProviderProfileID, Instructions: definition.Instructions, Budget: dto.Budget{ToolCalls: definition.Budget.ToolCalls, ModelCalls: definition.Budget.ModelCalls, ElapsedSeconds: definition.Budget.ElapsedSeconds, FollowUpTurns: definition.Budget.FollowUpTurns}}}
}
