import React from 'react';
import { expect, it } from 'vitest';
import usePanResponder from 'react-native-image-viewing/dist/hooks/usePanResponder';
import type { GestureResponderEvent, PanResponderGestureState, PanResponderCallbacks } from 'react-native';
import { animatedValueForTest } from '../../test-support/react-native';
import { MobileRenderHarness } from '../../test-support/render';

it('reveals controls for a tap but not a double tap, drag, or retired responder', async () => {
  const h = new MobileRenderHarness();
  let singles = 0;
  const zooms: boolean[] = [];
  let handlers!: PanResponderCallbacks;
  function Probe() {
    handlers = usePanResponder({ initialScale: 1, initialTranslate: { x: 0, y: 0 },
      onZoom: (scaled: boolean) => { zooms.push(scaled); }, doubleTapToZoomEnabled: true,
      onSingleTap: () => { singles++; }, onLongPress: () => {}, delayLongPress: 800 })[0] as unknown as PanResponderCallbacks;
    return null;
  }
  const event = { nativeEvent: { touches: [{ pageX: 100, pageY: 200 }] } } as GestureResponderEvent;
  const gesture = { numberActiveTouches: 1, dx: 0, dy: 0 } as PanResponderGestureState;
  const press = () => {
    handlers.onPanResponderGrant!(event, gesture);
    handlers.onPanResponderStart!(event, gesture);
    handlers.onPanResponderRelease!(event, gesture);
  };
  try {
    await h.render(<React.StrictMode><Probe /></React.StrictMode>);
    await h.run(press);
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toBe(1);
    await h.run(() => { press(); press(); });
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toBe(1);
    expect(zooms).toContain(true);
    await h.run(() => {
      handlers.onPanResponderGrant!(event, gesture);
      handlers.onPanResponderStart!(event, gesture);
      handlers.onPanResponderMove!(event, { ...gesture, dx: 40 });
      handlers.onPanResponderRelease!(event, gesture);
    });
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toBe(1);
    await h.run(() => {
      handlers.onPanResponderGrant!(event, gesture);
      handlers.onPanResponderStart!(event, gesture);
      handlers.onPanResponderTerminate!(event, gesture);
      handlers.onPanResponderRelease!(event, gesture);
    });
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toBe(1);
    await h.unmount();
    await h.run(press);
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toBe(1);
  } finally { await h.unmount(); }
});


it('preserves zoom scale and reaches current callbacks across unrelated rerenders', async () => {
  const h = new MobileRenderHarness();
  let viewer!: ReturnType<typeof usePanResponder>;
  const singles: string[] = [];
  function Probe({ label }: { label: string }) {
    viewer = usePanResponder({ initialScale: 1, initialTranslate: { x: 0, y: 0 },
      onZoom: () => {}, doubleTapToZoomEnabled: true,
      onSingleTap: () => { singles.push(label); }, onLongPress: () => {}, delayLongPress: 800 });
    return null;
  }
  const event = { nativeEvent: { touches: [{ pageX: 100, pageY: 200 }] } } as GestureResponderEvent;
  const gesture = { numberActiveTouches: 1, dx: 0, dy: 0 } as PanResponderGestureState;
  const press = () => {
    const handlers = viewer[0] as unknown as PanResponderCallbacks;
    handlers.onPanResponderGrant!(event, gesture);
    handlers.onPanResponderStart!(event, gesture);
    handlers.onPanResponderRelease!(event, gesture);
  };
  try {
    await h.render(<Probe label="before" />);
    await h.run(() => { press(); press(); });
    expect(animatedValueForTest(viewer[1])).toBe(2);
    await h.render(<Probe label="after" />);
    expect(animatedValueForTest(viewer[1])).toBe(2);
    await h.run(press);
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toEqual(['after']);
    expect(animatedValueForTest(viewer[1])).toBe(2);
  } finally { await h.unmount(); }
});

it.each(['image', 'geometry'] as const)('retires pending taps and zoom state when %s changes', async change => {
  const h = new MobileRenderHarness();
  let viewer!: ReturnType<typeof usePanResponder>;
  const singles: string[] = [];
  let zoomed = false;
  function Probe({ replaced }: { replaced: boolean }) {
    viewer = usePanResponder({ initialScale: replaced && change === 'geometry' ? 0.5 : 1,
      initialTranslate: { x: 0, y: 0 }, scope: replaced && change === 'image' ? 'replacement' : 'original',
      onZoom: value => { zoomed = value; }, doubleTapToZoomEnabled: true,
      onSingleTap: () => { singles.push(replaced ? 'replacement' : 'original'); }, onLongPress: () => {}, delayLongPress: 800 });
    return null;
  }
  const event = { nativeEvent: { touches: [{ pageX: 100, pageY: 200 }] } } as GestureResponderEvent;
  const gesture = { numberActiveTouches: 1, dx: 0, dy: 0 } as PanResponderGestureState;
  const press = (handlers: PanResponderCallbacks) => {
    handlers.onPanResponderGrant!(event, gesture);
    handlers.onPanResponderStart!(event, gesture);
    handlers.onPanResponderRelease!(event, gesture);
  };
  try {
    await h.render(<Probe replaced={false} />);
    const old = viewer[0] as unknown as PanResponderCallbacks;
    await h.run(() => { press(old); press(old); });
    expect(animatedValueForTest(viewer[1])).toBe(2);
    expect(zoomed).toBe(true);
    await h.run(() => press(old));
    await h.render(<Probe replaced />);
    expect(animatedValueForTest(viewer[1])).toBe(change === 'geometry' ? 0.5 : 1);
    expect(zoomed).toBe(false);
    await h.run(() => press(old));
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toEqual([]);
    await h.run(() => press(viewer[0] as unknown as PanResponderCallbacks));
    await new Promise(resolve => setTimeout(resolve, 350));
    expect(singles).toEqual(['replacement']);
  } finally { await h.unmount(); }
});
