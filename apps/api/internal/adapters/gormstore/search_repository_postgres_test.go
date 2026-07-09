package gormstore

import (
	"context"
	"os"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestPostgresStoreSearchAssetsReturnsInventoryScopedResultsAndWritesReadAudit(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set STUFF_STASH_TEST_POSTGRES_DSN to run Postgres search audit verification")
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
	authorizer := memory.NewAuthorizer()
	tenantID := tenant.ID("01KX2A1B7D5VYYA70W2NQ4TE5A")
	inventoryID := inventory.InventoryID("01KX2A1BEGDWWN7P149T1P3R06")
	assetID := "01KX2A34G3TS98AR86N4KQW9QK"
	createAuditID := "01KX2A3AR94F9DJWVEYEYWK31N"
	principal := identity.Principal{ID: identity.PrincipalID("oidc_oVHefizjP_fV2vLDR_PdhwYJq-36AQl_MNw5vGxvfSM")}

	cleanupSearchTestRows(t, ctx, store, tenantID)
	saveTenant(t, ctx, store, tenantID, "Household")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Home")
	if err := authorizer.GrantTenantOwner(ctx, principal, tenantID); err != nil {
		t.Fatalf("grant tenant owner: %v", err)
	}
	if err := authorizer.GrantInventoryOwner(ctx, principal, tenantID, inventoryID); err != nil {
		t.Fatalf("grant inventory owner: %v", err)
	}

	item := assetItem(assetID, tenantID.String(), inventoryID.String(), asset.KindItem, "")
	title, ok := asset.NewTitle("Spring lawn fertilizer in the garage")
	if !ok {
		t.Fatal("expected valid title")
	}
	item.Title = title
	if err := store.CreateAsset(ctx, item, postgresAuditRecord(t, createAuditID, tenantID, inventoryID, audit.ActionAssetCreated), nil); err != nil {
		t.Fatalf("create asset: %v", err)
	}

	application := app.New(app.Dependencies{
		Authorizer:       authorizer,
		Users:            store,
		Tenants:          store,
		Inventories:      store,
		Assets:           store,
		Search:           store,
		Attachments:      store,
		Audit:            store,
		DefaultPageLimit: 50,
		MaxPageLimit:     100,
	})

	result, err := application.SearchAssets(ctx, app.SearchAssetsInput{
		Principal:      principal,
		TenantID:       tenantID,
		InventoryIDs:   []inventory.InventoryID{inventoryID},
		Source:         audit.SourceAPI,
		RequestID:      "search-request-one",
		Query:          "fertilizer",
		LifecycleState: string(ports.AssetLifecycleFilterActive),
		Mode:           search.ModeFuzzy.String(),
		CheckoutState:  string(ports.AssetCheckoutStateFilterAny),
		Limit:          20,
	})
	if err != nil {
		t.Fatalf("search assets: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Asset.ID != item.ID {
		t.Fatalf("expected fertilizer search result, got %+v", result.Items)
	}

	records, err := store.ListInventoryAuditRecords(ctx, tenantID, inventoryID, ports.AuditRecordPageRequest{Limit: 20})
	if err != nil {
		t.Fatalf("list inventory audit records: %v", err)
	}
	if !auditRecordsIncludeAction(records, audit.ActionAssetSearched) {
		t.Fatalf("expected inventory search read audit record, got %+v", records)
	}
}

func TestPostgresStoreSearchAssetsMatchesPersistedMetadata(t *testing.T) {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set STUFF_STASH_TEST_POSTGRES_DSN to run Postgres search verification")
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
	tenantID := tenant.ID("01ARZ3NDEKTSV4RRFFQ69G5FC0")
	inventoryID := inventory.InventoryID("01ARZ3NDEKTSV4RRFFQ69G5FC1")
	medicineTypeID := "01ARZ3NDEKTSV4RRFFQ69G5FC2"
	drillID := "01ARZ3NDEKTSV4RRFFQ69G5FC3"
	aspirinID := "01ARZ3NDEKTSV4RRFFQ69G5FC4"
	attachmentID := "01ARZ3NDEKTSV4RRFFQ69G5FC5"
	medicineTypeAuditID := "01ARZ3NDEKTSV4RRFFQ69G5FC6"
	drillAuditID := "01ARZ3NDEKTSV4RRFFQ69G5FC7"
	attachmentAuditID := "01ARZ3NDEKTSV4RRFFQ69G5FC8"
	aspirinAuditID := "01ARZ3NDEKTSV4RRFFQ69G5FC9"
	cleanupSearchTestRows(t, ctx, store, tenantID)
	saveTenant(t, ctx, store, tenantID, "Postgres Search")
	saveInventory(t, ctx, store, inventoryID.String(), tenantID, "Search Inventory")

	medicineType := customAssetType(t, medicineTypeID, tenantID.String(), inventoryID.String(), customfield.ScopeInventory, "medicine")
	if err := store.SaveCustomAssetType(ctx, medicineType, postgresAuditRecord(t, medicineTypeAuditID, tenantID, inventoryID, audit.ActionCustomAssetTypeCreated)); err != nil {
		t.Fatalf("save custom asset type: %v", err)
	}

	drill := assetItem(drillID, tenantID.String(), inventoryID.String(), asset.KindItem, "")
	drillTitle, _ := asset.NewTitle("Cordless Drill")
	drill.Title = drillTitle
	drillFields, ok := asset.NewCustomFields(map[string]any{"serial": "bag-42"})
	if !ok {
		t.Fatalf("expected valid custom fields")
	}
	drill.CustomFields = drillFields
	if err := store.CreateAsset(ctx, drill, postgresAuditRecord(t, drillAuditID, tenantID, inventoryID, audit.ActionAssetCreated), nil); err != nil {
		t.Fatalf("create drill: %v", err)
	}
	saveSearchAttachmentWithAuditID(t, ctx, store, attachmentID, attachmentAuditID, drill, "warranty-card.png", media.ContentTypePNG)

	aspirin := assetItem(aspirinID, tenantID.String(), inventoryID.String(), asset.KindItem, "")
	aspirinTitle, _ := asset.NewTitle("Aspirin")
	aspirin.Title = aspirinTitle
	aspirin.CustomAssetTypeID = asset.CustomAssetTypeID(medicineType.ID.String())
	if err := store.CreateAsset(ctx, aspirin, postgresAuditRecord(t, aspirinAuditID, tenantID, inventoryID, audit.ActionAssetCreated), nil); err != nil {
		t.Fatalf("create aspirin: %v", err)
	}

	attachmentResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{inventoryID}, "warranty", search.ModeFuzzy, ports.AssetLifecycleFilterActive, "", "", 10)
	if len(attachmentResults) != 1 || attachmentResults[0].Asset.ID != drill.ID || attachmentResults[0].Matches[0].Field != search.MatchFieldAttachmentFileName {
		t.Fatalf("expected Postgres attachment search to find drill, got %+v", attachmentResults)
	}

	customFieldResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{inventoryID}, "bag-42", search.ModeExact, ports.AssetLifecycleFilterActive, "", "", 10)
	if len(customFieldResults) != 1 || customFieldResults[0].Asset.ID != drill.ID || customFieldResults[0].Matches[0].Field != search.MatchFieldCustomField {
		t.Fatalf("expected Postgres exact custom field search to find drill, got %+v", customFieldResults)
	}

	typeResults := searchPersistedAssets(t, ctx, store, tenantID, []inventory.InventoryID{inventoryID}, "medicine", search.ModeExact, ports.AssetLifecycleFilterActive, medicineType.ID.String(), "", 10)
	if len(typeResults) != 1 || typeResults[0].Asset.ID != aspirin.ID || typeResults[0].Matches[0].Field != search.MatchFieldCustomAssetTypeKey {
		t.Fatalf("expected Postgres custom asset type search to find aspirin, got %+v", typeResults)
	}
}

