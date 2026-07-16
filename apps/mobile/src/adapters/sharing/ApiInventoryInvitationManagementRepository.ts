import type { InventoryAccessInvitation, StuffStashClient } from '@stuff-stash/api-client';
import type {
  CreatedInventoryInvitation,
  InventoryInvitationManagementRepository,
  InventoryInvitationSummary,
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

  async list(scope: InventorySharingScope): Promise<readonly InventoryInvitationSummary[]> {
    const invitations: InventoryInvitationSummary[] = [];
    const seenCursors = new Set<string>();
    let cursor: string | undefined;
    for (let pageNumber = 0; pageNumber < 100; pageNumber += 1) {
      const page = await this.client.listInventoryAccessInvitations(
        scope.tenantId,
        scope.inventoryId,
        { limit: 50, cursor, status: 'all' }
      );
      invitations.push(...page.items.map((invitation) => mapSafeInvitation(invitation, scope)));
      cursor = page.pagination.nextCursor ?? undefined;
      if (!cursor) return invitations;
      if (seenCursors.has(cursor)) {
        throw new Error('Stuff Stash returned an invalid invitation page.');
      }
      seenCursors.add(cursor);
    }
    throw new Error('Stuff Stash returned too many invitation pages.');
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
