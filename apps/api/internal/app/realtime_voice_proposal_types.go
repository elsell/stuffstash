package app

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) validateRealtimeVoiceProposalTypes(ctx context.Context, session RealtimeVoiceSession, commands []ports.ActionPlanCommandRecord) error {
	for _, command := range commands {
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
