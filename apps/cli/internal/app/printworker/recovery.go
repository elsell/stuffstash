package printworker

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (w *Worker) recover(ctx context.Context, journal ports.LockedPrintState, record printing.JournalRecord, printer ports.PrinterConnection) error {
	if record.Binding != w.Config.Binding {
		return ports.ErrRecoveryRequired
	}
	status, err := w.Jobs.Attempt(ctx, record.AttemptID)
	if errors.Is(err, ports.ErrAttemptNotFound) && record.Phase == printing.JournalClaiming {
		return journal.Save(ctx, nil)
	}
	if err != nil {
		return err
	}
	if status.AttemptID != record.AttemptID || status.SessionID != record.SessionID || record.JobID != "" && status.JobID != record.JobID {
		return ports.ErrRecoveryRequired
	}
	switch status.Phase {
	case printing.RemoteCompleted, printing.RemoteFailed, printing.RemoteCanceled:
		return journal.Save(ctx, nil)
	case printing.RemoteUncertain:
		evidence := printing.Evidence{Outcome: printing.Uncertain, CompletedCopies: record.CompletedCopies, Reason: printing.Interrupted}
		switch record.Phase {
		case printing.JournalCompleted:
			evidence.Outcome = printing.Completed
			evidence.Reason = printing.NoReason
		case printing.JournalClaiming, printing.JournalPrepared, printing.JournalStarted, printing.JournalNoOutput:
			if record.CompletedCopies == 0 {
				evidence.Outcome = printing.NoOutput
				evidence.Reason = printing.NoReason
			}
		}
		if evidence.Outcome == printing.Uncertain {
			return w.confirmIdle(ctx, printer, []printing.AttemptStatus{status})
		}
		if err = w.Jobs.Reconcile(ctx, record.AttemptID, status.Revision, evidence); err != nil {
			return err
		}
		return journal.Save(ctx, nil)
	default:
		return ports.ErrRecoveryRequired
	}
}

// Lost local evidence cannot establish completion or zero output. Scoped server
// identity plus an independently locked idle device can still permit a human
// acknowledgement, without manufacturing a journal or submitting output.
func (w *Worker) recoverMissingJournal(ctx context.Context, printer ports.PrinterConnection, attempts []printing.AttemptStatus) error {
	for _, attempt := range attempts {
		if attempt.Phase != printing.RemoteUncertain || attempt.AttemptID == "" || attempt.Revision == 0 {
			return ports.ErrRecoveryRequired
		}
	}
	return w.confirmIdle(ctx, printer, attempts)
}
func (w *Worker) confirmIdle(ctx context.Context, printer ports.PrinterConnection, attempts []printing.AttemptStatus) error {
	idle, ok := printer.(ports.PrinterIdleConfirmation)
	if !ok {
		return ports.ErrRecoveryRequired
	}
	check, cancel := context.WithTimeout(ctx, w.Config.ReadinessTimeout)
	err := idle.ConfirmIdle(check)
	cancel()
	if err != nil {
		return err
	}
	for _, attempt := range attempts {
		if err = w.Jobs.ConfirmIdle(ctx, attempt.AttemptID, attempt.Revision); err != nil {
			return err
		}
	}
	return ports.ErrRecoveryRequired
}
