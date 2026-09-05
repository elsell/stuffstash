package ports

import (
	"context"
	"errors"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

const MaxWorkflowPageLimit = 100

var ErrInvalidWorkflowPage = errors.New("invalid workflow page")

type WorkflowHeadPageRequest struct {
	AfterID model.WorkflowID
	Limit   int
}
type WorkflowRevisionPageRequest struct {
	AfterNumber int
	Limit       int
}
type WorkflowDiscoveryRepository interface {
	ListWorkflowHeads(context.Context, tenant.ID, WorkflowHeadPageRequest) ([]WorkflowHeadRecord, error)
	ListWorkflowRevisions(context.Context, tenant.ID, model.WorkflowID, WorkflowRevisionPageRequest) ([]model.WorkflowRevision, error)
}
