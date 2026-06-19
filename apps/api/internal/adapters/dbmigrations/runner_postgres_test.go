package dbmigrations

import (
	"os"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/gormstore"
)

func TestPostgresRunnerAppliesAndReportsNoopMigrations(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set STUFF_STASH_TEST_POSTGRES_DSN to run Postgres migration verification")
	}

	db, err := gormstore.OpenPostgres(dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("postgres db handle: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatalf("close postgres: %v", err)
		}
	})

	runner := NewRunner(db)
	if err := runner.Up(); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	status, err := runner.Status()
	if err != nil {
		t.Fatalf("migration status: %v", err)
	}
	if status.Empty || status.Dirty || status.Version != status.Latest {
		t.Fatalf("expected clean current migration status, got %+v", status)
	}

	if err := runner.Up(); err != nil {
		t.Fatalf("migrate up no-op: %v", err)
	}
	status, err = runner.Status()
	if err != nil {
		t.Fatalf("migration status after no-op: %v", err)
	}
	if status.Empty || status.Dirty || status.Version != status.Latest {
		t.Fatalf("expected clean current migration status after no-op, got %+v", status)
	}
}
