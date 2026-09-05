import type { InventoryAccessInvitation, StuffStashClient } from '@stuff-stash/api-client';
import type {
  CreatedInventoryInvitation,
  InventoryInvitationManagementRepository,
  InventoryInvitationSummary,
  InventoryInvitationRead,
  InventoryInvitationPage,
  InventorySharingScope
} from '../../application/sharing/InventorySharing';
import { parseCreatedInventoryInvitationLink } from '../../application/invitations/InvitationLinkParser';

type InvitationManagementClient = Pick<
  StuffStashClient,
  'listInventoryAccessInvitations' | 'createInventoryAccessInvitation' | 'cancelInventoryAccessInvitation'
>;

export class ApiInventoryInvitationManagementRepository implements InventoryInvitationManagementRepository {
  constructor(
    private readonly client: InvitationManagementClient,
    private readonly trustedInvitationOrigin?: string,
    private readonly allowInsecureLocalHTTP = false
  ) {}

  async list(scope: InventorySharingScope, request: InventoryInvitationRead = {}): Promise<InventoryInvitationPage> {
    request.signal?.throwIfAborted();
    const page = await this.client.listInventoryAccessInvitations(
      scope.tenantId, scope.inventoryId, { limit: 50, cursor: request.cursor, status: 'all' }, request.signal
    );
    request.signal?.throwIfAborted();
    const nextCursor = page.pagination.nextCursor ?? undefined;
    if (nextCursor && nextCursor === request.cursor) throw new Error('Stuff Stash returned an invalid invitation page.');
    return { items: page.items.map((invitation) => mapSafeInvitation(invitation, scope)), nextCursor };
  }

  async create(
    scope: InventorySharingScope,
    input: { readonly email: string; readonly relationship: 'viewer' | 'editor' }
  ): Promise<CreatedInventoryInvitation> {
    const invitation = await this.client.createInventoryAccessInvitation(
      scope.tenantId,
      scope.inventoryId,
      input
    );
    if (!invitation.inviteUrl) {
      throw new Error('Stuff Stash did not return the one-time invitation link.');
    }
    let reference;
    try {
      reference = parseCreatedInventoryInvitationLink(
        invitation.inviteUrl,
        this.trustedInvitationOrigin,
        this.allowInsecureLocalHTTP
      );
    } catch {
      throw new Error('Stuff Stash did not return the one-time invitation link.');
    }
    if (
      invitation.tenantId !== scope.tenantId ||
      invitation.inventoryId !== scope.inventoryId ||
      reference.tenantId !== scope.tenantId ||
      reference.inventoryId !== scope.inventoryId ||
      reference.invitationId !== invitation.id
    ) {
      throw new Error('Stuff Stash did not return the one-time invitation link.');
    }
    return { ...mapSafeInvitation(invitation, scope), inviteUrl: invitation.inviteUrl };
  }

  async cancel(scope: InventorySharingScope, invitationId: string): Promise<void> {
    await this.client.cancelInventoryAccessInvitation(scope.tenantId, scope.inventoryId, invitationId);
  }
}

function mapSafeInvitation(
  invitation: InventoryAccessInvitation,
  scope: InventorySharingScope
): InventoryInvitationSummary {
  if (invitation.tenantId !== scope.tenantId || invitation.inventoryId !== scope.inventoryId) {
    throw new Error('Stuff Stash returned an invalid invitation response.');
  }
  return {
    id: invitation.id,
    email: invitation.email,
    relationship: invitation.relationship,
    status: invitation.status,
    isExpired: invitation.isExpired,
    expiresAt: invitation.expiresAt
  };
}
