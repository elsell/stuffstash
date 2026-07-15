export type InvitationRelationship = 'viewer' | 'editor';
export type InvitationStatus = 'pending' | 'accepted' | 'revoked' | 'cancelled' | 'expired';

export interface InvitationLinkMaterial {
  tenantId: string;
  inventoryId: string;
  invitationId: string;
  token: string;
}

export interface InvitationPreview {
  inventoryId: string;
  inventoryName: string;
  relationship: InvitationRelationship;
  status: InvitationStatus;
  isExpired: boolean;
  expiresAt: string;
}

export interface InvitationAcceptance {
  tenantId: string;
  inventoryId: string;
  status: 'accepted';
}

export type InvitationFailureKind = 'invalid' | 'email_mismatch' | 'authentication_required' | 'unavailable';

export class InvitationFailure extends Error {
  constructor(readonly kind: InvitationFailureKind) {
    super(kind);
    this.name = 'InvitationFailure';
  }
}
