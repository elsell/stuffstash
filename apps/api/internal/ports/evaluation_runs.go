package ports

import (
	"context"
	"errors"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"time"
)

var ErrEvaluationRunConflict = errors.New("evaluation run changed")
var ErrInvalidEvaluationRunPage = errors.New("invalid evaluation run page")

const MaxEvaluationRunPageLimit = 100

type EvaluationRunReference struct {
	TenantID tenant.ID
	ID       model.EvaluationRunID
}
type EvaluationRunHead struct {
	EvaluationRunReference
	State          model.EvaluationRunState
	Version        int
	WorkflowID     model.WorkflowID
	RevisionID     model.WorkflowRevisionID
	TotalCases     int
	CompletedCases int
	PassedCases    int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
type EvaluationRunPageRequest struct {
	AfterID model.EvaluationRunID
	Limit   int
}
type EvaluationRunRepository interface {
	EvaluationRun(context.Context, tenant.ID, model.EvaluationRunID) (model.EvaluationRun, bool, error)
	SaveEvaluationRun(context.Context, model.EvaluationRun, int, audit.Record) error
	ListEvaluationRuns(context.Context, tenant.ID, EvaluationRunPageRequest) ([]EvaluationRunHead, error)
	RunnableEvaluationRuns(context.Context, time.Time, int) ([]EvaluationRunReference, error)
}
