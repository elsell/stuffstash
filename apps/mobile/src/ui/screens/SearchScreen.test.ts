import type { RefObject } from 'react';
import type { TextInput } from 'react-native';
import { describe, expect, it, vi } from 'vitest';
import {
  buildSearchFilterGroupPlacement,
  focusSearchInput
} from './SearchScreenPresentation';
import { SearchHeader } from './SearchScreen';

vi.mock('expo-router', () => ({
  router: { push: vi.fn() },
  useFocusEffect: vi.fn()
}));

vi.mock('react-native-safe-area-context', () => ({
  SafeAreaView: 'SafeAreaView'
}));

vi.mock('react-native', () => ({
  ActivityIndicator: 'ActivityIndicator',
  FlatList: 'FlatList',
  Pressable: 'Pressable',
  StyleSheet: {
    create: (styles: unknown) => styles
  },
  Text: 'Text',
  TextInput: 'TextInput',
  View: 'View'
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

  it('keeps sort as the final refinement group in browse mode', () => {
    expect(buildSearchFilterGroupPlacement('browse')).toEqual([
      { key: 'status', isLast: false },
      { key: 'type', isLast: false },
      { key: 'sort', isLast: true }
    ]);
  });

  it('hides sort and finishes on type in search mode', () => {
    expect(buildSearchFilterGroupPlacement('search')).toEqual([
      { key: 'status', isLast: false },
      { key: 'type', isLast: true }
    ]);
  });

  it('wires the focused search input ref into the rendered header', () => {
    const inputRef = { current: null } as RefObject<TextInput | null>;

    const header = renderHeader({
      query: 'bike pump',
      searchInputFocused: true,
      searchInputRef: inputRef
    });
    const input = findFirstByType(header, 'TextInput');

    expect(input?.props?.ref).toBe(inputRef);
    expect(input?.props?.value).toBe('bike pump');
    expect(input?.props?.style).toEqual(
      expect.arrayContaining([expect.objectContaining({ borderWidth: 2 })])
    );
  });

  it('renders sort in browse mode but not search mode', () => {
    expect(collectText(renderHeader({ resultsMode: 'browse' }))).toContain('Sort');
    expect(collectText(renderHeader({ resultsMode: 'search' }))).not.toContain('Sort');
  });
});

function renderHeader(
  overrides: Partial<Parameters<typeof SearchHeader>[0]> = {}
): ReturnType<typeof SearchHeader> {
  return SearchHeader({
    isLoading: false,
    kind: 'all',
    lifecycleState: 'active',
    query: '',
    resultCount: 0,
    resultsMode: 'browse',
    searchInputFocused: false,
    searchInputRef: { current: null } as RefObject<TextInput | null>,
    sort: 'updated_desc',
    onChangeKind: vi.fn(),
    onChangeLifecycleState: vi.fn(),
    onChangeQuery: vi.fn(),
    onChangeSort: vi.fn(),
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
