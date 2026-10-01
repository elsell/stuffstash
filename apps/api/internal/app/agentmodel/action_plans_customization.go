package agentmodel

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ActionPlanCustomizationArguments struct {
	Key                string   `json:"key"`
	DisplayName        string   `json:"displayName"`
	Description        string   `json:"description,omitempty"`
	ExpirationEnabled  bool     `json:"expirationEnabled,omitempty"`
	FieldType          string   `json:"fieldType,omitempty"`
	Applicability      string   `json:"applicability,omitempty"`
	EnumOptions        []string `json:"enumOptions,omitempty"`
	CustomAssetTypeIDs []string `json:"customAssetTypeIds,omitempty"`
}

func IsCustomizationCommand(kind actionplan.CommandKind) bool {
	return kind == actionplan.CommandKindCreateCustomAssetType || kind == actionplan.CommandKindCreateCustomFieldDefinition
}
func parseActionPlanCustomizationArguments(command ports.ActionPlanCommandRecord) (ActionPlanCustomizationArguments, error) {
	var args ActionPlanCustomizationArguments
	decoder := json.NewDecoder(bytes.NewReader(command.ArgumentsJSON))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&args) != nil {
		return args, apperrors.ErrValidation
	}
	var raw map[string]json.RawMessage
	if json.Unmarshal(command.ArgumentsJSON, &raw) != nil {
		return args, apperrors.ErrValidation
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
			return args, apperrors.ErrValidation
		}
		if _, ok := customfield.NewApplicability(args.Applicability); !ok {
			return args, apperrors.ErrValidation
		}
	} else {
		return args, apperrors.ErrValidation
	}
	for key, value := range raw {
		if !allowed[key] || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return args, apperrors.ErrValidation
		}
	}
	if key, ok := customfield.NewKey(args.Key); !ok || key.String() != args.Key {
		return args, apperrors.ErrValidation
	}
	if _, ok := customfield.NewDisplayName(args.DisplayName); !ok {
		return args, apperrors.ErrValidation
	}
	if len(args.CustomAssetTypeIDs) > 10 || len(args.EnumOptions) > 50 {
		return args, apperrors.ErrValidation
	}
	return args, nil
}

type PreparedActionPlanCustomization struct {
	AssetType  *ports.PreparedCustomAssetType
	Definition *ports.PreparedCustomFieldDefinition
	TypeInput  ports.CreateCustomAssetTypeInput
	FieldInput ports.CreateCustomFieldDefinitionInput
	Changes    []string
}

func (a ActionPlanService) PrepareActionPlanCustomization(ctx context.Context, input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (PreparedActionPlanCustomization, error) {
	var result PreparedActionPlanCustomization
	if a.deps.ActionPlanCustomizations == nil {
		return result, apperrors.ErrValidation
	}
	args, err := parseActionPlanCustomizationArguments(command)
	if err != nil {
		return result, err
	}
	if command.Kind == actionplan.CommandKindCreateCustomAssetType {
		result.TypeInput = ports.CreateCustomAssetTypeInput{Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID, Source: audit.SourceConversation, RequestID: command.ID, Key: args.Key, DisplayName: args.DisplayName, Description: args.Description, ExpirationEnabled: args.ExpirationEnabled}
		prepared, err := a.deps.CustomizationPreparation.PrepareInventoryCustomAssetType(ctx, result.TypeInput)
		if err != nil {
			return result, err
		}
		result.AssetType = &prepared
		tracking := "Off"
		if args.ExpirationEnabled {
			tracking = "On"
		}
		result.Changes = []string{"Type name: " + prepared.Item.DisplayName.String(), "Key: " + args.Key, "Expiration tracking: " + tracking}
		if prepared.Item.Description.String() != "" {
			result.Changes = append(result.Changes, "Description: "+prepared.Item.Description.String())
		}
	} else {
		result.FieldInput = ports.CreateCustomFieldDefinitionInput{Principal: input.Principal, TenantID: input.TenantID, InventoryID: input.InventoryID, Source: audit.SourceConversation, RequestID: command.ID, Key: args.Key, DisplayName: args.DisplayName, Type: args.FieldType, Applicability: args.Applicability, EnumOptions: args.EnumOptions, CustomAssetTypeIDs: args.CustomAssetTypeIDs}
		prepared, err := a.deps.CustomizationPreparation.PrepareInventoryCustomFieldDefinition(ctx, result.FieldInput)
		if err != nil {
			return result, err
		}
		result.Definition = &prepared
		names := []string{}
		for _, id := range prepared.Item.CustomAssetTypeIDs {
			item, found, err := a.deps.CustomAssetTypes.CustomAssetTypeByID(ctx, input.TenantID, input.InventoryID, id)
			if err != nil {
				return result, err
			}
			if !found {
				return result, apperrors.ErrValidation
			}
			names = append(names, item.DisplayName.String())
		}
		target := "All assets"
		if len(names) > 0 {
			target = strings.Join(names, ", ")
		}
		result.Changes = []string{"Field name: " + prepared.Item.DisplayName.String(), "Key: " + args.Key, "Field type: " + strings.ReplaceAll(args.FieldType, "_", " "), "Applies to: " + target}
		if len(args.EnumOptions) > 0 {
			result.Changes = append(result.Changes, "Choices: "+strings.Join(args.EnumOptions, ", "))
		}
	}
	for _, change := range result.Changes {
		if len(change) > 4608 {
			return PreparedActionPlanCustomization{}, apperrors.ErrValidation
		}
	}
	return result, nil
}
func (a ActionPlanService) executeApprovedCustomization(ctx context.Context, input ActionPlanDecisionInput, command ports.ActionPlanCommandRecord) (ActionPlanExecutionResult, error) {
	prepared, err := a.PrepareActionPlanCustomization(ctx, input, command)
	if err != nil {
		return ActionPlanExecutionResult{}, err
	}
	transition := ports.ActionPlanStateTransition{PrincipalID: input.Principal.ID, From: actionplan.StateApproved, To: actionplan.StateExecuted, At: a.deps.Clock.Now()}
	var record ports.ActionPlanRecord
	var found bool
	if prepared.AssetType != nil {
		record, found, err = a.deps.ActionPlanCustomizations.ExecuteCreateCustomAssetTypeActionPlan(ctx, input.TenantID, input.InventoryID, input.PlanID, transition, prepared.AssetType.Item, prepared.AssetType.AuditRecord)
	} else {
		record, found, err = a.deps.ActionPlanCustomizations.ExecuteCreateCustomFieldActionPlan(ctx, input.TenantID, input.InventoryID, input.PlanID, transition, prepared.Definition.Item, prepared.Definition.AuditRecord)
	}
	if err != nil {
		return ActionPlanExecutionResult{}, err
	}
	if !found {
		return ActionPlanExecutionResult{}, apperrors.ErrNotFound
	}
	if prepared.AssetType != nil {
		a.deps.CustomizationPreparation.RecordCustomAssetTypeCreated(ctx, prepared.TypeInput, prepared.AssetType.Item)
	} else {
		a.deps.CustomizationPreparation.RecordCustomFieldDefinitionCreated(ctx, prepared.FieldInput, prepared.Definition.Item)
	}
	return ActionPlanExecutionResult{Record: record}, nil
}
