package gormstore

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ ports.ArchiveJobCommands = Store{}

func (s Store) CreateArchiveJobAudited(ctx context.Context, r archivejob.Record, record audit.Record) (archivejob.Record, error) {
	if !validArchiveJobAudit(r, record, audit.ActionArchiveJobCreated) {
		return archivejob.Record{}, archivejob.ErrInvalid
	}
	var result archivejob.Record
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = NewStore(tx).CreateArchiveJob(ctx, r)
		if err != nil {
			return err
		}
		// Idempotent retries return the original job without duplicate history.
		var count int64
		if err = tx.Model(&auditRecordModel{}).Where(clause.Eq{Column: "tenant_id", Value: result.TenantID}).Where(clause.Eq{Column: "target_type", Value: audit.TargetArchiveJob.String()}).Where(clause.Eq{Column: "target_id", Value: result.ID}).Where(clause.Eq{Column: "action", Value: audit.ActionArchiveJobCreated.String()}).Count(&count).Error; err != nil {
			return err
		}
		if count != 0 {
			return nil
		}
		record.TargetID = result.ID
		return createAuditRecord(tx, record)
	})
	if err != nil {
		return archivejob.Record{}, err
	}
	return result, nil
}
func (s Store) UpdateArchiveJobAudited(ctx context.Context, r archivejob.Record, expected int64, record audit.Record) (bool, error) {
	if !validArchiveJobAudit(r, record, audit.ActionArchiveJobUpdated) {
		return false, archivejob.ErrInvalid
	}
	changed := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		changed, err = NewStore(tx).UpdateArchiveJob(ctx, r, expected)
		if err != nil || !changed {
			return err
		}
		return createAuditRecord(tx, record)
	})
	return changed && err == nil, err
}
func validArchiveJobAudit(r archivejob.Record, a audit.Record, action audit.Action) bool {
	return a.ID != "" && !a.OccurredAt.IsZero() && a.TenantID.String() == r.TenantID && a.InventoryID.String() == r.SourceInventoryID && a.PrincipalID.String() == r.PrincipalID && a.TargetType == audit.TargetArchiveJob && a.TargetID == r.ID && a.Action == action
}
