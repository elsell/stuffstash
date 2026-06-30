package app

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceSessionResolvesAndUsesSessionProviders(t *testing.T) {
	t.Parallel()

	language := &resolvedLanguageInference{
		response: "The tools are in the office.",
	}
	resolver := &fakeRealtimeVoiceProviderResolver{
		providers: ports.RealtimeVoiceProviderSet{
			SpeechToTextProfileID:      "stt-profile",
			LanguageInferenceProfileID: "lm-profile",
			TextToSpeechProfileID:      "tts-profile",
			LanguagePromptTemplate:     "Prefer concise spoken answers.",
			SpeechToText:               resolvedSpeechToText{transcript: "Where are my tools?"},
			LanguageInference:          language,
			TextToSpeech:               &resolvedTextToSpeech{},
		},
	}
	application := newRealtimeVoiceResolutionTestApp(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), defaultRealtimeVoiceSessionInput())
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	if resolver.calls != 1 || resolver.lastInput.TenantID != tenant.ID("tenant-home") || resolver.lastInput.InventoryID != inventory.InventoryID("inventory-home") {
		t.Fatalf("resolver was not called with session scope: calls=%d input=%+v", resolver.calls, resolver.lastInput)
	}
	if session.SpeechToTextProfileID != "stt-profile" || session.LanguageInferenceProfileID != "lm-profile" || session.TextToSpeechProfileID != "tts-profile" {
		t.Fatalf("expected selected provider profile IDs on session, got %+v", session)
	}

	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(context.Background(), RealtimeVoiceQueryInput{
		Session:     session,
		AudioChunks: [][]byte{[]byte("audio")},
	}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("run realtime voice query: %v", err)
	}
	if len(events) == 0 || events[0].Type != RealtimeVoiceEventTranscriptFinal || events[0].Text != "Where are my tools?" {
		t.Fatalf("expected transcript from resolved speech-to-text provider, got %+v", events)
	}
	tts := resolver.providers.TextToSpeech.(*resolvedTextToSpeech)
	if tts.lastText != "The tools are in the office." {
		t.Fatalf("expected resolved text-to-speech provider to receive final response, got %q", tts.lastText)
	}
	if language.lastPromptTemplate != "Prefer concise spoken answers." {
		t.Fatalf("expected language prompt template to be passed to model, got %q", language.lastPromptTemplate)
	}
}

func defaultRealtimeVoiceSessionInput() RealtimeVoiceSessionInput {
	return RealtimeVoiceSessionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		Source:      RealtimeVoiceSourceMobile,
		InputAudio:  ports.RealtimeAudioFormat{MimeType: "audio/mp4", Channels: 1},
		OutputAudio: RealtimeVoiceOutputAudio{MimeTypes: []string{"audio/mpeg"}},
	}
}

func TestRealtimeVoiceSessionFailsWhenProviderResolverUnavailable(t *testing.T) {
	t.Parallel()

	application := newRealtimeVoiceResolutionTestApp(t, nil)
	_, err := application.StartRealtimeVoiceSession(context.Background(), RealtimeVoiceSessionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		Source:      RealtimeVoiceSourceMobile,
		InputAudio:  ports.RealtimeAudioFormat{MimeType: "audio/mp4", Channels: 1},
		OutputAudio: RealtimeVoiceOutputAudio{MimeTypes: []string{"audio/mpeg"}},
	})
	if err == nil {
		t.Fatalf("expected missing provider resolver to fail")
	}
}

func newRealtimeVoiceResolutionTestApp(t *testing.T, resolver ports.RealtimeVoiceProviderResolver) App {
	application, _ := newRealtimeVoiceResolutionTestAppWithStoreAndSessions(t, resolver, newFakeRealtimeSessionRepository())
	return application
}

func newRealtimeVoiceResolutionTestAppWithSessions(t *testing.T, resolver ports.RealtimeVoiceProviderResolver, sessions ports.RealtimeSessionRepository) App {
	application, _ := newRealtimeVoiceResolutionTestAppWithStoreAndSessions(t, resolver, sessions)
	return application
}

func newRealtimeVoiceResolutionTestAppWithStore(t *testing.T, resolver ports.RealtimeVoiceProviderResolver) (App, *memory.Store) {
	return newRealtimeVoiceResolutionTestAppWithStoreAndSessions(t, resolver, newFakeRealtimeSessionRepository())
}

