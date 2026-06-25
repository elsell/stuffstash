package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const (
	RealtimeVoiceSourceMobile = "mobile_voice"

	RealtimeVoiceEventTranscriptFinal             = "transcript.final"
	RealtimeVoiceEventAgentProgress               = "agent.progress"
	RealtimeVoiceEventToolCallStarted             = "tool.call.started"
	RealtimeVoiceEventToolCallCompleted           = "tool.call.completed"
	RealtimeVoiceEventToolCallFailed              = "tool.call.failed"
	RealtimeVoiceEventAssistantResponseStarted    = "assistant.response.started"
	RealtimeVoiceEventAssistantResponseCompleted  = "assistant.response.completed"
	RealtimeVoiceEventTextToSpeechAudioStarted    = "tts.audio.started"
	RealtimeVoiceEventTextToSpeechAudioChunk      = "tts.audio.chunk"
	RealtimeVoiceEventTextToSpeechAudioCompleted  = "tts.audio.completed"
	RealtimeVoiceEventSessionCompleted            = "session.completed"
	RealtimeVoiceToolSearchAuthorizedAssets       = "search_authorized_assets"
	realtimeVoiceSearchAuthorizedAssetsPublicName = "Search inventory"
)

type RealtimeVoiceSessionInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Source      string
	InputAudio  ports.RealtimeAudioFormat
	OutputAudio RealtimeVoiceOutputAudio
}

type RealtimeVoiceOutputAudio struct {
	MimeTypes []string
}

type RealtimeVoiceSession struct {
	ID          string
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	Principal   identity.Principal
	Source      string
	InputAudio  ports.RealtimeAudioFormat
	OutputAudio RealtimeVoiceOutputAudio
}

type RealtimeVoiceQueryInput struct {
	Session     RealtimeVoiceSession
	AudioChunks [][]byte
}

type RealtimeVoiceEvent struct {
	Type       string
	SessionID  string
	ToolCallID string
	ToolLabel  string
	Status     string
	Code       string
	Message    string
	Text       string
	Response   *ports.StructuredAgentResponse
	Audio      []byte
	AudioMime  string
	ChunkID    string
	FinalChunk bool
}

type RealtimeVoiceEventSink func(RealtimeVoiceEvent) error

func (a App) WithRealtimeVoiceProviders(stt ports.SpeechToTextProvider, lm ports.LanguageInferenceProvider, tts ports.TextToSpeechProvider) App {
	a.speechToText = stt
	a.languageInference = lm
	a.textToSpeech = tts
	return a
}

func (a App) StartRealtimeVoiceSession(ctx context.Context, input RealtimeVoiceSessionInput) (RealtimeVoiceSession, error) {
	if err := a.ensureRealtimeVoiceDependencies(); err != nil {
		return RealtimeVoiceSession{}, err
	}
	if input.Source != RealtimeVoiceSourceMobile {
		return RealtimeVoiceSession{}, apperrors.ErrInvalidInput
	}
	if input.InputAudio.MimeType != "audio/mp4" || input.InputAudio.Channels != 1 {
		return RealtimeVoiceSession{}, apperrors.ErrInvalidInput
	}
	if err := a.authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionView, input.TenantID); err != nil {
		a.recordAuthorizationDenied(ctx, input.Principal, input.TenantID)
		return RealtimeVoiceSession{}, err
	}
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return RealtimeVoiceSession{}, err
	}

	sessionID := a.newRealtimeVoiceID()
	if strings.TrimSpace(sessionID) == "" {
		return RealtimeVoiceSession{}, apperrors.ErrInvalidInput
	}
	return RealtimeVoiceSession{
		ID:          sessionID,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Principal:   input.Principal,
		Source:      input.Source,
		InputAudio:  input.InputAudio,
		OutputAudio: input.OutputAudio,
	}, nil
}

