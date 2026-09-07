package gormstore

import (
	"context"
	"database/sql"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func (s Store) ThumbnailQueueStatus(ctx context.Context, now time.Time) (ports.ThumbnailQueueStatus, error) {
	status := ports.ThumbnailQueueStatus{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, entry := range []struct {
			state  ports.ThumbnailJobStatus
			target *int64
		}{{ports.ThumbnailJobPending, &status.Pending}, {ports.ThumbnailJobFailed, &status.Failed}, {ports.ThumbnailJobCompleted, &status.Completed}} {
			if err := tx.Model(&thumbnailJobModel{}).Where(&thumbnailJobModel{Status: string(entry.state)}).Count(entry.target).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&thumbnailJobModel{}).Where(&thumbnailJobModel{Status: string(ports.ThumbnailJobPending)}).Where(clause.Gt{Column: "claimed_until", Value: now}).Count(&status.Leased).Error; err != nil {
			return err
		}
		var oldest thumbnailJobModel
		err := tx.Where(&thumbnailJobModel{Status: string(ports.ThumbnailJobPending)}).Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}}).First(&oldest).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			status.OldestPendingAt = oldest.CreatedAt
		}
		var backfill thumbnailBackfillModel
		err = tx.First(&backfill, int(media.CurrentThumbnailRevision)).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			status.BackfillComplete = backfill.Complete
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return status, err
}
func (s Store) RetryFailedThumbnailJobs(ctx context.Context, limit int, now time.Time) (int, error) {
	if limit < 1 || limit > 1000 || now.IsZero() {
		return 0, errors.New("invalid thumbnail retry batch")
	}
	retried := 0
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var jobs []thumbnailJobModel
		if err := tx.Clauses(skipLockedForUpdate()).Where(&thumbnailJobModel{Status: string(ports.ThumbnailJobFailed)}).Order(clause.OrderByColumn{Column: clause.Column{Name: "updated_at"}}).Order(clause.OrderByColumn{Column: clause.Column{Name: "attachment_id"}}).Limit(limit).Find(&jobs).Error; err != nil {
			return err
		}
		for _, job := range jobs {
			if err := tx.Model(&job).Updates(map[string]any{"status": string(ports.ThumbnailJobPending), "attempts": 0, "failure": "", "claim_id": "", "claimed_until": nil, "next_attempt_at": now, "updated_at": now}).Error; err != nil {
				return err
			}
			retried++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return retried, nil
}