func newRealtimeVoiceResolutionTestAppWithStoreAndSessions(t *testing.T, resolver ports.RealtimeVoiceProviderResolver, sessions ports.RealtimeSessionRepository) (App, *memory.Store) {
	return newRealtimeVoiceResolutionTestAppWithStoreSessionsAndIDs(t, resolver, sessions, &realtimeVoiceResolutionIDGenerator{})
}

func newRealtimeVoiceResolutionTestAppWithStoreSessionsAndIDs(t *testing.T, resolver ports.RealtimeVoiceProviderResolver, sessions ports.RealtimeSessionRepository, ids ports.IDGenerator) (App, *memory.Store) {
	t.Helper()

	ctx := context.Background()
	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	name, ok := tenant.NewName("Home")
	if !ok {
		t.Fatalf("invalid tenant name")
	}
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: tenant.ID("tenant-home"), Name: name}); err != nil {
		t.Fatalf("save tenant: %v", err)
	}
	inventoryName, ok := inventory.NewName("Home")
	if !ok {
		t.Fatalf("invalid inventory name")
	}
	if err := store.SaveInventory(ctx, inventory.Inventory{
		ID:       inventory.InventoryID("inventory-home"),
		TenantID: inventory.TenantID("tenant-home"),
		Name:     inventoryName,
	}); err != nil {
		t.Fatalf("save inventory: %v", err)
	}
	principal := identity.Principal{ID: identity.PrincipalID("user-1")}
	if err := authorizer.GrantTenantOwner(ctx, principal, tenant.ID("tenant-home")); err != nil {
		t.Fatalf("grant tenant owner: %v", err)
	}
	if err := authorizer.GrantInventoryOwner(ctx, principal, tenant.ID("tenant-home"), inventory.InventoryID("inventory-home")); err != nil {
		t.Fatalf("grant inventory owner: %v", err)
	}

	application := New(Dependencies{
		Authorizer:                    authorizer,
		Tenants:                       store,
		Inventories:                   store,
		Assets:                        store,
		Search:                        store,
		Audit:                         store,
		Observer:                      &fakeObserver{},
		ActionPlans:                   store,
		RealtimeVoiceProviderResolver: resolver,
		RealtimeSessions:              sessions,
		IDs:                           ids,
		Clock:                         fixedRealtimeClock{now: time.Date(2026, 6, 26, 16, 0, 0, 0, time.UTC)},
	})
	return application, store
}

type fakeRealtimeVoiceProviderResolver struct {
	providers ports.RealtimeVoiceProviderSet
	calls     int
	lastInput ports.RealtimeVoiceProviderResolutionInput
}

func (f *fakeRealtimeVoiceProviderResolver) ResolveRealtimeVoiceProviders(_ context.Context, input ports.RealtimeVoiceProviderResolutionInput) (ports.RealtimeVoiceProviderSet, error) {
	f.calls++
	f.lastInput = input
	return f.providers, nil
}

type resolvedSpeechToText struct {
	transcript string
	err        error
}

func (r resolvedSpeechToText) Transcribe(context.Context, ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	if r.err != nil {
		return ports.SpeechToTextResult{}, r.err
	}
	return ports.SpeechToTextResult{Transcript: r.transcript}, nil
}

type resolvedLanguageInference struct {
	response           string
	lastPromptTemplate string
}

func (r *resolvedLanguageInference) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	r.lastPromptTemplate = input.PromptTemplate
	return ports.LanguageInferenceTurn{Final: &ports.StructuredAgentResponse{
		Kind:            ports.StructuredAgentResponseKindAnswer,
		SpokenResponse:  r.response,
		DisplayResponse: r.response,
	}}, nil
}

type failingResolvedLanguageInference struct {
	err error
}

func (r *failingResolvedLanguageInference) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{}, r.err
}

type resolvedTextToSpeech struct {
	lastText string
}

func (r *resolvedTextToSpeech) Synthesize(_ context.Context, input ports.TextToSpeechInput) (ports.TextToSpeechResult, error) {
	r.lastText = input.Text
	return ports.TextToSpeechResult{MimeType: "audio/mpeg", Chunks: [][]byte{[]byte("speech")}}, nil
}

type realtimeVoiceResolutionIDGenerator struct {
	counter int
}

func (g *realtimeVoiceResolutionIDGenerator) NewID() string {
	g.counter++
	return fmt.Sprintf("voice-resolution-id-%d", g.counter)
}
