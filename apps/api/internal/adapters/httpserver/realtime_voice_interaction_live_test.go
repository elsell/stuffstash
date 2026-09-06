package httpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"nhooyr.io/websocket"
)

// Synthesized recordings exercise actual recognition and synthesis. These cases
// are isolated integration evidence, not physical-device or microphone evidence.
func TestGoogleLiveRealtimeInteractionCorpus(t *testing.T) {
	if os.Getenv("STUFF_STASH_GOOGLE_LIVE_TESTS") != "1" {
		t.Skip("explicit live Google opt-in required")
	}
	for _, scenario := range []struct {
		name  string
		audio []string
	}{
		{"move-existing", []string{"move-existing"}},
		{"create-additional", []string{"create-additional"}},
		{"dependent-move", []string{"dependent-move"}},
		{"ambiguous-move", []string{"ambiguous-move", "resolve-ambiguity"}},
		{"followup-color", []string{"find-drill", "followup-color"}},
	} {
		t.Run(scenario.name, func(t *testing.T) { runLiveInteraction(t, scenario.name, scenario.audio) })
	}
}

func runLiveInteraction(t *testing.T, name string, audioNames []string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	store := memory.NewStore()
	application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}, store, memory.NewAuthorizer())
	providers := liveGoogleVoiceProviders(t, ctx)
	application = application.WithRealtimeVoiceProviders(providers.SpeechToText, providers.ConversationModel, providers.TextToSpeech)
	fixture := seedLiveInteractionInventory(t, ctx, application)
	list := app.ListAssetsInput{Principal: identity.Principal{ID: "user-1"}, TenantID: "tenant-home", InventoryID: "inventory-home", Source: audit.SourceAPI, Limit: 100}
	before, err := application.ListAssets(ctx, list)
	if err != nil || before.HasMore {
		t.Fatalf("fixture snapshot: %v", err)
	}
	t.Cleanup(func() {
		checkCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		after, err := application.ListAssets(checkCtx, list)
		if err != nil || after.HasMore || !reflect.DeepEqual(before.Items, after.Items) {
			t.Errorf("unapproved interaction changed inventory: %v", err)
		}
	})
	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	defer server.Close()
	connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:user-1"}}})
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close(websocket.StatusNormalClosure, "")
	connection.SetReadLimit(maxRealtimeVoiceFrameBytes)
	start := realtimeVoiceStartMessage("tenant-home", "inventory-home")
	start["conversationContinuity"], start["developerDiagnostics"] = true, true
	writeRealtimeMessage(t, ctx, connection, start)
	started := readRealtimeMessage(t, ctx, connection)
	if started["type"] != "session.started" {
		t.Fatalf("session start: %+v", started)
	}
	sessionID, _ := started["sessionId"].(string)
	var turns []liveInteractionTurn
	for index, audioName := range audioNames {
		raw, err := os.ReadFile(filepath.Join(liveVoiceRequired(t, "STUFF_STASH_VOICE_INTERACTION_AUDIO_DIR"), audioName+".m4a"))
		if err != nil || len(raw) == 0 {
			t.Fatalf("read corpus recording %s: %v", audioName, err)
		}
		began := time.Now()
		seq := 2 + index*2
		writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "audio.chunk", "seq": seq, "sessionId": sessionID, "chunkId": audioName, "audioBase64": base64.StdEncoding.EncodeToString(raw), "isFinalChunk": true})
		writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "audio.end", "seq": seq + 1, "sessionId": sessionID})
		turn := readLiveInteractionTurn(t, ctx, connection, began, name, index)
		turns = append(turns, turn)
		if index < len(audioNames)-1 {
			if turn.terminal != "session.completed" || turn.followUp != true {
				t.Fatalf("interaction stopped before required follow-up: %+v", turn)
			}
			requireLiveInteractionSpeech(t, turn)
			if name == "ambiguous-move" {
				requireLiveSpokenFacts(t, turn, "Office", "Kitchen")
			}
			if name == "followup-color" {
				requireLiveSpokenFacts(t, turn, "Office")
			}
		}
	}
	last := turns[len(turns)-1]
	if name == "followup-color" {
		requireLiveInteractionSpeech(t, last)
		requireLiveSpokenFacts(t, last, "red")
		return
	}
	if last.terminal != "action.plan.proposed" || last.planID == "" {
		t.Fatalf("expected reviewable proposal: %+v", last)
	}
	plan, found, err := store.ActionPlanByID(ctx, "tenant-home", "inventory-home", last.planID)
	if err != nil || !found {
		t.Fatalf("read scoped proposal: %v", err)
	}
	checkLiveInteractionProposal(t, name, plan, fixture)
}

type liveInteractionTurn struct {
	terminal, spoken, planID string
	audioBytes               int
	followUp                 bool
}

func readLiveInteractionTurn(t *testing.T, ctx context.Context, connection *websocket.Conn, began time.Time, scenario string, index int) liveInteractionTurn {
	t.Helper()
	var result liveInteractionTurn
	var events []map[string]any
	var audio []byte
	for result.terminal == "" {
		event := readRealtimeMessage(t, ctx, connection)
		if encoded, ok := event["audioBase64"].(string); ok {
			raw, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatal(err)
			}
			audio = append(audio, raw...)
			delete(event, "audioBase64")
			event["audioBytes"] = len(raw)
		}
		event["elapsedMs"] = time.Since(began).Milliseconds()
		events = append(events, event)
		raw, _ := json.Marshal(event)
		t.Logf("VOICE_INTERACTION_TRACE scenario=%s turn=%d %s", scenario, index+1, raw)
		switch event["type"] {
		case "assistant.response.completed":
			response, _ := event["response"].(map[string]any)
			result.spoken, _ = response["spokenResponse"].(string)
		case "action.plan.proposed":
			plan, _ := event["actionPlan"].(map[string]any)
			result.planID, _ = plan["planId"].(string)
			result.terminal = "action.plan.proposed"
		case "session.completed", "session.failed":
			result.terminal, _ = event["type"].(string)
			result.followUp, _ = event["followUpAvailable"].(bool)
		}
	}
	result.audioBytes = len(audio)
	directory := filepath.Join(liveVoiceRequired(t, "STUFF_STASH_VOICE_EVIDENCE_DIR"), scenario)
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	label := "first"
	if index > 0 {
		label = "followup"
	}
	raw, _ := json.MarshalIndent(events, "", "  ")
	if err := os.WriteFile(filepath.Join(directory, label+"-events.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	if len(audio) > 0 {
		if err := os.WriteFile(filepath.Join(directory, label+"-answer.mp3"), audio, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return result
}
func requireLiveInteractionSpeech(t *testing.T, turn liveInteractionTurn) {
	t.Helper()
	if turn.terminal != "session.completed" || strings.TrimSpace(turn.spoken) == "" || turn.audioBytes == 0 {
		t.Fatalf("missing completed model-authored speech: %+v", turn)
	}
}
func requireLiveSpokenFacts(t *testing.T, turn liveInteractionTurn, facts ...string) {
	t.Helper()
	for _, fact := range facts {
		if !strings.Contains(strings.ToLower(turn.spoken), strings.ToLower(fact)) {
			t.Errorf("spoken answer omits %q: %s", fact, turn.spoken)
		}
	}
}
