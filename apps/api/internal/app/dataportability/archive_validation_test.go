package dataportability

import (
	"context"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func validArchiveDocument() ports.InventoryExportDocument {
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	return ports.InventoryExportDocument{SchemaVersion: 2, ExportedAt: now, TenantID: "source", InventoryID: "inventory", InventoryName: "Home", Tags: []assettag.Tag{{ID: "tag", Key: "medicine", DisplayName: "Medicine", LifecycleState: assettag.LifecycleStateActive, CreatedAt: now, UpdatedAt: now}}, CustomAssetTypes: []customfield.AssetType{{ID: "type", Scope: customfield.ScopeTenant, Key: "medicine", DisplayName: "Medicine", LifecycleState: customfield.AssetTypeLifecycleActive}}, CustomFieldDefinitions: []customfield.Definition{{ID: "field", Scope: customfield.ScopeTenant, Key: "strength", DisplayName: "Strength", Type: customfield.FieldTypeNumber, Applicability: customfield.ApplicabilityCustomAssetTypes, CustomAssetTypeIDs: []customfield.AssetTypeID{"type"}, LifecycleState: customfield.DefinitionLifecycleArchived}}, Assets: []ports.InventoryExportAsset{{ID: "parent", Title: "Shelf", Kind: "container", LifecycleState: "active", CreatedAt: now, UpdatedAt: now}, {ID: "child", Title: "Medicine", Kind: "item", ParentAssetID: "parent", CustomAssetTypeID: "type", CustomFields: map[string]any{"strength": float64(10)}, TagIDs: []string{"tag"}, LifecycleState: "archived", CreatedAt: now, UpdatedAt: now}}}
}
func TestArchiveValidationPreservesArchivedAssignmentsAndRejectsBadGraphs(t *testing.T) {
	if err := ValidateArchiveDocument(context.Background(), validArchiveDocument(), 100); err != nil {
		t.Fatal(err)
	}
	for name, change := range map[string]func(*ports.InventoryExportDocument){
		"noncanonical field key": func(d *ports.InventoryExportDocument) { d.CustomFieldDefinitions[0].Key = " strength " },
		"noncanonical type key":  func(d *ports.InventoryExportDocument) { d.CustomAssetTypes[0].Key = " medicine " },
		"noncanonical tag key":   func(d *ports.InventoryExportDocument) { d.Tags[0].Key = " medicine " },
		"ambiguous value key":    func(d *ports.InventoryExportDocument) { d.Assets[1].CustomFields[" strength "] = float64(20) },
		"cycle":                  func(d *ports.InventoryExportDocument) { d.Assets[0].ParentAssetID = "parent" },
		"missing parent":         func(d *ports.InventoryExportDocument) { d.Assets[1].ParentAssetID = "foreign" },
		"item parent":            func(d *ports.InventoryExportDocument) { d.Assets[0].Kind = "item" },
		"duplicate id":           func(d *ports.InventoryExportDocument) { d.Assets[1].ID = "parent" },
		"missing tag":            func(d *ports.InventoryExportDocument) { d.Assets[1].TagIDs = []string{"foreign"} },
		"missing type":           func(d *ports.InventoryExportDocument) { d.Assets[1].CustomAssetTypeID = "foreign" },
		"missing field target": func(d *ports.InventoryExportDocument) {
			d.CustomFieldDefinitions[0].CustomAssetTypeIDs = []customfield.AssetTypeID{"foreign"}
		},
		"wrong value type": func(d *ports.InventoryExportDocument) { d.Assets[1].CustomFields["strength"] = "ten" },
		"duplicate field key": func(d *ports.InventoryExportDocument) {
			f := d.CustomFieldDefinitions[0]
			f.ID = "other"
			d.CustomFieldDefinitions = append(d.CustomFieldDefinitions, f)
		},
		"invalid expiration": func(d *ports.InventoryExportDocument) {
			d.Assets[1].ExpirationDate = "2026-02-30"
			d.Assets[1].ExpirationPrecision = "day"
		},
	} {
		t.Run(name, func(t *testing.T) {
			d := validArchiveDocument()
			change(&d)
			if err := ValidateArchiveDocument(context.Background(), d, 100); err == nil {
				t.Fatal("invalid graph accepted")
			}
		})
	}
	if err := ValidateArchiveDocument(context.Background(), validArchiveDocument(), 3); err == nil {
		t.Fatal("record limit ignored")
	}
}
