package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

func TestStateChangingOperationsWriteAuditHistory(t *testing.T) {
	assets := &fakeAssetRepository{}
	customFields := &fakeCustomFieldRepository{}
	inventories := &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}}
	outbox := &fakeOutbox{}
	application := New(Dependencies{
		Observer:                  &fakeObserver{},
		Authorizer:                &fakeAuthorizer{},
		Tenants:                   &fakeTenantRepository{exists: true},
		TenantUnitOfWork:          &fakeTenantRepository{exists: true},
		Inventories:               inventories,
		InventoryUnitOfWork:       inventories,
		InventoryAccess:           inventories,
		InventoryAccessUnitOfWork: inventories,
		CustomFields:              customFields,
		CustomFieldUnitOfWork:     customFields,
		Assets:                    assets,
		AssetUnitOfWork:           assets,
		Undoables:                 assets,
		Audit:                     &fakeAuditRepository{},
		Outbox:                    outbox,
		IDs: &fakeIDGenerator{ids: []string{
			"tenant-created", "audit-tenant", "tenant-owner-event", "claim-tenant",
			"inventory-created", "audit-inventory", "inventory-owner-event", "claim-inventory",
			"audit-share", "share-event", "claim-share",
			"definition-created", "audit-definition",
			"garage", "audit-location",
			"drill", "audit-asset",
			"audit-asset-updated", "audit-asset-moved",
		}},
	})

	if _, err := application.CreateTenant(context.Background(), CreateTenantInput{
		Principal: identity.Principal{ID: identity.PrincipalID("owner")},
		Name:      "Home",
	}); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	if _, err := application.CreateInventory(context.Background(), CreateInventoryInput{
		Principal: identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:  tenant.ID("tenant-one"),
		Name:      "Tools",
	}); err != nil {
		t.Fatalf("create inventory: %v", err)
	}
	if _, err := application.GrantInventoryAccess(context.Background(), GrantInventoryAccessInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		TargetUserID: "viewer",
		Relationship: "viewer",
	}); err != nil {
		t.Fatalf("grant inventory access: %v", err)
	}
	if _, err := application.CreateInventoryCustomFieldDefinition(context.Background(), CreateCustomFieldDefinitionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Key:         "serial",
		DisplayName: "Serial",
		Type:        "text",
	}); err != nil {
		t.Fatalf("create custom field definition: %v", err)
	}
	locationResult, err := application.CreateAssetWithOperation(context.Background(), CreateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Kind:        "location",
		Title:       "Garage",
	})
	if err != nil {
		t.Fatalf("create location asset: %v", err)
	}
	location := locationResult.Asset
	itemResult, err := application.CreateAssetWithOperation(context.Background(), CreateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Kind:        "item",
		Title:       "Drill",
	})
	if err != nil {
		t.Fatalf("create item asset: %v", err)
	}
	item := itemResult.Asset
	title := "Cordless Drill"
	if _, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     item.ID,
		Title:       &title,
		ParentAssetID: AssetParentUpdate{
			Present: true,
			Value:   location.ID.String(),
		},
	}); err != nil {
		t.Fatalf("update and move asset: %v", err)
	}

	auditRecords := append([]audit.Record{}, outbox.auditRecords...)
	auditRecords = append(auditRecords, inventories.auditRecords...)
	auditRecords = append(auditRecords, customFields.auditRecords...)
	auditRecords = append(auditRecords, assets.auditRecords...)
	collectedAudits := &fakeAuditRepository{items: auditRecords}
	for _, expected := range []audit.Action{
		audit.ActionTenantCreated,
		audit.ActionInventoryCreated,
		audit.ActionInventoryAccessGranted,
		audit.ActionCustomFieldDefinitionCreated,
		audit.ActionAssetCreated,
		audit.ActionAssetUpdated,
	} {
		if !collectedAudits.hasAction(expected) {
			t.Fatalf("expected audit action %s in %+v", expected, collectedAudits.items)
		}
	}
	moved, ok := collectedAudits.recordForAction(audit.ActionAssetUpdated)
	if !ok {
		t.Fatalf("expected coherent asset update audit record")
	}
	if moved.Source != audit.SourceAPI || moved.TargetType != audit.TargetAsset || moved.TargetID != item.ID.String() || moved.Metadata["new_parent"] != location.ID.String() {
		t.Fatalf("unexpected asset moved record: %+v", moved)
	}
}

