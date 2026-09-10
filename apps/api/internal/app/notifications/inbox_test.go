package notifications

import (
	"context"
	"fmt"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/domain/asset"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/customfield"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/domainvalue/expirationdate"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

type inboxClock struct{ now time.Time }

func (c inboxClock) Now() time.Time { return c.now }

type inboxIDs struct{ n int }

func (g *inboxIDs) NewID() string { g.n++; return fmt.Sprintf("generated-%d", g.n) }

func TestNotificationItemRevalidatesAssetAndPersonalScope(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	auth := memory.NewAuthorizer()
	name, _ := tenant.NewName("Home")
	if err := store.SaveTenant(ctx, tenant.Tenant{ID: "home", Name: name}); err != nil {
		t.Fatal(err)
	}
	invName, _ := inventory.NewName("Main")
	if err := store.SaveInventory(ctx, inventory.Inventory{ID: "main", TenantID: "home", Name: invName}); err != nil {
		t.Fatal(err)
	}
	owner := identity.Principal{ID: "owner"}
	if err := auth.GrantInventoryOwner(ctx, owner, "home", "main"); err != nil {
		t.Fatal(err)
	}
	if err := auth.GrantInventoryViewer(ctx, identity.Principal{ID: "viewer"}, "home", "main"); err != nil {
		t.Fatal(err)
	}
	kind := customfield.AssetType{ID: "medicine", TenantID: "home", InventoryID: "main", Scope: customfield.ScopeInventory, Key: "medicine", DisplayName: "Medicine", LifecycleState: customfield.AssetTypeLifecycleActive, ExpirationEnabled: true}
	if err := store.SaveCustomAssetType(ctx, kind, audit.Record{ID: "seed-type"}); err != nil {
		t.Fatal(err)
	}
	date, _ := expirationdate.ParseDate("2026-09-30", expirationdate.Day)
	title, _ := asset.NewTitle("Bottle")
	item := asset.Asset{ID: "bottle", TenantID: "home", InventoryID: "main", Kind: asset.KindItem, CustomAssetTypeID: "medicine", Title: title, Expiration: date, LifecycleState: asset.LifecycleStateActive}
	if err := store.CreateAsset(ctx, item, audit.Record{ID: "seed-asset"}, nil); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	entry := ports.NotificationRecord{ID: "notice", Scope: ports.NotificationScope{TenantID: "home", InventoryID: "main", PrincipalID: "owner"}, Milestone: notification.Milestone{AssetID: "bottle", Date: date, Kind: notification.MilestoneUpcoming}, CreatedAt: now}
	if _, _, err := store.InsertNotification(ctx, entry, audit.Record{ID: "seed-notice", TenantID: "home", InventoryID: "main", PrincipalID: "owner"}); err != nil {
		t.Fatal(err)
	}
	service := New(Dependencies{Authorizer: auth, Inventories: store, Types: store, Assets: store, Inbox: store, Preferences: store, Audit: store, IDs: &inboxIDs{}, Clock: inboxClock{now}})
	input := ScopeInput{Principal: owner, TenantID: "home", InventoryID: "main"}
	view, err := service.GetNotification(ctx, input, "notice")
	if err != nil || view.Asset.Title.String() != "Bottle" {
		t.Fatalf("legitimate detail %+v %v", view, err)
	}
	page, err := service.ListInbox(ctx, input, "", 10, false)
	if err != nil || len(page.Items) != 1 || page.NextCursor != "" {
		t.Fatalf("inbox page %+v %v", page, err)
	}
	if _, err := service.ListInbox(ctx, input, "", 0, false); err == nil {
		t.Fatal("invalid page accepted")
	}
	for i := 0; i < 201; i++ {
		hidden := entry
		hidden.ID = fmt.Sprintf("withdrawn-%03d", i)
		hidden.Milestone.AssetID = fmt.Sprintf("missing-%03d", i)
		if _, _, err := store.InsertNotification(ctx, hidden, audit.Record{ID: audit.ID(fmt.Sprintf("seed-hidden-%03d", i)), TenantID: "home", InventoryID: "main", PrincipalID: "owner"}); err != nil {
			t.Fatal(err)
		}
	}
	sparse, err := service.ListInbox(ctx, input, "", 10, false)
	if err != nil || len(sparse.Items) != 0 || sparse.NextCursor == "" {
		t.Fatalf("sparse continuation %+v %v", sparse, err)
	}
	continued, err := service.ListInbox(ctx, input, sparse.NextCursor, 10, false)
	if err != nil || len(continued.Items) != 1 || continued.NextCursor != "" {
		t.Fatalf("sparse history lost visible entry %+v %v", continued, err)
	}
	foreign := input
	foreign.Principal.ID = "viewer"
	if _, err := service.GetNotification(ctx, foreign, "notice"); err == nil {
		t.Fatal("another member read personal notification")
	}
	if err := service.MarkRead(ctx, foreign, "notice"); err == nil {
		t.Fatal("another member marked notification read")
	}
	foreign = input
	foreign.TenantID = "elsewhere"
	if _, err := service.GetNotification(ctx, foreign, "notice"); err == nil {
		t.Fatal("cross tenant detail")
	}
	if err := service.MarkRead(ctx, input, "notice"); err != nil {
		t.Fatal(err)
	}
	read, _, _ := store.NotificationByID(ctx, entry.Scope, "notice")
	if read.ReadAt == nil {
		t.Fatal("read state not saved")
	}
	page, err = service.ListInbox(ctx, input, sparse.NextCursor, 10, true)
	if err != nil || len(page.Items) != 0 {
		t.Fatal("read notification included in unread list")
	}
	kind.ExpirationEnabled = false
	if err := store.UpdateCustomAssetType(ctx, kind, audit.Record{ID: "disable-type"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetNotification(ctx, input, "notice"); err == nil {
		t.Fatal("disabled type visible")
	}
	if err := service.MarkRead(ctx, input, "notice"); err == nil {
		t.Fatal("obsolete notification mutable")
	}
	page, err = service.ListInbox(ctx, input, sparse.NextCursor, 10, false)
	if err != nil || len(page.Items) != 0 {
		t.Fatal("obsolete notification included in list")
	}
	kind.ExpirationEnabled = true
	if err := store.UpdateCustomAssetType(ctx, kind, audit.Record{ID: "enable-type"}); err != nil {
		t.Fatal(err)
	}
	item.Expiration, _ = expirationdate.ParseDate("2027-09-30", expirationdate.Day)
	if err := store.UpdateAsset(ctx, item, []audit.Record{{ID: "change-date"}}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetNotification(ctx, input, "notice"); err == nil {
		t.Fatal("obsolete date visible")
	}
	item.Expiration = date
	if err := store.UpdateAsset(ctx, item, []audit.Record{{ID: "restore-date"}}, nil); err != nil {
		t.Fatal(err)
	}
	item.LifecycleState = asset.LifecycleStateArchived
	if err := store.UpdateAssetLifecycle(ctx, item, audit.Record{ID: "archive-item", TenantID: "home", InventoryID: "main"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetNotification(ctx, input, "notice"); err == nil {
		t.Fatal("archived asset visible")
	}

	if _, err := service.InitializePreferences(ctx, input, "UTC"); err != nil {
		t.Fatal(err)
	}
	item.ID = "new-bottle"
	item.LifecycleState = asset.LifecycleStateActive
	if err := store.CreateAsset(ctx, item, audit.Record{ID: "new-bottle-create"}, nil); err != nil {
		t.Fatal(err)
	}
	generated, err := service.GenerateRecipientPage(ctx, input, "", 100)
	if err != nil || generated.Created != 1 || generated.NextCursor != "" {
		t.Fatalf("generation %+v %v", generated, err)
	}
	generated, err = service.GenerateRecipientPage(ctx, input, "", 100)
	if err != nil || generated.Created != 0 {
		t.Fatal("generation duplicated milestone")
	}
	outsider := input
	outsider.Principal.ID = "outsider"
	if _, err := service.GenerateRecipientPage(ctx, outsider, "", 100); err == nil {
		t.Fatal("unauthorized recipient generated")
	}
	if _, err := service.GenerateRecipientPage(ctx, input, "", 0); err == nil {
		t.Fatal("unbounded generation accepted")
	}

	item.ID = "new-bottle-z"
	if err := store.CreateAsset(ctx, item, audit.Record{ID: "new-bottle-z-create"}, nil); err != nil {
		t.Fatal(err)
	}
	first, err := service.GenerateRecipientPage(ctx, input, "", 1)
	if err != nil || first.NextCursor != "new-bottle" || first.Created != 0 {
		t.Fatalf("generation first page %+v %v", first, err)
	}
	second, err := service.GenerateRecipientPage(ctx, input, first.NextCursor, 1)
	if err != nil || second.NextCursor != "" || second.Created != 1 {
		t.Fatalf("generation continuation %+v %v", second, err)
	}
	viewerInput := input
	viewerInput.Principal.ID = "viewer"
	if _, err := service.InitializePreferences(ctx, viewerInput, "UTC"); err != nil {
		t.Fatal(err)
	}
	if err := auth.RevokeInventoryViewer(ctx, viewerInput.Principal, "home", "main"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GenerateRecipientPage(ctx, viewerInput, "", 100); err == nil {
		t.Fatal("revoked registered recipient generated")
	}

	cursor := GenerationCursor{}
	completed := false
	for i := 0; i < 10; i++ {
		next, done, err := service.GenerateSweepPage(ctx, cursor, 1)
		if err != nil {
			t.Fatal(err)
		}
		cursor = next
		if done {
			completed = true
			break
		}
	}
	if !completed {
		t.Fatal("sweep did not advance through recipients and revoked membership")
	}

}
