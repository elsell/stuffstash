package app

import (
	"context"
	"fmt"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) completeRealtimeVoiceResponse(ctx context.Context, session RealtimeVoiceSession, response ports.StructuredAgentResponse, toolCallIDs []string, emit RealtimeVoiceEventSink) error {
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
		return realtimeVoiceProviderStageError{code: realtimeVoiceFailureTextToSpeech, err: err}
	}
	if speech.MimeType == "" || len(speech.Chunks) == 0 {
		return ports.ErrInvalidProviderInput
	}
	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventTextToSpeechAudioStarted, SessionID: session.ID, AudioMime: speech.MimeType}); err != nil {
		return err
	}
	for index, chunk := range speech.Chunks {
		if len(chunk) == 0 {
			continue
		}
		if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventTextToSpeechAudioChunk, SessionID: session.ID, ChunkID: fmt.Sprintf("tts-%d", index+1), Audio: chunk, FinalChunk: index == len(speech.Chunks)-1}); err != nil {
			return err
		}
	}
	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventTextToSpeechAudioCompleted, SessionID: session.ID}); err != nil {
		return err
	}
	if err := a.markRealtimeVoiceSessionOutcome(ctx, session, ports.RealtimeSessionStateCompleted, ""); err != nil {
		return err
	}
	return emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventSessionCompleted, SessionID: session.ID})
}

func (a App) recoverRealtimeVoiceResponse(ctx context.Context, session RealtimeVoiceSession, toolCallIDs []string, emit RealtimeVoiceEventSink) error {
	if err := emitRealtimeVoiceProgress(session, realtimeVoiceProgressRecovering, "Recovering safely.", emit); err != nil {
		return err
	}
	return a.completeRealtimeVoiceResponse(ctx, session, ports.StructuredAgentResponse{
		Kind:            ports.StructuredAgentResponseKindSafeFailure,
		SpokenResponse:  "I could not finish that voice request safely. Please try again with a little more detail.",
		DisplayResponse: "I could not finish that voice request safely. Please try again with a little more detail.",
	}, toolCallIDs, emit)
}
