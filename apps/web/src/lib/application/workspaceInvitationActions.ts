import type { AccessInvitationRouteAction } from './workspaceRoute';
import { workspaceRouteHref } from './workspaceRoute';
import type { InventoryAccessInvitation, InvitationStatusFilter } from '$lib/domain/inventory';

export function invitationActionIsAvailable(
  action: AccessInvitationRouteAction,
  invitation: InventoryAccessInvitation
): boolean {
  return action === 'delete' || ((action === 'expire' || action === 'cancel') && invitation.status === 'pending' && !invitation.isExpired);
}

export function accessInvitationsHref(
  tenantId: string | null,
  inventoryId: string | null,
  invitationStatus: InvitationStatusFilter
): string {
  return workspaceRouteHref({ mode: 'settings', settingsSection: 'access', invitationStatus }, tenantId, inventoryId);
}

export function invitationActionHref(
  tenantId: string | null,
  inventoryId: string | null,
  invitationStatus: InvitationStatusFilter,
  invitation: InventoryAccessInvitation,
  action: Exclude<AccessInvitationRouteAction, null>
): string {
  return workspaceRouteHref(
    {
      mode: 'settings',
      settingsSection: 'access',
      invitationStatus,
      accessInvitationAction: action,
      accessInvitationId: invitation.id
    },
    tenantId,
    inventoryId
  );
}
