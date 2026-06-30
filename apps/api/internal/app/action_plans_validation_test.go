package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func TestCreateActionPlanRejectsUnsupportedExecutableCommandArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		kind      actionplan.CommandKind
		arguments map[string]any
	}{
		{
			name: "create parent title",
			kind: actionplan.CommandKindCreateAsset,
			arguments: map[string]any{
				"title":       "Apple TV remote",
				"kind":        "item",
				"parentTitle": "Living room",
			},
		},
		{
			name: "create location title",
			kind: actionplan.CommandKindCreateAsset,
			arguments: map[string]any{
				"title":         "Apple TV remote",
				"kind":          "item",
				"locationTitle": "Living room",
			},
		},
		{
			name: "create location with container kind",
			kind: actionplan.CommandKindCreateLocation,
			arguments: map[string]any{
				"title": "Box under the TV",
				"kind":  "container",
			},
		},
		{
			name: "move parent title",
			kind: actionplan.CommandKindMoveAsset,
			arguments: map[string]any{
				"assetId":     "asset-1",
				"parentTitle": "Living room",
			},
		},
		{
			name: "archive title",
			kind: actionplan.CommandKindArchiveAsset,
			arguments: map[string]any{
				"title": "Apple TV remote",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			application := newActionPlanTestApp(&fakeActionPlanRepository{}, &fakeIDGenerator{ids: []string{"plan-1", "command-1"}}, nil)
			_, err := application.CreateActionPlan(context.Background(), CreateActionPlanInput{
				Principal:           identity.Principal{ID: identity.PrincipalID("user-1")},
				TenantID:            tenant.ID("tenant-home"),
				InventoryID:         inventory.InventoryID("inventory-home"),
				Source:              "mobile_voice",
				ConfirmationSummary: "Apply change?",
				Commands: []ActionPlanCommandInput{{
					Kind:      tt.kind,
					Summary:   "Apply change",
					Arguments: tt.arguments,
				}},
			})
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}
