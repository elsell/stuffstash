import { describe, expect, it } from 'vitest';
import { assetTagChipLayoutPresentation, assetTagChipPresentation } from './AssetTagChipsPresentation';

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
});

function tag(id: string) {
  return {
    id: `tag-${id}`,
    label: id,
    color: '#2F80ED'
  };
}
