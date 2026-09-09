package app

import (
	"context"
	"testing"
)

func TestTypedTurnSkipsSpeechButNextAudioTurnSpeaks(t *testing.T) {
	resolver := successfulRealtimeVoiceResolver()
	speech := &resolvedTextToSpeech{}
	resolver.providers.TextToSpeech = speech
	application := newRealtimeVoiceResolutionTestApp(t, resolver)
	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatal(err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, Text: "Where are my tools?"}, func(event RealtimeVoiceEvent) error { events = append(events, event); return nil })
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Type == RealtimeVoiceEventTextToSpeechAudioStarted || event.Type == RealtimeVoiceEventTextToSpeechAudioChunk {
			t.Fatal("typed turn emitted speech")
		}
	}
	if speech.lastText != "" {
		t.Fatal("typed turn synthesized speech")
	}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if speech.lastText == "" {
		t.Fatal("audio turn did not synthesize speech")
	}
}
