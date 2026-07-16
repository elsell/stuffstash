import { beforeEach, describe, expect, it, vi } from 'vitest';
import { InventoryInvitationLinkProvider } from './InventoryInvitationLinkContext';

const harness = vi.hoisted(() => ({
  effects: [] as Array<() => void | (() => void)>,
  hookIndex: 0,
  initialURL: undefined as undefined | Promise<string | null>,
  listener: undefined as undefined | ((event: { url: string }) => void),
  values: [] as unknown[]
}));

vi.mock('react', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react')>()),
  useCallback: <T,>(callback: T) => callback,
  useEffect: (effect: () => void | (() => void)) => harness.effects.push(effect),
  useMemo: <T,>(factory: () => T) => factory(),
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

vi.mock('expo-linking', () => ({
  addEventListener: (_type: string, listener: (event: { url: string }) => void) => {
    harness.listener = listener;
    return { remove: vi.fn() };
  },
  getInitialURL: () => harness.initialURL
}));

vi.mock('../../config/mobileRuntimeConfig', () => ({
  loadMobileRuntimeConfigSeed: () => ({
    invitationOrigin: 'https://stash.example.test',
    invitationAllowInsecureLocalHTTP: false
  })
}));

describe('InventoryInvitationLinkProvider', () => {
  beforeEach(() => {
    harness.effects = [];
    harness.hookIndex = 0;
    harness.listener = undefined;
    harness.values = [];
  });

  it('keeps a foreground invitation authoritative when initial URL resolution finishes later', async () => {
    const initial = deferred<string | null>();
    harness.initialURL = initial.promise;
    render();
    harness.effects.forEach((effect) => effect());

    harness.listener?.({ url: invitationURL('foreground', 'F') });
    initial.resolve(invitationURL('initial', 'I'));
    await flushPromises();

    const value = providerValue(render());
    expect(value.reference).toMatchObject({
      invitationId: 'foreground',
      acceptanceToken: 'F'.repeat(43)
    });
    expect(value.initialized).toBe(true);
  });

  it('does not clear a foreground invitation when the late initial lookup has no URL', async () => {
    const initial = deferred<string | null>();
    harness.initialURL = initial.promise;
    render();
    harness.effects.forEach((effect) => effect());

    harness.listener?.({ url: invitationURL('foreground', 'F') });
    initial.resolve(null);
    await flushPromises();

    expect(providerValue(render()).reference?.invitationId).toBe('foreground');
  });
});

function render(): unknown {
  harness.hookIndex = 0;
  harness.effects = [];
  return InventoryInvitationLinkProvider({ children: 'content' });
}

function providerValue(tree: unknown): {
  readonly initialized: boolean;
  readonly reference?: { readonly invitationId: string; readonly acceptanceToken: string };
} {
  return (tree as { props: { value: unknown } }).props.value as never;
}

function invitationURL(invitationId: string, tokenCharacter: string): string {
  return `stuffstash://invitations/accept?tenant=tenant-one&inventory=inventory-one&invitation=${invitationId}#token=${tokenCharacter.repeat(43)}`;
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
