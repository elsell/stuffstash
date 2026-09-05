package app

import (
	"context"
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
)

type SaveConversationWorkflowInput = agentmodelapp.SaveConversationWorkflowInput

func (a App) SaveConversationWorkflowRevision(ctx context.Context, input SaveConversationWorkflowInput) (agentmodel.WorkflowRevision, error) {
	return a.conversationWorkflowService.SaveRevision(ctx, input)
}
