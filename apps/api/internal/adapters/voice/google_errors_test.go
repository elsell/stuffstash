package voice

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2/google"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiLanguageInferenceReportsSafeHTTPStatusWithoutBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"quota exhausted for secret-project and bearer should-not-leak"}}`))
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

	_, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{Transcript: "Provider diagnostic.", FinalOnly: true})
	if err == nil {
		t.Fatalf("expected provider error")
	}
	if !strings.Contains(err.Error(), "status 429") {
		t.Fatalf("expected safe status in error, got %v", err)
	}
	if strings.Contains(err.Error(), "secret-project") || strings.Contains(err.Error(), "should-not-leak") {
		t.Fatalf("provider error leaked response body: %v", err)
	}
}

func TestGoogleGeminiLanguageInferenceReportsSafeTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"final\":{\"kind\":\"answer\",\"spokenResponse\":\"ready\",\"displayResponse\":\"ready\"}}"}]}}]}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID:   "project",
		Location:    "us-central1",
		Model:       "gemini-test",
		BaseURL:     server.URL,
		TokenSource: staticTokenSource{},
		HTTPTimeout: time.Millisecond,
	})

	_, err := provider.NextTurn(context.Background(), ports.LanguageInferenceInput{Transcript: "Provider diagnostic.", FinalOnly: true})
	if err == nil {
		t.Fatalf("expected provider timeout")
	}
	var safe interface{ SafeRealtimeVoiceDiagnostic() string }
	if !errors.As(err, &safe) || safe.SafeRealtimeVoiceDiagnostic() != "provider_timeout" {
		t.Fatalf("expected safe timeout diagnostic, got %T %v", err, err)
	}
}

