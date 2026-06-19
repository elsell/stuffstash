package app

import (
	"context"
	"errors"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/gormstore"
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

func TestCreateTenantEnqueuesAndDrainsOwnerGrant(t *testing.T) {
	authorizer := &fakeAuthorizer{}
	outbox := &fakeOutbox{}
	application := New(Dependencies{
		Observer:    &fakeObserver{},
		Authorizer:  authorizer,
		Tenants:     &fakeTenantRepository{},
		Inventories: &fakeInventoryRepository{},
		Audit:       &fakeAuditRepository{},
		Outbox:      outbox,
		IDs:         &fakeIDGenerator{ids: []string{"tenant-one", "audit-one", "event-one"}},
	})

	item, err := application.CreateTenant(context.Background(), CreateTenantInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		Name:      "Home",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	if item.ID != tenant.ID("tenant-one") {
		t.Fatalf("expected tenant-one, got %q", item.ID)
	}
	if len(outbox.processed) != 1 || outbox.processed[0] != "event-one" {
		t.Fatalf("expected event-one processed, got %+v", outbox.processed)
	}
	if !slices.Contains(authorizer.tenantOwnerGrants, "user-one:tenant-one") {
		t.Fatalf("expected tenant owner grant, got %+v", authorizer.tenantOwnerGrants)
	}
}

func TestCreateInventoryEnqueuesAndDrainsOwnerGrant(t *testing.T) {
	authorizer := &fakeAuthorizer{}
	outbox := &fakeOutbox{}
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: authorizer,
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: &fakeInventoryRepository{},
		Audit:       &fakeAuditRepository{},
		Outbox:      outbox,
		IDs:         &fakeIDGenerator{ids: []string{"inventory-one", "audit-one", "event-one"}},
	})

	item, err := application.CreateInventory(context.Background(), CreateInventoryInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:  tenant.ID("tenant-one"),
		Name:      "Tools",
	})
	if err != nil {
		t.Fatalf("create inventory: %v", err)
	}
	if item.ID != inventory.InventoryID("inventory-one") {
		t.Fatalf("expected inventory-one, got %q", item.ID)
	}
	if len(outbox.processed) != 1 || outbox.processed[0] != "event-one" {
		t.Fatalf("expected event-one processed, got %+v", outbox.processed)
	}
	if !slices.Contains(authorizer.inventoryOwnerGrants, "user-one:tenant-one:inventory-one") {
		t.Fatalf("expected inventory owner grant, got %+v", authorizer.inventoryOwnerGrants)
	}
}

func TestCreateTenantKeepsOutboxEventPendingWhenGrantFails(t *testing.T) {
	expected := errors.New("spicedb unavailable")
	authorizer := &fakeAuthorizer{grantTenantOwnerErr: expected}
	observer := &fakeObserver{}
	outbox := &fakeOutbox{}
	application := New(Dependencies{
		Observer:    observer,
		Authorizer:  authorizer,
		Tenants:     &fakeTenantRepository{},
		Inventories: &fakeInventoryRepository{},
		Audit:       &fakeAuditRepository{},
		Outbox:      outbox,
		IDs:         &fakeIDGenerator{ids: []string{"tenant-one", "audit-one", "event-one"}},
	})

	_, err := application.CreateTenant(context.Background(), CreateTenantInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		Name:      "Home",
	})
	if err != nil {
		t.Fatalf("create tenant should persist outbox despite grant failure: %v", err)
	}
	if len(outbox.processed) != 0 {
		t.Fatalf("expected no processed events, got %+v", outbox.processed)
	}
	if len(outbox.failed) != 1 || outbox.failed[0] != "event-one" {
		t.Fatalf("expected event-one failure recorded, got %+v", outbox.failed)
	}
	if len(outbox.events) != 1 {
		t.Fatalf("expected event to remain pending, got %+v", outbox.events)
	}
	if !observer.hasEvent(ports.EventAuthorizationOutboxFailed) {
		t.Fatalf("expected outbox failure observability event, got %+v", observer.events)
	}
}

func TestDrainAuthorizationOutboxContinuesAfterFailedEvent(t *testing.T) {
	expected := errors.New("spicedb unavailable")
	authorizer := &fakeAuthorizer{grantTenantOwnerErr: expected}
	outbox := &fakeOutbox{
		events: []ports.AuthorizationOutboxEvent{
			{
				ID:          "tenant-event",
				Kind:        ports.AuthorizationOutboxGrantTenantOwner,
				PrincipalID: identity.PrincipalID("user-one"),
				TenantID:    tenant.ID("tenant-one"),
			},
			{
				ID:          "inventory-event",
				Kind:        ports.AuthorizationOutboxGrantInventoryOwner,
				PrincipalID: identity.PrincipalID("user-one"),
				TenantID:    tenant.ID("tenant-one"),
				InventoryID: inventory.InventoryID("inventory-one"),
			},
		},
	}
	application := New(Dependencies{
		Observer:    &fakeObserver{},
		Authorizer:  authorizer,
		Tenants:     &fakeTenantRepository{},
		Inventories: &fakeInventoryRepository{},
		Audit:       &fakeAuditRepository{},
		Outbox:      outbox,
		IDs:         &fakeIDGenerator{},
	})

	err := application.DrainAuthorizationOutbox(context.Background(), 10)
	if !errors.Is(err, expected) {
		t.Fatalf("expected drain error %v, got %v", expected, err)
	}
	if !slices.Contains(outbox.failed, "tenant-event") {
		t.Fatalf("expected tenant event failure, got %+v", outbox.failed)
	}
	if !slices.Contains(outbox.processed, "inventory-event") {
		t.Fatalf("expected inventory event to process after failed tenant event, got %+v", outbox.processed)
	}
}

func TestDrainAuthorizationOutboxDeadLettersUnrecoverableEventAndContinues(t *testing.T) {
	observer := &fakeObserver{}
	outbox := &fakeOutbox{
		events: []ports.AuthorizationOutboxEvent{
			{
				ID:          "bad-event",
				Kind:        ports.AuthorizationOutboxEventKind("unknown"),
				PrincipalID: identity.PrincipalID("user-one"),
				TenantID:    tenant.ID("tenant-one"),
			},
			{
				ID:          "inventory-event",
				Kind:        ports.AuthorizationOutboxGrantInventoryOwner,
				PrincipalID: identity.PrincipalID("user-one"),
				TenantID:    tenant.ID("tenant-one"),
				InventoryID: inventory.InventoryID("inventory-one"),
			},
		},
	}
	application := New(Dependencies{
		Observer:    observer,
		Authorizer:  &fakeAuthorizer{},
		Tenants:     &fakeTenantRepository{},
		Inventories: &fakeInventoryRepository{},
		Audit:       &fakeAuditRepository{},
		Outbox:      outbox,
		IDs:         &fakeIDGenerator{},
	})

	if err := application.DrainAuthorizationOutbox(context.Background(), 10); err != nil {
		t.Fatalf("dead-lettering unrecoverable events should not fail the batch: %v", err)
	}
	if !slices.Contains(outbox.deadLettered, "bad-event") {
		t.Fatalf("expected bad event to be dead-lettered, got %+v", outbox.deadLettered)
	}
	if !slices.Contains(outbox.processed, "inventory-event") {
		t.Fatalf("expected inventory event to process after dead-letter, got %+v", outbox.processed)
	}
	if !observer.hasEvent(ports.EventAuthorizationOutboxDeadLettered) {
		t.Fatalf("expected outbox dead-letter observability event, got %+v", observer.events)
	}

	events, err := outbox.ClaimPendingAuthorizationOutboxEvents(context.Background(), "claim-two", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected dead-lettered event to stay out of pending claims, got %+v", events)
	}
}

