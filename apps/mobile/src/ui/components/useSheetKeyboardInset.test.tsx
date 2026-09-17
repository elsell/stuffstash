import React from 'react';
import { expect, it } from 'vitest';
import type { SheetBoundaryPort } from './SheetBoundaryPort';
import type { SheetBottomBoundary } from './keyboardBoundaryInset';
import { MobileRenderHarness } from '../../test-support/render';
import { emitKeyboardEventForTest } from '../../test-support/react-native';
import { useSheetKeyboardInset } from './useSheetKeyboardInset';

it('ignores superseded measurements and clears the inset when the keyboard hides', async () => {
  const h = new MobileRenderHarness();
  const measurements: ((boundary: SheetBottomBoundary | null) => void)[] = [];
  const boundaryRef: { current: SheetBoundaryPort } = { current: { measureInKeyboardWindow: () => new Promise(resolve => { measurements.push(resolve); }) } };
  let state!: ReturnType<typeof useSheetKeyboardInset>;
  function Fixture() { state = useSheetKeyboardInset(boundaryRef); return null; }
  try {
    await h.render(<Fixture />);
    await h.run(() => emitKeyboardEventForTest('keyboardWillChangeFrame', { endCoordinates: { screenX: 0, screenY: 544, width: 390, height: 300 } }));
    await h.run(() => emitKeyboardEventForTest('keyboardWillChangeFrame', { endCoordinates: { screenX: 0, screenY: 644, width: 390, height: 200 } }));
    await h.run(() => measurements[1]({ x: 0, y: 844, width: 390 }));
    expect(state.bottomInset).toBe(200);
    await h.run(() => measurements[0]({ x: 0, y: 844, width: 390 }));
    expect(state.bottomInset).toBe(200);
    await h.run(() => state.measure());
    await h.run(() => emitKeyboardEventForTest('keyboardDidHide'));
    await h.run(() => measurements[2]({ x: 0, y: 844, width: 390 }));
    expect(state.bottomInset).toBe(0);
  } finally { await h.unmount(); }
});

it('remeasures the settled iOS keyboard and sheet geometry after the anticipated frame', async () => {
  const h = new MobileRenderHarness();
  let boundaryY = 844;
  const boundaryRef: { current: SheetBoundaryPort } = {
    current: { measureInKeyboardWindow: async () => ({ x: 0, y: boundaryY, width: 390 }) }
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

it('uses keyboard-window sheet coordinates and clears unavailable or rejected measurements', async () => {
  const h = new MobileRenderHarness();
  let frame: SheetBottomBoundary | null = { x: 0, y: 874, width: 402 };
  let reject = false;
  const boundaryRef: { current: SheetBoundaryPort | null } = { current: {
    measureInKeyboardWindow: async () => { if (reject) throw new Error('View detached'); return frame; }
  } };
  let state!: ReturnType<typeof useSheetKeyboardInset>;
  function Fixture() { state = useSheetKeyboardInset(boundaryRef); return null; }
  try {
    await h.render(<Fixture />);
    await h.run(() => emitKeyboardEventForTest('keyboardDidShow', { endCoordinates: { screenX: 0, screenY: 495, width: 402, height: 379 } }));
    expect(state.bottomInset).toBe(379);
    frame = null;
    await h.run(() => state.measure()); expect(state.bottomInset).toBe(0);
    frame = { x: 0, y: 874, width: 402 };
    await h.run(() => state.measure()); expect(state.bottomInset).toBe(379);
    reject = true;
    await h.run(() => state.measure()); expect(state.bottomInset).toBe(0);
    reject = false;
    await h.run(() => state.measure()); expect(state.bottomInset).toBe(379);
    boundaryRef.current = null;
    await h.run(() => state.measure()); expect(state.bottomInset).toBe(0);
  } finally { await h.unmount(); }
});
