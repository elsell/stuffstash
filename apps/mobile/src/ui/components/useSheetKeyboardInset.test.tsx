import React from 'react';
import { expect, it } from 'vitest';
import type { View } from 'react-native';
import { MobileRenderHarness } from '../../test-support/render';
import { emitKeyboardEventForTest } from '../../test-support/react-native';
import { useSheetKeyboardInset } from './useSheetKeyboardInset';

it('ignores superseded measurements and clears the inset when the keyboard hides', async () => {
  const h = new MobileRenderHarness();
  const measurements: ((x: number, y: number, width: number, height: number) => void)[] = [];
  const boundaryRef: { current: Pick<View, 'measureInWindow'> } = { current: { measureInWindow: callback => { measurements.push(callback); } } };
  let state!: ReturnType<typeof useSheetKeyboardInset>;
  function Fixture() { state = useSheetKeyboardInset(boundaryRef); return null; }
  try {
    await h.render(<Fixture />);
    await h.run(() => emitKeyboardEventForTest('keyboardWillChangeFrame', { endCoordinates: { screenX: 0, screenY: 544, width: 390, height: 300 } }));
    await h.run(() => emitKeyboardEventForTest('keyboardWillChangeFrame', { endCoordinates: { screenX: 0, screenY: 644, width: 390, height: 200 } }));
    await h.run(() => measurements[1](0, 844, 390, 0));
    expect(state.bottomInset).toBe(200);
    await h.run(() => measurements[0](0, 844, 390, 0));
    expect(state.bottomInset).toBe(200);
    await h.run(() => state.measure());
    await h.run(() => emitKeyboardEventForTest('keyboardDidHide'));
    await h.run(() => measurements[2](0, 844, 390, 0));
    expect(state.bottomInset).toBe(0);
  } finally { await h.unmount(); }
});

it('remeasures the settled iOS keyboard and sheet geometry after the anticipated frame', async () => {
  const h = new MobileRenderHarness();
  let boundaryY = 844;
  const boundaryRef: { current: Pick<View, 'measureInWindow'> } = {
    current: { measureInWindow: callback => callback(0, boundaryY, 390, 0) }
  };
  let state!: ReturnType<typeof useSheetKeyboardInset>;
  function Fixture() { state = useSheetKeyboardInset(boundaryRef); return null; }
  try {
    await h.render(<Fixture />);
    await h.run(() => emitKeyboardEventForTest('keyboardWillChangeFrame', { endCoordinates: { screenX: 0, screenY: 544, width: 390, height: 300 } }));
    expect(state.bottomInset).toBe(300);
    boundaryY = 820;
    await h.run(() => emitKeyboardEventForTest('keyboardDidChangeFrame', { endCoordinates: { screenX: 0, screenY: 490, width: 390, height: 354 } }));
    expect(state.bottomInset).toBe(330);
    boundaryY = 800;
    await h.run(() => emitKeyboardEventForTest('keyboardDidShow', { endCoordinates: { screenX: 0, screenY: 490, width: 390, height: 354 } }));
    expect(state.bottomInset).toBe(310);
    await h.run(() => emitKeyboardEventForTest('keyboardDidHide'));
    expect(state.bottomInset).toBe(0);
  } finally { await h.unmount(); }
});