func TestDrainAuthorizationOutboxDeadLettersDurableInvalidEvent(t *testing.T) {
	ctx := context.Background()
	store := newAppTestGORMStore(t, ctx)
	tenantName, ok := tenant.NewName("Home")
	if !ok {
		t.Fatalf("expected valid tenant name")
	}
	if err := store.SaveTenantAndEnqueueOwnerGrant(ctx, "event-one", tenant.Tenant{
		ID:   tenant.ID("tenant-one"),
		Name: tenantName,
	}, identity.Principal{}, auditRecord("audit-one", "tenant-one", "", audit.ActionTenantCreated)); err != nil {
		t.Fatalf("save tenant and enqueue invalid owner grant: %v", err)
	}

	observer := &fakeObserver{}
	authorizer := &fakeAuthorizer{}
	application := New(Dependencies{
		Observer:    observer,
		Authorizer:  authorizer,
		Tenants:     store,
		Inventories: store,
		Audit:       store,
		Outbox:      store,
		IDs:         &fakeIDGenerator{ids: []string{"claim-one"}},
	})

	if err := application.DrainAuthorizationOutbox(ctx, 10); err != nil {
		t.Fatalf("dead-lettering durable invalid events should not fail the batch: %v", err)
	}
	if len(authorizer.tenantOwnerGrants) != 0 {
		t.Fatalf("expected invalid event not to reach authorizer, got %+v", authorizer.tenantOwnerGrants)
	}
	if !observer.hasEvent(ports.EventAuthorizationOutboxDeadLettered) {
		t.Fatalf("expected outbox dead-letter observability event, got %+v", observer.events)
	}
	deadLetterEvent, ok := observer.eventNamed(ports.EventAuthorizationOutboxDeadLettered)
	if !ok || !strings.Contains(deadLetterEvent.Fields["reason"], "principal id") {
		t.Fatalf("expected actionable dead-letter reason, got %+v", deadLetterEvent)
	}

	events, err := store.ClaimPendingAuthorizationOutboxEvents(ctx, "claim-two", 10, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("claim outbox events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected durable dead-lettered event to stay out of pending claims, got %+v", events)
	}
}

func TestListInventoriesReturnsAuthorizationBackendFailures(t *testing.T) {
	expected := errors.New("authorization backend unavailable")
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: expected,
		},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: &fakeInventoryRepository{
			items: []inventory.Inventory{
				inventoryItem("inventory-one", "tenant-one", "Tools"),
			},
		},
		Audit:  &fakeAuditRepository{},
		Outbox: &fakeOutbox{},
	})

	_, err := application.ListInventories(context.Background(), ListInventoriesInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:  tenant.ID("tenant-one"),
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected backend error, got %v", err)
	}
}

func TestListInventoriesSkipsForbiddenInventories(t *testing.T) {
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: ports.ErrForbidden,
		},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: &fakeInventoryRepository{
			items: []inventory.Inventory{
				inventoryItem("inventory-one", "tenant-one", "Tools"),
			},
		},
		Audit:  &fakeAuditRepository{},
		Outbox: &fakeOutbox{},
	})

	result, err := application.ListInventories(context.Background(), ListInventoriesInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:  tenant.ID("tenant-one"),
	})
	if err != nil {
		t.Fatalf("list inventories: %v", err)
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected forbidden inventory to be hidden, got %+v", result.Items)
	}
}

func TestListInventoriesPaginatesAfterAuthorizationFiltering(t *testing.T) {
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Visible One"),
			inventoryItem("inventory-two", "tenant-one", "Hidden"),
			inventoryItem("inventory-three", "tenant-one", "Visible Two"),
		},
	}
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: &selectiveInventoryAuthorizer{forbidden: map[inventory.InventoryID]struct{}{"inventory-two": {}}},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories:      repository,
		Audit:            &fakeAuditRepository{},
		Outbox:           &fakeOutbox{},
		DefaultPageLimit: 1,
		MaxPageLimit:     1,
	})

	firstPage, err := application.ListInventories(context.Background(), ListInventoriesInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:  tenant.ID("tenant-one"),
	})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(firstPage.Items) != 1 || firstPage.Items[0].ID != inventory.InventoryID("inventory-one") || !firstPage.HasMore || firstPage.NextCursor == nil {
		t.Fatalf("expected first visible inventory page, got %+v", firstPage)
	}
	if repository.calls != 1 {
		t.Fatalf("expected one bounded repository scan, got %d", repository.calls)
	}
	if len(repository.limits) != 1 || repository.limits[0] != 3 {
		t.Fatalf("expected bounded scan limit 3, got %+v", repository.limits)
	}

	secondPage, err := application.ListInventories(context.Background(), ListInventoriesInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:  tenant.ID("tenant-one"),
		Cursor:    *firstPage.NextCursor,
	})
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(secondPage.Items) != 1 || secondPage.Items[0].ID != inventory.InventoryID("inventory-three") || secondPage.HasMore {
		t.Fatalf("expected second visible inventory page, got %+v", secondPage)
	}
	if repository.calls != 2 {
		t.Fatalf("expected second request to make one more bounded scan, got %d", repository.calls)
	}
	if len(repository.limits) != 2 || repository.limits[1] != 3 {
		t.Fatalf("expected second bounded scan limit 3, got %+v", repository.limits)
	}
}

func TestListInventoriesReturnsEmptyBoundedPageWhenScanWindowIsHidden(t *testing.T) {
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("a-hidden-one", "tenant-one", "Hidden One"),
			inventoryItem("b-hidden-two", "tenant-one", "Hidden Two"),
			inventoryItem("c-hidden-three", "tenant-one", "Hidden Three"),
			inventoryItem("d-visible", "tenant-one", "Visible"),
		},
	}
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &selectiveInventoryAuthorizer{forbidden: map[inventory.InventoryID]struct{}{
			"a-hidden-one":   {},
			"b-hidden-two":   {},
			"c-hidden-three": {},
		}},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories:      repository,
		Audit:            &fakeAuditRepository{},
		Outbox:           &fakeOutbox{},
		DefaultPageLimit: 1,
		MaxPageLimit:     1,
	})

	firstPage, err := application.ListInventories(context.Background(), ListInventoriesInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:  tenant.ID("tenant-one"),
	})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(firstPage.Items) != 0 || !firstPage.HasMore || firstPage.NextCursor == nil {
		t.Fatalf("expected empty bounded page with continuation cursor, got %+v", firstPage)
	}
	if repository.calls != 1 || len(repository.limits) != 1 || repository.limits[0] != 3 {
		t.Fatalf("expected one bounded scan of 3 raw inventories, calls=%d limits=%+v", repository.calls, repository.limits)
	}

	secondPage, err := application.ListInventories(context.Background(), ListInventoriesInput{
		Principal: identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:  tenant.ID("tenant-one"),
		Cursor:    *firstPage.NextCursor,
	})
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(secondPage.Items) != 1 || secondPage.Items[0].ID != inventory.InventoryID("d-visible") || secondPage.HasMore {
		t.Fatalf("expected visible inventory after hidden scan window, got %+v", secondPage)
	}
}

