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

func TestGenerationQueuesOnlyEnabledRecipientDevices(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
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

			service := New(Dependencies{Deliveries: store, Devices: store, Authorizer: auth, Inventories: store, Types: store, Assets: store, Inbox: store, Preferences: store, Audit: store, IDs: &inboxIDs{}, Clock: inboxClock{now}})
			input := ScopeInput{Principal: owner, TenantID: "home", InventoryID: "main"}
			if _, err := service.InitializePreferences(ctx, input, "UTC"); err != nil {
				t.Fatal(err)
			}
			preferences, _, _ := store.NotificationPreferences(ctx, input.Scope())
			preferences.Settings.PushEnabled = enabled
			preferences.Revision++
			if err := store.SaveNotificationPreferences(ctx, preferences, preferences.Revision-1, audit.Record{ID: "enable-push", TenantID: "home", InventoryID: "main", PrincipalID: "owner"}); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 101; i++ {
				token, _ := notification.ParseDeviceToken(fmt.Sprintf("token-%03d", i))
				device := ports.NotificationDevice{ID: fmt.Sprintf("device-%03d", i), Scope: input.Scope(), InstallationID: fmt.Sprint(i), Transport: notification.PushFCM, Token: token, Active: true, Revision: 1, CreatedAt: now, UpdatedAt: now}
				if err := store.SaveNotificationDevice(ctx, device, 0, audit.Record{ID: audit.ID(fmt.Sprintf("device-audit-%03d", i)), TenantID: "home", InventoryID: "main", PrincipalID: "owner"}); err != nil {
					t.Fatal(err)
				}
			}

			for _, extra := range []struct {
				id        string
				principal string
				active    bool
			}{{"inactive", "owner", false}, {"foreign", "viewer", true}} {
				token, _ := notification.ParseDeviceToken(extra.id)
				scope := input.Scope()
				scope.PrincipalID = identity.PrincipalID(extra.principal)
				device := ports.NotificationDevice{ID: extra.id, Scope: scope, InstallationID: extra.id, Transport: notification.PushFCM, Token: token, Active: extra.active, Revision: 1, CreatedAt: now, UpdatedAt: now}
				if err := store.SaveNotificationDevice(ctx, device, 0, audit.Record{ID: audit.ID(extra.id + "-audit"), TenantID: "home", InventoryID: "main", PrincipalID: audit.PrincipalID(extra.principal)}); err != nil {
					t.Fatal(err)
				}
			}
			generated, err := service.GenerateRecipientPage(ctx, input, "", 100)
			if err != nil || generated.Created != 1 {
				t.Fatalf("generate: %+v %v", generated, err)
			}
			policy := notification.RetryPolicy{MaxAttempts: 2, InitialDelay: time.Second, MaximumDelay: time.Minute}
			count := 0
			for _, fence := range []string{"one", "two"} {
				rows, err := store.ClaimNotificationDeliveries(ctx, now, fence, time.Minute, policy, 100)
				if err != nil {
					t.Fatal(err)
				}
				count += len(rows)
				for _, row := range rows {
					if row.Scope != input.Scope() || row.DeviceRevision != 1 {
						t.Fatal("wrong recipient destination")
					}
				}
			}
			expected := 0
			if enabled {
				expected = 101
			}
			if count != expected {
				t.Fatalf("queued %d, want %d", count, expected)
			}

			lateToken, _ := notification.ParseDeviceToken("late-token")
			late := ports.NotificationDevice{ID: "late", Scope: input.Scope(), InstallationID: "late", Transport: notification.PushFCM, Token: lateToken, Active: true, Revision: 1, CreatedAt: now, UpdatedAt: now}
			if err := store.SaveNotificationDevice(ctx, late, 0, audit.Record{ID: "late-audit", TenantID: "home", InventoryID: "main", PrincipalID: "owner"}); err != nil {
				t.Fatal(err)
			}
			if _, err := service.GenerateRecipientPage(ctx, input, "", 100); err != nil {
				t.Fatal(err)
			}
			rows, err := store.ClaimNotificationDeliveries(ctx, now, "three", time.Minute, policy, 100)
			if err != nil || len(rows) != 0 {
				t.Fatal("duplicate generation queued more deliveries")
			}
		})
	}
}
