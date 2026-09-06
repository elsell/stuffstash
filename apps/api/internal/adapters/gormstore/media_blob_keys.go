package gormstore

import (
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type mediaBlobKeyModel struct {
	StorageKey string `gorm:"primaryKey;size:512"`
}

func (mediaBlobKeyModel) TableName() string { return "media_blob_keys" }

func reserveMediaBlobKey(tx *gorm.DB, key media.StorageKey) error {
	if key == "" {
		return ports.ErrConflict
	}
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&mediaBlobKeyModel{StorageKey: key.String()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ports.ErrConflict
	}
	return nil
}

const mediaBlobKeySeedBatchSize = 100

func seedMediaBlobKeys(db *gorm.DB) error {
	var attachments []attachmentModel
	if err := db.Select("id", "storage_key").FindInBatches(&attachments, mediaBlobKeySeedBatchSize, func(tx *gorm.DB, _ int) error {
		keys := make([]mediaBlobKeyModel, 0, len(attachments))
		for _, attachment := range attachments {
			keys = append(keys, mediaBlobKeyModel{StorageKey: attachment.StorageKey})
		}
		return tx.Session(&gorm.Session{NewDB: true}).Clauses(clause.OnConflict{DoNothing: true}).Create(&keys).Error
	}).Error; err != nil {
		return err
	}
	var deletions []blobDeletionEventModel
	return db.Select("id", "storage_key").FindInBatches(&deletions, mediaBlobKeySeedBatchSize, func(tx *gorm.DB, _ int) error {
		keys := make([]mediaBlobKeyModel, 0, len(deletions))
		for _, deletion := range deletions {
			keys = append(keys, mediaBlobKeyModel{StorageKey: deletion.StorageKey})
		}
		return tx.Session(&gorm.Session{NewDB: true}).Clauses(clause.OnConflict{DoNothing: true}).Create(&keys).Error
	}).Error
}
