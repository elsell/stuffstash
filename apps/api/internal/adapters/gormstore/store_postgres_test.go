package gormstore

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"github.com/stuffstash/stuff-stash/migrations"
	"gorm.io/gorm"
)

func TestPostgresStoreClaimsOutboxEventOnceAcrossWorkers(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set STUFF_STASH_TEST_POSTGRES_DSN to run Postgres outbox concurrency verification")
	}

	ctx := context.Background()
	db, err := OpenPostgres(dsn)
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
	if err := runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatalf("migrate postgres: %v", err)
	}

	store := NewStore(db)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	eventID := "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	cleanupAuthorizationOutboxTestRows(t, ctx, store, eventID, tenantID)
	saveTenantWithOutbox(t, ctx, store, eventID, tenantID, "Concurrency Home")

	claims := make(chan string, 2)
	var wg sync.WaitGroup
	for _, claimID := range []string{"claim-one", "claim-two"} {
		wg.Add(1)
		go func(claimID string) {
			defer wg.Done()
			events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, claimID, 1, time.Now(), time.Now().Add(time.Minute))
			if err != nil {
				t.Errorf("claim %s: %v", claimID, err)
				return
			}
			for _, event := range events {
				claims <- event.ClaimID
			}
		}(claimID)
	}
	wg.Wait()
	close(claims)

	claimedBy := []string{}
	for claimID := range claims {
		claimedBy = append(claimedBy, claimID)
	}
	if len(claimedBy) != 1 {
		t.Fatalf("expected exactly one worker to claim event, got %+v", claimedBy)
	}

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-three", 1, time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim while lease active: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected active lease to hide event from third worker, got %+v", events)
	}
}

func TestPostgresStorePersistsAssetCustomFieldsAsJSONB(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set STUFF_STASH_TEST_POSTGRES_DSN to run Postgres asset persistence verification")
	}

	ctx := context.Background()
	db, err := OpenPostgres(dsn)
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
	if err := runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatalf("migrate postgres: %v", err)
	}

	store := NewStore(db)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FB0")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FB1")
	assetID := "01ARZ3NDEKTSV4RRFFQ69G5FB2"
	cleanupAssetTestRows(t, ctx, store, tenantID, inventoryID, assetID)
	saveTenant(t, ctx, store, tenantID, "Postgres Assets")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	item := assetItem(assetID, tenantID.String(), inventoryID.String(), asset.KindItem, "")
	customFields, ok := asset.NewCustomFields(map[string]any{"serial": "abc", "count": float64(2)})
	if !ok {
		t.Fatalf("expected valid custom fields")
	}
	item.CustomFields = customFields

	if err := store.CreateAsset(ctx, item, postgresAuditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FB7", tenantID, inventoryID, audit.ActionAssetCreated), nil); err != nil {
		t.Fatalf("create asset: %v", err)
	}

	found, ok, err := store.AssetByID(ctx, tenantID, inventoryID, item.ID)
	if err != nil {
		t.Fatalf("find asset: %v", err)
	}
	values := found.CustomFields.Values()
	if !ok || values["serial"] != "abc" || values["count"] != float64(2) {
		t.Fatalf("expected JSONB custom fields to round-trip, found=%t values=%+v", ok, values)
	}
}

func TestPostgresStoreRestrictsInventoryDeleteWithAuditRecords(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set STUFF_STASH_TEST_POSTGRES_DSN to run Postgres audit FK verification")
	}

	ctx := context.Background()
	db, err := OpenPostgres(dsn)
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
	if err := runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatalf("migrate postgres: %v", err)
	}

	store := NewStore(db)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FB8")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FB9")
	cleanupAuditInventoryDeleteTestRows(t, ctx, store, tenantID, inventoryID)
	saveTenant(t, ctx, store, tenantID, "Postgres Audit")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Audit Inventory")
	if err := store.SaveAuditRecord(ctx, postgresAuditRecord(t, "01ARZ3NDEKTSV4RRFFQ69G5FBA", tenantID, inventoryID, audit.ActionAssetCreated)); err != nil {
		t.Fatalf("save audit record: %v", err)
	}

	err = store.db.WithContext(ctx).Delete(&inventoryModel{ID: inventoryID.String()}).Error
	if err == nil {
		t.Fatalf("expected inventory delete to be restricted by audit records")
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "23001" {
		t.Fatalf("expected foreign key violation, got %v", err)
	}
}

