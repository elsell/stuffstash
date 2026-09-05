package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/conversationworkflows/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

func DefinitionToDomain(value dto.Definition) agentmodel.WorkflowDefinitionInput {
	steps := make([]agentmodel.WorkflowStep, len(value.Steps))
	for i, step := range value.Steps {
		steps[i] = agentmodel.WorkflowStep{Kind: agentmodel.WorkflowStepKind(step.Kind), ProviderProfileID: step.ProviderProfileID, Instructions: step.Instructions, Attempts: step.Attempts}
	}
	return agentmodel.WorkflowDefinitionInput{Name: value.Name, Retrieval: agentmodel.WorkflowRetrievalStrategy(value.Retrieval), Response: agentmodel.WorkflowResponseMode(value.Response), Budget: agentmodel.WorkflowBudget{EvidenceRounds: value.Budget.EvidenceRounds, ModelCalls: value.Budget.ModelCalls, ElapsedSeconds: value.Budget.ElapsedSeconds, FollowUpTurns: value.Budget.FollowUpTurns}, Steps: steps}
}
func RevisionToResponse(revision agentmodel.WorkflowRevision) dto.Revision {
	value := revision.Snapshot()
	definition := value.Definition.Settings()
	steps := make([]dto.Step, len(definition.Steps))
	for i, step := range definition.Steps {
		steps[i] = dto.Step{Kind: string(step.Kind), ProviderProfileID: step.ProviderProfileID, Instructions: step.Instructions, Attempts: step.Attempts}
	}
	return dto.Revision{ID: string(value.ID), WorkflowID: string(value.WorkflowID), Number: value.Number, AuthorID: string(value.AuthorID), CreatedAt: value.CreatedAt, Definition: dto.Definition{Name: definition.Name, Retrieval: string(definition.Retrieval), Response: string(definition.Response), Budget: dto.Budget{EvidenceRounds: definition.Budget.EvidenceRounds, ModelCalls: definition.Budget.ModelCalls, ElapsedSeconds: definition.Budget.ElapsedSeconds, FollowUpTurns: definition.Budget.FollowUpTurns}, Steps: steps}}
}
