import { describe, expect, it } from 'vitest';
import { buildVoiceResponseEntityLinks, voiceResponseEntityOpenLabel } from './VoiceResponseEntityLinks';

describe('VoiceResponseEntityLinks', () => {
  it('links exact entity titles while preserving the response text', () => {
    const presentation = buildVoiceResponseEntityLinks('The Water bottle is in the Office.', [
      { type: 'asset_reference', assetId: 'water-bottle', title: 'Water bottle', assetKind: 'item' },
      { type: 'asset_reference', assetId: 'office', title: 'Office', assetKind: 'location' }
    ]);

    expect(presentation.segments.map((segment) => segment.text).join('')).toBe('The Water bottle is in the Office.');
    expect(presentation.segments.filter((segment) => segment.reference).map((segment) => segment.reference?.assetId)).toEqual([
      'water-bottle',
      'office'
    ]);
    expect(presentation.fallbackReferences).toEqual([]);
  });

  it('prefers longer non-overlapping titles and keeps unplaced references available', () => {
    const presentation = buildVoiceResponseEntityLinks('The Garage shelf is clear.', [
      { type: 'asset_reference', assetId: 'garage', title: 'Garage', assetKind: 'location' },
      { type: 'asset_reference', assetId: 'garage-shelf', title: 'Garage shelf', assetKind: 'container' },
      { type: 'asset_reference', assetId: 'drill', title: 'Drill', assetKind: 'item' }
    ]);

    expect(presentation.segments.filter((segment) => segment.reference).map((segment) => segment.reference?.assetId)).toEqual(['garage-shelf']);
    expect(presentation.fallbackReferences.map((reference) => reference.assetId)).toEqual(['garage', 'drill']);
  });

  it('does not assign duplicate titles to an arbitrary asset', () => {
    const references = [
      { type: 'asset_reference' as const, assetId: 'drill-one', title: 'Drill', assetKind: 'item' as const, context: 'Garage toolbox' },
      { type: 'asset_reference' as const, assetId: 'drill-two', title: 'Drill', assetKind: 'item' as const, context: 'Basement cabinet' }
    ];
    const presentation = buildVoiceResponseEntityLinks('Did you mean the Drill?', references);

    expect(presentation.segments.some((segment) => segment.reference)).toBe(false);
    expect(presentation.fallbackReferences).toEqual(references);
    expect(presentation.fallbackReferences.map((reference) => voiceResponseEntityOpenLabel(reference, references))).toEqual([
      'Open Drill in Garage toolbox',
      'Open Drill in Basement cabinet'
    ]);
  });

  it('preserves original offsets when Unicode case folding changes string length', () => {
    const presentation = buildVoiceResponseEntityLinks('The İSTANBUL BOX is in Garage.', [
      { type: 'asset_reference', assetId: 'box', title: 'İSTANBUL BOX', assetKind: 'container' },
      { type: 'asset_reference', assetId: 'garage', title: 'Garage', assetKind: 'location' }
    ]);

    expect(presentation.segments.map((segment) => segment.text).join('')).toBe('The İSTANBUL BOX is in Garage.');
    expect(presentation.segments.filter((segment) => segment.reference).map((segment) => segment.reference?.assetId)).toEqual(['box', 'garage']);
  });

  it('does not turn a case-distinct ordinary word into an entity link', () => {
    const references = [{ type: 'asset_reference' as const, assetId: 'us', title: 'US', assetKind: 'item' as const }];
    const presentation = buildVoiceResponseEntityLinks('Please help us find it.', references);

    expect(presentation.segments.some((segment) => segment.reference)).toBe(false);
    expect(presentation.fallbackReferences).toEqual(references);
  });

  it('adds stable ordinals when duplicate title and context are identical', () => {
    const references = [
      { type: 'asset_reference' as const, assetId: 'drill-one', title: 'Drill', assetKind: 'item' as const, context: 'Toolbox' },
      { type: 'asset_reference' as const, assetId: 'drill-two', title: 'Drill', assetKind: 'item' as const, context: 'Toolbox' }
    ];

    expect(references.map((reference) => voiceResponseEntityOpenLabel(reference, references))).toEqual([
      'Open Drill in Toolbox (1 of 2)',
      'Open Drill in Toolbox (2 of 2)'
    ]);
  });
});
