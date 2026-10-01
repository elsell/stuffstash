package agentmodel

import (
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func ValidActionPlanCommandID(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxActionPlanCommandIDLength {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			continue
		}
		return false
	}
	return true
}

func ValidateExecutableActionPlanArguments(kind actionplan.CommandKind, arguments []byte) error {
	command := ports.ActionPlanCommandRecord{Kind: kind, ArgumentsJSON: arguments}
	switch kind {
	case actionplan.CommandKindCreateCustomAssetType, actionplan.CommandKindCreateCustomFieldDefinition:
		_, err := parseActionPlanCustomizationArguments(command)
		return err
	case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation:
		_, err := ParseActionPlanCreateArguments(command)
		return err
	case actionplan.CommandKindUpdateAsset:
		_, err := ParseActionPlanUpdateArguments(command)
		return err
	case actionplan.CommandKindMoveAsset:
		_, err := ParseActionPlanMoveArguments(command)
		return err
	case actionplan.CommandKindArchiveAsset, actionplan.CommandKindRestoreAsset:
		_, err := ParseActionPlanAssetIDOnlyArguments(command)
		return err
	case actionplan.CommandKindCheckoutAsset, actionplan.CommandKindReturnAsset:
		_, err := ParseActionPlanCheckoutArguments(command)
		return err
	default:
		return apperrors.ErrValidation
	}
}

func boundedActionPlanStrings(values []string, maxCount int, maxLength int) ([]string, error) {
	if len(values) > maxCount {
		return nil, apperrors.ErrValidation
	}
	bounded := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if len(trimmed) > maxLength {
			return nil, apperrors.ErrValidation
		}
		bounded = append(bounded, trimmed)
	}
	return bounded, nil
}

func validateActionPlanApplicationRecord(record ports.ActionPlanRecord) error {
	if strings.TrimSpace(record.ID) == "" ||
		record.TenantID.String() == "" ||
		record.InventoryID.String() == "" ||
		record.PrincipalID.String() == "" ||
		strings.TrimSpace(record.Source) == "" ||
		strings.TrimSpace(record.ConfirmationSummary) == "" ||
		len(record.ConfirmationSummary) > MaxActionPlanSummaryLength ||
		record.State != actionplan.StateProposed ||
		record.CreatedAt.IsZero() ||
		record.UpdatedAt.IsZero() ||
		len(record.Commands) == 0 {
		return apperrors.ErrValidation
	}
	if len(record.IntentSummary) > MaxActionPlanSummaryLength || len(record.ModelInterpretationSummary) > MaxActionPlanSummaryLength {
		return apperrors.ErrValidation
	}
	return nil
}
