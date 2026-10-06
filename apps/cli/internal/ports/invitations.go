package ports

import "context"

type Invitation struct {
	ID                  string             `json:"id"`
	TenantID            string             `json:"tenantId"`
	InventoryID         string             `json:"inventoryId"`
	Email               string             `json:"email"`
	Relationship        AccessRelationship `json:"relationship"`
	Status              string             `json:"status"`
	ExpiresAt           string             `json:"expiresAt"`
	IsExpired           bool               `json:"isExpired"`
	InviterPrincipalID  string             `json:"inviterPrincipalId"`
	AcceptedPrincipalID *string            `json:"acceptedPrincipalId,omitempty"`
}
type InvitationAction string

const (
	CancelInvitation InvitationAction = "cancel"
	DeleteInvitation InvitationAction = "delete"
)

type InvitationsAPI interface {
	PreviewInvitation(context.Context, Scope, string, []byte) (Result[InvitationPreview], error)
	AcceptInvitation(context.Context, Scope, string, []byte) (Result[InvitationAcceptance], error)
	UpdateInvitationExpiration(context.Context, Scope, string, []byte) (Result[Invitation], error)
	Invitations(context.Context, Scope, Page, string) (Result[[]Invitation], error)
	Invitation(context.Context, Scope, string) (Result[Invitation], error)
	ChangeInvitation(context.Context, Scope, string, InvitationAction) error
}

type InvitationPreview struct {
	InventoryID   string             `json:"inventoryId"`
	InventoryName string             `json:"inventoryName"`
	Relationship  AccessRelationship `json:"relationship"`
	Status        string             `json:"status"`
	ExpiresAt     string             `json:"expiresAt"`
	IsExpired     bool               `json:"isExpired"`
}
type InvitationAcceptance struct {
	Invitation Invitation  `json:"invitation"`
	Grant      AccessGrant `json:"grant"`
}
