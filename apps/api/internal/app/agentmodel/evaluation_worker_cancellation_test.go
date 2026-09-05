package agentmodel

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	model "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	fixture "github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
)

type evaluationPulseDelay struct{ pulse chan struct{} }

func (d evaluationPulseDelay) Wait(ctx context.Context, _ time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-d.pulse:
		return nil
	}
}

type evaluationBlockingExecutor struct {
	before    func()
	cancelled bool
}

func (e *evaluationBlockingExecutor) Execute(ctx context.Context, _ ports.ConversationEvaluationInput) (ports.ConversationEvaluationResult, error) {
	e.before()
	<-ctx.Done()
	e.cancelled = true
	return ports.ConversationEvaluationResult{}, ctx.Err()
}

type evaluationRevocableAuthorizer struct {
	allowTenantConfigureAuthorizer
	revoked atomic.Bool
}

func (a *evaluationRevocableAuthorizer) CheckTenant(ctx context.Context, p identity.Principal, permission ports.TenantPermission, id tenant.ID) error {
	if a.revoked.Load() {
		return ports.ErrForbidden
	}
	return a.allowTenantConfigureAuthorizer.CheckTenant(ctx, p, permission, id)
}
func TestEvaluationWorkerCancelsActiveModelOnCancellationOrRevocation(t *testing.T) {
	for _, revoke := range []bool{false, true} {
		deps, store, _, clock := evaluationWorkerSetup(t)
		ref := workerReference(t, store)
		pulse := make(chan struct{}, 1)
		deps.Delay = evaluationPulseDelay{pulse: pulse}
		deps.PollInterval = time.Second
		authorizer := &evaluationRevocableAuthorizer{}
		deps.Authorizer = authorizer
		executor := &evaluationBlockingExecutor{before: func() {
			if revoke {
				authorizer.revoked.Store(true)
			} else {
				run, _, err := store.EvaluationRun(context.Background(), ref.TenantID, ref.ID)
				if err != nil {
					t.Fatal(err)
				}
				cancelled, err := run.Cancel(clock.Now())
				if err != nil {
					t.Fatal(err)
				}
				if err := store.SaveEvaluationRun(context.Background(), cancelled, run.Snapshot().Version, fixture.Record(t, "cancel-active", string(ref.ID), audit.ActionConversationEvaluationRunCancelled)); err != nil {
					t.Fatal(err)
				}
			}
			pulse <- struct{}{}
		}}
		deps.Executor = executor
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := NewEvaluationWorker(deps).Process(ctx, ref)
		cancel()
		if err != nil || !executor.cancelled {
			t.Fatalf("active executor not cancelled: %v", err)
		}
		saved, _, err := store.EvaluationRun(context.Background(), ref.TenantID, ref.ID)
		if err != nil || len(saved.Snapshot().Results) != 0 {
			t.Fatal("cancelled observation persisted")
		}
		if revoke {
			if saved.Snapshot().FailureCode != model.EvaluationRunFailureAccessRevoked {
				t.Fatal("revocation not recorded")
			}
		} else if saved.Snapshot().State != model.EvaluationRunCancelled {
			t.Fatal("cancelled state overwritten")
		}
	}
}
