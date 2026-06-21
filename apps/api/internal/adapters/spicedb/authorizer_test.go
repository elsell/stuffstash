package spicedb

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func TestAuthorizerChecksTenantPermission(t *testing.T) {
	gateway := &fakeGateway{
		permissionship: v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION,
	}
	authorizer := NewAuthorizer(gateway)

	err := authorizer.CheckTenant(context.Background(), principal("user-one"), ports.TenantPermissionCreateInventory, tenant.ID("tenant-one"))
	if err != nil {
		t.Fatalf("check tenant: %v", err)
	}

	request := gateway.checks[0]
	if request.Resource.ObjectType != "tenant" || request.Resource.ObjectId != "tenant-one" {
		t.Fatalf("unexpected resource: %+v", request.Resource)
	}
	if request.Permission != "create_inventory" {
		t.Fatalf("unexpected permission: %q", request.Permission)
	}
	if request.Subject.Object.ObjectType != "user" || request.Subject.Object.ObjectId != "user-one" {
		t.Fatalf("unexpected subject: %+v", request.Subject.Object)
	}
	if request.Consistency.GetFullyConsistent() != true {
		t.Fatalf("expected fully consistent permission check")
	}
}

func TestAuthorizerDeniesMissingPermission(t *testing.T) {
	gateway := &fakeGateway{
		permissionship: v1.CheckPermissionResponse_PERMISSIONSHIP_NO_PERMISSION,
	}
	authorizer := NewAuthorizer(gateway)

	err := authorizer.CheckInventory(context.Background(), principal("user-one"), ports.InventoryPermissionView, inventory.InventoryID("inventory-one"))
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestAuthorizerChecksCreateAssetPermission(t *testing.T) {
	gateway := &fakeGateway{
		permissionship: v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION,
	}
	authorizer := NewAuthorizer(gateway)

	err := authorizer.CheckInventory(context.Background(), principal("user-one"), ports.InventoryPermissionCreateAsset, inventory.InventoryID("inventory-one"))
	if err != nil {
		t.Fatalf("check create asset: %v", err)
	}

	request := gateway.checks[0]
	if request.Resource.ObjectType != "inventory" || request.Resource.ObjectId != "inventory-one" {
		t.Fatalf("unexpected resource: %+v", request.Resource)
	}
	if request.Permission != "create_asset" {
		t.Fatalf("unexpected permission: %q", request.Permission)
	}
	if request.Subject.Object.ObjectType != "user" || request.Subject.Object.ObjectId != "user-one" {
		t.Fatalf("unexpected subject: %+v", request.Subject.Object)
	}
}

func TestAuthorizerChecksEditAssetPermission(t *testing.T) {
	gateway := &fakeGateway{
		permissionship: v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION,
	}
	authorizer := NewAuthorizer(gateway)

	err := authorizer.CheckInventory(context.Background(), principal("user-one"), ports.InventoryPermissionEditAsset, inventory.InventoryID("inventory-one"))
	if err != nil {
		t.Fatalf("check edit asset: %v", err)
	}

	request := gateway.checks[0]
	if request.Resource.ObjectType != "inventory" || request.Resource.ObjectId != "inventory-one" {
		t.Fatalf("unexpected resource: %+v", request.Resource)
	}
	if request.Permission != "edit_asset" {
		t.Fatalf("unexpected permission: %q", request.Permission)
	}
	if request.Subject.Object.ObjectType != "user" || request.Subject.Object.ObjectId != "user-one" {
		t.Fatalf("unexpected subject: %+v", request.Subject.Object)
	}
}

func TestAuthorizerChecksSharePermission(t *testing.T) {
	gateway := &fakeGateway{
		permissionship: v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION,
	}
	authorizer := NewAuthorizer(gateway)

	err := authorizer.CheckInventory(context.Background(), principal("user-one"), ports.InventoryPermissionShare, inventory.InventoryID("inventory-one"))
	if err != nil {
		t.Fatalf("check share: %v", err)
	}

	request := gateway.checks[0]
	if request.Resource.ObjectType != "inventory" || request.Resource.ObjectId != "inventory-one" {
		t.Fatalf("unexpected resource: %+v", request.Resource)
	}
	if request.Permission != "share" {
		t.Fatalf("unexpected permission: %q", request.Permission)
	}
	if request.Subject.Object.ObjectType != "user" || request.Subject.Object.ObjectId != "user-one" {
		t.Fatalf("unexpected subject: %+v", request.Subject.Object)
	}
}

func TestAuthorizerPropagatesBackendFailure(t *testing.T) {
	expected := errors.New("backend unavailable")
	gateway := &fakeGateway{checkErr: expected}
	authorizer := NewAuthorizer(gateway)

	err := authorizer.CheckTenant(context.Background(), principal("user-one"), ports.TenantPermissionView, tenant.ID("tenant-one"))
	if !errors.Is(err, expected) {
		t.Fatalf("expected backend error, got %v", err)
	}
}

func TestAuthorizerGrantsTenantOwner(t *testing.T) {
	gateway := &fakeGateway{}
	authorizer := NewAuthorizer(gateway)

	err := authorizer.GrantTenantOwner(context.Background(), principal("user-one"), tenant.ID("tenant-one"))
	if err != nil {
		t.Fatalf("grant tenant owner: %v", err)
	}

	update := gateway.relationshipWrites[0].Updates[0]
	if update.Operation != v1.RelationshipUpdate_OPERATION_TOUCH {
		t.Fatalf("expected touch operation, got %s", update.Operation)
	}
	relationship := update.Relationship
	if relationship.Resource.ObjectType != "tenant" || relationship.Resource.ObjectId != "tenant-one" {
		t.Fatalf("unexpected resource: %+v", relationship.Resource)
	}
	if relationship.Relation != "owner" {
		t.Fatalf("unexpected relation: %q", relationship.Relation)
	}
	if relationship.Subject.Object.ObjectType != "user" || relationship.Subject.Object.ObjectId != "user-one" {
		t.Fatalf("unexpected subject: %+v", relationship.Subject.Object)
	}
}

func TestAuthorizerGrantsInventoryOwnerAndTenantLink(t *testing.T) {
	gateway := &fakeGateway{}
	authorizer := NewAuthorizer(gateway)

	err := authorizer.GrantInventoryOwner(context.Background(), principal("user-one"), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"))
	if err != nil {
		t.Fatalf("grant inventory owner: %v", err)
	}

	updates := gateway.relationshipWrites[0].Updates
	if len(updates) != 3 {
		t.Fatalf("expected 3 relationship updates, got %d", len(updates))
	}
	if updates[0].Relationship.Relation != "viewer" {
		t.Fatalf("expected viewer relationship first, got %q", updates[0].Relationship.Relation)
	}
	if updates[0].Relationship.Resource.ObjectType != "tenant" || updates[0].Relationship.Resource.ObjectId != "tenant-one" {
		t.Fatalf("unexpected tenant viewer resource: %+v", updates[0].Relationship.Resource)
	}
	if updates[0].Relationship.Subject.Object.ObjectType != "user" || updates[0].Relationship.Subject.Object.ObjectId != "user-one" {
		t.Fatalf("unexpected tenant viewer subject: %+v", updates[0].Relationship.Subject.Object)
	}
	if updates[1].Relationship.Relation != "tenant" {
		t.Fatalf("expected tenant relationship second, got %q", updates[1].Relationship.Relation)
	}
	if updates[1].Relationship.Subject.Object.ObjectType != "tenant" || updates[1].Relationship.Subject.Object.ObjectId != "tenant-one" {
		t.Fatalf("unexpected tenant subject: %+v", updates[1].Relationship.Subject.Object)
	}
	if updates[2].Relationship.Relation != "owner" {
		t.Fatalf("expected owner relationship third, got %q", updates[2].Relationship.Relation)
	}
	if updates[2].Relationship.Subject.Object.ObjectType != "user" || updates[2].Relationship.Subject.Object.ObjectId != "user-one" {
		t.Fatalf("unexpected owner subject: %+v", updates[2].Relationship.Subject.Object)
	}
}

func TestAuthorizerGrantsInventoryViewerAndTenantLink(t *testing.T) {
	assertDirectInventoryGrant(t, func(authorizer Authorizer) error {
		return authorizer.GrantInventoryViewer(context.Background(), principal("user-one"), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"))
	}, "viewer")
}

func TestAuthorizerGrantsInventoryEditorAndTenantLink(t *testing.T) {
	assertDirectInventoryGrant(t, func(authorizer Authorizer) error {
		return authorizer.GrantInventoryEditor(context.Background(), principal("user-one"), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"))
	}, "editor")
}

func TestAuthorizerRevokesDirectInventoryAccess(t *testing.T) {
	for _, item := range []struct {
		name         string
		revoke       func(Authorizer) error
		relationship string
	}{
		{name: "viewer", relationship: "viewer", revoke: func(authorizer Authorizer) error {
			return authorizer.RevokeInventoryViewer(context.Background(), principal("user-one"), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"))
		}},
		{name: "editor", relationship: "editor", revoke: func(authorizer Authorizer) error {
			return authorizer.RevokeInventoryEditor(context.Background(), principal("user-one"), tenant.ID("tenant-one"), inventory.InventoryID("inventory-one"))
		}},
	} {
		t.Run(item.name, func(t *testing.T) {
			gateway := &fakeGateway{}
			authorizer := NewAuthorizer(gateway)

			if err := item.revoke(authorizer); err != nil {
				t.Fatalf("revoke direct inventory access: %v", err)
			}

			updates := gateway.relationshipWrites[0].Updates
			if len(updates) != 1 {
				t.Fatalf("expected one relationship delete, got %d", len(updates))
			}
			if updates[0].Operation != v1.RelationshipUpdate_OPERATION_DELETE {
				t.Fatalf("expected delete operation, got %s", updates[0].Operation)
			}
			relationship := updates[0].Relationship
			if relationship.Resource.ObjectType != "inventory" || relationship.Resource.ObjectId != "inventory-one" {
				t.Fatalf("unexpected inventory resource: %+v", relationship.Resource)
			}
			if relationship.Relation != item.relationship {
				t.Fatalf("expected %q relationship, got %q", item.relationship, relationship.Relation)
			}
			if relationship.Subject.Object.ObjectType != "user" || relationship.Subject.Object.ObjectId != "user-one" {
				t.Fatalf("unexpected subject: %+v", relationship.Subject.Object)
			}
		})
	}
}

func TestAuthorizerBootstrapsSchema(t *testing.T) {
	gateway := &fakeGateway{}
	authorizer := NewAuthorizer(gateway)

	err := authorizer.BootstrapSchema(context.Background(), "definition user {}")
	if err != nil {
		t.Fatalf("bootstrap schema: %v", err)
	}
	if gateway.schemaWrites[0].Schema != "definition user {}" {
		t.Fatalf("unexpected schema: %q", gateway.schemaWrites[0].Schema)
	}
}

func TestAuthorizerListsViewableInventoryIDsWithLookupResources(t *testing.T) {
	gateway := &fakeGateway{
		lookupResponses: []*v1.LookupResourcesResponse{
			{
				ResourceObjectId: "inventory-three",
				Permissionship:   v1.LookupPermissionship_LOOKUP_PERMISSIONSHIP_HAS_PERMISSION,
			},
			{
				ResourceObjectId: "inventory-one",
				Permissionship:   v1.LookupPermissionship_LOOKUP_PERMISSIONSHIP_HAS_PERMISSION,
			},
			{
				ResourceObjectId: "inventory-outside-candidates",
				Permissionship:   v1.LookupPermissionship_LOOKUP_PERMISSIONSHIP_HAS_PERMISSION,
			},
			{
				ResourceObjectId: "inventory-two",
				Permissionship:   v1.LookupPermissionship_LOOKUP_PERMISSIONSHIP_CONDITIONAL_PERMISSION,
			},
		},
	}
	authorizer := NewAuthorizer(gateway)

	visible, err := authorizer.ListViewableInventoryIDs(context.Background(), principal("user-one"), tenant.ID("tenant-one"), []inventory.InventoryID{
		inventory.InventoryID("inventory-one"),
		inventory.InventoryID("inventory-two"),
		inventory.InventoryID("inventory-three"),
	})
	if err != nil {
		t.Fatalf("list viewable inventory ids: %v", err)
	}

	expected := []inventory.InventoryID{
		inventory.InventoryID("inventory-one"),
		inventory.InventoryID("inventory-three"),
	}
	if len(visible) != len(expected) {
		t.Fatalf("expected %d visible inventories, got %d: %#v", len(expected), len(visible), visible)
	}
	for index := range expected {
		if visible[index] != expected[index] {
			t.Fatalf("expected visible[%d] %q, got %q", index, expected[index], visible[index])
		}
	}
	if len(gateway.checks) != 0 {
		t.Fatalf("expected lookup resources instead of per-candidate checks, got %d checks", len(gateway.checks))
	}
	if len(gateway.lookupRequests) != 1 {
		t.Fatalf("expected one lookup request, got %d", len(gateway.lookupRequests))
	}
	request := gateway.lookupRequests[0]
	if request.ResourceObjectType != "inventory" {
		t.Fatalf("unexpected resource object type: %q", request.ResourceObjectType)
	}
	if request.Permission != "view" {
		t.Fatalf("unexpected permission: %q", request.Permission)
	}
	if request.Subject.Object.ObjectType != "user" || request.Subject.Object.ObjectId != "user-one" {
		t.Fatalf("unexpected subject: %+v", request.Subject.Object)
	}
	if request.Consistency.GetFullyConsistent() != true {
		t.Fatalf("expected fully consistent lookup")
	}
}

func TestAuthorizerSkipsLookupWhenNoInventoryCandidates(t *testing.T) {
	gateway := &fakeGateway{}
	authorizer := NewAuthorizer(gateway)

	visible, err := authorizer.ListViewableInventoryIDs(context.Background(), principal("user-one"), tenant.ID("tenant-one"), nil)
	if err != nil {
		t.Fatalf("list viewable inventory ids: %v", err)
	}
	if visible != nil {
		t.Fatalf("expected nil visible inventory ids, got %#v", visible)
	}
	if len(gateway.lookupRequests) != 0 {
		t.Fatalf("expected no lookup for empty candidates, got %d", len(gateway.lookupRequests))
	}
}

func TestAuthorizerPropagatesLookupResourcesFailure(t *testing.T) {
	expected := errors.New("lookup unavailable")
	gateway := &fakeGateway{lookupErr: expected}
	authorizer := NewAuthorizer(gateway)

	_, err := authorizer.ListViewableInventoryIDs(context.Background(), principal("user-one"), tenant.ID("tenant-one"), []inventory.InventoryID{
		inventory.InventoryID("inventory-one"),
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected lookup error, got %v", err)
	}
}

func TestAuthorizerPropagatesLookupResourcesStreamFailure(t *testing.T) {
	expected := errors.New("stream interrupted")
	gateway := &fakeGateway{lookupRecvErr: expected}
	authorizer := NewAuthorizer(gateway)

	_, err := authorizer.ListViewableInventoryIDs(context.Background(), principal("user-one"), tenant.ID("tenant-one"), []inventory.InventoryID{
		inventory.InventoryID("inventory-one"),
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected lookup stream error, got %v", err)
	}
}

func TestNewGatewayRequiresEndpoint(t *testing.T) {
	gateway, err := NewGateway("", "", false, "")
	if err == nil {
		t.Fatalf("expected missing endpoint error")
	}
	if gateway != nil {
		t.Fatalf("expected no gateway on configuration error")
	}
}

func TestNewGatewayAllowsUnauthenticatedLocalTesting(t *testing.T) {
	gateway, err := NewGateway("localhost:50051", "", false, "")
	if err != nil {
		t.Fatalf("create unauthenticated local gateway: %v", err)
	}
	t.Cleanup(func() {
		if err := gateway.Close(); err != nil {
			t.Fatalf("close gateway: %v", err)
		}
	})
}

func TestNewTLSConfigLoadsCustomCA(t *testing.T) {
	caPath := filepath.Join(t.TempDir(), "ca.crt")
	if err := os.WriteFile(caPath, []byte(testCACertificatePEM), 0o600); err != nil {
		t.Fatalf("write ca: %v", err)
	}

	config, err := newTLSConfig(caPath)
	if err != nil {
		t.Fatalf("create tls config: %v", err)
	}

	if config.MinVersion != tls.VersionTLS12 {
		t.Fatalf("expected TLS 1.2 minimum, got %d", config.MinVersion)
	}
	if config.RootCAs == nil {
		t.Fatalf("expected custom root CAs")
	}
}

func TestNewTLSConfigRejectsInvalidCA(t *testing.T) {
	caPath := filepath.Join(t.TempDir(), "ca.crt")
	if err := os.WriteFile(caPath, []byte("not a certificate"), 0o600); err != nil {
		t.Fatalf("write ca: %v", err)
	}

	_, err := newTLSConfig(caPath)
	if err == nil {
		t.Fatalf("expected invalid CA error")
	}
}

func principal(id string) identity.Principal {
	return identity.Principal{ID: identity.PrincipalID(id)}
}

func assertDirectInventoryGrant(t *testing.T, grant func(Authorizer) error, expectedInventoryRelation string) {
	t.Helper()

	gateway := &fakeGateway{}
	authorizer := NewAuthorizer(gateway)

	if err := grant(authorizer); err != nil {
		t.Fatalf("grant direct inventory access: %v", err)
	}

	updates := gateway.relationshipWrites[0].Updates
	if len(updates) != 3 {
		t.Fatalf("expected 3 relationship updates, got %d", len(updates))
	}
	if updates[0].Relationship.Resource.ObjectType != "tenant" || updates[0].Relationship.Resource.ObjectId != "tenant-one" {
		t.Fatalf("unexpected tenant viewer resource: %+v", updates[0].Relationship.Resource)
	}
	if updates[0].Relationship.Relation != "viewer" {
		t.Fatalf("expected tenant viewer relationship first, got %q", updates[0].Relationship.Relation)
	}
	if updates[0].Relationship.Subject.Object.ObjectType != "user" || updates[0].Relationship.Subject.Object.ObjectId != "user-one" {
		t.Fatalf("unexpected tenant viewer subject: %+v", updates[0].Relationship.Subject.Object)
	}
	if updates[1].Relationship.Resource.ObjectType != "inventory" || updates[1].Relationship.Resource.ObjectId != "inventory-one" {
		t.Fatalf("unexpected inventory tenant resource: %+v", updates[1].Relationship.Resource)
	}
	if updates[1].Relationship.Relation != "tenant" {
		t.Fatalf("expected inventory tenant relationship second, got %q", updates[1].Relationship.Relation)
	}
	if updates[1].Relationship.Subject.Object.ObjectType != "tenant" || updates[1].Relationship.Subject.Object.ObjectId != "tenant-one" {
		t.Fatalf("unexpected tenant subject: %+v", updates[1].Relationship.Subject.Object)
	}
	if updates[2].Relationship.Resource.ObjectType != "inventory" || updates[2].Relationship.Resource.ObjectId != "inventory-one" {
		t.Fatalf("unexpected inventory access resource: %+v", updates[2].Relationship.Resource)
	}
	if updates[2].Relationship.Relation != expectedInventoryRelation {
		t.Fatalf("expected inventory %q relationship third, got %q", expectedInventoryRelation, updates[2].Relationship.Relation)
	}
	if updates[2].Relationship.Subject.Object.ObjectType != "user" || updates[2].Relationship.Subject.Object.ObjectId != "user-one" {
		t.Fatalf("unexpected inventory access subject: %+v", updates[2].Relationship.Subject.Object)
	}
}

type fakeGateway struct {
	permissionship     v1.CheckPermissionResponse_Permissionship
	checkErr           error
	lookupErr          error
	lookupRecvErr      error
	writeErr           error
	schemaErr          error
	checks             []*v1.CheckPermissionRequest
	lookupRequests     []*v1.LookupResourcesRequest
	lookupResponses    []*v1.LookupResourcesResponse
	relationshipWrites []*v1.WriteRelationshipsRequest
	schemaWrites       []*v1.WriteSchemaRequest
}

func (f *fakeGateway) CheckPermission(_ context.Context, request *v1.CheckPermissionRequest) (*v1.CheckPermissionResponse, error) {
	f.checks = append(f.checks, request)
	if f.checkErr != nil {
		return nil, f.checkErr
	}
	return &v1.CheckPermissionResponse{Permissionship: f.permissionship}, nil
}

func (f *fakeGateway) LookupResources(_ context.Context, request *v1.LookupResourcesRequest) (lookupResourcesStream, error) {
	f.lookupRequests = append(f.lookupRequests, request)
	if f.lookupErr != nil {
		return nil, f.lookupErr
	}
	return &fakeLookupResourcesStream{
		responses: f.lookupResponses,
		recvErr:   f.lookupRecvErr,
	}, nil
}

func (f *fakeGateway) WriteRelationships(_ context.Context, request *v1.WriteRelationshipsRequest) (*v1.WriteRelationshipsResponse, error) {
	f.relationshipWrites = append(f.relationshipWrites, request)
	return &v1.WriteRelationshipsResponse{}, f.writeErr
}

func (f *fakeGateway) WriteSchema(_ context.Context, request *v1.WriteSchemaRequest) (*v1.WriteSchemaResponse, error) {
	f.schemaWrites = append(f.schemaWrites, request)
	return &v1.WriteSchemaResponse{}, f.schemaErr
}

type fakeLookupResourcesStream struct {
	responses []*v1.LookupResourcesResponse
	recvErr   error
	index     int
}

func (s *fakeLookupResourcesStream) Recv() (*v1.LookupResourcesResponse, error) {
	if s.recvErr != nil {
		return nil, s.recvErr
	}
	if s.index >= len(s.responses) {
		return nil, io.EOF
	}
	response := s.responses[s.index]
	s.index++
	return response, nil
}

const testCACertificatePEM = `-----BEGIN CERTIFICATE-----
MIIDpDCCAoygAwIBAgIUNryriY/J75dF/XfR5ZXqO5yFkJYwDQYJKoZIhvcNAQEL
BQAwTjELMAkGA1UEBhMCVVMxDjAMBgNVBAgMBVN0YXRlMQ0wCwYDVQQHDARDaXR5
MRAwDgYDVQQKDAdIb21lbGFiMQ4wDAYDVQQDDAVsZW5ueTAeFw0yNTEwMzAyMDA5
MDRaFw0zNTEwMjgyMDA5MDRaME4xCzAJBgNVBAYTAlVTMQ4wDAYDVQQIDAVTdGF0
ZTENMAsGA1UEBwwEQ2l0eTEQMA4GA1UECgwHSG9tZWxhYjEOMAwGA1UEAwwFbGVu
bnkwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCTatL9743mv7xQ6yHv
Tt6Nf1arvvxwh9ZI5SgJnd6cglTQr2bzOfNKwC63K/JoJ28bKrOc4p9YBvQlvQvA
j7QEcYoUxevC/V9CSo06mQpbZCa3Emhe8yTUBxwOW0rinZv5HrLuHgBAAa909fPf
Qi3m2T+F7uVOCRmjwiUxdBXn9/3I8FKE3JIUolCphcBgH/0e7Cs8GMoKYymGidfN
jDR59qhejkjbBihWZCCe8icTWNMGgBmpMyNNeJFxxZ7f3vpuR6MDHwcfnBtlpG8J
yEXaHdv13LAxNBW+M+LqZt0ZsojCBYCyhGudFqz7fO4SgdYZBSpbmd2DLOBKKokf
IGU/AgMBAAGjejB4MB0GA1UdDgQWBBRjzf2X95ZhToDaZvBQw9c2Jt2lUTAfBgNV
HSMEGDAWgBRjzf2X95ZhToDaZvBQw9c2Jt2lUTAPBgNVHRMBAf8EBTADAQH/MCUG
A1UdEQQeMByCBWxlbm55ggcqLmxvY2FshwTAqALkhwR/AAABMA0GCSqGSIb3DQEB
CwUAA4IBAQAvS0O1XwJFBJvPxWEjiJrcI+bPxxO6noPq2LPBLREpKMb4b+Mqq1/K
2MsRwZOtoreuU1xdSj3AUktTWh5+95CCyW8ClwfCgAjcYfyhKI+sHBVwTN8I8au0
AurFv2MUhq8xw1Q5lEpceeMouf5btu8NR/ykJSyj06REBROnl7dDfBJUW9apHg4d
1LVEmk8vqKro1g1rmIbPBVtYvpUsyMxzg2WqhqjCNIHHyu3iKdx1mGQkW1+ynfgS
HcoOxtVkkxpvvGT7J1qDCD8Fkd59cBKlyWdcaxbRS4WODgVylOaMHdybyJwb8Zm0
ttoTWy9Dk5K+U4SjA882EfsjA2ro6pkD
-----END CERTIFICATE-----`