func TestCreateAndListAssets(t *testing.T) {
	assets := &fakeAssetRepository{}
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: &fakeAuthorizer{},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: &fakeInventoryRepository{
			items: []inventory.Inventory{
				inventoryItem("inventory-one", "tenant-one", "Tools"),
			},
		},
		Assets:           assets,
		Audit:            &fakeAuditRepository{},
		Outbox:           &fakeOutbox{},
		IDs:              &fakeIDGenerator{ids: []string{"asset-one", "asset-two"}},
		DefaultPageLimit: 1,
		MaxPageLimit:     2,
	})

	location, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Kind:        "location",
		Title:       "Garage",
	})
	if err != nil {
		t.Fatalf("create location asset: %v", err)
	}
	if location.Kind != asset.KindLocation || location.LifecycleState != asset.LifecycleStateActive {
		t.Fatalf("unexpected location asset: %+v", location)
	}

	item, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:     identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:      tenant.ID("tenant-one"),
		InventoryID:   inventory.InventoryID("inventory-one"),
		Kind:          "item",
		Title:         "Drill",
		Description:   "Cordless",
		ParentAssetID: location.ID.String(),
	})
	if err != nil {
		t.Fatalf("create item asset: %v", err)
	}
	if item.ParentAssetID != location.ID {
		t.Fatalf("expected parent %q, got %q", location.ID, item.ParentAssetID)
	}

	result, err := application.ListAssets(context.Background(), ListAssetsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
	})
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(result.Items) != 1 || !result.HasMore || result.NextCursor == nil || result.Limit != 1 {
		t.Fatalf("expected paginated first page, got %+v", result)
	}

	nextPage, err := application.ListAssets(context.Background(), ListAssetsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Cursor:      *result.NextCursor,
	})
	if err != nil {
		t.Fatalf("list next assets page: %v", err)
	}
	if len(nextPage.Items) != 1 || nextPage.HasMore || nextPage.Items[0].ID != item.ID {
		t.Fatalf("expected second page with item, got %+v", nextPage)
	}
}

func TestCreateAssetRejectsItemParentAndCustomFields(t *testing.T) {
	itemParent := assetItem("asset-parent", "tenant-one", "inventory-one", asset.KindItem, "")
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: &fakeAuthorizer{},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: &fakeInventoryRepository{
			items: []inventory.Inventory{
				inventoryItem("inventory-one", "tenant-one", "Tools"),
			},
		},
		Assets: &fakeAssetRepository{
			items: map[asset.ID]asset.Asset{itemParent.ID: itemParent},
		},
		Audit:  &fakeAuditRepository{},
		Outbox: &fakeOutbox{},
		IDs:    &fakeIDGenerator{ids: []string{"asset-one"}},
	})

	_, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:     identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:      tenant.ID("tenant-one"),
		InventoryID:   inventory.InventoryID("inventory-one"),
		Kind:          "item",
		Title:         "Bit set",
		ParentAssetID: itemParent.ID.String(),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input for item parent, got %v", err)
	}

	_, err = application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("user-one")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		Kind:         "item",
		Title:        "Bit set",
		CustomFields: map[string]any{"serial": "abc"},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input for custom fields, got %v", err)
	}
}

func TestGrantInventoryAccessRequiresShareAndRejectsInvalidGrants(t *testing.T) {
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		},
	}
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: ports.ErrForbidden,
		},
		Tenants: &fakeTenantRepository{
			exists: true,
		},
		Inventories: repository,
		Audit:       &fakeAuditRepository{},
		Outbox:      &fakeOutbox{},
		IDs:         &fakeIDGenerator{ids: []string{"event-one"}},
	})

	_, err := application.GrantInventoryAccess(context.Background(), GrantInventoryAccessInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		TargetUserID: "viewer",
		Relationship: "viewer",
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized without share permission, got %v", err)
	}
	if len(repository.accessGrants) != 0 {
		t.Fatalf("expected no durable grant without share permission, got %+v", repository.accessGrants)
	}

	allowed := New(Dependencies{
		Observer:    &fakeObserver{},
		Authorizer:  &fakeAuthorizer{},
		Tenants:     &fakeTenantRepository{exists: true},
		Inventories: repository,
		Audit:       &fakeAuditRepository{},
		Outbox:      &fakeOutbox{},
		IDs:         &fakeIDGenerator{ids: []string{"event-two"}},
	})
	for _, item := range []struct {
		name          string
		targetUserID  string
		relationship  string
		expectedError error
	}{
		{name: "self grant", targetUserID: "owner", relationship: "viewer", expectedError: ErrInvalidInput},
		{name: "bad principal", targetUserID: "user/one", relationship: "viewer", expectedError: ErrInvalidInput},
		{name: "bad relationship", targetUserID: "viewer", relationship: "owner", expectedError: ErrInvalidInput},
	} {
		t.Run(item.name, func(t *testing.T) {
			_, err := allowed.GrantInventoryAccess(context.Background(), GrantInventoryAccessInput{
				Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
				TenantID:     tenant.ID("tenant-one"),
				InventoryID:  inventory.InventoryID("inventory-one"),
				TargetUserID: item.targetUserID,
				Relationship: item.relationship,
			})
			if !errors.Is(err, item.expectedError) {
				t.Fatalf("expected %v, got %v", item.expectedError, err)
			}
		})
	}
}

