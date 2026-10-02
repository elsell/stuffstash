package dataportability

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type ArchiveDependencies struct {
	Observer         ports.Observer
	Audit            ports.AuditRepository
	Plans            ports.ArchivePlanCodec
	MaxMetadataBytes int64
	MaxRecords       int
	Jobs             ports.ArchiveJobRepository
	Artifacts        ports.ArchiveArtifactRepository
	Commands         ports.ArchiveJobCommands
	Authorizer       ports.Authorizer
	Inventories      ports.InventoryRepository
	Tenants          ports.TenantRepository
	IDs              ports.IDGenerator
	Clock            ports.Clock
	Storage          ports.StreamingBlobStorage
	Scratch          ports.ArchiveScratchSpace
	MaxArchiveBytes  int64
	Retention        time.Duration
	CleanupTimeout   time.Duration
}
type ArchiveService struct{ deps ArchiveDependencies }
type ArchiveAccess struct {
	Principal   identity.Principal
	TenantID    tenant.ID
	InventoryID inventory.InventoryID
}

func NewArchiveService(d ArchiveDependencies) (ArchiveService, error) {
	if d.Audit == nil || d.Plans == nil || d.MaxRecords <= 0 || d.MaxMetadataBytes <= 0 || d.MaxMetadataBytes > d.MaxArchiveBytes || d.Jobs == nil || d.Artifacts == nil || d.Commands == nil || d.Authorizer == nil || d.Inventories == nil || d.Tenants == nil || d.IDs == nil || d.Clock == nil || d.Storage == nil || d.Scratch == nil || d.MaxArchiveBytes <= 0 || d.MaxArchiveBytes > ports.MaxSinglePutBytes || d.Retention <= 0 || d.CleanupTimeout <= 0 {
		return ArchiveService{}, apperrors.ErrInvalidInput
	}
	return ArchiveService{deps: d}, nil
}
func (a ArchiveAccess) scope() ports.ArchiveJobScope {
	return ports.ArchiveJobScope{TenantID: a.TenantID.String(), SourceInventoryID: a.InventoryID.String(), PrincipalID: a.Principal.ID.String()}
}
func (s ArchiveService) authorize(ctx context.Context, a ArchiveAccess) error {
	if a.Principal.ID == "" {
		return apperrors.ErrUnauthenticated
	}
	if a.TenantID == "" {
		return apperrors.ErrInvalidInput
	}
	if a.InventoryID != "" {
		item, found, err := s.deps.Inventories.InventoryByID(ctx, a.TenantID, a.InventoryID)
		if err != nil {
			return err
		}
		if !found || item.TenantID.String() != a.TenantID.String() {
			return apperrors.ErrNotFound
		}
		return s.deps.Authorizer.CheckInventory(ctx, a.Principal, ports.InventoryPermissionView, a.InventoryID)
	}
	exists, err := s.deps.Tenants.TenantExists(ctx, a.TenantID)
	if err != nil {
		return err
	}
	if !exists {
		return apperrors.ErrNotFound
	}
	return s.deps.Authorizer.CheckTenant(ctx, a.Principal, ports.TenantPermissionCreateInventory, a.TenantID)
}
func (s ArchiveService) Job(ctx context.Context, a ArchiveAccess, id string) (archivejob.Record, error) {
	if err := s.authorize(ctx, a); err != nil {
		return archivejob.Record{}, err
	}
	job, found, err := s.deps.Jobs.ArchiveJobByID(ctx, a.scope(), id)
	if err != nil {
		return archivejob.Record{}, err
	}
	if !found || job.PrincipalID != a.Principal.ID.String() {
		return archivejob.Record{}, apperrors.ErrNotFound
	}
	if err = s.auditArchiveRead(ctx, a, id, "get"); err != nil {
		return archivejob.Record{}, err
	}
	return job, nil
}
func (s ArchiveService) List(ctx context.Context, a ArchiveAccess, after string, limit int) ([]archivejob.Record, error) {
	if err := s.authorize(ctx, a); err != nil {
		return nil, err
	}
	jobs, err := s.deps.Jobs.ListArchiveJobs(ctx, a.scope(), after, limit)
	if err != nil {
		return nil, err
	}
	if err = s.auditArchiveRead(ctx, a, "", "list"); err != nil {
		return nil, err
	}
	return jobs, nil
}
func (s ArchiveService) CreateExport(ctx context.Context, a ArchiveAccess, key string, photos, files bool) (archivejob.Record, error) {
	if a.InventoryID == "" {
		return archivejob.Record{}, apperrors.ErrInvalidInput
	}
	if err := s.authorize(ctx, a); err != nil {
		return archivejob.Record{}, err
	}
	now := s.deps.Clock.Now()
	job, err := archivejob.New(archivejob.Request{RequestKey: key, ID: s.deps.IDs.NewID(), TenantID: a.TenantID.String(), PrincipalID: a.Principal.ID.String(), Kind: archivejob.Export, SourceInventoryID: a.InventoryID.String(), Photos: photos, OtherFiles: files}, now, now.Add(s.deps.Retention))
	if err != nil {
		return archivejob.Record{}, err
	}
	return s.create(ctx, job)
}
func (s ArchiveService) create(ctx context.Context, job archivejob.Record) (archivejob.Record, error) {
	record, err := s.jobAudit(job, audit.ActionArchiveJobCreated)
	if err != nil {
		return archivejob.Record{}, err
	}
	result, err := s.deps.Commands.CreateArchiveJobAudited(ctx, job, record)
	if err == nil && result.ID == job.ID {
		s.observe(ctx, ports.EventArchiveJobCreated, result)
	}
	return result, err
}
func (s ArchiveService) jobAudit(job archivejob.Record, action audit.Action) (audit.Record, error) {
	record, ok := audit.NewRecord(audit.ID(s.deps.IDs.NewID()), audit.TenantID(job.TenantID), audit.InventoryID(job.SourceInventoryID), audit.PrincipalID(job.PrincipalID), action, audit.SourceAPI, audit.TargetArchiveJob, job.ID, s.deps.Clock.Now(), job.RequestKey, map[string]string{"kind": string(job.Kind), "state": string(job.State)})
	if !ok {
		return audit.Record{}, apperrors.ErrInvalidInput
	}
	return record, nil
}
func (s ArchiveService) transition(ctx context.Context, prior, next archivejob.Record, err error) (archivejob.Record, error) {
	if err != nil {
		return archivejob.Record{}, err
	}
	record, err := s.jobAudit(next, audit.ActionArchiveJobUpdated)
	if err != nil {
		return archivejob.Record{}, err
	}
	changed, err := s.deps.Commands.UpdateArchiveJobAudited(ctx, next, prior.Revision, record)
	if err != nil {
		return archivejob.Record{}, err
	}
	if !changed {
		return archivejob.Record{}, ports.ErrArchiveJobConflict
	}
	s.observe(ctx, ports.EventArchiveJobUpdated, next)
	return next, nil
}
func (s ArchiveService) Cancel(ctx context.Context, a ArchiveAccess, id string) (archivejob.Record, error) {
	job, err := s.Job(ctx, a, id)
	if err != nil {
		return archivejob.Record{}, err
	}
	next, err := job.Cancel(s.deps.Clock.Now())
	return s.transition(ctx, job, next, err)
}
func (s ArchiveService) Retry(ctx context.Context, a ArchiveAccess, id string) (archivejob.Record, error) {
	job, err := s.Job(ctx, a, id)
	if err != nil {
		return archivejob.Record{}, err
	}
	next, err := job.Retry(s.deps.Clock.Now())
	return s.transition(ctx, job, next, err)
}
func (s ArchiveService) Approve(ctx context.Context, a ArchiveAccess, id, name string) (archivejob.Record, error) {
	job, err := s.Job(ctx, a, id)
	if err != nil {
		return archivejob.Record{}, err
	}
	if _, ok := inventory.NewName(name); !ok {
		return archivejob.Record{}, apperrors.ErrInvalidInput
	}
	next, err := job.Approve(name, s.deps.Clock.Now())
	return s.transition(ctx, job, next, err)
}

func (s ArchiveService) observe(ctx context.Context, name ports.EventName, job archivejob.Record) {
	if s.deps.Observer != nil {
		s.deps.Observer.Record(ctx, ports.Event{Name: name, Message: "archive job changed", Fields: map[string]string{"job_id": job.ID, "tenant_id": job.TenantID, "kind": string(job.Kind), "state": string(job.State)}})
	}
}
