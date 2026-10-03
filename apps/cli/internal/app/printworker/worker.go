package printworker

import (
	"context"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

type Config struct {
	RecoveryOnly                                   bool
	Binding, PrinterID, SessionID                  string
	Media                                          printing.Media
	MaxArtifactBytes                               int64
	LeaseSafety, ObserveInterval, ReadinessTimeout time.Duration
}
type Worker struct {
	Jobs     ports.PrintJobs
	Clock    ports.Clock
	Waiter   ports.Waiter
	Identity ports.PrintIdentity
	Observer ports.Observer
	Config   Config
}

// Step owns at most one new attempt. The caller always holds the journal
// reservation; output-capable calls also hold the physical connection OS lock.
// RecoveryOnly deliberately needs no hardware connection.
// Recovery always returns before a subsequent claim can be considered.
func (w *Worker) Step(ctx context.Context, journal ports.LockedPrintState, printer ports.PrinterConnection) error {
	if w.Config.Binding == "" || w.Config.PrinterID == "" || w.Config.SessionID == "" || w.Config.MaxArtifactBytes <= 0 || w.Config.LeaseSafety <= 0 || w.Config.ObserveInterval <= 0 || w.Config.ReadinessTimeout <= 0 {
		return errors.New("invalid print worker configuration")
	}
	record, err := journal.Load(ctx)
	if err != nil && !errors.Is(err, ports.ErrJournalMissing) {
		return err
	}
	if record != nil {
		err = w.recover(ctx, journal, *record, printer)
		if err == nil {
			w.Observer.Event(ctx, "cli.print.recovered")
		}
		return err
	}
	unsettled, err := w.Jobs.Unsettled(ctx, w.Config.PrinterID)
	if err != nil {
		return err
	}
	if len(unsettled) > 0 {
		return w.recoverMissingJournal(ctx, printer, unsettled)
	}
	if w.Config.RecoveryOnly {
		return nil
	}
	readContext, readCancel := context.WithTimeout(ctx, w.Config.ReadinessTimeout)
	readiness, err := printer.Readiness(readContext)
	readCancel()
	if err != nil {
		return err
	}
	if readiness.State != printing.Ready {
		return ports.Failure("unavailable", "printer needs attention before printing")
	}
	attempt, err := w.Identity.Attempt()
	if err != nil {
		return err
	}
	token, err := w.Identity.ClaimToken()
	if err != nil {
		return err
	}
	record = &printing.JournalRecord{Binding: w.Config.Binding, AttemptID: attempt, SessionID: w.Config.SessionID, ClaimToken: token, Phase: printing.JournalClaiming}
	if err = journal.Save(ctx, record); err != nil {
		return err
	}
	control := printing.AttemptControl{AttemptID: attempt, SessionID: w.Config.SessionID, ClaimToken: token}
	claim, err := w.Jobs.Claim(ctx, w.Config.PrinterID, control)
	if err != nil {
		return err
	}
	if claim == nil {
		return journal.Save(ctx, nil)
	}
	if err = w.validateClaim(*claim, control); err != nil {
		return err
	}
	record.JobID = claim.JobID
	record.Revision = claim.Control.Revision
	record.Phase = printing.JournalPrepared
	if err = journal.Save(ctx, record); err != nil {
		return err
	}
	w.Observer.Event(ctx, "cli.print.claimed")
	return w.print(ctx, journal, printer, *claim, record)
}
func (w *Worker) callContext(ctx context.Context, expires time.Time) (context.Context, context.CancelFunc, error) {
	remaining := expires.Sub(w.Clock.Now()) - w.Config.LeaseSafety
	if remaining <= 0 {
		return nil, nil, ports.Failure("lease_expired", "print lease is not current; no output was started")
	}
	next, cancel := context.WithTimeout(ctx, remaining)
	return next, cancel, nil
}
func (w *Worker) renew(ctx context.Context, claim *printing.Claim, phase printing.RemotePhase) error {
	request, cancel, err := w.callContext(ctx, claim.LeaseExpiresAt)
	if err != nil {
		return err
	}
	defer cancel()
	status, err := w.Jobs.Renew(request, claim.Control)
	if err != nil {
		return err
	}
	return w.acceptStatus(claim, status, phase)
}
func (w *Worker) acceptStatus(claim *printing.Claim, status printing.AttemptStatus, expected printing.RemotePhase) error {
	if status.AttemptID != claim.Control.AttemptID || status.SessionID != claim.Control.SessionID || status.JobID != claim.JobID || status.Revision <= claim.Control.Revision || status.Phase != expected || !status.LeaseExpiresAt.After(w.Clock.Now().Add(w.Config.LeaseSafety)) {
		return ports.Failure("invalid_response", "API did not acknowledge a current owned print attempt")
	}
	claim.Control.Revision = status.Revision
	claim.LeaseExpiresAt = status.LeaseExpiresAt
	return nil
}
