package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/voice"
)

func TestGoogleVoiceProjectsProductionProposalCatalog(t *testing.T) {
	projected := false
	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Tools []struct {
				FunctionDeclarations []struct {
					Name       string
					Parameters map[string]any `json:"parametersJsonSchema"`
				}
			}
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			http.Error(w, "bad request", 400)
			return
		}
		for _, call := range request.Tools[0].FunctionDeclarations {
			if call.Name != "propose_inventory_change" {
				continue
			}
			args := call.Parameters
			commands := args["properties"].(map[string]any)["commands"].(map[string]any)
			if commands["maxItems"] != nil {
				t.Error("Google decoding bound unexpectedly restored")
			}
			items := commands["items"].(map[string]any)
			if items["anyOf"] != nil {
				t.Error("production command union was not projected")
				continue
			}
			projected = true
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"candidates": []any{map[string]any{"finishReason": "STOP", "content": map[string]any{"role": "model", "parts": []any{map[string]any{"functionCall": map[string]any{"name": "present_answer", "args": map[string]any{"spoken": "Hello.", "display": "Hello."}}}}}}}})
	}))
	defer providerServer.Close()
	model := voice.NewGoogleGeminiLanguageInference(voice.GoogleGeminiConfig{BaseURL: providerServer.URL, APIKey: "fixture", Model: "fixture"})
	application := newSeededTestAppWithVoice(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}, fakeSpeechToText{transcript: "Hello"}, model, fakeTextToSpeech{chunks: [][]byte{[]byte("audio")}})
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	defer server.Close()
	runRealtimeVoiceQuestion(t, server.URL, "tenant-home", "inventory-home", "user-1")
	if !projected {
		t.Fatal("production proposal not observed with flat shape")
	}
}
