import { describe, expect, it } from 'vitest';
import type { Asset } from '$lib/domain/inventory';
import { browseFailureMessage, buildPlaceBrowseSummaries } from './workspaceBrowsePresentation';

const asset = (id: string, title: string, kind: Asset['kind'], parentAssetId: string | null, updatedAt?: string): Asset => ({
  id, tenantId: 't', inventoryId: 'i', title, kind, parentAssetId, updatedAt, description: '', lifecycleState: 'active'
});

describe('Browse place presentation', () => {
  it('precomputes recursive counts and recent contained names', () => {
    const garage = asset('garage', 'Garage', 'location', null);
    const assets = [garage, asset('bin', 'Bin', 'container', 'garage', '2026-01-01'), asset('drill', 'Drill', 'item', 'bin', '2026-02-01')];
    expect(buildPlaceBrowseSummaries([garage], assets)).toEqual([
      { asset: garage, containedCount: 2, recentContainedNames: ['Drill', 'Bin'] }
    ]);
  });

  it('replaces unsafe Browse failures with calm phase-specific recovery copy', () => {
    const unsafe = Object.assign(new Error('database host 10.0.0.8 rejected the query'), { safeForUser: false });

    expect(browseFailureMessage(unsafe, 'initial')).toBe('Browse could not be loaded. Try again.');
    expect(browseFailureMessage(unsafe, 'replacement')).toBe('Browse could not be refreshed. Try again.');
    expect(browseFailureMessage(unsafe, 'append')).toBe('More results could not be loaded. Try again.');
    expect(browseFailureMessage(unsafe, 'map')).toBe('Map could not be loaded. Try again.');
  });

  it('preserves an explicitly safe Browse failure reason', () => {
    const safe = Object.assign(new Error('The selected tag is no longer available.'), { safeForUser: true });
    expect(browseFailureMessage(safe, 'replacement')).toBe('The selected tag is no longer available.');
  });
});
