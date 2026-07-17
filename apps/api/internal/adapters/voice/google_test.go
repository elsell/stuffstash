package voice

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiSpeechToTextTranscribesInlineAudio(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing bearer token: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Goog-User-Project") != "project" {
			t.Fatalf("missing quota project: %q", r.Header.Get("X-Goog-User-Project"))
		}
		if !strings.Contains(r.URL.Path, "/publishers/google/models/gemini-test:generateContent") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		payload, _ := json.Marshal(request)
		if !strings.Contains(string(payload), "audio/mp4") || !strings.Contains(string(payload), base64.StdEncoding.EncodeToString([]byte("audio"))) {
			t.Fatalf("request did not include inline audio: %s", string(payload))
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse("Where are my tools?"))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiSpeechToText(GoogleGeminiConfig{
		ProjectID: "project", Location: "us-central1", Model: "gemini-test", QuotaProject: "project",
		BaseURL: server.URL, TokenSource: staticTokenSource{}, HTTPClient: server.Client(),
	})
	result, err := provider.Transcribe(context.Background(), ports.SpeechToTextInput{
		AudioFormat: ports.RealtimeAudioFormat{MimeType: "audio/mp4"}, AudioChunks: [][]byte{[]byte("audio")},
	})
	if err != nil {
		t.Fatalf("transcribe: %v", err)
	}
	if result.Transcript != "Where are my tools?" {
		t.Fatalf("unexpected transcript %q", result.Transcript)
	}
}

func TestGoogleGeminiSpeechToTextUsesAPIKeyBackend(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("api-key backend must not send bearer authorization: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Goog-Api-Key") != "test-api-key" {
			t.Fatalf("missing api key header: %q", r.Header.Get("X-Goog-Api-Key"))
		}
		if r.URL.Path != "/v1beta/models/gemini-test:generateContent" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse("Where are my tools?"))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiSpeechToText(GoogleGeminiConfig{
		Model: "gemini-test", BaseURL: server.URL, APIKey: "test-api-key", HTTPClient: server.Client(),
	})
	result, err := provider.Transcribe(context.Background(), ports.SpeechToTextInput{
		AudioFormat: ports.RealtimeAudioFormat{MimeType: "audio/mp4"}, AudioChunks: [][]byte{[]byte("audio")},
	})
	if err != nil {
		t.Fatalf("transcribe with api key: %v", err)
	}
	if result.Transcript != "Where are my tools?" {
		t.Fatalf("unexpected transcript %q", result.Transcript)
	}
}

func TestGoogleGeminiSpeechToTextProbeUsesModelEndpointWithoutTenantAudio(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/publishers/google/models/gemini-test:generateContent") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		payload, _ := json.Marshal(request)
		if strings.Contains(string(payload), "inlineData") || !strings.Contains(string(payload), "Provider diagnostic") {
			t.Fatalf("probe request should use safe text-only diagnostic: %s", string(payload))
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse("ready"))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiSpeechToText(GoogleGeminiConfig{
		ProjectID: "project", Location: "us-central1", Model: "gemini-test", QuotaProject: "project",
		BaseURL: server.URL, TokenSource: staticTokenSource{}, HTTPClient: server.Client(),
	})
	if err := provider.ProbeSpeechToText(context.Background()); err != nil {
		t.Fatalf("probe speech-to-text: %v", err)
	}
}

