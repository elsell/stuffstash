package spicedb

import (
	"context"
	"errors"
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

func TestNewGatewayRequiresEndpoint(t *testing.T) {
	gateway, err := NewGateway("", "", false)
	if err == nil {
		t.Fatalf("expected missing endpoint error")
	}
	if gateway != nil {
		t.Fatalf("expected no gateway on configuration error")
	}
}

func TestNewGatewayAllowsUnauthenticatedLocalTesting(t *testing.T) {
	gateway, err := NewGateway("localhost:50051", "", false)
	if err != nil {
		t.Fatalf("create unauthenticated local gateway: %v", err)
	}
	t.Cleanup(func() {
		if err := gateway.Close(); err != nil {
			t.Fatalf("close gateway: %v", err)
		}
	})
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
	writeErr           error
	schemaErr          error
	checks             []*v1.CheckPermissionRequest
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

func (f *fakeGateway) WriteRelationships(_ context.Context, request *v1.WriteRelationshipsRequest) (*v1.WriteRelationshipsResponse, error) {
	f.relationshipWrites = append(f.relationshipWrites, request)
	return &v1.WriteRelationshipsResponse{}, f.writeErr
}

func (f *fakeGateway) WriteSchema(_ context.Context, request *v1.WriteSchemaRequest) (*v1.WriteSchemaResponse, error) {
	f.schemaWrites = append(f.schemaWrites, request)
	return &v1.WriteSchemaResponse{}, f.schemaErr
}
