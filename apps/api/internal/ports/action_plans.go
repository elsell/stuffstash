package ports

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/actionplan"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

type ActionPlanCommandRecord struct {
	ID            string
	Kind          actionplan.CommandKind
	Summary       string
	ArgumentsJSON []byte
}

type ActionPlanRecord struct {
	ID                         string
	TenantID                   tenant.ID
	InventoryID                inventory.InventoryID
	PrincipalID                identity.PrincipalID
	Source                     string
	RealtimeSessionID          string
	State                      actionplan.State
	IntentSummary              string
	ModelInterpretationSummary string
	ConfirmationSummary        string
	Commands                   []ActionPlanCommandRecord
	Risks                      []string
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
	ApprovedAt                 time.Time
	CancelledAt                time.Time
	ExecutedAt                 time.Time
	FailedAt                   time.Time
}

type ActionPlanStateTransition struct {
	PrincipalID identity.PrincipalID
	From        actionplan.State
	To          actionplan.State
	At          time.Time
}

type ActionPlanRepository interface {
	SaveActionPlan(ctx context.Context, record ActionPlanRecord) error
	ActionPlanByID(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string) (ActionPlanRecord, bool, error)
	UpdateActionPlanState(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ActionPlanStateTransition) (ActionPlanRecord, bool, error)
	ExecuteCreateAssetActionPlan(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, planID string, transition ActionPlanStateTransition, item asset.Asset, auditRecord audit.Record, undoableOperation *UndoableOperation) (ActionPlanRecord, bool, error)
}
