package app

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/stuffstash/stuff-stash/internal/adapters/voice"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestGoogleGeminiLiveMoveWaterBottleToMissingKitchenProposesPlan(t *testing.T) {
	if os.Getenv("STUFF_STASH_GOOGLE_LIVE_TESTS") != "1" {
		t.Skip("set STUFF_STASH_GOOGLE_LIVE_TESTS=1 to run the live Gemini regression")
	}
	projectID := strings.TrimSpace(os.Getenv("STUFF_STASH_GOOGLE_CLOUD_PROJECT"))
	if projectID == "" {
		t.Skip("set STUFF_STASH_GOOGLE_CLOUD_PROJECT to run the live Gemini regression")
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
	resolver := successfulRealtimeVoiceResolver()
	resolver.providers.SpeechToText = resolvedSpeechToText{transcript: "Move my water bottle to the kitchen."}
	resolver.providers.LanguageInference = voice.NewGoogleGeminiLanguageInference(voice.GoogleGeminiConfig{
		ProjectID:    projectID,
		Location:     location,
		Model:        model,
		QuotaProject: projectID,
		TokenSource:  liveGoogleTokenSource(t, ctx),
	})
	resolver.providers.TextToSpeech = &resolvedTextToSpeech{}
	application, store := newRealtimeVoiceResolutionTestAppWithStoreSessionsAndIDs(t, resolver, newFakeRealtimeSessionRepository(), &fakeIDGenerator{})
	office := assetItem("office-1", "tenant-home", "inventory-home", asset.KindLocation, "")
	officeTitle, _ := asset.NewTitle("Office")
	office.Title = officeTitle
	if err := store.CreateAsset(ctx, office, audit.Record{ID: audit.ID("audit-office"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "office-1", OccurredAt: time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed office: %v", err)
	}
	waterBottle := assetItem("water-bottle-1", "tenant-home", "inventory-home", asset.KindItem, "office-1")
	waterBottleTitle, _ := asset.NewTitle("Water bottle")
	waterBottle.Title = waterBottleTitle
	if err := store.CreateAsset(ctx, waterBottle, audit.Record{ID: audit.ID("audit-water-bottle"), TenantID: audit.TenantID("tenant-home"), InventoryID: audit.InventoryID("inventory-home"), Action: audit.ActionAssetCreated, TargetType: audit.TargetAsset, TargetID: "water-bottle-1", OccurredAt: time.Date(2026, 6, 29, 12, 1, 0, 0, time.UTC)}, nil); err != nil {
		t.Fatalf("seed water bottle: %v", err)
	}

	sessionInput := defaultRealtimeVoiceSessionInput()
	sessionInput.DeveloperDiagnostics = true
	session, err := application.StartRealtimeVoiceSession(ctx, sessionInput)
	if err != nil {
		t.Fatalf("start realtime voice session: %v", err)
	}
	directSearch, err := application.SearchAssets(ctx, SearchAssetsInput{
		Principal:      session.Principal,
		TenantID:       session.TenantID,
		InventoryIDs:   []inventory.InventoryID{session.InventoryID},
		Query:          "water bottle",
		Mode:           "fuzzy",
		LifecycleState: "active",
		Limit:          10,
	})
	if err != nil || len(directSearch.Items) == 0 {
		t.Fatalf("direct search should find seeded water bottle, result=%+v err=%T %[2]v", directSearch, err)
	}
	events := []RealtimeVoiceEvent{}
	err = application.RunRealtimeVoiceQuery(ctx, RealtimeVoiceQueryInput{Session: session, AudioChunks: [][]byte{[]byte("audio")}}, func(event RealtimeVoiceEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("run realtime voice query: %v\n%s", err, liveGeminiVoiceDiagnostics(events))
	}
	var proposed *RealtimeVoiceActionPlanProposal
	for index := range events {
		event := events[index]
		if event.Type == RealtimeVoiceEventActionPlanProposed {
			proposed = event.ActionPlan
		}
		if event.Type == RealtimeVoiceEventAssistantResponseCompleted || event.Type == RealtimeVoiceEventSessionCompleted {
			t.Fatalf("expected live Gemini proposal to pause before final completion, got %+v\n%s", event, liveGeminiVoiceDiagnostics(events))
		}
	}
	if proposed == nil {
		t.Fatalf("expected live Gemini to propose Kitchen move plan, got events:\n%s", liveGeminiVoiceDiagnostics(events))
	}
	assertLiveGeminiVoiceKitchenMoveProposal(t, *proposed, events)
}

func liveGoogleTokenSource(t *testing.T, ctx context.Context) oauth2.TokenSource {
	t.Helper()

	if token := strings.TrimSpace(os.Getenv("STUFF_STASH_GOOGLE_ACCESS_TOKEN")); token != "" {
		return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token, TokenType: "Bearer"})
	}
	tokenSource, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		t.Skipf("Google ADC unavailable for live Gemini regression: %v", err)
	}
	return tokenSource
}

func assertLiveGeminiVoiceKitchenMoveProposal(t *testing.T, proposed RealtimeVoiceActionPlanProposal, events []RealtimeVoiceEvent) {
	t.Helper()

	if len(proposed.Commands) < 2 {
		t.Fatalf("expected create and move commands, got %+v\n%s", proposed, liveGeminiVoiceDiagnostics(events))
	}
	kitchenCommandID := ""
	sawKitchenCreate := false
	sawWaterBottleMove := false
	for _, command := range proposed.Commands {
		if command.Kind == string(actionplan.CommandKindCreateLocation) && strings.EqualFold(command.Title, "Kitchen") && command.AssetKind == asset.KindLocation.String() {
			kitchenCommandID = command.ID
			sawKitchenCreate = true
		}
	}
	for _, command := range proposed.Commands {
		if command.Kind == string(actionplan.CommandKindMoveAsset) && command.Operation == "move" {
			if command.ParentCommandID == kitchenCommandID && command.ParentAssetID == "" {
				sawWaterBottleMove = true
			}
		}
	}
	if kitchenCommandID == "" || !sawKitchenCreate || !sawWaterBottleMove {
		t.Fatalf("expected create Kitchen plus move Water bottle into that command, got %+v\n%s", proposed, liveGeminiVoiceDiagnostics(events))
	}
}

func liveGeminiVoiceDiagnostics(events []RealtimeVoiceEvent) string {
	builder := strings.Builder{}
	for _, event := range events {
		switch event.Type {
		case RealtimeVoiceEventAgentDiagnostic:
			builder.WriteString(event.Message)
			builder.WriteString(": ")
			builder.WriteString(event.Detail)
			builder.WriteString("\n")
		case RealtimeVoiceEventToolCallFailed:
			builder.WriteString("tool failed: ")
			builder.WriteString(event.Code)
			builder.WriteString(" ")
			builder.WriteString(event.Message)
			builder.WriteString("\n")
		case RealtimeVoiceEventActionPlanProposed:
			builder.WriteString("action plan proposed\n")
		}
	}
	return safeRealtimeVoiceDiagnosticText(builder.String(), 12000)
}

var _ ports.LanguageInferenceProvider = voice.GoogleGeminiLanguageInference{}
