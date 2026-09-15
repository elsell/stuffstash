import React from 'react';
import { beforeEach, expect, it } from 'vitest';
import useImageDimensions from 'react-native-image-viewing/dist/hooks/useImageDimensions';
import { imageSizeRequestsForTest, resetNativeTestState } from '../../test-support/react-native';
import { MobileRenderHarness } from '../../test-support/render';

beforeEach(resetNativeTestState);

const useDimensions = useImageDimensions;

it('retries failed dimensions with the original headers and ignores the old request', async () => {
  const h = new MobileRenderHarness();
  const source = { uri: 'https://photo.invalid/retry', headers: { Authorization: 'synthetic' } };
  let dimensions: ReturnType<typeof useDimensions>;
  function Probe({ attempt }: { attempt: number }) { dimensions = useDimensions(source, attempt); return null; }
  try {
    await h.render(<Probe attempt={0} />);
    const first = imageSizeRequestsForTest()[0];
    await h.run(() => first.fail());
    expect(dimensions!).toEqual({ width: 0, height: 0 });
    await h.render(<Probe attempt={1} />);
    expect(dimensions!).toBeNull();
    const requests = imageSizeRequestsForTest();
    expect(requests).toHaveLength(2);
    expect(requests[1].headers).toEqual(source.headers);
    await h.run(() => first.succeed(1, 1));
    expect(dimensions!).toBeNull();
    await h.run(() => requests[1].succeed(640, 480));
    expect(dimensions!).toEqual({ width: 640, height: 480 });
  } finally { await h.unmount(); }
});

it('clears old dimensions on source replacement and ignores its late completion', async () => {
  const h = new MobileRenderHarness();
  const firstSource = { uri: 'https://photo.invalid/first' };
  const secondSource = { uri: 'https://photo.invalid/second' };
  let dimensions: ReturnType<typeof useDimensions>;
  function Probe({ source }: { source: typeof firstSource }) { dimensions = useDimensions(source); return null; }
  try {
    await h.render(<Probe source={firstSource} />);
    const first = imageSizeRequestsForTest()[0];
    await h.run(() => first.succeed(100, 80));
    await h.render(<Probe source={secondSource} />);
    expect(dimensions!).toBeNull();
    await h.run(() => first.succeed(500, 400));
    expect(dimensions!).toBeNull();
    await h.run(() => imageSizeRequestsForTest()[1].fail());
    expect(dimensions!).toEqual({ width: 0, height: 0 });
  } finally { await h.unmount(); }
});

it('keeps the current dimensions when a superseded pending request finishes late', async () => {
  const h = new MobileRenderHarness();
  const oldSource = { uri: 'https://photo.invalid/pending-old' };
  const newSource = { uri: 'https://photo.invalid/pending-new' };
  let dimensions: ReturnType<typeof useDimensions>;
  function Probe({ source }: { source: typeof oldSource }) { dimensions = useDimensions(source); return null; }
  try {
    await h.render(<Probe source={oldSource} />);
    const oldRequest = imageSizeRequestsForTest()[0];
    await h.render(<Probe source={newSource} />);
    await h.run(() => imageSizeRequestsForTest()[1].succeed(640, 480));
    expect(dimensions!).toEqual({ width: 640, height: 480 });
    await h.run(() => oldRequest.succeed(20, 10));
    expect(dimensions!).toEqual({ width: 640, height: 480 });
  } finally { await h.unmount(); }
});
