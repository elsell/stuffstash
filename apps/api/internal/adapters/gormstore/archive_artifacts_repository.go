package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

type archiveArtifactModel struct {
	StorageKey        string    `gorm:"primaryKey;size:512"`
	JobID             string    `gorm:"not null;index"`
	TenantID          string    `gorm:"not null"`
	SourceInventoryID string    `gorm:"not null"`
	Kind              string    `gorm:"not null"`
	ExpiresAt         time.Time `gorm:"not null;index"`
}

func (archiveArtifactModel) TableName() string { return "archive_artifacts" }

var _ ports.ArchiveArtifactRepository = Store{}

func (s Store) RegisterArchiveArtifact(ctx context.Context, job archivejob.Record, key media.StorageKey, kind ports.ArchiveArtifactKind, now time.Time) error {
	if !validArchiveArtifact(job, key, kind) || now.IsZero() || now.Before(job.CreatedAt) || !now.Before(job.ExpiresAt) {
		return archivejob.ErrInvalid
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing archiveJobModel
		err := archiveJobQuery(tx, ports.ArchiveJobScope{TenantID: job.TenantID, SourceInventoryID: job.SourceInventoryID}).Clauses(clause.Locking{Strength: "UPDATE"}).Where(clause.Eq{Column: "id", Value: job.ID}).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if kind != ports.ArchiveArtifactSource {
				return ports.ErrArchiveJobConflict
			}
		} else if err != nil {
			return err
		} else {
			current, err := archiveJobRecordFromModel(existing)
			if err != nil {
				return err
			}
			if current.Request != job.Request || !current.ExpiresAt.Equal(job.ExpiresAt) || current.Phase != job.Phase || current.DestinationInventoryID != job.DestinationInventoryID || !now.Before(current.ExpiresAt) {
				return ports.ErrArchiveJobConflict
			}
			if kind != ports.ArchiveArtifactSource && (current.State != archivejob.Running || current.LeaseToken != job.LeaseToken || !now.Before(current.LeaseUntil)) {
				return ports.ErrArchiveJobConflict
			}
			if kind == ports.ArchiveArtifactSource && current.State != archivejob.Queued {
				return ports.ErrArchiveJobConflict
			}
		}
		row := archiveArtifactModel{StorageKey: key.String(), JobID: job.ID, TenantID: job.TenantID, SourceInventoryID: job.SourceInventoryID, Kind: string(kind), ExpiresAt: job.ExpiresAt}
		if err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
			return err
		}
		var retained archiveArtifactModel
		if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(clause.Eq{Column: "storage_key", Value: key.String()}).First(&retained).Error; err != nil {
			return err
		}
		var reserved int64
		if err = tx.Model(&mediaBlobKeyModel{}).Where(clause.Eq{Column: "storage_key", Value: key.String()}).Count(&reserved).Error; err != nil {
			return err
		}
		if reserved != 0 {
			return ports.ErrConflict
		}
		if retained.StorageKey != row.StorageKey || retained.JobID != row.JobID || retained.TenantID != row.TenantID || retained.SourceInventoryID != row.SourceInventoryID || retained.Kind != row.Kind || !retained.ExpiresAt.Equal(row.ExpiresAt) {
			return ports.ErrConflict
		}
		return nil
	})
}
func validArchiveArtifact(job archivejob.Record, key media.StorageKey, kind ports.ArchiveArtifactKind) bool {
	if job.ID == "" || job.TenantID == "" || key == "" {
		return false
	}
	prefix := "archives/" + job.TenantID + "/" + job.ID + "/"
	switch kind {
	case ports.ArchiveArtifactSource:
		return job.Kind == archivejob.Restore && key.String() == job.SourceArtifactID && key.String() == prefix+"source.zip"
	case ports.ArchiveArtifactPlan:
		return job.Kind == archivejob.Restore && job.Phase == archivejob.Validation && key.String() == prefix+job.LeaseToken+"/plan.json"
	case ports.ArchiveArtifactExport:
		return job.Kind == archivejob.Export && key.String() == prefix+job.LeaseToken+"/export.zip"
	case ports.ArchiveArtifactRestoredMedia:
		return job.Kind == archivejob.Restore && job.Phase == archivejob.Execution && job.DestinationInventoryID != "" && strings.HasPrefix(key.String(), job.TenantID+"/"+job.DestinationInventoryID+"/")
	}
	return false
}
func (s Store) ListDueArchiveArtifacts(ctx context.Context, now time.Time, limit int) ([]ports.ArchiveArtifact, error) {
	if now.IsZero() || limit <= 0 || limit > 200 {
		return nil, archivejob.ErrInvalid
	}
	var rows []archiveArtifactModel
	err := s.db.WithContext(ctx).Where(clause.Lte{Column: "expires_at", Value: now}).Order(clause.OrderByColumn{Column: clause.Column{Name: "expires_at"}}).Order(clause.OrderByColumn{Column: clause.Column{Name: "storage_key"}}).Limit(limit).Find(&rows).Error
	result := make([]ports.ArchiveArtifact, 0, len(rows))
	for _, r := range rows {
		result = append(result, ports.ArchiveArtifact{Key: media.StorageKey(r.StorageKey), JobID: r.JobID, TenantID: r.TenantID, SourceInventoryID: r.SourceInventoryID, Kind: ports.ArchiveArtifactKind(r.Kind), ExpiresAt: r.ExpiresAt})
	}
	return result, err
}
func (s Store) ListExpiredArchiveJobs(ctx context.Context, now time.Time, limit int) ([]archivejob.Record, error) {
	if now.IsZero() || limit <= 0 || limit > 200 {
		return nil, archivejob.ErrInvalid
	}
	return readArchiveJobs(s.db.WithContext(ctx).Model(&archiveJobModel{}).Where(clause.Lte{Column: "expires_at", Value: now}).Where(clause.Neq{Column: "state", Value: string(archivejob.Expired)}).Order(clause.OrderByColumn{Column: clause.Column{Name: "expires_at"}}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(limit))
}
func (s Store) RetireArchiveArtifact(ctx context.Context, a ports.ArchiveArtifact, now time.Time, eventID string) error {
	if eventID == "" || now.IsZero() || a.Key == "" || now.Before(a.ExpiresAt) {
		return archivejob.ErrInvalid
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock in the same order as registration/publication: job before artifact.
		var job archiveJobModel
		err := archiveJobQuery(tx, ports.ArchiveJobScope{TenantID: a.TenantID, SourceInventoryID: a.SourceInventoryID}).Clauses(clause.Locking{Strength: "UPDATE"}).Where(clause.Eq{Column: "id", Value: a.JobID}).First(&job).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil && (job.State != string(archivejob.Expired) || now.Before(job.ExpiresAt)) {
			return ports.ErrArchiveJobConflict
		}
		var row archiveArtifactModel
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(clause.Eq{Column: "storage_key", Value: a.Key.String()}).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if row.JobID != a.JobID || row.TenantID != a.TenantID || row.SourceInventoryID != a.SourceInventoryID || row.Kind != string(a.Kind) || !row.ExpiresAt.Equal(a.ExpiresAt) || now.Before(row.ExpiresAt) {
			return ports.ErrArchiveJobConflict
		}
		var references int64
		if err = tx.Model(&attachmentModel{}).Where(clause.Eq{Column: "storage_key", Value: a.Key.String()}).Count(&references).Error; err != nil {
			return err
		}
		if references == 0 {
			if err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&mediaBlobKeyModel{StorageKey: a.Key.String()}).Error; err != nil {
				return err
			}
			if err = tx.Create(&blobDeletionEventModel{ID: eventID, StorageKey: a.Key.String(), CreatedAt: now, UpdatedAt: now}).Error; err != nil {
				return err
			}
		}
		return tx.Where(clause.Eq{Column: "storage_key", Value: a.Key.String()}).Delete(&archiveArtifactModel{}).Error
	})
}
