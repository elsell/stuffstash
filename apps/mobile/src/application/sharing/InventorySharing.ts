import type { ReadRequest } from '../shared/ReadRequest';
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

export type InventoryInvitationPage = { readonly items: readonly InventoryInvitationSummary[]; readonly nextCursor?: string };
export type InventoryInvitationRead = ReadRequest & { readonly cursor?: string };
export interface InventoryInvitationMutationObserver {
  onInvitationsChanged(scope: InventorySharingScope, cancelledInvitationId?: string): void;
}
const noInvitationObserver: InventoryInvitationMutationObserver = { onInvitationsChanged: () => undefined };

export interface InventoryInvitationManagementRepository {
  list(scope: InventorySharingScope, request?: InventoryInvitationRead): Promise<InventoryInvitationPage>;
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

  async execute(scope: InventorySharingScope, request: InventoryInvitationRead = {}): Promise<InventoryInvitationPage> {
    requireShare(scope);
    return await this.invitations.list(scope, request);
  }
}

export class CreateInventoryInvitationCommand {
  constructor(private readonly invitations: InventoryInvitationManagementRepository, private readonly observer: InventoryInvitationMutationObserver = noInvitationObserver) {}

  async execute(
    scope: InventorySharingScope,
    input: { readonly email: string; readonly relationship: InventoryInvitationRelationship }
  ): Promise<CreatedInventoryInvitation> {
    requireShare(scope);
    const email = input.email.trim();
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      throw new Error('Enter a valid email address.');
    }
    const result = await this.invitations.create(scope, { email, relationship: input.relationship });
    this.observer.onInvitationsChanged(scope);
    return result;
  }
}

export class CancelInventoryInvitationCommand {
  constructor(private readonly invitations: InventoryInvitationManagementRepository, private readonly observer: InventoryInvitationMutationObserver = noInvitationObserver) {}

  async execute(scope: InventorySharingScope, invitationId: string): Promise<void> {
    requireShare(scope);
    const id = invitationId.trim();
    if (id.length === 0) {
      throw new Error('Invitation ID must not be empty.');
    }
    await this.invitations.cancel(scope, id);
    this.observer.onInvitationsChanged(scope, id);
  }
}

function requireShare(scope: InventorySharingScope): void {
  if (!scope.permissions.includes('share')) {
    throw new InventorySharingPermissionError();
  }
}
