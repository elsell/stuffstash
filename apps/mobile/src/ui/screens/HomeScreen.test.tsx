import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { HomeDashboardViewModel } from '../../application/home/HomeDashboardQuery';
import { HomeScreen } from './HomeScreen';

const testState = vi.hoisted(() => ({
  stateValues: [] as unknown[],
  stateIndex: 0
}));
const routerPush = vi.hoisted(() => vi.fn());

vi.mock('react', () => ({
  useCallback: <T,>(callback: T) => callback,
  useEffect: vi.fn(),
  useRef: <T,>(value: T) => ({ current: value }),
  useState: <T,>(initialValue?: T) => {
    const index = testState.stateIndex++;
    const value = index < testState.stateValues.length ? testState.stateValues[index] : initialValue;
    return [value, vi.fn()];
  }
}));

vi.mock('expo-router', () => ({
  router: {
    navigate: vi.fn(),
    push: routerPush
  },
  useFocusEffect: vi.fn()
}));

vi.mock('lucide-react-native', () => ({
  Settings: 'SettingsIcon'
}));

vi.mock('react-native-safe-area-context', () => ({
  SafeAreaView: 'SafeAreaView'
}));

vi.mock('react-native', () => ({
  ActivityIndicator: 'ActivityIndicator',
  Image: 'Image',
  Modal: 'Modal',
  Pressable: 'Pressable',
  RefreshControl: 'RefreshControl',
  ScrollView: 'ScrollView',
  StyleSheet: { create: (styles: unknown) => styles },
  Text: 'Text',
  TextInput: 'TextInput',
  View: 'View'
}));

vi.mock('../components/AssetCard', () => ({
  AssetCard: (props: unknown) => ({ type: 'AssetCard', props })
}));

vi.mock('../components/BrandMark', () => ({ BrandMark: 'BrandMark' }));
vi.mock('../components/IdentityIcon', () => ({ IdentityLabel: 'IdentityLabel' }));
vi.mock('../feedback/AppFeedback', () => ({
  useAppFeedback: () => ({ showDialog: vi.fn(), showNotice: vi.fn() })
}));
vi.mock('../theme/appearance', () => ({
  useAppearanceAwarePalette: () => ({
    accent: '#6B90AA',
    action: '#0066CC',
    background: '#F7FAFB',
    border: '#C5D0D7',
    onAction: '#FFFFFF',
    surface: '#FFFFFF',
    surfaceMuted: '#E8F0F5',
    text: '#243038',
    textMuted: '#52616B'
  })
}));

const recentAsset = {
  id: 'asset-recent',
  title: 'Recent bowl',
  kindLabel: 'Item',
  customTypeLabel: undefined,
  description: 'Recently changed',
  locationTrailLabel: 'Kitchen / Cabinet',
  parentLocationTrail: [{ id: 'asset-kitchen', title: 'Kitchen', isImmediateParent: true }],
  updatedAtLabel: 'Updated today',
  photoLabel: 'Needs photo',
  imagePlaceholderLabel: 'Item',
  tags: [{ id: 'tag-kitchen', label: 'Kitchen', color: '#2F80ED' }]
} as const;

const checkedOutAsset = {
  ...recentAsset,
  id: 'asset-checked-out',
  title: 'Cordless drill',
  checkedOutLabel: 'Checked out'
} as const;

const dashboard: HomeDashboardViewModel = {
  tenantId: 'tenant-home',
  tenantName: 'Home',
  inventoryId: 'inventory-home',
  inventoryName: 'Home Inventory',
  tenants: [{ id: 'tenant-home', name: 'Home' }],
  inventories: [{
    id: 'inventory-home',
    tenantId: 'tenant-home',
    tenantName: 'Home',
    name: 'Home Inventory',
    roleLabel: 'Owner',
    updatedAtLabel: 'Updated today'
  }],
  canAdd: true,
  topLocations: [],
  locations: [],
  recentAssets: [recentAsset],
  checkedOutAssets: [checkedOutAsset],
  assetTags: []
};

describe('HomeScreen asset cards', () => {
  beforeEach(() => {
    testState.stateIndex = 0;
    testState.stateValues = [{ status: 'ready', dashboard }, false, undefined, undefined];
    routerPush.mockClear();
  });

  it('wires recent and checked-out entries through the shared compact card', () => {
    const cards = renderHomeCards();
    const recent = cards.find((card) => card.props?.asset === recentAsset);
    const checkedOut = cards.find((card) => card.props?.asset === checkedOutAsset);

    expect(recent?.props).toMatchObject({ density: 'compact' });
    expect(recent?.props?.showTags).toBeUndefined();
    expect(checkedOut?.props).toMatchObject({ density: 'compact', showTags: false });
    expect(checkedOut?.props?.footerAction).toMatchObject({ disabled: false, label: 'Return' });

    (recent?.props?.onPress as (() => void) | undefined)?.();
    (recent?.props?.onParentLocationPress as ((location: { id: string }) => void) | undefined)?.({ id: 'asset-kitchen' });

    expect(routerPush).toHaveBeenNthCalledWith(1, {
      pathname: '/assets/[assetId]',
      params: { assetId: 'asset-recent' }
    });
    expect(routerPush).toHaveBeenNthCalledWith(2, {
      pathname: '/assets/[assetId]',
      params: { assetId: 'asset-kitchen' }
    });
  });

  it('connects Return to the checkout command and disables the action while returning', async () => {
    const execute = vi.fn().mockResolvedValue({
      id: 'checkout-one',
      assetId: 'asset-checked-out',
      undoableOperationId: 'operation-one'
    });
    const cards = renderHomeCards(execute);
    const checkedOut = cards.find((card) => card.props?.asset === checkedOutAsset);

    (checkedOut?.props?.footerAction as { onPress?: () => void } | undefined)?.onPress?.();
    await Promise.resolve();

    expect(execute).toHaveBeenCalledWith({ action: 'return', assetId: 'asset-checked-out' });

    testState.stateIndex = 0;
    testState.stateValues = [{ status: 'ready', dashboard }, false, 'asset-checked-out', undefined];
    const returningCard = renderHomeCards(execute).find((card) => card.props?.asset === checkedOutAsset);

    expect(returningCard?.props?.footerAction).toMatchObject({ disabled: true, label: 'Returning...' });
  });
});

type ElementNode = {
  readonly type?: unknown;
  readonly props?: Record<string, unknown>;
};

function renderHomeCards(execute = vi.fn()): readonly ElementNode[] {
  const tree = HomeScreen({
    assetCheckoutCommand: {
      execute,
      undoOperation: vi.fn(),
      updateReturnedCheckoutDetails: vi.fn()
    } as never,
    dashboardQuery: { execute: vi.fn().mockResolvedValue(dashboard) } as never
  });
  return findAllByType(tree, 'AssetCard');
}

function findAllByType(node: unknown, type: unknown): readonly ElementNode[] {
  if (Array.isArray(node)) {
    return node.flatMap((child) => findAllByType(child, type));
  }
  if (!isElementNode(node)) {
    return [];
  }
  if (node.type === type) {
    return [node];
  }
  if (typeof node.type === 'function') {
    return findAllByType(node.type(node.props), type);
  }
  const children = node.props?.children;
  return findAllByType(Array.isArray(children) ? children : [children], type);
}

function isElementNode(node: unknown): node is ElementNode {
  return Boolean(node && typeof node === 'object' && 'props' in node);
}
