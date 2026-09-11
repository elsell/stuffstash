package notifications

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/adapters/memory"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/notification"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
	"time"
)

type acceptingPushTokens struct{}

func (acceptingPushTokens) NormalizeDeviceToken(_ context.Context, _ notification.PushTransport, token notification.DeviceToken) (notification.DeviceToken, error) {
	return token, nil
}
func TestDeviceCommandsAuthorizeReviseAndRevoke(t *testing.T) {
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
	viewer := identity.Principal{ID: "viewer"}
	if err := auth.GrantInventoryViewer(ctx, viewer, "home", "main"); err != nil {
		t.Fatal(err)
	}
	service := New(Dependencies{Authorizer: auth, Inventories: store, Preferences: store, Devices: store, PushTokens: acceptingPushTokens{}, Audit: store, IDs: &inboxIDs{}, Clock: inboxClock{time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)}})
	scope := ScopeInput{Principal: viewer, TenantID: "home", InventoryID: "main"}
	token, _ := notification.ParseDeviceToken("native-token")
	input := RegisterDeviceInput{InstallationID: "installation", Transport: notification.PushAPNS, Token: token}
	device, err := service.RegisterDevice(ctx, scope, input)
	if err != nil || device.Revision != 1 || !device.Active {
		t.Fatalf("register %+v %v", device, err)
	}
	duplicate, err := service.RegisterDevice(ctx, scope, input)
	if err != nil || duplicate.ID != device.ID || duplicate.Revision != 1 {
		t.Fatal("retry not idempotent")
	}
	input.Token, _ = notification.ParseDeviceToken("rotated-token")
	if _, err := service.RegisterDevice(ctx, scope, input); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("stale replacement %v", err)
	}
	input.Revision = 1
	updated, err := service.RegisterDevice(ctx, scope, input)
	if err != nil || updated.Revision != 2 {
		t.Fatal(err)
	}
	outsider := scope
	outsider.Principal = identity.Principal{ID: "outsider"}
	if _, err := service.RegisterDevice(ctx, outsider, input); err == nil {
		t.Fatal("unauthorized registration")
	}
	if _, err := service.RevokeDevice(ctx, outsider, device.ID, 2); err == nil {
		t.Fatal("unauthorized revoke")
	}
	peer := scope
	peer.Principal = identity.Principal{ID: "peer"}
	if err := auth.GrantInventoryViewer(ctx, peer.Principal, "home", "main"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RevokeDevice(ctx, peer, device.ID, 2); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("another member revoke: %v", err)
	}
	if _, err := service.GetDeviceByInstallation(ctx, peer, "installation"); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("another member lookup: %v", err)
	}
	if _, err := service.RevokeDevice(ctx, scope, device.ID, 1); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("stale revoke %v", err)
	}
	if err := auth.RevokeInventoryViewer(ctx, viewer, "home", "main"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetDeviceByInstallation(ctx, scope, "installation"); err != nil {
		t.Fatalf("own cleanup lookup after access loss: %v", err)
	}
	revoked, err := service.RevokeDevice(ctx, scope, device.ID, 2)
	if err != nil || revoked.Active || revoked.Revision != 3 {
		t.Fatalf("revoke %+v %v", revoked, err)
	}
	input.Revision = 0
	if _, err := service.RegisterDevice(ctx, peer, input); err != nil {
		t.Fatalf("next user registration: %v", err)
	}
}
