package app

import (
	"context"
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeReviewScope(session RealtimeVoiceSession) ActionPlanDecisionInput {
	return ActionPlanDecisionInput{Principal: session.Principal, TenantID: session.TenantID, InventoryID: session.InventoryID}
}
func (a App) realtimeVoiceActionPlanProposal(ctx context.Context, session RealtimeVoiceSession, record ports.ActionPlanRecord) (RealtimeVoiceActionPlanProposal, error) {
	return a.actionPlanService().ReviewProposal(ctx, realtimeReviewScope(session), record)
}
func (a App) realtimeVoiceActionPlanCommand(ctx context.Context, session RealtimeVoiceSession, command ports.ActionPlanCommandRecord) (RealtimeVoiceActionPlanCommand, error) {
	return a.actionPlanService().ReviewCommand(ctx, realtimeReviewScope(session), command)
}
func actionPlanCommandOperation(kind actionplan.CommandKind) string {
	return agentmodelapp.ActionPlanCommandOperation(kind)
}
