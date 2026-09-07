import { createMobilePerformanceSession } from '../../adapters/observability/MobilePerformanceSession';
import React from 'react';
import { Pressable, Text } from 'react-native';
import { onlineManager } from '@tanstack/react-query';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { setAppStateForTest } from '../../test-support/react-native';
import { createMobileQueryClient, mobileQueryKeys, resetMobileInventorySelection } from '../../adapters/serverState/MobileQueryClient';
import { MobileServerStateProvider } from './MobileServerStateProvider';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';
const settle = (h: MobileRenderHarness) => h.run(() => new Promise(r => setTimeout(r, 10)));

it('refreshes mounted consumers on selection reset without showing the previous inventory', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let selected = 'Garage'; let resolve!: (name: string) => void;
  function Surface() { const result = useMobileInventoryServerQuery({ key: mobileQueryKeys.home, query: async () => selected === 'Garage' ? selected : new Promise<string>(r => { resolve = r; }) }); return <Text>{result.data ?? 'Loading'}</Text>; }
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: selected })}><Surface /></MobileServerStateProvider>);
    await settle(h); await settle(h); expect(h.allText()).toEqual(['Garage']);
    selected = 'Kitchen'; await h.run(() => resetMobileInventorySelection(client, 'scope')); await settle(h); await settle(h);
    expect(h.allText()).toEqual(['Loading']);
    await h.run(() => resolve('Kitchen')); await settle(h); expect(h.allText()).toEqual(['Kitchen']);
  } finally { await h.unmount(); }
});

it('cancels and clears the old composition on provider replacement', async () => {
  const h = new MobileRenderHarness(); const oldClient = createMobileQueryClient(); const nextClient = createMobileQueryClient(); let signal: AbortSignal | undefined;
  function Surface({ pending }: { pending: boolean }) { const result = useMobileInventoryServerQuery({ key: mobileQueryKeys.home, query: async s => { if (pending) { signal = s; return new Promise<string>(() => undefined); } return 'New session'; } }); return <Text>{result.data ?? 'Loading'}</Text>; }
  const view = (old: boolean) => <MobileServerStateProvider client={old ? oldClient : nextClient} scopeId={old ? 'old' : 'next'} loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><Surface pending={old} /></MobileServerStateProvider>;
  try { await h.render(view(true)); await settle(h); await settle(h); expect(signal?.aborted).toBe(false); await h.render(view(false)); await settle(h); await settle(h); expect(signal?.aborted).toBe(true); expect(oldClient.getQueryCache().getAll()).toHaveLength(0); expect(h.allText()).toEqual(['New session']); } finally { await h.unmount(); }
});

it('reconciles stale active data on foreground and reconnect', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let reads = 0;
  client.setDefaultOptions({ queries: { ...client.getDefaultOptions().queries, staleTime: 0 } });
  function Surface() { const result = useMobileInventoryServerQuery({ key: mobileQueryKeys.home, query: async () => `Read ${++reads}` }); return <Text>{result.data ?? 'Loading'}</Text>; }
  try { await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><Surface /></MobileServerStateProvider>); await settle(h); await settle(h); expect(reads).toBe(1); await h.run(() => setAppStateForTest('background')); await h.run(() => setAppStateForTest('active')); await settle(h); expect(reads).toBe(2); await h.run(() => onlineManager.setOnline(false)); await h.run(() => onlineManager.setOnline(true)); await settle(h); expect(reads).toBe(3); } finally { await h.unmount(); onlineManager.setOnline(true); setAppStateForTest('active'); }
});

it('hides warmed resource values after authoritative denial while keeping transient failures usable', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let status = 0;
  function Surface() { const result = useMobileInventoryServerQuery({ key: mobileQueryKeys.home, query: async () => { if (status) throw Object.assign(new Error('unavailable'), { status }); return 'Private inventory'; } }); return <Text>{result.data ?? 'Unavailable'}</Text>; }
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><Surface /></MobileServerStateProvider>); await settle(h); await settle(h);
    status = 400; await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.home('scope', 'tenant', 'inventory') })); await settle(h); expect(h.allText()).toEqual(['Private inventory']);
    status = 403; await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.home('scope', 'tenant', 'inventory') })); await settle(h); expect(h.allText()).toEqual(['Unavailable']);
    status = 500; await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.home('scope', 'tenant', 'inventory') })); await settle(h); expect(h.allText()).toEqual(['Unavailable']);
    status = 0; await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.home('scope', 'tenant', 'inventory') })); await settle(h); expect(h.allText()).toEqual(['Private inventory']);
  } finally { await h.unmount(); }
});

it('recovers shared resource reads by retrying denied discovery before the resource', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let denied = true; let reads = 0;
  function Surface() { const result = useMobileInventoryServerQuery({ key: mobileQueryKeys.home, query: async () => { reads++; return 'Restored inventory'; } }); return <><Text>{result.data ?? 'Unavailable'}</Text><Pressable accessibilityLabel="Try again" onPress={() => result.refetch()} /></>; }
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => { if (denied) throw Object.assign(new Error('denied'), { status: 403 }); return { tenantId: 'tenant', inventoryId: 'inventory' }; }}><Surface /></MobileServerStateProvider>); await settle(h); await settle(h);
    expect(h.allText()).toEqual(['Unavailable']); expect(reads).toBe(0);
    denied = false; await h.press(h.byLabel('Try again')); await settle(h); await settle(h);
    expect(h.allText()).toEqual(['Restored inventory']); expect(reads).toBe(1);
  } finally { await h.unmount(); }
});


it('disposes performance collection on session replacement and unmount', async () => {
  const h = new MobileRenderHarness();
  const first = createMobileQueryClient(); const second = createMobileQueryClient();
  const disposed: string[] = [];
  const releaseFirst = () => () => { disposed.push('first'); };
  const releaseSecond = () => () => { disposed.push('second'); };
  const view = (next: boolean) => <MobileServerStateProvider client={next ? second : first} scopeId={next ? 'second' : 'first'} acquirePerformance={next ? releaseSecond : releaseFirst} loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><Text>Ready</Text></MobileServerStateProvider>;
  await h.render(view(false));
  await h.render(view(true));
  expect(disposed).toEqual(['first']);
  await h.unmount();
  expect(disposed).toEqual(['first', 'second']);
});


it('keeps measurement active across StrictMode effect replay and stops after unmount', async () => {
  const h = new MobileRenderHarness();
  const client = createMobileQueryClient();
  const active = new Set<() => void>();
  const session = createMobilePerformanceSession({ platform: 'ios', enabled: true, baseUrl: 'https://api.example.test', tokenProvider: () => null,
    fetch: async () => new Response(), clock: { now: () => 0 },
    scheduler: { schedule(callback) { active.add(callback); return () => { active.delete(callback); }; } }
  });
  let acquisitions = 0;
  const acquire = () => { acquisitions++; return session.acquire(); };
  try {
    await h.render(<React.StrictMode><MobileServerStateProvider client={client} scopeId="session" acquirePerformance={acquire} loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><Text>Ready</Text></MobileServerStateProvider></React.StrictMode>);
    expect(acquisitions).toBeGreaterThan(1);
    session.observer.start({ operation: 'image', surface: 'detail', variant: 'large' })('success');
    expect(active.size).toBe(1);
  } finally { await h.unmount(); }
  expect(active.size).toBe(0);
});
