import { describe, expect, it } from 'vitest';
import type { ParentTargetViewModel } from '$lib/domain/inventory';
import { normalizeParentTargetQuery, parentTargetSuggestions, searchParentTargets } from './workspaceParentTargets';

describe('workspace parent target helpers', () => {
  it('builds bounded location-first suggestions without the selected target', () => {
    const suggestions = parentTargetSuggestions(
      [
        parentTarget('garage-shelf', 'Garage shelf', 'Garage', 'container'),
        parentTarget('hall-closet', 'Hall closet', 'Hall', 'location'),
        parentTarget('attic', 'Attic', 'Upstairs', 'location'),
        parentTarget('toolbox', 'Toolbox', 'Garage', 'container')
      ],
      'attic',
      3
    );

    expect(suggestions.map((target) => target.id)).toEqual(['hall-closet', 'garage-shelf', 'toolbox']);
  });

  it('ranks search results within kind by title strength before containment trail matches', () => {
    const result = searchParentTargets(
      [
        parentTarget('garage-shelf', 'Garage shelf', 'Garage'),
        parentTarget('shelf-rack', 'Shelf rack', 'Storage'),
        parentTarget('storage-shelf', 'Shelf', 'Storage'),
        parentTarget('bin', 'Utility bin', 'Garage / Shelf')
      ],
      'shelf',
      4
    );

    expect(result.visibleTargets.map((target) => target.id)).toEqual(['storage-shelf', 'shelf-rack', 'garage-shelf', 'bin']);
  });

  it('groups visible search results by location and container after limiting', () => {
    const result = searchParentTargets(
      [
        parentTarget('hall', 'Hall closet', 'Hall', 'location'),
        parentTarget('attic', 'Attic closet', 'Upstairs', 'location'),
        parentTarget('bin', 'Closet bin', 'Hall', 'container')
      ],
      'closet',
      2
    );

    expect(result.matchingTargets).toHaveLength(3);
    expect(result.visibleTargets.map((target) => target.id)).toEqual(['attic', 'hall']);
    expect(result.locationResults.map((target) => target.id)).toEqual(['attic', 'hall']);
    expect(result.containerResults).toEqual([]);
  });

  it('normalizes search queries before matching', () => {
    expect(normalizeParentTargetQuery('  Garage Shelf  ')).toBe('garage shelf');
    expect(searchParentTargets([parentTarget('garage', 'Garage shelf', 'Root')], '  GARAGE  ', 8).visibleTargets).toHaveLength(1);
  });
});

function parentTarget(
  id: string,
  title: string,
  containmentTrail: string,
  kind: ParentTargetViewModel['kind'] = 'container'
): ParentTargetViewModel {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind,
    title,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    containmentTrail
  };
}
