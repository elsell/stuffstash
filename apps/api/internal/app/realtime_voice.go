package app

import (
	"context"

	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) WithRealtimeVoiceProviderResolver(resolver ports.RealtimeVoiceProviderResolver) App {
	a.realtimeVoiceProviders = resolver
	return a
}

func (a App) WithRealtimeVoiceProviders(stt ports.SpeechToTextProvider, model ports.ConversationModel, tts ports.TextToSpeechProvider) App {
	a.realtimeVoiceProviders = staticRealtimeVoiceProviderResolver{providers: ports.RealtimeVoiceProviderSet{SpeechToText: stt, ConversationModel: model, TextToSpeech: tts}}
	return a
}

func (a App) realtimeSessionService() agentmodelapp.RealtimeSessionService {
	return agentmodelapp.NewRealtimeSessionService(agentmodelapp.RealtimeSessionDependencies{
		Workflows: a.conversationWorkflowService, Providers: a.realtimeVoiceProviders,
		Access: a.inventoryService(), Authorizer: a.authorizer, Sessions: a.realtimeSessions,
		Clock: a.clock, IDs: a.ids, ContextBytes: a.conversationContextBytes,
	})
}

func (a App) StartRealtimeVoiceSession(ctx context.Context, input RealtimeVoiceSessionInput) (RealtimeVoiceSession, error) {
	if err := a.ensureRealtimeVoiceDependencies(); err != nil {
		return RealtimeVoiceSession{}, err
	}
	prepared, err := a.realtimeSessionService().Start(ctx, agentmodelapp.RealtimeSessionStartInput{
		Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID, Source: input.Source, InputAudio: input.InputAudio,
	})
	if err != nil {
		return RealtimeVoiceSession{}, err
	}
	providers := prepared.Providers
	return RealtimeVoiceSession{
		ID: prepared.ID, TenantID: input.TenantID, InventoryID: input.InventoryID, Principal: input.Principal,
		Source: input.Source, InputAudio: input.InputAudio, OutputAudio: input.OutputAudio,
		ConversationContinuity: input.ConversationContinuity, DeveloperDiagnostics: input.DeveloperDiagnostics,
		SpeechToTextProfileID: providers.SpeechToTextProfileID, LanguageInferenceProfileID: providers.LanguageInferenceProfileID,
		TextToSpeechProfileID: providers.TextToSpeechProfileID, LanguagePromptTemplate: providers.LanguagePromptTemplate,
		WorkflowRevisionID: prepared.WorkflowRevisionID, workflow: prepared.Workflow, conversationMemory: prepared.Memory,
		conversationModel: providers.ConversationModel, speechToText: providers.SpeechToText, textToSpeech: providers.TextToSpeech,
	}, nil
}

func (a App) RunRealtimeVoiceQuery(ctx context.Context, input RealtimeVoiceQueryInput, emit RealtimeVoiceEventSink) error {
	service := agentmodelapp.RealtimeQueryService{
		Sessions: a.realtimeSessionService(), Observer: a.observer,
		CleanupTimeout: a.realtimeVoiceToolCallTimeout, Configured: a.ensureRealtimeVoiceDependencies() == nil,
	}
	return service.Run(ctx, agentmodelapp.RealtimeQueryInput{
		Session: a.realtimeConversationSession(input.Session), Text: input.Text,
		InputAudio: input.Session.InputAudio, AudioChunks: input.AudioChunks,
		SpeechToText: input.Session.speechToText, TextToSpeech: input.Session.textToSpeech,
	}, emit, func(ctx context.Context, transcript string, silentReply bool) error {
		session := input.Session
		session.silentReply = silentReply
		return a.runRealtimeVoiceConversation(ctx, session, transcript, input.ConversationTurns, emit)
	})
}

func (a App) ensureRealtimeVoiceAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return a.realtimeSessionService().EnsureAccess(ctx, principal, tenantID, inventoryID)
}

func emitRealtimeVoiceDiagnostic(sessionID, title, detail string, emit RealtimeVoiceEventSink) error {
	return agentmodelapp.EmitRealtimeVoiceDiagnostic(sessionID, title, detail, emit)
}
func safeRealtimeVoiceDiagnosticText(value string, maxLength int) string {
	return agentmodelapp.SafeRealtimeVoiceDiagnosticText(value, maxLength)
}

func (a App) ensureRealtimeVoiceDependencies() error {
	if a.authorizer == nil || a.tenants == nil || a.inventories == nil || a.assets == nil || a.search == nil || a.realtimeVoiceProviders == nil || a.realtimeSessions == nil {
		return apperrors.ErrInvalidInput
	}
	return nil
}

func (a App) MarkRealtimeVoiceSessionFailed(ctx context.Context, session RealtimeVoiceSession, safeFailureCode string) error {
	if err := a.ensureRealtimeVoiceDependencies(); err != nil {
		return err
	}
	return a.markRealtimeVoiceSessionOutcome(ctx, session, ports.RealtimeSessionStateFailed, safeFailureCode)
}

func (a App) MarkRealtimeVoiceSessionCompleted(ctx context.Context, session RealtimeVoiceSession) error {
	if err := a.ensureRealtimeVoiceDependencies(); err != nil {
		return err
	}
	return a.markRealtimeVoiceSessionOutcome(ctx, session, ports.RealtimeSessionStateCompleted, "")
}

func (a App) MarkRealtimeVoiceSessionCancelled(ctx context.Context, session RealtimeVoiceSession) error {
	if err := a.ensureRealtimeVoiceDependencies(); err != nil {
		return err
	}
	return a.markRealtimeVoiceSessionOutcome(ctx, session, ports.RealtimeSessionStateCancelled, "")
}

func (a App) markRealtimeVoiceSessionOutcome(ctx context.Context, session RealtimeVoiceSession, state ports.RealtimeSessionState, safeFailureCode string) error {
	return a.realtimeSessionService().UpdateOutcome(ctx, session.TenantID, session.InventoryID, session.ID, state, safeFailureCode)
}

func RealtimeVoiceSafeErrorCode(err error) string {
	return realtimeVoiceErrorCode(err)
}
