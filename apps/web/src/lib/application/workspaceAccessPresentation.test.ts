import { describe, expect, it } from 'vitest';
import { inventoryAccessListStatus, inventoryAccessRelationshipLabel, inventoryAccessRelationshipOptions } from './workspaceAccessPresentation';

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

  it('builds access list loading and empty status presentation', () => {
    expect(inventoryAccessListStatus({ kind: 'grants', busy: true, loaded: false, count: 0 })).toEqual({
      kind: 'loading',
      message: 'Loading grants...',
      role: 'status'
    });
    expect(inventoryAccessListStatus({ kind: 'invitations', busy: true, loaded: false, count: 0 })).toEqual({
      kind: 'loading',
      message: 'Loading invitations...',
      role: 'status'
    });
    expect(inventoryAccessListStatus({ kind: 'grants', busy: false, loaded: true, count: 0 })).toEqual({
      kind: 'empty',
      message: 'No direct grants.'
    });
    expect(inventoryAccessListStatus({ kind: 'grants', busy: false, loaded: false, count: 0 })).toEqual({
      kind: 'none',
      message: ''
    });
    expect(inventoryAccessListStatus({ kind: 'invitations', busy: false, loaded: true, count: 1 })).toEqual({
      kind: 'none',
      message: ''
    });
  });
});
