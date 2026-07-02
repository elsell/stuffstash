import { describe, expect, it } from 'vitest';
import { computeVoiceLevelBarHeights } from './VoiceLevelMeterPresentation';

describe('VoiceLevelMeterPresentation', () => {
  it('bounds voice levels and maps louder input to taller bars', () => {
    expect(computeVoiceLevelBarHeights(Number.NaN, 'regular')).toEqual(computeVoiceLevelBarHeights(0, 'regular'));
    expect(computeVoiceLevelBarHeights(-1, 'compact')).toEqual(computeVoiceLevelBarHeights(0, 'compact'));
    expect(computeVoiceLevelBarHeights(2, 'regular')).toEqual(computeVoiceLevelBarHeights(1, 'regular'));

    const quiet = computeVoiceLevelBarHeights(0, 'regular');
    const loud = computeVoiceLevelBarHeights(1, 'regular');

    expect(loud.every((height, index) => height >= (quiet[index] ?? 0))).toBe(true);
    expect(loud).not.toEqual(quiet);
  });
});
