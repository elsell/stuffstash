package importworker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestInProcessWorkerRecordsBackgroundExecutionFailure(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	job := importjob.NewPreviewedRecord(
		importjob.ID("job-one"),
		importjob.TenantID("tenant-one"),
		importjob.InventoryID("inventory-one"),
		importjob.PrincipalID("owner"),
		importjob.SourceRef{Type: importjob.SourceTypeLegacyHomebox, Name: "Homebox", Fingerprint: "sha256:test"},
		importjob.Counts{},
		nil,
		time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC),
	)
	job.Status = importjob.StatusRunning
	if err := store.SaveImportJob(ctx, job); err != nil {
		t.Fatalf("save import job: %v", err)
	}
	repository := failingImportJobRepository{Store: store, err: errors.New("terminal update failed")}
	observer := &recordingObserver{seen: make(chan ports.Event, 1)}
	application := app.New(app.Dependencies{
		Observer:   observer,
		ImportJobs: repository,
	})
	worker := NewInProcess(application, observer)

	if _, err := worker.ExecuteImportJob(ctx, ports.ImportJobCommand{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		RequestID:   "start-request",
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		JobID:       importjob.ID("job-one"),
	}); err != nil {
		t.Fatalf("dispatch import job: %v", err)
	}

	select {
	case event := <-observer.seen:
		if event.Name != ports.EventImportJobWorkerFailed {
			t.Fatalf("expected worker failure event, got %+v", event)
		}
		if event.Fields["job_id"] != "job-one" ||
			event.Fields["tenant_id"] != "tenant-one" ||
			event.Fields["inventory_id"] != "inventory-one" ||
			event.Fields["error_class"] != "import_worker_execution_failed" {
			t.Fatalf("unexpected event fields: %+v", event.Fields)
		}
		if _, leaked := event.Fields["error"]; leaked {
			t.Fatalf("worker failure event leaked raw error: %+v", event.Fields)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for worker failure event")
	}
}

type failingImportJobRepository struct {
	*memory.Store
	err error
}

func (r failingImportJobRepository) UpdateImportJobIfStatus(context.Context, importjob.Record, importjob.Status) (bool, error) {
	return false, r.err
}

type recordingObserver struct {
	seen chan ports.Event
}

func (o *recordingObserver) Record(_ context.Context, event ports.Event) {
	if event.Name != ports.EventImportJobWorkerFailed {
		return
	}
	select {
	case o.seen <- event:
	default:
	}
}
