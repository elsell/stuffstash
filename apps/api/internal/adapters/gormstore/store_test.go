package gormstore

import (
	"context"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testAuditRecordSequence uint64

func newTestStore(t *testing.T, ctx context.Context) Store {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite fake: %v", err)
	}
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate sqlite fake: %v", err)
	}

	return NewStore(db)
}

func saveTenant(t *testing.T, ctx context.Context, store Store, id tenant.ID, name string) {
	t.Helper()

	tenantName, ok := tenant.NewName(name)
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: id, Name: tenantName}); err != nil {
		t.Fatalf("save tenant: %v", err)
	}
}

func saveTenantWithOutbox(t *testing.T, ctx context.Context, store Store, eventID string, id tenant.ID, name string) {
	t.Helper()

	tenantName, ok := tenant.NewName(name)
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	if err := store.SaveTenantAndEnqueueOwnerGrant(ctx, eventID, tenant.Tenant{
		ID:   id,
		Name: tenantName,
	}, identity.Principal{ID: identity.PrincipalID("user-one")}, auditRecord(t, eventID, id, "", audit.ActionTenantCreated)); err != nil {
		t.Fatalf("save tenant with outbox: %v", err)
	}
}

func saveInventory(t *testing.T, ctx context.Context, store Store, id string, tenantID tenant.ID, name string) {
	t.Helper()

	inventoryName, ok := inventory.NewName(name)
	if !ok {
		t.Fatalf("expected valid inventory name")
	}
	item := inventory.Inventory{
		ID:       inventory.InventoryID(id),
		TenantID: inventory.TenantID(tenantID.String()),
		Name:     inventoryName,
	}
	if err := store.SaveInventory(ctx, item); err != nil {
		t.Fatalf("save inventory: %v", err)
	}
}

func saveInventoryWithOutbox(t *testing.T, ctx context.Context, store Store, eventID string, id string, tenantID tenant.ID, name string) {
	t.Helper()

	inventoryName, ok := inventory.NewName(name)
	if !ok {
		t.Fatalf("expected valid inventory name")
	}
	item := inventory.Inventory{
		ID:       inventory.InventoryID(id),
		TenantID: inventory.TenantID(tenantID.String()),
		Name:     inventoryName,
	}
	if err := store.SaveInventoryAndEnqueueOwnerGrant(ctx, eventID, item, tenantID, identity.Principal{ID: identity.PrincipalID("user-one")}, auditRecord(t, eventID, tenantID, item.ID, audit.ActionInventoryCreated)); err != nil {
		t.Fatalf("save inventory with outbox: %v", err)
	}
}

func createAsset(t *testing.T, ctx context.Context, store Store, item asset.Asset) error {
	t.Helper()

	return store.CreateAsset(ctx, item, auditRecord(t, auditIDWithSuffix(item.ID.String(), "C"), tenant.ID(item.TenantID.String()), inventory.InventoryID(item.InventoryID.String()), audit.ActionAssetCreated), nil)
}

func updateAsset(t *testing.T, ctx context.Context, store Store, item asset.Asset) error {
	t.Helper()

	return store.UpdateAsset(ctx, item, []audit.Record{
		auditRecord(t, auditIDWithSuffix(item.ID.String(), "U"), tenant.ID(item.TenantID.String()), inventory.InventoryID(item.InventoryID.String()), audit.ActionAssetUpdated),
	}, nil)
}

func saveCustomFieldDefinition(t *testing.T, ctx context.Context, store Store, definition customfield.Definition) error {
	t.Helper()

	return store.SaveCustomFieldDefinition(ctx, definition, auditRecord(t, auditIDWithSuffix(definition.ID.String(), "D"), tenant.ID(definition.TenantID.String()), inventory.InventoryID(definition.InventoryID.String()), audit.ActionCustomFieldDefinitionCreated))
}

func saveCustomAssetType(t *testing.T, ctx context.Context, store Store, assetType customfield.AssetType) error {
	t.Helper()

	return store.SaveCustomAssetType(ctx, assetType, auditRecord(t, auditIDWithSuffix(assetType.ID.String(), "T"), tenant.ID(assetType.TenantID.String()), inventory.InventoryID(assetType.InventoryID.String()), audit.ActionCustomAssetTypeCreated))
}

func saveInventoryAccessGrantAndEnqueue(t *testing.T, ctx context.Context, store Store, eventID string, grant ports.InventoryAccessGrant) error {
	t.Helper()

	return store.SaveInventoryAccessGrantAndEnqueue(ctx, eventID, grant, auditRecord(t, eventID, grant.TenantID, grant.InventoryID, audit.ActionInventoryAccessGranted))
}

