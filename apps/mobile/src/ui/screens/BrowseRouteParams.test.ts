import { describe, expect, it } from 'vitest';
import { parseBrowseRouteParams } from './BrowseRouteParams';

describe('parseBrowseRouteParams', () => {
  it('parses the supported Browse deep-link parameters', () => {
    expect(
      parseBrowseRouteParams({
        scope: 'containers',
        query: '  camping  ',
        tagId: ['outdoors', 'summer'],
        lifecycleState: 'archived',
        checkoutState: 'checked_out',
        sort: 'id_asc'
      })
    ).toEqual({
      initialScope: 'containers',
      initialQuery: '  camping  ',
      initialTagIds: ['outdoors', 'summer'],
      initialLifecycleState: 'archived',
      initialCheckoutState: 'checked_out',
      initialSort: 'id_asc'
    });
  });

  it('uses the first scalar value for single-value parameters supplied as arrays', () => {
    expect(
      parseBrowseRouteParams({
        scope: ['items', 'places'],
        query: ['drill', 'ignored'],
        lifecycleState: ['all', 'active'],
        checkoutState: ['available', 'any'],
        sort: ['updated_desc', 'id_asc']
      })
    ).toMatchObject({
      initialScope: 'items',
      initialQuery: 'drill',
      initialLifecycleState: 'all',
      initialCheckoutState: 'available',
      initialSort: 'updated_desc'
    });
  });

  it('normalizes multiple tag IDs without losing their route order', () => {
    expect(
      parseBrowseRouteParams({ tagId: [' tools ', '', 'favorites', 'tools', 'favorites'] }).initialTagIds
    ).toEqual(['tools', 'favorites']);
  });

  it('falls back to the default Browse state for absent or unsupported values', () => {
    expect(
      parseBrowseRouteParams({
        scope: 'unknown',
        lifecycleState: 'deleted',
        checkoutState: 'overdue',
        sort: 'name_asc'
      })
    ).toEqual({
      initialScope: 'all',
      initialQuery: '',
      initialTagIds: [],
      initialLifecycleState: 'active',
      initialCheckoutState: 'any',
      initialSort: 'updated_desc'
    });
  });
});
