import { describe, expect, it } from 'vitest';
import type { WorkspaceRouteState } from './workspaceRoute';
import { defaultWorkspaceRoute } from './workspaceRoute';
import { resolveWorkspaceAddRoute } from './workspaceAddRoute';

describe('workspace add route helper', () => {
  it('keeps ordinary routes closed with default add state', () => {
    expect(resolveWorkspaceAddRoute(route(), addInput())).toEqual({
      open: false,
      kind: 'item',
      parentAssetId: null,
      deniedMessage: '',
      replacementRoute: null
    });
  });

  it('opens permitted add routes with a valid route-backed parent', () => {
    expect(
      resolveWorkspaceAddRoute(
        route({ action: 'add', addKind: 'container', addParentAssetId: 'location-garage' }),
        addInput({ validParentIds: ['location-garage'] })
      )
    ).toEqual({
      open: true,
      kind: 'container',
      parentAssetId: 'location-garage',
      deniedMessage: '',
      replacementRoute: null
    });
  });

  it('returns a denied add state when creation is unavailable', () => {
    expect(resolveWorkspaceAddRoute(route({ action: 'add', addKind: 'location' }), addInput({ createAllowed: false }))).toEqual({
      open: false,
      kind: 'location',
      parentAssetId: null,
      deniedMessage: 'You do not have permission to add assets in this inventory.',
      replacementRoute: null
    });
  });

  it('normalizes invalid route-backed parents without changing the requested add kind', () => {
    expect(
      resolveWorkspaceAddRoute(
        route({ action: 'add', addKind: 'item', addParentAssetId: 'missing-location' }),
        addInput({ validParentIds: ['location-garage'] })
      )
    ).toEqual({
      open: true,
      kind: 'item',
      parentAssetId: null,
      deniedMessage: '',
      replacementRoute: {
        action: 'add',
        addKind: 'item',
        addParentAssetId: null,
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household'
      }
    });
  });
});

function route(overrides: Partial<WorkspaceRouteState> = {}): WorkspaceRouteState {
  return {
    ...defaultWorkspaceRoute,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    ...overrides
  };
}

function addInput(
  overrides: Partial<Parameters<typeof resolveWorkspaceAddRoute>[1]> = {}
): Parameters<typeof resolveWorkspaceAddRoute>[1] {
  return {
    createAllowed: true,
    validParentIds: [],
    selectedTenantId: 'tenant-home',
    selectedInventoryId: 'inventory-household',
    ...overrides
  };
}
