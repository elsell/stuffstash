import { useState } from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { VoiceResultRailNavigation } from './VoiceResultRailNavigation';

it('moves through cards once and prevents moving beyond either end', async () => {
  const h = new MobileRenderHarness(); const moves: number[] = [];
  function Rail() {
    const [position, setPosition] = useState(0);
    return <VoiceResultRailNavigation position={position} count={2} onMove={next => { moves.push(next); setPosition(next); }} />;
  }
  try {
    await h.render(<Rail />);
    expect(h.byText('1 of 2')).toBeDefined();
    expect(h.byLabel('Previous')?.props.accessibilityState.disabled).toBe(true);
    await h.press(h.byLabel('Previous')); expect(moves).toEqual([]);
    await h.press(h.byLabel('Next')); expect(moves).toEqual([1]);
    expect(h.byText('2 of 2')).toBeDefined();
    expect(h.byLabel('Next')?.props.accessibilityState.disabled).toBe(true);
    await h.press(h.byLabel('Next')); expect(moves).toEqual([1]);
    await h.press(h.byLabel('Previous')); expect(moves).toEqual([1, 0]);
  } finally { await h.unmount(); }
});
