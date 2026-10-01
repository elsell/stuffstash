import { t } from '../../presentation/localization';
export type InventoryInvitationRelationship = 'viewer' | 'editor';

export type InventoryInvitationStatus =
  | 'pending'
  | 'accepted'
  | 'revoked'
  | 'cancelled'
  | 'expired';

export type InventoryInvitationReference = {
  readonly tenantId: string;
  readonly inventoryId: string;
  readonly invitationId: string;
  readonly acceptanceToken: string;
};

export type InventoryInvitationPreview = {
  readonly inventoryId: string;
  readonly inventoryName: string;
  readonly relationship: InventoryInvitationRelationship;
  readonly status: InventoryInvitationStatus;
  readonly isExpired: boolean;
  readonly expiresAt: string;
};

export type InventoryInvitationAcceptance = {
  readonly tenantId: string;
  readonly inventoryId: string;
  readonly invitationId: string;
  readonly principalId: string;
  readonly relationship: InventoryInvitationRelationship;
  readonly status: 'accepted';
};

export interface InventoryInvitationRepository {
  preview(input: InventoryInvitationReference): Promise<InventoryInvitationPreview>;
  accept(input: InventoryInvitationReference): Promise<InventoryInvitationAcceptance>;
}

export class InventoryInvitationAuthenticationRequiredError extends Error {
  constructor() {
    super(t('mobile.InventoryInvitationRepository.signInToViewThisInvitation'));
    this.name = 'InventoryInvitationAuthenticationRequiredError';
  }
}

export class InventoryInvitationEmailMismatchError extends Error {
  constructor() {
    super(t('mobile.InventoryInvitationRepository.thisInvitationBelongsToAnotherSignedInAccount'));
    this.name = 'InventoryInvitationEmailMismatchError';
  }
}

export class InventoryInvitationInvalidError extends Error {
  constructor() {
    super(t('mobile.InventoryInvitationRepository.thisInvitationLinkIsInvalid'));
    this.name = 'InventoryInvitationInvalidError';
  }
}

export class InventoryInvitationInvalidResponseError extends Error {
  constructor() {
    super(t('mobile.InventoryInvitationRepository.stuffStashReturnedAnInvalidInvitationResponse'));
    this.name = 'InventoryInvitationInvalidResponseError';
  }
}
