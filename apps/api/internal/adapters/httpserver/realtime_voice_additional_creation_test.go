package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"nhooyr.io/websocket"
)

type additionalItemModel struct {
	mode  agentmodel.CreationMode
	quote string
}

func (m additionalItemModel) NextTurn(_ context.Context, input ports.LanguageInferenceInput) (ports.LanguageInferenceTurn, error) {
	intent := agentmodel.Intent{RequestShape: agentmodel.RequestShapeSingleTarget, Kind: agentmodel.IntentKindChange, Operation: agentmodel.OperationCreate, SubjectMention: "Charger", NewAssetKind: "item", CreationMode: m.mode, CreationEvidence: m.quote}
	step := agentmodel.InvestigationStep{Intent: intent}
	if input.Investigation.Phase == agentmodel.InvestigationPhaseInitial {
		step.Decision = agentmodel.InvestigationDecisionSearch
		step.SearchRequests = []agentmodel.SearchRequest{{ReferenceKey: agentmodel.SemanticReferenceSubject, ReadKind: agentmodel.InvestigationReadSearchAssets, Mention: "Charger", SearchProbes: []string{"Charger"}}}
	} else {
		step.Decision = agentmodel.InvestigationDecisionFinish
		ids := []string{}
		for _, observation := range input.Investigation.Observations {
			ids = append(ids, observation.CandidateID)
		}
		status := agentmodel.ResolutionStrong
		if len(ids) > 1 {
			status = agentmodel.ResolutionAmbiguous
		}
		step.Resolutions = []agentmodel.Resolution{{ReferenceKey: agentmodel.SemanticReferenceSubject, Status: status, CandidateIDs: ids}}
	}
	return ports.LanguageInferenceTurn{Investigation: &step}, nil
}

func TestRealtimeAdditionalItemRequiresUserEvidenceAccessAndApproval(t *testing.T) {
	for _, scenario := range []struct {
		name, user, transcript string
		mode                   agentmodel.CreationMode
		quote, terminal        string
		approve                bool
	}{
		{"additional owner", "user-1", "I bought another charger", agentmodel.CreationModeAdditional, "I bought another charger", "action.plan.proposed", true},
		{"ordinary recording", "user-1", "Record my charger", agentmodel.CreationModeRecord, "", "session.completed", false},
		{"fabricated acquisition", "user-1", "Record my charger", agentmodel.CreationModeAdditional, "I bought another charger", "session.failed", false},
		{"viewer additional", "viewer", "I bought another charger", agentmodel.CreationModeAdditional, "I bought another charger", "session.failed", false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			store := memory.NewStore()
			authorizer := memory.NewAuthorizer()
			application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}}}, store, authorizer).WithRealtimeVoiceProviders(fakeSpeechToText{transcript: scenario.transcript}, additionalItemModel{mode: scenario.mode, quote: scenario.quote}, fakeTextToSpeech{chunks: [][]byte{[]byte("audio")}}).WithRealtimeVoiceResponseGenerator(httpTestVoiceResponseGenerator{})
			seedVoiceAsset(t, application, "user-1", "tenant-home", "inventory-home", "item", "Charger", "")
			if err := authorizer.GrantInventoryViewer(context.Background(), identity.Principal{ID: "viewer"}, "tenant-home", "inventory-home"); err != nil {
				t.Fatal(err)
			}
			original, err := store.ListAssetsByInventory(context.Background(), "tenant-home", "inventory-home", ports.AssetListPageRequest{Limit: 10})
			if err != nil || len(original) != 1 {
				t.Fatalf("fixture: %v %v", original, err)
			}
			server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			connection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": []string{"Bearer dev:" + scenario.user}}})
			if err != nil {
				t.Fatal(err)
			}
			defer connection.Close(websocket.StatusNormalClosure, "")
			writeRealtimeMessage(t, ctx, connection, realtimeVoiceStartMessage("tenant-home", "inventory-home"))
			started := readRealtimeMessage(t, ctx, connection)
			if started["type"] != "session.started" {
				t.Fatalf("start: %+v", started)
			}
			sessionID := started["sessionId"].(string)
			writeRealtimeAudioTurn(t, ctx, connection, sessionID, 2, "creation-audio")
			events := readRealtimeMessagesUntil(t, ctx, connection, scenario.terminal)
			before, err := store.ListAssetsByInventory(ctx, "tenant-home", "inventory-home", ports.AssetListPageRequest{Limit: 10})
			if err != nil || len(before) != 1 || before[0].ID != original[0].ID {
				t.Fatal("voice request mutated assets before approval")
			}
			if !scenario.approve {
				assertNoRealtimeEventType(t, events, "action.plan.proposed")
				return
			}
			proposal := findRealtimeEvent(t, events, "action.plan.proposed")["actionPlan"].(map[string]any)
			if !strings.Contains(proposal["confirmationSummary"].(string), "additional") {
				t.Fatalf("approval hides additional intent: %+v", proposal)
			}
			writeRealtimeMessage(t, ctx, connection, map[string]any{"type": "action.plan.approve", "seq": 4, "sessionId": sessionID, "planId": proposal["planId"]})
			readRealtimeMessagesUntil(t, ctx, connection, "action.plan.executed")
			after, err := store.ListAssetsByInventory(ctx, "tenant-home", "inventory-home", ports.AssetListPageRequest{Limit: 10})
			if err != nil || len(after) != 2 {
				t.Fatalf("approved new instance not created: %v %v", after, err)
			}
			if after[0].ID == after[1].ID || after[0].Title.String() != "Charger" || after[1].Title.String() != "Charger" {
				t.Fatal("additional item overwrote the existing identity")
			}
		})
	}
}
