package app

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	customfieldapp "github.com/stuffstash/stuff-stash/internal/app/customfields"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type actionPlanCustomizationArguments struct {
	Key                string   `json:"key"`
	DisplayName        string   `json:"displayName"`
	Description        string   `json:"description,omitempty"`
	ExpirationEnabled  bool     `json:"expirationEnabled,omitempty"`
	FieldType          string   `json:"fieldType,omitempty"`
	Applicability      string   `json:"applicability,omitempty"`
	EnumOptions        []string `json:"enumOptions,omitempty"`
	CustomAssetTypeIDs []string `json:"customAssetTypeIds,omitempty"`
}

func isCustomizationCommand(kind actionplan.CommandKind) bool {
	return kind == actionplan.CommandKindCreateCustomAssetType || kind == actionplan.CommandKindCreateCustomFieldDefinition
}
func parseActionPlanCustomizationArguments(command ports.ActionPlanCommandRecord) (actionPlanCustomizationArguments, error) {
	var args actionPlanCustomizationArguments
	decoder := json.NewDecoder(bytes.NewReader(command.ArgumentsJSON))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&args) != nil {
		return args, ErrValidation
	}
	var raw map[string]json.RawMessage
	if json.Unmarshal(command.ArgumentsJSON, &raw) != nil {
		return args, ErrValidation
	}
	allowed := map[string]bool{"key": true, "displayName": true}
	if command.Kind == actionplan.CommandKindCreateCustomAssetType {
		allowed["description"] = true
		allowed["expirationEnabled"] = true
	} else if command.Kind == actionplan.CommandKindCreateCustomFieldDefinition {
		for _, key := range []string{"fieldType", "applicability", "enumOptions", "customAssetTypeIds"} {
			allowed[key] = true
		}
		if _, ok := customfield.NewFieldType(args.FieldType); !ok {
			return args, ErrValidation
		}
		if _, ok := customfield.NewApplicability(args.Applicability); !ok {
			return args, ErrValidation
		}
	} else {
		return args, ErrValidation
	}
	for key, value := range raw {
		if !allowed[key] || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return args, ErrValidation
		}
	}
	if key, ok := customfield.NewKey(args.Key); !ok || key.String() != args.Key {
		return args, ErrValidation
	}
	if _, ok := customfield.NewDisplayName(args.DisplayName); !ok {
		return args, ErrValidation
	}
	if len(args.CustomAssetTypeIDs) > 10 || len(args.EnumOptions) > 50 {
		return args, ErrValidation
	}
	return args, nil
}

type preparedActionPlanCustomization struct {
	assetType  *customfieldapp.PreparedCustomAssetType
	definition *customfieldapp.PreparedCustomFieldDefinition
	typeInput  CreateCustomAssetTypeInput
	fieldInput CreateCustomFieldDefinitionInput
	changes    []string
}

func (a App) prepareActionPlanCustomization(ctx context.Context, input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (preparedActionPlanCustomization, error) {
	var result preparedActionPlanCustomization
	if a.actionPlanCustomizations == nil {
		return result, ErrValidation
	}
	args, err := parseActionPlanCustomizationArguments(command)
	if err != nil {
		return result, err
	}
	if command.Kind == actionplan.CommandKindCreateCustomAssetType {
		result.typeInput = CreateCustomAssetTypeInput{Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID, Source: audit.SourceConversation, RequestID: command.ID, Key: args.Key, DisplayName: args.DisplayName, Description: args.Description, ExpirationEnabled: args.ExpirationEnabled}
		prepared, err := a.customFieldService.PrepareInventoryCustomAssetType(ctx, result.typeInput)
		if err != nil {
			return result, err
		}
		result.assetType = &prepared
		tracking := "Off"
		if args.ExpirationEnabled {
			tracking = "On"
		}
		result.changes = []string{"Type name: " + prepared.Item.DisplayName.String(), "Key: " + args.Key, "Expiration tracking: " + tracking}
		if prepared.Item.Description.String() != "" {
			result.changes = append(result.changes, "Description: "+prepared.Item.Description.String())
		}
	} else {
		result.fieldInput = CreateCustomFieldDefinitionInput{Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID, Source: audit.SourceConversation, RequestID: command.ID, Key: args.Key, DisplayName: args.DisplayName, Type: args.FieldType, Applicability: args.Applicability, EnumOptions: args.EnumOptions, CustomAssetTypeIDs: args.CustomAssetTypeIDs}
		prepared, err := a.customFieldService.PrepareInventoryCustomFieldDefinition(ctx, result.fieldInput)
		if err != nil {
			return result, err
		}
		result.definition = &prepared
		names := []string{}
		for _, id := range prepared.Item.CustomAssetTypeIDs {
			item, found, err := a.customAssetTypes.CustomAssetTypeByID(ctx, input.TenantID, input.InventoryID, id)
			if err != nil {
				return result, err
			}
			if !found {
				return result, ErrValidation
			}
			names = append(names, item.DisplayName.String())
		}
		target := "All assets"
		if len(names) > 0 {
			target = strings.Join(names, ", ")
		}
		result.changes = []string{"Field name: " + prepared.Item.DisplayName.String(), "Key: " + args.Key, "Field type: " + strings.ReplaceAll(args.FieldType, "_", " "), "Applies to: " + target}
		if len(args.EnumOptions) > 0 {
			result.changes = append(result.changes, "Choices: "+strings.Join(args.EnumOptions, ", "))
		}
	}
	for _, change := range result.changes {
		if len(change) > 4608 {
			return preparedActionPlanCustomization{}, ErrValidation
		}
	}
	return result, nil
}
func (a App) executeApprovedCustomization(ctx context.Context, input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (ActionPlanExecutionResult, error) {
	prepared, err := a.prepareActionPlanCustomization(ctx, input, command)
	if err != nil {
		return ActionPlanExecutionResult{}, err
	}
	transition := ports.ActionPlanStateTransition{PrincipalID: input.Principal.ID, From: actionplan.StateApproved, To: actionplan.StateExecuted, At: a.clock.Now()}
	var record ports.ActionPlanRecord
	var found bool
	if prepared.assetType != nil {
		record, found, err = a.actionPlanCustomizations.ExecuteCreateCustomAssetTypeActionPlan(ctx, input.TenantID, input.InventoryID, input.PlanID, transition, prepared.assetType.Item, prepared.assetType.AuditRecord)
	} else {
		record, found, err = a.actionPlanCustomizations.ExecuteCreateCustomFieldActionPlan(ctx, input.TenantID, input.InventoryID, input.PlanID, transition, prepared.definition.Item, prepared.definition.AuditRecord)
	}
	if err != nil {
		return ActionPlanExecutionResult{}, err
	}
	if !found {
		return ActionPlanExecutionResult{}, ErrNotFound
	}
	if prepared.assetType != nil {
		a.customFieldService.RecordCustomAssetTypeCreated(ctx, prepared.typeInput, prepared.assetType.Item)
	} else {
		a.customFieldService.RecordCustomFieldDefinitionCreated(ctx, prepared.fieldInput, prepared.definition.Item)
	}
	return ActionPlanExecutionResult{Record: record}, nil
}
