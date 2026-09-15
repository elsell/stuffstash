import React from 'react';
import { Pressable, Text } from 'react-native';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { useInvitationLinkSnapshot, type InvitationLinkSource } from './useInvitationLinkSnapshot';

class LinkSource implements InvitationLinkSource {
  listener?: (url: string) => void;
  removed = false;
  resolve!: (url: string | null) => void;
  reject!: (error: Error) => void;
  initial = new Promise<string | null>((resolve, reject) => { this.resolve = resolve; this.reject = reject; });
  getInitialURL() { return this.initial; }
  subscribe(listener: (url: string) => void) { this.listener = listener; return () => { this.removed = true; this.listener = undefined; }; }
}
const url = (id: string) => `stuffstash://invitations/accept?tenant=tenant-one&inventory=inventory-one&invitation=${id}#token=${'F'.repeat(43)}`;
function Probe({ source }: { source: InvitationLinkSource }) {
  const state = useInvitationLinkSnapshot(source, 'https://stash.example.test', false);
  return <><Text>{JSON.stringify(state.snapshot)}</Text><Pressable accessibilityLabel="Clear invitation" onPress={state.clear} /></>;
}
it.each([false, true])('finishes failed initial lookup and retains foreground capture=%s', async foreground => {
  const h = new MobileRenderHarness(); const source = new LinkSource();
  try {
    await h.render(<Probe source={source} />);
    if (foreground) await h.run(() => source.listener?.(url('foreground')));
    await h.run(() => source.reject(new Error('system lookup failed')));
    const state = JSON.parse(h.allText()[0]);
    expect(state.initialized).toBe(true);
    expect(state.reference?.invitationId).toBe(foreground ? 'foreground' : undefined);
    await h.run(() => source.listener?.(url('later')));
    expect(JSON.parse(h.allText()[0]).reference.invitationId).toBe('later');
  } finally { await h.unmount(); }
});
it('retains foreground capture against late initial URL and removes its subscription', async () => {
  const h = new MobileRenderHarness(); const source = new LinkSource();
  await h.render(<Probe source={source} />);
  await h.run(() => source.listener?.(url('foreground')));
  await h.run(() => source.resolve(url('initial')));
  expect(JSON.parse(h.allText()[0]).reference.invitationId).toBe('foreground');
  await h.unmount(); expect(source.removed).toBe(true);
});
it('keeps a foreground invitation when a late initial lookup returns no URL', async () => {
  const h = new MobileRenderHarness(); const source = new LinkSource();
  try {
    await h.render(<Probe source={source} />);
    await h.run(() => source.listener?.(url('foreground')));
    await h.run(() => source.resolve(null));
    expect(JSON.parse(h.allText()[0]).reference.invitationId).toBe('foreground');
  } finally { await h.unmount(); }
});
it('disposes the source before a late failed lookup completes', async () => {
  const h = new MobileRenderHarness(); const source = new LinkSource();
  await h.render(<Probe source={source} />);
  await h.unmount();
  expect(source.removed).toBe(true);
  await h.run(() => source.reject(new Error('late failure')));
});

it('does not resurrect a cleared invitation from a late initial result', async () => {
  const h = new MobileRenderHarness(); const source = new LinkSource();
  try {
    await h.render(<Probe source={source} />);
    await h.press(h.byLabel('Clear invitation'));
    await h.run(() => source.resolve(url('initial')));
    expect(JSON.parse(h.allText()[0]).reference).toBeUndefined();
    await h.run(() => source.listener?.(url('new')));
    expect(JSON.parse(h.allText()[0]).reference.invitationId).toBe('new');
  } finally { await h.unmount(); }
});