func cleanupSearchTestRows(t *testing.T, ctx context.Context, store Store, tenantID tenant.ID) {
	t.Helper()

	if err := store.db.WithContext(ctx).Where(&attachmentModel{TenantID: tenantID.String()}).Delete(&attachmentModel{}).Error; err != nil {
		t.Fatalf("clean attachment rows: %v", err)
	}
	if err := store.db.WithContext(ctx).Where(&assetModel{TenantID: tenantID.String()}).Delete(&assetModel{}).Error; err != nil {
		t.Fatalf("clean asset rows: %v", err)
	}
	if err := store.db.WithContext(ctx).Where(&customAssetTypeModel{TenantID: tenantID.String()}).Delete(&customAssetTypeModel{}).Error; err != nil {
		t.Fatalf("clean custom asset type rows: %v", err)
	}
	if err := store.db.WithContext(ctx).Where(&auditRecordModel{TenantID: tenantID.String()}).Delete(&auditRecordModel{}).Error; err != nil {
		t.Fatalf("clean audit record rows: %v", err)
	}
	if err := store.db.WithContext(ctx).Where(&inventoryModel{TenantID: tenantID.String()}).Delete(&inventoryModel{}).Error; err != nil {
		t.Fatalf("clean inventory rows: %v", err)
	}
	if err := store.db.WithContext(ctx).Delete(&tenantModel{ID: tenantID.String()}).Error; err != nil {
		t.Fatalf("clean tenant row: %v", err)
	}
}
