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
      newTags: [{ displayName: 'Camp / Kitchen', color: '#2F80ED' }]
    });
  });

  it('drops invalid stale colors before creating pending tags', () => {
    const result = reconcilePendingAssetTagDrafts(
      [],
      [],
      [
        { displayName: ' Travel ', color: ' blue ' },
        { displayName: 'Medical', color: '00aa88' }
      ]
    );

    expect(result).toEqual({
      tagIds: [],
      newTags: [
        { displayName: 'Travel' },
        { displayName: 'Medical', color: '#00AA88' }
      ]
    });
  });
});

function tag(id: string, key: string, displayName: string): AssetTag {
  return { id, key, displayName };
}
