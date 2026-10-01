package inventoryexport

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestExportJSONPreservesPortableInventoryData(t *testing.T) {
	snapshot := ports.InventoryExportDocument{
		SchemaVersion: 1, ExportedAt: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		TenantID: "tenant", InventoryID: "inventory", InventoryName: "Home",
		Assets: []ports.InventoryExportAsset{{ID: "item", Title: "=SUM(1,2)\n日本語", Kind: "item", ParentAssetID: "box", LifecycleState: "archived", CustomFields: map[string]any{"amount": float64(2), "owned": true}, TagIDs: []string{"tag"}}},
	}
	body, err := (Encoder{}).Encode(context.Background(), snapshot, ports.InventoryExportJSON, 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ports.InventoryExportDocument
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.SchemaVersion != 1 || len(decoded.Assets) != 1 || decoded.Assets[0].Title != snapshot.Assets[0].Title || decoded.Assets[0].ParentAssetID != "box" || decoded.Assets[0].LifecycleState != "archived" || decoded.Assets[0].CustomFields["amount"] != float64(2) {
		t.Fatalf("lost inventory data: %#v", decoded)
	}
}

func TestExportCSVQuotesTextAndNeutralizesFormulas(t *testing.T) {
	titles := []string{"=SUM(1,2)", "  +12", "-cmd", "@SUM(A1)", "\tformula", "\rformula", "普通の箱, \"blue\"\nsecond line"}
	doc := ports.InventoryExportDocument{SchemaVersion: 1, Assets: []ports.InventoryExportAsset{}}
	for _, title := range titles {
		doc.Assets = append(doc.Assets, ports.InventoryExportAsset{ID: "item", Title: title, CustomFields: map[string]any{"text": "=kept inside JSON"}})
	}
	body, err := (Encoder{}).Encode(context.Background(), doc, ports.InventoryExportCSV, 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(string(body))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	titleColumn := -1
	for i, name := range rows[0] {
		if name == "title" {
			titleColumn = i
		}
	}
	if titleColumn < 0 || len(rows) != len(titles)+1 {
		t.Fatalf("invalid table: %#v", rows)
	}
	for i, title := range titles {
		expected := title
		if i < len(titles)-1 {
			expected = "'" + title
		}
		if rows[i+1][titleColumn] != expected {
			t.Fatalf("title %d: %q", i, rows[i+1][titleColumn])
		}
	}
}

func TestExportEncoderNeverReturnsPartialSuccess(t *testing.T) {
	doc := ports.InventoryExportDocument{SchemaVersion: 1, Assets: []ports.InventoryExportAsset{{Title: strings.Repeat("x", 2048)}}}
	for _, format := range []ports.InventoryExportFormat{ports.InventoryExportJSON, ports.InventoryExportCSV} {
		body, err := (Encoder{}).Encode(context.Background(), doc, format, 100)
		if !errors.Is(err, ports.ErrInventoryExportLimit) || len(body) != 0 {
			t.Fatalf("limit published partial %s: %d %v", format, len(body), err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		body, err = (Encoder{}).Encode(ctx, doc, format, 4096)
		if !errors.Is(err, context.Canceled) || len(body) != 0 {
			t.Fatalf("cancel published partial %s", format)
		}
	}
}

func TestExportExcludesStorageKeysAndBoundsEveryByte(t *testing.T) {
	doc := ports.InventoryExportDocument{SchemaVersion: 1, Assets: []ports.InventoryExportAsset{{Title: "photo", Attachments: []media.Attachment{{ID: "photo-id", StorageKey: "private-storage-key", FileName: "photo.jpg", ContentType: "image/jpeg", LifecycleState: media.LifecycleStateArchived}}}}}
	for _, format := range []ports.InventoryExportFormat{ports.InventoryExportJSON, ports.InventoryExportCSV} {
		content, err := (Encoder{}).Encode(context.Background(), doc, format, 10000)
		if err != nil || strings.Contains(string(content), "private-storage-key") || !strings.Contains(string(content), "photo.jpg") {
			t.Fatalf("unsafe export: %s %v", content, err)
		}
		for _, limit := range []int{len(content) - 1, len(content)} {
			bounded, err := (Encoder{}).Encode(context.Background(), doc, format, limit)
			if limit < len(content) {
				if !errors.Is(err, ports.ErrInventoryExportLimit) || len(bounded) != 0 {
					t.Fatalf("last byte bypassed size limit: %v", err)
				}
			} else if err != nil || len(bounded) != limit {
				t.Fatalf("exact size rejected: %v", err)
			}
		}
	}
}
