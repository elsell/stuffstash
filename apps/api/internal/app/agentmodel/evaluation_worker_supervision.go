package agentmodel

import (
	"context"
	"time"

	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (w EvaluationWorker) executeSupervised(ctx context.Context, ref ports.EvaluationRunReference, run model.EvaluationRun, input ports.ConversationEvaluationInput, budget time.Duration) (ports.ConversationEvaluationResult, error, error) {
	caseCtx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	finished := make(chan error, 1)
	go func() {
		err := w.superviseCase(caseCtx, ref, run, input)
		if err != nil {
			cancel()
		}
		finished <- err
	}()
	result, executionErr := w.deps.Executor.Execute(caseCtx, input)
	if executionErr == nil {
		executionErr = caseCtx.Err()
	}
	cancel()
	supervisionErr := <-finished
	return result, executionErr, supervisionErr
}
func (w EvaluationWorker) superviseCase(ctx context.Context, ref ports.EvaluationRunReference, run model.EvaluationRun, input ports.ConversationEvaluationInput) error {
	for {
		if err := w.deps.Delay.Wait(ctx, w.deps.PollInterval); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		current, err := w.stillOwned(ctx, ref, run)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			return err
		}
		if !current {
			return ports.ErrEvaluationRunConflict
		}
		err = w.deps.Authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, ref.TenantID)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
