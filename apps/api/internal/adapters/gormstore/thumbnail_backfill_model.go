package gormstore

import "time"

type thumbnailBackfillModel struct {
	Revision  int       `gorm:"primaryKey"`
	Cursor    string    `gorm:"not null;size:26"`
	Complete  bool      `gorm:"not null"`
	UpdatedAt time.Time `gorm:"autoUpdateTime:false"`
}

func (thumbnailBackfillModel) TableName() string { return "thumbnail_backfills" }
