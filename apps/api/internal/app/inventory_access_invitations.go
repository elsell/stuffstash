package app

import (
	"context"

	inventoryapp "github.com/stuffstash/stuff-stash/internal/app/inventories"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type CreateInventoryAccessInvitationInput = inventoryapp.CreateInventoryAccessInvitationInput
type CreateInventoryAccessInvitationResult = inventoryapp.CreateInventoryAccessInvitationResult
type AcceptInventoryAccessInvitationInput = inventoryapp.AcceptInventoryAccessInvitationInput
type PreviewInventoryAccessInvitationInput = inventoryapp.PreviewInventoryAccessInvitationInput
type InventoryAccessInvitationPreview = inventoryapp.InventoryAccessInvitationPreview
type RevokeInventoryAccessInvitationInput = inventoryapp.RevokeInventoryAccessInvitationInput
type GetInventoryAccessInvitationInput = inventoryapp.GetInventoryAccessInvitationInput
type ListInventoryAccessInvitationsInput = inventoryapp.ListInventoryAccessInvitationsInput
type ListInventoryAccessInvitationsResult = inventoryapp.ListInventoryAccessInvitationsResult
type UpdateInventoryAccessInvitationExpirationInput = inventoryapp.UpdateInventoryAccessInvitationExpirationInput

func (a App) CreateInventoryAccessInvitation(ctx context.Context, input CreateInventoryAccessInvitationInput) (CreateInventoryAccessInvitationResult, error) {
	return a.inventoryService().CreateInventoryAccessInvitation(ctx, input)
}

func normalizeInvitationPublicBaseURL(value string) string {
	return inventoryapp.NormalizeInvitationPublicBaseURL(value)
}

func buildInventoryInvitationURL(baseURL string, invitation ports.InventoryAccessInvitation, acceptanceToken string, allowInsecureLocalHTTP bool) (string, error) {
	return inventoryapp.BuildInventoryInvitationURL(baseURL, invitation, acceptanceToken, allowInsecureLocalHTTP)
}

func (a App) PreviewInventoryAccessInvitation(ctx context.Context, input PreviewInventoryAccessInvitationInput) (InventoryAccessInvitationPreview, error) {
	return a.inventoryService().PreviewInventoryAccessInvitation(ctx, input)
}

func (a App) AcceptInventoryAccessInvitation(ctx context.Context, input AcceptInventoryAccessInvitationInput) (ports.InventoryAccessInvitation, ports.InventoryAccessGrant, error) {
	return a.inventoryService().AcceptInventoryAccessInvitation(ctx, input)
}

func (a App) RevokeInventoryAccessInvitation(ctx context.Context, input RevokeInventoryAccessInvitationInput) (bool, error) {
	return a.inventoryService().RevokeInventoryAccessInvitation(ctx, input)
}

func (a App) GetInventoryAccessInvitation(ctx context.Context, input GetInventoryAccessInvitationInput) (ports.InventoryAccessInvitation, error) {
	return a.inventoryService().GetInventoryAccessInvitation(ctx, input)
}

func (a App) ListInventoryAccessInvitations(ctx context.Context, input ListInventoryAccessInvitationsInput) (ListInventoryAccessInvitationsResult, error) {
	return a.inventoryService().ListInventoryAccessInvitations(ctx, input)
}

func (a App) UpdateInventoryAccessInvitationExpiration(ctx context.Context, input UpdateInventoryAccessInvitationExpirationInput) (ports.InventoryAccessInvitation, error) {
	return a.inventoryService().UpdateInventoryAccessInvitationExpiration(ctx, input)
}

func (a App) CancelInventoryAccessInvitation(ctx context.Context, input RevokeInventoryAccessInvitationInput) (bool, error) {
	return a.inventoryService().CancelInventoryAccessInvitation(ctx, input)
}

func (a App) DeleteInventoryAccessInvitation(ctx context.Context, input RevokeInventoryAccessInvitationInput) (bool, error) {
	return a.inventoryService().DeleteInventoryAccessInvitation(ctx, input)
}

func hashInventoryInvitationToken(token string) string {
	return inventoryapp.HashInventoryInvitationToken(token)
}
