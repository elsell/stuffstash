package inventoryarchive

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/assettag"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/media"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestMetadataRetainsRelationshipsWithoutStorageCredentials(t *testing.T) {
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	doc := ports.InventoryExportDocument{SchemaVersion: 2, ExportedAt: now, TenantID: "source", InventoryID: "inventory", InventoryName: "Home 日本", Tags: []assettag.Tag{{ID: "tag", Key: "medicine", DisplayName: "Medicine", Color: "#ABCDEF", LifecycleState: assettag.LifecycleStateArchived, CreatedAt: now, UpdatedAt: now}}, CustomAssetTypes: []customfield.AssetType{{ID: "type", Scope: customfield.ScopeTenant, Key: "medicine", DisplayName: "Medicine", ExpirationEnabled: true}}, CustomFieldDefinitions: []customfield.Definition{{ID: "field", Scope: customfield.ScopeTenant, Key: "strength", Type: customfield.FieldTypeNumber, Applicability: customfield.ApplicabilityCustomAssetTypes, CustomAssetTypeIDs: []customfield.AssetTypeID{"type"}}}, Assets: []ports.InventoryExportAsset{{ID: "asset", Title: "薬", ParentAssetID: "parent", CustomAssetTypeID: "type", CustomFields: map[string]any{"strength": float64(12.5)}, TagIDs: []string{"tag"}, CurrentCheckout: &asset.Checkout{ID: "checkout", CheckedOutAt: now, CreatedAt: now, UpdatedAt: now, CheckedOutByPrincipal: "source-actor", CheckoutDetails: "At home"}, Attachments: []media.Attachment{{ID: "photo", StorageKey: "SECRET-INTERNAL-KEY", FileName: "photo.jpg", SHA256: media.SHA256(strings.Repeat("a", 64)), SizeBytes: 12, CreatedAt: now}}}}}
	codec := MetadataCodec{}
	encoded, err := codec.EncodeMetadata(context.Background(), doc, 100000)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "SECRET") || strings.Contains(string(encoded), "StorageKey") {
		t.Fatal("storage key exposed")
	}
	decoded, err := codec.DecodeMetadata(context.Background(), encoded, 100000)
	if err != nil {
		t.Fatal(err)
	}
	a := decoded.Assets[0]
	if decoded.InventoryName != doc.InventoryName || a.Title != "薬" || a.ParentAssetID != "parent" || a.TagIDs[0] != "tag" || a.CustomFields["strength"] != json.Number("12.5") || a.Attachments[0].StorageKey != "" || a.CurrentCheckout.CheckedOutByPrincipal != "source-actor" || !a.CurrentCheckout.CreatedAt.Equal(now) || !decoded.CustomAssetTypes[0].ExpirationEnabled || decoded.CustomFieldDefinitions[0].CustomAssetTypeIDs[0] != "type" {
		t.Fatalf("metadata lost during roundtrip: %#v", decoded)
	}
}
func TestMetadataRejectsAmbiguousOrUnsupportedDocuments(t *testing.T) {
	codec := MetadataCodec{}
	for _, raw := range []string{
		`{"schemaVersion":2,"assets":[],"aſſets":[],"tags":[],"customAssetTypes":[],"customFieldDefinitions":[]}`,
		`{"schemaVersion":2,"schemaVersion":2,"assets":[],"tags":[],"customAssetTypes":[],"customFieldDefinitions":[]}`,
		`{"schemaVersion":1,"assets":[],"tags":[],"customAssetTypes":[],"customFieldDefinitions":[]}`,
		`{"schemaVersion":2,"assets":[],"tags":[],"customAssetTypes":[]}`,
		`{"schemaVersion":2,"assets":[],"tags":[],"customAssetTypes":[],"customFieldDefinitions":[],"storageKey":"secret"}`,
		`{"schemaVersion":2,"assets":[{"customFields":{"foo":1,"foo":2}}],"tags":[],"customAssetTypes":[],"customFieldDefinitions":[]}`,
	} {
		if _, err := codec.DecodeMetadata(context.Background(), []byte(raw), 10000); err == nil {
			t.Fatalf("accepted ambiguous metadata: %s", raw)
		}
	}
	if _, err := codec.DecodeMetadata(context.Background(), []byte(strings.Repeat(" ", 100)), 10); err == nil {
		t.Fatal("byte limit bypass")
	}
}
