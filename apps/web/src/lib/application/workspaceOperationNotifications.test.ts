import { describe, expect, it } from 'vitest';
import { operationRefreshWarning, safeOperationFailureDescription } from './workspaceOperationNotifications';

describe('workspace operation notifications', () => {
  it('models refresh degradation as a persistent warning', () => {
    expect(operationRefreshWarning('operation-one', 'Undid change to Drill.', { label: 'Redo', onClick: () => {} })).toMatchObject({
      id: 'asset-operation-refresh:operation-one',
      kind: 'warning',
      duration: Infinity,
      important: true,
      description: expect.stringContaining('Undid change to Drill.'),
      action: { label: 'Redo' }
    });
  });

  it('preserves only explicitly safe failure detail', () => {
    const safe = Object.assign(new Error('This change is stale because the asset changed later.'), { safeForUser: true });
    const unsafe = Object.assign(new Error('database host 10.0.0.8 rejected operation'), { safeForUser: false });

    expect(safeOperationFailureDescription(safe)).toContain('stale');
    expect(safeOperationFailureDescription(unsafe)).not.toContain('database host');
  });
});
