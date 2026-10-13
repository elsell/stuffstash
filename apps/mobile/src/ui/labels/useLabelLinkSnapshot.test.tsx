import { Text } from 'react-native';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { useLabelLinkSnapshot, type LabelLinkSource } from './useLabelLinkSnapshot';
import { parseMobileLabelLink } from '../../adapters/labels/LabelLinkParser';
const url = 'https://old.example/l/v1/01ARZ3NDEKTSV4RRFFQ69G5FAV/01ARZ3NDEKTSV4RRFFQ69G5FAW';
it('retains a parsed scan through session replacement and returns its revision before rendering', async () => {
  const h = new MobileRenderHarness(); let finish!: (url: string) => void;
  const source: LabelLinkSource = { getInitialURL: () => new Promise(resolve => { finish = resolve; }), subscribe: () => () => {} };
  let value!: ReturnType<typeof useLabelLinkSnapshot>;
  function Probe({ session }: { session: string }) { value = useLabelLinkSnapshot(source, parseMobileLabelLink); return <Text key={session}>{session}</Text>; }
  const reference = parseMobileLabelLink(url);
  try {
    await h.render(<Probe session="signed-in" />);
    await h.run(() => value.capture('stuffstash://labels/v9/invalid/invalid'));
    const previousRevision = value.revision; let retainedRevision = previousRevision;
    await h.run(() => { retainedRevision = value.retain(reference); expect(retainedRevision).toBeGreaterThan(previousRevision); });
    expect(value).toMatchObject({ reference, revision: retainedRevision, initialized: true, invalid: false });
    await h.render(<Probe session="expired" />);
    await h.run(() => finish('stuffstash://labels/v9/invalid/invalid'));
    expect(value).toMatchObject({ reference, revision: retainedRevision, initialized: true, invalid: false });
  } finally { await h.unmount(); }
});

it('retains foreground identity through auth child replacement, ignores stale initial lookup and cannot resurrect cleared intent', async () => {
  const h = new MobileRenderHarness(); let receive!: (url: string) => void; let finish!: (url: string) => void; let removed = false;
  const source: LabelLinkSource = { getInitialURL: async () => new Promise(resolve => { finish = resolve; }), subscribe: listener => { receive = listener; return () => { removed = true; }; } };
  let value!: ReturnType<typeof useLabelLinkSnapshot>;
  function Probe({ session }: { session: string }) { value = useLabelLinkSnapshot(source, parseMobileLabelLink); return <Text>{session}</Text>; }
  try {
    await h.render(<Probe session="signed-out" />);
    expect(value.initialized).toBe(false);
    await h.run(() => receive(url));
    expect(value.initialized).toBe(true);
    expect(value.reference?.instanceId).toBe('01ARZ3NDEKTSV4RRFFQ69G5FAV');
    await h.render(<Probe session="another-server" />); expect(value.reference).toBeDefined();
    await h.run(value.clear); await h.run(() => finish(url)); expect(value.reference).toBeUndefined();
    await h.run(() => receive('stuffstash://auth/callback?code=secret')); expect(value.reference).toBeUndefined();
    await h.run(() => receive('stuffstash://labels/v9/invalid/invalid')); expect(value.invalid).toBe(true);
  } finally { await h.unmount(); }
  expect(removed).toBe(true);
});

it.each(['valid', 'empty', 'failure'] as const)('finishes initial-link readiness after a %s initial lookup', async result => {
  const h = new MobileRenderHarness(); let finish!: (value: string | null) => void; let fail!: () => void;
  const source: LabelLinkSource = { getInitialURL: () => new Promise((resolve, reject) => { finish = resolve; fail = () => reject(new Error('unavailable')); }), subscribe: () => () => {} };
  let value!: ReturnType<typeof useLabelLinkSnapshot>;
  function Probe() { value = useLabelLinkSnapshot(source, parseMobileLabelLink); return null; }
  try {
    await h.render(<Probe />);
    expect(value.initialized).toBe(false);
    await h.run(() => result === 'failure' ? fail() : finish(result === 'valid' ? url : null));
    expect(value.initialized).toBe(true);
    expect(Boolean(value.reference)).toBe(result === 'valid');
  } finally { await h.unmount(); }
});
