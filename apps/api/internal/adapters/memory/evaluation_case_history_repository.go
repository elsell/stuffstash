package memory

import (
	"cmp"
	"context"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"slices"
)

func (s *Store) ListEvaluationCaseRevisions(ctx context.Context, scope tenant.ID, caseID domain.EvaluationCaseID, page ports.EvaluationCaseRevisionPageRequest) ([]domain.EvaluationCaseRevision, error) {
	if page.Limit < 1 || page.Limit > ports.MaxEvaluationCasePageLimit || page.AfterNumber < 0 {
		return nil, ports.ErrInvalidEvaluationCasePage
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows := []domain.EvaluationCaseRevision{}
	for key, value := range s.evaluationCaseRevisions {
		if key.tenant == scope && key.id == caseID && value.Snapshot().Number > page.AfterNumber {
			rows = append(rows, value)
		}
	}
	slices.SortFunc(rows, func(a, b domain.EvaluationCaseRevision) int {
		return cmp.Compare(a.Snapshot().Number, b.Snapshot().Number)
	})
	return rows[:min(len(rows), page.Limit)], nil
}
