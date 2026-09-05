package agentmodel

import (
	"context"
	"strconv"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (w EvaluationWorker) stillOwned(ctx context.Context, ref ports.EvaluationRunReference, expected model.EvaluationRun) (bool, error) {
	current, found, err := w.deps.Runs.EvaluationRun(ctx, ref.TenantID, ref.ID)
	if err != nil || !found {
		return false, err
	}
	value := current.Snapshot()
	previous := expected.Snapshot()
	return value.Version == previous.Version && value.State == model.EvaluationRunRunning && value.LeaseToken == previous.LeaseToken && w.deps.Clock.Now().Before(value.LeaseUntil), nil
}
func (w EvaluationWorker) fail(ctx context.Context, run model.EvaluationRun, token string, code model.EvaluationRunFailureCode) error {
	failed, err := run.Fail(token, code, w.deps.Clock.Now())
	if err != nil {
		return err
	}
	return w.persist(ctx, failed, run.Snapshot().Version)
}
func (w EvaluationWorker) persist(ctx context.Context, run model.EvaluationRun, expected int) error {
	value := run.Snapshot()
	metadata := map[string]string{"state": string(value.State), "completed_cases": strconv.Itoa(len(value.Results)), "total_cases": strconv.Itoa(len(value.Input.Cases))}
	record, ok := audit.NewRecord(audit.ID(w.deps.IDs.NewID()), audit.TenantID(value.Input.TenantID), "", audit.PrincipalID(value.Input.AuthorID), audit.ActionConversationEvaluationRunProgressed, audit.SourceSystem, audit.TargetConversationEvaluationRun, string(value.Input.ID), value.UpdatedAt, "", metadata)
	if !ok {
		return apperrors.ErrValidation
	}
	if err := w.deps.Runs.SaveEvaluationRun(ctx, run, expected, record); err != nil {
		return err
	}
	w.deps.Observer.Record(ctx, ports.Event{Name: ports.EventConversationEvaluationRunProgressed, Message: "evaluation run progressed", Fields: map[string]string{"tenant_id": string(value.Input.TenantID), "run_id": string(value.Input.ID), "state": string(value.State), "completed_cases": strconv.Itoa(len(value.Results))}})
	return nil
}
