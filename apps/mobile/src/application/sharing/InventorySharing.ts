export type InventoryInvitationRelationship = 'viewer' | 'editor';
export type InventoryInvitationStatus = 'pending' | 'accepted' | 'revoked' | 'cancelled' | 'expired';

export type InventorySharingScope = {
  readonly tenantId: string;
  readonly inventoryId: string;
  readonly inventoryName: string;
  readonly permissions: readonly string[];
};

export type InventoryInvitationSummary = {
  readonly id: string;
  readonly email: string;
  readonly relationship: InventoryInvitationRelationship;
  readonly status: InventoryInvitationStatus;
  readonly isExpired: boolean;
  readonly expiresAt: string;
};

export type CreatedInventoryInvitation = InventoryInvitationSummary & {
  readonly inviteUrl: string;
};

export interface InventoryInvitationManagementRepository {
  list(scope: InventorySharingScope): Promise<readonly InventoryInvitationSummary[]>;
  create(
    scope: InventorySharingScope,
    input: { readonly email: string; readonly relationship: InventoryInvitationRelationship }
  ): Promise<CreatedInventoryInvitation>;
  cancel(scope: InventorySharingScope, invitationId: string): Promise<void>;
}

export interface InvitationLinkActions {
  copy(link: string): Promise<void>;
  share(input: { readonly link: string; readonly inventoryName: string }): Promise<void>;
}

export class InventorySharingPermissionError extends Error {
  constructor() {
    super('You do not have permission to manage invitations for this inventory.');
    this.name = 'InventorySharingPermissionError';
  }
}

export class ListInventoryInvitationsQuery {
  constructor(private readonly invitations: InventoryInvitationManagementRepository) {}

  async execute(scope: InventorySharingScope): Promise<readonly InventoryInvitationSummary[]> {
    requireShare(scope);
    return await this.invitations.list(scope);
  }
}

export class CreateInventoryInvitationCommand {
  constructor(private readonly invitations: InventoryInvitationManagementRepository) {}

  async execute(
    scope: InventorySharingScope,
    input: { readonly email: string; readonly relationship: InventoryInvitationRelationship }
  ): Promise<CreatedInventoryInvitation> {
    requireShare(scope);
    const email = input.email.trim();
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      throw new Error('Enter a valid email address.');
    }
    return await this.invitations.create(scope, { email, relationship: input.relationship });
  }
}

export class CancelInventoryInvitationCommand {
  constructor(private readonly invitations: InventoryInvitationManagementRepository) {}

  async execute(scope: InventorySharingScope, invitationId: string): Promise<void> {
    requireShare(scope);
    const id = invitationId.trim();
    if (id.length === 0) {
      throw new Error('Invitation ID must not be empty.');
    }
    await this.invitations.cancel(scope, id);
  }
}

function requireShare(scope: InventorySharingScope): void {
  if (!scope.permissions.includes('share')) {
    throw new InventorySharingPermissionError();
  }
}
