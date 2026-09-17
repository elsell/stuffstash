import React from 'react';
import { Text, type View } from 'react-native';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../src/test-support/render';
import { emitKeyboardEventForTest } from '../src/test-support/react-native';
import { useFilterGeometryProbe } from './FilterGeometryProbe';

it('retains paired keyboard and measured sheet coordinates and rejects late measurements after hide', async () => {
  const h = new MobileRenderHarness();
  const callbacks: ((x: number, y: number, width: number, height: number) => void)[] = [];
  const ref: { current: Pick<View, 'measureInWindow'> } = { current: { measureInWindow: callback => callbacks.push(callback) } };
  function Probe() { return <Text testID="geometry">{JSON.stringify(useFilterGeometryProbe(ref))}</Text>; }
  try {
    await h.render(<Probe />);
    const frame = { screenX: 0, screenY: 495, width: 402, height: 379 };
    await h.run(() => emitKeyboardEventForTest('keyboardDidShow', { endCoordinates: frame }));
    await h.run(() => callbacks.at(-1)!(0, 812, 402, 0));
    expect(JSON.parse(h.byTestId('geometry')!.props.children)).toEqual({ boundary: { x: 0, y: 812, width: 402 }, keyboard: frame, calculatedInset: 317 });
    await h.run(() => emitKeyboardEventForTest('keyboardDidChangeFrame', { endCoordinates: frame }));
    const late = callbacks.at(-1)!;
    await h.run(() => emitKeyboardEventForTest('keyboardDidHide'));
    await h.run(() => late(0, 874, 402, 0));
    expect(JSON.parse(h.byTestId('geometry')!.props.children)).toBeNull();
  } finally { await h.unmount(); }
});
