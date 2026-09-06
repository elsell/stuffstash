package app

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const realtimeConversationProposeTool = "propose_inventory_change"

type conversationProposalArguments struct {
	Summary  string                   `json:"summary"`
	Commands []ActionPlanCommandInput `json:"commands"`
	Risks    []string                 `json:"risks"`
}

func (e *realtimeConversationTools) propose(ctx context.Context, call ports.AgentToolCall) (ports.ConversationToolOutcome, error) {
	if err := e.application.ensureActiveInventoryAccess(ctx, e.session.Principal, e.session.TenantID, e.session.InventoryID, ports.InventoryPermissionEditAsset); err != nil {
		return ports.ConversationToolOutcome{}, err
	}
	encoded, err := json.Marshal(call.Arguments)
	if err != nil {
		return ports.ConversationToolOutcome{}, ports.ErrInvalidProviderInput
	}
	var input conversationProposalArguments
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || strings.TrimSpace(input.Summary) == "" || len(input.Summary) > maxActionPlanSummaryLength {
		return ports.ConversationToolOutcome{}, ports.ErrInvalidProviderInput
	}
	for _, command := range input.Commands {
		if !validActionPlanCommandID(command.ID) {
			return ports.ConversationToolOutcome{}, ports.ErrInvalidProviderInput
		}
	}
	commands, err := e.application.actionPlanCommands(input.Commands)
	if err != nil {
		return ports.ConversationToolOutcome{}, err
	}
	checked := map[string]bool{}
	for _, command := range input.Commands {
		for _, key := range []string{"assetId", "parentAssetId"} {
			raw, _ := command.Arguments[key].(string)
			id := strings.TrimSpace(raw)
			if id == "" || checked[id] {
				continue
			}
			if _, ok := e.visible[id]; !ok {
				return ports.ConversationToolOutcome{}, ports.ErrInvalidProviderInput
			}
			if _, err := e.application.GetAsset(ctx, GetAssetInput{Principal: e.session.Principal, TenantID: e.session.TenantID, InventoryID: e.session.InventoryID, AssetID: asset.ID(id), Source: audit.SourceConversation}); err != nil {
				return ports.ConversationToolOutcome{}, err
			}
			checked[id] = true
		}
	}
	// Resolve fallible review metadata before persisting a draft. A failed read
	// may be retried safely because no plan has yet been created.
	proposal, err := e.application.realtimeVoiceActionPlanProposal(ctx, e.session, ports.ActionPlanRecord{Commands: commands, ConfirmationSummary: input.Summary, Risks: input.Risks})
	if err != nil {
		return ports.ConversationToolOutcome{}, err
	}
	record, err := e.application.CreateActionPlan(ctx, CreateActionPlanInput{Principal: e.session.Principal, TenantID: e.session.TenantID, InventoryID: e.session.InventoryID, Source: e.session.Source, RealtimeSessionID: e.session.ID, IntentSummary: input.Summary, ConfirmationSummary: input.Summary, Commands: input.Commands, Risks: input.Risks})
	if err != nil {
		return ports.ConversationToolOutcome{}, err
	}
	proposal.PlanID = record.ID
	proposal.ConfirmationSummary = record.ConfirmationSummary
	proposal.Risks = append([]string(nil), record.Risks...)
	e.proposal = &proposal
	return ports.ConversationToolOutcome{ApprovalPlanID: record.ID, Result: ports.AgentToolResult{CallID: call.ID, Name: call.Name, Call: call, Content: `{"status":"awaiting_user_approval","executed":false}`}}, nil
}

func realtimeConversationProposalTool() ports.ConversationToolDefinition {
	return ports.ConversationToolDefinition{Name: realtimeConversationProposeTool, Description: "Prepare an inventory change for user approval; never execute it. Search for existing items first. Use existing assetId/parentAssetId only from tool results. Commands may depend on earlier create commands via parentCommandId. Put all related commands in one ordered proposal; execution pauses for review immediately. Move existing items rather than duplicating them. An explicitly additional physical item may be created.", Parameters: json.RawMessage(`{"type":"object","properties":{"summary":{"type":"string"},"risks":{"type":"array","items":{"type":"string"}},"commands":{"type":"array","minItems":1,"maxItems":10,"items":{"type":"object","properties":{"id":{"type":"string"},"kind":{"type":"string","enum":["create_asset","create_location","move_asset","archive_asset","restore_asset","checkout_asset","return_asset"]},"summary":{"type":"string"},"arguments":{"type":"object","properties":{"assetId":{"type":"string"},"parentAssetId":{"type":"string"},"parentCommandId":{"type":"string"},"title":{"type":"string"},"kind":{"type":"string","enum":["item","container","location"]},"description":{"type":"string"},"details":{"type":"string"}},"additionalProperties":false}},"required":["id","kind","summary","arguments"],"additionalProperties":false}}},"required":["summary","commands"],"additionalProperties":false}`)}
}
