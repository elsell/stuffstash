import type { RefObject } from 'react';
import type { TextInput } from 'react-native';
import { describe, expect, it } from 'vitest';
import {
  browseScopeToKind,
  browseColumnCount,
  browseContinuationCriteria,
  browseGridCardWidth,
  browseLoadingFlagsForRefresh,
  buildBrowseScopeOptions,
  buildBrowseFilterTokens,
  browseFilterCount,
  cancelPendingBrowseSearch,
  canLoadNextBrowsePage,
  focusSearchInput,
  locationRowsFromAssetCards,
  commitBrowseFilterDraft,
  openBrowseFilterDraft,
  parseBrowseScope,
  removeBrowseFilter,
  searchResultSummaryLabel,
  sortLabel,
  shouldAutoFocusSearchInput
} from './SearchScreenPresentation';
import { createBrowseHeaderStyles } from './BrowseHeader';
import { darkPalette, lightPalette } from '../theme/tokens';

describe('SearchScreen presentation helpers', () => {
  it('uses calm Browse fields while native controls own refinement styling', () => {
    const styles = createBrowseHeaderStyles(darkPalette);

    expect(styles).not.toHaveProperty('searchBar');
    expect(styles.resultToolsRow).toMatchObject({ alignItems: 'center', minHeight: 44 });
    expect(styles.activeFilterToken).toMatchObject({ minHeight: 44 });
    expect(styles.activeFilterTokenPill).toMatchObject({ minHeight: 32 });
    expect(styles).not.toHaveProperty('toolButton');
  });

  it('focuses the search input only after an explicit search action', () => {
    let focusCount = 0;
    const inputRef = {
      current: {
        focus: () => {
          focusCount += 1;
        }
      }
    } as RefObject<TextInput | null>;

    focusSearchInput(inputRef);

    expect(focusCount).toBe(1);
  });

  it('does not auto-focus the browse search input on any browse entry', () => {
    expect(shouldAutoFocusSearchInput([])).toBe(false);
    expect(shouldAutoFocusSearchInput([''])).toBe(false);
    expect(shouldAutoFocusSearchInput(['tag-workshop'])).toBe(false);
    expect(shouldAutoFocusSearchInput([' ', 'tag-camping'])).toBe(false);
  });

  it('counts only non-default filters and describes them by user-facing labels', () => {
    const tags = [
      { id: 'tag-camping', key: 'camping', label: 'Camping' },
      { id: 'tag-tools', key: 'tools', label: 'Tools' }
    ];

    expect(browseFilterCount({ scope: 'all', lifecycleState: 'active', checkoutState: 'any', tagIds: [] })).toBe(0);
    expect(browseFilterCount({ scope: 'containers', lifecycleState: 'active', checkoutState: 'any', tagIds: [] })).toBe(1);
    expect(browseFilterCount({
      scope: 'items',
      lifecycleState: 'archived',
      checkoutState: 'checked_out',
      tagIds: ['tag-camping', 'tag-tools']
    })).toBe(5);
    expect(buildBrowseFilterTokens({
      scope: 'places',
      lifecycleState: 'archived',
      checkoutState: 'checked_out',
      tagIds: ['tag-tools', 'tag-camping']
    }, tags)).toEqual([
      { key: 'scope', label: 'Places', type: 'scope' },
      { key: 'lifecycle', label: 'Archived', type: 'lifecycle' },
      { key: 'checkout', label: 'Checked out', type: 'checkout' },
      { key: 'tag:tag-tools', label: 'Tools', type: 'tag', tagId: 'tag-tools' },
      { key: 'tag:tag-camping', label: 'Camping', type: 'tag', tagId: 'tag-camping' }
    ]);
    expect(sortLabel('updated_desc')).toBe('Recently changed');
    expect(sortLabel('id_asc')).toBe('Default order');
  });

  it('removes an applied Type without disturbing the other applied filters', () => {
    expect(removeBrowseFilter({
      scope: 'containers',
      lifecycleState: 'archived',
      checkoutState: 'available',
      tagIds: ['tag-tools']
    }, { key: 'scope', label: 'Containers', type: 'scope' })).toEqual({
      scope: 'all',
      lifecycleState: 'archived',
      checkoutState: 'available',
      tagIds: ['tag-tools']
    });
  });

  it('offers browse scopes that collapse search and locations into one surface', () => {
    expect(buildBrowseScopeOptions()).toEqual([
      { label: 'All', value: 'all' },
      { label: 'Places', value: 'places' },
      { label: 'Containers', value: 'containers' },
      { label: 'Items', value: 'items' }
    ]);
    expect(browseScopeToKind('all')).toBe('all');
    expect(browseScopeToKind('places')).toBe('location');
    expect(browseScopeToKind('containers')).toBe('container');
    expect(browseScopeToKind('items')).toBe('item');
  });

  it('keeps the two-column asset grid on ordinary phones and adapts only when space or text requires it', () => {
    expect(browseColumnCount({ fontScale: 1, scope: 'all', width: 390 })).toBe(2);
    expect(browseColumnCount({ fontScale: 1, scope: 'items', width: 393 })).toBe(2);
    expect(browseColumnCount({ fontScale: 1, scope: 'places', width: 393 })).toBe(1);
    expect(browseColumnCount({ fontScale: 1.4, scope: 'all', width: 393 })).toBe(1);
    expect(browseColumnCount({ fontScale: 1, scope: 'all', width: 340 })).toBe(1);
    expect(browseGridCardWidth(393, 2)).toBe(175);
    expect(browseGridCardWidth(393, 1)).toBeUndefined();
  });

  it('continues only the loaded page criteria and blocks pagination after a failed replacement', () => {
    const loadedCriteria = {
      query: 'drill',
      lifecycleState: 'active' as const,
      checkoutState: 'available' as const,
      scope: 'items' as const,
      sort: 'updated_desc' as const,
      tagIds: ['tag-tools']
    };

    expect(browseContinuationCriteria(loadedCriteria)).toEqual(loadedCriteria);
    expect(canLoadNextBrowsePage('ready')).toBe(true);
    expect(canLoadNextBrowsePage('error', 'pagination')).toBe(true);
    expect(canLoadNextBrowsePage('error', 'replacement')).toBe(false);
    expect(canLoadNextBrowsePage('loading')).toBe(false);
  });

  it('cancels pending debounced text search before applying a refinement', () => {
    const timer = { current: 42 as unknown as ReturnType<typeof setTimeout> };
    const cleared: unknown[] = [];
    const clearTimer = (timer: ReturnType<typeof setTimeout>) => { cleared.push(timer); };

    expect(cancelPendingBrowseSearch(timer, '  drill  ', clearTimer)).toBe('drill');
    expect(cleared).toEqual([42]);
    expect(timer.current).toBeUndefined();
  });

  it('clears in-flight pagination state when pull-to-refresh takes ownership', () => {
    expect(browseLoadingFlagsForRefresh()).toEqual({
      isLoadingMore: false,
      isRefreshing: true
    });
  });

  it('parses Browse scope route params safely', () => {
    expect(parseBrowseScope('places')).toBe('places');
    expect(parseBrowseScope(['containers'])).toBe('containers');
    expect(parseBrowseScope('unknown')).toBe('all');
    expect(parseBrowseScope(undefined)).toBe('all');
  });

  it('enriches API-backed place asset rows with location summary metadata', () => {
    const rows = locationRowsFromAssetCards([
      {
        id: 'kitchen',
        title: 'Kitchen',
        kindLabel: 'Location',
        customTypeLabel: undefined,
        description: 'Cooking and pantry storage',
        locationTrailLabel: 'Home / Kitchen',
        parentLocationTrail: [
          { id: 'asset-home', title: 'Home', isImmediateParent: true }
        ],
        updatedAtLabel: 'Updated today',
        photoLabel: 'Needs photo',
        imagePlaceholderLabel: 'Place',
        photo: undefined
      },
      {
        id: 'garage',
        title: 'Garage',
        kindLabel: 'Location',
        customTypeLabel: undefined,
        description: 'Tools and seasonal bins',
        locationTrailLabel: 'Home / Garage',
        parentLocationTrail: [
          { id: 'asset-home', title: 'Home', isImmediateParent: true }
        ],
        updatedAtLabel: 'Updated today',
        photoLabel: 'Photo ready',
        imagePlaceholderLabel: 'Place',
        photo: { uri: 'https://photos/garage.jpg' }
      },
      {
        id: 'attic',
        title: 'Attic',
        kindLabel: 'Location',
        customTypeLabel: undefined,
        description: 'Long-term storage',
        locationTrailLabel: 'Home / Attic',
        parentLocationTrail: [
          { id: 'asset-home', title: 'Home', isImmediateParent: true }
        ],
        updatedAtLabel: 'Updated today',
        photoLabel: 'Needs photo',
        imagePlaceholderLabel: 'Place',
        photo: undefined
      }
    ], [
      {
        id: 'kitchen',
        title: 'Kitchen',
        description: 'Cooking and pantry storage',
        containedAssetCountLabel: '12 assets',
        recentAssetLabel: 'Water bottle, travel mug',
        photoLabel: 'Needs photo'
      },
      {
        id: 'garage',
        title: 'Garage',
        description: 'Tools and seasonal bins',
        containedAssetCountLabel: '8 assets',
        recentAssetLabel: 'Drill, socket set',
        photoLabel: 'Photo ready'
      }
    ]);

    expect(rows).toEqual([
      {
        id: 'kitchen',
        title: 'Kitchen',
        description: 'Cooking and pantry storage',
        containedAssetCountLabel: '12 assets',
        recentAssetLabel: 'Water bottle, travel mug',
        photo: undefined
      },
      {
        id: 'garage',
        title: 'Garage',
        description: 'Tools and seasonal bins',
        containedAssetCountLabel: '8 assets',
        recentAssetLabel: 'Drill, socket set',
        photo: { uri: 'https://photos/garage.jpg' }
      },
      {
        id: 'attic',
        title: 'Attic',
        description: 'Long-term storage',
        containedAssetCountLabel: 'Contents not summarized',
        recentAssetLabel: 'Home / Attic',
        photo: undefined
      }
    ]);
  });

  it('summarizes the active browse state with scope, query, and sort', () => {
    expect(searchResultSummaryLabel({
      lifecycleState: 'active',
      query: 'drill',
      resultCount: 4,
      scope: 'containers',
      sort: 'updated_desc'
    })).toBe('4 shown for “drill” · relevance');
    expect(searchResultSummaryLabel({
      lifecycleState: 'all',
      query: '',
      resultCount: 2,
      scope: 'places',
      sort: 'id_asc'
    })).toBe('2 shown · Default order');
    expect(searchResultSummaryLabel({
      hasTagFilters: true,
      lifecycleState: 'active',
      query: '',
      resultCount: 3,
      scope: 'all',
      sort: 'updated_desc'
    })).toBe('3 shown · relevance');
  });

  it('commits an applied Type and restores the applied Type after cancelling a later draft', () => {
    const initialApplied = { scope: 'places' as const, lifecycleState: 'active' as const, checkoutState: 'any' as const, tagIds: ['tag-tools'] };
    const firstDraft = { ...openBrowseFilterDraft(initialApplied), scope: 'items' as const };
    const applied = commitBrowseFilterDraft(firstDraft);

    expect(applied).toEqual({ ...initialApplied, scope: 'items' });

    const cancelledDraft = { ...openBrowseFilterDraft(applied), scope: 'containers' as const };
    expect(cancelledDraft.scope).toBe('containers');
    expect(openBrowseFilterDraft(applied)).toEqual(applied);
    expect(openBrowseFilterDraft(applied)).not.toBe(applied);
  });

});
