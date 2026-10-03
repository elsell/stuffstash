package printworker

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (r *Runtime) runPrinter(ctx context.Context, registration printing.RegisteredPrinter, retired *atomic.Bool) error {
	failures := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var err error
		if retired.Load() {
			err = r.retiredSession(ctx, registration, retired)
		} else {
			err = r.deviceSession(ctx, registration, retired)
		}
		if terminal(err) || ctx.Err() != nil {
			return err
		}
		failures++
		r.Worker.Observer.Event(ctx, "cli.print.printer_unavailable")
		if reportErr := r.Registry.Report(ctx, registration.ID, failureReadiness(err)); terminal(reportErr) {
			return reportErr
		}
		if err = r.Worker.Waiter.Wait(ctx, r.Backoff.Delay(failures)); err != nil {
			return err
		}
	}
}

func (r *Runtime) deviceSession(ctx context.Context, registration printing.RegisteredPrinter, retired *atomic.Bool) error {
	journal, err := r.State.Acquire(ctx, registration.DeviceID)
	if err != nil {
		return err
	}
	defer journal.Close()
	worker := r.Worker
	worker.Config.PrinterID = registration.ID
	worker.Config.Media = registration.Media
	// Binding identifies the installation plus its server-authorized destination;
	// media edits do not change ownership of existing recovery evidence.
	worker.Config.Binding += ":" + registration.ID
	worker.Config.RecoveryOnly = true
	if err = worker.Step(ctx, journal, nil); err != nil {
		return err
	}
	connection, err := r.Devices.Open(ctx, registration)
	if err != nil {
		return err
	}
	defer connection.Close()
	for {
		worker.Config.RecoveryOnly = retired.Load()
		if worker.Config.RecoveryOnly {
			if err = worker.Step(ctx, journal, nil); err != nil {
				return err
			}
			if err = worker.Waiter.Wait(ctx, r.PollInterval); err != nil {
				return err
			}
			continue
		}
		readContext, readCancel := context.WithTimeout(ctx, worker.Config.ReadinessTimeout)
		readiness, err := connection.Readiness(readContext)
		readCancel()
		if err != nil {
			return err
		}
		if err = r.Registry.Report(ctx, registration.ID, readiness); err != nil {
			return err
		}
		if readiness.State == printing.Ready {
			err = worker.Step(ctx, journal, connection)
			if err != nil {
				return err
			}
		}
		if err = worker.Waiter.Wait(ctx, r.PollInterval); err != nil {
			return err
		}
	}
}

func failureReadiness(err error) printing.Readiness {
	if errors.Is(err, ports.ErrRecoveryRequired) || errors.Is(err, ports.ErrJournalCorrupt) {
		return printing.Readiness{State: printing.NeedsAttention}
	}
	if errors.Is(err, ports.ErrDeviceInUse) {
		return printing.Readiness{State: printing.Busy, Reason: printing.ActiveSubmission}
	}
	var failure *ports.Error
	if errors.As(err, &failure) {
		if failure.Category == "permission_denied" {
			return printing.Readiness{State: printing.NeedsAttention, Reason: printing.PermissionDenied}
		}
		if failure.Category == "configuration" {
			return printing.Readiness{State: printing.NeedsAttention, Reason: printing.UnsupportedTransport}
		}
	}
	return printing.Readiness{State: printing.Unavailable, Reason: printing.Disconnected}
}

// Retirement must not require an online printer to settle durable evidence.
func (r *Runtime) retiredSession(ctx context.Context, registration printing.RegisteredPrinter, retired *atomic.Bool) error {
	journal, err := r.State.Acquire(ctx, registration.DeviceID)
	if err != nil {
		return err
	}
	defer journal.Close()
	worker := r.Worker
	worker.Config.PrinterID = registration.ID
	worker.Config.Media = registration.Media
	worker.Config.Binding += ":" + registration.ID
	worker.Config.RecoveryOnly = true
	for retired.Load() {
		if err = worker.Step(ctx, journal, nil); err != nil {
			return err
		}
		if err = worker.Waiter.Wait(ctx, r.PollInterval); err != nil {
			return err
		}
	}
	return nil
}
