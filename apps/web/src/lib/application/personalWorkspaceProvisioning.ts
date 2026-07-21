import type { Principal } from '$lib/domain/inventory';
import type { WorkspaceData } from '$lib/domain/inventory';
import type { InventoryRepository } from '$lib/ports/inventoryRepository';

interface ProvisioningLock {
  request<T>(name: string, callback: () => Promise<T>): Promise<T>;
}

export function personalWorkspaceNames(principal: Principal): { tenantName: string; inventoryName: string } {
  const displayName = principal.displayName?.trim();
  return {
    tenantName: displayName ? `${displayName}\u2019s household` : 'My household',
    inventoryName: 'Home'
  };
}

export function provisionPersonalWorkspace(
  repository: InventoryRepository,
  principal: Principal,
  lock: ProvisioningLock | undefined = typeof navigator === 'undefined' ? undefined : navigator.locks
): Promise<WorkspaceData> {
  const provision = () => repository.provisionPersonalWorkspace(personalWorkspaceNames(principal));
  return lock ? lock.request('stuffstash.personal-workspace-provisioning', provision) : provision();
}