func TestGrantAndListInventoryAccessGrants(t *testing.T) {
	observer := &fakeObserver{}
	repository := &fakeInventoryRepository{
		items: []inventory.Inventory{
			inventoryItem("inventory-one", "tenant-one", "Tools"),
		},
	}
	application := New(Dependencies{
		Observer:     observer,
		Authorizer:   &fakeAuthorizer{},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  repository,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{ids: []string{"event-one", "event-two"}},
		MaxPageLimit: 1,
	})

	_, err := application.GrantInventoryAccess(context.Background(), GrantInventoryAccessInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		TargetUserID: "viewer",
		Relationship: "viewer",
	})
	if err != nil {
		t.Fatalf("grant viewer: %v", err)
	}
	_, err = application.GrantInventoryAccess(context.Background(), GrantInventoryAccessInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		TargetUserID: "editor",
		Relationship: "editor",
	})
	if err != nil {
		t.Fatalf("grant editor: %v", err)
	}

	firstPage, err := application.ListInventoryAccessGrants(context.Background(), ListInventoryAccessGrantsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Limit:       1,
	})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(firstPage.Items) != 1 || firstPage.Items[0].PrincipalID != identity.PrincipalID("editor") || !firstPage.HasMore || firstPage.NextCursor == nil {
		t.Fatalf("expected first grant page with editor, got %+v", firstPage)
	}

	secondPage, err := application.ListInventoryAccessGrants(context.Background(), ListInventoryAccessGrantsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Limit:       1,
		Cursor:      *firstPage.NextCursor,
	})
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(secondPage.Items) != 1 || secondPage.Items[0].PrincipalID != identity.PrincipalID("viewer") || secondPage.HasMore {
		t.Fatalf("expected second grant page with viewer, got %+v", secondPage)
	}
	if !observer.hasEvent(ports.EventInventoryAccessGranted) || !observer.hasEvent(ports.EventInventoryAccessListed) {
		t.Fatalf("expected grant/list observability events, got %+v", observer.events)
	}
}

func TestCreateAndListCustomFieldDefinitions(t *testing.T) {
	observer := &fakeObserver{}
	customFields := &fakeCustomFieldRepository{}
	application := New(Dependencies{
		Observer:     observer,
		Authorizer:   &fakeAuthorizer{},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: customFields,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{ids: []string{"tenant-definition", "inventory-definition"}},
		MaxPageLimit: 1,
	})

	tenantDefinition, err := application.CreateTenantCustomFieldDefinition(context.Background(), CreateCustomFieldDefinitionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		Key:         "serial",
		DisplayName: "Serial",
		Type:        "text",
	})
	if err != nil {
		t.Fatalf("create tenant definition: %v", err)
	}
	if tenantDefinition.Scope != customfield.ScopeTenant || tenantDefinition.Key != customfield.Key("serial") {
		t.Fatalf("unexpected tenant definition: %+v", tenantDefinition)
	}

	inventoryDefinition, err := application.CreateInventoryCustomFieldDefinition(context.Background(), CreateCustomFieldDefinitionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Key:         "condition",
		DisplayName: "Condition",
		Type:        "enum",
		EnumOptions: []string{"new", "used"},
	})
	if err != nil {
		t.Fatalf("create inventory definition: %v", err)
	}
	if inventoryDefinition.Scope != customfield.ScopeInventory || len(inventoryDefinition.EnumOptions) != 2 {
		t.Fatalf("unexpected inventory definition: %+v", inventoryDefinition)
	}

	firstPage, err := application.ListInventoryCustomFieldDefinitions(context.Background(), ListCustomFieldDefinitionsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Limit:       1,
	})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(firstPage.Items) != 1 || firstPage.Items[0].ID != tenantDefinition.ID || !firstPage.HasMore || firstPage.NextCursor == nil {
		t.Fatalf("expected first page with inherited tenant definition, got %+v", firstPage)
	}

	secondPage, err := application.ListInventoryCustomFieldDefinitions(context.Background(), ListCustomFieldDefinitionsInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Limit:       1,
		Cursor:      *firstPage.NextCursor,
	})
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(secondPage.Items) != 1 || secondPage.Items[0].ID != inventoryDefinition.ID || secondPage.HasMore {
		t.Fatalf("expected second page with inventory definition, got %+v", secondPage)
	}
	if !observer.hasEvent(ports.EventCustomFieldDefinitionCreated) || !observer.hasEvent(ports.EventCustomFieldDefinitionsListed) {
		t.Fatalf("expected custom field observability events, got %+v", observer.events)
	}
}

func TestCustomFieldDefinitionsRejectUnauthorizedAndDuplicateKeys(t *testing.T) {
	customFields := &fakeCustomFieldRepository{}
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkTenantErr: ports.ErrForbidden,
		},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: customFields,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{ids: []string{"definition-one"}},
	})

	_, err := application.CreateTenantCustomFieldDefinition(context.Background(), CreateCustomFieldDefinitionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:    tenant.ID("tenant-one"),
		Key:         "serial",
		DisplayName: "Serial",
		Type:        "text",
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized tenant definition create, got %v", err)
	}

	allowed := New(Dependencies{
		Observer:     &fakeObserver{},
		Authorizer:   &fakeAuthorizer{},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: customFields,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{ids: []string{"definition-two", "definition-three"}},
	})
	_, err = allowed.CreateTenantCustomFieldDefinition(context.Background(), CreateCustomFieldDefinitionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		Key:         "serial",
		DisplayName: "Serial",
		Type:        "text",
	})
	if err != nil {
		t.Fatalf("create first definition: %v", err)
	}
	_, err = allowed.CreateInventoryCustomFieldDefinition(context.Background(), CreateCustomFieldDefinitionInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Key:         "serial",
		DisplayName: "Serial",
		Type:        "text",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected duplicate effective key rejection, got %v", err)
	}
}

func TestCreateAssetValidatesCustomFieldsAgainstDefinitions(t *testing.T) {
	customFields := &fakeCustomFieldRepository{}
	serialDefinition := customFieldDefinition("serial-definition", "tenant-one", "", customfield.ScopeTenant, "serial", customfield.FieldTypeText, nil)
	conditionDefinition := customFieldDefinition("condition-definition", "tenant-one", "inventory-one", customfield.ScopeInventory, "condition", customfield.FieldTypeEnum, []string{"new", "used"})
	if err := customFields.SaveCustomFieldDefinition(context.Background(), serialDefinition, auditRecord("audit-serial", "tenant-one", "", audit.ActionCustomFieldDefinitionCreated)); err != nil {
		t.Fatalf("save serial definition: %v", err)
	}
	if err := customFields.SaveCustomFieldDefinition(context.Background(), conditionDefinition, auditRecord("audit-condition", "tenant-one", "inventory-one", audit.ActionCustomFieldDefinitionCreated)); err != nil {
		t.Fatalf("save condition definition: %v", err)
	}
	assets := &fakeAssetRepository{}
	application := New(Dependencies{
		Observer:     &fakeObserver{},
		Authorizer:   &fakeAuthorizer{},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: customFields,
		Assets:       assets,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{ids: []string{"asset-one"}},
	})

	item, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Kind:        "item",
		Title:       "Drill",
		CustomFields: map[string]any{
			"serial":    "abc",
			"condition": "used",
		},
	})
	if err != nil {
		t.Fatalf("create asset with custom fields: %v", err)
	}
	if item.CustomFields.Values()["serial"] != "abc" || item.CustomFields.Values()["condition"] != "used" {
		t.Fatalf("expected custom fields to be saved, got %+v", item.CustomFields.Values())
	}

	_, err = application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Kind:        "item",
		Title:       "Bad Drill",
		CustomFields: map[string]any{
			"condition": "broken",
		},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid enum value rejection, got %v", err)
	}
}

