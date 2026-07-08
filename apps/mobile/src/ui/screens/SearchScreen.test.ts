import type { RefObject } from 'react';
import type { TextInput } from 'react-native';
import { describe, expect, it, vi } from 'vitest';
import {
  browseScopeToKind,
  buildBrowseScopeOptions,
  focusSearchInput,
  locationRowsFromAssetCards,
  parseBrowseScope,
  searchResultSummaryLabel
} from './SearchScreenPresentation';
import { SearchHeader } from './SearchScreen';

vi.mock('expo-router', () => ({
  router: { push: vi.fn() },
  useFocusEffect: vi.fn()
}));

vi.mock('lucide-react-native', () => ({
  Camera: 'CameraIcon',
  CheckCircle2: 'CheckCircle2Icon',
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
  ActivityIndicator: 'ActivityIndicator',
  AccessibilityInfo: {
    addEventListener: vi.fn(() => ({ remove: vi.fn() })),
    isReduceMotionEnabled: vi.fn(() => Promise.resolve(false))
  },
  FlatList: 'FlatList',
  Image: 'Image',
  Modal: 'Modal',
  Pressable: 'Pressable',
  RefreshControl: 'RefreshControl',
  ScrollView: 'ScrollView',
  StyleSheet: {
    create: (styles: unknown) => styles
  },
  Text: 'Text',
  TextInput: 'TextInput',
  View: 'View',
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
  it('focuses the search input when a current input is available', () => {
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
        photoLabel: 'Needs photo',
        photo: undefined
      },
      {
        id: 'garage',
        title: 'Garage',
        description: 'Tools and seasonal bins',
        containedAssetCountLabel: '8 assets',
        recentAssetLabel: 'Drill, socket set',
        photoLabel: 'Photo ready',
        photo: { uri: 'https://photos/garage.jpg' }
      },
      {
        id: 'attic',
        title: 'Attic',
        description: 'Long-term storage',
        containedAssetCountLabel: 'Contents not summarized',
        recentAssetLabel: 'Home / Attic',
        photoLabel: 'Needs photo',
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
    })).toBe('Showing 4 active containers for "drill" · relevance order');
    expect(searchResultSummaryLabel({
      lifecycleState: 'all',
      query: '',
      resultCount: 2,
      scope: 'places',
      sort: 'id_asc'
    })).toBe('Showing 2 all places · stable order');
  });

  it('renders the Browse header with a search field, scopes, and compact refinements', () => {
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
    expect(text).toEqual(expect.arrayContaining([
      'Browse',
      'List',
      'Map',
      'All',
      'Places',
      'Containers',
      'Items',
      'Active',
      'Archived',
      'Recent'
    ]));
  });

  it('keeps lifecycle and sort chips available for API-backed Places', () => {
    const text = collectText(renderHeader({ scope: 'places' }));

    expect(text).toContain('Places');
    expect(text).toContain('Active');
    expect(text).toContain('Recent');
    expect(text).toContain('Stable');
  });

  it('does not render sort chips while a submitted query is in search mode', () => {
    const text = collectText(renderHeader({
      query: 'mug',
      submittedQuery: 'mug'
    }));

    expect(text).toContain('Active');
    expect(text).not.toContain('Recent');
    expect(text).not.toContain('Stable');
    expect(text).toContain('Showing 0 active things for "mug" · relevance order');
  });

  it('renders colored tag browse filters in the Browse header', () => {
    const selectedTags: string[] = [];
    const header = renderHeader({
      tagFilters: [
        { id: 'tag-tools', key: 'tools', label: 'Tools', color: '#2F80ED' },
        { id: 'tag-camping', key: 'camping', label: 'Camping', color: '#2E7D32' }
      ],
      onSelectTag: (tag) => {
        selectedTags.push(tag.label);
      }
    });

    const text = collectText(header);
    expect(text).toContain('Tags');
    expect(text).toContain('Tools');
    expect(findFirstByProp(header, 'accessibilityLabel', 'Browse by tag')).toBeTruthy();

    const tools = findFirstByProp(header, 'accessibilityLabel', 'Search for tag Tools');
    expect(tools).toBeTruthy();
    const onPress = tools?.props?.onPress;
    if (typeof onPress !== 'function') {
      throw new Error('Missing tag filter press handler');
    }
    onPress();

    expect(selectedTags).toEqual(['Tools']);
  });
});

function renderHeader(
  overrides: Partial<Parameters<typeof SearchHeader>[0]> = {}
): ReturnType<typeof SearchHeader> {
  return SearchHeader({
    isLoading: false,
    lifecycleState: 'active',
    checkoutState: 'any',
    query: '',
    resultCount: 0,
    scope: 'all',
    selectedSurface: 'list',
    searchInputFocused: false,
    searchInputRef: { current: null } as RefObject<TextInput | null>,
    sort: 'updated_desc',
    submittedQuery: '',
    onChangeSurface: vi.fn(),
    onChangeLifecycleState: vi.fn(),
    onChangeCheckoutState: vi.fn(),
    onChangeQuery: vi.fn(),
    onChangeScope: vi.fn(),
    onChangeSort: vi.fn(),
    onClearQuery: vi.fn(),
    onSearchBlur: vi.fn(),
    onSearchFocus: vi.fn(),
    onSubmit: vi.fn(),
    ...overrides
  });
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
