package gormstore

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func (s Store) BackfillThumbnailJobs(ctx context.Context, limit int, now time.Time) (ports.ThumbnailBackfillProgress, error) {
	progress := ports.ThumbnailBackfillProgress{}
	if limit < 1 || limit > 1000 || now.IsZero() {
		return progress, errors.New("invalid thumbnail backfill batch")
	}
	now = now.UTC().Truncate(time.Microsecond)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		state := thumbnailBackfillModel{Revision: int(media.CurrentThumbnailRevision), UpdatedAt: now}
		// The insert also obtains SQLite's writer lock before reading cursor state.
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&state).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&state, state.Revision).Error; err != nil {
			return err
		}
		progress.Cursor = media.ID(state.Cursor)
		progress.Complete = state.Complete
		if state.Complete {
			return nil
		}
		var attachments []attachmentModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(clause.Gt{Column: "id", Value: state.Cursor}).Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Limit(limit).Find(&attachments).Error; err != nil {
			return err
		}
		for _, model := range attachments {
			attachment, valid := model.toDomain()
			if !valid {
				return errors.New("invalid attachment during thumbnail backfill")
			}
			if attachment.ContentType.IsImage() {
				job, err := media.NewThumbnailJob(attachment, media.ThumbnailJobBackfill, now)
				if err != nil {
					return err
				}
				queued := thumbnailJobModelFromDomain(job)
				inserted := tx.Omit(clause.Associations).Clauses(clause.OnConflict{DoNothing: true}).Create(&queued)
				if inserted.Error != nil {
					return inserted.Error
				}
				progress.Enqueued += int(inserted.RowsAffected)
			}
			progress.Scanned++
			progress.Cursor = attachment.ID
		}
		progress.Complete = len(attachments) < limit
		return tx.Model(&state).Updates(map[string]any{"cursor": progress.Cursor.String(), "complete": progress.Complete, "updated_at": now}).Error
	})
	if err != nil {
		return ports.ThumbnailBackfillProgress{}, err
	}
	return progress, nil
}
