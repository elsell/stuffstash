package agentmodel

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/ports"
)

// Changes are application-authored disclosure; the model's summary is not an
// authority for which values will change. Values remain plain text in clients.
func (a ActionPlanService) actionPlanDetailChanges(ctx context.Context, session ActionPlanDecisionInput, title, description *string, fields map[string]any) ([]string, error) {
	changes := []string{}
	if title != nil {
		changes = append(changes, "Name: "+strings.TrimSpace(*title))
	}
	if description != nil {
		value := strings.TrimSpace(*description)
		if value == "" {
			value = "Clear value"
		}
		changes = append(changes, "Description: "+value)
	}
	if len(fields) > 0 {
		if a.deps.CustomFields == nil {
			return nil, ports.ErrInvalidProviderInput
		}
		definitions, err := a.deps.CustomFields.ListEffectiveCustomFieldDefinitions(ctx, session.TenantID, session.InventoryID)
		if err != nil {
			return nil, err
		}
		labels := map[string]string{}
		for _, definition := range definitions {
			if definition.IsActive() {
				labels[definition.Key.String()] = definition.DisplayName.String()
			}
		}
		keys := make([]string, 0, len(fields))
		for key := range fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			label, ok := labels[key]
			if !ok {
				return nil, ports.ErrInvalidProviderInput
			}
			value := "Clear value"
			if fields[key] != nil {
				value = fmt.Sprint(fields[key])
			}
			changes = append(changes, label+": "+value)
		}
	}
	for _, change := range changes {
		if len(change) > 4608 {
			return nil, ports.ErrInvalidProviderInput
		}
	}
	return changes, nil
}
