package gormstore

import (
	"context"
	"fmt"
	"gorm.io/gorm/clause"
	"testing"
)

func TestMediaBlobKeyUpgradeSeedsBatchesAndPreservesReservations(t *testing.T) {
	ctx := context.Background()
	store, attachment, record, job := thumbnailQueueFixture(t, ctx)
	if err := store.SaveAttachment(ctx, attachment, record, &job); err != nil {
		t.Fatal(err)
	}
	var original attachmentModel
	if err := store.db.First(&original, &attachmentModel{ID: attachment.ID.String()}).Error; err != nil {
		t.Fatal(err)
	}
	for i := range 105 {
		copy := original
		copy.ID = fmt.Sprintf("legacy-image-%03d", i)
		copy.StorageKey = fmt.Sprintf("legacy-key-%03d", i)
		if err := store.db.Omit(clause.Associations).Create(&copy).Error; err != nil {
			t.Fatal(err)
		}
		key := fmt.Sprintf("deleted-key-%03d", i)
		if i == 0 {
			key = copy.StorageKey
		}
		if err := store.db.Create(&blobDeletionEventModel{ID: fmt.Sprintf("legacy-delete-%03d", i), StorageKey: key}).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := store.db.Migrator().DropTable(&mediaBlobKeyModel{}); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, store.db); err != nil {
		t.Fatal("upgrade failed", err)
	}
	var count int64
	if err := store.db.Model(&mediaBlobKeyModel{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 210 {
		t.Fatalf("seed lost keys or duplicated overlap: %d", count)
	}
	if err := store.db.Create(&mediaBlobKeyModel{StorageKey: "retired-only"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, store.db); err != nil {
		t.Fatal("repeat migration failed", err)
	}
	if err := store.db.Model(&mediaBlobKeyModel{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 211 {
		t.Fatal("repeat migration lost reservation", count)
	}
}
