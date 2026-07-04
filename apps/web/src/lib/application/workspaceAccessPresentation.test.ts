import { describe, expect, it } from 'vitest';
import { inventoryAccessRelationshipLabel, inventoryAccessRelationshipOptions } from './workspaceAccessPresentation';

describe('workspace access presentation helpers', () => {
  it('builds relationship selector options from canonical frontend-domain values', () => {
    expect(inventoryAccessRelationshipOptions()).toEqual([
      { value: 'viewer', label: 'Viewer' },
      { value: 'editor', label: 'Editor' }
    ]);
  });

  it('formats relationship labels', () => {
    expect(inventoryAccessRelationshipLabel('viewer')).toBe('Viewer');
    expect(inventoryAccessRelationshipLabel('editor')).toBe('Editor');
  });
});
