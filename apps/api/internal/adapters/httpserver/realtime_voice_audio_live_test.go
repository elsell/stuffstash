package httpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"

	"github.com/stuffstash/stuff-stash/internal/app"
)

// This opt-in regression uses real audio and providers, but isolated inventory and
// local development authentication. It does not claim deployed or device coverage.
func TestGoogleLiveRealtimeAudioLocatesBabyClothes(t *testing.T) {
	runGoogleLiveRealtimeAudio(t, "STUFF_STASH_VOICE_BABY_CLOTHES_AUDIO_FILE", seedLiveBabyClothes)
}

func TestGoogleLiveRealtimeAudioFindsChemicals(t *testing.T) {
	runGoogleLiveRealtimeAudio(t, "STUFF_STASH_VOICE_CHEMICALS_AUDIO_FILE", seedLiveChemicals)
}

func TestGoogleLiveRealtimeAudioFindsChemicalsWithoutLiteralMatches(t *testing.T) {
	runGoogleLiveRealtimeAudio(t, "STUFF_STASH_VOICE_CHEMICALS_AUDIO_FILE", seedLiveChemicalsWithoutLiteralMatches)
}

type liveAudioFixture struct {
	expectedIDs       []string
	excludedIDs       []string
	spokenLocations   []string
	artifactLocations map[string]string
}

type liveAudioSeeder func(*testing.T, context.Context, app.App) liveAudioFixture

func runGoogleLiveRealtimeAudio(t *testing.T, audioFileKey string, seed liveAudioSeeder) {
	if os.Getenv("STUFF_STASH_GOOGLE_LIVE_TESTS") != "1" {
		t.Skip("explicit live Google opt-in required")
	}
	inputAudio, err := os.ReadFile(liveVoiceRequired(t, audioFileKey))
	if err != nil || len(inputAudio) == 0 {
		t.Fatalf("read recorded MP4 input: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	providers := liveGoogleVoiceProviders(t, ctx)
	application := newSeededTestApp(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}).WithRealtimeVoiceProviders(providers.SpeechToText, providers.ConversationModel, providers.TextToSpeech)
	fixture := seed(t, ctx, application)
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	defer server.Close()
	startedAt := time.Now()
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}}})
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close(websocket.StatusNormalClosure, "")
	connection.SetReadLimit(maxRealtimeVoiceFrameBytes)
	start := realtimeVoiceStartMessage("tenant-home", "inventory-home")
	start["developerDiagnostics"] = true
	writeRealtimeMessage(t, ctx, connection, start)
	started := readRealtimeMessage(t, ctx, connection)
	if started["type"] != "session.started" {
		t.Fatalf("start: %+v", started)
	}
	sessionID := started["sessionId"]
	writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "audio.chunk", "seq": 2, "sessionId": sessionID, "chunkId": "question-1", "audioBase64": base64.StdEncoding.EncodeToString(inputAudio), "isFinalChunk": true})
	writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "audio.end", "seq": 3, "sessionId": sessionID})
	var events []map[string]any
	var outputAudio []byte
	var response map[string]any
	completed := false
	for !completed {
		event := readRealtimeMessage(t, ctx, connection)
		if encoded, ok := event["audioBase64"].(string); ok {
			chunk, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatal(err)
			}
			outputAudio = append(outputAudio, chunk...)
			delete(event, "audioBase64")
			event["audioBytes"] = len(chunk)
		}
		event["elapsedMs"] = time.Since(startedAt).Milliseconds()
		events = append(events, event)
		safe, _ := json.Marshal(event)
		t.Logf("VOICE_TRACE %s", safe)
		if event["type"] == "assistant.response.completed" {
			response, _ = event["response"].(map[string]any)
		}
		if event["type"] == "session.failed" || event["type"] == "session.completed" || event["type"] == app.RealtimeVoiceEventActionPlanProposed {
			completed = true
		}
	}
	if directory := os.Getenv("STUFF_STASH_VOICE_EVIDENCE_DIR"); directory != "" {
		if err := os.MkdirAll(directory, 0700); err != nil {
			t.Fatal(err)
		}
		trace, _ := json.MarshalIndent(events, "", "  ")
		if err := os.WriteFile(filepath.Join(directory, "events.json"), trace, 0600); err != nil {
			t.Fatal(err)
		}
		if len(outputAudio) > 0 {
			if err := os.WriteFile(filepath.Join(directory, "answer.mp3"), outputAudio, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	if events[len(events)-1]["type"] != "session.completed" || response == nil || len(outputAudio) == 0 {
		t.Fatal("voice must complete with a grounded answer and actual synthesized audio")
	}
	if response["kind"] != "answer" {
		t.Fatalf("expected location answer, got %v", response["kind"])
	}
	spoken, _ := response["spokenResponse"].(string)
	for _, location := range fixture.spokenLocations {
		if !strings.Contains(strings.ToLower(spoken), strings.ToLower(location)) {
			t.Errorf("spoken answer omits recorded location %q: %s", location, spoken)
		}
	}
	seen := map[string]bool{}
	artifacts, _ := response["artifacts"].([]any)
	for _, raw := range artifacts {
		item, _ := raw.(map[string]any)
		id, _ := item["assetId"].(string)
		seen[id] = true
		if expected, ok := fixture.artifactLocations[id]; ok && item["context"] != expected {
			t.Errorf("asset %s omits recorded card location %q: %v", id, expected, item["context"])
		}
	}
	for _, id := range fixture.expectedIDs {
		if !seen[id] {
			t.Errorf("answer omits relevant asset %s", id)
		}
	}
	for _, id := range fixture.excludedIDs {
		if seen[id] {
			t.Errorf("answer includes irrelevant asset %s", id)
		}
	}
}
