import { afterEach, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import type { VoiceResponseArtifact } from '../../application/voice/RealtimeVoiceSession';
import { useVoiceReferenceNavigation } from './useVoiceReferenceNavigation';

const reference: VoiceResponseArtifact = { type: 'asset_reference', assetId: 'drill', title: 'Drill', assetKind: 'item' };
let h: MobileRenderHarness;
afterEach(async () => { await h?.unmount(); setScreenFocused(true); });

it.each(['current', 'left', 'returned', 'scope'] as const)('owns reference navigation after media stops: %s', async scenario => {
  h = new MobileRenderHarness(); setScreenFocused(true);
  let release!: () => void; let pauses = 0;
  const stopped = new Promise<void>(resolve => { release = resolve; });
  const opened: string[] = [];
  let open!: ReturnType<typeof useVoiceReferenceNavigation>;
  const pauseMedia = async () => { pauses++; await stopped; };
  function Surface({ scope }: { scope: string }) {
    open = useVoiceReferenceNavigation({ scopeIdentity: scope, pauseMedia, onOpen: item => { opened.push(item.assetId); } });
    return null;
  }
  await h.render(<Surface scope="first" />);
  const retained = open;
  let pending!: Promise<void>;
  await h.run(() => { pending = open(reference); });
  if (scenario === 'left' || scenario === 'returned') await h.run(() => setScreenFocused(false));
  if (scenario === 'returned') await h.run(() => setScreenFocused(true));
  if (scenario === 'scope') await h.render(<Surface scope="second" />);
  await h.run(async () => { release(); await pending; });
  expect(opened).toEqual(scenario === 'current' ? ['drill'] : []);
  if (scenario !== 'current') {
    await h.run(() => retained(reference));
    expect(pauses).toBe(1);
    expect(opened).toEqual([]);
    if (scenario === 'left') await h.run(() => setScreenFocused(true));
    await h.run(() => open(reference));
    expect(opened).toEqual(['drill']);
  }
});
