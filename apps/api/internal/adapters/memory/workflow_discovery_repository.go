package memory

import (
	"cmp"
	"context"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"slices"
)

func (s *Store) ListWorkflowHeads(ctx context.Context, scope tenant.ID, page ports.WorkflowHeadPageRequest) ([]ports.WorkflowHeadRecord, error) {
	if page.Limit < 1 || page.Limit > ports.MaxWorkflowPageLimit {
		return nil, ports.ErrInvalidWorkflowPage
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows := []ports.WorkflowHeadRecord{}
	for key, head := range s.workflowHeads {
		if key.tenant == scope && key.workflow > page.AfterID {
			rows = append(rows, head)
		}
	}
	slices.SortFunc(rows, func(a, b ports.WorkflowHeadRecord) int { return cmp.Compare(a.ID, b.ID) })
	return rows[:min(len(rows), page.Limit)], nil
}
func (s *Store) ListWorkflowRevisions(ctx context.Context, scope tenant.ID, workflow model.WorkflowID, page ports.WorkflowRevisionPageRequest) ([]model.WorkflowRevision, error) {
	if page.Limit < 1 || page.Limit > ports.MaxWorkflowPageLimit || page.AfterNumber < 0 {
		return nil, ports.ErrInvalidWorkflowPage
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows := []model.WorkflowRevision{}
	for key, revision := range s.workflowRevisions {
		if key.tenant == scope && key.workflow == workflow && revision.Snapshot().Number > page.AfterNumber {
			rows = append(rows, revision)
		}
	}
	slices.SortFunc(rows, func(a, b model.WorkflowRevision) int { return cmp.Compare(a.Snapshot().Number, b.Snapshot().Number) })
	return rows[:min(len(rows), page.Limit)], nil
}
