package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) validateRealtimeVoiceProposalTypes(ctx context.Context, session RealtimeVoiceSession, commands []ports.ActionPlanCommandRecord) error {
	for _, command := range commands {
		if command.Kind == actionplan.CommandKindUpdateAsset {
			if len(commands) != 1 {
				return ports.ErrInvalidProviderInput
			}
			if err := a.validateRealtimeVoiceExpirationCorrection(ctx, session, command); err != nil {
				return err
			}
			continue
		}
		if command.Kind != actionplan.CommandKindCreateAsset && command.Kind != actionplan.CommandKindCreateLocation {
			continue
		}
		args, err := parseActionPlanCreateArguments(command)
		if err != nil {
			return err
		}
		if args.CustomAssetTypeID == "" {
			continue
		}
		if a.customAssetTypes == nil {
			return ports.ErrInvalidProviderInput
		}
		kind, found, err := a.customAssetTypes.CustomAssetTypeByID(ctx, session.TenantID, session.InventoryID, customfield.AssetTypeID(args.CustomAssetTypeID))
		if err != nil {
			return err
		}
		if !found || !kind.IsActive() || args.Expiration != nil && !kind.ExpirationEnabled {
			return ports.ErrInvalidProviderInput
		}
	}
	return nil
}

func (a App) validateRealtimeVoiceExpirationCorrection(ctx context.Context, session RealtimeVoiceSession, command ports.ActionPlanCommandRecord) error {
	args, err := parseActionPlanExpirationArguments(command)
	if err != nil {
		return err
	}
	item, err := a.GetAsset(ctx, GetAssetInput{Principal: session.Principal, TenantID: session.TenantID, InventoryID: session.InventoryID, AssetID: args.AssetID, Source: audit.SourceConversation})
	if err != nil {
		return err
	}
	if item.LifecycleState != asset.LifecycleStateActive {
		return ports.ErrInvalidProviderInput
	}
	if args.Expiration == nil {
		return nil
	}
	if a.customAssetTypes == nil {
		return ports.ErrInvalidProviderInput
	}
	kind, found, err := a.customAssetTypes.CustomAssetTypeByID(ctx, session.TenantID, session.InventoryID, customfield.AssetTypeID(item.CustomAssetTypeID))
	if err != nil {
		return err
	}
	if !found || !kind.IsActive() || !kind.ExpirationEnabled {
		return ports.ErrInvalidProviderInput
	}
	return nil
}
