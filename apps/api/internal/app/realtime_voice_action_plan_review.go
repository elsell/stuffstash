package app

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) realtimeVoiceActionPlanProposal(ctx context.Context, session RealtimeVoiceSession, record ports.ActionPlanRecord) (RealtimeVoiceActionPlanProposal, error) {
	commands := make([]RealtimeVoiceActionPlanCommand, 0, len(record.Commands))
	for _, command := range record.Commands {
		proposalCommand, err := a.realtimeVoiceActionPlanCommand(ctx, session, command)
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

func (a App) realtimeVoiceActionPlanCommand(ctx context.Context, session RealtimeVoiceSession, command ports.ActionPlanCommandRecord) (RealtimeVoiceActionPlanCommand, error) {
	proposal := RealtimeVoiceActionPlanCommand{
		ID:        command.ID,
		Kind:      string(command.Kind),
		Summary:   command.Summary,
		Operation: actionPlanCommandOperation(command.Kind),
	}
	if command.Kind == actionplan.CommandKindCreateAsset || command.Kind == actionplan.CommandKindCreateLocation {
		args, err := parseActionPlanCreateArguments(command)
		if err == nil {
			proposal.Title = args.Title
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
	} else if command.Kind == actionplan.CommandKindMoveAsset {
		args, err := parseActionPlanMoveArguments(command)
		if err == nil {
			moved, err := a.realtimeVoiceReviewAsset(ctx, session, args.AssetID.String())
			if err != nil {
				return RealtimeVoiceActionPlanCommand{}, err
			}
			proposal.AssetKind = moved.Kind.String()
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
		assetID, err := parseActionPlanAssetIDOnlyArguments(command)
		if err == nil {
			item, err := a.realtimeVoiceReviewAsset(ctx, session, assetID.String())
			if err != nil {
				return RealtimeVoiceActionPlanCommand{}, err
			}
			proposal.AssetKind = item.Kind.String()
			proposal.Title = item.Title.String()
		}
	} else if command.Kind == actionplan.CommandKindCheckoutAsset || command.Kind == actionplan.CommandKindReturnAsset {
		args, err := parseActionPlanCheckoutArguments(command)
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

func (a App) realtimeVoiceReviewAsset(ctx context.Context, session RealtimeVoiceSession, rawAssetID string) (asset.Asset, error) {
	assetID, ok := asset.NewID(rawAssetID)
	if !ok {
		return asset.Asset{}, ports.ErrInvalidProviderInput
	}
	item, found, err := a.assets.AssetByID(ctx, session.TenantID, session.InventoryID, assetID)
	if err != nil {
		return asset.Asset{}, err
	}
	if !found {
		return asset.Asset{}, ports.ErrInvalidProviderInput
	}
	return item, nil
}

func actionPlanCommandOperation(kind actionplan.CommandKind) string {
	switch kind {
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