func TestGoogleGeminiLanguageInferenceUsesAPIKeyBackend(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" || r.Header.Get("X-Goog-Api-Key") != "test-api-key" {
			t.Fatalf("unexpected authentication headers")
		}
		if r.URL.Path != "/v1beta/models/gemini-test:generateContent" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{
          "decision":"search",
          "intent":{"requestShape":"single_target","kind":"read","operation":"locate","subjectMention":"tools","newAssetKind":"","destinationPath":[],"destinationKinds":[],"details":""},
          "searchRequests":[{"referenceKey":"subject","readKind":"search_assets","mention":"tools","kindHint":"","visibleAssetId":"","searchProbes":["tools"],"lifecycleScope":"active"}],
          "vocabularyRequests":[],"resolutions":[],"rationale":"Gather candidates."
        }`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		Model: "gemini-test", BaseURL: server.URL, APIKey: "test-api-key", HTTPClient: server.Client(),
	})
	turn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{
		Transcript: "Where are my tools?",
		Investigation: &agentmodel.InvestigationInput{
			Phase: agentmodel.InvestigationPhaseInitial, PromptVersion: "voice-investigation-v1", SchemaVersion: "voice-investigation-v1",
			Transcript: "Where are my tools?", MaxEvidenceRounds: agentmodel.MaxEvidenceRounds,
		},
	})
	if err != nil || turn.Investigation == nil {
		t.Fatalf("language inference with API key: turn=%+v err=%v", turn, err)
	}
}

func TestGoogleGeminiLanguageInferenceProbeUsesSeparateTextDiagnostic(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		payload, _ := json.Marshal(request)
		if strings.Contains(string(payload), "responseJsonSchema") || !strings.Contains(string(payload), "Provider diagnostic") {
			t.Fatalf("probe must use the separate text-only diagnostic path: %s", payload)
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse("ready"))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID: "project", Location: "us-central1", Model: "gemini-test", QuotaProject: "project",
		BaseURL: server.URL, TokenSource: staticTokenSource{}, HTTPClient: server.Client(),
	})
	if err := provider.ProbeLanguageInference(context.Background()); err != nil {
		t.Fatalf("probe language inference: %v", err)
	}
}

func TestGoogleTextToSpeechSynthesizesMP3(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing bearer token: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Goog-User-Project") != "project" {
			t.Fatalf("missing quota project: %q", r.Header.Get("X-Goog-User-Project"))
		}
		if r.URL.Path != "/v1/text:synthesize" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		payload, _ := json.Marshal(request)
		if !strings.Contains(string(payload), "Your tools are in Garage.") || !strings.Contains(string(payload), "MP3") {
			t.Fatalf("unexpected synthesize request: %s", string(payload))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"audioContent": base64.StdEncoding.EncodeToString([]byte("mp3-bytes"))})
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleTextToSpeech(GoogleTextToSpeechConfig{
		LanguageCode: "en-US", VoiceName: "en-US-Neural2-F", QuotaProject: "project",
		BaseURL: server.URL, TokenSource: staticTokenSource{}, HTTPClient: server.Client(),
	})
	result, err := provider.Synthesize(context.Background(), ports.TextToSpeechInput{Text: "Your tools are in Garage."})
	if err != nil {
		t.Fatalf("synthesize: %v", err)
	}
	if result.MimeType != "audio/mpeg" || string(result.Chunks[0]) != "mp3-bytes" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestGoogleTextToSpeechProbeSynthesizesSafeDiagnosticPhrase(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/text:synthesize" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		payload, _ := json.Marshal(request)
		if !strings.Contains(string(payload), "Stuff Stash provider test.") {
			t.Fatalf("probe request should synthesize safe diagnostic phrase: %s", string(payload))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"audioContent": base64.StdEncoding.EncodeToString([]byte("mp3-bytes"))})
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleTextToSpeech(GoogleTextToSpeechConfig{
		LanguageCode: "en-US", VoiceName: "en-US-Neural2-F", QuotaProject: "project",
		BaseURL: server.URL, TokenSource: staticTokenSource{}, HTTPClient: server.Client(),
	})
	if err := provider.ProbeTextToSpeech(context.Background()); err != nil {
		t.Fatalf("probe text-to-speech: %v", err)
	}
}

func geminiTextResponse(text string) map[string]any {
	return map[string]any{"candidates": []map[string]any{{"content": map[string]any{"parts": []map[string]any{{"text": text}}}}}}
}
