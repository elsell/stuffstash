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
  commitBrowseFilterDraft,
  openBrowseFilterDraft,
  parseBrowseScope,
  removeBrowseFilter,
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
  Bell: 'BellIcon',
  UserCircle: 'UserCircleIcon',
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

vi.mock('../components/NativeActionMenu', () => ({
  NativeActionMenu: (props: { readonly trigger?: { readonly kind: string; readonly label?: string } }) => ({
    type: 'NativeActionMenu',
    props: {
      ...props,
      children: props.trigger?.kind === 'label' ? props.trigger.label : 'SortIcon'
    }
  })
}));

vi.mock('../components/NativeRefinementButton', () => ({
  NativeRefinementButton: (props: { readonly label: string }) => ({
    type: 'NativeRefinementButton',
    props: { ...props, children: props.label }
  })
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

  it('renders Browse as a content-first inventory surface with Type disclosed through Filters', () => {
    const header = renderHeader({

    });
    const input = findFirstByProp(header, 'placeholder', 'Search names, places, or tags');
    const text = collectText(header);

    expect(input).toBeUndefined();
    expect(header.props?.style).toMatchObject({ marginBottom: 16 });
    expect(findFirstByProp(header, 'accessibilityLabel', 'Browse view')?.props?.accessibilityRole).toBe('tablist');
    expect(findFirstByProp(header, 'selectedIndex', 0)?.props?.values).toEqual(['List', 'Map']);
    expect(findFirstByProp(header, 'accessibilityLabel', 'Filter by type')).toBeUndefined();
    expect(findFirstByProp(header, 'accessibilityLabel', 'Browse by kind')).toBeUndefined();
    expect(text).not.toContain('Home inventory');
    expect(text).not.toContain('No tags');
    expect(text).not.toContain('Any');
  });

  it('keeps creation actions out of List and Map content headers', () => {
    expect(findFirstByProp(renderHeader(), 'accessibilityLabel', 'Add an asset')).toBeUndefined();
    expect(findFirstByProp(InventoryMapHeaderActions({
      palette: lightPalette, selectedSurface: 'map', onChangeSurface: () => {}
    }), 'accessibilityLabel', 'Add an asset')).toBeUndefined();
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

  it('describes submitted search results without presenting a false total', () => {
    const header = renderHeader({

      resultCount: 20,
      submittedQuery: 'mug'
    });
    const text = collectText(header);

    expect(text).toContain('20 shown for “mug” · relevance');
  });



  it('uses a compact native filter disclosure', () => {
    let toggles = 0;
    const header = renderHeader({

      onToggleFilters: () => { toggles++; }
    });

    const filters = findFirstByProp(header, 'accessibilityLabel', 'Filters');
    expect(filters?.props?.label).toBe('Filters');
    expect(filters?.props?.systemImage).toBe('line.3.horizontal.decrease');

    const onPress = filters?.props?.onPress;
    if (typeof onPress !== 'function') {
      throw new Error('Missing filter toggle press handler');
    }
    onPress();

    expect(toggles).toBe(1);
  });

  it('presents one native filter button with applied-count feedback', () => {
    const header = renderHeader({
      scope: 'containers',
      lifecycleState: 'archived',
      selectedTagIds: []
    });
    const filters = findFirstByProp(header, 'accessibilityLabel', 'Filters, 2 applied');
    const sort = findFirstByProp(header, 'accessibilityLabel', 'Sort, Recently changed');
    const styles = createBrowseHeaderStyles(lightPalette) as unknown as Record<string, Record<string, unknown>>;

    expect(filters?.props?.badgeCount).toBe(2);
    expect(filters?.props?.iconOnly).toBe(true);
    expect(filters?.props?.label).toBe('Filters');
    expect(filters?.props?.systemImage).toBe('line.3.horizontal.decrease');
    expect(sort).toBeUndefined();
    expect(styles.resultToolsRow).toMatchObject({ alignItems: 'center', minHeight: 44 });
    expect(styles).not.toHaveProperty('toolButton');
    expect(styles).not.toHaveProperty('toolButtonDisabled');
  });

  it('does not repeat inventory context in scrolling content', () => {
    expect(collectText(renderHeader())).not.toContain('Home inventory');
  });

  it('shows removable applied-filter labels and clear all when multiple refinements are active', () => {
    const clears: string[] = [];
    const header = renderHeader({

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

    palette: lightPalette,

    resultCount: 0,
    scope: 'all',
    selectedSurface: 'list',
    selectedTagIds: [],

    sort: 'updated_desc',
    submittedQuery: '',
    onChangeSurface: vi.fn(),

    onClearFilters: vi.fn(),
    onRemoveFilter: vi.fn(),
    onToggleFilters: vi.fn(),
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
