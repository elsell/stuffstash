import React from 'react';
import { expect, it } from 'vitest';
import useDoubleTapToZoom from 'react-native-image-viewing/dist/hooks/useDoubleTapToZoom';
import { MobileRenderHarness } from '../../test-support/render';

type Tap = ReturnType<typeof useDoubleTapToZoom>;
type TapEvent = Parameters<Tap>[0];
type ViewerRef = Parameters<typeof useDoubleTapToZoom>[0];
const touch = (x: number, y: number) => ({ nativeEvent: { pageX: x, pageY: y } }) as TapEvent;

// Use real timers and the installed adapter. Zoom calls are recorded by an
// in-memory scroll responder; no native viewer test double handles these taps.
it('reveals commands on a single tap and zooms on a double tap without toggling commands', async () => {
  const h = new MobileRenderHarness();
  let singles = 0;
  const zooms: unknown[] = [];
  const ref = { current: { getScrollResponder: () => ({ scrollResponderZoomTo: (value: unknown) => zooms.push(value) }) } };
  let tap!: Tap;
  function Probe() {
    tap = useDoubleTapToZoom(ref as unknown as ViewerRef, true, { width: 400, height: 800 }, () => { singles++; });
    return null;
  }
  const event = () => touch(200, 400);
  try {
    await h.render(<Probe />);
    await h.run(() => tap(event()));
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toBe(1);
    expect(zooms).toHaveLength(0);
    await h.run(() => { tap(event()); tap(event()); });
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toBe(1);
    expect(zooms).toHaveLength(1);
  } finally { await h.unmount(); }
});

it('cancels a pending single tap when its image unmounts', async () => {
  const h = new MobileRenderHarness();
  let singles = 0;
  let tap!: Tap;
  const ref = { current: { getScrollResponder: () => ({ scrollResponderZoomTo: () => {} }) } };
  function Probe() {
    tap = useDoubleTapToZoom(ref as unknown as ViewerRef, false, { width: 400, height: 800 }, () => { singles++; });
    return null;
  }
  await h.render(<Probe />);
  await h.run(() => tap(touch(100, 100)));
  await h.unmount();
  await new Promise(resolve => setTimeout(resolve, 350));
  expect(singles).toBe(0);
});

it('retires callbacks when the image scope changes or the viewer unmounts', async () => {
  const h = new MobileRenderHarness();
  const singles: string[] = [];
  let tap!: Tap;
  const ref = { current: { getScrollResponder: () => ({ scrollResponderZoomTo: () => {} }) } };
  function Probe({ scope }: { scope: string }) {
    tap = useDoubleTapToZoom(ref as unknown as ViewerRef, false, { width: 400, height: 800 },
      () => { singles.push(scope); }, true, scope);
    return null;
  }
  try {
    await h.render(<Probe scope="first" />);
    const old = tap;
    await h.run(() => old(touch(100, 100)));
    await h.render(<Probe scope="second" />);
    await h.run(() => old(touch(100, 100)));
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toEqual([]);
    await h.run(() => tap(touch(100, 100)));
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toEqual(['second']);
    await h.unmount();
    await h.run(() => tap(touch(100, 100)));
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toEqual(['second']);
  } finally { await h.unmount(); }
});

it('does not interpret taps in separate viewers as a double tap', async () => {
  const h = new MobileRenderHarness();
  const singles: string[] = [];
  const taps = new Map<string, Tap>();
  const zooms: string[] = [];
  function Probe({ id }: { id: string }) {
    const ref = React.useRef({ getScrollResponder: () => ({ scrollResponderZoomTo: () => { zooms.push(id); } }) });
    taps.set(id, useDoubleTapToZoom(ref as unknown as ViewerRef, false, { width: 400, height: 800 },
      () => { singles.push(id); }));
    return null;
  }
  try {
    await h.render(<><Probe id="one" /><Probe id="two" /></>);
    await h.run(() => { taps.get('one')!(touch(100, 100)); taps.get('two')!(touch(100, 100)); });
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toEqual(['one', 'two']);
    expect(zooms).toEqual([]);
  } finally { await h.unmount(); }
});
