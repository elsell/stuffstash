package voice

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiSpeechToTextTranscribesInlineAudio(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing bearer token: %q", r.Header.Get("Authorization"))
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
		_ = json.NewEncoder(w).Encode(map[string]any{
			"candidates": []map[string]any{{
				"content": map[string]any{
					"parts": []map[string]any{{"text": "Where are my tools?"}},
				},
			}},
		})
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiSpeechToText(GoogleGeminiConfig{
		ProjectID:   "project",
		Location:    "us-central1",
		Model:       "gemini-test",
		BaseURL:     server.URL,
		TokenSource: staticTokenSource{},
		HTTPClient:  server.Client(),
	})

	result, err := provider.Transcribe(context.Background(), ports.SpeechToTextInput{
		AudioFormat: ports.RealtimeAudioFormat{MimeType: "audio/mp4"},
		AudioChunks: [][]byte{[]byte("audio")},
	})
	if err != nil {
		t.Fatalf("transcribe: %v", err)
	}
	if result.Transcript != "Where are my tools?" {
		t.Fatalf("unexpected transcript %q", result.Transcript)
	}
}

func TestGoogleGeminiLanguageInferenceMapsToolAndFinalTurns(t *testing.T) {
	t.Parallel()

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"toolCalls":[{"id":"call-1","name":"search_authorized_assets","arguments":{"query":"tools"}}]}`))
			return
		}
		_ = json.NewEncoder(w).Encode(geminiTextResponse(`{"final":{"kind":"answer","spokenResponse":"Your tools are in Garage.","displayResponse":"Your tools are in Garage."}}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID:   "project",
		Location:    "us-central1",
		Model:       "gemini-test",
		BaseURL:     server.URL,
		TokenSource: staticTokenSource{},
		HTTPClient:  server.Client(),
	})

	toolTurn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{Transcript: "Where are my tools?"})
	if err != nil {
		t.Fatalf("first turn: %v", err)
	}
	if len(toolTurn.ToolCalls) != 1 || toolTurn.ToolCalls[0].Name != "search_authorized_assets" || toolTurn.ToolCalls[0].Arguments["query"] != "tools" {
		t.Fatalf("unexpected tool turn: %+v", toolTurn)
	}

	finalTurn, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{
		Transcript:    "Where are my tools?",
		ToolResults:   []ports.AgentToolResult{{CallID: "call-1", Name: "search_authorized_assets", Content: "Tools (container)"}},
		PreviousTurns: 1,
	})
	if err != nil {
		t.Fatalf("final turn: %v", err)
	}
	if finalTurn.Final == nil || finalTurn.Final.SpokenResponse != "Your tools are in Garage." {
		t.Fatalf("unexpected final turn: %+v", finalTurn)
	}
}

func TestGoogleTextToSpeechSynthesizesMP3(t *testing.T) {
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
		if !strings.Contains(string(payload), "Your tools are in Garage.") || !strings.Contains(string(payload), "MP3") {
			t.Fatalf("unexpected synthesize request: %s", string(payload))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"audioContent": base64.StdEncoding.EncodeToString([]byte("mp3-bytes")),
		})
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleTextToSpeech(GoogleTextToSpeechConfig{
		LanguageCode: "en-US",
		VoiceName:    "en-US-Neural2-F",
		BaseURL:      server.URL,
		TokenSource:  staticTokenSource{},
		HTTPClient:   server.Client(),
	})
	result, err := provider.Synthesize(context.Background(), ports.TextToSpeechInput{Text: "Your tools are in Garage."})
	if err != nil {
		t.Fatalf("synthesize: %v", err)
	}
	if result.MimeType != "audio/mpeg" || string(result.Chunks[0]) != "mp3-bytes" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func geminiTextResponse(text string) map[string]any {
	return map[string]any{
		"candidates": []map[string]any{{
			"content": map[string]any{
				"parts": []map[string]any{{"text": text}},
			},
		}},
	}
}

type staticTokenSource struct{}

func (staticTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "test-token", TokenType: "Bearer"}, nil
}
