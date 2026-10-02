package gormstore

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/archivejob"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ ports.ArchiveJobRepository = Store{}

func (s Store) CreateArchiveJob(ctx context.Context, r archivejob.Record) (archivejob.Record, error) {
	initial, err := archivejob.New(r.Request, r.CreatedAt, r.ExpiresAt)
	if err != nil || initial != r {
		return archivejob.Record{}, archivejob.ErrInvalid
	}
	m, err := archiveJobModelFromRecord(r)
	if err != nil {
		return archivejob.Record{}, err
	}
	result := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&m)
	if result.Error != nil {
		return archivejob.Record{}, result.Error
	}
	if result.RowsAffected == 1 {
		return r, nil
	}
	var existing archiveJobModel
	err = s.db.WithContext(ctx).Where(clause.Eq{Column: "tenant_id", Value: r.TenantID}).Where(clause.Eq{Column: "principal_id", Value: r.PrincipalID}).Where(clause.Eq{Column: "request_key", Value: r.RequestKey}).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return archivejob.Record{}, ports.ErrArchiveJobConflict
	}
	if err != nil {
		return archivejob.Record{}, err
	}
	previous, err := archiveJobRecordFromModel(existing)
	if err != nil {
		return archivejob.Record{}, err
	}
	sameRequest := r.Request
	sameRequest.ID = previous.ID
	if sameRequest.Kind == archivejob.Restore && sameRequest.SourceSHA256 == previous.SourceSHA256 {
		sameRequest.SourceArtifactID = previous.SourceArtifactID
	}
	if previous.Request != sameRequest {
		return archivejob.Record{}, ports.ErrArchiveJobConflict
	}
	return previous, nil
}
func (s Store) ArchiveJobByID(ctx context.Context, scope ports.ArchiveJobScope, id string) (archivejob.Record, bool, error) {
	if scope.TenantID == "" || id == "" {
		return archivejob.Record{}, false, ports.ErrArchiveJobScope
	}
	var m archiveJobModel
	err := archiveJobQuery(s.db.WithContext(ctx), scope).Where(clause.Eq{Column: "id", Value: id}).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return archivejob.Record{}, false, nil
	}
	if err != nil {
		return archivejob.Record{}, false, err
	}
	r, err := archiveJobRecordFromModel(m)
	return r, err == nil, err
}
func (s Store) ListArchiveJobs(ctx context.Context, scope ports.ArchiveJobScope, after string, limit int) ([]archivejob.Record, error) {
	if scope.TenantID == "" || limit <= 0 || limit > 200 {
		return nil, ports.ErrArchiveJobScope
	}
	q := archiveJobQuery(s.db.WithContext(ctx), scope)
	if after != "" {
		q = q.Where(clause.Gt{Column: "id", Value: after})
	}
	return readArchiveJobs(q.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(limit))
}
func (s Store) UpdateArchiveJob(ctx context.Context, r archivejob.Record, expected int64) (bool, error) {
	if r.TenantID == "" || r.ID == "" || expected <= 0 || r.Revision != expected+1 {
		return false, archivejob.ErrInvalid
	}
	var updated bool
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		scope := ports.ArchiveJobScope{TenantID: r.TenantID, SourceInventoryID: r.SourceInventoryID}
		var current archiveJobModel
		err := archiveJobQuery(tx, scope).Clauses(clause.Locking{Strength: "UPDATE"}).Where(clause.Eq{Column: "id", Value: r.ID}).Where(clause.Eq{Column: "revision", Value: expected}).First(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		previous, err := archiveJobRecordFromModel(current)
		if err != nil {
			return err
		}
		if err := archivejob.ValidateSuccessor(previous, r); err != nil {
			return err
		}
		m, err := archiveJobModelFromRecord(r)
		if err != nil {
			return err
		}
		result := archiveJobQuery(tx.Model(&archiveJobModel{}), scope).Where(clause.Eq{Column: "id", Value: r.ID}).Where(clause.Eq{Column: "revision", Value: expected}).Updates(map[string]any{"state_json": m.StateJSON, "state": m.State, "revision": m.Revision, "updated_at": m.UpdatedAt, "lease_until": m.LeaseUntil})
		updated = result.RowsAffected == 1
		return result.Error
	})
	return updated, err
}
func (s Store) ListRunnableArchiveJobs(ctx context.Context, now time.Time, limit int) ([]archivejob.Record, error) {
	if now.IsZero() || limit <= 0 || limit > 200 {
		return nil, ports.ErrArchiveJobScope
	}
	q := s.db.WithContext(ctx).Model(&archiveJobModel{}).Where(clause.Gt{Column: "expires_at", Value: now}).Where(clause.Or(clause.Eq{Column: "state", Value: string(archivejob.Queued)}, clause.And(clause.Eq{Column: "state", Value: string(archivejob.Running)}, clause.Lte{Column: "lease_until", Value: now})))
	return readArchiveJobs(q.Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(limit))
}
func archiveJobQuery(db *gorm.DB, scope ports.ArchiveJobScope) *gorm.DB {
	return db.Where(clause.Eq{Column: "tenant_id", Value: scope.TenantID}).Where(clause.Eq{Column: "source_inventory_id", Value: scope.SourceInventoryID})
}
func readArchiveJobs(q *gorm.DB) ([]archivejob.Record, error) {
	var models []archiveJobModel
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]archivejob.Record, 0, len(models))
	for _, m := range models {
		r, err := archiveJobRecordFromModel(m)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}
func archiveJobModelFromRecord(r archivejob.Record) (archiveJobModel, error) {
	request, err := json.Marshal(r.Request)
	if err != nil {
		return archiveJobModel{}, err
	}
	state, err := json.Marshal(r)
	if err != nil {
		return archiveJobModel{}, err
	}
	return archiveJobModel{ID: r.ID, TenantID: r.TenantID, SourceInventoryID: r.SourceInventoryID, PrincipalID: r.PrincipalID, RequestKey: r.RequestKey, RequestJSON: string(request), StateJSON: string(state), State: string(r.State), Revision: r.Revision, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, ExpiresAt: r.ExpiresAt, LeaseUntil: r.LeaseUntil}, nil
}
func archiveJobRecordFromModel(m archiveJobModel) (archivejob.Record, error) {
	var r archivejob.Record
	if err := json.Unmarshal([]byte(m.StateJSON), &r); err != nil {
		return r, err
	}
	if err := json.Unmarshal([]byte(m.RequestJSON), &r.Request); err != nil {
		return r, err
	}
	if r.ID != m.ID || r.TenantID != m.TenantID || r.SourceInventoryID != m.SourceInventoryID || r.PrincipalID != m.PrincipalID || r.RequestKey != m.RequestKey || r.Revision != m.Revision || string(r.State) != m.State || !r.CreatedAt.Equal(m.CreatedAt) || !r.UpdatedAt.Equal(m.UpdatedAt) || !r.ExpiresAt.Equal(m.ExpiresAt) || !r.LeaseUntil.Equal(m.LeaseUntil) {
		return archivejob.Record{}, archivejob.ErrInvalid
	}
	return r, nil
}
