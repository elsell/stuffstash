package dataportability

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (w ArchiveWorker) restoreReservations(ctx context.Context, job archivejob.Record) (ports.ArchiveKeyReservations, error) {
	result := ports.ArchiveKeyReservations{Fields: map[string]bool{}, Types: map[string]bool{}}
	err := collectPages(ctx, w.deps.MaxRecords, func(after string, limit int) ([]customfield.Definition, error) {
		return w.deps.Fields.ListTenantCustomFieldDefinitions(ctx, tenant.ID(job.TenantID), ports.CustomFieldDefinitionPageRequest{AfterDefinitionKey: after, Limit: limit, Lifecycle: ports.CustomizationLifecycleAll})
	}, func(f customfield.Definition) string { return f.CursorKey() }, func(f customfield.Definition) error {
		if f.TenantID.String() != job.TenantID || f.Scope != customfield.ScopeTenant {
			return ErrArchiveMetadata
		}
		result.Fields[f.Key.String()] = true
		return nil
	})
	if err != nil {
		return result, err
	}
	err = collectPages(ctx, w.deps.MaxRecords, func(after string, limit int) ([]customfield.AssetType, error) {
		return w.deps.Types.ListTenantCustomAssetTypes(ctx, tenant.ID(job.TenantID), ports.CustomAssetTypePageRequest{AfterAssetTypeKey: after, Limit: limit, Lifecycle: ports.CustomizationLifecycleAll})
	}, func(t customfield.AssetType) string { return t.CursorKey() }, func(t customfield.AssetType) error {
		if t.TenantID.String() != job.TenantID || t.Scope != customfield.ScopeTenant {
			return ErrArchiveMetadata
		}
		result.Types[t.Key.String()] = true
		return nil
	})
	return result, err
}
