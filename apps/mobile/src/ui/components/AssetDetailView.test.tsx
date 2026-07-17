import { describe, expect, it, vi } from 'vitest';
import { createElement } from 'react';
import type { AssetCardViewModel, AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import {
  AssetDetailView,
  assetDetailNavigationTitle,
  containedAssetRowAccessibilityLabel,
  containedWorkspaceItems
} from './AssetDetailView';
import { AppTextInput } from './AppTextInput';

vi.mock('react', () => ({
  useState: <Value,>(initial: Value) => [initial, vi.fn()]
}));

vi.mock('react', async (importOriginal) => ({
  ...await importOriginal<typeof import('react')>(),
  useState: <T,>(initial: T) => [initial, vi.fn()] as const
}));

vi.mock('lucide-react-native', () => ({
  Camera: 'CameraIcon',
  CheckCircle2: 'CheckCircle2Icon',
  ChevronRight: 'ChevronRightIcon',
  MoreHorizontal: 'MoreHorizontalIcon',
  MoveRight: 'MoveRightIcon',
  Pencil: 'PencilIcon',
  Plus: 'PlusIcon'
}));

vi.mock('../theme/appearance', () => ({
  appearanceAwarePalette: () => ({
    accent: '#6B90AA',
    accentStrong: '#303A41',
    action: '#0066CC',
    background: '#F7FAFB',
    border: '#C5D0D7',
    onAction: '#FFFFFF',
    onScrim: '#FFFFFF',
    surface: '#FFFFFF',
    surfaceMuted: '#E8F0F5',
    text: '#243038',
    textMuted: '#52616B'
  }),
  useAppearanceAwarePalette: () => ({
    accent: '#6B90AA',
    accentStrong: '#303A41',
    action: '#0066CC',
    background: '#F7FAFB',
    border: '#C5D0D7',
    onAction: '#FFFFFF',
    onScrim: '#FFFFFF',
    surface: '#FFFFFF',
    surfaceMuted: '#E8F0F5',
    text: '#243038',
    textMuted: '#52616B'
  })
}));

vi.mock('../theme/AppearanceContext', () => ({
  useAppearancePalette: () => ({
    accent: '#6B90AA',
    accentStrong: '#303A41',
    action: '#0066CC',
    actionPressed: '#004F9F',
    border: '#C5D0D7',
    controlBorder: '#6F7E88',
    onAction: '#FFFFFF',
    selected: '#E8F0F5',
    surface: '#FFFFFF',
    surfaceMuted: '#E8F0F5',
    text: '#243038',
    textMuted: '#52616B',
    warning: '#8A4F00',
    warningSurface: '#FFF3DF'
  })
}));

vi.mock('react-native', () => ({
  DynamicColorIOS: ({ light }: { light: string }) => light,
  FlatList: (props: Record<string, unknown>) => {
    const data = (props.data as readonly unknown[] | undefined) ?? [];
    const renderItem = props.renderItem as ((input: { item: unknown; index: number }) => unknown) | undefined;
    return {
      type: 'View',
      props: {
        children: [
          props.ListHeaderComponent,
          ...data.map((item, index) => renderItem?.({ item, index })),
          data.length === 0 ? props.ListEmptyComponent : undefined,
          props.ListFooterComponent
        ]
      }
    };
  },
  Image: 'Image',
  Platform: { OS: 'ios' },
  Pressable: 'Pressable',
  ScrollView: 'ScrollView',
  StyleSheet: {
    create: (styles: unknown) => styles,
    hairlineWidth: 1
  },
  Text: 'Text',
  TextInput: 'TextInput',
  View: 'View',
  useColorScheme: () => 'light',
  useWindowDimensions: () => ({ width: 390, height: 844, scale: 3, fontScale: 1 })
}));

describe('AssetDetailView', () => {
  it('makes detail tag chips search actions', () => {
    const searchedTags: string[] = [];
    const tree = AssetDetailView({
      asset: assetDetail(),
      onTagPress: (tag) => {
        searchedTags.push(tag.label);
      }
    });

    const tagChip = findFirstByProp(tree, 'accessibilityLabel', 'Search for tag Camping');
    expect(tagChip?.props?.accessibilityRole).toBe('button');
    press(tagChip);

    expect(searchedTags).toEqual(['Camping']);
  });

  it('presents the asset title before classification and placement before description', () => {
    const tree = AssetDetailView({ asset: assetDetail() });
    const text = collectText(tree);
    const identityHeading = findFirstByProp(tree, 'accessibilityRole', 'header');
    const placementIndex = text.findIndex((value) => value.includes('Garage'));

    expect(collectText(identityHeading)).toContain('Family tent');
    expect(text.indexOf('Family tent')).toBeLessThan(placementIndex);
    expect(placementIndex).toBeLessThan(text.indexOf('Sleeps four.'));
  });

  it('shows only the applicable availability action and hides unavailable maintenance actions', () => {
    const available = collectText(AssetDetailView({
      asset: assetDetail(),
      onCheckout: vi.fn(),
      onEdit: vi.fn(),
      onMove: vi.fn(),
      onAddPhotos: vi.fn()
    }));
    const checkedOut = collectText(AssetDetailView({
      asset: {
        ...assetDetail(),
        isCheckedOut: true,
        checkoutLabel: 'Checked out Jul 14, 2026',
        canCheckout: false,
        canReturn: true
      },
      onCheckout: vi.fn(),
      onReturn: vi.fn(),
      onEdit: vi.fn(),
      onMove: vi.fn(),
      onAddPhotos: vi.fn()
    }));
    const readOnly = collectText(AssetDetailView({
      asset: {
        ...assetDetail(),
        canCheckout: false,
        canEdit: false,
        canMove: false,
        canAddPhotos: false
      },
      onCheckout: vi.fn(),
      onEdit: vi.fn(),
      onMove: vi.fn(),
      onAddPhotos: vi.fn()
    }));

    expect(available).toContain('Check out');
    expect(available).not.toContain('Return');
    expect(checkedOut).toContain('Return');
    expect(checkedOut).not.toContain('Check out');
    expect(readOnly).not.toContain('Check out');
    expect(readOnly).not.toContain('Return');
    expect(readOnly).not.toContain('Edit');
    expect(readOnly).not.toContain('Move');
  });

  it('opens structured placement breadcrumbs using parent asset identity', () => {
    const openedParents: string[] = [];
    const tree = AssetDetailView({
      asset: assetDetail(),
      onParentLocationPress: (parent) => openedParents.push(parent.id)
    });

    const garage = findFirstByProp(tree, 'accessibilityLabel', 'Open location Garage');
    const campBin = findFirstByProp(tree, 'accessibilityLabel', 'Open location Camp / bin');
    press(campBin);

    expect(garage?.props?.accessibilityRole).toBe('button');
    expect(campBin?.props?.accessibilityRole).toBe('button');
    expect(openedParents).toEqual(['asset-camp-bin']);
  });

  it('uses a calm root placement label when there is no parent trail', () => {
    const text = collectText(AssetDetailView({
      asset: {
        ...assetDetail(),
        locationTrailLabel: 'Household',
        parentLocationTrailLabel: 'Inventory root',
        parentLocationTrail: []
      }
    }));

    expect(text).toContain('No location');
    expect(text).not.toContain('Inventory root');
  });

  it('omits synthetic placement for a root place and keeps one gallery-level photo action', () => {
    const tree = AssetDetailView({
      asset: {
        ...assetDetail(),
        title: 'Garage',
        kind: 'location',
        kindLabel: 'Place',
        canContainAssets: true,
        canAddContainedAssets: true,
        locationTrailLabel: 'Garage',
        parentLocationTrailLabel: 'Inventory root',
        parentLocationTrail: [],
        imagePlaceholderLabel: 'Place'
      },
      onAddHere: vi.fn(),
      onAddPhotos: vi.fn(),
      onEdit: vi.fn(),
      onMove: vi.fn(),
      onMoveThingsHere: vi.fn()
    });
    const text = collectText(tree);

    expect(text.filter((value) => value === 'Add photos')).toHaveLength(1);
    expect(text.indexOf('Add photos')).toBeLessThan(text.indexOf('Garage'));
    expect(text).not.toContain('No location');
    expect(text).not.toContain('Inventory root');
    expect(text).toContain('Move place');
    expect(text).not.toContain('Move');
    expect(styleValue(findFirstByProp(tree, 'accessibilityLabel', 'No photos')?.props?.style, 'width')).toBe(358);
  });

  it('uses place route language only for locations', () => {
    expect(assetDetailNavigationTitle({ kind: 'location' })).toBe('Place');
    expect(assetDetailNavigationTitle({ kind: 'container' })).toBe('Details');
    expect(assetDetailNavigationTitle({ kind: 'item' })).toBe('Details');
  });

  it('keeps place section headings adjacent to title-first rows with relative paths', () => {
    const tree = AssetDetailView({
      asset: placeDetail({
        spaces: [containedCard('space-shelf', 'Utility shelf', 'container')],
        items: [{
          ...containedCard('item-drill', 'Cordless drill', 'item'),
          relativePath: [{ id: 'space-shelf', title: 'Utility shelf' }],
          relativePathLabel: 'Utility shelf'
        }]
      }),
      onAddHere: vi.fn(),
      onChildPress: vi.fn(),
      onMoveThingsHere: vi.fn()
    });
    const text = collectText(tree);

    expect(text.indexOf('Add item here')).toBeLessThan(text.indexOf('Spaces in Garage'));
    expect(text.indexOf('Spaces in Garage')).toBeLessThan(text.indexOf('Utility shelf'));
    expect(text.indexOf('Items in Garage')).toBeLessThan(text.indexOf('Cordless drill'));
    expect(text.indexOf('Cordless drill')).toBeLessThan(text.lastIndexOf('Item'));
    expect(text.lastIndexOf('Item')).toBeLessThan(text.lastIndexOf('Utility shelf'));
    expect(text.indexOf('Move place')).toBeGreaterThan(text.indexOf('Cordless drill'));
    expect(findFirstByProp(
      tree,
      'accessibilityLabel',
      'Open asset Cordless drill. Item. Utility shelf'
    )?.props?.accessibilityRole).toBe('button');
  });

  it('builds distinguishing contained-row labels from visible kind, path, and checkout context', () => {
    expect(containedAssetRowAccessibilityLabel({
      id: 'asset-bin',
      title: 'Storage bin',
      eyebrowLabel: 'Container · Holiday storage',
      supportingLabel: 'Checked out · Attic',
      imagePlaceholderLabel: 'Container'
    })).toBe('Open asset Storage bin. Container · Holiday storage. Checked out · Attic');
  });

  it('filters large place contents by title and relative path while preserving sections and recovery', () => {
    const asset = placeDetail({
      spaces: [containedCard('space-shelf', 'Utility shelf', 'container')],
      items: [{
        ...containedCard('item-drill', 'Cordless drill', 'item'),
        relativePath: [{ id: 'space-shelf', title: 'Utility shelf' }],
        relativePathLabel: 'Utility shelf'
      }]
    });
    const pathMatches = containedWorkspaceItems(asset, 'utility');
    const noMatches = containedWorkspaceItems(asset, 'freezer');

    expect(pathMatches.filter((item) => item.kind === 'section').map((item) => item.heading.summary))
      .toEqual(['1 space', '1 item']);
    expect(pathMatches.filter((item) => item.kind === 'row').map((item) => item.row.title))
      .toEqual(['Utility shelf', 'Cordless drill']);
    expect(noMatches.filter((item) => item.kind === 'section').map((item) => item.heading.summary))
      .toEqual(['0 of 1 space', '0 of 1 item']);
    expect(noMatches.some((item) => item.kind === 'empty' && item.canClearSearch)).toBe(true);
  });

  it('reserves inline contents search for places with at least twenty rows', () => {
    const twentySpaces = Array.from({ length: 20 }, (_, index) => (
      containedCard(`space-${index.toString()}`, `Shelf ${index.toString()}`, 'container')
    ));
    const large = AssetDetailView({ asset: placeDetail({ spaces: twentySpaces }), onAddHere: vi.fn() });
    const small = AssetDetailView({ asset: placeDetail({ spaces: twentySpaces.slice(0, 19) }), onAddHere: vi.fn() });

    expect(findFirstByProp(large, 'accessibilityLabel', 'Search contents')?.type).toBe(AppTextInput);
    expect(large.props.keyboardDismissMode).toBe('interactive');
    expect(large.props.keyboardShouldPersistTaps).toBe('handled');
    expect(findFirstByProp(small, 'accessibilityLabel', 'Search contents')).toBeUndefined();
  });

  it('omits empty maintenance chrome for viewer and archived places with no applicable actions', () => {
    const viewer = AssetDetailView({
      asset: {
        ...placeDetail(),
        canEdit: false,
        canMove: false,
        canCheckout: false,
        canReturn: false
      }
    });
    const archived = AssetDetailView({
      asset: {
        ...placeDetail(),
        isActive: false,
        canEdit: false,
        canMove: false,
        canCheckout: false,
        canReturn: false
      }
    });

    expect(findFirstByProp(viewer, 'accessibilityLabel', 'Manage this asset')).toBeUndefined();
    expect(findFirstByProp(archived, 'accessibilityLabel', 'Manage this asset')).toBeUndefined();
  });

  it('lets shared detail text and action rows grow with accessibility Dynamic Type', () => {
    const tree = AssetDetailView({
      asset: placeDetail({
        spaces: [containedCard('space-shelf', 'Utility shelf', 'container')],
        items: [{
          ...containedCard('item-drill', 'Cordless drill', 'item'),
          relativePath: [{ id: 'space-shelf', title: 'Utility shelf' }],
          relativePathLabel: 'Utility shelf'
        }]
      }),
      onAddHere: vi.fn(),
      onEdit: vi.fn(),
      onMove: vi.fn(),
      onMoveThingsHere: vi.fn()
    });

    for (const label of ['Garage', 'Add item here', 'Spaces in Garage', 'Cordless drill', 'Move place']) {
      expect(styleValue(findFirstTextNode(tree, label)?.props?.style, 'lineHeight')).toBeUndefined();
    }
    expect(styleValue(findFirstByProp(tree, 'accessibilityLabel', 'Asset maintenance')?.props?.style, 'flexWrap'))
      .toBe('wrap');
    expect(styleValue(findFirstByProp(tree, 'accessibilityLabel', 'Add item here')?.props?.style, 'paddingVertical'))
      .toBeGreaterThanOrEqual(10);
    expect(findFirstTextNode(tree, 'Container')?.props?.allowFontScaling).toBe(false);
  });

  it('puts the primary spatial action before quieter container utility actions', () => {
    const text = collectText(AssetDetailView({
      asset: {
        ...assetDetail(),
        kind: 'container',
        kindLabel: 'Container',
        canContainAssets: true,
        canAddContainedAssets: true
      },
      onAddHere: vi.fn(),
      onAddPhotos: vi.fn(),
      onCheckout: vi.fn(),
      onEdit: vi.fn(),
      onMove: vi.fn(),
      onMoveThingsHere: vi.fn()
    }));

    const addHereIndex = text.indexOf('Add item here');
    expect(addHereIndex).toBeGreaterThan(-1);
    expect(addHereIndex).toBeLessThan(text.indexOf('Check out'));
    expect(addHereIndex).toBeLessThan(text.indexOf('Edit'));
    expect(text.indexOf('Add photos')).toBeLessThan(addHereIndex);
    expect(text.filter((value) => value === 'Add photos')).toHaveLength(1);
  });

  it('keeps overflow actions reachable when detail is embedded without a native header', () => {
    const overflowMenu = createElement('NativeActionMenu', { accessibilityLabel: 'More actions' });
    const tree = AssetDetailView({ asset: assetDetail(), overflowMenu });

    const moreButton = findFirstByProp(tree, 'accessibilityLabel', 'More actions');
    expect(moreButton?.type).toBe('NativeActionMenu');
  });

  it('keeps callback-less maintenance actions disabled when they cannot run', () => {
    const callbackLessTree = AssetDetailView({ asset: assetDetail() });

    const editButton = findFirstByProp(callbackLessTree, 'accessibilityLabel', 'Edit');
    const moveButton = findFirstByProp(callbackLessTree, 'accessibilityLabel', 'Move');

    expect(editButton?.props?.disabled).toBe(true);
    expect(editButton?.props?.accessibilityState).toEqual({ disabled: true });
    expect(moveButton?.props?.disabled).toBe(true);
    expect(moveButton?.props?.accessibilityState).toEqual({ disabled: true });
  });
});

function assetDetail(): AssetDetailViewModel {
  return {
    id: 'asset-tent',
    title: 'Family tent',
    kind: 'item',
    kindLabel: 'Item',
    description: 'Sleeps four.',
    locationTrailLabel: 'Garage / Camp bin',
    parentLocationTrailLabel: 'Garage / Camp bin',
    parentLocationTrail: [
      { id: 'asset-garage', title: 'Garage', isImmediateParent: false },
      { id: 'asset-camp-bin', title: 'Camp / bin', isImmediateParent: true }
    ],
    lifecycleLabel: 'Active',
    isActive: true,
    canEdit: true,
    canMove: true,
    canAddPhotos: true,
    canArchive: true,
    canRestore: false,
    canDeletePermanently: false,
    isCheckedOut: false,
    checkoutLabel: 'Available',
    tags: [{ id: 'tag-camping', label: 'Camping', color: '#2F80ED' }],
    canCheckout: true,
    canReturn: false,
    containedAssets: [],
    containedAssetsLabel: '0 things inside',
    containedSpaces: [],
    containedSpacesLabel: '0 spaces',
    containedItems: [],
    containedItemsLabel: '0 items',
    canContainAssets: false,
    canAddContainedAssets: false,
    updatedAtLabel: 'Updated today',
    photoLabel: 'Needs photo',
    imagePlaceholderLabel: 'Item',
    photos: []
  };
}

function placeDetail({
  items = [],
  spaces = []
}: {
  readonly items?: AssetDetailViewModel['containedItems'];
  readonly spaces?: AssetDetailViewModel['containedSpaces'];
} = {}): AssetDetailViewModel {
  return {
    ...assetDetail(),
    title: 'Garage',
    kind: 'location',
    kindLabel: 'Place',
    parentLocationTrail: [],
    parentLocationTrailLabel: 'Inventory root',
    containedAssets: spaces,
    containedAssetsLabel: `${spaces.length.toString()} things inside`,
    containedSpaces: spaces,
    containedSpacesLabel: `${spaces.length.toString()} ${spaces.length === 1 ? 'space' : 'spaces'}`,
    containedItems: items,
    containedItemsLabel: `${items.length.toString()} ${items.length === 1 ? 'item' : 'items'}`,
    canContainAssets: true,
    canAddContainedAssets: true,
    canCheckout: false,
    imagePlaceholderLabel: 'Place'
  };
}

function containedCard(
  id: string,
  title: string,
  kind: 'item' | 'container' | 'location'
): AssetCardViewModel {
  return {
    id,
    title,
    kindLabel: kind === 'location' ? 'Place' : kind === 'container' ? 'Container' : 'Item',
    description: '',
    locationTrailLabel: 'Garage',
    parentLocationTrail: [],
    updatedAtLabel: 'Updated today',
    photoLabel: 'Needs photo',
    imagePlaceholderLabel: kind === 'location' ? 'Place' : kind === 'container' ? 'Container' : 'Item'
  };
}

function collectText(node: unknown): string[] {
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

function styleValue(style: unknown, key: string): unknown {
  if (typeof style === 'function') {
    return styleValue(style({ pressed: false }), key);
  }
  if (Array.isArray(style)) {
    return style.reduce<unknown>((found, entry) => found ?? styleValue(entry, key), undefined);
  }
  return style && typeof style === 'object' ? (style as Record<string, unknown>)[key] : undefined;
}

function press(node: ElementNode | undefined): void {
  const onPress = node?.props?.onPress;
  if (typeof onPress !== 'function') {
    throw new Error('Missing press handler');
  }
  onPress();
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

function findFirstTextNode(node: unknown, value: string): ElementNode | undefined {
  if (Array.isArray(node)) {
    return node.reduce<ElementNode | undefined>(
      (found, child) => found ?? findFirstTextNode(child, value),
      undefined
    );
  }
  if (!isElementNode(node)) {
    return undefined;
  }
  if (node.type === 'Text' && childrenOf(node).includes(value)) {
    return node;
  }
  if (typeof node.type === 'function') {
    return findFirstTextNode(node.type(node.props), value);
  }
  return childrenOf(node).reduce<ElementNode | undefined>(
    (found, child) => found ?? findFirstTextNode(child, value),
    undefined
  );
}

function childrenOf(node: ElementNode): readonly unknown[] {
  const children = node.props?.children;
  return Array.isArray(children) ? children : [children];
}

function isElementNode(node: unknown): node is ElementNode {
  return Boolean(node && typeof node === 'object' && 'props' in node);
}

type ElementNode = {
  readonly type?: unknown;
  readonly props?: {
    readonly children?: unknown;
    readonly [key: string]: unknown;
  };
};
