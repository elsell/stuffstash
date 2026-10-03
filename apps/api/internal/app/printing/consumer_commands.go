package printing

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s ConsumerService) dispatchAllowed(ctx context.Context, j p.Job) error {
	scope := LabelScope{Principal: identity.Principal{ID: identity.PrincipalID(j.RequestedBy)}, TenantID: tenant.ID(j.Scope.TenantID), InventoryID: inventory.InventoryID(j.Scope.InventoryID)}
	if err := s.Jobs.access(ctx, scope, true); err != nil {
		return err
	}
	if j.Kind == p.JobAssetLabel {
		a, err := s.Jobs.labels.asset(ctx, scope, asset.ID(j.AssetID))
		if err != nil {
			return err
		}
		if a.LifecycleState != asset.LifecycleStateActive {
			return apperrors.ErrNotFound
		}
	}
	return nil
}
func (s ConsumerService) Start(ctx context.Context, token string, proof ClaimProof) (p.Job, error) {
	c, authority, j, err := s.lookup(ctx, token, proof.AttemptID)
	if err != nil {
		return p.Job{}, err
	}
	owner, err := proof.authority(c.ID)
	if err != nil {
		return p.Job{}, err
	}
	if _, err = ownedAttempt(j, owner, true); err != nil {
		return p.Job{}, err
	}
	now := s.Jobs.labels.deps.Clock.Now()
	if err = s.dispatchAllowed(ctx, j); err != nil {
		if errors.Is(err, ports.ErrForbidden) || errors.Is(err, apperrors.ErrNotFound) || errors.Is(err, apperrors.ErrUnauthorized) {
			_, cancelErr := s.Jobs.jobs.UpdatePrintJob(ctx, ports.PrintJobUpdate{Scope: j.Scope, PrinterID: j.PrinterID, JobID: j.ID, Authority: &authority, Now: now, Change: func(job *p.Job, _ p.Printer) error { return job.Cancel(now, proof.Revision) }, Audit: s.transitionAudit(c)})
			if cancelErr != nil {
				return p.Job{}, jobError(cancelErr)
			}
			return p.Job{}, apperrors.ErrConflict
		}
		return p.Job{}, err
	}
	result, err := s.Jobs.jobs.UpdatePrintJob(ctx, ports.PrintJobUpdate{Scope: j.Scope, PrinterID: j.PrinterID, JobID: j.ID, Authority: &authority, Now: now, StartReportMaxAge: s.Jobs.config.ReadinessMaxAge, Change: func(job *p.Job, _ p.Printer) error {
		if !job.Artifact.ExpiresAt.After(now) {
			return apperrors.ErrConflict
		}
		return job.Start(owner, now, proof.Revision)
	}, Audit: s.transitionAudit(c)})
	s.observe(ctx, j.Revision, result, err)
	return result, jobError(err)
}
func (s ConsumerService) Renew(ctx context.Context, token string, proof ClaimProof) (p.Job, error) {
	c, authority, j, err := s.lookup(ctx, token, proof.AttemptID)
	if err != nil {
		return p.Job{}, err
	}
	owner, err := proof.authority(c.ID)
	if err != nil {
		return p.Job{}, err
	}
	result, err := s.Jobs.jobs.RenewPrintJob(ctx, ports.PrintLeaseRenewal{Authority: authority, Owner: owner, JobID: j.ID, Revision: proof.Revision, Now: s.Jobs.labels.deps.Clock.Now(), Lease: s.Jobs.config.Lease})
	return result, jobError(err)
}
func (s ConsumerService) Report(ctx context.Context, token string, proof ClaimProof, outcome p.Outcome) (p.Job, error) {
	c, authority, j, err := s.lookup(ctx, token, proof.AttemptID)
	if err != nil {
		return p.Job{}, err
	}
	owner, err := proof.authority(c.ID)
	if err != nil {
		return p.Job{}, err
	}
	now := s.Jobs.labels.deps.Clock.Now()
	result, err := s.Jobs.jobs.UpdatePrintJob(ctx, ports.PrintJobUpdate{Scope: j.Scope, PrinterID: j.PrinterID, JobID: j.ID, Authority: &authority, Now: now, Change: func(job *p.Job, _ p.Printer) error { return job.Report(owner, now, proof.Revision, outcome) }, Audit: s.transitionAudit(c)})
	s.observe(ctx, j.Revision, result, err)
	return result, jobError(err)
}
func (s ConsumerService) Content(ctx context.Context, token string, proof ClaimProof) ([]byte, error) {
	c, _, j, err := s.lookup(ctx, token, proof.AttemptID)
	if err != nil {
		return nil, err
	}
	owner, err := proof.authority(c.ID)
	if err != nil {
		return nil, err
	}
	a, err := ownedAttempt(j, owner, true)
	if err != nil {
		return nil, err
	}
	now := s.Jobs.labels.deps.Clock.Now()
	if j.Revision != proof.Revision || (j.Status != p.JobClaimed && j.Status != p.JobPrinting) || !a.LeaseExpiresAt.After(now) || !j.Artifact.ExpiresAt.After(now) {
		return nil, apperrors.ErrConflict
	}
	if err = s.dispatchAllowed(ctx, j); err != nil {
		return nil, err
	}
	content, err := s.Jobs.jobs.GetPrintJobContent(ctx, j.Scope, j.ID, now)
	if err != nil {
		return nil, jobError(err)
	}
	record, err := s.audit(c, j, audit.ActionPrintJobContentDownloaded)
	if err != nil {
		return nil, err
	}
	if err = s.Jobs.labels.deps.Audit.SaveAuditRecord(ctx, record); err != nil {
		return nil, err
	}
	return content, nil
}
