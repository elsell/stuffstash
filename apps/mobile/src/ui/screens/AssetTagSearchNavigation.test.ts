import { describe, expect, it, vi } from 'vitest';
import { router } from 'expo-router';
import { InventoryAssetList } from './InventoryAssetsRouteScreen';
import { LocationAssetList } from './LocationAssetsRouteScreen';
import {
  assetTagSearchHref,
  navigateToAssetTagSearch
} from './AssetTagSearchNavigation';

const mocks = vi.hoisted(() => ({
  push: vi.fn(),
  palette: { background: '#111416', text: '#F4F7F8', textMuted: '#B8C4CB', accent: '#8EB3CC' }
}));

vi.mock('expo-router', () => ({
  router: { push: mocks.push },
  Stack: { Screen: 'StackScreen' }
}));

vi.mock('react-native-safe-area-context', () => ({
  SafeAreaView: 'SafeAreaView'
}));

vi.mock('react-native', () => ({
  ActivityIndicator: 'ActivityIndicator',
  FlatList: 'FlatList',
  StyleSheet: {
    create: (styles: unknown) => styles
  },
  Text: 'Text',
  View: 'View'
}));

vi.mock('../components/IdentityIcon', () => ({
  IdentityLabel: 'IdentityLabel'
}));

vi.mock('../components/AssetCard', () => ({
  AssetCard: (props: unknown) => ({
    type: 'AssetCard',
    props
  })
}));

vi.mock('../theme/AppearanceContext', () => ({
  useAppearancePalette: () => mocks.palette
}));

describe('asset tag search navigation', () => {
  it('builds tag-backed search route params for chip navigation', () => {
    expect(assetTagSearchHref({ id: 'tag-camp-kitchen', label: 'Camp Kitchen' })).toEqual({
      pathname: '/search',
      params: { tagId: 'tag-camp-kitchen', tagLabel: 'Camp Kitchen' }
    });
  });

  it('pushes tag-backed search from shared tag chip navigation', () => {
    mocks.push.mockClear();

    navigateToAssetTagSearch(router, { id: 'tag-emergency-kit', label: 'Emergency Kit' });

    expect(router.push).toHaveBeenCalledWith(assetTagSearchHref({ id: 'tag-emergency-kit', label: 'Emergency Kit' }));
  });

  it('wires inventory asset card tag presses to tag-backed search', () => {
    mocks.push.mockClear();
    const list = InventoryAssetList({
      inventoryAssets: {
        inventoryName: 'Household',
        assets: [assetCard('asset-one')]
      },
      isRefreshing: false,
      onRefresh: vi.fn()
    }) as ElementNode;

    const card = renderFirstAssetCard(list);
    (card.props?.onTagPress as (tag: { id: string; label: string }) => void)({ id: 'tag-camp-kitchen', label: 'Camp Kitchen' });

    expect(router.push).toHaveBeenCalledWith(assetTagSearchHref({ id: 'tag-camp-kitchen', label: 'Camp Kitchen' }));
    expect(card.props?.palette).toBe(mocks.palette);
  });

  it('wires location asset card tag presses to tag-backed search', () => {
    mocks.push.mockClear();
    const list = LocationAssetList({
      locationAssets: {
        inventoryName: 'Household',
        locationId: 'location-garage',
        locationTitle: 'Garage',
        assets: [assetCard('asset-two')],
        assetDetails: []
      },
      isRefreshing: false,
      onRefresh: vi.fn()
    }) as ElementNode;

    const card = renderFirstAssetCard(list);
    (card.props?.onTagPress as (tag: { id: string; label: string }) => void)({ id: 'tag-shop-tools', label: 'Shop Tools' });

    expect(router.push).toHaveBeenCalledWith(assetTagSearchHref({ id: 'tag-shop-tools', label: 'Shop Tools' }));
    expect(card.props?.palette).toBe(mocks.palette);
  });
});

function renderFirstAssetCard(tree: ElementNode): ElementNode {
  const flatList = findFirstByType(tree, 'FlatList');
  const renderItem = flatList?.props?.renderItem as ((input: { item: unknown }) => ElementNode) | undefined;
  if (!renderItem) {
    throw new Error('Missing FlatList renderItem');
  }
  const item = (flatList?.props?.data as readonly unknown[] | undefined)?.[0];
  return renderItem({ item });
}

function assetCard(id: string) {
  return {
    id,
    title: 'Camp stove',
    kindLabel: 'Item',
    customTypeLabel: undefined,
    description: 'Cooking kit',
    locationTrailLabel: 'Garage / Camp bin',
    parentLocationTrail: [
      { id: 'asset-garage', title: 'Garage', isImmediateParent: true }
    ],
    updatedAtLabel: 'Updated today',
    photoLabel: 'Photo ready',
    imagePlaceholderLabel: 'Item',
    tags: [{ id: 'tag-camp-kitchen', label: 'Camp Kitchen' }]
  };
}

type ElementNode = {
  readonly type?: unknown;
  readonly props?: {
    readonly children?: unknown;
    readonly [key: string]: unknown;
  };
};

function findFirstByType(node: unknown, type: unknown): ElementNode | undefined {
  if (!node || typeof node !== 'object') {
    return undefined;
  }
  const element = node as ElementNode;
  if (element.type === type) {
    return element;
  }
  const children = element.props?.children;
  if (Array.isArray(children)) {
    for (const child of children) {
      const match = findFirstByType(child, type);
      if (match) {
        return match;
      }
    }
  } else if (children) {
    return findFirstByType(children, type);
  }
  return undefined;
}
