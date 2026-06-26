package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

const (
	maxActionPlanCommands       = 10
	maxActionPlanSummaryLength  = 500
	maxActionPlanArgumentBytes  = 4096
	maxActionPlanRiskCount      = 10
	maxActionPlanRiskTextLength = 300
)

type CreateActionPlanInput struct {
	Principal                  identity.Principal
	TenantID                   tenant.ID
	InventoryID                inventory.InventoryID
	Source                     string
	RealtimeSessionID          string
	IntentSummary              string
	ModelInterpretationSummary string
	ConfirmationSummary        string
	Commands                   []ActionPlanCommandInput
	Risks                      []string
}

type ActionPlanCommandInput struct {
	Kind      actionplan.CommandKind
	Summary   string
	Arguments map[string]any
}

type ActionPlanDecisionInput struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
	PlanID      string
}

func (a App) CreateActionPlan(ctx context.Context, input CreateActionPlanInput) (ports.ActionPlanRecord, error) {
	if err := a.ensureActionPlanDependencies(); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionView); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	planID := strings.TrimSpace(a.ids.NewID())
	commands, err := a.actionPlanCommands(input.Commands)
	if err != nil {
		return ports.ActionPlanRecord{}, err
	}
	risks, err := boundedActionPlanStrings(input.Risks, maxActionPlanRiskCount, maxActionPlanRiskTextLength)
	if err != nil {
		return ports.ActionPlanRecord{}, err
	}
	now := a.clock.Now()
	record := ports.ActionPlanRecord{
		ID:                         planID,
		TenantID:                   input.TenantID,
		InventoryID:                input.InventoryID,
		PrincipalID:                input.Principal.ID,
		Source:                     strings.TrimSpace(input.Source),
		RealtimeSessionID:          strings.TrimSpace(input.RealtimeSessionID),
		State:                      actionplan.StateProposed,
		IntentSummary:              strings.TrimSpace(input.IntentSummary),
		ModelInterpretationSummary: strings.TrimSpace(input.ModelInterpretationSummary),
		ConfirmationSummary:        strings.TrimSpace(input.ConfirmationSummary),
		Commands:                   commands,
		Risks:                      risks,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}
	if err := validateActionPlanApplicationRecord(record); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	if err := a.actionPlans.SaveActionPlan(ctx, record); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	return record, nil
}

func (a App) ApproveActionPlan(ctx context.Context, input ActionPlanDecisionInput) (ports.ActionPlanRecord, error) {
	return a.transitionActionPlan(ctx, input, actionplan.StateApproved)
}

func (a App) CancelActionPlan(ctx context.Context, input ActionPlanDecisionInput) (ports.ActionPlanRecord, error) {
	return a.transitionActionPlan(ctx, input, actionplan.StateCancelled)
}

func (a App) transitionActionPlan(ctx context.Context, input ActionPlanDecisionInput, to actionplan.State) (ports.ActionPlanRecord, error) {
	if err := a.ensureActionPlanDependencies(); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	permission := ports.InventoryPermissionEditAsset
	if to == actionplan.StateCancelled {
		permission = ports.InventoryPermissionView
	}
	if err := a.ensureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, permission); err != nil {
		return ports.ActionPlanRecord{}, err
	}
	record, found, err := a.actionPlans.UpdateActionPlanState(ctx, input.TenantID, input.InventoryID, strings.TrimSpace(input.PlanID), ports.ActionPlanStateTransition{
		PrincipalID: input.Principal.ID,
		From:        actionplan.StateProposed,
		To:          to,
		At:          a.clock.Now(),
	})
	if err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return ports.ActionPlanRecord{}, ErrConflict
		}
		return ports.ActionPlanRecord{}, err
	}
	if !found {
		return ports.ActionPlanRecord{}, ErrNotFound
	}
	return record, nil
}

func (a App) actionPlanCommands(inputs []ActionPlanCommandInput) ([]ports.ActionPlanCommandRecord, error) {
	if len(inputs) == 0 || len(inputs) > maxActionPlanCommands {
		return nil, ErrValidation
	}
	commands := make([]ports.ActionPlanCommandRecord, 0, len(inputs))
	for _, input := range inputs {
		if !input.Kind.Valid() {
			return nil, ErrValidation
		}
		summary := strings.TrimSpace(input.Summary)
		if summary == "" || len(summary) > maxActionPlanSummaryLength {
			return nil, ErrValidation
		}
		if err := validateSafeActionPlanArguments(input.Arguments); err != nil {
			return nil, err
		}
		arguments, err := json.Marshal(input.Arguments)
		if err != nil || len(arguments) > maxActionPlanArgumentBytes {
			return nil, ErrValidation
		}
		if string(arguments) == "null" {
			arguments = []byte("{}")
		}
		commands = append(commands, ports.ActionPlanCommandRecord{
			ID:            strings.TrimSpace(a.ids.NewID()),
			Kind:          input.Kind,
			Summary:       summary,
			ArgumentsJSON: arguments,
		})
	}
	return commands, nil
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

func (a App) ensureActionPlanDependencies() error {
	if a.actionPlans == nil || a.tenants == nil || a.inventories == nil || a.authorizer == nil || a.ids == nil || a.clock == nil {
		return ErrInvalidInput
	}
	return nil
}
