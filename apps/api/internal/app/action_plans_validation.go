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

func validateSafeActionPlanArguments(arguments any) error {
	if arguments == nil {
		return nil
	}
	switch value := arguments.(type) {
	case map[string]any:
		for key, nested := range value {
			if unsafeActionPlanArgumentKey(key) {
				return ErrValidation
			}
			if err := validateSafeActionPlanArguments(nested); err != nil {
				return err
			}
		}
	case []any:
		for _, nested := range value {
			if err := validateSafeActionPlanArguments(nested); err != nil {
				return err
			}
		}
	case string:
		if unsafeActionPlanArgumentString(value) {
			return ErrValidation
		}
	case bool, float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return nil
	default:
		return ErrValidation
	}
	return nil
}

func unsafeActionPlanArgumentKey(key string) bool {
	normalized := normalizeActionPlanSafetyText(key)
	unsafeTokens := []string{
		"audio",
		"approval",
		"approved",
		"apikey",
		"bearer",
		"credential",
		"generatedspeech",
		"modelresponse",
		"password",
		"prompt",
		"providerid",
		"providerresponse",
		"providersessionid",
		"secret",
		"sessiontoken",
		"token",
		"transcript",
	}
	for _, token := range unsafeTokens {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func unsafeActionPlanArgumentString(value string) bool {
	normalized := normalizeActionPlanSafetyText(value)
	unsafePhrases := []string{
		"apikey",
		"bearer",
		"beginprivatekey",
		"credential",
		"modelresponse",
		"providerresponse",
		"rawprompt",
		"systemprompt",
	}
	for _, phrase := range unsafePhrases {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	return false
}

func normalizeActionPlanSafetyText(value string) string {
	replacer := strings.NewReplacer("_", "", "-", "", " ", "", ".", "", ":", "")
	return replacer.Replace(strings.ToLower(strings.TrimSpace(value)))
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
