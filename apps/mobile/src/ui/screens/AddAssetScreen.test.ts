import { describe, expect, it } from 'vitest';
import {
  resolveParentAssetId,
  resolveSelectedParent
} from './AddAssetResolution';
import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';

const garage: ParentLookupResult = {
  id: 'asset-garage',
  title: 'Garage',
  kind: 'location',
  subtitle: 'No parent',
  selectionHint: 'Location',
  willPromoteToContainer: false
};

describe('AddAssetScreen parent resolution', () => {
  it('resolves an exact typed parent without requiring the row to be tapped', () => {
    expect(resolveParentAssetId([garage], ' garage ', undefined)).toBe('asset-garage');
  });

  it('carries an exact typed parent into the next add after save', () => {
    const resolvedParentAssetId = resolveParentAssetId([garage], 'Garage', undefined);

    expect(resolveSelectedParent([garage], resolvedParentAssetId, 'Garage', undefined)).toEqual(
      garage
    );
  });

  it('keeps a newly created place selected even before search returns it again', () => {
    const createdPlace: ParentLookupResult = {
      id: 'asset-basement-shelf',
      title: 'Basement shelf',
      kind: 'location',
      subtitle: 'New location',
      selectionHint: 'Location',
      willPromoteToContainer: false
    };

    expect(
      resolveSelectedParent([], 'asset-basement-shelf', 'Basement shelf', createdPlace)
    ).toEqual(createdPlace);
  });

  it('rejects unknown typed parents so users create or clear the parent first', () => {
    expect(() => resolveParentAssetId([garage], 'Basement shelf', undefined)).toThrow(
      'Create this parent or clear the Put in field.'
    );
  });

});
