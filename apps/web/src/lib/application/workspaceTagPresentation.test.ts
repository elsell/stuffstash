import { describe, expect, it } from 'vitest';
import type { AssetTag } from '$lib/domain/inventory';
import { sortAssetTagsByDisplayName, visibleAssetTagOptions } from './workspaceTagPresentation';

describe('workspace tag presentation', () => {
  it('naturally sorts tag display names without mutating the source', () => {
    const tags = [tag('tag-10', 'Tag 10'), tag('tag-2', 'tag 2'), tag('tag-1', 'Tag 1')];

    expect(sortAssetTagsByDisplayName(tags).map(({ id }) => id)).toEqual(['tag-1', 'tag-2', 'tag-10']);
    expect(tags.map(({ id }) => id)).toEqual(['tag-10', 'tag-2', 'tag-1']);
  });

  it('keeps selected options visible when the naturally sorted list is collapsed', () => {
    const tags = Array.from({ length: 14 }, (_, index) => tag(`tag-${index + 1}`, `Tag ${index + 1}`));

    expect(visibleAssetTagOptions(tags, false, ['tag-14']).map(({ id }) => id)).toEqual([
      ...Array.from({ length: 12 }, (_, index) => `tag-${index + 1}`),
      'tag-14'
    ]);
  });

  it('returns every option in natural order when expanded', () => {
    const tags = [tag('tag-10', 'Tag 10'), tag('tag-2', 'Tag 2'), tag('tag-1', 'Tag 1')];

    expect(visibleAssetTagOptions(tags, true, ['tag-10'], 1).map(({ id }) => id)).toEqual(['tag-1', 'tag-2', 'tag-10']);
  });
});

function tag(id: string, displayName: string): AssetTag {
  return { id, key: id, displayName };
}
