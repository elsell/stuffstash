import { describe, expect, it } from 'vitest';
import {
  browseRouteParamsForState,
  consumeLocalBrowseRouteEffect,
  parseBrowseRouteParams
} from './BrowseRouteParams';

describe('parseBrowseRouteParams', () => {
  it('parses the supported Browse deep-link parameters', () => {
    expect(
      parseBrowseRouteParams({
        surface: 'map',
        scope: 'containers',
        query: '  camping  ',
        tagId: ['outdoors', 'summer'],
        lifecycleState: 'archived',
        checkoutState: 'checked_out',
        sort: 'id_asc'
      })
    ).toEqual({
      initialSurface: 'map',
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
        surface: ['list', 'map'],
        scope: ['items', 'places'],
        query: ['drill', 'ignored'],
        lifecycleState: ['all', 'active'],
        checkoutState: ['available', 'any'],
        sort: ['updated_desc', 'id_asc']
      })
    ).toMatchObject({
      initialSurface: 'list',
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
        surface: 'grid',
        scope: 'unknown',
        lifecycleState: 'deleted',
        checkoutState: 'overdue',
        sort: 'name_asc'
      })
    ).toEqual({
      initialSurface: 'list',
      initialScope: 'all',
      initialQuery: '',
      initialTagIds: [],
      initialLifecycleState: 'active',
      initialCheckoutState: 'any',
      initialSort: 'updated_desc'
    });
  });
});

describe('browseRouteParamsForState', () => {
  it('writes typed applied Browse state and omits default values', () => {
    expect(browseRouteParamsForState({
      surface: 'map',
      scope: 'containers',
      query: ' drill ',
      tagIds: ['tag-tools'],
      lifecycleState: 'archived',
      checkoutState: 'available',
      sort: 'id_asc'
    })).toEqual({
      surface: 'map',
      scope: 'containers',
      query: 'drill',
      tagId: ['tag-tools'],
      lifecycleState: 'archived',
      checkoutState: 'available',
      sort: 'id_asc'
    });

    expect(browseRouteParamsForState({
      surface: 'list',
      scope: 'all',
      query: '',
      tagIds: [],
      lifecycleState: 'active',
      checkoutState: 'any',
      sort: 'updated_desc'
    })).toEqual({
      surface: undefined,
      scope: undefined,
      query: undefined,
      tagId: undefined,
      lifecycleState: undefined,
      checkoutState: undefined,
      sort: undefined
    });
  });

  it('consumes a locally-originated route echo exactly once', () => {
    const localRouteKeys = new Set(['map-route']);

    expect(consumeLocalBrowseRouteEffect(localRouteKeys, 'map-route')).toBe(true);
    expect(localRouteKeys.size).toBe(0);
    expect(consumeLocalBrowseRouteEffect(localRouteKeys, 'map-route')).toBe(false);
  });
});
