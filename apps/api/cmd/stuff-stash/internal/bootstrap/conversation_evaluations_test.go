package bootstrap

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/config"
	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	fixture "github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
)

func TestBuildApplicationRejectsInvalidEvaluationConfiguration(t *testing.T) {
	t.Setenv("STUFF_STASH_EVALUATION_CONCURRENCY", "9")
	_, err := buildApplication(context.Background(), config.Load(), nil, nil, nil, repositories{})
	if err == nil || !strings.Contains(err.Error(), "STUFF_STASH_EVALUATION_CONCURRENCY") {
		t.Fatalf("invalid worker config not rejected first: %v", err)
	}
}

type evaluationRuntimeDrainer struct {
	started chan struct{}
	active  atomic.Int32
	calls   atomic.Int32
}

func (d *evaluationRuntimeDrainer) DrainEvaluationRuns(ctx context.Context, limit, concurrency int) error {
	d.active.Add(1)
	defer d.active.Add(-1)
	d.calls.Add(1)
	d.started <- struct{}{}
	<-ctx.Done()
	return ctx.Err()
}
func TestEvaluationSchedulerStopsAndJoinsActiveDrain(t *testing.T) {
	drainer := &evaluationRuntimeDrainer{started: make(chan struct{}, 1)}
	stop, err := startEvaluationWorker(context.Background(), drainer, nil, config.Load())
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-drainer.started:
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not start")
	}
	stop()
	if drainer.active.Load() != 0 || drainer.calls.Load() != 1 {
		t.Fatal("scheduler returned before drain stopped")
	}
}
func TestEvaluationSchedulerCanBeDisabled(t *testing.T) {
	t.Setenv("STUFF_STASH_EVALUATION_WORKER_ENABLED", "false")
	drainer := &evaluationRuntimeDrainer{started: make(chan struct{}, 1)}
	stop, err := startEvaluationWorker(context.Background(), drainer, nil, config.Load())
	if err != nil {
		t.Fatal(err)
	}
	stop()
	if drainer.calls.Load() != 0 {
		t.Fatal("disabled scheduler drained")
	}
}

func TestEvaluationRuntimeUsesConfiguredAuthorizationAndRunStorage(t *testing.T) {
	ctx := context.Background()
	cfg := config.Load()
	cfg.RepositoryMode = "memory"
	repos, closeStore, err := buildRepositories(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer closeStore()
	name, _ := tenant.NewName("Evaluation home")
	if err := repos.tenantUnitOfWork.SaveTenant(ctx, tenant.Tenant{ID: fixture.TenantID, Name: name}); err != nil {
		t.Fatal(err)
	}
	if err := repos.evaluationRuns.SaveEvaluationRun(ctx, fixture.Run(t, "runtime"), 0, fixture.Record(t, "runtime-created", "runtime", audit.ActionConversationEvaluationRunCreated)); err != nil {
		t.Fatal(err)
	}
	settings, err := cfg.ConversationEvaluations.Settings()
	if err != nil {
		t.Fatal(err)
	}
	limits, err := cfg.ConversationWorkflows.Limits()
	if err != nil {
		t.Fatal(err)
	}
	services := buildEvaluationRuntime(cfg, settings, limits, nil, memory.NewAuthorizer(), repos, nil)
	if err := services.worker.Drain(ctx, 1, 1); err != nil {
		t.Fatal(err)
	}
	run, found, err := repos.evaluationRuns.EvaluationRun(ctx, fixture.TenantID, "runtime")
	if err != nil || !found || run.Snapshot().FailureCode != agentmodel.EvaluationRunFailureAccessRevoked {
		t.Fatal("runtime did not enforce configured authorization")
	}
}
