import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  InventoryInvitationEmailMismatchError,
  type InventoryInvitationPreview,
  type InventoryInvitationReference
} from '../../application/invitations/InventoryInvitationRepository';
import { InventoryInvitationScreen } from './InventoryInvitationScreen';

const harness = vi.hoisted(() => ({
  effects: [] as Array<() => void | (() => void)>,
  fontScale: 1,
  hookIndex: 0,
  values: [] as unknown[]
}));

vi.mock('react', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react')>()),
  useEffect: (effect: () => void | (() => void)) => harness.effects.push(effect),
  useRef: <T,>(initial: T) => {
    const index = harness.hookIndex++;
    if (index >= harness.values.length) harness.values[index] = { current: initial };
    return harness.values[index] as { current: T };
  },
  useState: <T,>(initial: T) => {
    const index = harness.hookIndex++;
    if (index >= harness.values.length) harness.values[index] = initial;
    const setValue = (next: T | ((current: T) => T)) => {
      const current = harness.values[index] as T;
      harness.values[index] = typeof next === 'function'
        ? (next as (value: T) => T)(current)
        : next;
    };
    return [harness.values[index] as T, setValue] as const;
  }
}));

vi.mock('react-native', () => ({
  ActivityIndicator: 'ActivityIndicator',
  Pressable: 'Pressable',
  ScrollView: 'ScrollView',
  StyleSheet: { create: <T,>(styles: T) => styles },
  Text: 'Text',
  View: 'View',
  useWindowDimensions: () => ({ fontScale: harness.fontScale, height: 844, width: 390 })
}));

vi.mock('lucide-react-native', () => ({
  CheckCircle2: 'CheckCircleIcon',
  MailCheck: 'MailCheckIcon'
}));

vi.mock('../components/BrandMark', () => ({ BrandMark: 'BrandMark' }));
vi.mock('../theme/appearance', () => ({ useAppearanceAwarePalette: () => palette }));

const referenceA: InventoryInvitationReference = {
  tenantId: 'tenant-a', inventoryId: 'inventory-a', invitationId: 'invitation-a', acceptanceToken: 'A'.repeat(43)
};
const referenceB: InventoryInvitationReference = {
  tenantId: 'tenant-b', inventoryId: 'inventory-b', invitationId: 'invitation-b', acceptanceToken: 'B'.repeat(43)
};

