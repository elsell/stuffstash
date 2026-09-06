package app

import (
	"context"
	"testing"
)

func TestRealtimeVoiceRejectsMissingConversationModelBeforeStartingSession(t *testing.T) {
	t.Parallel()
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.ConversationModel = nil
	repository := newFakeRealtimeSessionRepository()
	application := newRealtimeVoiceResolutionTestAppWithSessions(t, resolver, repository)
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err == nil || session.ID != "" {
		t.Fatalf("missing conversation provider started a voice session: id=%q err=%v", session.ID, err)
	}
}
