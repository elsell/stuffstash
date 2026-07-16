import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { HomeDashboardViewModel } from '../../application/home/HomeDashboardQuery';
import { HomeScreen } from './HomeScreen';

const testState = vi.hoisted(() => ({
  stateValues: [] as unknown[],
  stateIndex: 0
}));
const routerPush = vi.hoisted(() => vi.fn());
const useFocusEffectMock = vi.hoisted(() => vi.fn());

vi.mock('react', async (importOriginal) => ({
  ...await importOriginal<typeof import('react')>(),
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
  useFocusEffect: useFocusEffectMock
}));

vi.mock('lucide-react-native', () => ({
  ChevronDown: 'ChevronDownIcon',
  Plus: 'PlusIcon',
  UserCircle: 'UserCircleIcon',
  Settings: 'SettingsIcon'
}));

vi.mock('react-native-safe-area-context', () => ({
  SafeAreaView: 'SafeAreaView'
}));

vi.mock('react-native', () => ({
  ActivityIndicator: 'ActivityIndicator',
  Image: 'Image',
  Modal: 'Modal',
  Platform: { OS: 'ios' },
  Pressable: 'Pressable',
  RefreshControl: 'RefreshControl',
  ScrollView: 'ScrollView',
  StyleSheet: { create: (styles: unknown) => styles },
  Text: 'Text',
  TextInput: 'TextInput',
  View: 'View'
}));

