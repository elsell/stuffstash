import { describe, expect, it, vi } from 'vitest';
import {
  BrowseEmptyState,
  BrowseLoadError,
  BrowsePaginationRetry
} from './BrowseResultStates';
import { lightPalette } from '../theme/tokens';

vi.mock('react-native', () => ({
  DynamicColorIOS: undefined,
  Platform: { OS: 'android' },
  Pressable: 'Pressable',
  StyleSheet: {
    create: (styles: unknown) => styles
  },
  Text: 'Text',
  useColorScheme: () => 'light',
  View: 'View'
}));

describe('BrowseEmptyState', () => {
  it('offers Add for an empty inventory', () => {
    const onAdd = vi.fn();
    const state = BrowseEmptyState({
      kind: 'inventory',
      inventoryName: 'Home inventory',
      palette: lightPalette,
      onAdd
    });

    expect(collectText(state)).toEqual(expect.arrayContaining([
      'No items in Home inventory',
      'Add your first item, container, or place.',
      'Add item'
    ]));
    pressButtonWithText(state, 'Add item');
    expect(onAdd).toHaveBeenCalledTimes(1);
  });

  it('explains a read-only empty inventory without exposing Add', () => {
    const state = BrowseEmptyState({
      kind: 'inventory',
      inventoryName: 'Shared inventory',
      palette: lightPalette
    });

    expect(collectText(state)).toEqual(expect.arrayContaining([
      'No items in Shared inventory',
      'An inventory editor can add the first item, container, or place.'
    ]));
    expect(collectText(state)).not.toContain('Add item');
    expect(findPressableWithText(state, 'Add item')).toBeUndefined();
  });

  it('quotes the submitted query and offers Clear search', () => {
    const onClearSearch = vi.fn();
    const state = BrowseEmptyState({
      kind: 'search',
      palette: lightPalette,
      query: 'bike pump',
      onClearSearch
    });

    expect(collectText(state)).toEqual(expect.arrayContaining([
      'No results for “bike pump”',
      'Try another search or clear it to browse everything.',
      'Clear search'
    ]));
    pressButtonWithText(state, 'Clear search');
    expect(onClearSearch).toHaveBeenCalledTimes(1);
  });

  it('offers Clear filters when refinements have no matches', () => {
    const onClearFilters = vi.fn();
    const state = BrowseEmptyState({
      kind: 'filters',
      palette: lightPalette,
      onClearFilters
    });

    expect(collectText(state)).toEqual(expect.arrayContaining([
      'No items match these filters',
      'Remove a filter to see more of your inventory.',
      'Clear filters'
    ]));
    pressButtonWithText(state, 'Clear filters');
    expect(onClearFilters).toHaveBeenCalledTimes(1);
  });
});

describe('BrowseLoadError', () => {
  it('keeps the failure specific and provides a 44-point Retry action', () => {
    const onRetry = vi.fn();
    const state = BrowseLoadError({
      message: 'The server could not be reached.',
      palette: lightPalette,
      onRetry
    });
    const retry = findPressableWithText(state, 'Retry');

    expect(collectText(state)).toEqual(expect.arrayContaining([
      'Could not load this inventory',
      'The server could not be reached.',
      'Retry'
    ]));
    expect(retry?.props?.accessibilityRole).toBe('button');
    expect(styleValue(resolvePressableStyle(retry?.props?.style, false), 'minHeight')).toBeGreaterThanOrEqual(44);
    expect(styleValue(resolvePressableStyle(retry?.props?.style, true), 'backgroundColor')).not.toBe(
      styleValue(resolvePressableStyle(retry?.props?.style, false), 'backgroundColor')
    );

    retry?.props?.onPress?.();
    expect(onRetry).toHaveBeenCalledTimes(1);
  });
});

describe('BrowsePaginationRetry', () => {
  it('preserves a compact footer retry near the loaded results', () => {
    const onRetry = vi.fn();
    const footer = BrowsePaginationRetry({
      message: 'Could not load more items.',
      palette: lightPalette,
      onRetry
    });

    expect(collectText(footer)).toEqual(expect.arrayContaining([
      'Could not load more items.',
      'Try again'
    ]));
    const retry = findPressableWithText(footer, 'Try again');
    expect(styleValue(resolvePressableStyle(retry?.props?.style, false), 'minHeight')).toBeGreaterThanOrEqual(44);
    retry?.props?.onPress?.();
    expect(onRetry).toHaveBeenCalledTimes(1);
  });
});

type TestNode = {
  readonly type?: unknown;
  readonly props?: {
    readonly children?: unknown;
    readonly onPress?: () => void;
    readonly style?: unknown;
    readonly accessibilityRole?: string;
    readonly [key: string]: unknown;
  };
};

function collectText(node: unknown): string[] {
  if (typeof node === 'string') return [node];
  if (Array.isArray(node)) return node.flatMap(collectText);
  if (!node || typeof node !== 'object') return [];
  return collectText((node as TestNode).props?.children);
}

function findPressableWithText(node: unknown, text: string): TestNode | undefined {
  if (!node || typeof node !== 'object') return undefined;
  if (Array.isArray(node)) {
    for (const child of node) {
      const match = findPressableWithText(child, text);
      if (match) return match;
    }
    return undefined;
  }
  const element = node as TestNode;
  if (element.type === 'Pressable' && collectText(element).includes(text)) return element;
  return findPressableWithText(element.props?.children, text);
}

function pressButtonWithText(node: unknown, text: string): void {
  const button = findPressableWithText(node, text);
  if (!button?.props?.onPress) throw new Error(`Missing ${text} action`);
  button.props.onPress();
}

function styleValue(style: unknown, key: string): unknown {
  const styles = Array.isArray(style) ? style : [style];
  return styles.reduce<unknown>((value, entry) => {
    if (entry && typeof entry === 'object' && key in entry) {
      return (entry as Record<string, unknown>)[key];
    }
    return value;
  }, undefined);
}

function resolvePressableStyle(style: unknown, pressed: boolean): unknown {
  return typeof style === 'function'
    ? (style as (state: { readonly pressed: boolean }) => unknown)({ pressed })
    : style;
}