describe('InventoryInvitationScreen behavior', () => {
  beforeEach(() => {
    harness.effects = [];
    harness.fontScale = 1;
    harness.hookIndex = 0;
    harness.values = [];
  });

  it('previews, explicitly accepts, and opens the accepted inventory', async () => {
    const accept = vi.fn().mockResolvedValue({ inventoryId: 'inventory-a', relationship: 'viewer' });
    const onAccepted = vi.fn().mockResolvedValue(undefined);
    const props = screenProps({
      acceptCommand: { execute: accept },
      onAccepted,
      previewQuery: { execute: vi.fn().mockResolvedValue(preview('Kitchen inventory')) }
    });

    let tree = render(props);
    await runEffects();
    tree = render(props);

    press(findTextButton(tree, 'Accept invitation'));
    expect(accept).toHaveBeenCalledWith(referenceA);
    await flushPromises();
    tree = render(props);
    await press(findTextButton(tree, 'Open inventory'));

    expect(onAccepted).toHaveBeenCalledWith('inventory-a');
  });

  it('ignores a late preview after a newer invitation becomes current', async () => {
    const older = deferred<InventoryInvitationPreview>();
    const newer = deferred<InventoryInvitationPreview>();
    const execute = vi.fn((reference: InventoryInvitationReference) =>
      reference.invitationId === referenceA.invitationId ? older.promise : newer.promise
    );
    let props = screenProps({ previewQuery: { execute } });

    render(props);
    const cleanup = await runEffects();
    cleanup();
    props = screenProps({ previewQuery: { execute }, reference: referenceB });
    render(props);
    await runEffects();
    newer.resolve(preview('Garage inventory', { inventoryId: 'inventory-b' }));
    await flushPromises();
    older.resolve(preview('Old inventory'));
    await flushPromises();

    const tree = render(props);
    expect(findText(tree, 'Join Garage inventory')).toBeDefined();
    expect(findText(tree, 'Join Old inventory')).toBeUndefined();
  });

  it('ignores a late acceptance after a newer invitation becomes current', async () => {
    const acceptance = deferred<{
      tenantId: string;
      inventoryId: string;
      invitationId: string;
      principalId: string;
      relationship: 'viewer';
      status: 'accepted';
    }>();
    const previewQuery = {
      execute: vi.fn((reference: InventoryInvitationReference) => Promise.resolve(
        reference.invitationId === referenceA.invitationId
          ? preview('Kitchen inventory')
          : preview('Garage inventory', { inventoryId: 'inventory-b' })
      ))
    };
    let props = screenProps({ acceptCommand: { execute: vi.fn(() => acceptance.promise) }, previewQuery });

    render(props);
    const cleanup = await runEffects();
    let tree = render(props);
    press(findTextButton(tree, 'Accept invitation'));

    cleanup();
    props = screenProps({ previewQuery, reference: referenceB });
    render(props);
    await runEffects();
    acceptance.resolve({
      inventoryId: 'inventory-a',
      invitationId: 'invitation-a',
      principalId: 'principal-one',
      relationship: 'viewer',
      status: 'accepted',
      tenantId: 'tenant-a'
    });
    await flushPromises();

    tree = render(props);
    expect(findText(tree, 'Join Garage inventory')).toBeDefined();
    expect(findText(tree, 'You’re in')).toBeUndefined();
  });

  it('offers an account switch for a mismatched signed-in identity', async () => {
    const onSwitchAccount = vi.fn();
    const props = screenProps({
      onSwitchAccount,
      previewQuery: { execute: vi.fn().mockRejectedValue(new InventoryInvitationEmailMismatchError()) }
    });

    render(props);
    await runEffects();
    const tree = render(props);
    press(findTextButton(tree, 'Switch account'));

    expect(onSwitchAccount).toHaveBeenCalledOnce();
  });

  it.each([
    ['expired', true, 'Invitation expired'],
    ['revoked', false, 'Invitation revoked'],
    ['cancelled', false, 'Invitation cancelled']
  ] as const)('renders the %s terminal state without an acceptance action', async (status, isExpired, title) => {
    const props = screenProps({
      previewQuery: { execute: vi.fn().mockResolvedValue(preview('Kitchen inventory', { isExpired, status })) }
    });

    render(props);
    await runEffects();
    const tree = render(props);

    expect(findText(tree, title)).toBeDefined();
    expect(findTextButton(tree, 'Accept invitation')).toBeUndefined();
  });

  it('renders an already accepted invitation as success', async () => {
    const props = screenProps({
      previewQuery: { execute: vi.fn().mockResolvedValue(preview('Kitchen inventory', { status: 'accepted' })) }
    });

    render(props);
    await runEffects();
    const tree = render(props);

    expect(findText(tree, 'You’re in')).toBeDefined();
    expect(findTextButton(tree, 'Open inventory')).toBeDefined();
  });

  it('uses an adaptive scroll layout and does not cap accessibility Dynamic Type', async () => {
    harness.fontScale = 3;
    const props = screenProps({ previewQuery: { execute: vi.fn().mockResolvedValue(preview('Kitchen inventory')) } });

    render(props);
    await runEffects();
    const tree = render(props);
    const scrollView = findByType(tree, 'ScrollView');

    expect(styleValue(scrollView?.props?.contentContainerStyle, 'justifyContent')).toBe('flex-start');
    expect(findAllByType(tree, 'Text').every((node) => node.props?.maxFontSizeMultiplier === undefined)).toBe(true);
  });
});

function screenProps(overrides: Partial<Parameters<typeof InventoryInvitationScreen>[0]> = {}): Parameters<typeof InventoryInvitationScreen>[0] {
  return {
    acceptCommand: { execute: vi.fn() },
    initialized: true,
    invalidLink: false,
    onAccepted: vi.fn().mockResolvedValue(undefined),
    onDismiss: vi.fn(),
    onSwitchAccount: vi.fn(),
    previewQuery: { execute: vi.fn().mockResolvedValue(preview('Kitchen inventory')) },
    reference: referenceA,
    ...overrides
  };
}

