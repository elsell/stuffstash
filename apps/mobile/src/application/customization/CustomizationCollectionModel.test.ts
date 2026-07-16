import { describe, expect, it } from 'vitest';
import { beginLifecycleTransition, commitLifecycleTransition, rollbackLifecycleTransition } from './CustomizationCollectionModel';

describe('customization collection lifecycle transitions', () => {
  const active = { lifecycle: 'active' as const, rows: [{ id: 'active-row' }] };

  it('keeps the committed lifecycle and rows while the next collection loads', () => {
    expect(beginLifecycleTransition(active, 'archived')).toEqual({
      lifecycle: 'active',
      pendingLifecycle: 'archived',
      rows: [{ id: 'active-row' }]
    });
  });

  it('commits the selected lifecycle and its rows together', () => {
    const loading = beginLifecycleTransition(active, 'archived');
    expect(commitLifecycleTransition(loading, 'archived', [{ id: 'archived-row' }])).toEqual({
      lifecycle: 'archived',
      pendingLifecycle: undefined,
      rows: [{ id: 'archived-row' }]
    });
  });

  it('rolls back pending selection without discarding the prior rows', () => {
    const loading = beginLifecycleTransition(active, 'archived');
    expect(rollbackLifecycleTransition(loading)).toEqual({
      lifecycle: 'active',
      pendingLifecycle: undefined,
      rows: [{ id: 'active-row' }]
    });
  });
});
