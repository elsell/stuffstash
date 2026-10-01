package app

import (
	"bytes"
	"encoding/json"
	"strings"

	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type actionPlanUpdateArguments struct {
	AssetID            asset.ID
	Title, Description *string
	Expiration         *assetapp.ExpirationInput
	ExpirationPresent  bool
	CustomFields       map[string]any
}

func parseActionPlanUpdateArguments(command ports.ActionPlanCommandRecord) (actionPlanUpdateArguments, error) {
	var raw map[string]json.RawMessage
	if json.Unmarshal(command.ArgumentsJSON, &raw) != nil || len(raw) < 2 {
		return actionPlanUpdateArguments{}, ErrValidation
	}
	var args actionPlanUpdateArguments
	for key, value := range raw {
		switch key {
		case "assetId":
			id, err := actionPlanStringArgument(value)
			if err != nil {
				return args, err
			}
			var ok bool
			args.AssetID, ok = asset.NewID(id)
			if !ok {
				return args, ErrValidation
			}
		case "title", "description":
			var text string
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) || json.Unmarshal(value, &text) != nil {
				return args, ErrValidation
			}
			if key == "title" {
				if _, ok := asset.NewTitle(text); !ok {
					return args, ErrValidation
				}
				args.Title = &text
			} else {
				args.Description = &text
			}
		case "expiration":
			args.ExpirationPresent = true
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				continue
			}
			var date assetapp.ExpirationInput
			decoder := json.NewDecoder(bytes.NewReader(value))
			decoder.DisallowUnknownFields()
			if decoder.Decode(&date) != nil {
				return args, ErrValidation
			}
			if _, err := expirationdate.ParseDate(date.Date, expirationdate.Precision(date.Precision)); err != nil {
				return args, ErrValidation
			}
			args.Expiration = &date
		case "customFields":
			fields, err := parseActionPlanCustomFields(value)
			if err != nil {
				return args, err
			}
			args.CustomFields = fields
		default:
			return args, ErrValidation
		}
	}
	if args.AssetID == "" {
		return args, ErrValidation
	}
	return args, nil
}

func parseActionPlanCustomFields(raw json.RawMessage) (map[string]any, error) {
	var fields map[string]any
	if json.Unmarshal(raw, &fields) != nil || len(fields) == 0 || len(fields) > 10 {
		return nil, ErrValidation
	}
	for rawKey, value := range fields {
		key, ok := customfield.NewKey(rawKey)
		if !ok || key.String() != rawKey || strings.TrimSpace(rawKey) != rawKey {
			return nil, ErrValidation
		}
		switch value.(type) {
		case nil, string, bool, float64:
		default:
			return nil, ErrValidation
		}
	}
	return fields, nil
}

func actionPlanAssetUpdateInput(input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (UpdateAssetInput, string, error) {
	if command.Kind == actionplan.CommandKindMoveAsset {
		update, err := actionPlanMoveAssetInput(input, command)
		return update, "move", err
	}
	args, err := parseActionPlanUpdateArguments(command)
	if err != nil {
		return UpdateAssetInput{}, "", err
	}
	return UpdateAssetInput{Principal: input.Principal, Source: audit.SourceConversation, RequestID: command.ID, TenantID: input.TenantID, InventoryID: input.InventoryID, AssetID: args.AssetID, Title: args.Title, Description: args.Description, CustomFieldPatch: args.CustomFields, Expiration: assetapp.ExpirationUpdate{Present: args.ExpirationPresent, Value: args.Expiration}}, "update", nil
}