func TestPostgresStoreSerializesEffectiveCustomFieldKeysAcrossScopes(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set STUFF_STASH_TEST_POSTGRES_DSN to run Postgres custom field concurrency verification")
	}

	ctx := context.Background()
	db, err := OpenPostgres(dsn)
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
	if err := runEmbeddedPostgresMigrations(db); err != nil {
		t.Fatalf("migrate postgres: %v", err)
	}

	store := NewStore(db)
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FB3")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FB4")
	cleanupCustomFieldDefinitionTestRows(t, ctx, store, tenantID, inventoryID)
	saveTenant(t, ctx, store, tenantID, "Postgres Custom Fields")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Tools")

	tenantDefinition := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FB5", tenantID, "", customfield.ScopeTenant, "serial", customfield.FieldTypeText, nil)
	inventoryDefinition := customFieldDefinition(t, "01ARZ3NDEKTSV4RRFFQ69G5FB6", tenantID, inventoryID, customfield.ScopeInventory, "serial", customfield.FieldTypeText, nil)

	tx := store.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		t.Fatalf("begin transaction: %v", tx.Error)
	}
	defer tx.Rollback()

	if err := tx.Create(customFieldDefinitionModelForTest(tenantDefinition)).Error; err != nil {
		t.Fatalf("insert first definition in open transaction: %v", err)
	}

	secondInsert := make(chan error, 1)
	go func() {
		secondInsert <- store.db.WithContext(ctx).Create(customFieldDefinitionModelForTest(inventoryDefinition)).Error
	}()

	select {
	case err := <-secondInsert:
		t.Fatalf("expected second insert to wait for tenant/key advisory lock, got %v", err)
	case <-time.After(500 * time.Millisecond):
	}

	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit first definition: %v", err)
	}

	err = <-secondInsert
	if err == nil {
		t.Fatalf("expected second insert to conflict after first transaction commits")
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "23505" {
		t.Fatalf("expected postgres unique violation, got %v", err)
	}

	effective, err := store.ListInventoryCustomFieldDefinitions(ctx, tenantID, inventoryID, ports.CustomFieldDefinitionPageRequest{Limit: 10})
	if err != nil {
		t.Fatalf("list effective definitions: %v", err)
	}
	if len(effective) != 1 || effective[0].Key.String() != "serial" {
		t.Fatalf("expected one effective serial definition, got %+v", effective)
	}
}

func customFieldDefinitionModelForTest(definition customfield.Definition) *customFieldDefinitionModel {
	model := &customFieldDefinitionModel{
		ID:          definition.ID.String(),
		TenantID:    definition.TenantID.String(),
		Scope:       definition.Scope.String(),
		FieldKey:    definition.Key.String(),
		DisplayName: definition.DisplayName.String(),
		FieldType:   definition.Type.String(),
		EnumOptions: "[]",
	}
	if definition.InventoryID.String() != "" {
		inventoryID := definition.InventoryID.String()
		model.InventoryID = &inventoryID
	}
	return model
}

func cleanupAuthorizationOutboxTestRows(t *testing.T, ctx context.Context, store Store, eventID string, tenantID tenant.ID) {
	t.Helper()

	if err := store.db.WithContext(ctx).Delete(&authorizationOutboxEventModel{ID: eventID}).Error; err != nil {
		t.Fatalf("clean outbox row: %v", err)
	}
	if err := store.db.WithContext(ctx).Where(&auditRecordModel{TenantID: tenantID.String()}).Delete(&auditRecordModel{}).Error; err != nil {
		t.Fatalf("clean audit record rows: %v", err)
	}
	if err := store.db.WithContext(ctx).Delete(&tenantModel{ID: tenantID.String()}).Error; err != nil {
		t.Fatalf("clean tenant row: %v", err)
	}
}

