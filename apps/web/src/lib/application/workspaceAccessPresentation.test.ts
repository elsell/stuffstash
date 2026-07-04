import { describe, expect, it } from 'vitest';
import {
  inventoryAccessListStatus,
  inventoryAccessManagerAccessStatus,
  inventoryAccessManagerOperationStatus,
  inventoryAccessRelationshipLabel,
  inventoryAccessRelationshipOptions
} from './workspaceAccessPresentation';

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

  it('builds access manager missing-context and denied statuses', () => {
    expect(inventoryAccessManagerAccessStatus({ hasInventory: false, canShare: true })).toEqual({
      kind: 'missing-context',
      message: 'Select an inventory before managing sharing.'
    });
    expect(inventoryAccessManagerAccessStatus({ hasInventory: true, canShare: false })).toEqual({
      kind: 'denied',
      message: 'You can view this inventory, but you cannot manage sharing.',
      role: 'alert'
    });
    expect(inventoryAccessManagerAccessStatus({ hasInventory: true, canShare: true })).toBeNull();
  });

  it('builds access manager operation error status', () => {
    expect(inventoryAccessManagerOperationStatus('Access service unavailable.')).toEqual({
      kind: 'error',
      message: 'Access service unavailable.',
      role: 'alert'
    });
    expect(inventoryAccessManagerOperationStatus('')).toBeNull();
  });
});
