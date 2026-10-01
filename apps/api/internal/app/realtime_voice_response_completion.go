package app

import (
	"context"

	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func realtimeResponseSession(session RealtimeVoiceSession) agentmodelapp.RealtimeResponseSession {
	return agentmodelapp.RealtimeResponseSession{ID: session.ID, TenantID: session.TenantID, InventoryID: session.InventoryID, Principal: session.Principal, Source: session.Source, OutputMimeTypes: session.OutputAudio.MimeTypes, SilentReply: session.silentReply, ConversationContinuity: session.ConversationContinuity, DeveloperDiagnostics: session.DeveloperDiagnostics, TextToSpeech: session.textToSpeech}
}
func (a App) completeRealtimeVoiceResponse(ctx context.Context, session RealtimeVoiceSession, response ports.StructuredAgentResponse, toolCallIDs []string, toolResults []ports.AgentToolResult, emit RealtimeVoiceEventSink, continueAfterClarification ...bool) error {
	return agentmodelapp.NewRealtimeResponseService(a.ids, a.realtimeSessionService()).Complete(ctx, realtimeResponseSession(session), response, toolCallIDs, toolResults, emit, continueAfterClarification...)
}
func (a App) recoverRealtimeVoiceResponse(ctx context.Context, session RealtimeVoiceSession, toolCallIDs []string, toolResults []ports.AgentToolResult, emit RealtimeVoiceEventSink) error {
	return agentmodelapp.NewRealtimeResponseService(a.ids, a.realtimeSessionService()).Recover(ctx, realtimeResponseSession(session), toolCallIDs, toolResults, emit)
}
func realtimeVoicePlayableSpeechChunks(chunks [][]byte) [][]byte {
	return agentmodelapp.RealtimeVoicePlayableSpeechChunks(chunks)
}
func realtimeVoiceShouldContinueAfterClarification(response ports.StructuredAgentResponse, continueAfterClarification ...bool) bool {
	return agentmodelapp.RealtimeVoiceShouldContinueAfterClarification(response, continueAfterClarification...)
}