vi.mock('../components/AssetCard', () => ({
  AssetCard: (props: Record<string, unknown>) => ({
    type: 'AssetCard',
    props: {
      ...props,
      children: props.showUpdatedAt
        ? { type: 'Text', props: { children: (props.asset as typeof recentAsset).updatedAtLabel } }
        : undefined
    }
  })
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
  recentAssets: [recentAsset],
  checkedOutAssets: [checkedOutAsset],
  assetTags: []
};

describe('HomeScreen asset cards', () => {
  beforeEach(() => {
    testState.stateIndex = 0;
    testState.stateValues = [{ status: 'ready', dashboard }, false, undefined, undefined];
    routerPush.mockClear();
    useFocusEffectMock.mockClear();
  });

  it('wires recent and checked-out entries through the shared compact card', () => {
    const cards = renderHomeCards();
    const recent = cards.find((card) => card.props?.asset === recentAsset);
    const checkedOut = cards.find((card) => card.props?.asset === checkedOutAsset);

    expect(recent?.props).toMatchObject({ density: 'row', showUpdatedAt: true });
    expect(recent?.props?.showTags).toBeUndefined();
    expect(checkedOut?.props).toMatchObject({ density: 'row', showTags: false });
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

  it('presents compact inventory context and primary toolbar actions with explicit labels', () => {
    const tree = renderReadyHome();
    const context = findByAccessibilityLabel(
      tree,
      'Current inventory Home Inventory, tenant Home. Switch inventory'
    );
    const account = findByAccessibilityLabel(tree, 'Open account and settings');
    const add = findByAccessibilityLabel(tree, 'Add an asset');

    expect(context?.props?.accessibilityRole).toBe('button');
    expect(findTextNode(tree, 'Home Inventory')?.props?.numberOfLines).toBeUndefined();
    expect(findTextNode(tree, 'Home')?.props?.numberOfLines).toBeUndefined();
    expect(account?.props?.accessibilityRole).toBe('button');
    expect(add?.props?.accessibilityRole).toBe('button');
    expect(controlSize(context, 'minHeight')).toBeGreaterThanOrEqual(44);
    expect(controlSize(account, 'minHeight')).toBeGreaterThanOrEqual(44);
    expect(controlSize(account, 'minWidth')).toBeGreaterThanOrEqual(44);
    expect(controlSize(add, 'minHeight')).toBeGreaterThanOrEqual(44);
    expect(controlSize(add, 'minWidth')).toBeGreaterThanOrEqual(44);

    (context?.props?.onPress as (() => void) | undefined)?.();
    (account?.props?.onPress as (() => void) | undefined)?.();
    (add?.props?.onPress as (() => void) | undefined)?.();

    expect(routerPush).toHaveBeenNthCalledWith(1, '/tenant-switcher');
    expect(routerPush).toHaveBeenNthCalledWith(2, '/settings');
    expect(routerPush).toHaveBeenNthCalledWith(3, '/add');
  });

  it('removes the permanent Locations section even when location data is available', () => {
    const tree = renderReadyHome();

    expect(allText(tree)).not.toContain('Locations');
    expect(allText(tree)).not.toContain('Kitchen');
    expect(findAllByType(tree, 'Image')).toHaveLength(0);
  });

  it('exposes recent activity context and descriptive section semantics', () => {
    const tree = renderReadyHome();
    const cards = findAllByType(tree, 'AssetCard');

    expect(allText(tree)).toContain('Updated today');
    expect(cards.find((card) => card.props?.asset === recentAsset)?.props?.showUpdatedAt).toBe(true);
    expect(findTextNode(tree, 'Recently changed')?.props?.accessibilityRole).toBe('header');
    expect(findTextNode(tree, 'Checked out')?.props?.accessibilityRole).toBe('header');
    expect(findByAccessibilityLabel(tree, 'View all recently changed assets')).toBeDefined();
    expect(findByAccessibilityLabel(tree, 'View all checked-out assets')).toBeDefined();
    const scrollViews = findAllByType(tree, 'ScrollView');
    expect(scrollViews.filter((node) => node.props?.horizontal === true)).toHaveLength(0);
    expect(scrollViews[0]?.props?.contentInsetAdjustmentBehavior).toBe('automatic');
  });

  it('omits the checked-out section entirely when no assets need attention', () => {
    const tree = renderReadyHome({ ...dashboard, checkedOutAssets: [] });

    expect(allText(tree)).not.toContain('Checked out');
    expect(allText(tree)).not.toContain('Nothing checked out.');
    expect(findByAccessibilityLabel(tree, 'View all checked-out assets')).toBeUndefined();
  });

  it('keeps checked-out assets compact and gives each Return action asset-specific context', () => {
    const checkedOut = findAllByType(renderReadyHome(), 'AssetCard')
      .find((card) => card.props?.asset === checkedOutAsset);

    expect(checkedOut?.props).toMatchObject({ density: 'row', showTags: false });
    expect(checkedOut?.props?.footerAction).toMatchObject({
      accessibilityLabel: 'Return Cordless drill',
      disabled: false,
      label: 'Return'
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

describe('HomeScreen recovery', () => {
  it('offers an explicit Retry action after the initial load fails', async () => {
    testState.stateIndex = 0;
    testState.stateValues = [{ status: 'error', message: 'Network unavailable' }, false];
    const execute = vi.fn().mockResolvedValue(dashboard);
    const tree = renderHome(execute);
    const retry = findByAccessibilityLabel(tree, 'Retry loading Home');

    expect(retry?.props?.accessibilityRole).toBe('button');
    (retry?.props?.onPress as (() => void) | undefined)?.();
    await Promise.resolve();

    expect(execute).toHaveBeenCalledTimes(1);

    const focusRefresh = useFocusEffectMock.mock.calls.at(-1)?.[0] as (() => void) | undefined;
    focusRefresh?.();
    await Promise.resolve();

    expect(execute).toHaveBeenCalledTimes(2);
  });
});

type ElementNode = {
  readonly type?: unknown;
  readonly props?: Record<string, unknown>;
};

function renderHomeCards(execute = vi.fn()): readonly ElementNode[] {
  const tree = renderHome(undefined, execute);
  return findAllByType(tree, 'AssetCard');
}

function renderReadyHome(readyDashboard: HomeDashboardViewModel = dashboard): unknown {
  testState.stateIndex = 0;
  testState.stateValues = [{ status: 'ready', dashboard: readyDashboard }, false, undefined, undefined];
  return renderHome();
}

function renderHome(dashboardExecute?: ReturnType<typeof vi.fn>, checkoutExecute = vi.fn()): unknown {
  return HomeScreen({
    assetCheckoutCommand: {
      execute: checkoutExecute,
      undoOperation: vi.fn(),
      updateReturnedCheckoutDetails: vi.fn()
    } as never,
    dashboardQuery: { execute: dashboardExecute ?? vi.fn().mockResolvedValue(dashboard) } as never
  });
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

function findByAccessibilityLabel(node: unknown, label: string): ElementNode | undefined {
  return findAll(node).find((candidate) => candidate.props?.accessibilityLabel === label);
}

function findTextNode(node: unknown, value: string): ElementNode | undefined {
  return findAllByType(node, 'Text').find((candidate) => candidate.props?.children === value);
}

function allText(node: unknown): readonly string[] {
  return findAllByType(node, 'Text')
    .flatMap((candidate) => typeof candidate.props?.children === 'string' ? [candidate.props.children] : []);
}

function findAll(node: unknown): readonly ElementNode[] {
  if (Array.isArray(node)) {
    return node.flatMap(findAll);
  }
  if (!isElementNode(node)) {
    return [];
  }
  if (typeof node.type === 'function') {
    return [node, ...findAll(node.type(node.props))];
  }
  const children = node.props?.children;
  return [node, ...findAll(Array.isArray(children) ? children : [children])];
}

function controlSize(node: ElementNode | undefined, property: 'minHeight' | 'minWidth'): number | undefined {
  const styles = flattenStyles(node?.props?.style);
  const value = styles[property];
  return typeof value === 'number' ? value : undefined;
}

function flattenStyles(style: unknown): Record<string, unknown> {
  if (Array.isArray(style)) {
    return Object.assign({}, ...style.map(flattenStyles));
  }
  return style && typeof style === 'object' ? style as Record<string, unknown> : {};
}
