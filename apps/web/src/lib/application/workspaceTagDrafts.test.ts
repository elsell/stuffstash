import { describe, expect, it } from 'vitest';
import type { AssetTag } from '$lib/domain/inventory';
import { reconcilePendingAssetTagDrafts } from './workspaceTagDrafts';

describe('reconcilePendingAssetTagDrafts', () => {
  it('reuses known tags by normalized key instead of leaving duplicate drafts to create', () => {
    const result = reconcilePendingAssetTagDrafts(
      [tag('tag-workshop', 'workshop', 'Workshop')],
      ['tag-selected'],
      [{ displayName: ' workshop ' }]
    );

    expect(result).toEqual({
      tagIds: ['tag-selected', 'tag-workshop'],
      newTags: []
    });
  });

  it('normalizes pending drafts and drops duplicate pending keys', () => {
    const result = reconcilePendingAssetTagDrafts(
      [],
      [' tag-selected ', ''],
      [
        { displayName: ' Camp / Kitchen ', color: ' #2f80ed ' },
        { displayName: 'camp-kitchen' },
        { displayName: ' ' }
      ]
    );

    expect(result).toEqual({
      tagIds: ['tag-selected'],
      newTags: [{ displayName: 'Camp / Kitchen', color: '#2f80ed' }]
    });
  });
});

function tag(id: string, key: string, displayName: string): AssetTag {
  return { id, key, displayName };
}