func TestUpdateAssetMovesAndValidatesCustomFields(t *testing.T) {
	customFields := &fakeCustomFieldRepository{}
	serialDefinition := customFieldDefinition("serial-definition", "tenant-one", "", customfield.ScopeTenant, "serial", customfield.FieldTypeText, nil)
	if err := customFields.SaveCustomFieldDefinition(context.Background(), serialDefinition, auditRecord("audit-serial", "tenant-one", "", audit.ActionCustomFieldDefinitionCreated)); err != nil {
		t.Fatalf("save serial definition: %v", err)
	}
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		asset.ID("garage"): assetItem("garage", "tenant-one", "inventory-one", asset.KindLocation, ""),
		asset.ID("shelf"):  assetItem("shelf", "tenant-one", "inventory-one", asset.KindLocation, "garage"),
		asset.ID("drill"):  assetItem("drill", "tenant-one", "inventory-one", asset.KindItem, "garage"),
	}}
	observer := &fakeObserver{}
	application := New(Dependencies{
		Observer:     observer,
		Authorizer:   &fakeAuthorizer{},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: customFields,
		Assets:       assets,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{},
	})

	title := "Cordless Drill"
	description := "Blue case"
	updated, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("drill"),
		Title:       &title,
		Description: &description,
		ParentAssetID: AssetParentUpdate{
			Present: true,
			Value:   "shelf",
		},
		CustomFields: map[string]any{"serial": "abc"},
	})
	if err != nil {
		t.Fatalf("update asset: %v", err)
	}
	if updated.Title.String() != title || updated.Description.String() != description || updated.ParentAssetID != asset.ID("shelf") {
		t.Fatalf("unexpected updated asset: %+v", updated)
	}
	if updated.CustomFields.Values()["serial"] != "abc" {
		t.Fatalf("expected updated custom fields, got %+v", updated.CustomFields.Values())
	}
	if assets.items[asset.ID("drill")].ParentAssetID != asset.ID("shelf") {
		t.Fatalf("expected persisted parent shelf, got %+v", assets.items[asset.ID("drill")])
	}
	if !observer.hasEvent(ports.EventAssetUpdated) {
		t.Fatalf("expected asset updated observability event, got %+v", observer.events)
	}

	_, err = application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:    identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:     tenant.ID("tenant-one"),
		InventoryID:  inventory.InventoryID("inventory-one"),
		AssetID:      asset.ID("drill"),
		CustomFields: map[string]any{"serial": 42},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid custom field update rejection, got %v", err)
	}
}

func TestUpdateAssetRejectsInvalidMovement(t *testing.T) {
	assets := &fakeAssetRepository{items: map[asset.ID]asset.Asset{
		asset.ID("garage"):   assetItem("garage", "tenant-one", "inventory-one", asset.KindLocation, ""),
		asset.ID("shelf"):    assetItem("shelf", "tenant-one", "inventory-one", asset.KindLocation, "garage"),
		asset.ID("box"):      assetItem("box", "tenant-one", "inventory-one", asset.KindContainer, "shelf"),
		asset.ID("wrench"):   assetItem("wrench", "tenant-one", "inventory-one", asset.KindItem, "box"),
		asset.ID("supplies"): assetItem("supplies", "tenant-one", "inventory-one", asset.KindItem, ""),
	}}
	application := New(Dependencies{
		Observer:     &fakeObserver{},
		Authorizer:   &fakeAuthorizer{},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: &fakeCustomFieldRepository{},
		Assets:       assets,
		Audit:        &fakeAuditRepository{},
		Outbox:       &fakeOutbox{},
		IDs:          &fakeIDGenerator{},
	})

	for _, item := range []struct {
		name    string
		assetID asset.ID
		parent  string
	}{
		{name: "self parent", assetID: asset.ID("box"), parent: "box"},
		{name: "cycle through descendant", assetID: asset.ID("garage"), parent: "box"},
		{name: "item parent", assetID: asset.ID("wrench"), parent: "supplies"},
	} {
		t.Run(item.name, func(t *testing.T) {
			_, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
				Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
				TenantID:    tenant.ID("tenant-one"),
				InventoryID: inventory.InventoryID("inventory-one"),
				AssetID:     item.assetID,
				ParentAssetID: AssetParentUpdate{
					Present: true,
					Value:   item.parent,
				},
			})
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected invalid movement rejection, got %v", err)
			}
		})
	}

	updated, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("box"),
		ParentAssetID: AssetParentUpdate{
			Present: true,
			Null:    true,
		},
	})
	if err != nil {
		t.Fatalf("move container to root: %v", err)
	}
	if updated.ParentAssetID.String() != "" || assets.items[asset.ID("wrench")].ParentAssetID != asset.ID("box") {
		t.Fatalf("expected box moved to root with child preserved, box=%+v wrench=%+v", updated, assets.items[asset.ID("wrench")])
	}

	_, err = application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("editor")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("box"),
		ParentAssetID: AssetParentUpdate{
			Present: true,
			Value:   " ",
		},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected blank parent rejection, got %v", err)
	}
}

func TestUpdateAssetRequiresEditPermission(t *testing.T) {
	application := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: ports.ErrForbidden,
		},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}},
		CustomFields: &fakeCustomFieldRepository{},
		Assets: &fakeAssetRepository{items: map[asset.ID]asset.Asset{
			asset.ID("drill"): assetItem("drill", "tenant-one", "inventory-one", asset.KindItem, ""),
		}},
		Audit:  &fakeAuditRepository{},
		Outbox: &fakeOutbox{},
		IDs:    &fakeIDGenerator{},
	})

	title := "Cordless Drill"
	_, err := application.UpdateAsset(context.Background(), UpdateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		AssetID:     asset.ID("drill"),
		Title:       &title,
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized update, got %v", err)
	}
}

func TestStateChangingOperationsWriteAuditHistory(t *testing.T) {
	assets := &fakeAssetRepository{}
	customFields := &fakeCustomFieldRepository{}
	inventories := &fakeInventoryRepository{items: []inventory.Inventory{inventoryItem("inventory-one", "tenant-one", "Tools")}}
	outbox := &fakeOutbox{}
	application := New(Dependencies{
		Observer:     &fakeObserver{},
		Authorizer:   &fakeAuthorizer{},
		Tenants:      &fakeTenantRepository{exists: true},
		Inventories:  inventories,
		CustomFields: customFields,
		Assets:       assets,
		Audit:        &fakeAuditRepository{},
		Outbox:       outbox,
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
	location, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Kind:        "location",
		Title:       "Garage",
	})
	if err != nil {
		t.Fatalf("create location asset: %v", err)
	}
	item, err := application.CreateAsset(context.Background(), CreateAssetInput{
		Principal:   identity.Principal{ID: identity.PrincipalID("owner")},
		TenantID:    tenant.ID("tenant-one"),
		InventoryID: inventory.InventoryID("inventory-one"),
		Kind:        "item",
		Title:       "Drill",
	})
	if err != nil {
		t.Fatalf("create item asset: %v", err)
	}
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
		audit.ActionAssetMoved,
	} {
		if !collectedAudits.hasAction(expected) {
			t.Fatalf("expected audit action %s in %+v", expected, collectedAudits.items)
		}
	}
	moved, ok := collectedAudits.recordForAction(audit.ActionAssetMoved)
	if !ok {
		t.Fatalf("expected asset moved audit record")
	}
	if moved.Source != audit.SourceAPI || moved.TargetType != audit.TargetAsset || moved.TargetID != item.ID.String() || moved.Metadata["new_parent"] != location.ID.String() {
		t.Fatalf("unexpected asset moved record: %+v", moved)
	}
}

