package ports

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

var ErrWorkflowConflict = errors.New("conversation workflow changed")
var ErrWorkflowNotFound = errors.New("conversation workflow not found")

type WorkflowHeadRecord struct {
	Name             string
	LatestRevisionID agentmodel.WorkflowRevisionID
	TenantID         tenant.ID
	ID               agentmodel.WorkflowID
	LatestRevision   int
	ActiveRevisionID agentmodel.WorkflowRevisionID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type WorkflowSelectionReference struct {
	WorkflowID agentmodel.WorkflowID
	RevisionID agentmodel.WorkflowRevisionID
}

type ConversationWorkflowRepository interface {
	SelectedWorkflowRevision(context.Context, tenant.ID) (WorkflowSelectionReference, bool, error)
	WorkflowHead(context.Context, tenant.ID, agentmodel.WorkflowID) (WorkflowHeadRecord, bool, error)
	WorkflowRevision(context.Context, tenant.ID, agentmodel.WorkflowID, agentmodel.WorkflowRevisionID) (agentmodel.WorkflowRevision, bool, error)
	AppendWorkflowRevision(context.Context, agentmodel.WorkflowRevision, int, audit.Record) error
	ActivateWorkflowRevision(context.Context, tenant.ID, agentmodel.WorkflowID, agentmodel.WorkflowRevisionID, WorkflowSelectionReference, time.Time, audit.Record) error
}
