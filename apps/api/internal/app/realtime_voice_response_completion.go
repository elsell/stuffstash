package app

import (
	"context"
	"fmt"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) completeRealtimeVoiceResponse(ctx context.Context, session RealtimeVoiceSession, response ports.StructuredAgentResponse, toolCallIDs []string, toolResults []ports.AgentToolResult, emit RealtimeVoiceEventSink, continueAfterClarification ...bool) error {
	if err := emitRealtimeVoiceProgress(session, realtimeVoiceProgressAnswering, "Preparing a response.", emit); err != nil {
		return err
	}
	if response.Kind == "" {
		response.Kind = ports.StructuredAgentResponseKindAnswer
	}
	response.ResponseID = a.newRealtimeVoiceID()
	response.SessionID = session.ID
	response.TenantID = session.TenantID
	response.InventoryID = session.InventoryID
	response.Source = session.Source
	response.ToolCallIDs = append([]string{}, toolCallIDs...)
	if response.DisplayResponse == "" {
		response.DisplayResponse = response.SpokenResponse
	}

	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventAssistantResponseStarted, SessionID: session.ID, Response: &response}); err != nil {
		return err
	}
	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventAssistantResponseCompleted, SessionID: session.ID, Response: &response}); err != nil {
		return err
	}

	speech, err := session.textToSpeech.Synthesize(ctx, ports.TextToSpeechInput{
		TenantID:    session.TenantID,
		InventoryID: session.InventoryID,
		Principal:   session.Principal,
		Text:        response.SpokenResponse,
		MimeTypes:   session.OutputAudio.MimeTypes,
	})
	if err != nil {
		if diagnosticErr := emitRealtimeVoiceTextToSpeechFailureDiagnostic(session, toolResults, realtimeVoiceFailureTextToSpeech, err, emit); diagnosticErr != nil {
			return diagnosticErr
		}
		return realtimeVoiceProviderStageError{code: realtimeVoiceFailureTextToSpeech, err: err}
	}
	speechChunks := realtimeVoicePlayableSpeechChunks(speech.Chunks)
	if speech.MimeType == "" || len(speechChunks) == 0 {
		err := ports.ErrInvalidProviderInput
		if diagnosticErr := emitRealtimeVoiceTextToSpeechFailureDiagnostic(session, toolResults, realtimeVoiceFailureTextToSpeech, err, emit); diagnosticErr != nil {
			return diagnosticErr
		}
		return realtimeVoiceProviderStageError{code: realtimeVoiceFailureTextToSpeech, err: err}
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
	if !realtimeVoiceShouldContinueAfterClarification(response, continueAfterClarification...) {
		if err := a.markRealtimeVoiceSessionOutcome(ctx, session, ports.RealtimeSessionStateCompleted, ""); err != nil {
			return err
		}
	}
	return emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventSessionCompleted, SessionID: session.ID})
}

func realtimeVoicePlayableSpeechChunks(chunks [][]byte) [][]byte {
	playable := make([][]byte, 0, len(chunks))
	for _, chunk := range chunks {
		if len(chunk) > 0 {
			playable = append(playable, chunk)
		}
	}
	return playable
}

func (a App) recoverRealtimeVoiceResponse(ctx context.Context, session RealtimeVoiceSession, toolCallIDs []string, toolResults []ports.AgentToolResult, emit RealtimeVoiceEventSink) error {
	if err := emitRealtimeVoiceProgress(session, realtimeVoiceProgressRecovering, "Recovering safely.", emit); err != nil {
		return err
	}
	return a.completeRealtimeVoiceResponse(ctx, session, ports.StructuredAgentResponse{
		Kind:            ports.StructuredAgentResponseKindSafeFailure,
		SpokenResponse:  "I could not finish that voice request safely. Please try again with a little more detail.",
		DisplayResponse: "I could not finish that voice request safely. Please try again with a little more detail.",
	}, toolCallIDs, toolResults, emit)
}

func realtimeVoiceShouldContinueAfterClarification(response ports.StructuredAgentResponse, continueAfterClarification ...bool) bool {
	return len(continueAfterClarification) > 0 && continueAfterClarification[0] && response.Kind == ports.StructuredAgentResponseKindClarification
}
