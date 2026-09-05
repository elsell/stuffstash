package memory

import (
	"context"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"slices"
	"strings"
	"time"
)

func (s *Store) ListEvaluationRuns(ctx context.Context, tenantID tenant.ID, page ports.EvaluationRunPageRequest) ([]ports.EvaluationRunHead, error) {
	if page.Limit < 1 || page.Limit > ports.MaxEvaluationRunPageLimit {
		return nil, ports.ErrInvalidEvaluationRunPage
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows := []ports.EvaluationRunHead{}
	for key, run := range s.evaluationRuns {
		if key.tenant == tenantID && key.id > page.AfterID {
			rows = append(rows, memoryEvaluationRunHead(run))
		}
	}
	slices.SortFunc(rows, func(a, b ports.EvaluationRunHead) int { return strings.Compare(string(a.ID), string(b.ID)) })
	if len(rows) > page.Limit {
		rows = rows[:page.Limit]
	}
	return rows, nil
}
func (s *Store) RunnableEvaluationRuns(ctx context.Context, now time.Time, limit int) ([]ports.EvaluationRunReference, error) {
	if now.IsZero() || limit < 1 || limit > ports.MaxEvaluationRunPageLimit {
		return nil, ports.ErrInvalidEvaluationRunPage
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows := []ports.EvaluationRunHead{}
	for _, run := range s.evaluationRuns {
		value := run.Snapshot()
		if value.State == model.EvaluationRunQueued || (value.State == model.EvaluationRunRunning && !now.Before(value.LeaseUntil)) {
			rows = append(rows, memoryEvaluationRunHead(run))
		}
	}
	slices.SortFunc(rows, func(a, b ports.EvaluationRunHead) int {
		if result := a.CreatedAt.Compare(b.CreatedAt); result != 0 {
			return result
		}
		if result := strings.Compare(a.TenantID.String(), b.TenantID.String()); result != 0 {
			return result
		}
		return strings.Compare(string(a.ID), string(b.ID))
	})
	if len(rows) > limit {
		rows = rows[:limit]
	}
	result := make([]ports.EvaluationRunReference, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.EvaluationRunReference)
	}
	return result, nil
}
func memoryEvaluationRunHead(run model.EvaluationRun) ports.EvaluationRunHead {
	value := run.Snapshot()
	workflow := value.Input.Workflow.Snapshot()
	head := ports.EvaluationRunHead{EvaluationRunReference: ports.EvaluationRunReference{TenantID: tenant.ID(value.Input.TenantID), ID: value.Input.ID}, State: value.State, Version: value.Version, WorkflowID: workflow.WorkflowID, RevisionID: workflow.ID, TotalCases: len(value.Input.Cases), CompletedCases: len(value.Results), CreatedAt: value.Input.CreatedAt, UpdatedAt: value.UpdatedAt}
	for _, result := range value.Results {
		if result.Verdict.Passed {
			head.PassedCases++
		}
	}
	return head
}
