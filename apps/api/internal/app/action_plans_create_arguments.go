package app

import (
	"bytes"
	"encoding/json"
	assetapp "github.com/stuffstash/stuff-stash/internal/app/assets"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"strings"
)

func actionPlanCreateAssetInput(input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (CreateAssetInput, error) {
	args, err := parseActionPlanCreateArguments(command)
	if err != nil {
		return CreateAssetInput{}, err
	}
	kind := args.Kind
	if command.Kind == actionplan.CommandKindCreateLocation {
		kind = "location"
	}
	if strings.TrimSpace(kind) == "" {
		kind = "item"
	}
	return CreateAssetInput{
		Principal:         input.Principal,
		Source:            audit.SourceConversation,
		RequestID:         command.ID,
		TenantID:          input.TenantID,
		InventoryID:       input.InventoryID,
		Kind:              kind,
		Title:             args.Title,
		Description:       args.Description,
		ParentAssetID:     args.ParentAssetID,
		CustomFields:      map[string]any{},
		CustomAssetTypeID: args.CustomAssetTypeID,
		Expiration:        args.Expiration,
	}, nil
}

type actionPlanCreateArguments struct {
	CustomAssetTypeID string
	Expiration        *assetapp.ExpirationInput
	Title             string
	Kind              string
	Description       string
	ParentAssetID     string
	ParentCommandID   string
}

func parseActionPlanCreateArguments(command ports.ActionPlanCommandRecord) (actionPlanCreateArguments, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(command.ArgumentsJSON, &raw); err != nil {
		return actionPlanCreateArguments{}, ErrValidation
	}
	args := actionPlanCreateArguments{}
	for key, value := range raw {
		switch key {
		case "customAssetTypeId":
			text, err := actionPlanStringArgument(value)
			if err != nil || text == "" {
				return actionPlanCreateArguments{}, ErrValidation
			}
			args.CustomAssetTypeID = text
		case "expiration":
			var expiration assetapp.ExpirationInput
			decoder := json.NewDecoder(bytes.NewReader(value))
			decoder.DisallowUnknownFields()
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) || decoder.Decode(&expiration) != nil {
				return actionPlanCreateArguments{}, ErrValidation
			}
			if _, err := expirationdate.ParseDate(expiration.Date, expirationdate.Precision(expiration.Precision)); err != nil {
				return actionPlanCreateArguments{}, ErrValidation
			}
			args.Expiration = &expiration
		case "title", "name":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanCreateArguments{}, err
			}
			if strings.TrimSpace(args.Title) == "" {
				args.Title = text
			}
		case "kind":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanCreateArguments{}, err
			}
			args.Kind = text
		case "description":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanCreateArguments{}, err
			}
			args.Description = text
		case "parentAssetId":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanCreateArguments{}, err
			}
			args.ParentAssetID = text
		case "parentCommandId":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return actionPlanCreateArguments{}, err
			}
			args.ParentCommandID = text
		default:
			return actionPlanCreateArguments{}, ErrValidation
		}
	}
	if strings.TrimSpace(args.Title) == "" || args.Expiration != nil && args.CustomAssetTypeID == "" {
		return actionPlanCreateArguments{}, ErrValidation
	}
	if strings.TrimSpace(args.ParentAssetID) != "" && strings.TrimSpace(args.ParentCommandID) != "" {
		return actionPlanCreateArguments{}, ErrValidation
	}
	switch strings.TrimSpace(args.Kind) {
	case "", "item", "container", "location":
		if command.Kind == actionplan.CommandKindCreateLocation && strings.TrimSpace(args.Kind) != "" && strings.TrimSpace(args.Kind) != "location" {
			return actionPlanCreateArguments{}, ErrValidation
		}
		return args, nil
	default:
		return actionPlanCreateArguments{}, ErrValidation
	}
}
