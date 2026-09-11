package app

import (
	"bytes"
	"encoding/json"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type actionPlanExpirationArguments struct {
	AssetID    asset.ID
	Expiration *assetapp.ExpirationInput
}

func parseActionPlanExpirationArguments(command ports.ActionPlanCommandRecord) (actionPlanExpirationArguments, error) {
	var raw map[string]json.RawMessage
	if json.Unmarshal(command.ArgumentsJSON, &raw) != nil || len(raw) != 2 {
		return actionPlanExpirationArguments{}, ErrValidation
	}
	id, err := actionPlanStringArgument(raw["assetId"])
	if err != nil {
		return actionPlanExpirationArguments{}, err
	}
	assetID, ok := asset.NewID(id)
	if !ok {
		return actionPlanExpirationArguments{}, ErrValidation
	}
	encoded, ok := raw["expiration"]
	if !ok {
		return actionPlanExpirationArguments{}, ErrValidation
	}
	args := actionPlanExpirationArguments{AssetID: assetID}
	if bytes.Equal(bytes.TrimSpace(encoded), []byte("null")) {
		return args, nil
	}
	var date assetapp.ExpirationInput
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&date) != nil {
		return actionPlanExpirationArguments{}, ErrValidation
	}
	if _, err := expirationdate.ParseDate(date.Date, expirationdate.Precision(date.Precision)); err != nil {
		return actionPlanExpirationArguments{}, ErrValidation
	}
	args.Expiration = &date
	return args, nil
}
func actionPlanAssetUpdateInput(input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (UpdateAssetInput, string, error) {
	if command.Kind == actionplan.CommandKindMoveAsset {
		update, err := actionPlanMoveAssetInput(input, command)
		return update, "move", err
	}
	args, err := parseActionPlanExpirationArguments(command)
	if err != nil {
		return UpdateAssetInput{}, "", err
	}
	return UpdateAssetInput{Principal: input.Principal, Source: audit.SourceConversation, RequestID: command.ID, TenantID: input.TenantID, InventoryID: input.InventoryID, AssetID: args.AssetID, Expiration: assetapp.ExpirationUpdate{Present: true, Value: args.Expiration}}, "update", nil
}
