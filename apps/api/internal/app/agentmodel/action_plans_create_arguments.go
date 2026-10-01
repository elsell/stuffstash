package agentmodel

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func actionPlanCreateAssetInput(input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (ports.CreateAssetInput, error) {
	args, err := ParseActionPlanCreateArguments(command)
	if err != nil {
		return ports.CreateAssetInput{}, err
	}
	kind := args.Kind
	if command.Kind == actionplan.CommandKindCreateLocation {
		kind = "location"
	}
	if strings.TrimSpace(kind) == "" {
		kind = "item"
	}
	return ports.CreateAssetInput{
		Principal:         input.Principal,
		Source:            audit.SourceConversation,
		RequestID:         command.ID,
		TenantID:          input.TenantID,
		InventoryID:       input.InventoryID,
		Kind:              kind,
		Title:             args.Title,
		Description:       args.Description,
		ParentAssetID:     args.ParentAssetID,
		CustomFields:      args.CustomFields,
		CustomAssetTypeID: args.CustomAssetTypeID,
		Expiration:        args.Expiration,
	}, nil
}

type ActionPlanCreateArguments struct {
	CustomFields      map[string]any
	CustomAssetTypeID string
	Expiration        *ports.ExpirationInput
	Title             string
	Kind              string
	Description       string
	ParentAssetID     string
	ParentCommandID   string
}

func ParseActionPlanCreateArguments(command ports.ActionPlanCommandRecord) (ActionPlanCreateArguments, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(command.ArgumentsJSON, &raw); err != nil {
		return ActionPlanCreateArguments{}, apperrors.ErrValidation
	}
	args := ActionPlanCreateArguments{}
	for key, value := range raw {
		switch key {
		case "customFields":
			fields, err := parseActionPlanCustomFields(value)
			if err != nil {
				return ActionPlanCreateArguments{}, err
			}
			for _, fieldValue := range fields {
				if fieldValue == nil {
					return ActionPlanCreateArguments{}, apperrors.ErrValidation
				}
			}
			args.CustomFields = fields
		case "customAssetTypeId":
			text, err := actionPlanStringArgument(value)
			if err != nil || text == "" {
				return ActionPlanCreateArguments{}, apperrors.ErrValidation
			}
			args.CustomAssetTypeID = text
		case "expiration":
			var expiration ports.ExpirationInput
			decoder := json.NewDecoder(bytes.NewReader(value))
			decoder.DisallowUnknownFields()
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) || decoder.Decode(&expiration) != nil {
				return ActionPlanCreateArguments{}, apperrors.ErrValidation
			}
			if _, err := expirationdate.ParseDate(expiration.Date, expirationdate.Precision(expiration.Precision)); err != nil {
				return ActionPlanCreateArguments{}, apperrors.ErrValidation
			}
			args.Expiration = &expiration
		case "title", "name":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return ActionPlanCreateArguments{}, err
			}
			if strings.TrimSpace(args.Title) == "" {
				args.Title = text
			}
		case "kind":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return ActionPlanCreateArguments{}, err
			}
			args.Kind = text
		case "description":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return ActionPlanCreateArguments{}, err
			}
			args.Description = text
		case "parentAssetId":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return ActionPlanCreateArguments{}, err
			}
			args.ParentAssetID = text
		case "parentCommandId":
			text, err := actionPlanStringArgument(value)
			if err != nil {
				return ActionPlanCreateArguments{}, err
			}
			args.ParentCommandID = text
		default:
			return ActionPlanCreateArguments{}, apperrors.ErrValidation
		}
	}
	if strings.TrimSpace(args.Title) == "" || args.Expiration != nil && args.CustomAssetTypeID == "" {
		return ActionPlanCreateArguments{}, apperrors.ErrValidation
	}
	if strings.TrimSpace(args.ParentAssetID) != "" && strings.TrimSpace(args.ParentCommandID) != "" {
		return ActionPlanCreateArguments{}, apperrors.ErrValidation
	}
	switch strings.TrimSpace(args.Kind) {
	case "", "item", "container", "location":
		if command.Kind == actionplan.CommandKindCreateLocation && strings.TrimSpace(args.Kind) != "" && strings.TrimSpace(args.Kind) != "location" {
			return ActionPlanCreateArguments{}, apperrors.ErrValidation
		}
		return args, nil
	default:
		return ActionPlanCreateArguments{}, apperrors.ErrValidation
	}
}
