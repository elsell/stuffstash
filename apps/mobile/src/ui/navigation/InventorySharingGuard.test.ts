import { describe, expect, it } from 'vitest';
import type { SettingsLoadState } from '../screens/SettingsScreenState';
import { decideInventorySharingAccess } from './InventorySharingAccess';

// @ts-expect-error Vitest provides raw route sources to structural boundary tests.
const sharingRouteSources = import.meta.glob('../../app/settings/sharing.tsx', {
  eager: true,
  import: 'default',
  query: '?raw'
}) as Record<string, string>;

describe('InventorySharingGuard', () => {
  it('allows only the selected inventory share permission', () => {
    expect(decideInventorySharingAccess(state(['view', 'share']))).toEqual({
      status: 'allowed',
      scope: {
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        inventoryName: 'Household',
        permissions: ['view', 'share']
      }
    });
    expect(decideInventorySharingAccess(state(['view']))).toEqual({
      status: 'unavailable', inventoryName: 'Household'
    });
  });

  it('does not turn load failures into authorization denial', () => {
    expect(decideInventorySharingAccess({ status: 'loading' })).toEqual({ status: 'loading' });
    expect(decideInventorySharingAccess({ status: 'error', message: 'Scope unavailable.' }))
      .toEqual({ status: 'error', message: 'Scope unavailable.' });
  });

  it('guards the direct Sharing route before mounting invitation API behavior', () => {
    const source = Object.values(sharingRouteSources)[0] ?? '';
    expect(source).toContain('<InventorySharingGuard');
    expect(source.indexOf('<InventorySharingGuard')).toBeLessThan(source.indexOf('<InventorySharingScreen'));
  });
});

function state(permissions: readonly string[]): SettingsLoadState {
  return {
    status: 'ready',
    settings: {
      principal: { id: 'owner-one', primaryLabel: 'owner@example.com' },
      selectedTenant: { id: 'tenant-home', name: 'Home', permissions: ['view'] },
      selectedInventory: { id: 'inventory-home', name: 'Household', permissions },
      serverUrl: 'https://stash.example', appVersion: '0.0.0', authenticationMode: 'oidc-sso'
    }
  };
}
