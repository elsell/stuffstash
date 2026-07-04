export type WorkspaceRouteRecoveryPresentation = {
  title: string;
  message: string;
  actionLabel?: string;
  role?: 'alert';
};

export function workspaceUnavailableRoutePresentation(message: string): WorkspaceRouteRecoveryPresentation {
  return {
    title: 'Workspace unavailable',
    message,
    actionLabel: 'Go home',
    role: 'alert'
  };
}

export function workspaceNoInventoryPresentation(
  selectedTenantId: string | null,
  canCreateStarter: boolean
): WorkspaceRouteRecoveryPresentation {
  if (!canCreateStarter) {
    return {
      title: 'No inventory yet',
      message: 'You can view this tenant, but you cannot create inventories in it.'
    };
  }

  return {
    title: 'No inventory yet',
    message: selectedTenantId ? 'Create the first inventory for this tenant.' : 'Create your first tenant and inventory.',
    actionLabel: 'Create Household'
  };
}
