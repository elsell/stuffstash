package printing

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *JobService) Resolve(ctx context.Context, scope LabelScope, id p.JobID, revision uint64, outcome p.ReportedOutcome, acknowledge bool) (p.Job, error) {
	if err := s.access(ctx, scope, true); err != nil {
		return p.Job{}, err
	}
	j, err := s.jobs.GetPrintJob(ctx, jobScope(scope), id)
	if err != nil {
		return p.Job{}, jobError(err)
	}
	now := s.labels.deps.Clock.Now()
	result, err := s.jobs.UpdatePrintJob(ctx, ports.PrintJobUpdate{Scope: j.Scope, PrinterID: j.PrinterID, JobID: id, Now: now, Change: func(job *p.Job, _ p.Printer) error {
		return job.Resolve(string(scope.Principal.ID), outcome, acknowledge, now, revision)
	}, Audit: func(before, after p.Job) (audit.Record, error) {
		record, err := s.audit(scope, audit.ActionPrintJobResolved, id)
		record.Metadata = map[string]string{"reported_outcome": string(outcome)}
		return record, err
	}})
	if err == nil && result.Revision != j.Revision && s.labels.deps.Observer != nil {
		s.labels.deps.Observer.Record(ctx, ports.Event{Name: ports.EventPrintJobResolved})
	}
	return result, jobError(err)
}
func (s ConsumerService) ConfirmIdle(ctx context.Context, token string, id p.AttemptID, revision uint64) (p.Job, error) {
	c, authority, j, err := s.lookup(ctx, token, id)
	if err != nil {
		return p.Job{}, err
	}
	now := s.Now()
	result, err := s.Jobs.jobs.UpdatePrintJob(ctx, ports.PrintJobUpdate{Scope: j.Scope, PrinterID: j.PrinterID, JobID: j.ID, Authority: &authority, Now: now, Change: func(job *p.Job, _ p.Printer) error { return job.ConfirmIdle(id, c.ID, now, revision) }, Audit: func(before, after p.Job) (audit.Record, error) {
		return s.audit(c, after, audit.ActionPrintJobIdleConfirmed)
	}})
	if err == nil && result.Revision != j.Revision && s.Jobs.labels.deps.Observer != nil {
		s.Jobs.labels.deps.Observer.Record(ctx, ports.Event{Name: ports.EventPrintJobIdleConfirmed})
	}
	return result, jobError(err)
}
