package gormstore

import (
	"context"
	"os"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"gorm.io/gorm/clause"
)

func TestPostgresArchivePublicationPreservesArchivedReferences(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := OpenPostgres(dsn)
	if err != nil {
		t.Fatal(err)
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	if err = runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatal(err)
	}
	const tid = "archive-publish-tenant"
	const iid = "archive-publish-inventory"
	cleanup := func() {
		for _, model := range []any{&authorizationOutboxEventModel{}, &auditRecordModel{}, &attachmentModel{}, &assetModel{}, &customFieldDefinitionAssetTypeModel{}, &customFieldDefinitionModel{}, &customAssetTypeModel{}, &inventoryModel{}, &archiveJobModel{}, &tenantModel{}} {
			column := "tenant_id"
			if _, ok := model.(*tenantModel); ok {
				column = "id"
			}
			if err := db.Where(clause.Eq{Column: column, Value: tid}).Delete(model).Error; err != nil {
				t.Fatal(err)
			}
		}
		if err := db.Where(clause.Eq{Column: "storage_key", Value: tid + "/" + iid + "/pub-asset/pub-photo"}).Delete(&mediaBlobKeyModel{}).Error; err != nil {
			t.Fatal(err)
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	ctx := context.Background()
	s := NewStore(db)
	saveTenant(t, ctx, s, tid, "Home")
	input, clock := archivePublicationFixture(t, s, tid, iid, "archive-publish-job")
	if _, err = NewArchiveRestorePublisher(s, clock, 100).PublishArchiveRestore(ctx, input); err != nil {
		t.Fatal(err)
	}
	custom, found, err := s.CustomAssetTypeByID(ctx, tid, iid, "pub-type")
	if err != nil || !found || custom.LifecycleState != customfield.AssetTypeLifecycleArchived {
		t.Fatalf("archived type lost: %v", err)
	}
	field, found, err := s.CustomFieldDefinitionByID(ctx, tid, iid, "pub-field")
	if err != nil || !found || len(field.CustomAssetTypeIDs) != 1 || field.CustomAssetTypeIDs[0] != "pub-type" {
		t.Fatalf("archived target lost: %v", err)
	}
	item, found, err := s.AssetByID(ctx, tid, iid, "pub-asset")
	if err != nil || !found || item.CustomAssetTypeID != "pub-type" || item.LifecycleState != "archived" {
		t.Fatalf("asset reference lost: %v", err)
	}
}
