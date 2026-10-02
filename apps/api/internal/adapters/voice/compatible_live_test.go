package voice

import (
	"context"
	"encoding/json"
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
	config.Profile.RuntimeOptionsJSON = []byte(`{"httpTimeout":"90s"}`)
	config.Credential = []byte("synthetic-local-marker")
	factory := ProviderProfileFactory{CompatibleEndpoints: []string{endpoint}}
	type stage struct {
		Name      string `json:"name"`
		Passed    bool   `json:"passed"`
		ElapsedMS int64  `json:"elapsedMs"`
	}
	stages := []stage{}
	run := func(name string, check func(*testing.T, context.Context)) {
		t.Run(name, func(t *testing.T) {
			start := time.Now()
			defer func() { stages = append(stages, stage{name, !t.Failed(), time.Since(start).Milliseconds()}) }()
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
	run("profile-diagnostic", func(t *testing.T, ctx context.Context) {
		result, err := NewProviderProfileTester(factory).TestProviderProfile(ctx, ports.ProviderProfileTestInput{Profile: config.Profile, CredentialPurpose: config.CredentialPurpose, Credential: config.Credential, TestedAt: time.Now()})
		if err != nil || result.Status != ports.ProviderProfileTestStatusSucceeded {
			t.Fatal("real model profile diagnostic failed")
		}
	})
	run("read-tool-result-answer", func(t *testing.T, ctx context.Context) {
		provider, err := factory.ConversationModelProvider(ctx, config)
		if err != nil {
			t.Fatal("provider construction failed")
		}
		input := ports.ConversationModelInput{
			Instructions: "Use find to look up the tent before answering. Pass query tent. After receiving the tool result, call answer with a short text stating the returned location. Never invent a location. /no_think",
			Messages:     []ports.ConversationMessage{{Role: ports.ConversationRoleUser, Text: "Where is my tent?"}},
			Tools: []ports.ConversationToolDefinition{
				{Name: "find", Description: "Look up an inventory item by name", Parameters: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"],"additionalProperties":false}`)},
				{Name: "answer", Description: "Answer the user using the lookup result", ResponseTool: true, Parameters: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}`)},
			},
		}
		turn, err := provider.Converse(ctx, input)
		if err != nil || len(turn.ToolCalls) != 1 || turn.ToolCalls[0].Name != "find" {
			t.Fatal("model did not request the required lookup")
		}
		query, ok := turn.ToolCalls[0].Arguments["query"].(string)
		if !ok || !strings.EqualFold(strings.TrimSpace(query), "tent") {
			t.Fatal("model did not preserve lookup input")
		}
		input.Messages = append(input.Messages,
			ports.ConversationMessage{Role: ports.ConversationRoleAssistant, Text: turn.Text, ToolCalls: turn.ToolCalls},
			ports.ConversationMessage{Role: ports.ConversationRoleTool, ToolResults: []ports.AgentToolResult{{CallID: turn.ToolCalls[0].ID, Name: "find", Content: `{"item":"tent","location":"Garage shelf"}`}}},
		)
		turn, err = provider.Converse(ctx, input)
		if err != nil || len(turn.ToolCalls) != 1 || turn.ToolCalls[0].Name != "answer" {
			t.Fatal("model did not return the answer tool")
		}
		answer, ok := turn.ToolCalls[0].Arguments["text"].(string)
		if !ok || !strings.Contains(strings.ToLower(answer), "garage shelf") {
			t.Fatal("model answer did not preserve the synthetic location")
		}
	})
}
