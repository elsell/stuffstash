package gormstore

import (
	"gorm.io/gorm/clause"
	"os"
	"testing"
)

func TestPostgresPrintingScopeDeletionRetainsEvidence(t *testing.T) {
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
	t.Cleanup(func() {
		for _, model := range []any{&printingAttemptIndex{}, &printingJobModel{}, &printingReportModel{}, &printingBindingModel{}, &printingConnectorModel{}, &printingPrinterModel{}, &printSettingsModel{}, &auditRecordModel{}, &inventoryModel{}} {
			if err := db.Where(clause.Eq{Column: "tenant_id", Value: "tenant"}).Delete(model).Error; err != nil {
				t.Error(err)
			}
		}
		if err := db.Delete(&tenantModel{ID: "tenant"}).Error; err != nil {
			t.Error(err)
		}
	})
	exerciseInventoryPrintingDeletion(t, NewStore(db))
}
