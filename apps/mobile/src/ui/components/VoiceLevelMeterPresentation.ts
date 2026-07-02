export type VoiceLevelMeterSize = 'compact' | 'regular';

export function computeVoiceLevelBarHeights(level: number, size: VoiceLevelMeterSize): readonly number[] {
  const boundedLevel = boundedVoiceLevel(level);
  const maxHeight = size === 'compact' ? 18 : 22;
  const minHeight = size === 'compact' ? 4 : 6;
  const barScales = [
    0.42 + boundedLevel * 0.58,
    0.62 + boundedLevel * 0.38,
    0.36 + boundedLevel * 0.64
  ];
  return barScales.map((scale) => Math.max(minHeight, Math.round(scale * maxHeight)));
}

function boundedVoiceLevel(value: number): number {
  if (!Number.isFinite(value)) {
    return 0;
  }
  return Math.max(0, Math.min(1, value));
}