func auditIDWithSuffix(id string, suffix string) string {
	sequence := atomic.AddUint64(&testAuditRecordSequence, 1)
	return id + "-" + suffix + "-" + strconv.FormatUint(sequence, 10)
}

func auditRecordsIncludeAction(records []audit.Record, action audit.Action) bool {
	for _, record := range records {
		if record.Action == action {
			return true
		}
	}
	return false
}

func auditRecord(t *testing.T, id string, tenantID tenant.ID, inventoryID inventory.InventoryID, action audit.Action) audit.Record {
	t.Helper()
	return auditRecordAt(t, id, tenantID, inventoryID, action, time.Now())
}

func auditRecordAt(t *testing.T, id string, tenantID tenant.ID, inventoryID inventory.InventoryID, action audit.Action, occurredAt time.Time) audit.Record {
	t.Helper()

	record, ok := audit.NewRecord(
		audit.ID(id),
		audit.TenantID(tenantID.String()),
		audit.InventoryID(inventoryID.String()),
		audit.PrincipalID("user-one"),
		action,
		audit.SourceAPI,
		audit.TargetAsset,
		id+"-target",
		occurredAt,
		"",
		map[string]string{"note": "safe"},
	)
	if !ok {
		t.Fatalf("expected valid audit record")
	}
	return record
}

func customFieldDefinition(t *testing.T, id string, tenantID tenant.ID, inventoryID inventory.InventoryID, scope customfield.Scope, keyValue string, fieldType customfield.FieldType, rawOptions []string) customfield.Definition {
	t.Helper()

	definitionID, ok := customfield.NewID(id)
	if !ok {
		t.Fatalf("expected valid definition id")
	}
	key, ok := customfield.NewKey(keyValue)
	if !ok {
		t.Fatalf("expected valid custom field key")
	}
	displayName, ok := customfield.NewDisplayName("Field " + keyValue)
	if !ok {
		t.Fatalf("expected valid display name")
	}
	options := make([]customfield.Key, 0, len(rawOptions))
	for _, raw := range rawOptions {
		option, ok := customfield.NewKey(raw)
		if !ok {
			t.Fatalf("expected valid enum option")
		}
		options = append(options, option)
	}
	definition, ok := customfield.NewDefinition(
		definitionID,
		customfield.TenantID(tenantID.String()),
		customfield.InventoryID(inventoryID.String()),
		scope,
		key,
		displayName,
		fieldType,
		options,
		customfield.ApplicabilityAllAssets,
		nil,
	)
	if !ok {
		t.Fatalf("expected valid custom field definition")
	}
	return definition
}

func customAssetType(t *testing.T, id string, tenantID string, inventoryID string, scope customfield.Scope, keyValue string) customfield.AssetType {
	t.Helper()

	assetTypeID, ok := customfield.NewAssetTypeID(id)
	if !ok {
		t.Fatalf("expected valid custom asset type id")
	}
	key, ok := customfield.NewKey(keyValue)
	if !ok {
		t.Fatalf("expected valid custom asset type key")
	}
	displayName, ok := customfield.NewDisplayName("Type " + keyValue)
	if !ok {
		t.Fatalf("expected valid custom asset type display name")
	}
	description, ok := customfield.NewDescription("")
	if !ok {
		t.Fatalf("expected valid custom asset type description")
	}
	assetType, ok := customfield.NewAssetType(assetTypeID, customfield.TenantID(tenantID), customfield.InventoryID(inventoryID), scope, key, displayName, description)
	if !ok {
		t.Fatalf("expected valid custom asset type")
	}
	return assetType
}

func assetItem(id string, tenantID string, inventoryID string, kind asset.Kind, parentID string) asset.Asset {
	title, ok := asset.NewTitle("Asset " + id)
	if !ok {
		panic("invalid test asset title")
	}
	parent := asset.ID("")
	if parentID != "" {
		var parentOK bool
		parent, parentOK = asset.NewID(parentID)
		if !parentOK {
			panic("invalid parent id")
		}
	}
	return asset.Asset{
		ID:             asset.ID(id),
		TenantID:       asset.TenantID(tenantID),
		InventoryID:    asset.InventoryID(inventoryID),
		ParentAssetID:  parent,
		Kind:           kind,
		Title:          title,
		CustomFields:   asset.NewEmptyCustomFields(),
		LifecycleState: asset.LifecycleStateActive,
	}
}
