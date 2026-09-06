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
