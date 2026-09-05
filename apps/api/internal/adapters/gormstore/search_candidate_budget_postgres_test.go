package gormstore

import (
	"context"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

func TestPostgresSearchBoundsCandidateHydration(t *testing.T) {
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
	tenantID, inventoryID := seedSearchBenchmark(t, db, 10000)
	store := NewStore(db)
	for _, scenario := range []struct {
		query             string
		expected, maxRows int
	}{
		{"needle", 1, 10}, {"not-present", 0, 10}, {"Inventory item", 20, 500}, {"manual", 20, 500},
	} {
		t.Run(scenario.query, func(t *testing.T) {
			var rows int64
			const callback = "test:search_candidate_rows"
			if err := db.Callback().Query().After("gorm:query").Register(callback, func(tx *gorm.DB) {
				if tx.RowsAffected > 0 {
					rows += tx.RowsAffected
				}
			}); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Callback().Query().Remove(callback) })
			query, _ := search.NewQuery(scenario.query)
			items, err := store.SearchAssets(context.Background(), tenantID, []inventory.InventoryID{inventoryID}, ports.AssetSearchPageRequest{Query: query, Mode: search.ModeFuzzy, Limit: 20})
			if err != nil || len(items) != scenario.expected {
				t.Fatalf("search result count: %d, %v", len(items), err)
			}
			if rows > int64(scenario.maxRows) {
				t.Fatalf("search hydrated %d rows for %d results; budget %d", rows, len(items), scenario.maxRows)
			}
		})
	}
	// Candidate matching must remain a superset of Go's Unicode and literal rules.
	for index, title := range map[int]string{1: `Literal%_\ marker`, 2: "Kettle", 9000: "Café", 4: "INDIGO"} {
		if err := db.Model(&assetModel{}).Where(&assetModel{ID: searchBenchmarkID(10000, index), TenantID: tenantID.String()}).Update("title", title).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Model(&assetModel{}).Where(&assetModel{ID: searchBenchmarkID(10000, 5), TenantID: tenantID.String()}).Update("custom_fields", `{"nested":{"key":"value"}}`).Error; err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		query string
		index int
	}{
		{"%", 1}, {"_", 1}, {`\`, 1}, {"kettle", 2}, {"café", 9000}, {"indigo", 4}, {"map[key:value]", 5},
	} {
		query, _ := search.NewQuery(scenario.query)
		items, err := store.SearchAssets(context.Background(), tenantID, []inventory.InventoryID{inventoryID}, ports.AssetSearchPageRequest{Query: query, Mode: search.ModeFuzzy, Limit: 20})
		if err != nil || len(items) != 1 || items[0].Asset.ID.String() != searchBenchmarkID(10000, scenario.index) {
			t.Fatalf("query %q changed semantics: %+v, %v", scenario.query, items, err)
		}
	}
	// NUL is a valid domain query but cannot be passed as PostgreSQL text.
	for _, mode := range []search.Mode{search.ModeExact, search.ModeFuzzy} {
		query, ok := search.NewQuery("\x00")
		if !ok {
			t.Fatal("NUL query must remain valid")
		}
		items, err := store.SearchAssets(context.Background(), tenantID, []inventory.InventoryID{inventoryID}, ports.AssetSearchPageRequest{Query: query, Mode: mode, Limit: 20})
		if err != nil || len(items) != 0 {
			t.Fatalf("NUL query changed semantics: %+v, %v", items, err)
		}
	}
	// Opaque cursors use bytewise concatenated IDs, including prefix inventory IDs.
	inventoryIDs := []inventory.InventoryID{"search-prefix-a", "search-prefix-a0"}
	expected := []string{}
	for index, id := range inventoryIDs {
		if err := db.Create(&inventoryModel{ID: id.String(), TenantID: tenantID.String(), Name: "Prefix fixture"}).Error; err != nil {
			t.Fatal(err)
		}
		model := assetModel{ID: searchBenchmarkID(10000, 50000+index), TenantID: tenantID.String(), InventoryID: id.String(), Kind: "item", Title: "cursor-fixture", CustomFields: "{}", LifecycleState: "active"}
		if err := db.Create(&model).Error; err != nil {
			t.Fatal(err)
		}
		expected = append(expected, id.String()+":"+model.ID)
	}
	sort.Strings(expected)
	cursor := ""
	for _, key := range expected {
		query, _ := search.NewQuery("cursor-fixture")
		items, err := store.SearchAssets(context.Background(), tenantID, inventoryIDs, ports.AssetSearchPageRequest{Query: query, Mode: search.ModeFuzzy, Limit: 1, AfterResultKey: cursor})
		if err != nil || len(items) != 1 || items[0].CursorKey() != key {
			t.Fatalf("cursor order changed: %+v, %v; expected %s", items, err, key)
		}
		cursor = items[0].CursorKey()
	}
	// Inventory scopes must hold even when a sibling has the same matching field.
	query, _ := search.NewQuery("cursor-fixture")
	items, err := store.SearchAssets(context.Background(), tenantID, inventoryIDs[:1], ports.AssetSearchPageRequest{Query: query, Limit: 20})
	if err != nil || len(items) != 1 || !strings.HasPrefix(items[0].CursorKey(), inventoryIDs[0].String()+":") {
		t.Fatalf("inventory candidate scope escaped: %+v, %v", items, err)
	}

	// Type and tag subqueries must discover metadata-only matches as well.
	typeID, tagID := searchBenchmarkID(10000, 70000), searchBenchmarkID(10000, 70001)
	if err := db.Create(&customAssetTypeModel{ID: typeID, TenantID: tenantID.String(), Scope: "tenant", CursorKey: typeID, TypeKey: "calibration", DisplayName: "Calibration equipment", Description: "Metrology reference", LifecycleState: "active"}).Error; err != nil {
		t.Fatal(err)
	}
	targetID := searchBenchmarkID(10000, 6)
	if err := db.Model(&assetModel{}).Where(&assetModel{ID: targetID, TenantID: tenantID.String()}).Update("custom_asset_type_id", typeID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&assetTagModel{ID: tagID, TenantID: tenantID.String(), InventoryID: inventoryID.String(), Key: "inspected", DisplayName: "Inspection completed", LifecycleState: "active"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&assetTagAssignmentModel{TenantID: tenantID.String(), InventoryID: inventoryID.String(), AssetID: targetID, TagID: tagID}).Error; err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"calibration", "Calibration equipment", "Metrology reference", "inspected", "Inspection completed", "INDIGO"} {
		query, _ := search.NewQuery(text)
		items, err := store.SearchAssets(context.Background(), tenantID, []inventory.InventoryID{inventoryID}, ports.AssetSearchPageRequest{Query: query, Mode: search.ModeExact, Limit: 20})
		expectedID := targetID
		if text == "INDIGO" {
			expectedID = searchBenchmarkID(10000, 4)
		}
		if err != nil || len(items) != 1 || items[0].Asset.ID.String() != expectedID {
			t.Fatalf("exact metadata query %q lost match: %+v, %v", text, items, err)
		}
	}
	if err := db.Model(&assetTagModel{}).Where(&assetTagModel{ID: tagID, TenantID: tenantID.String()}).Update("lifecycle_state", "archived").Error; err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"inspected", "Inventory"} {
		query, _ := search.NewQuery(text)
		items, err := store.SearchAssets(context.Background(), tenantID, []inventory.InventoryID{inventoryID}, ports.AssetSearchPageRequest{Query: query, Mode: search.ModeExact, Limit: 20})
		if err != nil || len(items) != 0 {
			t.Fatalf("query %q accepted a non-match or archived tag: %+v, %v", text, items, err)
		}
	}

}