func (a App) RunRealtimeVoiceQuery(ctx context.Context, input RealtimeVoiceQueryInput, emit RealtimeVoiceEventSink) error {
	if err := a.ensureRealtimeVoiceDependencies(); err != nil {
		return err
	}
	if len(input.AudioChunks) == 0 {
		return ports.ErrInvalidProviderInput
	}

	transcription, err := a.speechToText.Transcribe(ctx, ports.SpeechToTextInput{
		TenantID:    input.Session.TenantID,
		InventoryID: input.Session.InventoryID,
		Principal:   input.Session.Principal,
		AudioFormat: input.Session.InputAudio,
		AudioChunks: input.AudioChunks,
	})
	if err != nil {
		return err
	}
	transcript := strings.TrimSpace(transcription.Transcript)
	if transcript == "" {
		return ports.ErrInvalidProviderInput
	}
	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventTranscriptFinal, SessionID: input.Session.ID, Text: transcript}); err != nil {
		return err
	}
	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventAgentProgress, SessionID: input.Session.ID, Status: "thinking", Message: "Checking your inventory."}); err != nil {
		return err
	}

	toolResults := []ports.AgentToolResult{}
	toolCallIDs := []string{}
	for turn := 0; turn < 4; turn++ {
		modelTurn, err := a.languageInference.NextTurn(ctx, ports.LanguageInferenceInput{
			TenantID:      input.Session.TenantID,
			InventoryID:   input.Session.InventoryID,
			Principal:     input.Session.Principal,
			Transcript:    transcript,
			Tools:         realtimeVoiceToolDescriptors(),
			ToolResults:   toolResults,
			PreviousTurns: turn,
		})
		if err != nil {
			return err
		}
		if modelTurn.Final != nil {
			return a.completeRealtimeVoiceResponse(ctx, input.Session, *modelTurn.Final, toolCallIDs, emit)
		}
		if len(modelTurn.ToolCalls) == 0 {
			return ports.ErrInvalidProviderInput
		}
		for _, call := range modelTurn.ToolCalls {
			toolCallID := strings.TrimSpace(call.ID)
			if toolCallID == "" {
				toolCallID = a.newRealtimeVoiceID()
			}
			toolCallIDs = append(toolCallIDs, toolCallID)
			if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallStarted, SessionID: input.Session.ID, ToolCallID: toolCallID, ToolLabel: realtimeVoiceSearchAuthorizedAssetsPublicName, Status: "searching"}); err != nil {
				return err
			}
			result, err := a.executeRealtimeVoiceTool(ctx, input.Session, ports.AgentToolCall{
				ID:        toolCallID,
				Name:      call.Name,
				Arguments: call.Arguments,
			})
			if err != nil {
				_ = emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallFailed, SessionID: input.Session.ID, ToolCallID: toolCallID, ToolLabel: realtimeVoiceSearchAuthorizedAssetsPublicName, Code: "tool_failed", Message: "I could not check that safely."})
				return err
			}
			toolResults = append(toolResults, result)
			if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventToolCallCompleted, SessionID: input.Session.ID, ToolCallID: toolCallID, ToolLabel: realtimeVoiceSearchAuthorizedAssetsPublicName, Status: "completed"}); err != nil {
				return err
			}
		}
	}
	return ports.ErrInvalidProviderInput
}

func (a App) ensureRealtimeVoiceDependencies() error {
	if a.authorizer == nil || a.speechToText == nil || a.languageInference == nil || a.textToSpeech == nil {
		return apperrors.ErrInvalidInput
	}
	return nil
}

func (a App) executeRealtimeVoiceTool(ctx context.Context, session RealtimeVoiceSession, call ports.AgentToolCall) (ports.AgentToolResult, error) {
	if call.Name != RealtimeVoiceToolSearchAuthorizedAssets {
		return ports.AgentToolResult{}, ports.ErrForbidden
	}
	query, _ := call.Arguments["query"].(string)
	query = strings.TrimSpace(query)
	if query == "" {
		return ports.AgentToolResult{}, ports.ErrInvalidProviderInput
	}

	results, err := a.SearchAssets(ctx, SearchAssetsInput{
		Principal:      session.Principal,
		TenantID:       session.TenantID,
		InventoryIDs:   []inventory.InventoryID{session.InventoryID},
		Query:          query,
		Mode:           "fuzzy",
		LifecycleState: "active",
		Limit:          5,
	})
	if err != nil {
		return ports.AgentToolResult{}, err
	}
	safeSummaries := make([]string, 0, len(results.Items))
	for _, item := range results.Items {
		safeSummaries = append(safeSummaries, fmt.Sprintf("%s (%s)", item.Asset.Title.String(), item.Asset.Kind.String()))
	}
	if len(safeSummaries) == 0 {
		safeSummaries = append(safeSummaries, "No visible matches.")
	}
	return ports.AgentToolResult{
		CallID:  call.ID,
		Name:    call.Name,
		Content: strings.Join(safeSummaries, "\n"),
	}, nil
}

func (a App) completeRealtimeVoiceResponse(ctx context.Context, session RealtimeVoiceSession, response ports.StructuredAgentResponse, toolCallIDs []string, emit RealtimeVoiceEventSink) error {
	if strings.TrimSpace(response.SpokenResponse) == "" {
		return ports.ErrInvalidProviderInput
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

	speech, err := a.textToSpeech.Synthesize(ctx, ports.TextToSpeechInput{
		TenantID:    session.TenantID,
		InventoryID: session.InventoryID,
		Principal:   session.Principal,
		Text:        response.SpokenResponse,
		MimeTypes:   session.OutputAudio.MimeTypes,
	})
	if err != nil {
		return err
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
	return emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventSessionCompleted, SessionID: session.ID})
}

func realtimeVoiceToolDescriptors() []ports.AgentToolDescriptor {
	return []ports.AgentToolDescriptor{{
		Name:        RealtimeVoiceToolSearchAuthorizedAssets,
		Label:       realtimeVoiceSearchAuthorizedAssetsPublicName,
		Description: "Search visible assets in the selected inventory.",
		ReadOnly:    true,
	}}
}

func realtimeVoiceErrorCode(err error) string {
	switch {
	case errors.Is(err, ports.ErrUnauthenticated):
		return "unauthenticated"
	case errors.Is(err, ports.ErrForbidden), errors.Is(err, apperrors.ErrNotFound):
		return "forbidden"
	case errors.Is(err, ports.ErrInvalidProviderInput), errors.Is(err, apperrors.ErrInvalidInput):
		return "invalid_request"
	default:
		return "voice_session_failed"
	}
}

func (a App) newRealtimeVoiceID() string {
	if a.ids == nil {
		return ""
	}
	return a.ids.NewID()
}

func RealtimeVoiceSafeErrorCode(err error) string {
	return realtimeVoiceErrorCode(err)
}
