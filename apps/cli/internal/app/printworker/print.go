package printworker

import (
	"context"
	"errors"

	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (w *Worker) print(ctx context.Context, journal ports.LockedPrintState, printer ports.PrinterConnection, claim printing.Claim, record *printing.JournalRecord) error {
	if err := w.renew(ctx, &claim, printing.RemoteClaimed); err != nil {
		return err
	}
	request, cancel, err := w.callContext(ctx, claim.LeaseExpiresAt)
	if err != nil {
		return err
	}
	body, contentType, err := w.Jobs.Artifact(request, claim.Control, w.Config.MaxArtifactBytes)
	cancel()
	if err != nil {
		return err
	}
	if err = validateArtifact(claim, body, contentType); err != nil {
		return err
	}
	request, cancel, err = w.callContext(ctx, claim.LeaseExpiresAt)
	if err != nil {
		return err
	}
	started, err := w.Jobs.Start(request, claim.Control)
	cancel()
	if err != nil {
		return err
	}
	if err = w.acceptStatus(&claim, started, printing.RemotePrinting); err != nil {
		return err
	}
	record.Phase = printing.JournalStarted
	record.Revision = claim.Control.Revision
	if err = journal.Save(ctx, record); err != nil {
		return err
	}
	for record.CompletedCopies < claim.Copies {
		if err = w.renew(ctx, &claim, printing.RemotePrinting); err != nil {
			return err
		}
		request, cancel, err = w.callContext(ctx, claim.LeaseExpiresAt)
		if err != nil {
			return err
		}
		readiness, readyErr := printer.Readiness(request)
		cancel()
		if readyErr != nil || readiness.State != printing.Ready {
			outcome := printing.NoOutput
			if record.CompletedCopies > 0 {
				outcome = printing.Uncertain
			}
			reason := readiness.Reason
			if reason == printing.NoReason {
				reason = printing.Disconnected
			}
			return w.finish(ctx, journal, &claim, record, printing.Evidence{Outcome: outcome, CompletedCopies: record.CompletedCopies, Reason: reason})
		}
		record.Phase = printing.JournalSubmitting
		record.Copy = record.CompletedCopies + 1
		record.Revision = claim.Control.Revision
		// This acknowledged fsync is the last gate before any device output.
		if err = journal.Save(ctx, record); err != nil {
			return err
		}
		request, cancel, err = w.callContext(ctx, claim.LeaseExpiresAt)
		if err != nil {
			return err
		}
		submission, submitErr := printer.Submit(request, printing.Label{AttemptID: claim.Control.AttemptID, Copy: record.Copy, PNG: body, Media: claim.Media})
		cancel()
		if submitErr != nil {
			evidence := printing.Evidence{Outcome: printing.Uncertain, CompletedCopies: record.CompletedCopies, Reason: printing.Interrupted}
			var known *printing.SubmissionError
			if errors.As(submitErr, &known) {
				evidence.Reason = known.Reason
				if known.Outcome == printing.NoOutput && record.CompletedCopies == 0 {
					evidence.Outcome = printing.NoOutput
				}
			}
			return w.finish(ctx, journal, &claim, record, evidence)
		}
		for {
			if err = w.renew(ctx, &claim, printing.RemotePrinting); err != nil {
				return err
			}
			request, cancel, err = w.callContext(ctx, claim.LeaseExpiresAt)
			if err != nil {
				return err
			}
			observation, observeErr := printer.Observe(request, submission)
			cancel()
			if observeErr != nil {
				return w.finish(ctx, journal, &claim, record, printing.Evidence{Outcome: printing.Uncertain, CompletedCopies: record.CompletedCopies, Reason: printing.Interrupted})
			}
			if observation.Outcome == printing.Completed {
				break
			}
			if observation.Outcome != printing.Pending {
				outcome := printing.Uncertain
				if observation.Outcome == printing.NoOutput && record.CompletedCopies == 0 {
					outcome = printing.NoOutput
				}
				return w.finish(ctx, journal, &claim, record, printing.Evidence{Outcome: outcome, CompletedCopies: record.CompletedCopies, Reason: observation.Reason})
			}
			request, cancel, err = w.callContext(ctx, claim.LeaseExpiresAt)
			if err != nil {
				return err
			}
			err = w.Waiter.Wait(request, w.Config.ObserveInterval)
			cancel()
			if err != nil {
				return err
			}
		}
		record.CompletedCopies++
		record.Phase = printing.JournalStarted
		if record.CompletedCopies == claim.Copies {
			record.Phase = printing.JournalCompleted
		}
		record.Revision = claim.Control.Revision
		if err = journal.Save(ctx, record); err != nil {
			return err
		}
	}
	return w.finish(ctx, journal, &claim, record, printing.Evidence{Outcome: printing.Completed, CompletedCopies: record.CompletedCopies})
}
func (w *Worker) finish(ctx context.Context, journal ports.LockedPrintState, claim *printing.Claim, record *printing.JournalRecord, evidence printing.Evidence) error {
	switch evidence.Outcome {
	case printing.Completed:
		record.Phase = printing.JournalCompleted
	case printing.NoOutput:
		record.Phase = printing.JournalNoOutput
	default:
		record.Phase = printing.JournalUncertain
	}
	record.Revision = claim.Control.Revision
	if err := journal.Save(ctx, record); err != nil {
		return err
	}
	if err := w.Jobs.Outcome(ctx, claim.Control, evidence); err != nil {
		return err
	}
	w.Observer.Event(ctx, "cli.print.outcome_reported")
	if evidence.Outcome == printing.Uncertain {
		return ports.ErrRecoveryRequired
	}
	return journal.Save(ctx, nil)
}
