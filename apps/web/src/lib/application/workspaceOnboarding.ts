import { t } from '$lib/presentation/localization';
export type WorkspaceSetupMode = 'tenant_and_inventory' | 'inventory';

export interface WorkspaceSetupDraft {
  tenantName: string;
  inventoryName: string;
}

export interface WorkspaceSetupValidation {
  valid: boolean;
  tenantName: string;
  inventoryName: string;
  tenantError: string;
  inventoryError: string;
}

export function validateWorkspaceSetupDraft(mode: WorkspaceSetupMode, draft: WorkspaceSetupDraft): WorkspaceSetupValidation {
  const tenantName = draft.tenantName.trim();
  const inventoryName = draft.inventoryName.trim();
  const tenantError = mode === 'tenant_and_inventory' && !tenantName ? t('web.workspaceOnboarding.nameYourTenant') : '';
  const inventoryError = !inventoryName ? t('web.workspaceOnboarding.nameYourInventory') : '';
  return {
    valid: !tenantError && !inventoryError,
    tenantName,
    inventoryName,
    tenantError,
    inventoryError
  };
}

export function workspaceSetupTitle(mode: WorkspaceSetupMode): string {
  return mode === 'tenant_and_inventory' ? t('web.workspaceOnboarding.setUpYourWorkspace') : t('web.workspaceOnboarding.createAnInventory');
}

export function workspaceSetupDescription(mode: WorkspaceSetupMode, tenantName?: string): string {
  return mode === 'tenant_and_inventory'
    ? t('web.workspaceOnboarding.nameTheTenantAndFirstInventoryForThisStuff')
    : tenantName ? t('onboarding.firstInventoryNamed', { tenant: tenantName }) : t('onboarding.firstInventory');
}
