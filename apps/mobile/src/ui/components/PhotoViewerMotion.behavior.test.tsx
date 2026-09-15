import React from 'react';
import { expect, it } from 'vitest';
// Exercise the installed dependency hook rather than the viewer test double.
import useAnimatedComponents from 'react-native-image-viewing/dist/hooks/useAnimatedComponents';
import { animatedValueForTest, animationStopCount, deferAnimationsForTest, pendingAnimationCount, animationStartCount, holdReduceMotionSnapshotForTest, setReduceMotionEnabledForTest } from '../../test-support/react-native';
import { MobileRenderHarness } from '../../test-support/render';

it('keeps viewer chrome still until motion is allowed and honors later changes', async () => {
  const h = new MobileRenderHarness();
  const snapshot = holdReduceMotionSnapshotForTest();
  let toggle!: (visible: boolean) => void;
  function Probe() { [, , toggle] = useAnimatedComponents(); return null; }
  try {
    await h.render(<Probe />);
    const before = animationStartCount();
    await h.run(() => toggle(false));
    expect(animationStartCount()).toBe(before);
    await h.run(() => setReduceMotionEnabledForTest(true));
    await h.run(() => snapshot.resolve(false));
    await h.run(() => toggle(true));
    expect(animationStartCount()).toBe(before);
    await h.run(() => setReduceMotionEnabledForTest(false));
    await h.run(() => toggle(false));
    expect(animationStartCount()).toBeGreaterThan(before);
    await h.run(() => setReduceMotionEnabledForTest(true));
    const after = animationStartCount();
    await h.run(() => toggle(true));
    expect(animationStartCount()).toBe(after);
  } finally { await h.unmount(); }
});

it('keeps viewer motion suppressed when the native preference cannot be read', async () => {
  const h = new MobileRenderHarness();
  const snapshot = holdReduceMotionSnapshotForTest();
  let toggle!: (visible: boolean) => void;
  function Probe() { [, , toggle] = useAnimatedComponents(); return null; }
  try {
    await h.render(<Probe />);
    await h.run(() => snapshot.reject(new Error('Preference unavailable')));
    const before = animationStartCount();
    await h.run(() => toggle(false));
    await h.run(() => toggle(true));
    expect(animationStartCount()).toBe(before);
  } finally { await h.unmount(); }
});


it('settles in-flight chrome motion and preserves values across rerenders', async () => {
  const h = new MobileRenderHarness();
  let motion!: ReturnType<typeof useAnimatedComponents>;
  function Probe() { motion = useAnimatedComponents(); return null; }
  deferAnimationsForTest(true);
  try {
    await h.render(<Probe />);
    await h.run(() => setReduceMotionEnabledForTest(false));
    const headerY = motion[0][1].translateY;
    const footerY = motion[1][1].translateY;
    await h.run(() => motion[2](false));
    expect(pendingAnimationCount()).toBe(2);
    const stops = animationStopCount();
    await h.run(() => setReduceMotionEnabledForTest(true));
    expect(pendingAnimationCount()).toBe(0);
    expect(animationStopCount()).toBeGreaterThan(stops);
    expect(animatedValueForTest(headerY)).toBe(-300);
    expect(animatedValueForTest(footerY)).toBe(300);
    await h.render(<Probe />);
    expect(motion[0][1].translateY).toBe(headerY);
    expect(motion[1][1].translateY).toBe(footerY);
    await h.run(() => setReduceMotionEnabledForTest(false));
    await h.run(() => motion[2](true));
    expect(pendingAnimationCount()).toBe(2);
    await h.unmount();
    expect(pendingAnimationCount()).toBe(0);
  } finally { await h.unmount(); deferAnimationsForTest(false); }
});
