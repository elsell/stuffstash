package httpserver

import (
	"context"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"net/http"
	"testing"
	"time"
)

func notificationHTTPFixture(t *testing.T) (*http.Server, *memory.Store) {
	t.Helper()
	ctx := context.Background()
	store := memory.NewStore()
	authorizer := memory.NewAuthorizer()
	application := newSeededTestAppWithStoreAndAuthorizer(t, seededState{tenants: []seedTenant{{id: "home", name: "Home", owner: "owner"}}, inventories: []seedInventory{{id: "main", tenantID: "home", name: "Main", owner: "owner"}}}, store, authorizer)
	if err := authorizer.GrantInventoryViewer(ctx, identity.Principal{ID: "viewer"}, "home", "main"); err != nil {
		t.Fatal(err)
	}
	kind := customfield.AssetType{ID: "medicine", TenantID: "home", InventoryID: "main", Scope: customfield.ScopeInventory, Key: "medicine", DisplayName: "Medicine", LifecycleState: customfield.AssetTypeLifecycleActive, ExpirationEnabled: true}
	if err := store.SaveCustomAssetType(ctx, kind, audit.Record{ID: "seed-notice-type"}); err != nil {
		t.Fatal(err)
	}
	date, _ := expirationdate.ParseDate("2001-01", expirationdate.Month)
	title, _ := asset.NewTitle("Bottle")
	item := asset.Asset{ID: "bottle", TenantID: "home", InventoryID: "main", Kind: asset.KindItem, CustomAssetTypeID: "medicine", Title: title, Expiration: date, LifecycleState: asset.LifecycleStateActive}
	if err := store.CreateAsset(ctx, item, audit.Record{ID: "seed-notice-asset"}, nil); err != nil {
		t.Fatal(err)
	}
	value := ports.NotificationRecord{ID: "notice", Scope: ports.NotificationScope{TenantID: "home", InventoryID: "main", PrincipalID: "owner"}, Milestone: notification.Milestone{AssetID: "bottle", Date: date, Kind: notification.MilestoneExpired}, CreatedAt: time.Date(2001, 2, 1, 0, 0, 0, 0, time.UTC)}
	if _, _, err := store.InsertNotification(ctx, value, audit.Record{ID: "seed-notice", TenantID: "home", InventoryID: "main", PrincipalID: "owner"}); err != nil {
		t.Fatal(err)
	}
	return NewServer(":0", application), store
}

func coverNotificationInboxScenarios(t *testing.T, coverage executedScenarioCoverage, adversarial bool) {
	t.Helper()
	server, _ := notificationHTTPFixture(t)
	const base = "/tenants/home/inventories/main/notifications"
	const template = "/tenants/{tenantId}/inventories/{inventoryId}/notifications"
	token, status := "Bearer dev:owner", http.StatusOK
	if adversarial {
		token, status = "Bearer dev:outsider", http.StatusForbidden
	}
	coverage.request(t, server, http.MethodGet, template, base, token, nil, status)
	coverage.request(t, server, http.MethodGet, template+"/unread-count", base+"/unread-count", token, nil, status)
	coverage.request(t, server, http.MethodPut, template+"/read-all", base+"/read-all", token, nil, status)
	coverage.request(t, server, http.MethodGet, template+"/{notificationId}", base+"/notice", token, nil, status)
	coverage.request(t, server, http.MethodPut, template+"/{notificationId}/read", base+"/notice/read", token, nil, status)
	coverage.request(t, server, http.MethodDelete, template+"/{notificationId}/read", base+"/notice/read", token, nil, status)
}
