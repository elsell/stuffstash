package dataportability

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a ArchiveAccess) discoveryScope() ports.ArchiveJobScope {
	scope := a.scope()
	scope.AllInventories = a.InventoryID == ""
	return scope
}
func (s ArchiveService) authorizeDiscovery(ctx context.Context, a ArchiveAccess) error {
	if a.InventoryID != "" {
		return s.authorize(ctx, a)
	}
	if a.Principal.ID == "" {
		return apperrors.ErrUnauthenticated
	}
	if a.TenantID == "" {
		return apperrors.ErrInvalidInput
	}
	exists, err := s.deps.Tenants.TenantExists(ctx, a.TenantID)
	if err != nil {
		return err
	}
	if !exists {
		return apperrors.ErrNotFound
	}
	return s.deps.Authorizer.CheckTenant(ctx, a.Principal, ports.TenantPermissionView, a.TenantID)
}
func (s ArchiveService) authorizeJob(ctx context.Context, a ArchiveAccess, job archivejob.Record) error {
	a.InventoryID = inventory.InventoryID(job.SourceInventoryID)
	err := s.authorize(ctx, a)
	if errors.Is(err, ports.ErrForbidden) {
		return apperrors.ErrNotFound
	}
	return err
}
func (s ArchiveService) List(ctx context.Context, a ArchiveAccess, after string, limit int) ([]archivejob.Record, error) {
	if err := s.authorizeDiscovery(ctx, a); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		return nil, ports.ErrArchiveJobScope
	}
	result := make([]archivejob.Record, 0, limit)
	for len(result) < limit {
		jobs, err := s.deps.Jobs.ListArchiveJobs(ctx, a.discoveryScope(), after, limit)
		if err != nil {
			return nil, err
		}
		for _, job := range jobs {
			after = job.ID
			if err = s.authorizeJob(ctx, a, job); errors.Is(err, apperrors.ErrNotFound) {
				continue
			} else if err != nil {
				return nil, err
			}
			result = append(result, job)
			if len(result) == limit {
				break
			}
		}
		if len(jobs) < limit {
			break
		}
	}
	if err := s.auditArchiveRead(ctx, a, "", "list"); err != nil {
		return nil, err
	}
	return result, nil
}
