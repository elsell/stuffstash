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

type pushSenderFake struct {
	before   func()
	messages []ports.NotificationPushMessage
	outcome  ports.NotificationPushOutcome
}

func (f *pushSenderFake) SendNotification(_ context.Context, message ports.NotificationPushMessage) (ports.NotificationPushOutcome, error) {
	if f.before != nil {
		f.before()
	}
	f.messages = append(f.messages, message)
	return f.outcome, nil
}
func deliveryFixture(t *testing.T, sender *pushSenderFake) (Service, *memory.Store, ScopeInput) {
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

	service := New(Dependencies{PushSender: sender, Deliveries: store, Devices: store, Authorizer: auth, Inventories: store, Types: store, Assets: store, Inbox: store, Preferences: store, Audit: store, IDs: &inboxIDs{}, Clock: inboxClock{now}})
	input := ScopeInput{Principal: owner, TenantID: "home", InventoryID: "main"}
	if _, err := service.InitializePreferences(ctx, input, "UTC"); err != nil {
		t.Fatal(err)
	}
	preferences, _, _ := store.NotificationPreferences(ctx, input.Scope())
	preferences.Settings.PushEnabled = true
	preferences.Revision++
	if err := store.SaveNotificationPreferences(ctx, preferences, preferences.Revision-1, audit.Record{ID: "enable-push", TenantID: "home", InventoryID: "main", PrincipalID: "owner"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1; i++ {
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

	if _, err := service.GenerateRecipientPage(ctx, input, "", 100); err != nil {
		t.Fatal(err)
	}
	return service, store, input
}
func TestDeliveryRevalidatesAndSendsGenericContent(t *testing.T) {
	for _, mode := range []string{"send", "disabled", "revoked", "invalid-token", "archived", "date-changed", "access-removed"} {
		t.Run(mode, func(t *testing.T) {
			sender := &pushSenderFake{outcome: ports.NotificationPushAccepted}
			service, store, input := deliveryFixture(t, sender)
			ctx := context.Background()
			if mode == "disabled" {
				p, _, _ := store.NotificationPreferences(ctx, input.Scope())
				p.Settings.PushEnabled = false
				p.Revision++
				if err := store.SaveNotificationPreferences(ctx, p, p.Revision-1, audit.Record{ID: "disable", TenantID: "home", InventoryID: "main", PrincipalID: "owner"}); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "revoked" {
				if _, err := service.RevokeDevice(ctx, input, "device-000", 1); err != nil {
					t.Fatal(err)
				}
			}

			if mode == "archived" || mode == "date-changed" {
				item, _, _ := store.AssetByID(ctx, input.TenantID, input.InventoryID, "bottle")
				if mode == "archived" {
					item.LifecycleState = asset.LifecycleStateArchived
					if err := store.UpdateAssetLifecycle(ctx, item, audit.Record{ID: "archive", TenantID: "home", InventoryID: "main"}, nil); err != nil {
						t.Fatal(err)
					}
				} else {
					item.Expiration, _ = expirationdate.ParseDate("2027-09-30", expirationdate.Day)
					if err := store.UpdateAsset(ctx, item, []audit.Record{{ID: "date-change"}}, nil); err != nil {
						t.Fatal(err)
					}
				}
			}
			if mode == "access-removed" {
				service.deps.Authorizer = memory.NewAuthorizer()
			}
			if mode == "invalid-token" {
				sender.outcome = ports.NotificationPushInvalidDevice
			}
			policy := notification.RetryPolicy{MaxAttempts: 2, InitialDelay: time.Second, MaximumDelay: time.Minute}
			if _, err := service.DeliverPage(ctx, 10, time.Minute, policy); err != nil {
				t.Fatal(err)
			}
			expected := 1
			if mode != "send" && mode != "invalid-token" {
				expected = 0
			}
			if len(sender.messages) != expected {
				t.Fatalf("sent %d", len(sender.messages))
			}
			if expected == 1 {
				message := sender.messages[0]
				if message.Title != "Stuff Stash" || message.Body != "An item is expiring soon." || message.Scope != input.Scope() || message.NotificationID == "" {
					t.Fatalf("unexpected push: %+v", message)
				}
			}
			if mode == "invalid-token" {
				device, _, _ := store.NotificationDeviceByID(ctx, input.Scope(), "device-000")
				if device.Active {
					t.Fatal("invalid device remained active")
				}
			}
		})
	}
}

func TestDeliveryRetriesProviderFailureWithBackoff(t *testing.T) {
	sender := &pushSenderFake{outcome: ports.NotificationPushRetry}
	service, _, _ := deliveryFixture(t, sender)
	ctx := context.Background()
	policy := notification.RetryPolicy{MaxAttempts: 2, InitialDelay: time.Second, MaximumDelay: time.Minute}
	if _, err := service.DeliverPage(ctx, 10, time.Minute, policy); err != nil {
		t.Fatal(err)
	}
	if _, err := service.DeliverPage(ctx, 10, time.Minute, policy); err != nil {
		t.Fatal(err)
	}
	if len(sender.messages) != 1 {
		t.Fatal("retried before backoff")
	}
	service.deps.Clock = inboxClock{service.deps.Clock.Now().Add(time.Second)}
	if _, err := service.DeliverPage(ctx, 10, time.Minute, policy); err != nil {
		t.Fatal(err)
	}
	service.deps.Clock = inboxClock{service.deps.Clock.Now().Add(time.Hour)}
	if _, err := service.DeliverPage(ctx, 10, time.Minute, policy); err != nil {
		t.Fatal(err)
	}
	if len(sender.messages) != 2 {
		t.Fatal("failed to honor retry budget")
	}
}

func TestInvalidTokenDoesNotRetireRotatedRegistration(t *testing.T) {
	sender := &pushSenderFake{outcome: ports.NotificationPushInvalidDevice}
	service, store, input := deliveryFixture(t, sender)
	ctx := context.Background()
	sender.before = func() {
		device, _, _ := store.NotificationDeviceByID(ctx, input.Scope(), "device-000")
		device.Revision++
		device.Token, _ = notification.ParseDeviceToken("replacement-token")
		if err := store.SaveNotificationDevice(ctx, device, 1, audit.Record{ID: "rotated-device", TenantID: "home", InventoryID: "main", PrincipalID: "owner"}); err != nil {
			t.Fatal(err)
		}
	}
	policy := notification.RetryPolicy{MaxAttempts: 2, InitialDelay: time.Second, MaximumDelay: time.Minute}
	if _, err := service.DeliverPage(ctx, 10, time.Minute, policy); err != nil {
		t.Fatal(err)
	}
	device, _, _ := store.NotificationDeviceByID(ctx, input.Scope(), "device-000")
	if !device.Active || device.Revision != 2 {
		t.Fatal("invalid old token retired replacement")
	}
}
