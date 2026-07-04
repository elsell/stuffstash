import { describe, expect, it } from 'vitest';
import { workspaceAddAvailability } from './workspaceAddAvailability';

describe('workspaceAddAvailability', () => {
  it('allows add only when an inventory is selected and creation is permitted', () => {
    expect(workspaceAddAvailability({ hasInventory: true, canCreateAsset: true })).toEqual({
      canOpen: true,
      disabledReason: ''
    });
  });

  it('explains unavailable add when no inventory is selected', () => {
    expect(workspaceAddAvailability({ hasInventory: false, canCreateAsset: true })).toEqual({
      canOpen: false,
      disabledReason: 'Select an inventory before adding assets.'
    });
  });

  it('explains unavailable add when creation permission is missing', () => {
    expect(workspaceAddAvailability({ hasInventory: true, canCreateAsset: false })).toEqual({
      canOpen: false,
      disabledReason: 'Adding assets is unavailable for this inventory.'
    });
  });
});
