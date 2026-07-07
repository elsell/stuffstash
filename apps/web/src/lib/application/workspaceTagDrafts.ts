import { assetTagKeyFromDisplayName, type AssetTag, type AssetTagDraft } from '$lib/domain/inventory';

export interface ReconciledAssetTagDrafts {
  tagIds: string[];
  newTags: AssetTagDraft[];
}

export function reconcilePendingAssetTagDrafts(
  existingTags: readonly AssetTag[],
  selectedTagIds: readonly string[],
  pendingTags: readonly AssetTagDraft[]
): ReconciledAssetTagDrafts {
  const existingByKey = new Map(existingTags.map((tag) => [tag.key, tag]));
  const tagIds = new Set(selectedTagIds.map((tagId) => tagId.trim()).filter((tagId) => tagId.length > 0));
  const newTags: AssetTagDraft[] = [];
  const pendingKeys = new Set<string>();

  for (const pending of pendingTags) {
    const displayName = pending.displayName.trim();
    if (displayName.length === 0) {
      continue;
    }
    const key = assetTagKeyFromDisplayName(displayName);
    if (key.length === 0) {
      continue;
    }
    const existing = existingByKey.get(key);
    if (existing) {
      tagIds.add(existing.id);
      continue;
    }
    if (pendingKeys.has(key)) {
      continue;
    }
    pendingKeys.add(key);
    const color = normalizeTagColor(pending.color ?? '');
    newTags.push(color ? { displayName, color } : { displayName });
  }

  return { tagIds: Array.from(tagIds), newTags };
}

function normalizeTagColor(value: string): string | undefined {
  const trimmed = value.trim();
  if (trimmed.length === 0) {
    return undefined;
  }
  const match = /^#?([0-9a-fA-F]{6})$/.exec(trimmed);
  return match ? `#${match[1].toUpperCase()}` : undefined;
}
