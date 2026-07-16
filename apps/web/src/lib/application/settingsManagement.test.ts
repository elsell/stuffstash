import { describe, expect, it } from 'vitest';
import { collectSettingsPages, invalidateSharedSettingsLoads, mergeCanonicalSettingsRecord, normalizeTagColor, removeCanonicalSettingsRecord, settingsKeyFromName, sharePendingSettingsLoad, sortSettingsRecords, tagColorAccessibleLabel, utf8ByteLength } from './settingsManagement';

describe('settings management helpers', () => {
  it('loads bounded cursor pages without claiming a repeated cursor is complete', async () => {
    await expect(collectSettingsPages(async () => ({ items: [{ id: 'one', displayName: 'One' }], pagination: { limit: 1, nextCursor: 'same', hasMore: true } })))
      .rejects.toThrow('Settings collection could not be fully loaded.');
  });

  it('shares only concurrent settings loads for the same repository and key', async () => {
    const owner = {};
    let resolve!: (value: string) => void;
    let calls = 0;
    const load = () => { calls += 1; return new Promise<string>((done) => { resolve = done; }); };
    const first = sharePendingSettingsLoad(owner, 'types', load);
    const second = sharePendingSettingsLoad(owner, 'types', load);
    expect(second).toBe(first);
    expect(calls).toBe(1);
    resolve('loaded');
    await expect(first).resolves.toBe('loaded');
    const third = sharePendingSettingsLoad(owner, 'types', async () => 'fresh');
    expect(third).not.toBe(first);
    await expect(third).resolves.toBe('fresh');
  });

  it('reuses resolved settings queries until their semantic collection is invalidated', async () => {
    const owner = {};
    let calls = 0;
    const load = async () => { calls += 1; return calls; };
    await expect(sharePendingSettingsLoad(owner, 'custom-fields:inventory:active', load, true)).resolves.toBe(1);
    await expect(sharePendingSettingsLoad(owner, 'custom-fields:inventory:active', load, true)).resolves.toBe(1);
    invalidateSharedSettingsLoads(owner, 'custom-fields:inventory:');
    await expect(sharePendingSettingsLoad(owner, 'custom-fields:inventory:active', load, true)).resolves.toBe(2);
  });

  it('sorts names naturally and uses stable IDs as a tiebreaker', () => {
    expect(sortSettingsRecords([{ id: 'b', displayName: 'Bin 2' }, { id: 'a', displayName: 'bin 2' }, { id: 'c', displayName: 'Bin 10' }]).map((item) => item.id))
      .toEqual(['a', 'b', 'c']);
  });

  it('merges a restored record into the canonical active collection without losing unrelated scopes', () => {
    const canonical = [
      { id: 'tenant-type', displayName: 'Appliance', scope: 'tenant' },
      { id: 'inventory-type', displayName: 'Tool', scope: 'inventory' }
    ];

    expect(mergeCanonicalSettingsRecord(canonical, { id: 'restored-type', displayName: 'Furniture', scope: 'inventory' }))
      .toEqual([
        { id: 'tenant-type', displayName: 'Appliance', scope: 'tenant' },
        { id: 'restored-type', displayName: 'Furniture', scope: 'inventory' },
        { id: 'inventory-type', displayName: 'Tool', scope: 'inventory' }
      ]);
  });

  it('removes only an archived record from the canonical active collection', () => {
    const canonical = [{ id: 'keep', displayName: 'Keep' }, { id: 'archive', displayName: 'Archive' }];
    expect(removeCanonicalSettingsRecord(canonical, 'archive')).toEqual([{ id: 'keep', displayName: 'Keep' }]);
    expect(removeCanonicalSettingsRecord(canonical, 'already-archived')).toEqual(canonical);
  });

  it('normalizes suggested keys and optional tag colors', () => {
    expect(settingsKeyFromName(' Expiration Date ')).toBe('expiration-date');
    expect(normalizeTagColor('2f80ed')).toBe('#2F80ED');
    expect(normalizeTagColor('')).toBeUndefined();
    expect(normalizeTagColor('blue')).toBeNull();
  });

  it('counts UTF-8 bytes and describes colors with a perceptual name', () => {
    expect(utf8ByteLength('é')).toBe(2);
    expect(utf8ByteLength('📦')).toBe(4);
    expect(tagColorAccessibleLabel()).toBe('No color');
    expect(tagColorAccessibleLabel('#2F80ED')).toContain('Blue color');
  });
});
