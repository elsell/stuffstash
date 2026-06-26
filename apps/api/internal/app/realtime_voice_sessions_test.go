package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceSessionPersistsStartedAndCompletedSafeMetadata(t *testing.T) {
	t.Parallel()

	repository := newFakeRealtimeSessionRepository()
	application := newRealtimeVoiceResolutionTestAppWithSessions(t, successfulRealtimeVoiceResolver(), repository)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}

	started := repository.savedRecord(t, session.ID)
	if started.State != ports.RealtimeSessionStateStarted {
		t.Fatalf("expected started state, got %+v", started)
	}
	if started.TenantID != session.TenantID || started.InventoryID != session.InventoryID || started.PrincipalID != session.Principal.ID || started.Source != session.Source {
		t.Fatalf("expected session scope metadata, got %+v", started)
	}
	if started.SpeechToTextProfileID != "stt-profile" || started.LanguageInferenceProfileID != "lm-profile" || started.TextToSpeechProfileID != "tts-profile" {
		t.Fatalf("expected selected provider profile IDs, got %+v", started)
	}
	if started.StartedAt.IsZero() || !started.EndedAt.IsZero() || started.SafeFailureCode != "" {
		t.Fatalf("expected started timestamps without final outcome, got %+v", started)
	}

	if err := application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error {
		return nil
	}); err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}

	completed := repository.savedRecord(t, session.ID)
	if completed.State != ports.RealtimeSessionStateCompleted || completed.EndedAt.IsZero() || completed.SafeFailureCode != "" {
		t.Fatalf("expected completed outcome, got %+v", completed)
	}
	if completed.LastActivityAt.Before(completed.StartedAt) || completed.EndedAt.Before(completed.StartedAt) {
		t.Fatalf("expected monotonic session timestamps, got %+v", completed)
	}
}

func TestRealtimeVoiceSessionPersistsFailureWithSafeCode(t *testing.T) {
	t.Parallel()

	repository := newFakeRealtimeSessionRepository()
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{err: ports.ErrInvalidProviderInput}
	application := newRealtimeVoiceResolutionTestAppWithSessions(t, resolver, repository)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(RealtimeVoiceEvent) error {
		return nil
	})
	if !errors.Is(err, ports.ErrInvalidProviderInput) {
		t.Fatalf("expected provider error, got %v", err)
	}

	failed := repository.savedRecord(t, session.ID)
	if failed.State != ports.RealtimeSessionStateFailed || failed.SafeFailureCode != "invalid_request" || failed.EndedAt.IsZero() {
		t.Fatalf("expected failed outcome with safe code, got %+v", failed)
	}
}

func successfulRealtimeVoiceResolver() *fakeRealtimeVoiceProviderResolver {
	return &fakeRealtimeVoiceProviderResolver{
		providers: ports.RealtimeVoiceProviderSet{
			SpeechToTextProfileID:      "stt-profile",
			LanguageInferenceProfileID: "lm-profile",
			TextToSpeechProfileID:      "tts-profile",
			LanguagePromptTemplate:     "Prefer concise spoken answers.",
			SpeechToText:               resolvedSpeechToText{transcript: "Where are my tools?"},
			LanguageInference:          &resolvedLanguageInference{response: "The tools are in the office."},
			TextToSpeech:               &resolvedTextToSpeech{},
		},
	}
}

type fakeRealtimeSessionRepository struct {
	records map[string]ports.RealtimeSessionRecord
}

func newFakeRealtimeSessionRepository() *fakeRealtimeSessionRepository {
	return &fakeRealtimeSessionRepository{records: map[string]ports.RealtimeSessionRecord{}}
}

func (r *fakeRealtimeSessionRepository) SaveRealtimeSession(_ context.Context, record ports.RealtimeSessionRecord) error {
	r.records[record.ID] = record
	return nil
}

func (r *fakeRealtimeSessionRepository) UpdateRealtimeSessionOutcome(_ context.Context, _ tenant.ID, _ inventory.InventoryID, sessionID string, outcome ports.RealtimeSessionOutcome) error {
	record := r.records[sessionID]
	record.State = outcome.State
	record.LastActivityAt = outcome.At
	record.EndedAt = outcome.At
	record.SafeFailureCode = outcome.SafeFailureCode
	r.records[sessionID] = record
	return nil
}

func (r *fakeRealtimeSessionRepository) RealtimeSessionByID(_ context.Context, _ tenant.ID, _ inventory.InventoryID, sessionID string) (ports.RealtimeSessionRecord, bool, error) {
	record, found := r.records[sessionID]
	return record, found, nil
}

func (r *fakeRealtimeSessionRepository) savedRecord(t *testing.T, sessionID string) ports.RealtimeSessionRecord {
	t.Helper()

	record, found := r.records[sessionID]
	if !found {
		t.Fatalf("expected persisted session %q", sessionID)
	}
	return record
}

type fixedRealtimeClock struct {
	now time.Time
}

func (c fixedRealtimeClock) Now() time.Time {
	return c.now
}
