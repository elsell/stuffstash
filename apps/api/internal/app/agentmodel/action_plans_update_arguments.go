package agentmodel

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ActionPlanUpdateArguments struct {
	AssetID            asset.ID
	Title, Description *string
	Expiration         *ports.ExpirationInput
	ExpirationPresent  bool
	CustomFields       map[string]any
}

func ParseActionPlanUpdateArguments(command ports.ActionPlanCommandRecord) (ActionPlanUpdateArguments, error) {
	var raw map[string]json.RawMessage
	if json.Unmarshal(command.ArgumentsJSON, &raw) != nil || len(raw) < 2 {
		return ActionPlanUpdateArguments{}, apperrors.ErrValidation
	}
	var args ActionPlanUpdateArguments
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
				return args, apperrors.ErrValidation
			}
		case "title", "description":
			var text string
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) || json.Unmarshal(value, &text) != nil {
				return args, apperrors.ErrValidation
			}
			if key == "title" {
				if _, ok := asset.NewTitle(text); !ok {
					return args, apperrors.ErrValidation
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
			var date ports.ExpirationInput
			decoder := json.NewDecoder(bytes.NewReader(value))
			decoder.DisallowUnknownFields()
			if decoder.Decode(&date) != nil {
				return args, apperrors.ErrValidation
			}
			if _, err := expirationdate.ParseDate(date.Date, expirationdate.Precision(date.Precision)); err != nil {
				return args, apperrors.ErrValidation
			}
			args.Expiration = &date
		case "customFields":
			fields, err := parseActionPlanCustomFields(value)
			if err != nil {
				return args, err
			}
			args.CustomFields = fields
		default:
			return args, apperrors.ErrValidation
		}
	}
	if args.AssetID == "" {
		return args, apperrors.ErrValidation
	}
	return args, nil
}

func parseActionPlanCustomFields(raw json.RawMessage) (map[string]any, error) {
	var fields map[string]any
	if json.Unmarshal(raw, &fields) != nil || len(fields) == 0 || len(fields) > 10 {
		return nil, apperrors.ErrValidation
	}
	for rawKey, value := range fields {
		key, ok := customfield.NewKey(rawKey)
		if !ok || key.String() != rawKey || strings.TrimSpace(rawKey) != rawKey {
			return nil, apperrors.ErrValidation
		}
		switch value.(type) {
		case nil, string, bool, float64:
		default:
			return nil, apperrors.ErrValidation
		}
	}
	return fields, nil
}

func actionPlanAssetUpdateInput(input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (ports.UpdateAssetInput, string, error) {
	if command.Kind == actionplan.CommandKindMoveAsset {
		update, err := actionPlanMoveAssetInput(input, command)
		return update, "move", err
	}
	args, err := ParseActionPlanUpdateArguments(command)
	if err != nil {
		return ports.UpdateAssetInput{}, "", err
	}
	return ports.UpdateAssetInput{Principal: input.Principal, Source: audit.SourceConversation, RequestID: command.ID, TenantID: input.TenantID, InventoryID: input.InventoryID, AssetID: args.AssetID, Title: args.Title, Description: args.Description, CustomFieldPatch: args.CustomFields, Expiration: ports.ExpirationUpdate{Present: args.ExpirationPresent, Value: args.Expiration}}, "update", nil
}
