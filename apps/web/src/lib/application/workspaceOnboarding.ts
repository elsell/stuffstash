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
  const tenantError = mode === 'tenant_and_inventory' && !tenantName ? 'Name your tenant.' : '';
  const inventoryError = !inventoryName ? 'Name your inventory.' : '';
  return {
    valid: !tenantError && !inventoryError,
    tenantName,
    inventoryName,
    tenantError,
    inventoryError
  };
}

export function workspaceSetupTitle(mode: WorkspaceSetupMode): string {
  return mode === 'tenant_and_inventory' ? 'Set up your workspace' : 'Create an inventory';
}

export function workspaceSetupDescription(mode: WorkspaceSetupMode, tenantName?: string): string {
  return mode === 'tenant_and_inventory'
    ? 'Name the tenant and first inventory for this Stuff Stash instance.'
    : `Name the first inventory for ${tenantName ?? 'this tenant'}.`;
}