func cleanupAssetTestRows(t *testing.T, ctx context.Context, store Store, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID string) {
	t.Helper()

	if err := store.db.WithContext(ctx).Delete(&assetModel{ID: assetID}).Error; err != nil {
		t.Fatalf("clean asset row: %v", err)
	}
	if err := store.db.WithContext(ctx).Where(&auditRecordModel{TenantID: tenantID.String()}).Delete(&auditRecordModel{}).Error; err != nil {
		t.Fatalf("clean audit record rows: %v", err)
	}
	if err := store.db.WithContext(ctx).Delete(&inventoryModel{ID: inventoryID.String()}).Error; err != nil {
		t.Fatalf("clean inventory row: %v", err)
	}
	if err := store.db.WithContext(ctx).Delete(&tenantModel{ID: tenantID.String()}).Error; err != nil {
		t.Fatalf("clean tenant row: %v", err)
	}
}

func cleanupCustomFieldDefinitionTestRows(t *testing.T, ctx context.Context, store Store, tenantID tenant.ID, inventoryID inventory.InventoryID) {
	t.Helper()

	if err := store.db.WithContext(ctx).Where(&customFieldDefinitionModel{TenantID: tenantID.String()}).Delete(&customFieldDefinitionModel{}).Error; err != nil {
		t.Fatalf("clean custom field definition rows: %v", err)
	}
	if err := store.db.WithContext(ctx).Where(&auditRecordModel{TenantID: tenantID.String()}).Delete(&auditRecordModel{}).Error; err != nil {
		t.Fatalf("clean audit record rows: %v", err)
	}
	if err := store.db.WithContext(ctx).Delete(&inventoryModel{ID: inventoryID.String()}).Error; err != nil {
		t.Fatalf("clean inventory row: %v", err)
	}
	if err := store.db.WithContext(ctx).Delete(&tenantModel{ID: tenantID.String()}).Error; err != nil {
		t.Fatalf("clean tenant row: %v", err)
	}
}

func cleanupAuditInventoryDeleteTestRows(t *testing.T, ctx context.Context, store Store, tenantID tenant.ID, inventoryID inventory.InventoryID) {
	t.Helper()

	if err := store.db.WithContext(ctx).Where(&auditRecordModel{TenantID: tenantID.String()}).Delete(&auditRecordModel{}).Error; err != nil {
		t.Fatalf("clean audit record rows: %v", err)
	}
	if err := store.db.WithContext(ctx).Delete(&inventoryModel{ID: inventoryID.String()}).Error; err != nil {
		t.Fatalf("clean inventory row: %v", err)
	}
	if err := store.db.WithContext(ctx).Delete(&tenantModel{ID: tenantID.String()}).Error; err != nil {
		t.Fatalf("clean tenant row: %v", err)
	}
}

func runEmbeddedPostgresMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	databaseDriver, err := migratepostgres.WithInstance(sqlDB, &migratepostgres.Config{})
	if err != nil {
		return err
	}
	sourceDriver, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return err
	}
	instance, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", databaseDriver)
	if err != nil {
		return err
	}
	if err := instance.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func postgresAuditRecord(t *testing.T, id string, tenantID tenant.ID, inventoryID inventory.InventoryID, action audit.Action) audit.Record {
	t.Helper()

	record, ok := audit.NewRecord(
		audit.ID(id),
		audit.TenantID(tenantID.String()),
		audit.InventoryID(inventoryID.String()),
		audit.PrincipalID("user-one"),
		action,
		audit.SourceAPI,
		audit.TargetAsset,
		"postgres-test",
		time.Now(),
		"",
		map[string]string{"test": "postgres"},
	)
	if !ok {
		t.Fatalf("expected valid audit record")
	}
	return record
}
