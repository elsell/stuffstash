package spicedb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	authzed "github.com/authzed/authzed-go/v1"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/inventory"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	objectTypeUser      = "user"
	objectTypeTenant    = "tenant"
	objectTypeInventory = "inventory"

	relationOwner  = "owner"
	relationEditor = "editor"
	relationTenant = "tenant"
	relationViewer = "viewer"
)

type Gateway interface {
	CheckPermission(ctx context.Context, request *v1.CheckPermissionRequest) (*v1.CheckPermissionResponse, error)
	LookupResources(ctx context.Context, request *v1.LookupResourcesRequest) (lookupResourcesStream, error)
	WriteRelationships(ctx context.Context, request *v1.WriteRelationshipsRequest) (*v1.WriteRelationshipsResponse, error)
	WriteSchema(ctx context.Context, request *v1.WriteSchemaRequest) (*v1.WriteSchemaResponse, error)
}

type lookupResourcesStream interface {
	Recv() (*v1.LookupResourcesResponse, error)
}

type Authorizer struct {
	gateway Gateway
}

func NewAuthorizer(gateway Gateway) Authorizer {
	return Authorizer{gateway: gateway}
}

type ClientGateway struct {
	client *authzed.Client
}

func NewGateway(endpoint string, presharedKey string, tlsEnabled bool, caPath string) (*ClientGateway, error) {
	endpoint = strings.TrimSpace(endpoint)
	presharedKey = strings.TrimSpace(presharedKey)
	if endpoint == "" {
		return nil, errors.New("spicedb endpoint is required")
	}

	transport := grpc.WithTransportCredentials(insecure.NewCredentials())
	if tlsEnabled {
		tlsConfig, err := newTLSConfig(caPath)
		if err != nil {
			return nil, err
		}
		transport = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	}

	clientOptions := []grpc.DialOption{transport}
	if presharedKey != "" {
		clientOptions = append(clientOptions, grpc.WithPerRPCCredentials(bearerTokenCredentials{
			token:                    presharedKey,
			requireTransportSecurity: tlsEnabled,
		}))
	}

	client, err := authzed.NewClient(
		endpoint,
		clientOptions...,
	)
	if err != nil {
		return nil, err
	}

	return &ClientGateway{client: client}, nil
}

func newTLSConfig(caPath string) (*tls.Config, error) {
	config := &tls.Config{MinVersion: tls.VersionTLS12}
	caPath = strings.TrimSpace(caPath)
	if caPath == "" {
		return config, nil
	}
	certPEM, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("read spicedb ca certificate: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(certPEM) {
		return nil, errors.New("spicedb ca certificate must contain at least one PEM certificate")
	}
	config.RootCAs = pool
	return config, nil
}

func (g *ClientGateway) CheckPermission(ctx context.Context, request *v1.CheckPermissionRequest) (*v1.CheckPermissionResponse, error) {
	return g.client.CheckPermission(ctx, request)
}

func (g *ClientGateway) LookupResources(ctx context.Context, request *v1.LookupResourcesRequest) (lookupResourcesStream, error) {
	return g.client.LookupResources(ctx, request)
}

func (g *ClientGateway) WriteRelationships(ctx context.Context, request *v1.WriteRelationshipsRequest) (*v1.WriteRelationshipsResponse, error) {
	return g.client.WriteRelationships(ctx, request)
}

func (g *ClientGateway) WriteSchema(ctx context.Context, request *v1.WriteSchemaRequest) (*v1.WriteSchemaResponse, error) {
	return g.client.WriteSchema(ctx, request)
}

func (g *ClientGateway) Close() error {
	return g.client.Close()
}

func (a Authorizer) CheckTenant(ctx context.Context, principal identity.Principal, permission ports.TenantPermission, tenantID tenant.ID) error {
	return a.check(ctx, objectRef(objectTypeTenant, tenantID.String()), string(permission), userSubject(principal))
}

func (a Authorizer) CheckInventory(ctx context.Context, principal identity.Principal, permission ports.InventoryPermission, inventoryID inventory.InventoryID) error {
	return a.check(ctx, objectRef(objectTypeInventory, inventoryID.String()), string(permission), userSubject(principal))
}

func (a Authorizer) ListViewableInventoryIDs(ctx context.Context, principal identity.Principal, _ tenant.ID, candidates []inventory.InventoryID) ([]inventory.InventoryID, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	candidateSet := make(map[string]struct{}, len(candidates))
	for _, inventoryID := range candidates {
		candidateSet[inventoryID.String()] = struct{}{}
	}

	stream, err := a.gateway.LookupResources(ctx, &v1.LookupResourcesRequest{
		ResourceObjectType: objectTypeInventory,
		Permission:         string(ports.InventoryPermissionView),
		Subject:            userSubject(principal),
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true},
		},
	})
	if err != nil {
		return nil, err
	}

	visibleSet := make(map[string]struct{}, len(candidates))
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if response.GetPermissionship() != v1.LookupPermissionship_LOOKUP_PERMISSIONSHIP_HAS_PERMISSION {
			continue
		}
		resourceID := response.GetResourceObjectId()
		if _, ok := candidateSet[resourceID]; ok {
			visibleSet[resourceID] = struct{}{}
		}
	}

	visible := make([]inventory.InventoryID, 0, len(visibleSet))
	for _, inventoryID := range candidates {
		if _, ok := visibleSet[inventoryID.String()]; ok {
			visible = append(visible, inventoryID)
		}
	}

	return visible, nil
}

