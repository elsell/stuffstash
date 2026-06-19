package httpserver

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type failingGrantAuthorizer struct{}

func (failingGrantAuthorizer) CheckTenant(context.Context, identity.Principal, ports.TenantPermission, tenant.ID) error {
	return ports.ErrForbidden
}

func (failingGrantAuthorizer) CheckInventory(context.Context, identity.Principal, ports.InventoryPermission, inventory.InventoryID) error {
	return ports.ErrForbidden
}

func (failingGrantAuthorizer) GrantTenantOwner(context.Context, identity.Principal, tenant.ID) error {
	return errors.New("spicedb unavailable")
}

func (failingGrantAuthorizer) GrantInventoryOwner(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return errors.New("spicedb unavailable")
}

func (failingGrantAuthorizer) GrantInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return errors.New("spicedb unavailable")
}

func (failingGrantAuthorizer) GrantInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return errors.New("spicedb unavailable")
}

func (failingGrantAuthorizer) RevokeInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return errors.New("spicedb unavailable")
}

func (failingGrantAuthorizer) RevokeInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return errors.New("spicedb unavailable")
}

type failingRevokeAuthorizer struct {
	delegate ports.Authorizer
}

func (f failingRevokeAuthorizer) CheckTenant(ctx context.Context, principal identity.Principal, permission ports.TenantPermission, tenantID tenant.ID) error {
	return f.delegate.CheckTenant(ctx, principal, permission, tenantID)
}

func (f failingRevokeAuthorizer) CheckInventory(ctx context.Context, principal identity.Principal, permission ports.InventoryPermission, inventoryID inventory.InventoryID) error {
	return f.delegate.CheckInventory(ctx, principal, permission, inventoryID)
}

func (f failingRevokeAuthorizer) GrantTenantOwner(ctx context.Context, principal identity.Principal, tenantID tenant.ID) error {
	return f.delegate.GrantTenantOwner(ctx, principal, tenantID)
}

func (f failingRevokeAuthorizer) GrantInventoryOwner(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return f.delegate.GrantInventoryOwner(ctx, principal, tenantID, inventoryID)
}

func (f failingRevokeAuthorizer) GrantInventoryViewer(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return f.delegate.GrantInventoryViewer(ctx, principal, tenantID, inventoryID)
}

func (f failingRevokeAuthorizer) GrantInventoryEditor(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return f.delegate.GrantInventoryEditor(ctx, principal, tenantID, inventoryID)
}

func (f failingRevokeAuthorizer) RevokeInventoryViewer(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return errors.New("spicedb unavailable")
}

func (f failingRevokeAuthorizer) RevokeInventoryEditor(context.Context, identity.Principal, tenant.ID, inventory.InventoryID) error {
	return errors.New("spicedb unavailable")
}

type failingTenantGrantAuthorizer struct {
	delegate ports.Authorizer
}

func (f failingTenantGrantAuthorizer) CheckTenant(ctx context.Context, principal identity.Principal, permission ports.TenantPermission, tenantID tenant.ID) error {
	return f.delegate.CheckTenant(ctx, principal, permission, tenantID)
}

func (f failingTenantGrantAuthorizer) CheckInventory(ctx context.Context, principal identity.Principal, permission ports.InventoryPermission, inventoryID inventory.InventoryID) error {
	return f.delegate.CheckInventory(ctx, principal, permission, inventoryID)
}

func (f failingTenantGrantAuthorizer) GrantTenantOwner(context.Context, identity.Principal, tenant.ID) error {
	return errors.New("spicedb unavailable")
}

func (f failingTenantGrantAuthorizer) GrantInventoryOwner(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return f.delegate.GrantInventoryOwner(ctx, principal, tenantID, inventoryID)
}

func (f failingTenantGrantAuthorizer) GrantInventoryViewer(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return f.delegate.GrantInventoryViewer(ctx, principal, tenantID, inventoryID)
}

func (f failingTenantGrantAuthorizer) GrantInventoryEditor(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return f.delegate.GrantInventoryEditor(ctx, principal, tenantID, inventoryID)
}

func (f failingTenantGrantAuthorizer) RevokeInventoryViewer(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return f.delegate.RevokeInventoryViewer(ctx, principal, tenantID, inventoryID)
}

func (f failingTenantGrantAuthorizer) RevokeInventoryEditor(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return f.delegate.RevokeInventoryEditor(ctx, principal, tenantID, inventoryID)
}

type inventoryAccessGrantResponse struct {
	TenantID     string `json:"tenantId"`
	InventoryID  string `json:"inventoryId"`
	PrincipalID  string `json:"principalId"`
	Relationship string `json:"relationship"`
}

type inventoryAccessGrantBody struct {
	Data inventoryAccessGrantResponse `json:"data"`
	Meta responseMeta                 `json:"meta"`
}

type inventoryAccessGrantListBody struct {
	Data []inventoryAccessGrantResponse `json:"data"`
	Meta responseMeta                   `json:"meta"`
}

func decodeInventoryAccessGrant(t *testing.T, response *httptest.ResponseRecorder) inventoryAccessGrantBody {
	t.Helper()

	var body inventoryAccessGrantBody
	decodeBody(t, response, &body)
	return body
}

func decodeInventoryAccessGrantList(t *testing.T, response *httptest.ResponseRecorder) inventoryAccessGrantListBody {
	t.Helper()

	var body inventoryAccessGrantListBody
	decodeBody(t, response, &body)
	return body
}

func assertInventoryAccessGrant(t *testing.T, grant inventoryAccessGrantResponse, tenantID string, inventoryID string, principalID string, relationship string) {
	t.Helper()

	if grant.TenantID != tenantID || grant.InventoryID != inventoryID || grant.PrincipalID != principalID || grant.Relationship != relationship {
		t.Fatalf("expected access grant %s/%s/%s/%s, got %+v", tenantID, inventoryID, principalID, relationship, grant)
	}
}
