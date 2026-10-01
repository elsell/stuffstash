package app

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

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

func (a App) RunRealtimeVoiceQuery(ctx context.Context, input RealtimeVoiceQueryInput, emit RealtimeVoiceEventSink) (err error) {
	duration := time.Minute
	if input.Session.workflow != nil {
		duration = time.Duration(input.Session.workflow.Revision().Snapshot().Definition.Settings().Budget.ElapsedSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	if input.Session.conversationModel != nil && !input.Session.conversationMemory.Matches(realtimeConversationScope(input.Session)) {
		return ports.ErrForbidden
	}
	defer func() {
		if err != nil && strings.TrimSpace(input.Session.ID) != "" {
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.realtimeVoiceToolCallTimeout)
			defer cancel()
			if errors.Is(err, context.Canceled) {
				_ = a.markRealtimeVoiceSessionOutcome(cleanupCtx, input.Session, ports.RealtimeSessionStateCancelled, "")
				return
			}
			safeCode := realtimeVoiceErrorCode(err)
			if a.observer != nil {
				a.observer.Record(ctx, ports.Event{
					Name:    ports.EventRealtimeVoiceFailed,
					Message: "realtime voice failed safely",
					Fields: map[string]string{
						"tenant_id":         input.Session.TenantID.String(),
						"inventory_id":      input.Session.InventoryID.String(),
						"principal_id":      input.Session.Principal.ID.String(),
						"session_id":        input.Session.ID,
						"safe_failure_code": safeCode,
						"error":             safeRealtimeVoiceErrorDetail(err),
					},
				})
			}
			_ = a.markRealtimeVoiceSessionOutcome(cleanupCtx, input.Session, ports.RealtimeSessionStateFailed, safeCode)
		}
	}()
	if err := a.ensureRealtimeVoiceDependencies(); err != nil {
		return err
	}
	if (len(input.AudioChunks) == 0) == (strings.TrimSpace(input.Text) == "") || utf8.RuneCountInString(input.Text) > MaxRealtimeTextCharacters {
		return ports.ErrInvalidProviderInput
	}

	if input.Session.speechToText == nil || input.Session.textToSpeech == nil || input.Session.conversationModel == nil {
		return apperrors.ErrInvalidInput
	}
	if err := a.ensureRealtimeVoiceAccess(ctx, input.Session.Principal, input.Session.TenantID, input.Session.InventoryID); err != nil {
		return err
	}
	transcript := strings.TrimSpace(input.Text)
	if transcript == "" {
		transcription, err := input.Session.speechToText.Transcribe(ctx, ports.SpeechToTextInput{
			TenantID:    input.Session.TenantID,
			InventoryID: input.Session.InventoryID,
			Principal:   input.Session.Principal,
			AudioFormat: input.Session.InputAudio,
			AudioChunks: input.AudioChunks,
		})
		if err != nil {
			return realtimeVoiceProviderStageError{Code: realtimeVoiceFailureSpeechToText, Cause: err}
		}
		transcript = strings.TrimSpace(transcription.Transcript)
	}
	if transcript == "" {
		return ports.ErrInvalidProviderInput
	}
	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventTranscriptFinal, SessionID: input.Session.ID, Text: transcript}); err != nil {
		return err
	}
	input.Session.silentReply = strings.TrimSpace(input.Text) != ""
	return a.runRealtimeVoiceConversation(ctx, input.Session, transcript, input.ConversationTurns, emit)
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