func (a Authorizer) GrantTenantOwner(ctx context.Context, principal identity.Principal, tenantID tenant.ID) error {
	return a.touchRelationships(ctx, relationship(
		objectRef(objectTypeTenant, tenantID.String()),
		relationOwner,
		userSubject(principal),
	))
}

func (a Authorizer) GrantInventoryOwner(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return a.touchRelationships(ctx,
		relationship(
			objectRef(objectTypeTenant, tenantID.String()),
			relationViewer,
			userSubject(principal),
		),
		relationship(
			objectRef(objectTypeInventory, inventoryID.String()),
			relationTenant,
			objectSubject(objectTypeTenant, tenantID.String()),
		),
		relationship(
			objectRef(objectTypeInventory, inventoryID.String()),
			relationOwner,
			userSubject(principal),
		),
	)
}

func (a Authorizer) GrantInventoryViewer(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return a.grantDirectInventoryAccess(ctx, principal, tenantID, inventoryID, relationViewer)
}

func (a Authorizer) GrantInventoryEditor(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID) error {
	return a.grantDirectInventoryAccess(ctx, principal, tenantID, inventoryID, relationEditor)
}

func (a Authorizer) RevokeInventoryViewer(ctx context.Context, principal identity.Principal, _ tenant.ID, inventoryID inventory.InventoryID) error {
	return a.deleteRelationships(ctx, relationship(
		objectRef(objectTypeInventory, inventoryID.String()),
		relationViewer,
		userSubject(principal),
	))
}

func (a Authorizer) RevokeInventoryEditor(ctx context.Context, principal identity.Principal, _ tenant.ID, inventoryID inventory.InventoryID) error {
	return a.deleteRelationships(ctx, relationship(
		objectRef(objectTypeInventory, inventoryID.String()),
		relationEditor,
		userSubject(principal),
	))
}

func (a Authorizer) grantDirectInventoryAccess(ctx context.Context, principal identity.Principal, tenantID tenant.ID, inventoryID inventory.InventoryID, relation string) error {
	return a.touchRelationships(ctx,
		relationship(
			objectRef(objectTypeTenant, tenantID.String()),
			relationViewer,
			userSubject(principal),
		),
		relationship(
			objectRef(objectTypeInventory, inventoryID.String()),
			relationTenant,
			objectSubject(objectTypeTenant, tenantID.String()),
		),
		relationship(
			objectRef(objectTypeInventory, inventoryID.String()),
			relation,
			userSubject(principal),
		),
	)
}

func (a Authorizer) BootstrapSchema(ctx context.Context, schema string) error {
	_, err := a.gateway.WriteSchema(ctx, &v1.WriteSchemaRequest{Schema: schema})
	return err
}

func (a Authorizer) check(ctx context.Context, resource *v1.ObjectReference, permission string, subject *v1.SubjectReference) error {
	response, err := a.gateway.CheckPermission(ctx, &v1.CheckPermissionRequest{
		Resource:   resource,
		Permission: permission,
		Subject:    subject,
		Consistency: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true},
		},
	})
	if err != nil {
		return err
	}
	if response.GetPermissionship() != v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
		return ports.ErrForbidden
	}

	return nil
}

func (a Authorizer) touchRelationships(ctx context.Context, relationships ...*v1.Relationship) error {
	updates := make([]*v1.RelationshipUpdate, 0, len(relationships))
	for _, item := range relationships {
		updates = append(updates, &v1.RelationshipUpdate{
			Operation:    v1.RelationshipUpdate_OPERATION_TOUCH,
			Relationship: item,
		})
	}

	_, err := a.gateway.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{Updates: updates})
	return err
}

func (a Authorizer) deleteRelationships(ctx context.Context, relationships ...*v1.Relationship) error {
	updates := make([]*v1.RelationshipUpdate, 0, len(relationships))
	for _, item := range relationships {
		updates = append(updates, &v1.RelationshipUpdate{
			Operation:    v1.RelationshipUpdate_OPERATION_DELETE,
			Relationship: item,
		})
	}

	_, err := a.gateway.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{Updates: updates})
	return err
}

func relationship(resource *v1.ObjectReference, relation string, subject *v1.SubjectReference) *v1.Relationship {
	return &v1.Relationship{
		Resource: resource,
		Relation: relation,
		Subject:  subject,
	}
}

func objectRef(objectType string, objectID string) *v1.ObjectReference {
	return &v1.ObjectReference{
		ObjectType: objectType,
		ObjectId:   objectID,
	}
}

func userSubject(principal identity.Principal) *v1.SubjectReference {
	return objectSubject(objectTypeUser, principal.ID.String())
}

func objectSubject(objectType string, objectID string) *v1.SubjectReference {
	return &v1.SubjectReference{
		Object: objectRef(objectType, objectID),
	}
}

type bearerTokenCredentials struct {
	token                    string
	requireTransportSecurity bool
}

func (c bearerTokenCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + c.token}, nil
}

func (c bearerTokenCredentials) RequireTransportSecurity() bool {
	return c.requireTransportSecurity
}
