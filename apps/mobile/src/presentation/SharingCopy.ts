import type { createTranslator, en } from '@stuff-stash/localization';
import type { InventoryInvitationSummary, InventoryInvitationStatus, InventoryInvitationRelationship } from '../application/sharing/InventorySharing';

type Translator = ReturnType<typeof createTranslator<typeof en>>;
type Invitation = Pick<InventoryInvitationSummary, 'email' | 'relationship' | 'status' | 'isExpired' | 'expiresAt'>;
const roles = {
  viewer: 'mobile.InventorySharingScreen.viewer',
  editor: 'mobile.InventorySharingScreen.editor',
} as const satisfies Record<InventoryInvitationRelationship, keyof typeof en>;
const statuses = {
  pending: 'sharing.status.pending',
  accepted: 'sharing.status.accepted',
  revoked: 'sharing.status.revoked',
  cancelled: 'sharing.status.cancelled',
  expired: 'mobile.InventorySharingScreen.expired',
} as const satisfies Record<InventoryInvitationStatus, keyof typeof en>;

export function sharingCopy(translator: Translator) {
  function values(invitation: Invitation) {
    const date = new Date(invitation.expiresAt);
    return {
      email: invitation.email,
      access: translator.message(roles[invitation.relationship]),
      status: translator.message(statuses[invitation.isExpired ? 'expired' : invitation.status]),
      date: Number.isNaN(date.getTime()) ? invitation.expiresAt : translator.date(date, { dateStyle: 'medium' }),
    };
  }
  return {
    heading: (inventoryName: string) => translator.message('sharing.heading', { inventoryName }),
    createdMetadata: (invitation: Invitation) => translator.message('sharing.createdMetadata', values(invitation)),
    metadata: (invitation: Invitation) => translator.message('sharing.invitationMetadata', values(invitation)),
  };
}
