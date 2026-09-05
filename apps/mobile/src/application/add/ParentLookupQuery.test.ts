import { describe, expect, it } from 'vitest';
import { assetId, AssetSummary } from '../../domain/assets/AssetSummary';
import { ParentLookupQuery } from './ParentLookupQuery';

class FakeParentLookupRepository {
  searchedQuery: string | undefined;
  signal?: AbortSignal;
  async listParentCandidates(query: string, request: { signal?: AbortSignal } = {}) {
    this.searchedQuery = query || undefined;
    this.signal = request.signal;
    if (query === 'medicine') return [
      ...Array.from({ length: 6 }, (_, index) => asset(`fuzzy-${index}`, `Medicine backup ${index}`, 'item', 'Garage')),
      asset('asset-medicine-exact', 'Medicine', 'location', 'No parent')
    ];
    return [asset('asset-garage', 'Garage', 'location', 'No parent'), asset('asset-bin', 'Blue bin', 'container', 'Garage'), asset('asset-drill', 'Cordless drill', 'item', 'Blue bin')]
      .filter((asset) => asset.title.toLocaleLowerCase().includes(query.toLocaleLowerCase()));
  }
}

describe('ParentLookupQuery', () => {
  it('forwards cancellation through the focused port', async () => {
    const repository = new FakeParentLookupRepository();
    const signal = new AbortController().signal;
    await new ParentLookupQuery(repository).execute(' drill ', { signal });
    expect(repository.searchedQuery).toBe('drill');
    expect(repository.signal).toBe(signal);
  });
  it('uses recent inventory assets as bounded empty-query parent candidates', async () => {
    const repository = new FakeParentLookupRepository();
    const query = new ParentLookupQuery(repository);

    await expect(query.execute('   ')).resolves.toMatchObject([
      { id: 'asset-garage', title: 'Garage', kind: 'location', willPromoteToContainer: false },
      { id: 'asset-bin', title: 'Blue bin', kind: 'container', willPromoteToContainer: false },
      {
        id: 'asset-drill',
        title: 'Cordless drill',
        kind: 'item',
        selectionHint: 'Will become a container for this item',
        willPromoteToContainer: true,
        canSelectAsParent: true
      }
    ]);
    expect(repository.searchedQuery).toBeUndefined();
  });

  it('searches every asset kind for parent candidates', async () => {
    const repository = new FakeParentLookupRepository();
    const query = new ParentLookupQuery(repository);

    await expect(query.execute('drill')).resolves.toMatchObject([
      {
        id: 'asset-drill',
        title: 'Cordless drill',
        kind: 'item',
        selectionHint: 'Will become a container for this item',
        willPromoteToContainer: true,
        canSelectAsParent: true
      }
    ]);
    expect(repository.searchedQuery).toBe('drill');
  });

  it('keeps exact parent title matches before trimming compact search results', async () => {
    const repository = new FakeParentLookupRepository();
    const query = new ParentLookupQuery(repository);

    const results = await query.execute('medicine');

    expect(results[0]).toMatchObject({
      id: 'asset-medicine-exact',
      title: 'Medicine',
      kind: 'location',
      selectionHint: 'Location',
      willPromoteToContainer: false
    });
    expect(results).toHaveLength(6);
  });
});

function asset(
  id: string,
  title: string,
  kind: AssetSummary['kind'],
  locationLabel: string
): AssetSummary {
  return {
    id: assetId(id),
    title,
    kind,
    lifecycleState: 'active',
    locationLabel,
    locationTrail: ['Home', locationLabel, title].filter((value) => value !== 'No parent'),
    parentLocationTrail: locationLabel === 'No parent'
      ? []
      : [{ id: assetId(`asset-${locationLabel.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '')}`), title: locationLabel }],
    description: '',
    updatedAtLabel: 'Updated today',
    hasPhoto: false
  };
}