func TestListAuditRecordsPaginatesAndEnforcesScope(t *testing.T) {
	audits := &fakeAuditRepository{items: []audit.Record{
		auditRecord("audit-one", "tenant-one", "inventory-one", audit.ActionAssetCreated),
		auditRecord("audit-two", "tenant-one", "inventory-one", audit.ActionAssetUpdated),
		auditRecord("audit-three", "tenant-one", "inventory-two", audit.ActionAssetCreated),
		auditRecord("audit-four", "tenant-two", "inventory-three", audit.ActionAssetCreated),
	}}
	application := New(Dependencies{
		Observer:   &fakeObserver{},
		Authorizer: &fakeAuthorizer{},
		Tenants:    &fakeTenantRepository{exists: true},
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
	if len(secondPage.Items) != 1 || secondPage.HasMore || secondPage.Items[0].ID != audit.ID("audit-two") {
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
	if len(tenantPage.Items) != 3 {
		t.Fatalf("expected tenant audit list to exclude other tenant, got %+v", tenantPage.Items)
	}

	deniedInventory := New(Dependencies{
		Observer: &fakeObserver{},
		Authorizer: &fakeAuthorizer{
			checkInventoryErr: ports.ErrForbidden,
		},
		Tenants: &fakeTenantRepository{exists: true},
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
		Tenants: &fakeTenantRepository{exists: true},
		Audit:   audits,
		Outbox:  &fakeOutbox{},
	})
	_, err = deniedTenant.ListTenantAuditRecords(context.Background(), ListAuditRecordsInput{
		Principal: identity.Principal{ID: identity.PrincipalID("viewer")},
		TenantID:  tenant.ID("tenant-one"),
	})
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected forbidden tenant audit read, got %v", err)
	}
}

type fakeAuthorizer struct {
	checkInventoryErr     error
	checkTenantErr        error
	grantTenantOwnerErr   error
	tenantOwnerGrants     []string
	inventoryOwnerGrants  []string
	inventoryViewerGrants []string
	inventoryEditorGrants []string
}

func (f *fakeAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	if f.checkTenantErr != nil {
		return f.checkTenantErr
	}
	return nil
}

func (f *fakeAuthorizer) CheckInventory(context.Context, identity.Principal, ports.InventoryPermission, inventory.InventoryID) error {
	return f.checkInventoryErr
}

func (f *fakeAuthorizer) GrantTenantOwner(_ context.Context, principal identity.Principal, tenantID tenant.ID) error {
	if f.grantTenantOwnerErr != nil {
		return f.grantTenantOwnerErr
	}
	f.tenantOwnerGrants = append(f.tenantOwnerGrants, principal.ID.String()+":"+tenantID.String())
	return nil
}

func (f *fakeAuthorizer) GrantInventoryOwner(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryOwnerGrants = append(f.inventoryOwnerGrants, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

func (f *fakeAuthorizer) GrantInventoryViewer(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryViewerGrants = append(f.inventoryViewerGrants, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

func (f *fakeAuthorizer) GrantInventoryEditor(_ context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	f.inventoryEditorGrants = append(f.inventoryEditorGrants, principal.ID.String()+":"+tenantID.String()+":"+inventoryID.String())
	return nil
}

type fakeTenantRepository struct {
	exists bool
}

func (f *fakeTenantRepository) SaveTenant(context.Context, tenant.Tenant) error {
	return nil
}

func (f *fakeTenantRepository) TenantExists(context.Context, tenant.ID) (bool, error) {
	return f.exists, nil
}

type fakeInventoryRepository struct {
	items        []inventory.Inventory
	accessGrants []ports.InventoryAccessGrant
	auditRecords []audit.Record
	calls        int
	limits       []int
}

func (f *fakeInventoryRepository) InventoryByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) (inventory.Inventory, bool, error) {
	for _, item := range f.items {
		if item.ID == inventoryID && item.TenantID == inventory.TenantID(tenantID.String()) {
			return item, true, nil
		}
	}
	return inventory.Inventory{}, false, nil
}

type fakeAssetRepository struct {
	items        map[asset.ID]asset.Asset
	auditRecords []audit.Record
}

type fakeCustomFieldRepository struct {
	items        []customfield.Definition
	auditRecords []audit.Record
}

type fakeOutbox struct {
	events       []ports.AuthorizationOutboxEvent
	auditRecords []audit.Record
	processed    []string
	failed       []string
	deadLettered []string
}

type fakeAuditRepository struct {
	items []audit.Record
}

func (f *fakeAuditRepository) SaveAuditRecord(_ context.Context, record audit.Record) error {
	f.items = append(f.items, record)
	return nil
}

func (f *fakeAuditRepository) hasAction(action audit.Action) bool {
	_, ok := f.recordForAction(action)
	return ok
}

func (f *fakeAuditRepository) recordForAction(action audit.Action) (audit.Record, bool) {
	for _, record := range f.items {
		if record.Action == action {
			return record, true
		}
	}
	return audit.Record{}, false
}

func (f *fakeAuditRepository) ListTenantAuditRecords(_ context.Context, tenantID tenant.ID, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	items := []audit.Record{}
	for _, record := range f.items {
		if record.TenantID.String() == tenantID.String() && fakeAuditRecordAfter(record, page.AfterOccurredAt, page.AfterRecordID) {
			items = append(items, record)
		}
	}
	return pagedFakeAuditRecords(items, page.Limit), nil
}

func (f *fakeAuditRepository) ListInventoryAuditRecords(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AuditRecordPageRequest) ([]audit.Record, error) {
	items := []audit.Record{}
	for _, record := range f.items {
		if record.TenantID.String() == tenantID.String() && record.InventoryID.String() == inventoryID.String() && fakeAuditRecordAfter(record, page.AfterOccurredAt, page.AfterRecordID) {
			items = append(items, record)
		}
	}
	return pagedFakeAuditRecords(items, page.Limit), nil
}

func pagedFakeAuditRecords(items []audit.Record, limit int) []audit.Record {
	sort.Slice(items, func(left int, right int) bool {
		return items[left].Before(items[right])
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func fakeAuditRecordAfter(record audit.Record, occurredAt time.Time, id audit.ID) bool {
	if occurredAt.IsZero() || id.String() == "" {
		return true
	}
	if record.OccurredAt.After(occurredAt) {
		return true
	}
	return record.OccurredAt.Equal(occurredAt) && record.ID.String() > id.String()
}

func auditRecord(id string, tenantID string, inventoryID string, action audit.Action) audit.Record {
	record, ok := audit.NewRecord(
		audit.ID(id),
		audit.TenantID(tenantID),
		audit.InventoryID(inventoryID),
		audit.PrincipalID("owner"),
		action,
		audit.SourceAPI,
		audit.TargetAsset,
		id+"-target",
		time.Now(),
		"",
		map[string]string{},
	)
	if !ok {
		panic("invalid test audit record")
	}
	return record
}

func (f *fakeOutbox) SaveTenantAndEnqueueOwnerGrant(_ context.Context, eventID string, item tenant.Tenant, principal identity.Principal, auditRecord audit.Record) error {
	f.events = append(f.events, ports.AuthorizationOutboxEvent{
		ID:          eventID,
		Kind:        ports.AuthorizationOutboxGrantTenantOwner,
		PrincipalID: principal.ID,
		TenantID:    item.ID,
	})
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeOutbox) SaveInventoryAndEnqueueOwnerGrant(_ context.Context, eventID string, item inventory.Inventory, tenantID tenant.ID, principal identity.Principal, auditRecord audit.Record) error {
	f.events = append(f.events, ports.AuthorizationOutboxEvent{
		ID:          eventID,
		Kind:        ports.AuthorizationOutboxGrantInventoryOwner,
		PrincipalID: principal.ID,
		TenantID:    tenantID,
		InventoryID: item.ID,
	})
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeOutbox) ClaimPendingAuthorizationOutboxEvents(_ context.Context, claimID string, limit int, leaseUntil time.Time) ([]ports.AuthorizationOutboxEvent, error) {
	if limit <= 0 {
		limit = 25
	}
	now := time.Now()
	events := make([]ports.AuthorizationOutboxEvent, 0, len(f.events))
	for index, event := range f.events {
		if !event.DeadLetteredAt.IsZero() {
			continue
		}
		if !event.ClaimedUntil.IsZero() && event.ClaimedUntil.After(now) {
			continue
		}
		event.ClaimID = claimID
		event.ClaimedUntil = leaseUntil
		f.events[index] = event
		events = append(events, event)
		if len(events) == limit {
			break
		}
	}
	return events, nil
}

func (f *fakeOutbox) MarkAuthorizationOutboxEventProcessed(_ context.Context, eventID string, claimID string) error {
	for index, event := range f.events {
		if event.ID == eventID && event.ClaimID == claimID {
			f.processed = append(f.processed, eventID)
			f.events = append(f.events[:index], f.events[index+1:]...)
			return nil
		}
	}
	return ports.ErrAuthorizationOutboxClaimLost
}

func (f *fakeOutbox) MarkAuthorizationOutboxEventFailed(_ context.Context, eventID string, claimID string, _ string) error {
	for index, event := range f.events {
		if event.ID == eventID && event.ClaimID == claimID {
			f.failed = append(f.failed, eventID)
			event.ClaimID = ""
			event.ClaimedUntil = time.Time{}
			f.events[index] = event
			return nil
		}
	}
	return ports.ErrAuthorizationOutboxClaimLost
}

func (f *fakeOutbox) MarkAuthorizationOutboxEventDeadLettered(_ context.Context, eventID string, claimID string, reason string) error {
	for index, event := range f.events {
		if event.ID == eventID && event.ClaimID == claimID {
			f.deadLettered = append(f.deadLettered, eventID)
			event.DeadLetteredAt = time.Now()
			event.DeadLetterReason = reason
			event.ClaimID = ""
			event.ClaimedUntil = time.Time{}
			f.events[index] = event
			return nil
		}
	}
	return ports.ErrAuthorizationOutboxClaimLost
}

func (f *fakeInventoryRepository) SaveInventory(context.Context, inventory.Inventory) error {
	return nil
}

func (f *fakeInventoryRepository) SaveInventoryAccessGrantAndEnqueue(_ context.Context, _ string, grant ports.InventoryAccessGrant, auditRecord audit.Record) error {
	for _, existing := range f.accessGrants {
		if existing.TenantID == grant.TenantID && existing.InventoryID == grant.InventoryID && existing.CursorKey() == grant.CursorKey() {
			return nil
		}
	}
	f.accessGrants = append(f.accessGrants, grant)
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeInventoryRepository) ListInventoryAccessGrants(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.InventoryAccessGrantPageRequest) ([]ports.InventoryAccessGrant, error) {
	items := []ports.InventoryAccessGrant{}
	for _, grant := range f.accessGrants {
		key := grant.CursorKey()
		if grant.TenantID == tenantID && grant.InventoryID == inventoryID && key > page.AfterGrantKey {
			items = append(items, grant)
		}
	}
	sort.Slice(items, func(left int, right int) bool {
		return items[left].CursorKey() < items[right].CursorKey()
	})
	if page.Limit > 0 && len(items) > page.Limit {
		items = items[:page.Limit]
	}
	return items, nil
}

func (f *fakeInventoryRepository) ListInventoriesByTenant(_ context.Context, tenantID inventory.TenantID, page ports.InventoryListPageRequest) ([]inventory.Inventory, error) {
	f.calls++
	f.limits = append(f.limits, page.Limit)
	items := []inventory.Inventory{}
	for _, item := range f.items {
		if item.TenantID == tenantID && item.ID.String() > page.AfterInventoryID.String() {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(left int, right int) bool {
		return items[left].ID.String() < items[right].ID.String()
	})
	if page.Limit > 0 && len(items) > page.Limit {
		items = items[:page.Limit]
	}
	return items, nil
}

func (f *fakeCustomFieldRepository) SaveCustomFieldDefinition(_ context.Context, definition customfield.Definition, auditRecord audit.Record) error {
	for _, existing := range f.items {
		if customfield.DefinitionsConflict(existing, definition) {
			return ports.ErrConflict
		}
	}
	f.items = append(f.items, definition)
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeCustomFieldRepository) ListTenantCustomFieldDefinitions(_ context.Context, tenantID tenant.ID, page ports.CustomFieldDefinitionPageRequest) ([]customfield.Definition, error) {
	items := []customfield.Definition{}
	for _, item := range f.items {
		if item.TenantID.String() == tenantID.String() && item.Scope == customfield.ScopeTenant && item.CursorKey() > page.AfterDefinitionKey {
			items = append(items, item)
		}
	}
	return pagedFakeCustomFieldDefinitions(items, page.Limit), nil
}

func (f *fakeCustomFieldRepository) ListInventoryCustomFieldDefinitions(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.CustomFieldDefinitionPageRequest) ([]customfield.Definition, error) {
	items := []customfield.Definition{}
	for _, item := range f.items {
		if item.TenantID.String() != tenantID.String() || item.CursorKey() <= page.AfterDefinitionKey {
			continue
		}
		if item.Scope == customfield.ScopeTenant || item.InventoryID.String() == inventoryID.String() {
			items = append(items, item)
		}
	}
	return pagedFakeCustomFieldDefinitions(items, page.Limit), nil
}

func (f *fakeCustomFieldRepository) ListEffectiveCustomFieldDefinitions(ctx context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID) ([]customfield.Definition, error) {
	if inventoryID.String() == "" {
		return f.ListTenantCustomFieldDefinitions(ctx, tenantID, ports.CustomFieldDefinitionPageRequest{})
	}
	return f.ListInventoryCustomFieldDefinitions(ctx, tenantID, inventoryID, ports.CustomFieldDefinitionPageRequest{})
}

func pagedFakeCustomFieldDefinitions(items []customfield.Definition, limit int) []customfield.Definition {
	sort.Slice(items, func(left int, right int) bool {
		return items[left].CursorKey() < items[right].CursorKey()
	})
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

type selectiveInventoryAuthorizer struct {
	forbidden map[inventory.InventoryID]struct{}
}

func (s *selectiveInventoryAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	return nil
}

func (s *selectiveInventoryAuthorizer) CheckInventory(_ context.Context, _ identity.Principal, _ ports.InventoryPermission, inventoryID inventory.InventoryID) error {
	if _, ok := s.forbidden[inventoryID]; ok {
		return ports.ErrForbidden
	}
	return nil
}

func (s *selectiveInventoryAuthorizer) GrantTenantOwner(context.Context, identity.Principal, tenant.ID) error {
	return nil
}

func (s *selectiveInventoryAuthorizer) GrantInventoryOwner(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (s *selectiveInventoryAuthorizer) GrantInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (s *selectiveInventoryAuthorizer) GrantInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return nil
}

func (f *fakeAssetRepository) CreateAsset(_ context.Context, item asset.Asset, auditRecord audit.Record) error {
	if f.items == nil {
		f.items = map[asset.ID]asset.Asset{}
	}
	if _, exists := f.items[item.ID]; exists {
		return errors.New("asset already exists")
	}
	if item.ParentAssetID.String() != "" {
		parent, ok := f.items[item.ParentAssetID]
		if !ok || parent.TenantID != item.TenantID || parent.InventoryID != item.InventoryID || !parent.Kind.CanContainChildren() {
			return ports.ErrForbidden
		}
	}
	f.items[item.ID] = item
	f.auditRecords = append(f.auditRecords, auditRecord)
	return nil
}

func (f *fakeAssetRepository) UpdateAsset(_ context.Context, item asset.Asset, auditRecords []audit.Record) error {
	if f.items == nil {
		f.items = map[asset.ID]asset.Asset{}
	}
	existing, exists := f.items[item.ID]
	if !exists || existing.TenantID != item.TenantID || existing.InventoryID != item.InventoryID {
		return ports.ErrForbidden
	}
	if existing.Kind != item.Kind || existing.LifecycleState != item.LifecycleState {
		return ports.ErrForbidden
	}
	if item.ParentAssetID.String() != "" {
		parent, ok := f.items[item.ParentAssetID]
		if !ok || parent.TenantID != item.TenantID || parent.InventoryID != item.InventoryID || !parent.Kind.CanContainChildren() || parent.LifecycleState != asset.LifecycleStateActive {
			return ports.ErrForbidden
		}
		if parent.ID == item.ID {
			return ports.ErrForbidden
		}
		for current := parent; current.ParentAssetID.String() != ""; {
			next, ok := f.items[current.ParentAssetID]
			if !ok || next.TenantID != item.TenantID || next.InventoryID != item.InventoryID {
				return ports.ErrForbidden
			}
			if next.ID == item.ID {
				return ports.ErrForbidden
			}
			current = next
		}
	}
	f.items[item.ID] = item
	f.auditRecords = append(f.auditRecords, auditRecords...)
	return nil
}

func (f *fakeAssetRepository) AssetByID(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, assetID asset.ID) (asset.Asset, bool, error) {
	item, ok := f.items[assetID]
	if !ok || item.TenantID != asset.TenantID(tenantID.String()) || item.InventoryID != asset.InventoryID(inventoryID.String()) {
		return asset.Asset{}, false, nil
	}
	return item, true, nil
}

func (f *fakeAssetRepository) ListAssetsByInventory(_ context.Context, tenantID tenant.ID, inventoryID inventory.InventoryID, page ports.AssetListPageRequest) ([]asset.Asset, error) {
	items := []asset.Asset{}
	for _, item := range f.items {
		if item.TenantID == asset.TenantID(tenantID.String()) && item.InventoryID == asset.InventoryID(inventoryID.String()) && item.ID.String() > page.AfterAssetID.String() {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(left int, right int) bool {
		return items[left].ID.String() < items[right].ID.String()
	})
	if page.Limit > 0 && len(items) > page.Limit {
		items = items[:page.Limit]
	}
	return items, nil
}

type fakeObserver struct {
	events []ports.Event
}

func (f *fakeObserver) Record(_ context.Context, event ports.Event) {
	f.events = append(f.events, event)
}

func (f *fakeObserver) hasEvent(name ports.EventName) bool {
	for _, event := range f.events {
		if event.Name == name {
			return true
		}
	}
	return false
}

func (f *fakeObserver) eventNamed(name ports.EventName) (ports.Event, bool) {
	for _, event := range f.events {
		if event.Name == name {
			return event, true
		}
	}
	return ports.Event{}, false
}

type fakeIDGenerator struct {
	ids []string
}

func (f *fakeIDGenerator) NewID() string {
	if len(f.ids) == 0 {
		return "fixed-id"
	}
	id := f.ids[0]
	f.ids = f.ids[1:]
	return id
}

func newAppTestGORMStore(t *testing.T, ctx context.Context) gormstore.Store {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite fake: %v", err)
	}
	if err := gormstore.Migrate(ctx, db); err != nil {
		t.Fatalf("migrate sqlite fake: %v", err)
	}

	return gormstore.NewStore(db)
}

func inventoryItem(id string, tenantID string, name string) inventory.Inventory {
	inventoryName, ok := inventory.NewName(name)
	if !ok {
		panic("invalid test inventory name")
	}
	return inventory.Inventory{
		ID:       inventory.InventoryID(id),
		TenantID: inventory.TenantID(tenantID),
		Name:     inventoryName,
	}
}

func customFieldDefinition(id string, tenantID string, inventoryID string, scope customfield.Scope, keyValue string, fieldType customfield.FieldType, rawOptions []string) customfield.Definition {
	definitionID, ok := customfield.NewID(id)
	if !ok {
		panic("invalid custom field definition id")
	}
	key, ok := customfield.NewKey(keyValue)
	if !ok {
		panic("invalid custom field key")
	}
	displayName, ok := customfield.NewDisplayName("Field " + keyValue)
	if !ok {
		panic("invalid custom field display name")
	}
	options := make([]customfield.Key, 0, len(rawOptions))
	for _, raw := range rawOptions {
		option, ok := customfield.NewKey(raw)
		if !ok {
			panic("invalid custom field enum option")
		}
		options = append(options, option)
	}
	definition, ok := customfield.NewDefinition(
		definitionID,
		customfield.TenantID(tenantID),
		customfield.InventoryID(inventoryID),
		scope,
		key,
		displayName,
		fieldType,
		options,
		customfield.ApplicabilityAllAssets,
		nil,
	)
	if !ok {
		panic("invalid custom field definition")
	}
	return definition
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
