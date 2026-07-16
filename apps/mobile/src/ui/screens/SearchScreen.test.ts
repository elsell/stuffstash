import type { RefObject } from 'react';
import type { TextInput } from 'react-native';
import { router } from 'expo-router';
import { describe, expect, it, vi } from 'vitest';
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
  parseBrowseScope,
  searchResultSummaryLabel,
  sortLabel,
  shouldAutoFocusSearchInput
} from './SearchScreenPresentation';
import { SearchHeader } from './SearchScreen';
import { createBrowseHeaderStyles } from './BrowseHeader';
import { InventoryMapHeaderActions } from './InventoryMapScreen';
import { darkPalette, lightPalette } from '../theme/tokens';

vi.mock('expo-router', () => ({
  router: { navigate: vi.fn(), push: vi.fn() },
  useFocusEffect: vi.fn()
}));

vi.mock('lucide-react-native', () => ({
  Camera: 'CameraIcon',
  Check: 'CheckIcon',
  CheckCircle2: 'CheckCircle2Icon',
  ChevronDown: 'ChevronDownIcon',
  ChevronRight: 'ChevronRightIcon',
  Info: 'InfoIcon',
  Map: 'MapIcon',
  MoreHorizontal: 'MoreHorizontalIcon',
  MoveRight: 'MoveRightIcon',
  Package: 'PackageIcon',
  Pencil: 'PencilIcon',
  Plus: 'PlusIcon',
  Search: 'SearchIcon',
  SlidersHorizontal: 'SlidersIcon',
  X: 'XIcon'
}));

vi.mock('react-native-image-viewing', () => ({
  default: 'ImageViewing'
}));

vi.mock('react-native-safe-area-context', () => ({
  SafeAreaView: 'SafeAreaView',
  useSafeAreaInsets: () => ({ bottom: 34, left: 0, right: 0, top: 47 })
}));

vi.mock('react-native', () => ({
  ActionSheetIOS: { showActionSheetWithOptions: vi.fn() },
  ActivityIndicator: 'ActivityIndicator',
  Alert: { alert: vi.fn() },
  DynamicColorIOS: (variants: { readonly light: string }) => variants.light,
  AccessibilityInfo: {
    addEventListener: vi.fn(() => ({ remove: vi.fn() })),
    isReduceMotionEnabled: vi.fn(() => Promise.resolve(false))
  },
  FlatList: 'FlatList',
  Image: 'Image',
  Modal: 'Modal',
  Platform: { OS: 'ios' },
  Pressable: 'Pressable',
  RefreshControl: 'RefreshControl',
  ScrollView: 'ScrollView',
  StyleSheet: {
    create: (styles: unknown) => styles
  },
  Text: 'Text',
  TextInput: 'TextInput',
  View: 'View',
  useColorScheme: () => 'light',
  useWindowDimensions: () => ({ width: 390, height: 844 })
}));

type ElementNode = {
  readonly type?: unknown;
  readonly props?: {
    readonly children?: unknown;
    readonly [key: string]: unknown;
  };
};