func TestGoogleGeminiLiveMinimalStructuredAndToolTurns(t *testing.T) {
	if os.Getenv("STUFF_STASH_GOOGLE_LIVE_TESTS") != "1" {
		t.Skip("set STUFF_STASH_GOOGLE_LIVE_TESTS=1 to run the live Gemini probe")
	}
	projectID := strings.TrimSpace(os.Getenv("STUFF_STASH_GOOGLE_CLOUD_PROJECT"))
	if projectID == "" {
		t.Skip("set STUFF_STASH_GOOGLE_CLOUD_PROJECT to run the live Gemini probe")
	}
	location := strings.TrimSpace(os.Getenv("STUFF_STASH_GOOGLE_CLOUD_LOCATION"))
	if location == "" {
		location = "us-central1"
	}
	model := strings.TrimSpace(os.Getenv("STUFF_STASH_GOOGLE_GEMINI_MODEL"))
	if model == "" {
		model = "gemini-2.5-flash-lite"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	tokenSource, err := google.DefaultTokenSource(ctx, googleCloudPlatformScope)
	if err != nil {
		t.Skipf("Google ADC unavailable for live Gemini probe: %v", err)
	}
	provider := NewGoogleGeminiLanguageInference(GoogleGeminiConfig{
		ProjectID:    projectID,
		Location:     location,
		Model:        model,
		QuotaProject: projectID,
		TokenSource:  tokenSource,
		HTTPTimeout:  120 * time.Second,
	})

	finalTurn, err := provider.NextTurn(ctx, ports.LanguageInferenceInput{
		Transcript: "Say the provider is ready.",
		FinalOnly:  true,
	})
	if err != nil {
		t.Fatalf("structured final live turn failed: %v", err)
	}
	if finalTurn.Final == nil || strings.TrimSpace(finalTurn.Final.SpokenResponse) == "" {
		t.Fatalf("expected structured final response, got %+v", finalTurn)
	}
	t.Logf("live structured final: kind=%s spoken=%q", finalTurn.Final.Kind, finalTurn.Final.SpokenResponse)

	toolTurn, err := provider.NextTurn(ctx, ports.LanguageInferenceInput{
		Transcript:      "Where is my water bottle?",
		RequireToolCall: true,
		Tools: []ports.AgentToolDescriptor{{
			Name:        "search_authorized_assets",
			Description: "Search visible assets.",
			ReadOnly:    true,
			Parameters: ports.AgentToolParameters{
				Required: []string{"query"},
				Properties: map[string]ports.AgentToolParameter{
					"query": {Type: ports.AgentToolParameterTypeString, Description: "Search keywords."},
				},
			},
		}},
	})
	if err != nil {
		t.Fatalf("forced tool live turn failed: %v", err)
	}
	if len(toolTurn.ToolCalls) != 1 || toolTurn.ToolCalls[0].Name != "search_authorized_assets" {
		t.Fatalf("expected forced search tool call, got %+v", toolTurn)
	}
	t.Logf("live forced tool call: name=%s args=%v", toolTurn.ToolCalls[0].Name, toolTurn.ToolCalls[0].Arguments)

	plannerTurn, err := provider.NextTurn(ctx, ports.LanguageInferenceInput{
		Transcript: "Move my water bottle to the kitchen.",
		Tools:      []ports.AgentToolDescriptor{testGeminiActionPlanToolDescriptor()},
		ToolResults: []ports.AgentToolResult{{
			CallID: "call-water-bottle",
			Name:   "search_authorized_assets",
			Call: ports.AgentToolCall{
				ID:        "call-water-bottle",
				Name:      "search_authorized_assets",
				Arguments: map[string]any{"query": "water bottle"},
			},
			Content: `{"tool":"search_authorized_assets","count":1,"items":[{"assetId":"asset-water-bottle","title":"Water bottle","kind":"item","locationTitle":"Office"}]}`,
		}, {
			CallID: "call-kitchen",
			Name:   "search_authorized_assets",
			Call: ports.AgentToolCall{
				ID:        "call-kitchen",
				Name:      "search_authorized_assets",
				Arguments: map[string]any{"query": "kitchen"},
			},
			Content: `{"tool":"search_authorized_assets","count":0,"items":[]}`,
		}},
		PreviousTurns: 2,
		PlanOnly:      true,
	})
	if err != nil {
		t.Fatalf("structured planner live turn failed: %v", err)
	}
	if len(plannerTurn.ToolCalls) != 1 || plannerTurn.ToolCalls[0].Name != "propose_action_plan" {
		t.Fatalf("expected structured planner action plan, got %+v", plannerTurn)
	}
	commands, ok := plannerTurn.ToolCalls[0].Arguments["commands"].([]any)
	if !ok || len(commands) == 0 {
		t.Fatalf("expected planner commands from structured schema, got %+v", plannerTurn.ToolCalls[0].Arguments)
	}
	t.Logf("live structured planner action plan: %+v", plannerTurn.ToolCalls[0].Arguments)
	if !livePlannerCreatesKitchenAndMovesWithParentCommand(commands) {
		t.Fatalf("expected planner to create missing Kitchen and move by parentCommandId, got %+v", commands)
	}
}

func livePlannerCreatesKitchenAndMovesWithParentCommand(commands []any) bool {
	kitchenCommandID := ""
	for _, raw := range commands {
		command, ok := raw.(map[string]any)
		if !ok || command["kind"] != "create_location" {
			continue
		}
		args, ok := command["arguments"].(map[string]any)
		if !ok || !strings.EqualFold(stringAny(args["title"]), "Kitchen") {
			continue
		}
		kitchenCommandID = stringAny(command["id"])
		break
	}
	if kitchenCommandID == "" {
		return false
	}
	for _, raw := range commands {
		command, ok := raw.(map[string]any)
		if !ok || command["kind"] != "move_asset" {
			continue
		}
		args, ok := command["arguments"].(map[string]any)
		if !ok {
			continue
		}
		if stringAny(args["assetId"]) == "asset-water-bottle" && stringAny(args["parentCommandId"]) == kitchenCommandID {
			return true
		}
	}
	return false
}

func stringAny(value any) string {
	text, _ := value.(string)
	return text
}
