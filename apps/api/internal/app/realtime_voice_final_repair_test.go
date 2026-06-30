package app

import (
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestRealtimeVoiceRepairsWriteClaimAfterRejectedActionPlan(t *testing.T) {
	t.Parallel()

	rejected, err := realtimeVoiceToolErrorResult(ports.AgentToolCall{
		ID:   "tool-plan-bad",
		Name: RealtimeVoiceToolProposeActionPlan,
	}, "invalid_tool_request", "The action-plan request was invalid or incomplete.", true)
	if err != nil {
		t.Fatalf("build rejected tool result: %v", err)
	}
	response := ports.StructuredAgentResponse{
		Kind:            ports.StructuredAgentResponseKindAnswer,
		SpokenResponse:  "I have added the Apple TV remote to the box under the TV.",
		DisplayResponse: "Added Apple TV remote.",
	}

	if !realtimeVoiceShouldRepairWriteClaimAfterFailedProposal("Add an Apple TV remote to the box under the TV.", response, []ports.AgentToolResult{rejected}) {
		t.Fatalf("expected write claim after rejected action plan to be repaired")
	}
	repair, err := realtimeVoiceFinalWriteClaimRepairResult("repair-1")
	if err != nil {
		t.Fatalf("build repair result: %v", err)
	}
	if !strings.Contains(repair.Content, "No inventory change has been applied") || !strings.Contains(repair.Content, "propose_action_plan succeeds") {
		t.Fatalf("expected safe repair guidance, got %s", repair.Content)
	}
}

func TestRealtimeVoiceDoesNotRepairSafeWriteClarificationAfterRejectedActionPlan(t *testing.T) {
	t.Parallel()

	rejected, err := realtimeVoiceToolErrorResult(ports.AgentToolCall{
		ID:   "tool-plan-bad",
		Name: RealtimeVoiceToolProposeActionPlan,
	}, "invalid_tool_request", "The action-plan request was invalid or incomplete.", true)
	if err != nil {
		t.Fatalf("build rejected tool result: %v", err)
	}
	response := ports.StructuredAgentResponse{
		Kind:            ports.StructuredAgentResponseKindClarification,
		SpokenResponse:  "I need to know which room you mean.",
		DisplayResponse: "Which room do you mean?",
	}

	if realtimeVoiceShouldRepairWriteClaimAfterFailedProposal("Add an Apple TV remote to the box.", response, []ports.AgentToolResult{rejected}) {
		t.Fatalf("did not expect safe clarification to be repaired")
	}
}

func TestRealtimeVoiceReadQuestionsWithWriteVerbsAreNotWriteRequests(t *testing.T) {
	t.Parallel()

	for _, transcript := range []string{
		"Where did I put my water bottle?",
		"What did I put in the toolbox?",
		"Do I have any batteries?",
		"Find my drill.",
	} {
		if realtimeVoiceLooksLikeWriteRequest(transcript) {
			t.Fatalf("expected %q to stay read-only", transcript)
		}
		if realtimeVoiceLooksLikeMoveRequest(transcript) {
			t.Fatalf("expected %q not to look like a move request", transcript)
		}
	}
}

func TestRealtimeVoiceDoesNotRepairCreateClarificationWhenRequestedSourceWasNotVisible(t *testing.T) {
	t.Parallel()

	officeResult := ports.AgentToolResult{
		Name:    RealtimeVoiceToolSearchAuthorizedAssets,
		Content: `{"tool":"search_authorized_assets","query":"office","count":1,"items":[{"assetId":"office-1","title":"Office","kind":"location"}]}`,
	}
	response := ports.StructuredAgentResponse{
		Kind:            ports.StructuredAgentResponseKindClarification,
		SpokenResponse:  "I can't find your passport. Do you want me to create one?",
		DisplayResponse: "I can't find your passport. Do you want me to create one?",
	}

	if realtimeVoiceShouldRepairCreateClarification("Move my passport to the office.", response, []ports.AgentToolResult{officeResult}) {
		t.Fatalf("did not expect repair when requested source was not visible")
	}
}

func TestRealtimeVoiceRepairsCreateClarificationWhenRequestedSourceWasVisible(t *testing.T) {
	t.Parallel()

	waterBottleResult := ports.AgentToolResult{
		Name:    RealtimeVoiceToolSearchAuthorizedAssets,
		Content: `{"tool":"search_authorized_assets","query":"water bottle","count":1,"items":[{"assetId":"water-bottle-1","title":"Water bottle","kind":"item"}]}`,
	}
	response := ports.StructuredAgentResponse{
		Kind:            ports.StructuredAgentResponseKindClarification,
		SpokenResponse:  "I can't find the kitchen. Do you want me to create it?",
		DisplayResponse: "I can't find the kitchen. Do you want me to create it?",
	}

	if !realtimeVoiceShouldRepairCreateClarification("Move my water bottle to the kitchen.", response, []ports.AgentToolResult{waterBottleResult}) {
		t.Fatalf("expected repair when requested source was visible and destination creation should be proposed")
	}
}

func TestRealtimeVoiceRejectsRootItemCreateWhenTranscriptNamesDestination(t *testing.T) {
	t.Parallel()

	err := validateRealtimeVoiceRootCreate(actionplan.CommandKindCreateAsset, "item", "Add a phone charger to the office.")
	if err == nil {
		t.Fatalf("expected root item create with named destination to be rejected")
	}
	if err := validateRealtimeVoiceRootCreate(actionplan.CommandKindCreateAsset, "item", "Add a phone charger."); err != nil {
		t.Fatalf("expected root item create without destination to pass: %v", err)
	}
	if err := validateRealtimeVoiceRootCreate(actionplan.CommandKindCreateLocation, "location", "Move my water bottle to the kitchen."); err != nil {
		t.Fatalf("expected root location create to pass: %v", err)
	}
}

func TestRealtimeVoiceTitleMentionedInTranscriptHandlesPunctuation(t *testing.T) {
	t.Parallel()

	if !realtimeVoiceTitleMentionedInTranscript("Office", "Add a phone charger to the office.") {
		t.Fatalf("expected title mention to survive punctuation")
	}
	if realtimeVoiceTitleMentionedInTranscript("Garage", "Add a phone charger to the office.") {
		t.Fatalf("did not expect unrelated title mention")
	}
}
