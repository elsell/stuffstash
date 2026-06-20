package dto

import (
	"time"

	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
)

type GrantInventoryAccessInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Body          GrantBody
}

type GrantBody struct {
	PrincipalID  string `json:"principalId" doc:"User principal ID to grant access to"`
	Relationship string `json:"relationship" enum:"viewer,editor" doc:"Direct inventory relationship"`
}

type GrantInventoryAccessOutput struct {
	Body shared.SuccessEnvelope[GrantResponse]
}

type ListInventoryAccessInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
}

type ListInventoryAccessOutput struct {
	Body shared.SuccessEnvelope[[]GrantResponse]
}

type GetInventoryAccessGrantInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	PrincipalID   string `path:"principalId" doc:"User principal ID"`
	Relationship  string `path:"relationship" enum:"viewer,editor" doc:"Direct inventory relationship"`
}

type GetInventoryAccessGrantOutput struct {
	Body shared.SuccessEnvelope[GrantResponse]
}

type RevokeInventoryAccessInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	PrincipalID   string `path:"principalId" doc:"User principal ID to revoke access from"`
	Relationship  string `path:"relationship" enum:"viewer,editor" doc:"Direct inventory relationship"`
}

type RevokeInventoryAccessOutput struct{}

type CreateInventoryAccessInvitationInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Body          InvitationBody
}

type InvitationBody struct {
	Email        string `json:"email" doc:"Invitee email address"`
	Relationship string `json:"relationship" enum:"viewer,editor" doc:"Direct inventory relationship to grant on acceptance"`
}

type CreateInventoryAccessInvitationOutput struct {
	Body shared.SuccessEnvelope[InvitationResponse]
}

type AcceptInventoryAccessInvitationInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	InvitationID  string `path:"invitationId" doc:"Invitation ID"`
	Body          AcceptInvitationBody
}

type AcceptInvitationBody struct {
	AcceptanceToken string `json:"acceptanceToken" doc:"One-time invite acceptance token"`
}

type AcceptInventoryAccessInvitationOutput struct {
	Body shared.SuccessEnvelope[InvitationAcceptanceResponse]
}

type ListInventoryAccessInvitationsInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	Limit         int    `query:"limit" minimum:"1" doc:"Requested page size"`
	Cursor        string `query:"cursor" doc:"Opaque cursor from the previous page"`
	Status        string `query:"status" enum:"all,pending,accepted,revoked,cancelled,expired" doc:"Invitation status filter"`
}

type ListInventoryAccessInvitationsOutput struct {
	Body shared.SuccessEnvelope[[]InvitationResponse]
}

type RevokeInventoryAccessInvitationInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	InvitationID  string `path:"invitationId" doc:"Invitation ID"`
}

type UpdateInventoryAccessInvitationExpirationInput struct {
	Authorization string `header:"Authorization" doc:"Bearer dev:<principal-id>"`
	RequestID     string `header:"X-Request-ID" doc:"Optional request correlation ID"`
	TenantID      string `path:"tenantId" doc:"Tenant ID"`
	InventoryID   string `path:"inventoryId" doc:"Inventory ID"`
	InvitationID  string `path:"invitationId" doc:"Invitation ID"`
	Body          UpdateInvitationExpirationBody
}

type UpdateInvitationExpirationBody struct {
	ExpiresAt time.Time `json:"expiresAt" doc:"New invitation expiration timestamp"`
}

type GetInventoryAccessInvitationOutput struct {
	Body shared.SuccessEnvelope[InvitationResponse]
}

type UpdateInventoryAccessInvitationExpirationOutput struct {
	Body shared.SuccessEnvelope[InvitationResponse]
}

type RevokeInventoryAccessInvitationOutput struct{}

type GrantResponse struct {
	TenantID     string `json:"tenantId"`
	InventoryID  string `json:"inventoryId"`
	PrincipalID  string `json:"principalId"`
	Relationship string `json:"relationship"`
}

type InvitationResponse struct {
	ID                  string `json:"id"`
	TenantID            string `json:"tenantId"`
	InventoryID         string `json:"inventoryId"`
	Email               string `json:"email"`
	Relationship        string `json:"relationship"`
	Status              string `json:"status"`
	InviterPrincipalID  string `json:"inviterPrincipalId"`
	AcceptedPrincipalID string `json:"acceptedPrincipalId,omitempty"`
	ExpiresAt           string `json:"expiresAt"`
	IsExpired           bool   `json:"isExpired"`
	AcceptanceToken     string `json:"acceptanceToken,omitempty"`
}

type InvitationAcceptanceResponse struct {
	Invitation InvitationResponse `json:"invitation"`
	Grant      GrantResponse      `json:"grant"`
}
