package gormstore

import (
	"context"
	"os"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

func TestPostgresVoiceConfigurationUsesProductionMigrations(t *testing.T) {
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
	store := NewStore(db)
	ctx := context.Background()
	if _, found, err := store.VoiceProviderConfiguration(ctx, tenant.ID("unconfigured")); err != nil || found {
		t.Fatalf("unconfigured tenant must be a normal miss: found=%v err=%v", found, err)
	}
	cleanup := func() {
		for _, model := range []any{&voiceProviderConfigurationModel{}, &auditRecordModel{}, &tenantModel{}} {
			column := "tenant_id"
			if _, ok := model.(*tenantModel); ok {
				column = "id"
			}
			if err := db.Where(column+" IN ?", []string{"tenant-home", "tenant-other"}).Delete(model).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	t.Run("round_trip_and_tenant_scope", func(t *testing.T) { verifyVoiceConfigurationRoundTrip(t, ctx, store) })
	cleanup()
	t.Run("audit_rollback", func(t *testing.T) { verifyVoiceConfigurationAuditRollback(t, ctx, store) })
}
