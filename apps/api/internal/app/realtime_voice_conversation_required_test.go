package app

import (
	"context"
	"testing"
)

func TestRealtimeVoiceRejectsRetiredInferenceOnlyProviderBeforeStartingSession(t *testing.T) {
	t.Parallel()
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.ConversationModel = nil
	resolver.providers.LanguageInference = &resolvedLanguageInference{}
	resolver.providers.ResponseGenerator = &resolvedLanguageInference{}
	repository := newFakeRealtimeSessionRepository()
	application := newRealtimeVoiceResolutionTestAppWithSessions(t, resolver, repository)
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err == nil || session.ID != "" {
		t.Fatalf("retired inference-only provider started a voice session: id=%q err=%v", session.ID, err)
	}
}
