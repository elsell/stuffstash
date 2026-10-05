package ports

import "context"

type AccessRelationship string

const (
	AccessViewer AccessRelationship = "viewer"
	AccessEditor AccessRelationship = "editor"
)

type AccessGrant struct {
	TenantID     string             `json:"tenantId"`
	InventoryID  string             `json:"inventoryId"`
	PrincipalID  string             `json:"principalId"`
	Relationship AccessRelationship `json:"relationship"`
}
type AccessGrantsAPI interface {
	AccessGrants(context.Context, Scope, Page) (Result[[]AccessGrant], error)
	AccessGrant(context.Context, Scope, string, AccessRelationship) (Result[AccessGrant], error)
	RemoveAccessGrant(context.Context, Scope, string, AccessRelationship) error
}