describe('SearchScreen presentation helpers', () => {
  it('uses calm filled Browse controls and reserves outlines for search focus', () => {
    const styles = createBrowseHeaderStyles(darkPalette);

    expect(styles.searchBar.backgroundColor).toBe(darkPalette.surfaceMuted);
    expect(styles.searchBar).not.toHaveProperty('borderWidth');
    expect(styles.searchBarFocused.borderWidth).toBe(2);
    expect(styles.scopeButtonSelected).not.toHaveProperty('borderWidth');
    expect(styles.toolButton).not.toHaveProperty('borderWidth');
    expect(styles.toolButton.backgroundColor).toBe(darkPalette.surfaceMuted);
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

  it('counts only non-default secondary filters and describes them by user-facing labels', () => {
    const tags = [
      { id: 'tag-camping', key: 'camping', label: 'Camping' },
      { id: 'tag-tools', key: 'tools', label: 'Tools' }
    ];

    expect(browseFilterCount({ lifecycleState: 'active', checkoutState: 'any', tagIds: [] })).toBe(0);
    expect(browseFilterCount({
      lifecycleState: 'archived',
      checkoutState: 'checked_out',
      tagIds: ['tag-camping', 'tag-tools']
    })).toBe(4);
    expect(buildBrowseFilterTokens({
      lifecycleState: 'archived',
      checkoutState: 'checked_out',
      tagIds: ['tag-tools', 'tag-camping']
    }, tags)).toEqual([
      { key: 'lifecycle', label: 'Archived', type: 'lifecycle' },
      { key: 'checkout', label: 'Checked out', type: 'checkout' },
      { key: 'tag:tag-tools', label: 'Tools', type: 'tag', tagId: 'tag-tools' },
      { key: 'tag:tag-camping', label: 'Camping', type: 'tag', tagId: 'tag-camping' }
    ]);
    expect(sortLabel('updated_desc')).toBe('Recently changed');
    expect(sortLabel('id_asc')).toBe('Default order');
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
    const clearTimer = vi.fn();

    expect(cancelPendingBrowseSearch(timer, '  drill  ', clearTimer)).toBe('drill');
    expect(clearTimer).toHaveBeenCalledWith(timer.current ?? 42);
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

  it('renders Browse as a content-first inventory surface with visible scope and separate tools', () => {
    const inputRef = { current: null } as RefObject<TextInput | null>;
    const header = renderHeader({
      query: 'bike pump',
      searchInputFocused: true,
      searchInputRef: inputRef
    });
    const input = findFirstByType(header, 'TextInput');
    const text = collectText(header);

    expect(input?.props?.ref).toBe(inputRef);
    expect(input?.props?.value).toBe('bike pump');
    expect(findFirstByProp(header, 'accessibilityLabel', 'Sort unavailable during search')?.props?.accessibilityState)
      .toMatchObject({ disabled: true });
    expect(header.props?.style).toMatchObject({ marginBottom: 16 });
    expect(findFirstByProp(header, 'accessibilityLabel', 'Browse view')?.props?.accessibilityRole).toBe('tablist');
    expect(text).toEqual(expect.arrayContaining([
      'Browse',
      'Home inventory',
      'List',
      'Map',
      'All',
      'Places',
      'Containers',
      'Items',
      'Filters',
      'Sort'
    ]));
    expect(input?.props?.placeholder).toBe('Search names, places, or tags');
    expect(text).not.toContain('No tags');
    expect(text).not.toContain('Any');
  });

  it('offers permitted users a full-size Add action from a populated Browse header', () => {
    const permittedHeader = renderHeader({
      canAdd: true,
      resultCount: 3,
      onAdd: () => router.navigate('/add')
    });
    const add = findFirstByProp(permittedHeader, 'accessibilityLabel', 'Add an asset');

    expect(add?.props?.accessibilityRole).toBe('button');
    expect(controlSize(add, 'minHeight')).toBeGreaterThanOrEqual(44);
    expect(controlSize(add, 'minWidth')).toBeGreaterThanOrEqual(44);

    const onPress = add?.props?.onPress;
    if (typeof onPress !== 'function') throw new Error('Missing Browse Add handler');
    onPress();

    expect(router.navigate).toHaveBeenCalledWith('/add');
    expect(findFirstByProp(
      renderHeader({ canAdd: false, resultCount: 3 }),
      'accessibilityLabel',
      'Add an asset'
    )).toBeUndefined();
  });

  it('keeps the permission-aware Add action when Browse switches to Map', () => {
    const permittedActions = InventoryMapHeaderActions({
      canAdd: true,
      palette: lightPalette,
      selectedSurface: 'map',
      onAdd: () => router.navigate('/add'),
      onChangeSurface: vi.fn()
    });
    const add = findFirstByProp(permittedActions, 'accessibilityLabel', 'Add an asset');

    expect(add?.props?.accessibilityRole).toBe('button');
    expect(controlSize(add, 'minHeight')).toBeGreaterThanOrEqual(44);
    (add?.props?.onPress as (() => void) | undefined)?.();
    expect(router.navigate).toHaveBeenCalledWith('/add');

    expect(findFirstByProp(
      InventoryMapHeaderActions({
        canAdd: false,
        palette: lightPalette,
        selectedSurface: 'map',
        onAdd: () => router.navigate('/add'),
        onChangeSurface: vi.fn()
      }),
      'accessibilityLabel',
      'Add an asset'
    )).toBeUndefined();
  });

  it('uses Status, Availability, and Tags groups without hiding scope or mixing in Sort', () => {
    const text = collectText(renderHeader({ filtersExpanded: true, scope: 'places' }));

    expect(text).toContain('Places');
    expect(text).toContain('Status');
    expect(text).toContain('Availability');
    expect(text).toContain('Active');
    expect(text).toContain('Tags');
    expect(text).not.toContain('Checkout');
    expect(text).not.toContain('Stable');
  });

  it('describes submitted search results without presenting a false total', () => {
    const header = renderHeader({
      query: 'mug',
      resultCount: 20,
      submittedQuery: 'mug'
    });
    const text = collectText(header);

    expect(text).toContain('20 shown for “mug” · relevance');
    expect(findFirstByProp(header, 'accessibilityLabel', 'Sort unavailable during search')?.props?.accessibilityState)
      .toMatchObject({ disabled: true });
  });

  it('renders colored multi-select tag filters alphabetically in the filter sheet', () => {
    const selectedTags: string[][] = [];
    const header = renderHeader({
      filtersExpanded: true,
      selectedTagIds: ['tag-tools'],
      filterDraft: { lifecycleState: 'active', checkoutState: 'any', tagIds: ['tag-tools'] },
      tagFilters: [
        { id: 'tag-tools', key: 'tools', label: 'Tools', color: '#2F80ED' },
        { id: 'tag-camping', key: 'camping', label: 'Camping', color: '#2E7D32' },
        { id: 'tag-kids', key: 'kids', label: 'Kids' },
        { id: 'tag-office', key: 'office', label: 'Office' },
        { id: 'tag-travel', key: 'travel', label: 'Travel' }
      ],
      onChangeDraftTagIds: (tagIds) => {
        selectedTags.push([...tagIds]);
      }
    });

    const text = collectText(header);
    expect(text).toEqual(expect.arrayContaining(['Tags', 'Status', 'Availability']));
    expect(text).toContain('Tags');
    expect(text).toContain('Tools');
    const tagFilters = findFirstByProp(header, 'accessibilityLabel', 'Tag filters');
    const tagText = collectText(tagFilters);
    expect(tagText.indexOf('Camping')).toBeLessThan(tagText.indexOf('Kids'));
    expect(tagText.indexOf('Kids')).toBeLessThan(tagText.indexOf('Office'));
    expect(tagText.indexOf('Office')).toBeLessThan(tagText.indexOf('Tools'));
    expect(findFirstByProp(header, 'accessibilityLabel', 'Tag filters')).toBeTruthy();
    expect(findFirstByProp(header, 'accessibilityLabel', 'Sort unavailable during search')?.props?.accessibilityState)
      .toMatchObject({ disabled: true });

    const tools = findFirstByProp(header, 'accessibilityLabel', 'Filter by tag Tools');
    expect(tools).toBeTruthy();
    expect(tools?.props?.selected).toBe(true);
    const onPress = tools?.props?.onPress;
    if (typeof onPress !== 'function') {
      throw new Error('Missing tag filter press handler');
    }
    onPress();

    expect(selectedTags).toEqual([[]]);
  });

  it('uses a compact filter control without hiding primary scope', () => {
    const toggles: boolean[] = [];
    const header = renderHeader({
      filtersExpanded: false,
      onToggleFilters: (expanded) => {
        toggles.push(expanded);
      }
    });

    const filters = findFirstByProp(header, 'accessibilityLabel', 'Filters');
    expect(filters?.props?.accessibilityState).toMatchObject({ expanded: false });
    expect(collectText(header)).toContain('Places');

    const onPress = filters?.props?.onPress;
    if (typeof onPress !== 'function') {
      throw new Error('Missing filter toggle press handler');
    }
    onPress();

    expect(toggles).toEqual([true]);
  });

  it('keeps selected-inventory context recoverable when its metadata request fails', () => {
    const retry = vi.fn();
    const header = renderHeader({
      inventoryContext: undefined,
      inventoryContextStatus: 'error',
      onRetryInventoryContext: retry
    });

    expect(collectText(header)).toContain('Inventory context unavailable');
    const retryControl = findFirstByProp(header, 'accessibilityLabel', 'Retry inventory context');
    expect(retryControl).toBeTruthy();
    const onPress = retryControl?.props?.onPress;
    if (typeof onPress !== 'function') throw new Error('Missing inventory-context retry handler');
    onPress();
    expect(retry).toHaveBeenCalledTimes(1);
  });

  it('shows removable applied-filter labels and clear all when multiple refinements are active', () => {
    const clears: string[] = [];
    const header = renderHeader({
      filtersExpanded: true,
      lifecycleState: 'archived',
      selectedTagIds: ['tag-tools'],
      tagFilters: [{ id: 'tag-tools', key: 'tools', label: 'Tools' }],
      onClearFilters: () => {
        clears.push('filters');
      }
    });

    expect(collectText(header)).toEqual(expect.arrayContaining(['Archived', 'Tools', 'Clear all']));
    const clearFilters = findFirstByText(header, 'Clear all');
    const onPress = clearFilters?.props?.onPress;
    if (typeof onPress !== 'function') {
      throw new Error('Missing clear filters handler');
    }
    onPress();

    expect(clears).toEqual(['filters']);
  });
});

function renderHeader(
  overrides: Partial<Parameters<typeof SearchHeader>[0]> = {}
): ReturnType<typeof SearchHeader> {
  return SearchHeader({
    isLoading: false,
    lifecycleState: 'active',
    checkoutState: 'any',
    filterDraft: { lifecycleState: 'active', checkoutState: 'any', tagIds: [] },
    inventoryContext: 'Home inventory',
    palette: lightPalette,
    query: '',
    resultCount: 0,
    scope: 'all',
    selectedSurface: 'list',
    selectedTagIds: [],
    filtersExpanded: false,
    searchInputFocused: false,
    searchInputRef: { current: null } as RefObject<TextInput | null>,
    sort: 'updated_desc',
    submittedQuery: '',
    onChangeSurface: vi.fn(),
    onApplyFilters: vi.fn(),
    onChangeDraftLifecycleState: vi.fn(),
    onChangeDraftCheckoutState: vi.fn(),
    onChangeDraftTagIds: vi.fn(),
    onChangeQuery: vi.fn(),
    onChangeScope: vi.fn(),
    onChangeSort: vi.fn(),
    onClearQuery: vi.fn(),
    onClearFilters: vi.fn(),
    onRemoveFilter: vi.fn(),
    onToggleFilters: vi.fn(),
    onSearchBlur: vi.fn(),
    onSearchFocus: vi.fn(),
    onSubmit: vi.fn(),
    ...overrides
  });
}

function findFirstByText(node: unknown, text: string): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>((found, child) => found ?? findFirstByText(child, text), undefined);
  }
  if (!isElementNode(node)) return undefined;
  if (collectText(node).includes(text) && node.props?.onPress) return node;
  if (typeof node.type === 'function') return findFirstByText(node.type(node.props), text);
  return childrenOf(node).reduce<ElementNode | undefined>((found, child) => found ?? findFirstByText(child, text), undefined);
}

function findFirstByType(node: unknown, type: string): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>(
      (found, child) => found ?? findFirstByType(child, type),
      undefined
    );
  }

  if (!isElementNode(node)) {
    return undefined;
  }

  if (node.type === type) {
    return node;
  }

  return childrenOf(node).reduce<ElementNode | undefined>(
    (found, child) => found ?? findFirstByType(child, type),
    undefined
  );
}

