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
	TenantID         tenant.ID
	ID               agentmodel.WorkflowID
	LatestRevision   int
	ActiveRevisionID agentmodel.WorkflowRevisionID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ConversationWorkflowRepository interface {
	WorkflowHead(context.Context, tenant.ID, agentmodel.WorkflowID) (WorkflowHeadRecord, bool, error)
	WorkflowRevision(context.Context, tenant.ID, agentmodel.WorkflowID, agentmodel.WorkflowRevisionID) (agentmodel.WorkflowRevision, bool, error)
	AppendWorkflowRevision(context.Context, agentmodel.WorkflowRevision, int, audit.Record) error
	ActivateWorkflowRevision(context.Context, tenant.ID, agentmodel.WorkflowID, agentmodel.WorkflowRevisionID, agentmodel.WorkflowRevisionID, time.Time, audit.Record) error
}
