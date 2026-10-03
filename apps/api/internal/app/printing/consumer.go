package printing

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"regexp"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	p "github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ConsumerService struct {
	Jobs   *JobService
	Access ports.PrintConsumerAccess
}
type ClaimProof struct {
	AttemptID p.AttemptID
	SessionID p.SessionID
	Secret    string
	Revision  uint64
}

var consumerOpaqueID = regexp.MustCompile(`^[A-Za-z0-9_-]{16,100}$`)

func (proof ClaimProof) authority(connector p.ConnectorID) (p.AttemptAuthority, error) {
	bytes, err := base64.RawURLEncoding.DecodeString(proof.Secret)
	if err != nil || len(bytes) != 32 || base64.RawURLEncoding.EncodeToString(bytes) != proof.Secret || !consumerOpaqueID.MatchString(string(proof.AttemptID)) || !consumerOpaqueID.MatchString(string(proof.SessionID)) {
		return p.AttemptAuthority{}, apperrors.ErrInvalidInput
	}
	return p.AttemptAuthority{AttemptID: proof.AttemptID, ConnectorID: connector, SessionID: proof.SessionID, TokenDigest: sha256.Sum256([]byte(proof.Secret))}, nil
}
func (s ConsumerService) authenticate(ctx context.Context, token string) (p.Connector, error) {
	if s.Jobs == nil || s.Jobs.labels == nil || s.Access == nil {
		return p.Connector{}, ErrLabelsUnavailable
	}
	return s.Access.AuthenticateConsumer(ctx, token)
}
func (s ConsumerService) lookup(ctx context.Context, token string, id p.AttemptID) (p.Connector, p.ConsumerAuthority, p.Job, error) {
	c, err := s.authenticate(ctx, token)
	if err != nil {
		return c, p.ConsumerAuthority{}, p.Job{}, err
	}
	job, err := s.Jobs.jobs.FindPrintAttempt(ctx, c.Scope, c.ID, id)
	if err != nil {
		return c, p.ConsumerAuthority{}, job, jobError(err)
	}
	authority, err := s.Access.AuthorizePrinter(ctx, c, job.PrinterID, ports.PrinterPermissionConsume)
	return c, authority, job, err
}
func ownedAttempt(job p.Job, owner p.AttemptAuthority, current bool) (p.Attempt, error) {
	if current {
		if len(job.Attempts) == 0 || !job.Attempts[len(job.Attempts)-1].Authority.Equal(owner) {
			return p.Attempt{}, ports.ErrForbidden
		}
		return job.Attempts[len(job.Attempts)-1], nil
	}
	for _, attempt := range job.Attempts {
		if attempt.Authority.Equal(owner) {
			return attempt, nil
		}
	}
	return p.Attempt{}, ports.ErrForbidden
}
func (s ConsumerService) audit(c p.Connector, job p.Job, action audit.Action) (audit.Record, error) {
	target, targetID := audit.TargetPrintJob, string(job.ID)
	if job.ID == "" {
		target, targetID = audit.TargetInventory, job.Scope.InventoryID
	}
	return appsupport.NewAuditRecord(s.Jobs.labels.deps.IDs, s.Jobs.labels.deps.Clock, appsupport.AuditRecordInput{Principal: identity.Principal{ID: identity.PrincipalID(c.ServiceAccountID)}, TenantID: tenant.ID(job.Scope.TenantID), InventoryID: inventory.InventoryID(job.Scope.InventoryID), Source: audit.SourceAPI, Action: action, TargetType: target, TargetID: targetID, Metadata: map[string]string{"requestedBy": job.RequestedBy, "connectorId": string(c.ID)}})
}
func (s ConsumerService) transitionAudit(c p.Connector) ports.PrintJobAudit {
	return func(before, after p.Job) (audit.Record, error) {
		action := audit.ActionPrintJobClaimed
		switch after.Status {
		case p.JobPrinting:
			action = audit.ActionPrintJobStarted
		case p.JobCompleted:
			action = audit.ActionPrintJobCompleted
		case p.JobFailed:
			action = audit.ActionPrintJobFailed
		case p.JobUncertain:
			action = audit.ActionPrintJobUncertain
		case p.JobQueued:
			action = audit.ActionPrintJobReleased
		case p.JobCanceled:
			action = audit.ActionPrintJobCanceled
		}
		return s.audit(c, after, action)
	}
}
func (s ConsumerService) Claim(ctx context.Context, token string, printer p.PrinterID, proof ClaimProof) (p.Job, bool, error) {
	c, err := s.authenticate(ctx, token)
	if err != nil {
		return p.Job{}, false, err
	}
	owner, err := proof.authority(c.ID)
	if err != nil {
		return p.Job{}, false, err
	}
	authority, err := s.Access.AuthorizePrinter(ctx, c, printer, ports.PrinterPermissionConsume)
	if err != nil {
		return p.Job{}, false, err
	}
	replay := func() (p.Job, bool, error) {
		j, e := s.Jobs.jobs.FindPrintAttempt(ctx, c.Scope, c.ID, proof.AttemptID)
		if errors.Is(e, ports.ErrPrintJobNotFound) {
			return p.Job{}, false, nil
		}
		if e != nil {
			return p.Job{}, false, jobError(e)
		}
		if j.PrinterID != printer {
			return p.Job{}, false, apperrors.ErrConflict
		}
		if _, e = ownedAttempt(j, owner, true); e != nil {
			return p.Job{}, false, apperrors.ErrConflict
		}
		return j, true, nil
	}
	if j, found, e := replay(); e != nil || found {
		return j, found, e
	}
	j, found, err := s.Jobs.jobs.ClaimPrintJob(ctx, ports.PrintClaim{Authority: authority, Owner: owner, Now: s.Jobs.labels.deps.Clock.Now(), Lease: s.Jobs.config.Lease, ReportMaxAge: s.Jobs.config.ReadinessMaxAge, Audit: s.transitionAudit(c)})
	if err == nil && !found {
		return replay()
	}
	if found {
		s.observe(ctx, 0, j, err)
	}
	return j, found, jobError(err)
}
func (s ConsumerService) Read(ctx context.Context, token string, id p.AttemptID) (p.Job, error) {
	c, _, job, err := s.lookup(ctx, token, id)
	if err != nil {
		return p.Job{}, err
	}
	record, err := s.audit(c, job, audit.ActionPrintAttemptViewed)
	if err != nil {
		return p.Job{}, err
	}
	if err = s.Jobs.labels.deps.Audit.SaveAuditRecord(ctx, record); err != nil {
		return p.Job{}, err
	}
	return job, nil
}

func (s ConsumerService) Now() time.Time { return s.Jobs.labels.deps.Clock.Now() }

// Only committed state transitions emit events; retries are reads of prior evidence.
func (s ConsumerService) observe(ctx context.Context, revision uint64, job p.Job, err error) {
	if err != nil || job.Revision == revision || s.Jobs.labels.deps.Observer == nil {
		return
	}
	var name ports.EventName
	switch job.Status {
	case p.JobClaimed:
		name = ports.EventPrintJobClaimed
	case p.JobPrinting:
		name = ports.EventPrintJobStarted
	case p.JobCompleted:
		name = ports.EventPrintJobCompleted
	case p.JobFailed:
		name = ports.EventPrintJobFailed
	case p.JobUncertain:
		name = ports.EventPrintJobUncertain
	case p.JobQueued:
		name = ports.EventPrintJobReleased
	case p.JobCanceled:
		name = ports.EventPrintJobCanceled
	default:
		return
	}
	s.Jobs.labels.deps.Observer.Record(ctx, ports.Event{Name: name})
}
