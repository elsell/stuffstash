import { Text } from 'react-native';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { useLabelLinkSnapshot, type LabelLinkSource } from './useLabelLinkSnapshot';
import { parseMobileLabelLink } from '../../adapters/labels/LabelLinkParser';
const url = 'https://old.example/l/v1/01ARZ3NDEKTSV4RRFFQ69G5FAV/01ARZ3NDEKTSV4RRFFQ69G5FAW';
it('retains foreground identity through auth child replacement, ignores stale initial lookup and cannot resurrect cleared intent', async () => {
  const h = new MobileRenderHarness(); let receive!: (url: string) => void; let finish!: (url: string) => void; let removed = false;
  const source: LabelLinkSource = { getInitialURL: async () => new Promise(resolve => { finish = resolve; }), subscribe: listener => { receive = listener; return () => { removed = true; }; } };
  let value!: ReturnType<typeof useLabelLinkSnapshot>;
  function Probe({ session }: { session: string }) { value = useLabelLinkSnapshot(source, parseMobileLabelLink); return <Text>{session}</Text>; }
  try {
    await h.render(<Probe session="signed-out" />); await h.run(() => receive(url));
    expect(value.reference?.instanceId).toBe('01ARZ3NDEKTSV4RRFFQ69G5FAV');
    await h.render(<Probe session="another-server" />); expect(value.reference).toBeDefined();
    await h.run(value.clear); await h.run(() => finish(url)); expect(value.reference).toBeUndefined();
    await h.run(() => receive('stuffstash://auth/callback?code=secret')); expect(value.reference).toBeUndefined();
    await h.run(() => receive('stuffstash://labels/v9/invalid/invalid')); expect(value.invalid).toBe(true);
  } finally { await h.unmount(); }
  expect(removed).toBe(true);
});
