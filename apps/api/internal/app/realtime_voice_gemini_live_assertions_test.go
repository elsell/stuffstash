package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func assertLiveGeminiVoiceCorpusOutcome(t *testing.T, scenario liveGeminiVoiceCorpusCase, events []RealtimeVoiceEvent, spoken string) {
	t.Helper()

	var proposed *RealtimeVoiceActionPlanProposal
	var completed *ports.StructuredAgentResponse
	for index := range events {
		event := events[index]
		switch event.Type {
		case RealtimeVoiceEventActionPlanProposed:
			proposed = event.ActionPlan
		case RealtimeVoiceEventAssistantResponseCompleted:
			completed = event.Response
		}
		if event.Type == RealtimeVoiceEventSessionCompleted && scenario.expect == liveGeminiVoiceExpectPlan {
			t.Fatalf("expected action plan to pause before session completion, got session completion\n%s", liveGeminiVoiceDiagnostics(events))
		}
	}
	switch scenario.expect {
	case liveGeminiVoiceExpectPlan:
		if proposed == nil {
			t.Fatalf("expected action plan, got spoken %q\n%s", spoken, liveGeminiVoiceDiagnostics(events))
		}
		if completed != nil || strings.TrimSpace(spoken) != "" {
			t.Fatalf("expected action plan to pause without final speech, response=%+v spoken=%q\n%s", completed, spoken, liveGeminiVoiceDiagnostics(events))
		}
		if scenario.assertPlan != nil {
			scenario.assertPlan(t, *proposed, events)
		}
	case liveGeminiVoiceExpectAnswer:
		if proposed != nil {
			t.Fatalf("expected answer, got action plan %+v\n%s", proposed, liveGeminiVoiceDiagnostics(events))
		}
		if completed == nil || strings.TrimSpace(spoken) == "" {
			t.Fatalf("expected spoken answer, response=%+v spoken=%q\n%s", completed, spoken, liveGeminiVoiceDiagnostics(events))
		}
		assertLiveGeminiVoiceTextContains(t, spoken, scenario.terms, events)
		if scenario.assertAnswer != nil {
			scenario.assertAnswer(t, spoken, events)
		}
	case liveGeminiVoiceExpectFallForward:
		if proposed != nil {
			t.Fatalf("expected fall-forward response, got action plan %+v\n%s", proposed, liveGeminiVoiceDiagnostics(events))
		}
		if completed == nil || strings.TrimSpace(spoken) == "" {
			t.Fatalf("expected spoken fall-forward response, response=%+v spoken=%q\n%s", completed, spoken, liveGeminiVoiceDiagnostics(events))
		}
		if completed.Kind != ports.StructuredAgentResponseKindClarification && completed.Kind != ports.StructuredAgentResponseKindUnsupportedAction && completed.Kind != ports.StructuredAgentResponseKindSafeFailure && completed.Kind != ports.StructuredAgentResponseKindAnswer {
			t.Fatalf("expected fall-forward kind, got %+v\n%s", completed, liveGeminiVoiceDiagnostics(events))
		}
		assertLiveGeminiVoiceTextContains(t, spoken, scenario.terms, events)
	default:
		t.Fatalf("unknown expectation %q", scenario.expect)
	}
}

func assertLiveGeminiVoiceLocativeAnswer(t *testing.T, spoken string, events []RealtimeVoiceEvent) {
	t.Helper()

	normalized := strings.ToLower(spoken)
	if strings.Contains(normalized, "visible match") || strings.Contains(normalized, "candidate") || strings.Contains(normalized, "resolution") {
		t.Fatalf("expected household language rather than search mechanics, got %q\n%s", spoken, liveGeminiVoiceDiagnostics(events))
	}
	if !strings.Contains(normalized, " in ") && !strings.Contains(normalized, " at ") && !strings.Contains(normalized, " inside ") {
		t.Fatalf("expected answer to say where the collection is, got %q\n%s", spoken, liveGeminiVoiceDiagnostics(events))
	}
}

