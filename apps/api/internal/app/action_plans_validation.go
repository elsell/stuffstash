package app

import (
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func validActionPlanCommandID(value string) bool {
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

func validateExecutableActionPlanArguments(kind actionplan.CommandKind, arguments []byte) error {
	command := ports.ActionPlanCommandRecord{Kind: kind, ArgumentsJSON: arguments}
	switch kind {
	case actionplan.CommandKindCreateAsset, actionplan.CommandKindCreateLocation:
		_, err := parseActionPlanCreateArguments(command)
		return err
	case actionplan.CommandKindMoveAsset:
		_, err := parseActionPlanMoveArguments(command)
		return err
	case actionplan.CommandKindArchiveAsset, actionplan.CommandKindRestoreAsset:
		_, err := parseActionPlanAssetIDOnlyArguments(command)
		return err
	case actionplan.CommandKindCheckoutAsset, actionplan.CommandKindReturnAsset:
		_, err := parseActionPlanCheckoutArguments(command)
		return err
	default:
		return ErrValidation
	}
}

func boundedActionPlanStrings(values []string, maxCount int, maxLength int) ([]string, error) {
	if len(values) > maxCount {
		return nil, ErrValidation
	}
	bounded := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if len(trimmed) > maxLength {
			return nil, ErrValidation
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
		len(record.ConfirmationSummary) > maxActionPlanSummaryLength ||
		record.State != actionplan.StateProposed ||
		record.CreatedAt.IsZero() ||
		record.UpdatedAt.IsZero() ||
		len(record.Commands) == 0 {
		return ErrValidation
	}
	if len(record.IntentSummary) > maxActionPlanSummaryLength || len(record.ModelInterpretationSummary) > maxActionPlanSummaryLength {
		return ErrValidation
	}
	return nil
}
