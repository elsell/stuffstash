package app

import (
	"context"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceSessionResolvesAndUsesSessionProviders(t *testing.T) {
	t.Parallel()

	resolver := &fakeRealtimeVoiceProviderResolver{
		providers: ports.RealtimeVoiceProviderSet{
			SpeechToTextProfileID:      "stt-profile",
			LanguageInferenceProfileID: "lm-profile",
			TextToSpeechProfileID:      "tts-profile",
			SpeechToText:               resolvedSpeechToText{transcript: "Where are my tools?"},
			LanguageInference: resolvedLanguageInference{
				response: "The tools are in the office.",
			},
			TextToSpeech: &resolvedTextToSpeech{},
		},
	}
	application := newRealtimeVoiceResolutionTestApp(t, resolver)

	session, err := application.StartRealtimeVoiceSession(context.Background(), RealtimeVoiceSessionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-1")},
		TenantID:    tenant.ID("tenant-home"),
		InventoryID: inventory.InventoryID("inventory-home"),
		Source:      RealtimeVoiceSourceMobile,
		InputAudio:  ports.RealtimeAudioFormat{MimeType: "audio/mp4", Channels: 1},
		OutputAudio: RealtimeVoiceOutputAudio{MimeTypes: []string{"audio/mpeg"}},
	})
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

	return New(Dependencies{
		Authorizer:                    authorizer,
		Tenants:                       store,
		Inventories:                   store,
		Assets:                        store,
		Search:                        store,
		RealtimeVoiceProviderResolver: resolver,
		IDs:                           &realtimeVoiceResolutionIDGenerator{},
	})
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
}

func (r resolvedSpeechToText) Transcribe(context.Context, ports.SpeechToTextInput) (ports.SpeechToTextResult, error) {
	return ports.SpeechToTextResult{Transcript: r.transcript}, nil
}

type resolvedLanguageInference struct {
	response string
}

func (r resolvedLanguageInference) NextTurn(context.Context, ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	return ports.LanguageInferenceTurn{Final: &ports.StructuredAgentResponse{
		Kind:            ports.StructuredAgentResponseKindAnswer,
		SpokenResponse:  r.response,
		DisplayResponse: r.response,
	}}, nil
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
	return "voice-resolution-id"
}
