import React from 'react';
import { expect, it } from 'vitest';
import useImageIndexChange from 'react-native-image-viewing/dist/hooks/useImageIndexChange';
import { MobileRenderHarness } from '../../test-support/render';

it('moves to external selections without replaying a native swipe when its parent echoes it', async () => {
  const h = new MobileRenderHarness();
  const commands: number[] = [];
  let state!: ReturnType<typeof useImageIndexChange>;
  const selectPage = (index: number) => { commands.push(index); };
  function Probe({ index }: { index: number }) {
    state = useImageIndexChange(index, { width: 400, height: 800 }, selectPage);
    return null;
  }
  try {
    await h.render(<Probe index={0} />);
    expect(state[0]).toBe(0);
    expect(commands).toEqual([]);
    await h.render(<Probe index={1} />);
    expect(state[0]).toBe(1);
    expect(commands).toEqual([1]);
    await h.run(() => state[1]({ nativeEvent: { contentOffset: { x: 800 } } } as Parameters<typeof state[1]>[0]));
    expect(state[0]).toBe(2);
    await h.render(<Probe index={2} />);
    expect(state[0]).toBe(2);
    expect(commands).toEqual([1]);
    await h.render(<Probe index={0} />);
    expect(state[0]).toBe(0);
    expect(commands).toEqual([1, 0]);
  } finally { await h.unmount(); }
});
