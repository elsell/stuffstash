import { describe, expect, it } from 'vitest';
import { auditStatusPresentation } from './workspaceAuditPresentation';

describe('workspace audit presentation helpers', () => {
  it('builds missing-context and denied audit status presentation', () => {
    expect(baseStatus({ hasInventory: false })).toEqual({
      kind: 'missing-context',
      message: 'Select an inventory before viewing audit history.'
    });
    expect(baseStatus({ scope: 'tenant', canReadScope: false })).toEqual({
      kind: 'denied',
      message: 'Tenant audit history requires tenant configuration access.',
      role: 'alert'
    });
    expect(baseStatus({ canReadScope: false })).toEqual({
      kind: 'denied',
      message: 'Inventory audit history requires inventory view access.',
      role: 'alert'
    });
  });

  it('builds loading, error, empty, and ready audit status presentation', () => {
    expect(baseStatus({ error: 'Audit service unavailable.' })).toEqual({
      kind: 'error',
      message: 'Audit service unavailable.',
      role: 'alert'
    });
    expect(baseStatus({ busy: true, loaded: false })).toEqual({
      kind: 'loading',
      message: 'Loading audit history...',
      role: 'status'
    });
    expect(baseStatus({ busy: false, loaded: false, recordCount: 0 })).toEqual({
      kind: 'none',
      message: ''
    });
    expect(baseStatus({ loaded: true, recordCount: 0 })).toEqual({
      kind: 'empty',
      message: 'No audit records found.'
    });
    expect(baseStatus({ loaded: true, recordCount: 1 })).toEqual({
      kind: 'none',
      message: ''
    });
  });
});

function baseStatus(overrides: Partial<Parameters<typeof auditStatusPresentation>[0]> = {}) {
  return auditStatusPresentation({
    hasTenant: true,
    hasInventory: true,
    scope: 'inventory',
    canReadScope: true,
    error: '',
    busy: false,
    loaded: false,
    recordCount: 0,
    ...overrides
  });
}
