package agentmodel

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const MaxRealtimeTextCharacters = 8000

type RealtimeQueryInput struct {
	Session      RealtimeConversationSession
	Text         string
	InputAudio   ports.RealtimeAudioFormat
	AudioChunks  [][]byte
	SpeechToText ports.SpeechToTextProvider
	TextToSpeech ports.TextToSpeechProvider
}

type RealtimeQueryService struct {
	Sessions       RealtimeSessionService
	Observer       ports.Observer
	CleanupTimeout time.Duration
	Configured     bool
}

type RealtimeQueryContinuation func(context.Context, string, bool) error

func (a RealtimeQueryService) Run(ctx context.Context, input RealtimeQueryInput, emit RealtimeVoiceEventSink, continueConversation RealtimeQueryContinuation) (err error) {
	duration := time.Minute
	if input.Session.Workflow != nil {
		duration = time.Duration(input.Session.Workflow.Revision().Snapshot().Definition.Settings().Budget.ElapsedSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	if input.Session.Model != nil && !input.Session.Memory.Matches(realtimeConversationScope(input.Session)) {
		return ports.ErrForbidden
	}
	defer func() {
		if err != nil && strings.TrimSpace(input.Session.ID) != "" {
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.CleanupTimeout)
			defer cancel()
			if errors.Is(err, context.Canceled) {
				_ = a.Sessions.UpdateOutcome(cleanupCtx, input.Session.TenantID, input.Session.InventoryID, input.Session.ID, ports.RealtimeSessionStateCancelled, "")
				return
			}
			safeCode := RealtimeVoiceErrorCode(err)
			if a.Observer != nil {
				a.Observer.Record(ctx, ports.Event{
					Name:    ports.EventRealtimeVoiceFailed,
					Message: "realtime voice failed safely",
					Fields: map[string]string{
						"tenant_id":         input.Session.TenantID.String(),
						"inventory_id":      input.Session.InventoryID.String(),
						"principal_id":      input.Session.Principal.ID.String(),
						"session_id":        input.Session.ID,
						"safe_failure_code": safeCode,
						"error":             SafeRealtimeVoiceErrorDetail(err),
					},
				})
			}
			_ = a.Sessions.UpdateOutcome(cleanupCtx, input.Session.TenantID, input.Session.InventoryID, input.Session.ID, ports.RealtimeSessionStateFailed, safeCode)
		}
	}()
	if !a.Configured {
		return apperrors.ErrInvalidInput
	}
	if (len(input.AudioChunks) == 0) == (strings.TrimSpace(input.Text) == "") || utf8.RuneCountInString(input.Text) > MaxRealtimeTextCharacters {
		return ports.ErrInvalidProviderInput
	}

	if input.SpeechToText == nil || input.TextToSpeech == nil || input.Session.Model == nil {
		return apperrors.ErrInvalidInput
	}
	if err := a.Sessions.EnsureAccess(ctx, input.Session.Principal, input.Session.TenantID, input.Session.InventoryID); err != nil {
		return err
	}
	transcript := strings.TrimSpace(input.Text)
	if transcript == "" {
		transcription, err := input.SpeechToText.Transcribe(ctx, ports.SpeechToTextInput{
			TenantID:    input.Session.TenantID,
			InventoryID: input.Session.InventoryID,
			Principal:   input.Session.Principal,
			AudioFormat: input.InputAudio,
			AudioChunks: input.AudioChunks,
		})
		if err != nil {
			return RealtimeVoiceProviderStageError{Code: RealtimeVoiceFailureSpeechToText, Cause: err}
		}
		transcript = strings.TrimSpace(transcription.Transcript)
	}
	if transcript == "" {
		return ports.ErrInvalidProviderInput
	}
	if err := emit(RealtimeVoiceEvent{Type: RealtimeVoiceEventTranscriptFinal, SessionID: input.Session.ID, Text: transcript}); err != nil {
		return err
	}
	return continueConversation(ctx, transcript, strings.TrimSpace(input.Text) != "")
}
