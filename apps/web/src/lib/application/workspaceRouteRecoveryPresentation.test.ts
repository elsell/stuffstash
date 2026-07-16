import { describe, expect, it } from 'vitest';
import { workspaceNoInventoryPresentation, workspaceUnavailableRoutePresentation } from './workspaceRouteRecoveryPresentation';

describe('workspace route recovery presentation', () => {
  it('derives unavailable-route recovery copy and alert semantics', () => {
    expect(workspaceUnavailableRoutePresentation('That inventory is gone.')).toEqual({
      title: 'Workspace unavailable',
      message: 'That inventory is gone.',
      actionLabel: 'Go home',
      role: 'alert'
    });
  });

  it('derives no-inventory setup and denied presentation', () => {
    expect(workspaceNoInventoryPresentation('tenant-home', true)).toEqual({
      title: 'No inventory yet',
      message: 'Create the first inventory for this tenant.'
    });
    expect(workspaceNoInventoryPresentation(null, true).message).toBe('Create your first tenant and inventory.');
    expect(workspaceNoInventoryPresentation('tenant-home', false)).toEqual({
      title: 'No inventory yet',
      message: 'You can view this tenant, but you cannot create inventories in it.'
    });
  });
});
