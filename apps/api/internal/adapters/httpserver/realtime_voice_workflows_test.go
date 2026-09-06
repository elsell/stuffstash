package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/auth"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"nhooyr.io/websocket"
)

type voiceWorkflowReadStore struct {
	*memory.Store
	reads atomic.Int32
}

func (s *voiceWorkflowReadStore) SelectedWorkflowRevision(ctx context.Context, id tenant.ID) (ports.WorkflowSelectionReference, bool, error) {
	s.reads.Add(1)
	return s.Store.SelectedWorkflowRevision(ctx, id)
}

func TestRealtimeWorkflowSelectionRemainsBehindWebSocketAuthorization(t *testing.T) {
	for _, test := range []struct {
		name, token, tenant, inventory string
		allowed                        bool
	}{
		{"owner", "dev:user-1", "tenant-home", "inventory-home", true},
		{"explicit model owner", "dev:user-1", "tenant-home", "inventory-home", true},
		{"exhausted workflow owner", "dev:user-1", "tenant-home", "inventory-home", true},
		{"outsider", "dev:user-2", "tenant-home", "inventory-home", false},
		{"wrong inventory", "dev:user-1", "tenant-home", "inventory-other", false},
		{"cross tenant", "dev:user-1", "tenant-other", "inventory-other", false},
		{"unauthenticated", "", "tenant-home", "inventory-home", false},
		{"malformed token", "malformed", "tenant-home", "inventory-home", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			store := &voiceWorkflowReadStore{Store: memory.NewStore()}
			authorizer := memory.NewAuthorizer()
			seedMemoryStore(t, ctx, store.Store, authorizer, seededState{tenants: []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}, {id: "tenant-other", name: "Other", owner: "user-2"}}, inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home", owner: "user-1"}, {id: "inventory-other", tenantID: "tenant-other", name: "Other", owner: "user-2"}}})
			limits := agentmodel.WorkflowLimits{Budget: agentmodel.WorkflowBudget{EvidenceRounds: 2, ModelCalls: 4, ElapsedSeconds: 60, FollowUpTurns: 4}, MaxStepAttempts: 2, MaxNameRunes: 100, MaxInstructionRunes: 1000}
			application := app.New(app.Dependencies{Auth: auth.NewLocalDevAuthenticator(), Authorizer: authorizer, Users: store, Tenants: store, Inventories: store, Assets: store, Search: store, Audit: store, RealtimeSessions: store, ProviderProfiles: store, ConversationWorkflows: store, ConversationWorkflowLimits: limits}).WithRealtimeVoiceProviderResolver(nativeBoundaryResolver{model: &nativeBoundaryConversation{}})
			revision, err := application.SaveConversationWorkflowRevision(ctx, app.SaveConversationWorkflowInput{Principal: principal("user-1"), TenantID: "tenant-home", Source: audit.SourceAPI, Definition: agentmodel.WorkflowDefinitionInput{Name: "Home", Retrieval: agentmodel.WorkflowRetrievalPreciseFirst, Response: agentmodel.WorkflowResponseGrounded, Budget: limits.Budget, Steps: []agentmodel.WorkflowStep{{Kind: agentmodel.WorkflowStepInterpret, Attempts: 1}, {Kind: agentmodel.WorkflowStepAssess, Attempts: 1}, {Kind: agentmodel.WorkflowStepRespond, Attempts: 1}}}})
			if err != nil {
				t.Fatal(err)
			}

			if test.name == "explicit model owner" || test.name == "exhausted workflow owner" {
				snapshot := revision.Snapshot()
				settings := snapshot.Definition.Settings()
				if test.name == "exhausted workflow owner" {
					settings.Budget.ModelCalls = 2
				}
				for index := range settings.Steps {
					settings.Steps[index].ProviderProfileID = "explicit-model"
				}
				snapshot.Definition, err = agentmodel.NewWorkflowDefinition(settings, limits)
				if err != nil {
					t.Fatal(err)
				}
				snapshot.ID = "explicit-revision"
				snapshot.Number = 2
				revision, err = agentmodel.NewWorkflowRevision(snapshot)
				if err != nil {
					t.Fatal(err)
				}
				created, ok := audit.NewRecord("explicit-draft", "tenant-home", "", "user-1", audit.ActionConversationWorkflowRevisionCreated, audit.SourceAPI, "conversation_workflow", string(snapshot.WorkflowID), time.Now(), "", nil)
				if !ok {
					t.Fatal("invalid draft audit")
				}
				if err = store.AppendWorkflowRevision(ctx, revision, 1, created); err != nil {
					t.Fatal(err)
				}
				application = app.New(app.Dependencies{Auth: auth.NewLocalDevAuthenticator(), Authorizer: authorizer, Users: store, Tenants: store, Inventories: store, Assets: store, Search: store, Audit: store, RealtimeSessions: store, ProviderProfiles: store, ConversationWorkflows: store, ConversationWorkflowLimits: limits, RealtimeVoiceProviderResolver: workflowOnlyVoiceResolver{}})
			}
			snapshot := revision.Snapshot()
			record, ok := audit.NewRecord("activation", "tenant-home", "", "user-1", audit.ActionConversationWorkflowActivated, audit.SourceAPI, "conversation_workflow", string(snapshot.WorkflowID), time.Now(), "", nil)
			if !ok {
				t.Fatal("invalid audit")
			}
			if err = store.ActivateWorkflowRevision(ctx, "tenant-home", snapshot.WorkflowID, snapshot.ID, ports.WorkflowSelectionReference{}, time.Now(), record); err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
			defer server.Close()
			headers := http.Header{}
			if test.token != "" {
				headers.Set("Authorization", "Bearer "+test.token)
			}
			connection, response, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/realtime/voice", &websocket.DialOptions{HTTPHeader: headers})
			if test.token == "" || test.token == "malformed" {
				if err == nil || response == nil || response.StatusCode != http.StatusUnauthorized {
					t.Fatalf("expected unauthorized handshake: %v %v", response, err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				defer connection.Close(websocket.StatusNormalClosure, "")
				start := realtimeVoiceStartMessage(test.tenant, test.inventory)
				start["conversationContinuity"] = true
				writeRealtimeMessage(t, ctx, connection, start)
				event := readRealtimeMessage(t, ctx, connection)
				if test.allowed {
					if event["type"] != "session.started" {
						t.Fatalf("authorized workflow session failed: %+v", event)
					}
					if test.name == "exhausted workflow owner" {
						writeRealtimeAudioTurn(t, ctx, connection, event["sessionId"].(string), 2, "budget-turn")
						events := readRealtimeMessagesUntil(t, ctx, connection, "session.completed")
						if findRealtimeEvent(t, events, "session.completed")["followUpAvailable"] != false {
							t.Fatal("exhausted workflow advertised another recording")
						}
						_, _, closedErr := connection.Read(ctx)
						if websocket.CloseStatus(closedErr) != websocket.StatusNormalClosure {
							t.Fatalf("exhausted session not closed normally: %v", closedErr)
						}
					}

				} else {
					if event["type"] != "session.failed" || event["code"] != "forbidden" {
						t.Fatalf("denied session: %+v", event)
					}
				}
			}
			expected := int32(0)
			if test.allowed {
				expected = 1
			}
			if store.reads.Load() != expected {
				t.Fatalf("workflow selection reads=%d expected=%d", store.reads.Load(), expected)
			}
		})
	}
}

// This controlled resolver has working speech and an explicit model, but no default model.
type workflowOnlyVoiceResolver struct{}
type workflowHTTPModel struct{}

func (workflowOnlyVoiceResolver) ResolveRealtimeVoiceProviders(_ context.Context, input ports.RealtimeVoiceProviderResolutionInput) (ports.RealtimeVoiceProviderSet, error) {
	if !input.SkipDefaultLanguage {
		return ports.RealtimeVoiceProviderSet{}, ports.ErrInvalidProviderInput
	}
	return ports.RealtimeVoiceProviderSet{SpeechToText: fakeSpeechToText{transcript: "Where are my tools?"}, TextToSpeech: fakeTextToSpeech{chunks: [][]byte{[]byte("audio")}}}, nil
}
func (workflowOnlyVoiceResolver) ResolveWorkflowLanguageProvider(_ context.Context, input ports.WorkflowLanguageProviderResolutionInput) (ports.WorkflowLanguageProviderBinding, error) {
	if input.TenantID != "tenant-home" || input.ProfileID != "explicit-model" {
		return ports.WorkflowLanguageProviderBinding{}, ports.ErrInvalidProviderInput
	}
	return ports.WorkflowLanguageProviderBinding{ProfileID: input.ProfileID, Provider: workflowHTTPModel{}}, nil
}

func (workflowHTTPModel) Converse(_ context.Context, input ports.ConversationModelInput) (ports.ConversationModelTurn, error) {
	if len(input.Messages) == 1 {
		return ports.ConversationModelTurn{ToolCalls: []ports.AgentToolCall{{ID: "find-tools", Name: app.RealtimeVoiceToolSearchAuthorizedAssets, Arguments: map[string]any{"query": "tools"}}}}, nil
	}
	return ports.ConversationModelTurn{Text: "I could not find matching tools."}, nil
}
