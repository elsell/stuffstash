package app

import (
	"context"
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type SaveConversationWorkflowInput = agentmodelapp.SaveConversationWorkflowInput

func (a App) SaveConversationWorkflowRevision(ctx context.Context, input SaveConversationWorkflowInput) (agentmodel.WorkflowRevision, error) {
	return a.conversationWorkflowService.SaveRevision(ctx, input)
}

type ActivateWorkflowInput = agentmodelapp.ActivateWorkflowInput

func (a App) ActivateConversationWorkflow(ctx context.Context, input ActivateWorkflowInput) (agentmodel.WorkflowRevision, error) {
	return a.workflowActivation.Activate(ctx, input)
}

type GetWorkflowInput = agentmodelapp.GetWorkflowInput
type ListWorkflowsInput = agentmodelapp.ListWorkflowsInput
type ListWorkflowsResult = agentmodelapp.ListWorkflowsResult
type ListWorkflowRevisionsInput = agentmodelapp.ListWorkflowRevisionsInput
type ListWorkflowRevisionsResult = agentmodelapp.ListWorkflowRevisionsResult

func (a App) GetConversationWorkflow(ctx context.Context, input GetWorkflowInput) (agentmodel.WorkflowRevision, error) {
	return a.workflowQueries.Get(ctx, input)
}
func (a App) ListConversationWorkflows(ctx context.Context, input ListWorkflowsInput) (ListWorkflowsResult, error) {
	return a.workflowQueries.List(ctx, input)
}
func (a App) ListConversationWorkflowRevisions(ctx context.Context, input ListWorkflowRevisionsInput) (ListWorkflowRevisionsResult, error) {
	return a.workflowQueries.History(ctx, input)
}
func (a App) GetConversationWorkflowSelection(ctx context.Context, input EvaluationRunAccess) (*ports.WorkflowSelectionReference, error) {
	return a.workflowQueries.Selection(ctx, input)
}
