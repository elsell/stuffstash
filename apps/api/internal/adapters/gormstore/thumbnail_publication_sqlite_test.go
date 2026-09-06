package gormstore

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteWALThumbnailPublicationBlocksDeletion(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite("file:" + filepath.Join(t.TempDir(), "publication.db") + "?_journal_mode=WAL&_busy_timeout=50")
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	pool.SetMaxOpenConns(4)
	if err := Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	store, attachment, record, job := thumbnailQueueFixtureInStore(t, ctx, NewStore(db))
	if err := store.SaveAttachment(ctx, attachment, record, &job); err != nil {
		t.Fatal(err)
	}
	guard, err := NewThumbnailPublicationGuard(store, &publicationClock{now: job.CreatedAt})
	if err != nil {
		t.Fatal(err)
	}
	entered, finish := make(chan struct{}), make(chan struct{})
	published := make(chan error, 1)
	go func() {
		published <- guard.Publish(ctx, attachment, nil, func(context.Context) error { close(entered); <-finish; return nil })
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		close(finish)
		t.Fatal("publisher did not enter")
	}
	deleting, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	err = store.db.WithContext(deleting).Delete(&attachmentModel{ID: attachment.ID.String()}).Error
	cancel()
	close(finish)
	if publicationErr := <-published; publicationErr != nil {
		t.Fatal(publicationErr)
	}
	if err == nil {
		t.Fatal("WAL deletion committed while publisher held lifecycle ownership")
	}
	if err := store.db.Delete(&attachmentModel{ID: attachment.ID.String()}).Error; err != nil {
		t.Fatal("deletion remained blocked after publication", err)
	}
}
