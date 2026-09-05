package gormstore

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/stuffstash/stuff-stash/internal/testsupport/workflowdiscovery"
	"gorm.io/gorm/clause"
	"os"
	"testing"
)

func TestPostgresConversationWorkflowDiscoveryBackfill(t *testing.T) {
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
	migration, err := postgresMigrationInstance(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.Migrate(44); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatal(err)
	}
	// Always restore the current schema, even when a backfill assertion fails.
	t.Cleanup(func() {
		if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			t.Error(err)
		}
	})
	cleanup := func() {
		for _, value := range []any{&conversationWorkflowRevisionModel{}, &conversationWorkflowModel{}, &auditRecordModel{}, &tenantModel{}} {
			column := "tenant_id"
			if _, ok := value.(*tenantModel); ok {
				column = "id"
			}
			if err := db.Where(clause.IN{Column: column, Values: []any{"discovery-home", "discovery-legacy"}}).Delete(value).Error; err != nil {
				t.Error(err)
			}
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	ctx := context.Background()
	store := NewStore(db)
	saveTenant(t, ctx, store, "discovery-legacy", "Legacy")
	for number := 1; number <= 2; number++ {
		id := "old-first"
		if number == 2 {
			id = "old-second"
		}
		revision := persistedWorkflowRevision(t, "discovery-legacy", id, number)
		row, err := workflowRevisionModel(revision)
		if err != nil {
			t.Fatal(err)
		}
		if number == 1 {
			head := conversationWorkflowModel{TenantID: "discovery-legacy", ID: "workflow-one", LatestRevision: 2, CreatedAt: row.CreatedAt, UpdatedAt: row.CreatedAt}
			if err := db.Omit("Name", "LatestRevisionID", clause.Associations).Create(&head).Error; err != nil {
				t.Fatal(err)
			}
		}
		if err := db.Omit(clause.Associations).Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatal(err)
	}
	head, found, err := store.WorkflowHead(ctx, "discovery-legacy", "workflow-one")
	if err != nil || !found || head.Name != "Home workflow" || head.LatestRevisionID != "old-second" {
		t.Fatalf("backfill: %+v %v", head, err)
	}
	saveTenant(t, ctx, store, "discovery-home", "Home")
	workflowdiscovery.Verify(t, store)
}
