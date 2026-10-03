package printing

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s ConsumerService) Unsettled(ctx context.Context, token string, printer p.PrinterID, limit int, cursor string) (JobPage, error) {
	c, err := s.authenticate(ctx, token)
	if err != nil {
		return JobPage{}, err
	}
	if limit < 1 || limit > 100 {
		return JobPage{}, apperrors.ErrInvalidInput
	}
	if printer != "" {
		if _, err = s.Access.AuthorizePrinter(ctx, c, printer, ports.PrinterPermissionConsume); err != nil {
			return JobPage{}, err
		}
	}
	scope := c.Scope.TenantID + ":" + c.Scope.InventoryID + ":" + string(c.ID) + ":" + string(printer)
	after, err := appsupport.DecodePageCursor("print-consumer-attempts", scope, cursor)
	if err != nil {
		return JobPage{}, err
	}
	jobs, err := s.Jobs.jobs.ListPrintConsumerAttempts(ctx, c.Scope, c.ID, printer, limit+1, after)
	if err != nil {
		return JobPage{}, jobError(err)
	}
	page := JobPage{Items: []p.Job{}, Limit: limit, HasMore: len(jobs) > limit}
	if page.HasMore {
		jobs = jobs[:limit]
		page.NextCursor = appsupport.EncodePageCursor("print-consumer-attempts", scope, string(jobs[len(jobs)-1].ID))
	}
	for _, j := range jobs {
		if _, err = s.Access.AuthorizePrinter(ctx, c, j.PrinterID, ports.PrinterPermissionConsume); errors.Is(err, ports.ErrForbidden) || errors.Is(err, apperrors.ErrUnauthorized) {
			continue
		} else if err != nil {
			return JobPage{}, err
		}
		page.Items = append(page.Items, j)
	}
	record, err := s.audit(c, p.Job{Scope: c.Scope}, audit.ActionPrintAttemptsListed)
	if err != nil {
		return JobPage{}, err
	}
	if err = s.Jobs.labels.deps.Audit.SaveAuditRecord(ctx, record); err != nil {
		return JobPage{}, err
	}
	return page, nil
}
func (s ConsumerService) Reconcile(ctx context.Context, token string, id p.AttemptID, revision uint64, outcome p.Outcome) (p.Job, error) {
	c, authority, j, err := s.lookup(ctx, token, id)
	if err != nil {
		return p.Job{}, err
	}
	now := s.Now()
	result, err := s.Jobs.jobs.UpdatePrintJob(ctx, ports.PrintJobUpdate{Scope: j.Scope, PrinterID: j.PrinterID, JobID: j.ID, Authority: &authority, Now: now, Change: func(job *p.Job, _ p.Printer) error {
		if job.Revision != revision {
			// The domain accepts an identical settled outcome without changing state.
			// Keep this recovery path usable after a lost successful API response.
			if len(job.Attempts) > 0 {
				a := job.Attempts[len(job.Attempts)-1]
				if a.ID == id && a.Authority.ConnectorID == c.ID && a.Outcome.Kind != "" && a.Outcome == outcome {
					return job.Reconcile(id, c.ID, outcome, now, revision)
				}
			}
			return apperrors.ErrConflict
		}
		job.Expire(now)
		return job.Reconcile(id, c.ID, outcome, now, job.Revision)
	}, Audit: s.transitionAudit(c)})
	s.observe(ctx, j.Revision, result, err)
	return result, jobError(err)
}