func assertLiveGeminiVoiceTextContains(t *testing.T, text string, terms []string, events []RealtimeVoiceEvent) {
	t.Helper()

	normalized := strings.ToLower(text)
	for _, term := range terms {
		if !strings.Contains(normalized, strings.ToLower(term)) {
			t.Fatalf("expected spoken text %q to contain %q\n%s", text, term, liveGeminiVoiceDiagnostics(events))
		}
	}
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

func assertLiveGeminiVoiceCreateInOfficePlan(title string) func(*testing.T, RealtimeVoiceActionPlanProposal, []RealtimeVoiceEvent) {
	return func(t *testing.T, proposed RealtimeVoiceActionPlanProposal, events []RealtimeVoiceEvent) {
		t.Helper()

		assertLiveGeminiVoiceCreateInParentPlan(title, "office-1")(t, proposed, events)
	}
}

func assertLiveGeminiVoiceCreateInParentPlan(title string, parentAssetID string) func(*testing.T, RealtimeVoiceActionPlanProposal, []RealtimeVoiceEvent) {
	return func(t *testing.T, proposed RealtimeVoiceActionPlanProposal, events []RealtimeVoiceEvent) {
		t.Helper()

		for _, command := range proposed.Commands {
			if command.Kind == string(actionplan.CommandKindCreateAsset) && command.AssetKind == asset.KindItem.String() && liveGeminiVoiceTitleContains(command.Title, title) && command.ParentAssetID == parentAssetID {
				return
			}
		}
		t.Fatalf("expected create %q inside parent %q, got %+v\n%s", title, parentAssetID, proposed, liveGeminiVoiceDiagnostics(events))
	}
}

func assertLiveGeminiVoiceCreateItemInNewContainerUnderParentPlan(itemTitle string, containerTitle string, parentAssetID string) func(*testing.T, RealtimeVoiceActionPlanProposal, []RealtimeVoiceEvent) {
	return func(t *testing.T, proposed RealtimeVoiceActionPlanProposal, events []RealtimeVoiceEvent) {
		t.Helper()

		containerCommandID := ""
		for _, command := range proposed.Commands {
			if command.Kind == string(actionplan.CommandKindCreateAsset) && command.AssetKind == asset.KindContainer.String() && liveGeminiVoiceTitleContains(command.Title, containerTitle) && command.ParentAssetID == parentAssetID {
				containerCommandID = command.ID
				break
			}
		}
		if containerCommandID == "" {
			t.Fatalf("expected new container containing %q under parent %q, got %+v\n%s", containerTitle, parentAssetID, proposed, liveGeminiVoiceDiagnostics(events))
		}
		for _, command := range proposed.Commands {
			if command.Kind == string(actionplan.CommandKindCreateAsset) && command.AssetKind == asset.KindItem.String() && liveGeminiVoiceTitleContains(command.Title, itemTitle) && command.ParentCommandID == containerCommandID {
				return
			}
		}
		t.Fatalf("expected new item containing %q inside container command %q, got %+v\n%s", itemTitle, containerCommandID, proposed, liveGeminiVoiceDiagnostics(events))
	}
}

func assertLiveGeminiVoiceRemoteInNewTVBoxPlan(t *testing.T, proposed RealtimeVoiceActionPlanProposal, events []RealtimeVoiceEvent) {
	t.Helper()

	boxCommandID := ""
	for _, command := range proposed.Commands {
		if command.Kind == string(actionplan.CommandKindCreateAsset) && command.AssetKind == asset.KindContainer.String() && strings.Contains(strings.ToLower(command.Title), "box") && command.ParentAssetID == "living-room-1" {
			boxCommandID = command.ID
			break
		}
	}
	if boxCommandID == "" {
		t.Fatalf("expected new TV box container inside Living room, got %+v\n%s", proposed, liveGeminiVoiceDiagnostics(events))
	}
	for _, command := range proposed.Commands {
		if command.Kind == string(actionplan.CommandKindCreateAsset) && command.AssetKind == asset.KindItem.String() && strings.EqualFold(command.Title, "Apple TV remote") && command.ParentCommandID == boxCommandID {
			return
		}
	}
	t.Fatalf("expected Apple TV remote inside new TV box command %q, got %+v\n%s", boxCommandID, proposed, liveGeminiVoiceDiagnostics(events))
}

func assertLiveGeminiVoiceMoveToExistingLocationPlan(assetID string, parentAssetID string) func(*testing.T, RealtimeVoiceActionPlanProposal, []RealtimeVoiceEvent) {
	return func(t *testing.T, proposed RealtimeVoiceActionPlanProposal, events []RealtimeVoiceEvent) {
		t.Helper()

		for _, command := range proposed.Commands {
			if command.Kind == string(actionplan.CommandKindMoveAsset) && command.ParentAssetID == parentAssetID {
				return
			}
		}
		t.Fatalf("expected move %q into %q, got %+v\n%s", assetID, parentAssetID, proposed, liveGeminiVoiceDiagnostics(events))
	}
}

func assertLiveGeminiVoiceNestedMoveProposal(t *testing.T, proposed RealtimeVoiceActionPlanProposal, events []RealtimeVoiceEvent) {
	t.Helper()

	commandIDByTitle := map[string]string{}
	for _, command := range proposed.Commands {
		if command.Title != "" {
			commandIDByTitle[strings.ToLower(command.Title)] = command.ID
		}
	}
	kitchenID := commandIDByTitle["kitchen"]
	cabinetID := commandIDByTitle["big cabinet"]
	shelfID := commandIDByTitle["second shelf"]
	if kitchenID == "" || cabinetID == "" || shelfID == "" {
		t.Fatalf("expected Kitchen, Big cabinet, and Second shelf creates, got %+v\n%s", proposed, liveGeminiVoiceDiagnostics(events))
	}
	sawCabinetInKitchen := false
	sawShelfInCabinet := false
	sawMoveToShelf := false
	for _, command := range proposed.Commands {
		switch {
		case strings.EqualFold(command.Title, "Big cabinet") && command.ParentCommandID == kitchenID:
			sawCabinetInKitchen = true
		case strings.EqualFold(command.Title, "Second shelf") && command.ParentCommandID == cabinetID:
			sawShelfInCabinet = true
		case command.Kind == string(actionplan.CommandKindMoveAsset) && command.ParentCommandID == shelfID:
			sawMoveToShelf = true
		}
	}
	if !sawCabinetInKitchen || !sawShelfInCabinet || !sawMoveToShelf {
		t.Fatalf("expected nested create path and move into Second shelf, got %+v\n%s", proposed, liveGeminiVoiceDiagnostics(events))
	}
}

func assertLiveGeminiVoiceMoveToMissingNestedTitlesPlan(titles ...string) func(*testing.T, RealtimeVoiceActionPlanProposal, []RealtimeVoiceEvent) {
	return func(t *testing.T, proposed RealtimeVoiceActionPlanProposal, events []RealtimeVoiceEvent) {
		t.Helper()

		createdCommandIDByTitle := map[string]string{}
		for _, command := range proposed.Commands {
			if command.Kind == string(actionplan.CommandKindCreateAsset) || command.Kind == string(actionplan.CommandKindCreateLocation) {
				for _, title := range titles {
					if liveGeminiVoiceTitleContains(command.Title, title) {
						createdCommandIDByTitle[strings.ToLower(title)] = command.ID
					}
				}
			}
		}
		for _, title := range titles {
			if createdCommandIDByTitle[strings.ToLower(title)] == "" {
				t.Fatalf("expected created destination containing %q, got %+v\n%s", title, proposed, liveGeminiVoiceDiagnostics(events))
			}
		}
		deepestCommandID := createdCommandIDByTitle[strings.ToLower(titles[len(titles)-1])]
		for _, command := range proposed.Commands {
			if command.Kind == string(actionplan.CommandKindMoveAsset) && command.ParentCommandID == deepestCommandID {
				return
			}
		}
		t.Fatalf("expected move into deepest command %q, got %+v\n%s", deepestCommandID, proposed, liveGeminiVoiceDiagnostics(events))
	}
}

func assertLiveGeminiVoiceSingleExistingAssetPlan(kind actionplan.CommandKind, operation string, title string, assetKind string) func(*testing.T, RealtimeVoiceActionPlanProposal, []RealtimeVoiceEvent) {
	return func(t *testing.T, proposed RealtimeVoiceActionPlanProposal, events []RealtimeVoiceEvent) {
		t.Helper()

		if len(proposed.Commands) != 1 {
			t.Fatalf("expected one %s command for %q, got %+v\n%s", operation, title, proposed, liveGeminiVoiceDiagnostics(events))
		}
		command := proposed.Commands[0]
		if command.Kind != string(kind) || command.Operation != operation || !liveGeminiVoiceTitleContains(command.Title, title) || command.AssetKind != assetKind {
			t.Fatalf("expected %s %q %s command, got %+v\n%s", operation, title, assetKind, command, liveGeminiVoiceDiagnostics(events))
		}
	}
}

func liveGeminiVoiceTitleContains(actual string, expected string) bool {
	actual = strings.ToLower(actual)
	for _, word := range strings.Fields(strings.ToLower(expected)) {
		if !strings.Contains(actual, word) {
			return false
		}
	}
	return true
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

func liveGeminiVoiceFullTrace(events []RealtimeVoiceEvent, spoken string) string {
	builder := strings.Builder{}
	if strings.TrimSpace(spoken) != "" {
		builder.WriteString("spoken: ")
		builder.WriteString(spoken)
		builder.WriteString("\n")
	}
	for index, event := range events {
		builder.WriteString(fmt.Sprintf("%02d %s", index+1, event.Type))
		if strings.TrimSpace(event.Text) != "" {
			builder.WriteString(" text=")
			builder.WriteString(event.Text)
		}
		if strings.TrimSpace(event.Message) != "" {
			builder.WriteString(" message=")
			builder.WriteString(event.Message)
		}
		if strings.TrimSpace(event.ToolLabel) != "" {
			builder.WriteString(" tool=")
			builder.WriteString(event.ToolLabel)
		}
		if strings.TrimSpace(event.Code) != "" {
			builder.WriteString(" code=")
			builder.WriteString(event.Code)
		}
		builder.WriteString("\n")
		if strings.TrimSpace(event.Detail) != "" {
			builder.WriteString(safeRealtimeVoiceDiagnosticText(event.Detail, 8000))
			builder.WriteString("\n")
		}
		if event.Response != nil {
			payload, _ := json.MarshalIndent(event.Response, "", "  ")
			builder.Write(payload)
			builder.WriteString("\n")
		}
		if event.ActionPlan != nil {
			payload, _ := json.MarshalIndent(event.ActionPlan, "", "  ")
			builder.Write(payload)
			builder.WriteString("\n")
		}
	}
	return safeRealtimeVoiceDiagnosticText(builder.String(), 40000)
}
