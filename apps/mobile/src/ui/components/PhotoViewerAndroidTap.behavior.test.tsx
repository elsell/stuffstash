import React from 'react';
import { expect, it } from 'vitest';
import usePanResponder from 'react-native-image-viewing/dist/hooks/usePanResponder';
import type { GestureResponderEvent, PanResponderGestureState, PanResponderCallbacks } from 'react-native';
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
