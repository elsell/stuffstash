package gormstore

import (
	"context"
	"testing"

	searchapp "github.com/stuffstash/stuff-stash/internal/app/search"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/search"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/gorm"
)

func BenchmarkPostgresSearchPaths(b *testing.B) {
	db := searchBenchmarkDatabase(b)
	for _, shared := range []bool{true, false} {
		name := "distinct"
		if shared {
			name = "shared"
		}
		b.Run(name, func(b *testing.B) {
			tenantID, inventoryID := seedSearchBenchmark(b, db, 200)
			const resultCount, depth = 20, 4
			expectedPaths := map[string][]string{}
			for child := 1; child <= resultCount; child++ {
				firstAncestor := resultCount + (child-1)*depth + 1
				if shared {
					firstAncestor = resultCount + 1
				}

				expected := make([]string, 0, depth)
				for level := depth - 1; level >= 0; level-- {
					expected = append(expected, searchBenchmarkID(200, firstAncestor+level))
				}
				expectedPaths[searchBenchmarkID(200, child)] = expected
				for level := 0; level < depth; level++ {
					var parent any
					if level < depth-1 {
						parent = searchBenchmarkID(200, firstAncestor+level+1)
					}
					if err := db.Model(&assetModel{}).Where(&assetModel{ID: searchBenchmarkID(200, firstAncestor+level), TenantID: tenantID.String()}).Updates(map[string]any{"parent_asset_id": parent, "kind": "container"}).Error; err != nil {
						b.Fatal(err)
					}
				}
				if err := db.Model(&assetModel{}).Where(&assetModel{ID: searchBenchmarkID(200, child), TenantID: tenantID.String()}).Updates(map[string]any{"title": "Path match", "parent_asset_id": searchBenchmarkID(200, firstAncestor)}).Error; err != nil {
					b.Fatal(err)
				}
			}
			store := NewStore(db)
			query, _ := search.NewQuery("Path match")
			read := func() {
				results, err := store.SearchAssets(context.Background(), tenantID, []inventory.InventoryID{inventoryID}, ports.AssetSearchPageRequest{Query: query, Mode: search.ModeFuzzy, Limit: resultCount})
				if err != nil {
					b.Fatal(err)
				}
				results, err = searchapp.WithAncestorPaths(context.Background(), store, results)
				if err != nil || len(results) != resultCount {
					b.Fatalf("path search failed: %v, %d", err, len(results))
				}
				for _, result := range results {
					if len(result.AncestorPath) != depth {
						b.Fatalf("incomplete path: %+v", result.AncestorPath)
					}
					expected := expectedPaths[result.Asset.ID.String()]
					if len(expected) != depth {
						b.Fatal("unexpected search result")
					}
					for index, ancestor := range result.AncestorPath {
						if ancestor.ID.String() != expected[index] {
							b.Fatalf("incorrect ancestor chain: %+v, expected %v", result.AncestorPath, expected)
						}
					}

				}
			}
			read()
			var queries, rows int64
			const callback = "benchmark:search_paths"
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
			for index := 0; index < b.N; index++ {
				read()
			}
			b.StopTimer()
			b.ReportMetric(float64(queries)/float64(b.N), "queries/op")
			b.ReportMetric(float64(rows)/float64(b.N), "rows/op")
		})
	}
}
