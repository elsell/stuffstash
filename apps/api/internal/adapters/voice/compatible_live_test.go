package voice

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// Opt-in only: the isolated acceptance runner installs and verifies the pinned model.
func TestCompatibleLocalModelAcceptance(t *testing.T) {
	if os.Getenv("STUFF_STASH_LOCAL_MODEL_ACCEPTANCE") != "1" {
		t.Skip("real local model acceptance not enabled")
	}
	const endpoint = "http://127.0.0.1:11434/v1"
	config := compatibleTestConfig(t, endpoint)
	config.Profile.ModelName = agentmodel.ModelName("stuffstash-acceptance")
	config.Profile.RuntimeOptionsJSON = agentmodel.JSONObject(`{"httpTimeout":"90s"}`)
	config.Credential = []byte("synthetic-local-marker")
	factory := ProviderProfileFactory{CompatibleEndpoints: []string{endpoint}}
	type stage struct {
		Name      string                 `json:"name"`
		Passed    bool                   `json:"passed"`
		ElapsedMS int64                  `json:"elapsedMs"`
		Outcome   string                 `json:"outcome"`
		Response  *localResponseEvidence `json:"response,omitempty"`
	}
	stages := []stage{}
	outcome := "passed"
	var response *localResponseEvidence
	run := func(name string, check func(*testing.T, context.Context)) bool {
		outcome, response = "passed", nil
		return t.Run(name, func(t *testing.T) {
			start := time.Now()
			defer func() {
				stages = append(stages, stage{name, !t.Failed(), time.Since(start).Milliseconds(), outcome, response})
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			check(t, ctx)
		})
	}
	defer func() {
		path := os.Getenv("STUFF_STASH_LOCAL_MODEL_EVIDENCE")
		if path == "" {
			t.Error("evidence path is required")
			return
		}
		raw, err := json.MarshalIndent(struct {
			Stages []stage `json:"stages"`
		}{stages}, "", "  ")
		if err != nil {
			t.Error("could not encode bounded evidence")
			return
		}
		if err = os.WriteFile(path, raw, 0600); err != nil {
			t.Error("could not retain bounded evidence")
		}
	}()
	fail := func(t *testing.T, reason string) {
		outcome = reason
		t.Fatal(reason)
	}
	run("profile-diagnostic", func(t *testing.T, ctx context.Context) {
		result, err := NewProviderProfileTester(factory).TestProviderProfile(ctx, ports.ProviderProfileTestInput{Profile: config.Profile, CredentialPurpose: config.CredentialPurpose, Credential: config.Credential, TestedAt: time.Now()})
		if err != nil || result.Status != ports.ProviderProfileTestStatusSucceeded {
			fail(t, "profile-diagnostic-failed")
		}
	})
	var provider ports.ConversationModel
	var input ports.ConversationModelInput
	var turn ports.ConversationModelTurn
	if !run("lookup", func(t *testing.T, ctx context.Context) {
		var err error
		provider, err = factory.ConversationModelProvider(ctx, config)
		if err != nil {
			fail(t, "provider-construction-failed")
		}
		compatible := provider.(compatibleConversation)
		response = &localResponseEvidence{}
		compatible.client.Transport = localEvidenceTransport{response}
		provider = compatible
		input = ports.ConversationModelInput{
			Instructions: "Use find to look up the tent before answering. Pass query tent. After receiving the tool result, call answer with a short text stating the returned location. Never invent a location. /no_think",
			Messages:     []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Where is my tent?"}},
			Tools: []ports.ConversationToolDefinition{
				{Name: "find", Description: "Look up an inventory item by name", Parameters: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"],"additionalProperties":false}`)},
				{Name: "answer", Description: "Answer the user using the lookup result", ResponseTool: true, Parameters: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`)},
			},
		}
		turn, err = provider.Converse(ctx, input)
		if err != nil {
			fail(t, localAcceptanceError(err))
		}
		if len(turn.ToolCalls) != 1 {
			fail(t, "wrong-tool-count")
		}
		if turn.ToolCalls[0].Name != "find" {
			fail(t, "wrong-tool-name")
		}
		query, ok := turn.ToolCalls[0].Arguments["query"].(string)
		if !ok || !strings.EqualFold(strings.TrimSpace(query), "tent") {
			fail(t, "invalid-lookup-query")
		}
	}) {
		stages = append(stages, stage{Name: "tool-result-answer", Outcome: "not-run-lookup-failed"})
		return
	}
	run("tool-result-answer", func(t *testing.T, ctx context.Context) {
		compatible := provider.(compatibleConversation)
		response = &localResponseEvidence{}
		compatible.client.Transport = localEvidenceTransport{response}
		provider = compatible
		input.Messages = append(input.Messages,
			ports.ConversationMessage{Role: ports.ConversationRoleAssistant, Text: turn.Text, ToolCalls: turn.ToolCalls},
			ports.ConversationMessage{Role: ports.ConversationRoleTool, ToolResults: []ports.AgentToolResult{{CallID: turn.ToolCalls[0].ID, Name: "find", Content: `{"item":"tent","location":"Garage shelf"}`}}},
		)
		turn, err := provider.Converse(ctx, input)
		if err != nil {
			fail(t, localAcceptanceError(err))
		}
		if len(turn.ToolCalls) != 1 {
			fail(t, "wrong-tool-count")
		}
		if turn.ToolCalls[0].Name != "answer" {
			fail(t, "wrong-tool-name")
		}
		answer, ok := turn.ToolCalls[0].Arguments["text"].(string)
		if !ok || !strings.Contains(strings.ToLower(answer), "garage shelf") {
			fail(t, "invalid-answer-location")
		}
	})
}

// Only structural, allowlisted response facts leave this synthetic live test.
type localResponseEvidence struct {
	Status    int    `json:"httpStatus"`
	Finish    string `json:"finishReason"`
	ToolCalls int    `json:"toolCallCount"`
}
type localEvidenceTransport struct{ evidence *localResponseEvidence }

func (t localEvidenceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	t.evidence.Status = response.StatusCode
	response.Body = &localEvidenceBody{ReadCloser: response.Body, evidence: t.evidence}
	return response, nil
}

type localEvidenceBody struct {
	io.ReadCloser
	evidence *localResponseEvidence
	raw      []byte
}

func (b *localEvidenceBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	remaining := compatibleMaxBodyBytes + 1 - len(b.raw)
	if remaining > n {
		remaining = n
	}
	if remaining > 0 {
		b.raw = append(b.raw, p[:remaining]...)
	}
	return n, err
}
func (b *localEvidenceBody) Close() error {
	err := b.ReadCloser.Close()
	b.capture()
	return err
}
func (b *localEvidenceBody) capture() {
	var body struct {
		Choices []struct {
			Finish  string `json:"finish_reason"`
			Message struct {
				Calls []json.RawMessage `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	b.evidence.Finish = "invalid-or-unknown"
	if json.Unmarshal(b.raw, &body) == nil && len(body.Choices) == 1 {
		choice := body.Choices[0]
		switch choice.Finish {
		case "stop", "tool_calls", "length", "content_filter":
			b.evidence.Finish = choice.Finish
		}
		b.evidence.ToolCalls = len(choice.Message.Calls)
	}
}
func localAcceptanceError(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline-exceeded"
	}
	if errors.Is(err, ports.ErrInvalidProviderInput) {
		return "invalid-provider-response"
	}
	return "provider-request-failed"
}
