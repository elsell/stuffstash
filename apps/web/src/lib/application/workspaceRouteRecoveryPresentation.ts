import { t } from '$lib/presentation/localization';
export type WorkspaceRouteRecoveryPresentation = {
  title: string;
  message: string;
  actionLabel?: string;
  role?: 'alert';
};

export function workspaceUnavailableRoutePresentation(message: string): WorkspaceRouteRecoveryPresentation {
  return {
    title: t('web.workspaceRouteRecoveryPresentation.workspaceUnavailable'),
    message,
    actionLabel: t('web.workspaceRouteRecoveryPresentation.goHome'),
    role: 'alert'
  };
}

export function workspaceNoInventoryPresentation(
  selectedTenantId: string | null,
  canCreateStarter: boolean
): WorkspaceRouteRecoveryPresentation {
  if (!canCreateStarter) {
    return {
      title: t('web.workspaceRouteRecoveryPresentation.noInventoryYet'),
      message: t('web.workspaceRouteRecoveryPresentation.youCanViewThisTenantButYouCannotCreate')
    };
  }

  return {
    title: t('web.workspaceRouteRecoveryPresentation.noInventoryYet'),
    message: selectedTenantId ? t('web.workspaceRouteRecoveryPresentation.createTheFirstInventoryForThisTenant') : t('web.workspaceRouteRecoveryPresentation.createYourFirstTenantAndInventory')
  };
}
