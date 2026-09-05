package gormstore

import (
	"context"
	"os"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/testsupport/evaluationrun"
	"gorm.io/gorm/clause"
)

func TestPostgresEvaluationRunMigrationsAndClaims(t *testing.T) {
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
	if err := runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		for _, model := range []any{&evaluationRunModel{}, &auditRecordModel{}, &tenantModel{}} {
			column := "tenant_id"
			if _, ok := model.(*tenantModel); ok {
				column = "id"
			}
			if err := db.Where(clause.Eq{Column: column, Value: evaluationrun.TenantID}).Delete(model).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	ctx := context.Background()
	store := NewStore(db)
	saveTenant(t, ctx, store, evaluationrun.TenantID, "Home")
	evaluationrun.Verify(t, store)
	evaluationrun.VerifyConcurrentClaims(t, store)
	verifyEvaluationRunTimestampPrecision(t, store)
}
