package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

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
		Inventories:         repository,
		InventoryUnitOfWork: repository,
		Audit:               &fakeAuditRepository{},
		Outbox:              &fakeOutbox{},
		DefaultPageLimit:    1,
		MaxPageLimit:        1,
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
		Inventories:         repository,
		InventoryUnitOfWork: repository,
		Audit:               &fakeAuditRepository{},
		Outbox:              &fakeOutbox{},
		DefaultPageLimit:    1,
		MaxPageLimit:        1,
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
