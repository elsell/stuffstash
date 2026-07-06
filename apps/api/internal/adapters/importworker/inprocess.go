package importworker

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type InProcess struct {
	application app.App
	observer    ports.Observer
}

func NewInProcess(application app.App, observer ports.Observer) InProcess {
	return InProcess{application: application, observer: observer}
}

func (w InProcess) ExecuteImportJob(ctx context.Context, command ports.ImportJobCommand) (importjob.Record, error) {
	job, err := w.application.GetImportJobForWorker(ctx, command)
	if err != nil {
		return importjob.Record{}, err
	}
	go func() {
		if _, err := w.application.ExecuteImportJob(context.Background(), command); err != nil {
			w.recordFailure(command, err)
		}
	}()
	return job, nil
}

func (w InProcess) CancelImportJob(context.Context, importjob.ID, importjob.CancellationMode) error {
	return nil
}

func (w InProcess) recordFailure(command ports.ImportJobCommand, err error) {
	if w.observer == nil {
		return
	}
	w.observer.Record(context.Background(), ports.Event{
		Name:    ports.EventImportJobWorkerFailed,
		Message: "import job worker execution failed",
		Fields: map[string]string{
			"tenant_id":    tenant.ID(command.TenantID).String(),
			"inventory_id": inventory.InventoryID(command.InventoryID).String(),
			"job_id":       command.JobID.String(),
			"error_class":  "import_worker_execution_failed",
		},
	})
}
