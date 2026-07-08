import { describe, expect, it } from 'vitest';
import { assetTagChipLayoutPresentation, assetTagChipPresentation, assetTagChipStylePresentation } from './AssetTagChipsPresentation';

describe('assetTagChipPresentation', () => {
  it('shows every tag in full detail contexts', () => {
    expect(assetTagChipPresentation([tag('tools'), tag('camping'), tag('kids')])).toEqual({
      visibleTags: [tag('tools'), tag('camping'), tag('kids')],
      hiddenCount: 0,
      shouldRender: true
    });
  });

  it('keeps compact density independent from overflow summarization', () => {
    expect(assetTagChipPresentation([tag('tools'), tag('camping'), tag('kids')])).toEqual({
      visibleTags: [tag('tools'), tag('camping'), tag('kids')],
      hiddenCount: 0,
      shouldRender: true
    });
  });

  it('caps card contexts and reports the overflow count when requested', () => {
    expect(assetTagChipPresentation([tag('tools'), tag('camping'), tag('kids'), tag('garage')], 2)).toEqual({
      visibleTags: [tag('tools'), tag('camping')],
      hiddenCount: 2,
      shouldRender: true
    });
  });

  it('can summarize every tag into overflow when no visible chips fit', () => {
    expect(assetTagChipPresentation([tag('tools'), tag('camping')], 0)).toEqual({
      visibleTags: [],
      hiddenCount: 2,
      shouldRender: true
    });
  });

  it('does not render an empty tag row when no tags are assigned', () => {
    expect(assetTagChipPresentation([], 0)).toEqual({
      visibleTags: [],
      hiddenCount: 0,
      shouldRender: false
    });
  });

  it('shrinks visible compact chips while keeping full detail rows wrapping', () => {
    expect(assetTagChipLayoutPresentation()).toEqual({
      compactRow: false,
      shrinkVisibleChips: false
    });
    expect(assetTagChipLayoutPresentation(true)).toEqual({
      compactRow: true,
      shrinkVisibleChips: true
    });
  });

  it('uses the tag color as the chip color treatment when present', () => {
    expect(assetTagChipStylePresentation(tag('tools'))).toEqual({
      colored: true,
      backgroundColor: 'rgba(47, 128, 237, 0.14)',
      borderColor: '#2F80ED'
    });
    expect(assetTagChipStylePresentation({})).toEqual({
      colored: false
    });
  });

  it('keeps very light tag colors readable against the surface', () => {
    expect(assetTagChipStylePresentation({ color: '#FFFFFF' })).toEqual({
      colored: true,
      backgroundColor: 'rgba(255, 255, 255, 0.14)',
      borderColor: '#D9E1E6'
    });
    expect(assetTagChipStylePresentation({ color: '#FFF7B0' })).toEqual({
      colored: true,
      backgroundColor: 'rgba(255, 247, 176, 0.14)',
      borderColor: '#D9E1E6'
    });
  });
});

function tag(id: string) {
  return {
    id: `tag-${id}`,
    label: id,
    color: '#2F80ED'
  };
}
