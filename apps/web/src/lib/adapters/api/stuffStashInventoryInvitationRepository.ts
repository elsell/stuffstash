import { StuffStashAPIError, StuffStashClient, type TokenProvider } from '@stuff-stash/api-client';
import { InvitationFailure, type InvitationAcceptance, type InvitationLinkMaterial, type InvitationPreview } from '$lib/domain/invitation';
import type { InventoryInvitationRepository } from '$lib/ports/inventoryInvitationRepository';

export class StuffStashInventoryInvitationRepository implements InventoryInvitationRepository {
  private readonly client: StuffStashClient;

  constructor(apiBaseUrl: string, tokenProvider: TokenProvider, fetchImpl?: typeof fetch) {
    this.client = new StuffStashClient({ baseUrl: apiBaseUrl, tokenProvider, fetch: fetchImpl });
  }

  async preview(material: InvitationLinkMaterial): Promise<InvitationPreview> {
    try {
      const preview = await this.client.previewInventoryAccessInvitation(
        material.tenantId,
        material.inventoryId,
        material.invitationId,
        material.token
      );
      if (preview.inventoryId !== material.inventoryId) throw new InvitationFailure('invalid');
      return preview;
    } catch (error) {
      throw mapInvitationFailure(error);
    }
  }

  async accept(material: InvitationLinkMaterial): Promise<InvitationAcceptance> {
    try {
      const accepted = await this.client.acceptInventoryAccessInvitation(
        material.tenantId,
        material.inventoryId,
        material.invitationId,
        material.token
      );
      if (
        accepted.invitation.id !== material.invitationId ||
        accepted.invitation.tenantId !== material.tenantId ||
        accepted.invitation.inventoryId !== material.inventoryId ||
        accepted.invitation.status !== 'accepted' ||
        accepted.grant.tenantId !== material.tenantId ||
        accepted.grant.inventoryId !== material.inventoryId ||
        accepted.grant.relationship !== accepted.invitation.relationship
      ) {
        throw new InvitationFailure('invalid');
      }
      return {
        tenantId: accepted.invitation.tenantId,
        inventoryId: accepted.invitation.inventoryId,
        status: 'accepted'
      };
    } catch (error) {
      throw mapInvitationFailure(error);
    }
  }
}

function mapInvitationFailure(error: unknown): InvitationFailure {
  if (error instanceof InvitationFailure) return error;
  if (error instanceof StuffStashAPIError) {
    if (error.status === 401 || error.code === 'authentication_required') return new InvitationFailure('authentication_required');
    if (error.code === 'invitation_email_mismatch') return new InvitationFailure('email_mismatch');
    if (error.code === 'invitation_invalid') return new InvitationFailure('invalid');
  }
  return new InvitationFailure('unavailable');
}
