import { t } from '$lib/presentation/localization';
import type { AccessInvitationRouteAction } from './workspaceRoute';
import { workspaceRouteHref } from './workspaceRoute';
import type { InventoryAccessInvitation, InvitationStatusFilter } from '$lib/domain/inventory';

export interface InvitationActionOption {
  action: Exclude<AccessInvitationRouteAction, null>;
  label: string;
  ariaLabel?: string;
  href: string;
  disabled: boolean;
  destructive: boolean;
  iconOnly: boolean;
}

export interface InvitationActionConfirmation {
  title: string;
  description: string;
  buttonLabel: string;
  destructive: boolean;
  disabled: boolean;
}

type InvitationAction = Exclude<AccessInvitationRouteAction, null>;

interface InvitationActionMetadata {
  rowLabel: string;
  confirmationTitle: string;
  confirmationButtonLabel: string;
  description: (invitation: InventoryAccessInvitation) => string;
  destructive: boolean;
  iconOnly: boolean;
}

const invitationActionMetadata: Record<InvitationAction, InvitationActionMetadata> = {
  expire: {
    rowLabel: t('web.workspaceInvitationActions.expire'),
    confirmationTitle: t('web.workspaceInvitationActions.expireInvitation'),
    confirmationButtonLabel: t('web.workspaceInvitationActions.expire'),
    description: (invitation) => t('web.workspaceInvitationActions.expireDescription', { email: invitation.email }),
    destructive: false,
    iconOnly: false
  },
  cancel: {
    rowLabel: t('web.workspaceInvitationActions.cancel'),
    confirmationTitle: t('web.workspaceInvitationActions.cancelInvitation'),
    confirmationButtonLabel: t('web.workspaceInvitationActions.cancelInvitation'),
    description: (invitation) => t('web.workspaceInvitationActions.cancelDescription', { email: invitation.email }),
    destructive: false,
    iconOnly: false
  },
  delete: {
    rowLabel: t('web.workspaceInvitationActions.delete'),
    confirmationTitle: t('web.workspaceInvitationActions.deleteInvitation'),
    confirmationButtonLabel: t('web.workspaceInvitationActions.delete'),
    description: (invitation) => t('web.workspaceInvitationActions.deleteDescription', { email: invitation.email }),
    destructive: true,
    iconOnly: true
  }
};

const invitationActions = Object.keys(invitationActionMetadata) as InvitationAction[];

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
  return workspaceRouteHref({ mode: 'settings', settingsLevel: 'inventory', tenantId, inventoryId, settingsCollection: 'access', settingsSection: 'access', invitationStatus }, tenantId, inventoryId);
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
      settingsLevel: 'inventory',
      tenantId,
      inventoryId,
      settingsCollection: 'access',
      settingsSection: 'access',
      invitationStatus,
      accessInvitationAction: action,
      accessInvitationId: invitation.id
    },
    tenantId,
    inventoryId
  );
}

export function invitationActionOptions(input: {
  tenantId: string | null;
  inventoryId: string | null;
  invitationStatus: InvitationStatusFilter;
  invitation: InventoryAccessInvitation;
  busy: boolean;
}): InvitationActionOption[] {
  return invitationActions.map((action) => ({
    action,
    label: invitationActionMetadata[action].rowLabel,
    ariaLabel: action === 'delete' ? t('web.workspaceInvitationActions.deleteInvitationFor', { email: String(input.invitation.email) }) : undefined,
    href: invitationActionHref(input.tenantId, input.inventoryId, input.invitationStatus, input.invitation, action),
    disabled: input.busy || !invitationActionIsAvailable(action, input.invitation),
    destructive: invitationActionMetadata[action].destructive,
    iconOnly: invitationActionMetadata[action].iconOnly
  }));
}

export function invitationActionConfirmation(
  action: AccessInvitationRouteAction,
  invitation: InventoryAccessInvitation,
  busy: boolean
): InvitationActionConfirmation {
  if (!action) {
    return {
      title: t('web.workspaceInvitationActions.invitationAction'),
      description: t('web.workspaceInvitationActions.thisInvitationActionIsUnavailable'),
      buttonLabel: t('web.workspaceInvitationActions.continue'),
      destructive: false,
      disabled: true
    };
  }
  const metadata = invitationActionMetadata[action];
  return {
    title: metadata.confirmationTitle,
    description: metadata.description(invitation),
    buttonLabel: metadata.confirmationButtonLabel,
    destructive: metadata.destructive,
    disabled: busy || !invitationActionIsAvailable(action, invitation)
  };
}