function findFirstByProp(node: unknown, prop: string, value: unknown): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>(
      (found, child) => found ?? findFirstByProp(child, prop, value),
      undefined
    );
  }

  if (!isElementNode(node)) {
    return undefined;
  }

  if (node.props?.[prop] === value) {
    return node;
  }

  if (typeof node.type === 'function') {
    return findFirstByProp(node.type(node.props), prop, value);
  }

  return childrenOf(node).reduce<ElementNode | undefined>(
    (found, child) => found ?? findFirstByProp(child, prop, value),
    undefined
  );
}

function controlSize(node: ElementNode | undefined, key: 'minHeight' | 'minWidth'): number {
  const style = node?.props?.style;
  const resolved = typeof style === 'function' ? style({ pressed: false }) : style;
  const entries = Array.isArray(resolved) ? resolved : [resolved];

  return entries.reduce<number>((size, entry) => {
    if (!entry || typeof entry !== 'object') return size;
    const value = (entry as Record<string, unknown>)[key];
    return typeof value === 'number' ? Math.max(size, value) : size;
  }, 0);
}

function collectText(node: unknown): readonly string[] {
  if (typeof node === 'string') {
    return [node];
  }

  if (Array.isArray(node)) {
    return node.flatMap(collectText);
  }

  if (!isElementNode(node)) {
    return [];
  }

  if (typeof node.type === 'function') {
    return collectText(node.type(node.props));
  }

  return childrenOf(node).flatMap(collectText);
}

function childrenOf(node: ElementNode): readonly unknown[] {
  const children = node.props?.children;
  return Array.isArray(children) ? children : [children];
}

function isElementNode(node: unknown): node is ElementNode {
  return Boolean(node && typeof node === 'object' && 'props' in node);
}
