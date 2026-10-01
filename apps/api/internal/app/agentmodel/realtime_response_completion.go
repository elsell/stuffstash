package agentmodel

import (
	"context"
	"fmt"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type RealtimeResponseSession struct {
	ID                     string
	TenantID               tenant.ID
	InventoryID            inventory.InventoryID
	Principal              identity.Principal
	Source                 string
	OutputMimeTypes        []string
	SilentReply            bool
	ConversationContinuity bool
	DeveloperDiagnostics   bool
	TextToSpeech           ports.TextToSpeechProvider
}

type RealtimeOutcomeWriter interface {
	UpdateOutcome(context.Context, tenant.ID, inventory.InventoryID, string, ports.RealtimeSessionState, string) error
}

type RealtimeResponseService struct {
	ids      ports.IDGenerator
	outcomes RealtimeOutcomeWriter
}

func NewRealtimeResponseService(ids ports.IDGenerator, outcomes RealtimeOutcomeWriter) RealtimeResponseService {
	return RealtimeResponseService{ids: ids, outcomes: outcomes}
}

func (a RealtimeResponseService) newID() string {
	if a.ids == nil {
		return ""
	}
	return a.ids.NewID()
}

func (a RealtimeResponseService) Complete(ctx context.Context, session RealtimeResponseSession, response ports.StructuredAgentResponse, toolCallIDs []string, toolResults []ports.AgentToolResult, emit RealtimeVoiceEventSink, continueAfterClarification ...bool) error {
	if err := EmitRealtimeVoiceProgress(session.ID, RealtimeVoiceProgressAnswering, "Preparing a response.", emit); err != nil {
		return err
	}
	if response.Kind == "" {
		response.Kind = ports.StructuredAgentResponseKindAnswer
	}
	response.ResponseID = a.newID()
	response.SessionID = session.ID
	response.TenantID = session.TenantID
	response.InventoryID = session.InventoryID
	response.Source = session.Source
	response.ToolCallIDs = append([]string{}, toolCallIDs...)
	if response.DisplayResponse == "" {
		response.DisplayResponse = response.SpokenResponse
	}
	if err := ValidateRealtimeVoiceFinalResponse(response); err != nil {
		return err
	}

	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventAssistantResponseStarted, SessionID: session.ID, Response: &response}); err != nil {
		return err
	}
	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventAssistantResponseCompleted, SessionID: session.ID, Response: &response}); err != nil {
		return err
	}

	if !session.SilentReply {
		if err := emitRealtimeVoiceSpeech(ctx, session, response, toolResults, emit); err != nil {
			return err
		}
	}
	if !RealtimeVoiceShouldContinueAfterClarification(response, continueAfterClarification...) && !(session.ConversationContinuity && response.Kind == ports.StructuredAgentResponseKindAnswer) {
		if err := a.outcomes.UpdateOutcome(ctx, session.TenantID, session.InventoryID, session.ID, ports.RealtimeSessionStateCompleted, ""); err != nil {
			return err
		}
	}
	return emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventSessionCompleted, SessionID: session.ID})
}

func emitRealtimeVoiceSpeech(ctx context.Context, session RealtimeResponseSession, response ports.StructuredAgentResponse, toolResults []ports.AgentToolResult, emit RealtimeVoiceEventSink) error {
	speech, err := session.TextToSpeech.Synthesize(ctx, ports.TextToSpeechInput{
		TenantID:    session.TenantID,
		InventoryID: session.InventoryID,
		Principal:   session.Principal,
		Text:        response.SpokenResponse,
		MimeTypes:   session.OutputMimeTypes,
	})
	if err != nil {
		if diagnosticErr := emitRealtimeVoiceTextToSpeechFailureDiagnostic(session, toolResults, RealtimeVoiceFailureTextToSpeech, err, emit); diagnosticErr != nil {
			return diagnosticErr
		}
		return RealtimeVoiceProviderStageError{Code: RealtimeVoiceFailureTextToSpeech, Cause: err}
	}
	speechChunks := RealtimeVoicePlayableSpeechChunks(speech.Chunks)
	if speech.MimeType == "" || len(speechChunks) == 0 {
		err := ports.ErrInvalidProviderInput
		if diagnosticErr := emitRealtimeVoiceTextToSpeechFailureDiagnostic(session, toolResults, RealtimeVoiceFailureTextToSpeech, err, emit); diagnosticErr != nil {
			return diagnosticErr
		}
		return RealtimeVoiceProviderStageError{Code: RealtimeVoiceFailureTextToSpeech, Cause: err}
	}
	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventTextToSpeechAudioStarted, SessionID: session.ID, AudioMime: speech.MimeType}); err != nil {
		return err
	}
	for index, chunk := range speechChunks {
		if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventTextToSpeechAudioChunk, SessionID: session.ID, ChunkID: fmt.Sprintf("tts-%d", index+1), Audio: chunk, FinalChunk: index == len(speechChunks)-1}); err != nil {
			return err
		}
	}
	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventTextToSpeechAudioCompleted, SessionID: session.ID}); err != nil {
		return err
	}
	return nil
}

func RealtimeVoicePlayableSpeechChunks(chunks [][]byte) [][]byte {
	playable := make([][]byte, 0, len(chunks))
	for _, chunk := range chunks {
		if len(chunk) > 0 {
			playable = append(playable, chunk)
		}
	}
	return playable
}

func (a RealtimeResponseService) Recover(ctx context.Context, session RealtimeResponseSession, toolCallIDs []string, toolResults []ports.AgentToolResult, emit RealtimeVoiceEventSink) error {
	if err := EmitRealtimeVoiceProgress(session.ID, RealtimeVoiceProgressRecovering, "Recovering safely.", emit); err != nil {
		return err
	}
	return a.Complete(ctx, session, ports.StructuredAgentResponse{
		Kind:            ports.StructuredAgentResponseKindSafeFailure,
		SpokenResponse:  "I could not finish that voice request safely. Please try again with a little more detail.",
		DisplayResponse: "I could not finish that voice request safely. Please try again with a little more detail.",
	}, toolCallIDs, toolResults, emit)
}

func RealtimeVoiceShouldContinueAfterClarification(response ports.StructuredAgentResponse, continueAfterClarification ...bool) bool {
	return len(continueAfterClarification) > 0 && continueAfterClarification[0] && response.Kind == ports.StructuredAgentResponseKindClarification
}
