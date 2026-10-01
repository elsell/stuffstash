package voice

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func compatibleTestConfig(t *testing.T, endpoint string) ProviderProfileProviderConfig {
	t.Helper()
	profile := googleFactoryProfile(t, agentmodel.ProviderCapabilityLanguageInference, `{}`)
	profile.ProviderKind = agentmodel.ProviderKindLocalHTTP
	profile.EndpointURL = agentmodel.EndpointURL(endpoint)
	profile.ModelName = "local-test"
	return ProviderProfileProviderConfig{Profile: profile, CredentialPurpose: ports.ProviderCredentialPurposeAPIKey, Credential: []byte("test-secret"), CredentialVersionID: "v1"}
}
func compatibleInput() ports.ConversationModelInput {
	return ports.ConversationModelInput{Instructions: "Use authorized tools.", Messages: []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Find tent"}}, Tools: []ports.ConversationToolDefinition{{Name: "find", Parameters: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}}}`)}, {Name: "answer", ResponseTool: true, Parameters: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}}}`)}}}
}
func TestCompatibleConversationPreservesToolRoundTrip(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer test-secret" {
			t.Errorf("wrong endpoint or credential delivery")
		}
		var request struct {
			Model      string           `json:"model"`
			Messages   []map[string]any `json:"messages"`
			ToolChoice string           `json:"tool_choice"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if request.Model != "local-test" || request.ToolChoice != "required" || request.Messages[0]["role"] != "system" {
			t.Errorf("invalid wire request: %+v", request)
		}
		response := `{"choices":[{"finish_reason":"tool_calls","message":{"role":"assistant","tool_calls":[{"id":"read-1","type":"function","function":{"name":"find","arguments":"{\"query\":\"tent\"}"}}]}}]}`
		if calls == 2 {
			if len(request.Messages) != 4 || request.Messages[3]["tool_call_id"] != "read-1" || request.Messages[3]["content"] != "Found tent" {
				t.Errorf("lost tool round trip: %+v", request.Messages)
			}
			response = `{"choices":[{"finish_reason":"tool_calls","message":{"role":"assistant","tool_calls":[{"id":"answer-1","type":"function","function":{"name":"answer","arguments":"{\"text\":\"Tent found\"}"}}]}}]}`
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()
	factory := ProviderProfileFactory{CompatibleEndpoints: []string{server.URL + "/v1"}}
	provider, err := factory.ConversationModelProvider(context.Background(), compatibleTestConfig(t, server.URL+"/v1"))
	if err != nil {
		t.Fatal(err)
	}
	input := compatibleInput()
	turn, err := provider.Converse(context.Background(), input)
	if err != nil || len(turn.ToolCalls) != 1 {
		t.Fatalf("first turn: %+v %v", turn, err)
	}
	input.Messages = append(input.Messages, ports.ConversationMessage{Role: ports.ConversationRoleAssistant, ToolCalls: turn.ToolCalls, ProviderState: turn.ProviderState}, ports.ConversationMessage{Role: ports.ConversationRoleTool, ToolResults: []ports.AgentToolResult{{CallID: "read-1", Name: "find", Content: "Found tent"}}})
	turn, err = provider.Converse(context.Background(), input)
	if err != nil || len(turn.ToolCalls) != 1 || turn.ToolCalls[0].Name != "answer" || calls != 2 {
		t.Fatalf("second turn: %+v %v calls=%d", turn, err, calls)
	}
}
func TestCompatibleConversationRejectsUnsafeProviderOutput(t *testing.T) {
	for name, body := range map[string]string{
		"truncated":             `{"choices":[{"finish_reason":"length","message":{"content":"partial"}}]}`,
		"refusal":               `{"choices":[{"finish_reason":"stop","message":{"refusal":"no","content":"no"}}]}`,
		"missing required tool": `{"choices":[{"finish_reason":"stop","message":{"content":"done"}}]}`,
		"unknown tool":          `{"choices":[{"finish_reason":"tool_calls","message":{"tool_calls":[{"id":"1","type":"function","function":{"name":"delete_everything","arguments":"{}"}}]}}]}`,
		"non object arguments":  `{"choices":[{"finish_reason":"tool_calls","message":{"tool_calls":[{"id":"1","type":"function","function":{"name":"find","arguments":"[]"}}]}}]}`,
		"oversized":             strings.Repeat("x", 1024*1024+1),
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
			defer server.Close()
			provider, err := (ProviderProfileFactory{CompatibleEndpoints: []string{server.URL}}).ConversationModelProvider(context.Background(), compatibleTestConfig(t, server.URL))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = provider.Converse(context.Background(), compatibleInput()); err == nil {
				t.Fatal("unsafe output accepted")
			}
		})
	}
}
func TestCompatibleProviderEgressAndCancellation(t *testing.T) {
	reached := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true }))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer source.Close()
	config := compatibleTestConfig(t, source.URL)
	if _, err := (ProviderProfileFactory{}).ConversationModelProvider(context.Background(), config); err == nil {
		t.Fatal("unlisted endpoint accepted")
	}
	factory := ProviderProfileFactory{CompatibleEndpoints: []string{source.URL, target.URL}}
	provider, err := factory.ConversationModelProvider(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = provider.Converse(context.Background(), compatibleInput()); err == nil || reached {
		t.Fatal("redirect followed or accepted")
	}
	config.Profile.ProviderKind = agentmodel.ProviderKindOpenAICompatible
	if _, err = factory.ConversationModelProvider(context.Background(), config); err == nil {
		t.Fatal("remote HTTP accepted")
	}
	release := make(chan struct{})
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer func() { close(release); slow.Close() }()
	provider, err = (ProviderProfileFactory{CompatibleEndpoints: []string{slow.URL}}).ConversationModelProvider(context.Background(), compatibleTestConfig(t, slow.URL))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err = provider.Converse(ctx, compatibleInput())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline not propagated: %v", err)
	}
}

func TestCompatibleProfileDiagnosticAndIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"finish_reason":"tool_calls","message":{"tool_calls":[{"id":"probe","type":"function","function":{"name":"ready","arguments":"{\"status\":\"ready\"}"}}]}}]}`))
	}))
	defer server.Close()
	factory := ProviderProfileFactory{CompatibleEndpoints: []string{server.URL}}
	config := compatibleTestConfig(t, server.URL)
	result, err := NewProviderProfileTester(factory).TestProviderProfile(context.Background(), ports.ProviderProfileTestInput{Profile: config.Profile, CredentialPurpose: config.CredentialPurpose, Credential: config.Credential, TestedAt: time.Now()})
	if err != nil || result.Status != ports.ProviderProfileTestStatusSucceeded {
		t.Fatalf("profile diagnostic: %+v %v", result, err)
	}
	first, err := factory.ProviderConfigurationIdentity(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	config.CredentialVersionID = "v2"
	second, err := factory.ProviderConfigurationIdentity(context.Background(), config)
	if err != nil || first == second {
		t.Fatalf("credential rotation identity did not change: %v", err)
	}
	config.Profile.Capability = agentmodel.ProviderCapabilitySpeechToText
	if _, err = factory.SpeechToTextProvider(context.Background(), config); err == nil {
		t.Fatal("unsupported speech capability accepted")
	}
}
