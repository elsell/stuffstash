import { describe, expect, it } from 'vitest';
import type { ParentTargetViewModel } from '$lib/domain/inventory';
import {
  addAssetKindCopy,
  addDestinationSummary,
  addPhotoCountLabel,
  assetKindControlOptions,
  quickParentContainerLabel,
  quickParentContainerSummary,
  quickParentContainerTrail,
  quickParentKindOptions
} from './workspaceAddPresentation';

describe('workspace add presentation helpers', () => {
  it('builds stable asset-kind and quick-parent options', () => {
    expect(assetKindControlOptions()).toEqual([
      { value: 'item', label: 'Item' },
      { value: 'container', label: 'Container' },
      { value: 'location', label: 'Location' }
    ]);
    expect(quickParentKindOptions).toEqual([
      { value: 'location', label: 'Location' },
      { value: 'container', label: 'Container' }
    ]);
  });

  it('builds kind-specific add labels and placeholders', () => {
    expect(addAssetKindCopy('item')).toEqual({
      heading: 'Add item',
      kindLabel: 'Item',
      nameLabel: 'Item name',
      namePlaceholder: 'Tomato fertilizer',
      saveLabel: 'Save item',
      selectedKindLabel: 'item'
    });
    expect(addAssetKindCopy('container').namePlaceholder).toBe('Clear storage bin');
    expect(addAssetKindCopy('location').namePlaceholder).toBe('Garage shelf');
  });

  it('summarizes existing and quick-created parent destinations', () => {
    const parent = parentTarget('garage-shelf', 'Garage shelf', 'Garage');

    expect(addDestinationSummary({ quickParentEnabled: false, quickParentKind: 'location', quickParentTitle: '', selectedParent: null })).toBe(
      'Inventory root'
    );
    expect(addDestinationSummary({ quickParentEnabled: false, quickParentKind: 'location', quickParentTitle: '', selectedParent: parent })).toBe(
      'Garage shelf'
    );
    expect(
      addDestinationSummary({
        quickParentEnabled: true,
        quickParentKind: 'container',
        quickParentTitle: '  Clear bin  ',
        selectedParent: parent
      })
    ).toBe('New Container: Clear bin in Garage shelf / Garage');
    expect(addDestinationSummary({ quickParentEnabled: true, quickParentKind: 'location', quickParentTitle: '', selectedParent: null })).toBe(
      'New Location in Inventory root'
    );
  });

  it('summarizes quick-parent container context and photo counts', () => {
    const parent = parentTarget('hall-closet', 'Hall closet', 'Hall');

    expect(quickParentContainerLabel(null)).toBe('Inventory root');
    expect(quickParentContainerLabel(parent)).toBe('Hall closet');
    expect(quickParentContainerTrail(null)).toBe('');
    expect(quickParentContainerTrail(parent)).toBe('Hall');
    expect(quickParentContainerSummary(parent)).toBe('Hall closet / Hall');
    expect(addPhotoCountLabel(0)).toBe('No photos');
    expect(addPhotoCountLabel(1)).toBe('1 photo');
    expect(addPhotoCountLabel(2)).toBe('2 photos');
  });
});

function parentTarget(id: string, title: string, containmentTrail: string): ParentTargetViewModel {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind: 'container',
    title,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    containmentTrail
  };
}
