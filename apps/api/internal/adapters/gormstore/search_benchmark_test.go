package gormstore

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

func BenchmarkPostgresSearch(b *testing.B) {
	db := searchBenchmarkDatabase(b)
	for _, volume := range []int{100, 10000} {
		b.Run(fmt.Sprintf("assets_%d", volume), func(b *testing.B) {
			tenantID, inventoryID := seedSearchBenchmark(b, db, volume)
			store := NewStore(db)
			for _, scenario := range []struct {
				name, query string
				after       int
				count       int
			}{
				{"broad", "Inventory item", 0, 20},
				{"selective", "needle", 0, 1},
				{"empty", "not-present", 0, 0},
				{"attachment", "manual", 0, 20},
				{"next_page", "Inventory item", 20, 20},
			} {
				b.Run(scenario.name, func(b *testing.B) {
					query, _ := search.NewQuery(scenario.query)
					page := ports.AssetSearchPageRequest{Query: query, Mode: search.ModeFuzzy, Limit: 20}
					if scenario.after > 0 {
						page.AfterResultKey = inventoryID.String() + ":" + searchBenchmarkID(volume, scenario.after)
					}
					read := func() {
						items, err := store.SearchAssets(context.Background(), tenantID, []inventory.InventoryID{inventoryID}, page)
						if err != nil {
							b.Fatal(err)
						}
						if len(items) != scenario.count {
							b.Fatalf("expected %d results, got %d", scenario.count, len(items))
						}
						for _, item := range items {
							if item.TenantID != tenantID || item.Inventory.ID != inventoryID || item.CursorKey() <= page.AfterResultKey || len(item.Matches) == 0 {
								b.Fatal("invalid search result")
							}
						}
					}
					read()
					var queries, rows int64
					const callback = "benchmark:search_reads"
					if err := db.Callback().Query().After("gorm:query").Register(callback, func(tx *gorm.DB) {
						if tx.DryRun {
							return
						}
						queries++
						if tx.RowsAffected > 0 {
							rows += tx.RowsAffected
						}
					}); err != nil {
						b.Fatal(err)
					}
					b.Cleanup(func() { _ = db.Callback().Query().Remove(callback) })
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						read()
					}
					b.StopTimer()
					b.ReportMetric(float64(queries)/float64(b.N), "queries/op")
					b.ReportMetric(float64(rows)/float64(b.N), "rows/op")
				})
			}
		})
	}
}

func searchBenchmarkID(volume, index int) string { return fmt.Sprintf("%026d", volume*100000+index) }

func seedSearchBenchmark(b testing.TB, db *gorm.DB, volume int) (tenant.ID, inventory.InventoryID) {
	b.Helper()
	tenantID := tenant.ID(searchBenchmarkID(volume, 0))
	inventoryID := inventory.InventoryID(tenantID)
	if err := db.Create(&tenantModel{ID: tenantID.String(), Name: "Benchmark"}).Error; err != nil {
		b.Fatal(err)
	}
	if err := db.Create(&inventoryModel{ID: inventoryID.String(), TenantID: tenantID.String(), Name: "Benchmark"}).Error; err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		for _, model := range []any{&attachmentModel{}, &assetModel{}, &inventoryModel{}} {
			if err := db.Where(map[string]any{"tenant_id": tenantID.String()}).Delete(model).Error; err != nil {
				b.Error(err)
			}
		}
		if err := db.Delete(&tenantModel{ID: tenantID.String()}).Error; err != nil {
			b.Error(err)
		}
	})
	assets := make([]assetModel, 0, volume)
	attachments := make([]attachmentModel, 0, volume*2)
	now := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	for index := 1; index <= volume; index++ {
		id := searchBenchmarkID(volume, index)
		title := fmt.Sprintf("Inventory item %d", index)
		if index == volume {
			title += " needle"
		}
		assets = append(assets, assetModel{ID: id, TenantID: tenantID.String(), InventoryID: inventoryID.String(), Kind: "item", Title: title, CustomFields: "{}", LifecycleState: "active", CreatedAt: now, UpdatedAt: now})
		for photo := 0; photo < 2; photo++ {
			attachmentID := searchBenchmarkID(volume, volume+index*2+photo)
			attachments = append(attachments, attachmentModel{ID: attachmentID, TenantID: tenantID.String(), InventoryID: inventoryID.String(), AssetID: id, StorageKey: attachmentID, FileName: "manual.jpg", ContentType: "image/jpeg", SizeBytes: 1000, SHA256: strings.Repeat("a", 64), LifecycleState: "active", CreatedAt: now, UpdatedAt: now})
		}
	}
	if err := db.CreateInBatches(assets, 100).Error; err != nil {
		b.Fatal(err)
	}
	if err := db.CreateInBatches(attachments, 100).Error; err != nil {
		b.Fatal(err)
	}
	return tenantID, inventoryID
}

func searchBenchmarkDatabase(b *testing.B) *gorm.DB {
	dsn := os.Getenv("STUFF_STASH_TEST_POSTGRES_DSN")
	if dsn == "" {
		b.Skip("requires an isolated benchmark PostgreSQL database")
	}
	db, err := OpenPostgres(dsn)
	if err != nil {
		b.Fatal(err)
	}
	connection, err := db.DB()
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = connection.Close() })
	if err := runEmbeddedPostgresMigrations(db); err != nil {
		b.Fatal(err)
	}
	return db
}
