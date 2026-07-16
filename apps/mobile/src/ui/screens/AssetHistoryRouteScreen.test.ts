import { describe, expect, it } from 'vitest';
import type { AssetActivityRecordViewModel } from '../../application/assets/AssetActivityQuery';
import {
  groupHistoryRecords,
  historyLoadError,
  historyRevertConfirmation,
  historyRevertFailure,
  technicalDetailRows
} from './AssetHistoryPresentation';

describe('AssetHistoryRouteScreen presentation', () => {
  it('groups newest-first records with localized date context without changing row order', () => {
    const records = [record('newer', '2026-07-14T15:00:00Z'), record('older', '2026-07-13T15:00:00Z')];
    const sections = groupHistoryRecords(records);
    expect(sections).toHaveLength(2);
    expect(sections.map((section) => section.data[0]?.id)).toEqual(['newer', 'older']);
    expect(sections.every((section) => section.title.length > 0)).toBe(true);
  });

  it('uses safe non-retry states for denied and missing History', () => {
    expect(historyLoadError({ status: 403 })).toMatchObject({ title: 'History unavailable', canRetry: false });
    expect(historyLoadError({ status: 404 })).toMatchObject({ title: 'History unavailable', canRetry: false });
    expect(historyLoadError(new Error('Network unavailable'))).toEqual({ title: 'Could not load History', message: 'Network unavailable', canRetry: true });
  });

  it('builds the collapsed disclosure content from safe identifiers', () => {
    expect(technicalDetailRows({
      id: 'audit-one', principalId: 'principal-one', action: 'asset.updated', category: 'change', source: 'api',
      occurredAt: '2026-07-14T15:00:00Z', requestId: 'request-one', changes: [], technical: { count: '2' }
    })).toEqual([
      { label: 'Audit record', value: 'audit-one' }, { label: 'Action', value: 'asset.updated' },
      { label: 'Source', value: 'api' }, { label: 'Request', value: 'request-one' }, { label: 'count', value: '2' }
    ]);
  });

  it('explains that reverting a historical entry reverses only that change', () => {
    expect(historyRevertConfirmation({
      id: 'audit-one', principalId: 'principal-one', action: 'asset.updated', category: 'change', source: 'api',
      occurredAt: '2026-07-14T15:00:00Z', changes: [
        { field: 'title', previousValue: 'Drill', currentValue: 'Cordless drill' },
        { field: 'tags' }
      ], technical: {}
    })).toEqual({
      title: 'Revert this change?',
      message: 'This will reverse the name and tag changes from this entry. Other changes to the item will stay as they are.',
      confirmLabel: 'Revert Change'
    });
  });

  it('predicts the effect of non-field historical reversals', () => {
    expect(historyRevertConfirmation({
      id: 'audit-create', principalId: 'principal-one', action: 'asset.created', category: 'change', source: 'api',
      occurredAt: '2026-07-14T15:00:00Z', changes: [], technical: {}
    }).message).toBe('This item will be archived. Other changes to the item will stay as they are.');
    expect(historyRevertConfirmation({
      id: 'audit-return', principalId: 'principal-one', action: 'asset.returned', category: 'change', source: 'api',
      occurredAt: '2026-07-14T15:00:00Z', changes: [], technical: {}
    }).message).toBe('The item will be checked out again. Other changes to the item will stay as they are.');
  });

  it('uses safe, actionable copy when later state makes a revert stale', () => {
    expect(historyRevertFailure({ status: 409 })).toEqual({
      title: 'Change can’t be reverted',
      message: 'This item changed afterward, so this change can’t be safely reverted.',
      isTerminal: true
    });
    expect(historyRevertFailure(new Error('Network unavailable'))).toEqual({
      title: 'Could not revert change',
      message: 'Network unavailable',
      isTerminal: false
    });
    expect(historyRevertFailure({ status: 403 })).toEqual({
      title: 'Revert unavailable',
      message: 'You no longer have permission to revert this change.',
      isTerminal: true
    });
    expect(historyRevertFailure({ status: 404 })).toEqual({
      title: 'Change can’t be reverted',
      message: 'This change is no longer available.',
      isTerminal: true
    });
  });
});

function record(id: string, occurredAt: string): AssetActivityRecordViewModel {
  return { id, occurredAt, occurredAtLabel: occurredAt, title: 'Updated item', summary: 'Changed', actorLabel: 'Alex', sourceLabel: 'App' };
}
