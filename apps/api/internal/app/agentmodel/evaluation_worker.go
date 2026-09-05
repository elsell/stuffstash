package agentmodel

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type EvaluationWorkerDependencies struct {
	Runs       ports.EvaluationRunRepository
	Authorizer ports.Authorizer
	Providers  ports.EvaluationRunProviderResolver
	Executor   ports.ConversationEvaluationExecutor
	IDs        ports.IDGenerator
	Clock      ports.Clock
	Observer   ports.Observer
	LeaseGrace time.Duration
}
type EvaluationWorker struct{ deps EvaluationWorkerDependencies }

func NewEvaluationWorker(deps EvaluationWorkerDependencies) EvaluationWorker {
	if deps.Observer == nil {
		deps.Observer = noopObserver{}
	}
	return EvaluationWorker{deps: deps}
}
func (w EvaluationWorker) Process(ctx context.Context, ref ports.EvaluationRunReference) error {
	err := w.process(ctx, ref)
	if errors.Is(err, ports.ErrEvaluationRunConflict) {
		return nil
	}
	return err
}
func (w EvaluationWorker) process(ctx context.Context, ref ports.EvaluationRunReference) error {
	if w.deps.Runs == nil || w.deps.Authorizer == nil || w.deps.Providers == nil || w.deps.Executor == nil || w.deps.IDs == nil || w.deps.Clock == nil || w.deps.LeaseGrace <= 0 {
		return apperrors.ErrPrecondition
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	run, found, err := w.deps.Runs.EvaluationRun(ctx, ref.TenantID, ref.ID)
	if err != nil || !found {
		return err
	}
	initial := run.Snapshot()
	if initial.State != model.EvaluationRunQueued && initial.State != model.EvaluationRunRunning {
		return nil
	}
	now := w.deps.Clock.Now()
	if initial.State == model.EvaluationRunRunning {
		if now.Before(initial.LeaseUntil) {
			return nil
		}
		if initial.Attempts >= initial.Input.MaxAttempts {
			failed, err := run.FailExpired(now)
			if err != nil {
				return err
			}
			return w.persist(ctx, failed, initial.Version)
		}
	}
	const maxDuration = time.Duration(1<<63 - 1)
	seconds := initial.Input.Workflow.Snapshot().Definition.Settings().Budget.ElapsedSeconds
	if seconds <= 0 || uint64(seconds) > uint64(maxDuration/time.Second) {
		return apperrors.ErrPrecondition
	}
	budget := time.Duration(seconds) * time.Second
	if w.deps.LeaseGrace > maxDuration-budget {
		return apperrors.ErrPrecondition
	}
	lease := budget + w.deps.LeaseGrace
	token := w.deps.IDs.NewID()
	claimed, err := run.Claim(token, now, lease)
	if err != nil {
		return err
	}
	if err := w.persist(ctx, claimed, initial.Version); err != nil {
		return err
	}
	run = claimed
	principal := identity.Principal{ID: identity.PrincipalID(initial.Input.AuthorID)}
	if err := w.deps.Authorizer.CheckTenant(ctx, principal, ports.TenantPermissionConfigure, ref.TenantID); err != nil {
		if errors.Is(err, ports.ErrForbidden) || errors.Is(err, ports.ErrUnauthenticated) {
			return w.fail(ctx, run, token, model.EvaluationRunFailureAccessRevoked)
		}
		return err
	}
	providers, err := w.deps.Providers.ResolveEvaluationRunProviders(ctx, ref.TenantID, run)
	if err != nil {
		if errors.Is(err, ports.ErrEvaluationConfigurationChanged) {
			return w.fail(ctx, run, token, model.EvaluationRunFailureConfigurationChanged)
		}
		return err
	}
	for len(run.Snapshot().Results) < len(initial.Input.Cases) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if current, err := w.stillOwned(ctx, ref, run); err != nil || !current {
			return err
		}
		if err := w.deps.Authorizer.CheckTenant(ctx, principal, ports.TenantPermissionConfigure, ref.TenantID); err != nil {
			if errors.Is(err, ports.ErrForbidden) || errors.Is(err, ports.ErrUnauthenticated) {
				return w.fail(ctx, run, token, model.EvaluationRunFailureAccessRevoked)
			}
			return err
		}
		now = w.deps.Clock.Now()
		if now.Add(lease).After(run.Snapshot().LeaseUntil) {
			renewed, err := run.Renew(token, now, lease)
			if err != nil {
				return err
			}
			if err := w.persist(ctx, renewed, run.Snapshot().Version); err != nil {
				return err
			}
			run = renewed
		}
		index := len(run.Snapshot().Results)
		caseCtx, cancel := context.WithTimeout(ctx, budget)
		result, executionErr := w.deps.Executor.Execute(caseCtx, ports.ConversationEvaluationInput{Case: initial.Input.Cases[index].Snapshot().Definition, Revision: initial.Input.Workflow, Limits: initial.Input.Limits, Principal: principal, Providers: providers.Providers, WorkflowProviders: providers.WorkflowProviders})
		deadlineErr := caseCtx.Err()
		cancel()
		if err := ctx.Err(); err != nil {
			return err
		}
		if current, err := w.stillOwned(ctx, ref, run); err != nil || !current {
			return err
		}
		if executionErr != nil || deadlineErr != nil || result.Coverage != ports.EvaluationCoverageText {
			return w.fail(ctx, run, token, model.EvaluationRunFailureExecution)
		}
		next, err := run.RecordCase(token, index, result.Outcome, result.ModelCalls, result.Duration, w.deps.Clock.Now())
		if err != nil {
			return w.fail(ctx, run, token, model.EvaluationRunFailureExecution)
		}
		if err := w.persist(ctx, next, run.Snapshot().Version); err != nil {
			return err
		}
		run = next
	}
	return nil
}