func TestListAuditRecordsPaginatesAndEnforcesScope(t *testing.T) {
	audits := &fakeAuditRepository{items: []audit.Record{
		auditRecord("audit-tenant", "tenant-one", "", audit.ActionTenantCreated),
		auditRecord("audit-other-tenant", "tenant-two", "", audit.ActionTenantCreated),
		auditRecord("audit-one", "tenant-one", "inventory-one", audit.ActionAssetCreated),
		auditRecord("audit-two", "tenant-one", "inventory-one", audit.ActionAssetUpdated),
		auditRecord("audit-three", "tenant-one", "inventory-two", audit.ActionAssetCreated),
		auditRecord("audit-four", "tenant-two", "inventory-three", audit.ActionAssetCreated),
	}}
	application := New(Dependencies{
		Observer:         &fakeObserver{},
		Authorizer:       &fakeAuthorizer{},
		Tenants:          &fakeTenantRepository{exists: true},
		TenantUnitOfWork: &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
			inventoryItem("inventory-two", "tenant-one", "Supplies"),
		}},
		Audit:  audits,
		Outbox: &fakeOutbox{},
	})

	firstPage, err := application.ListInventoryAuditRecords(context.Background(), ListAuditRecordsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Limit:       1,
	})
	if err != nil {
		t.Fatalf("list first audit page: %v", err)
	}
	if len(firstPage.Items) != 1 || !firstPage.HasMore || firstPage.NextCursor == nil || firstPage.Items[0].ID != audit.ID("audit-one") {
		t.Fatalf("unexpected first audit page: %+v", firstPage)
	}
	secondPage, err := application.ListInventoryAuditRecords(context.Background(), ListAuditRecordsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Limit:       1,
		Cursor:      *firstPage.NextCursor,
	})
	if err != nil {
		t.Fatalf("list second audit page: %v", err)
	}
	if len(secondPage.Items) != 1 || secondPage.Items[0].ID != audit.ID("audit-two") {
		t.Fatalf("unexpected second audit page: %+v", secondPage)
	}

	tenantPage, err := application.ListTenantAuditRecords(context.Background(), ListAuditRecordsInput{
		Principal: identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:  tenant.ID("tenant-one"),
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("list tenant audit records: %v", err)
	}
	if len(tenantPage.Items) != 1 || tenantPage.Items[0].ID != audit.ID("audit-tenant") {
		t.Fatalf("expected tenant audit list to include only tenant-scoped records, got %+v", tenantPage.Items)
	}

	deniedInventory := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: ports.ErrForbidden,
		},
		Tenants:          &fakeTenantRepository{exists: true},
		TenantUnitOfWork: &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		}},
		Audit:  audits,
		Outbox: &fakeOutbox{},
	})
	_, err = deniedInventory.ListInventoryAuditRecords(context.Background(), ListAuditRecordsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("intruder")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
	})
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected forbidden inventory audit read, got %v", err)
	}

	deniedTenant := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkTenantErr: ports.ErrForbidden,
		},
		Tenants:          &fakeTenantRepository{exists: true},
		TenantUnitOfWork: &fakeTenantRepository{exists: true},
		Audit:            audits,
		Outbox:           &fakeOutbox{},
	})
	_, err = deniedTenant.ListTenantAuditRecords(context.Background(), ListAuditRecordsInput{
		Principal: identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:  tenant.ID("tenant-one"),
	})
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected forbidden tenant audit read, got %v", err)
	}
}

func TestListAuditRecordsFallsBackWhenPrincipalResolutionFails(t *testing.T) {
	audits := &fakeAuditRepository{items: []audit.Record{
		auditRecord("audit-one", "tenant-one", "inventory-one", audit.ActionAssetCreated),
	}}
	observer := &fakeObserver{}
	application := New(Dependencies{
		Observer:         observer,
		Authorizer:       &fakeAuthorizer{},
		Users:            &fakeUserRepository{err: errors.New("profile store unavailable")},
		Tenants:          &fakeTenantRepository{exists: true},
		TenantUnitOfWork: &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		}},
		Audit:  audits,
		Outbox: &fakeOutbox{},
	})

	page, err := application.ListInventoryAuditRecords(context.Background(), ListAuditRecordsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("list audit records: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].PrincipalID.String() != "owner" {
		t.Fatalf("expected audit row with raw principal fallback, got %+v", page.Items)
	}
	if len(page.ResolvedPrincipals) != 0 {
		t.Fatalf("expected unresolved principals when profile lookup fails, got %+v", page.ResolvedPrincipals)
	}
	if !observer.hasEvent(ports.EventAuditPrincipalResolutionFailed) {
		t.Fatalf("expected principal resolution failure event, got %+v", observer.events)
	}
}

func TestListAssetAuditHistoryRecordsConcreteAssetRead(t *testing.T) {
	audits := &fakeAuditRepository{items: []audit.Record{
		auditRecord("audit-one", "tenant-one", "inventory-one", audit.ActionAssetCreated),
		auditRecord("audit-two", "tenant-one", "inventory-one", audit.ActionAssetUpdated),
	}}
	audits.items[0].TargetID = "asset-one"
	audits.items[1].TargetID = "asset-one"
	application := New(Dependencies{
		Observer:         &fakeObserver{},
		Authorizer:       &fakeAuthorizer{},
		Tenants:          &fakeTenantRepository{exists: true},
		TenantUnitOfWork: &fakeTenantRepository{exists: true},
		Inventories: &fakeInventoryRepository{items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		}},
		Audit:  audits,
		Outbox: &fakeOutbox{},
	})

	page, err := application.ListAssetAuditHistory(context.Background(), ListAssetAuditHistoryInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     "asset-one",
		Limit:       1,
	})
	if err != nil {
		t.Fatalf("list asset audit history: %v", err)
	}
	if len(page.Items) != 1 || !page.HasMore {
		t.Fatalf("expected bounded asset audit page with more history, got %+v", page)
	}
	readRecord, ok := audits.recordForAction(audit.ActionAuditRecordListed)
	if !ok {
		t.Fatalf("expected asset audit read record, got %+v", audits.items)
	}
	if readRecord.TargetID != "asset-one" || readRecord.Metadata["target_id"] != "asset-one" || readRecord.Metadata["target_type"] != audit.TargetAsset.String() {
		t.Fatalf("expected read audit to name concrete asset, got %+v", readRecord)
	}
}

func fakeAuditResultsContain(items []audit.Record, id audit.ID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}
