package printworker

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (w *Worker) recover(ctx context.Context, journal ports.LockedPrintState, record printing.JournalRecord) error {
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
		if err = w.Jobs.Reconcile(ctx, record.AttemptID, status.Revision, evidence); err != nil {
			return err
		}
		if evidence.Outcome == printing.Uncertain {
			return ports.ErrRecoveryRequired
		}
		return journal.Save(ctx, nil)
	default:
		return ports.ErrRecoveryRequired
	}
}
