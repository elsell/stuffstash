import { describe, expect, it } from 'vitest';
import { auditRecordPresentation, auditStatusPresentation } from './workspaceAuditPresentation';

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

  it('builds human-readable audit row presentation with technical details separated', () => {
    const presentation = auditRecordPresentation({
      id: 'audit-one',
      tenantId: 'tenant-one',
      inventoryId: null,
      principalId: 'oidc_local_owner_123',
      action: 'asset.created',
      source: 'api',
      targetType: 'asset',
      targetId: 'asset-tenant-audit',
      occurredAt: '2026-06-24T12:00:00Z',
      requestId: 'request-tenant',
      metadata: { operation_id: 'operation-tenant' }
    });

    expect(presentation.title).toBe('Asset created');
    expect(presentation.actorLabel).toBe('Signed-in user');
    expect(presentation.sourceLabel).toBe('API');
    expect(presentation.targetLabel).toBe('Asset');
    expect(presentation.occurredAtLabel).toContain('2026');
    expect(presentation.primaryText).not.toContain('asset-tenant-audit');
    expect(presentation.primaryText).not.toContain('oidc_local_owner_123');
    expect(presentation.technicalDetails).toEqual([
      { label: 'Action code', value: 'asset.created' },
      { label: 'Target ID', value: 'asset-tenant-audit' },
      { label: 'Principal ID', value: 'oidc_local_owner_123' },
      { label: 'Source', value: 'api' },
      { label: 'Request ID', value: 'request-tenant' },
      { label: 'Metadata operation id', value: 'operation-tenant' }
    ]);
  });

  it('keeps unknown source codes out of the primary audit row', () => {
    const presentation = auditRecordPresentation({
      id: 'audit-one',
      tenantId: 'tenant-one',
      inventoryId: null,
      principalId: 'principal-owner',
      action: 'tenant.created',
      source: 'oidc_google_provider',
      targetType: 'tenant',
      targetId: 'tenant-one',
      occurredAt: 'not-a-date',
      metadata: {}
    });

    expect(presentation.sourceLabel).toBe('Recorded source');
    expect(presentation.occurredAtLabel).toBe('not-a-date');
    expect(presentation.primaryText).not.toContain('oidc_google_provider');
    expect(presentation.technicalDetails).toContainEqual({ label: 'Source', value: 'oidc_google_provider' });
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
