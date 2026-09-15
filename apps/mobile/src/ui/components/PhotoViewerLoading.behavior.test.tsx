import React from 'react';
import { beforeEach, expect, it } from 'vitest';
import useImageLoad from 'react-native-image-viewing/dist/hooks/useImageLoad';
import { imageSizeRequestsForTest, resetNativeTestState } from '../../test-support/react-native';
import { MobileRenderHarness } from '../../test-support/render';
beforeEach(resetNativeTestState);

it('recovers from decode failure with a fresh request and rejects stale image events', async () => {
  const h = new MobileRenderHarness();
  const source = { uri: 'https://photo.invalid/decode', headers: { Authorization: 'synthetic' } };
  let load!: ReturnType<typeof useImageLoad>;
  function Probe() { load = useImageLoad(source); return null; }
  try {
    await h.render(<Probe />);
    await h.run(() => imageSizeRequestsForTest()[0].succeed(640, 480));
    const oldEvents = load;
    await h.run(() => load.onError());
    expect(load.failed).toBe(true);
    await h.run(() => oldEvents.onLoad());
    expect(load.failed).toBe(true);
    const oldAttempt = load.attempt;
    await h.run(() => load.retry());
    expect(load.failed).toBe(false);
    expect(load.loaded).toBe(false);
    expect(load.attempt).not.toBe(oldAttempt);
    expect(imageSizeRequestsForTest()).toHaveLength(2);
    expect(imageSizeRequestsForTest()[1].headers).toEqual(source.headers);
    await h.run(() => imageSizeRequestsForTest()[1].succeed(640, 480));
    await h.run(() => load.onLoad());
    expect(load.loaded).toBe(true);
    await h.run(() => oldEvents.onError());
    expect(load.loaded).toBe(true);
    expect(load.failed).toBe(false);
  } finally { await h.unmount(); }
});

it('stops loading on dimension failure without retrying automatically', async () => {
  const h = new MobileRenderHarness();
  const source = { uri: 'https://photo.invalid/no-dimensions' };
  let load!: ReturnType<typeof useImageLoad>;
  function Probe() { load = useImageLoad(source); return null; }
  try {
    await h.render(<Probe />);
    await h.run(() => imageSizeRequestsForTest()[0].fail());
    expect(load.failed).toBe(true);
    await h.settle();
    expect(imageSizeRequestsForTest()).toHaveLength(1);
    await h.run(() => load.retry());
    expect(load.failed).toBe(false);
    expect(imageSizeRequestsForTest()).toHaveLength(2);
  } finally { await h.unmount(); }
});

it('isolates source changes and survives effect replay in Strict Mode', async () => {
  const h = new MobileRenderHarness();
  const first = { uri: 'https://photo.invalid/strict-first' };
  const next = { uri: 'https://photo.invalid/strict-next' };
  let load!: ReturnType<typeof useImageLoad>;
  function Probe({ source }: { source: typeof first }) { load = useImageLoad(source); return null; }
  try {
    await h.render(<React.StrictMode><Probe source={first} /></React.StrictMode>);
    await h.run(() => imageSizeRequestsForTest().at(-1)!.succeed(100, 100));
    await h.run(() => load.onLoad());
    expect(load.loaded).toBe(true);
    const old = load;
    await h.render(<React.StrictMode><Probe source={next} /></React.StrictMode>);
    expect(load.loaded).toBe(false);
    await h.run(() => old.onError());
    expect(load.failed).toBe(false);
    await h.run(() => imageSizeRequestsForTest().at(-1)!.succeed(200, 200));
    await h.run(() => load.onLoad());
    expect(load.loaded).toBe(true);
  } finally { await h.unmount(); }
});
