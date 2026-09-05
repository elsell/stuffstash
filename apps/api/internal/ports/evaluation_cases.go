package ports

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

var ErrEvaluationCaseConflict = errors.New("evaluation case changed")
var ErrInvalidEvaluationCasePage = errors.New("invalid evaluation case page")

const MaxEvaluationCasePageLimit = 100

type EvaluationCaseHeadRecord struct {
	TenantID         tenant.ID
	ID               agentmodel.EvaluationCaseID
	Title            string
	LatestRevision   int
	LatestRevisionID agentmodel.EvaluationCaseRevisionID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
type EvaluationCasePageRequest struct {
	AfterID agentmodel.EvaluationCaseID
	Limit   int
}
type EvaluationCaseRepository interface {
	EvaluationCaseHead(context.Context, tenant.ID, agentmodel.EvaluationCaseID) (EvaluationCaseHeadRecord, bool, error)
	EvaluationCaseRevision(context.Context, tenant.ID, agentmodel.EvaluationCaseID, agentmodel.EvaluationCaseRevisionID) (agentmodel.EvaluationCaseRevision, bool, error)
	ListEvaluationCases(context.Context, tenant.ID, EvaluationCasePageRequest) ([]EvaluationCaseHeadRecord, error)
	AppendEvaluationCaseRevision(context.Context, agentmodel.EvaluationCaseRevision, int, audit.Record) error
}
