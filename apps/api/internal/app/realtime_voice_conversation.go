package app

import (
	"context"
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) runRealtimeVoiceConversation(ctx context.Context, session RealtimeVoiceSession, transcript string, prior []ports.AgentConversationTurn, emit RealtimeVoiceEventSink) error {
	return a.realtimeConversationService().Run(ctx, a.realtimeConversationSession(session), transcript, prior, emit,
		func(ctx context.Context, response ports.StructuredAgentResponse, callIDs []string, results []ports.AgentToolResult) error {
			return a.completeRealtimeVoiceResponse(ctx, session, response, callIDs, results, emit)
		},
		func(modelCalls, toolCalls int, results []ports.AgentToolResult, err error) error {
			return emitRealtimeVoiceConversationFailureDiagnostic(session, modelCalls, toolCalls, results, err, emit)
		})
}

func (a App) realtimeConversationService() agentmodelapp.RealtimeConversationService {
	return agentmodelapp.NewRealtimeConversationService(agentmodelapp.RealtimeConversationDependencies{
		Queries: realtimeConversationQueries{realtimeToolQueries{a}}, Reads: a.realtimeReadTools(), Plans: a.actionPlanService(),
		CustomAssetTypes: a.customAssetTypes, CustomFields: a.customFields,
	})
}

func (a App) realtimeConversationSession(session RealtimeVoiceSession) agentmodelapp.RealtimeConversationSession {
	return agentmodelapp.RealtimeConversationSession{
		RealtimeReadToolScope: realtimeReadScope(session), ID: session.ID, Source: session.Source,
		LanguagePromptTemplate: session.LanguagePromptTemplate, Memory: session.conversationMemory,
		Model: session.conversationModel, Workflow: session.workflow, ContextBytes: a.conversationContextBytes,
	}
}

type realtimeConversationQueries struct{ realtimeToolQueries }

func (q realtimeConversationQueries) EnsureActiveInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, permission ports.InventoryPermission) error {
	return q.App.ensureActiveInventoryAccess(ctx, principal, tenantID, inventoryID, permission)
}

var _ agentmodelapp.RealtimeConversationQueries = realtimeConversationQueries{}