function preview(inventoryName: string, overrides: Partial<InventoryInvitationPreview> = {}): InventoryInvitationPreview {
  return {
    expiresAt: '2030-01-02T15:04:05Z',
    inventoryId: 'inventory-a',
    inventoryName,
    isExpired: false,
    relationship: 'viewer',
    status: 'pending',
    ...overrides
  };
}

function render(props: Parameters<typeof InventoryInvitationScreen>[0]): unknown {
  harness.hookIndex = 0;
  harness.effects = [];
  return InventoryInvitationScreen(props);
}

async function runEffects(): Promise<() => void> {
  const cleanups = harness.effects.map((effect) => effect()).filter((value): value is () => void => typeof value === 'function');
  await flushPromises();
  return () => cleanups.forEach((cleanup) => cleanup());
}

function press(node: TestNode | undefined): Promise<void> | void {
  return (node?.props?.onPress as (() => Promise<void> | void) | undefined)?.();
}

function findTextButton(node: unknown, text: string): TestNode | undefined {
  const textNode = findText(node, text);
  return findAncestor(node, textNode, (candidate) => candidate.type === 'Pressable');
}

function findText(node: unknown, text: string): TestNode | undefined {
  return findNode(node, (candidate) => candidate.type === 'Text' && textContent(candidate) === text);
}

function findByType(node: unknown, type: string): TestNode | undefined {
  return findNode(node, (candidate) => candidate.type === type);
}

function findAllByType(node: unknown, type: string): TestNode[] {
  const matches: TestNode[] = [];
  walk(node, (candidate) => { if (candidate.type === type) matches.push(candidate); });
  return matches;
}

function findNode(node: unknown, predicate: (candidate: TestNode) => boolean): TestNode | undefined {
  let result: TestNode | undefined;
  walk(node, (candidate) => { if (!result && predicate(candidate)) result = candidate; });
  return result;
}

function findAncestor(node: unknown, target: TestNode | undefined, predicate: (candidate: TestNode) => boolean): TestNode | undefined {
  if (!target) return undefined;
  let result: TestNode | undefined;
  const visit = (value: unknown, ancestors: TestNode[]) => {
    if (!value || typeof value !== 'object' || result) return;
    const candidate = value as TestNode;
    if (candidate === target) result = ancestors.toReversed().find(predicate);
    childList(candidate).forEach((child) => visit(child, [...ancestors, candidate]));
  };
  visit(node, []);
  return result;
}

function walk(node: unknown, visit: (candidate: TestNode) => void): void {
  if (Array.isArray(node)) return node.forEach((child) => walk(child, visit));
  if (!node || typeof node !== 'object') return;
  const candidate = node as TestNode;
  visit(candidate);
  childList(candidate).forEach((child) => walk(child, visit));
}

function childList(node: TestNode): unknown[] {
  const children = node.props?.children;
  return Array.isArray(children) ? children : [children];
}

function textContent(node: TestNode): string {
  return childList(node).filter((child): child is string => typeof child === 'string').join('');
}

function styleValue(style: unknown, property: string): unknown {
  const values = Array.isArray(style) ? style : [style];
  const entry = values.toReversed().find((value) => value && typeof value === 'object' && property in value);
  return entry ? (entry as Record<string, unknown>)[property] : undefined;
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => { resolve = done; });
  return { promise, resolve };
}

async function flushPromises(): Promise<void> {
  await Promise.resolve();
  await Promise.resolve();
}

type TestNode = {
  readonly props?: Record<string, unknown> & { readonly children?: unknown };
  readonly type?: unknown;
};

const palette = {
  action: '#0066CC', actionPressed: '#0055AA', background: '#F7FAFB', onAction: '#FFFFFF',
  success: '#26833A', surface: '#FFFFFF', surfaceMuted: '#E8F0F5', text: '#243038', textMuted: '#52616B'
};
