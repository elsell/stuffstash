package memory

import (
	"context"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"maps"
)

type evaluationRunKey struct {
	tenant tenant.ID
	id     model.EvaluationRunID
}

func (s *Store) EvaluationRun(ctx context.Context, tenantID tenant.ID, id model.EvaluationRunID) (model.EvaluationRun, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return model.EvaluationRun{}, false, err
	}
	value, found := s.evaluationRuns[evaluationRunKey{tenantID, id}]
	if !found {
		return model.EvaluationRun{}, false, nil
	}
	owned, err := model.RestoreEvaluationRun(value.Snapshot())
	return owned, err == nil, err
}
func (s *Store) SaveEvaluationRun(ctx context.Context, run model.EvaluationRun, expected int, record audit.Record) error {
	value, err := model.RestoreEvaluationRun(run.Snapshot())
	if err != nil {
		return err
	}
	snapshot := value.Snapshot()
	if expected < 0 || snapshot.Version != expected+1 {
		return ports.ErrEvaluationRunConflict
	}
	action := audit.ActionConversationEvaluationRunProgressed
	if expected == 0 {
		action = audit.ActionConversationEvaluationRunCreated
	} else if snapshot.State == model.EvaluationRunCancelled {
		action = audit.ActionConversationEvaluationRunCancelled
	}
	if record.TenantID.String() != string(snapshot.Input.TenantID) || record.InventoryID != "" || record.ID == "" || record.Action != action || record.TargetType != audit.TargetConversationEvaluationRun || record.TargetID != string(snapshot.Input.ID) {
		return model.ErrInvalidEvaluationRun
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	key := evaluationRunKey{tenant.ID(snapshot.Input.TenantID), snapshot.Input.ID}
	if _, exists := s.tenants[key.tenant]; !exists {
		return ports.ErrForbidden
	}
	previous, found := s.evaluationRuns[key]
	if expected == 0 {
		if found || snapshot.State != model.EvaluationRunQueued {
			return ports.ErrEvaluationRunConflict
		}
	} else if !found || previous.Snapshot().Version != expected || !value.IsSuccessorOf(previous) {
		return ports.ErrEvaluationRunConflict
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return ports.ErrEvaluationRunConflict
	}
	if s.evaluationRuns == nil {
		s.evaluationRuns = map[evaluationRunKey]model.EvaluationRun{}
	}
	record.Metadata = maps.Clone(record.Metadata)
	s.auditRecords[record.ID] = record
	s.evaluationRuns[key] = value
	return nil
}
