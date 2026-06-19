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
