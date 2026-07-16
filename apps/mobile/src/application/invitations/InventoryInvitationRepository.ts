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
    super('Sign in to view this invitation.');
    this.name = 'InventoryInvitationAuthenticationRequiredError';
  }
}

export class InventoryInvitationEmailMismatchError extends Error {
  constructor() {
    super('This invitation belongs to another signed-in account.');
    this.name = 'InventoryInvitationEmailMismatchError';
  }
}

export class InventoryInvitationInvalidError extends Error {
  constructor() {
    super('This invitation link is invalid.');
    this.name = 'InventoryInvitationInvalidError';
  }
}

export class InventoryInvitationInvalidResponseError extends Error {
  constructor() {
    super('Stuff Stash returned an invalid invitation response.');
    this.name = 'InventoryInvitationInvalidResponseError';
  }
}
