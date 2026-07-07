import { describe, expect, it } from 'vitest';
import {
  applyInlineAssetTagResolution,
  canResolveInlineAssetTag,
  createPendingAssetTags,
  reconcileCreatedAssetTags,
  reconcilePendingAssetTagDrafts,
  resolveInlineAssetTag
} from './AssetTagDraftResolution';

describe('asset tag draft resolution', () => {
  it('selects existing tags by backend key even with invalid color text', () => {
    const resolution = resolveInlineAssetTag({
      displayName: 'Camp / Kitchen',
      color: 'blue',
      activeTags: [{ id: 'tag-camp-kitchen', key: 'camp-kitchen' }],
      pendingTags: []
    });

    expect(resolution).toEqual({ status: 'select_existing', tagId: 'tag-camp-kitchen' });
    expect(canResolveInlineAssetTag({
      displayName: 'Camp / Kitchen',
      color: 'blue',
      activeTags: [{ id: 'tag-camp-kitchen', key: 'camp-kitchen' }],
      pendingTags: []
    })).toBe(true);
  });

  it('does not create duplicate pending tags with equivalent keys', () => {
    expect(resolveInlineAssetTag({
      displayName: 'Camp / Kitchen',
      color: '',
      activeTags: [],
      pendingTags: [{ displayName: 'Camp Kitchen' }]
    })).toEqual({ status: 'duplicate_pending' });
  });

  it('normalizes valid colors and rejects invalid colors for new tags', () => {
    expect(resolveInlineAssetTag({
      displayName: 'Travel',
      color: '2f80ed',
      activeTags: [],
      pendingTags: []
    })).toEqual({ status: 'create', tag: { displayName: 'Travel', color: '#2F80ED' } });

    expect(resolveInlineAssetTag({
      displayName: 'Travel',
      color: 'blue',
      activeTags: [],
      pendingTags: []
    })).toEqual({ status: 'invalid_color' });
  });

  it('keeps display names within the backend contract length', () => {
    expect(resolveInlineAssetTag({
      displayName: 'a'.repeat(80),
      color: '',
      activeTags: [],
      pendingTags: []
    })).toEqual({ status: 'create', tag: { displayName: 'a'.repeat(80) } });

    expect(resolveInlineAssetTag({
      displayName: 'a'.repeat(81),
      color: '',
      activeTags: [],
      pendingTags: []
    })).toEqual({ status: 'display_name_too_long' });

    expect(resolveInlineAssetTag({
      displayName: `${'a'.repeat(79)}é`,
      color: '',
      activeTags: [],
      pendingTags: []
    })).toEqual({ status: 'display_name_too_long' });
  });

  it('applies picker transitions consistently', () => {
    expect(applyInlineAssetTagResolution({
      resolution: { status: 'select_existing', tagId: 'tag-camp-kitchen' },
      selectedTagIds: [],
      pendingTags: []
    })).toEqual({
      shouldClearInputs: true,
      selectedTagIds: ['tag-camp-kitchen'],
      pendingTags: []
    });

    expect(applyInlineAssetTagResolution({
      resolution: { status: 'create', tag: { displayName: 'Travel' } },
      selectedTagIds: ['tag-camp-kitchen'],
      pendingTags: []
    })).toEqual({
      shouldClearInputs: true,
      selectedTagIds: ['tag-camp-kitchen'],
      pendingTags: [{ displayName: 'Travel' }]
    });
  });

  it('reconciles staged tags to active tag IDs by backend key', () => {
    expect(reconcileCreatedAssetTags(
      [{ displayName: 'Camp / Kitchen' }, { displayName: 'Uncreated' }],
      [{ id: 'tag-camp-kitchen', key: 'camp-kitchen' }]
    )).toEqual({
      createdTagIds: ['tag-camp-kitchen'],
      remainingTags: [{ displayName: 'Uncreated' }]
    });
  });

  it('reuses active tags before save and leaves only truly new drafts to create', () => {
    expect(reconcilePendingAssetTagDrafts({
      selectedTagIds: [' tag-selected ', ''],
      pendingTags: [
        { displayName: ' Camp / Kitchen ', color: ' #2f80ed ' },
        { displayName: 'Travel' },
        { displayName: 'travel' },
        { displayName: ' ' }
      ],
      activeTags: [{ id: 'tag-camp-kitchen', key: 'camp-kitchen' }]
    })).toEqual({
      tagIds: ['tag-selected', 'tag-camp-kitchen'],
      pendingTags: [{ displayName: 'Travel' }]
    });
  });

  it('drops invalid stale colors before creating pending tags', () => {
    expect(reconcilePendingAssetTagDrafts({
      selectedTagIds: [],
      pendingTags: [
        { displayName: ' Travel ', color: ' blue ' },
        { displayName: 'Medical', color: '00aa88' }
      ],
      activeTags: []
    })).toEqual({
      tagIds: [],
      pendingTags: [
        { displayName: 'Travel' },
        { displayName: 'Medical', color: '#00AA88' }
      ]
    });
  });

  it('sanitizes drafts at the tag creation boundary', async () => {
    const repository = new FakeAssetTagCreateRepository();

    await expect(createPendingAssetTags(repository, [
      { displayName: ' ### ', color: '#2f80ed' },
      { displayName: ' Travel ', color: ' blue ' },
      { displayName: 'Medical', color: '00aa88' }
    ])).resolves.toEqual(['tag-created-1', 'tag-created-2']);

    expect(repository.createdTags).toEqual([
      { displayName: 'Travel' },
      { displayName: 'Medical', color: '#00AA88' }
    ]);
  });
});

class FakeAssetTagCreateRepository {
  readonly createdTags: Array<{ readonly displayName: string; readonly color?: string }> = [];

  async createAssetTag(input: { readonly displayName: string; readonly color?: string }) {
    this.createdTags.push(input);
    return { id: `tag-created-${this.createdTags.length.toString()}` };
  }
}
