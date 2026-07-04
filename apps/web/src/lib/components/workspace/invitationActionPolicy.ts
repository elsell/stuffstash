import type { AccessInvitationRouteAction } from '$lib/application/workspaceRoute';
import type { InventoryAccessInvitation } from '$lib/domain/inventory';

export function invitationActionIsAvailable(
  action: AccessInvitationRouteAction,
  invitation: InventoryAccessInvitation
): boolean {
  return action === 'delete' || ((action === 'expire' || action === 'cancel') && invitation.status === 'pending' && !invitation.isExpired);
}
