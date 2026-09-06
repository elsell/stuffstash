package observability

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type observedAuthenticator struct {
	delegate  ports.Authenticator
	telemetry ports.Telemetry
}

func ObserveAuthenticator(delegate ports.Authenticator, telemetry ports.Telemetry) ports.Authenticator {
	if delegate == nil {
		return nil
	}
	if telemetry == nil {
		telemetry = ports.NoopTelemetry{}
	}
	return observedAuthenticator{delegate, telemetry}
}

func (a observedAuthenticator) Authenticate(ctx context.Context, authorizationHeader string) (principal identity.Principal, err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationAuthenticate)
	defer func() { finish(err) }()
	return a.delegate.Authenticate(ctx, authorizationHeader)
}

type observedAuthorizer struct {
	delegate  ports.Authorizer
	telemetry ports.Telemetry
}

func ObserveAuthorizer(delegate ports.Authorizer, telemetry ports.Telemetry) ports.Authorizer {
	if delegate == nil {
		return nil
	}
	if telemetry == nil {
		telemetry = ports.NoopTelemetry{}
	}
	return observedAuthorizer{delegate, telemetry}
}

func (a observedAuthorizer) CheckTenant(ctx context.Context, principal identity.Principal, permission ports.TenantPermission, tenantID tenant.ID) (err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationAuthorize)
	defer func() { finish(err) }()
	return a.delegate.CheckTenant(ctx, principal, permission, tenantID)
}

func (a observedAuthorizer) CheckInventory(ctx context.Context, principal identity.Principal, permission ports.InventoryPermission, inventoryID inventory.InventoryID) (err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationAuthorize)
	defer func() { finish(err) }()
	return a.delegate.CheckInventory(ctx, principal, permission, inventoryID)
}

func (a observedAuthorizer) ListViewableInventoryIDs(ctx context.Context, principal identity.Principal, tenantID tenant.ID, candidates []inventory.InventoryID) (result []inventory.InventoryID, err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationVisibility)
	defer func() { finish(err) }()
	return a.delegate.ListViewableInventoryIDs(ctx, principal, tenantID, candidates)
}

func (a observedAuthorizer) GrantTenantOwner(ctx context.Context, principal identity.Principal, tenantID tenant.ID) (err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationAccessChange)
	defer func() { finish(err) }()
	return a.delegate.GrantTenantOwner(ctx, principal, tenantID)
}

func (a observedAuthorizer) GrantInventoryOwner(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) (err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationAccessChange)
	defer func() { finish(err) }()
	return a.delegate.GrantInventoryOwner(ctx, principal, tenantID, inventoryID)
}

func (a observedAuthorizer) GrantInventoryViewer(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) (err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationAccessChange)
	defer func() { finish(err) }()
	return a.delegate.GrantInventoryViewer(ctx, principal, tenantID, inventoryID)
}

func (a observedAuthorizer) GrantInventoryEditor(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) (err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationAccessChange)
	defer func() { finish(err) }()
	return a.delegate.GrantInventoryEditor(ctx, principal, tenantID, inventoryID)
}

func (a observedAuthorizer) RevokeInventoryViewer(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) (err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationAccessChange)
	defer func() { finish(err) }()
	return a.delegate.RevokeInventoryViewer(ctx, principal, tenantID, inventoryID)
}

func (a observedAuthorizer) RevokeInventoryEditor(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) (err error) {
	ctx, finish := a.telemetry.Start(ctx, ports.OperationAccessChange)
	defer func() { finish(err) }()
	return a.delegate.RevokeInventoryEditor(ctx, principal, tenantID, inventoryID)
}
