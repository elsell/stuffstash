package agentmodel

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

// ReviewProposal projects an already-authorized plan within the caller's decision scope.
// Approval and execution remain separate commands with permission revalidation.
func (a ActionPlanService) ReviewProposal(ctx context.Context, session ActionPlanDecisionInput, record ports.ActionPlanRecord) (RealtimeVoiceActionPlanProposal, error) {
	commands := make([]RealtimeVoiceActionPlanCommand, 0, len(record.Commands))
	for _, command := range record.Commands {
		proposalCommand, err := a.ReviewCommand(ctx, session, command)
		if err != nil {
			return RealtimeVoiceActionPlanProposal{}, err
		}
		commands = append(commands, proposalCommand)
	}
	return RealtimeVoiceActionPlanProposal{
		PlanID:              record.ID,
		ConfirmationSummary: record.ConfirmationSummary,
		Commands:            commands,
		Risks:               append([]string{}, record.Risks...),
	}, nil
}

// ReviewCommand projects one command from an already-authorized review.
func (a ActionPlanService) ReviewCommand(ctx context.Context, session ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (RealtimeVoiceActionPlanCommand, error) {
	proposal := RealtimeVoiceActionPlanCommand{
		ID:        command.ID,
		Kind:      string(command.Kind),
		Summary:   command.Summary,
		Operation: ActionPlanCommandOperation(command.Kind),
	}
	if IsCustomizationCommand(command.Kind) {
		prepared, err := a.PrepareActionPlanCustomization(ctx, ActionPlanDecisionInput{Principal: session.Principal, TenantID: session.TenantID, InventoryID: session.InventoryID}, command)
		if err != nil {
			return RealtimeVoiceActionPlanCommand{}, err
		}
		proposal.Changes = prepared.Changes
		if prepared.AssetType != nil {
			proposal.Title = prepared.AssetType.Item.DisplayName.String()
		} else {
			proposal.Title = prepared.Definition.Item.DisplayName.String()
		}
		return proposal, nil
	}
	if command.Kind == actionplan.CommandKindCreateAsset || command.Kind == actionplan.CommandKindCreateLocation {
		args, err := ParseActionPlanCreateArguments(command)
		if err == nil {
			proposal.Title = args.Title
			proposal.Changes, err = a.actionPlanDetailChanges(ctx, session, nil, nil, args.CustomFields)
			if err != nil {
				return RealtimeVoiceActionPlanCommand{}, err
			}
			if args.Expiration != nil {
				proposal.Expiration = &RealtimeVoiceActionPlanExpiration{Date: args.Expiration.Date, Precision: args.Expiration.Precision}
			}
			proposal.AssetKind = args.Kind
			if command.Kind == actionplan.CommandKindCreateLocation {
				proposal.AssetKind = asset.KindLocation.String()
			}
			if proposal.AssetKind == "" {
				proposal.AssetKind = asset.KindItem.String()
			}
			proposal.ParentAssetID = args.ParentAssetID
			if args.ParentAssetID != "" {
				parent, err := a.realtimeVoiceReviewAsset(ctx, session, args.ParentAssetID)
				if err != nil {
					return RealtimeVoiceActionPlanCommand{}, err
				}
				proposal.ParentTitle = parent.Title.String()
				proposal.ParentKind = parent.Kind.String()
			}
			proposal.ParentCommandID = args.ParentCommandID
		}
	} else if command.Kind == actionplan.CommandKindUpdateAsset {
		args, err := ParseActionPlanUpdateArguments(command)
		if err != nil {
			return RealtimeVoiceActionPlanCommand{}, err
		}
		item, err := a.realtimeVoiceReviewAsset(ctx, session, args.AssetID.String())
		if err != nil {
			return RealtimeVoiceActionPlanCommand{}, err
		}
		proposal.Title = item.Title.String()
		proposal.AssetKind = item.Kind.String()
		proposal.Changes, err = a.actionPlanDetailChanges(ctx, session, args.Title, args.Description, args.CustomFields)
		if err != nil {
			return RealtimeVoiceActionPlanCommand{}, err
		}
		proposal.ExpirationCleared = args.ExpirationPresent && args.Expiration == nil
		if args.Expiration != nil {
			proposal.Expiration = &RealtimeVoiceActionPlanExpiration{Date: args.Expiration.Date, Precision: args.Expiration.Precision}
		}
	} else if command.Kind == actionplan.CommandKindMoveAsset {
		args, err := ParseActionPlanMoveArguments(command)
		if err == nil {
			moved, err := a.realtimeVoiceReviewAsset(ctx, session, args.AssetID.String())
			if err != nil {
				return RealtimeVoiceActionPlanCommand{}, err
			}
			proposal.AssetKind = moved.Kind.String()
			proposal.Title = moved.Title.String()
			proposal.ParentAssetID = args.ParentAssetID
			if args.ParentAssetID != "" {
				parent, err := a.realtimeVoiceReviewAsset(ctx, session, args.ParentAssetID)
				if err != nil {
					return RealtimeVoiceActionPlanCommand{}, err
				}
				proposal.ParentTitle = parent.Title.String()
				proposal.ParentKind = parent.Kind.String()
			}
			proposal.ParentCommandID = args.ParentCommandID
		}
	} else if command.Kind == actionplan.CommandKindArchiveAsset || command.Kind == actionplan.CommandKindRestoreAsset {
		assetID, err := ParseActionPlanAssetIDOnlyArguments(command)
		if err == nil {
			item, err := a.realtimeVoiceReviewAsset(ctx, session, assetID.String())
			if err != nil {
				return RealtimeVoiceActionPlanCommand{}, err
			}
			proposal.AssetKind = item.Kind.String()
			proposal.Title = item.Title.String()
		}
	} else if command.Kind == actionplan.CommandKindCheckoutAsset || command.Kind == actionplan.CommandKindReturnAsset {
		args, err := ParseActionPlanCheckoutArguments(command)
		if err == nil {
			item, err := a.realtimeVoiceReviewAsset(ctx, session, args.AssetID.String())
			if err != nil {
				return RealtimeVoiceActionPlanCommand{}, err
			}
			proposal.AssetKind = item.Kind.String()
			proposal.Title = item.Title.String()
		}
	}
	return proposal, nil
}

func (a ActionPlanService) realtimeVoiceReviewAsset(ctx context.Context, session ActionPlanDecisionInput, rawAssetID string) (asset.Asset, error) {
	assetID, ok := asset.NewID(rawAssetID)
	if !ok {
		return asset.Asset{}, ports.ErrInvalidProviderInput
	}
	item, found, err := a.deps.Assets.AssetByID(ctx, session.TenantID, session.InventoryID, assetID)
	if err != nil {
		return asset.Asset{}, err
	}
	if !found {
		return asset.Asset{}, ports.ErrInvalidProviderInput
	}
	return item, nil
}

func ActionPlanCommandOperation(kind actionplan.CommandKind) string {
	switch kind {
	case actionplan.CommandKindCreateCustomAssetType, actionplan.CommandKindCreateCustomFieldDefinition:
		return "configure"
	case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation:
		return "create"
	case actionplan.CommandKindMoveAsset:
		return "move"
	case actionplan.CommandKindArchiveAsset:
		return "archive"
	case actionplan.CommandKindRestoreAsset:
		return "restore"
	case actionplan.CommandKindCheckoutAsset:
		return "checkout"
	case actionplan.CommandKindReturnAsset:
		return "return"
	default:
		return "update"
	}
}
