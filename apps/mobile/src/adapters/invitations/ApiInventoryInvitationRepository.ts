import { StuffStashAPIError, type StuffStashClient } from '@stuff-stash/api-client';
import type {
  InventoryInvitationAcceptance,
  InventoryInvitationPreview,
  InventoryInvitationReference,
  InventoryInvitationRepository
} from '../../application/invitations/InventoryInvitationRepository';
import {
  InventoryInvitationAuthenticationRequiredError,
  InventoryInvitationEmailMismatchError,
  InventoryInvitationInvalidError,
  InventoryInvitationInvalidResponseError
} from '../../application/invitations/InventoryInvitationRepository';

type InventoryInvitationApiClient = Pick<
  StuffStashClient,
  'previewInventoryAccessInvitation' | 'acceptInventoryAccessInvitation'
>;

export class ApiInventoryInvitationRepository implements InventoryInvitationRepository {
  constructor(private readonly client: InventoryInvitationApiClient) {}

  async preview(input: InventoryInvitationReference): Promise<InventoryInvitationPreview> {
    const preview = await this.mapInvitationError(() =>
      this.client.previewInventoryAccessInvitation(
        input.tenantId,
        input.inventoryId,
        input.invitationId,
        input.acceptanceToken
      )
    );
    if (preview.inventoryId !== input.inventoryId) {
      throw new InventoryInvitationInvalidResponseError();
    }
    return {
      inventoryId: preview.inventoryId,
      inventoryName: preview.inventoryName,
      relationship: preview.relationship,
      status: preview.status,
      isExpired: preview.isExpired,
      expiresAt: preview.expiresAt
    };
  }

  async accept(input: InventoryInvitationReference): Promise<InventoryInvitationAcceptance> {
    const acceptance = await this.mapInvitationError(() =>
      this.client.acceptInventoryAccessInvitation(
        input.tenantId,
        input.inventoryId,
        input.invitationId,
        input.acceptanceToken
      )
    );
    if (
      acceptance.grant.tenantId !== input.tenantId ||
      acceptance.grant.inventoryId !== input.inventoryId ||
      acceptance.invitation.id !== input.invitationId ||
      acceptance.invitation.tenantId !== input.tenantId ||
      acceptance.invitation.inventoryId !== input.inventoryId ||
      acceptance.invitation.relationship !== acceptance.grant.relationship ||
      acceptance.invitation.status !== 'accepted'
    ) {
      throw new InventoryInvitationInvalidResponseError();
    }
    return {
      tenantId: acceptance.grant.tenantId,
      inventoryId: acceptance.grant.inventoryId,
      invitationId: acceptance.invitation.id,
      principalId: acceptance.grant.principalId,
      relationship: acceptance.grant.relationship,
      status: 'accepted'
    };
  }

  private async mapInvitationError<Result>(operation: () => Promise<Result>): Promise<Result> {
    try {
      return await operation();
    } catch (error) {
      if (!(error instanceof StuffStashAPIError)) {
        throw error;
      }
      if (error.code === 'invitation_email_mismatch') {
        throw new InventoryInvitationEmailMismatchError();
      }
      if (error.code === 'invitation_invalid') {
        throw new InventoryInvitationInvalidError();
      }
      if (error.status === 401) {
        throw new InventoryInvitationAuthenticationRequiredError();
      }
      throw error;
    }
  }
}
