import type { InventorySharingScope } from '../../application/sharing/InventorySharing';
import type { SettingsLoadState } from '../screens/SettingsScreenState';

export type InventorySharingAccessDecision =
  | { readonly status: 'loading' }
  | { readonly status: 'allowed'; readonly scope: InventorySharingScope }
  | { readonly status: 'unavailable'; readonly inventoryName: string }
  | { readonly status: 'error'; readonly message: string };

export function decideInventorySharingAccess(state: SettingsLoadState): InventorySharingAccessDecision {
  if (state.status !== 'ready') return state;
  const inventory = state.settings.selectedInventory;
  if (!inventory.permissions.includes('share')) {
    return { status: 'unavailable', inventoryName: inventory.name };
  }
  return {
    status: 'allowed',
    scope: {
      tenantId: state.settings.selectedTenant.id,
      inventoryId: inventory.id,
      inventoryName: inventory.name,
      permissions: inventory.permissions
    }
  };
}
