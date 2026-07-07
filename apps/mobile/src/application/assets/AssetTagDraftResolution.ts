import { assetTagKeyFromDisplayName } from '../../domain/assets/AssetSummary';

export type ActiveAssetTagReference = {
  readonly id: string;
  readonly key: string;
};

export type CreateAssetTagDraft = {
  readonly displayName: string;
  readonly color?: string;
};

export type AssetTagCreateRepository = {
  createAssetTag?: (input: CreateAssetTagDraft) => Promise<{ readonly id: string }>;
};

export type InlineAssetTagResolution =
  | { readonly status: 'empty' }
  | { readonly status: 'display_name_too_long' }
  | { readonly status: 'invalid_color' }
  | { readonly status: 'select_existing'; readonly tagId: string }
  | { readonly status: 'duplicate_pending' }
  | { readonly status: 'create'; readonly tag: CreateAssetTagDraft };

export type InlineAssetTagTransition = {
  readonly shouldClearInputs: boolean;
  readonly selectedTagIds: readonly string[];
  readonly pendingTags: readonly CreateAssetTagDraft[];
};

const maxDisplayNameLength = 80;

export function resolveInlineAssetTag(input: {
  readonly displayName: string;
  readonly color: string;
  readonly activeTags: readonly ActiveAssetTagReference[];
  readonly pendingTags: readonly CreateAssetTagDraft[];
}): InlineAssetTagResolution {
  const displayName = input.displayName.trim();
  const key = assetTagKeyFromDisplayName(displayName);
  if (displayName.length === 0 || key.length === 0) {
    return { status: 'empty' };
  }
  if (utf8ByteLength(displayName) > maxDisplayNameLength) {
    return { status: 'display_name_too_long' };
  }

  const existing = input.activeTags.find((tag) => tag.key === key);
  if (existing) {
    return { status: 'select_existing', tagId: existing.id };
  }

  if (input.pendingTags.some((tag) => assetTagKeyFromDisplayName(tag.displayName) === key)) {
    return { status: 'duplicate_pending' };
  }

  const color = normalizeTagColor(input.color);
  if (color === undefined && input.color.trim().length > 0) {
    return { status: 'invalid_color' };
  }
  return { status: 'create', tag: color ? { displayName, color } : { displayName } };
}

export function canResolveInlineAssetTag(input: {
  readonly displayName: string;
  readonly color: string;
  readonly activeTags: readonly ActiveAssetTagReference[];
  readonly pendingTags: readonly CreateAssetTagDraft[];
}): boolean {
  const resolution = resolveInlineAssetTag(input);
  return resolution.status === 'select_existing'
    || resolution.status === 'duplicate_pending'
    || resolution.status === 'create';
}

export function applyInlineAssetTagResolution(input: {
  readonly resolution: InlineAssetTagResolution;
  readonly selectedTagIds: readonly string[];
  readonly pendingTags: readonly CreateAssetTagDraft[];
}): InlineAssetTagTransition {
  if (input.resolution.status === 'select_existing') {
    return {
      shouldClearInputs: true,
      selectedTagIds: input.selectedTagIds.includes(input.resolution.tagId)
        ? input.selectedTagIds
        : [...input.selectedTagIds, input.resolution.tagId],
      pendingTags: input.pendingTags
    };
  }
  if (input.resolution.status === 'duplicate_pending') {
    return {
      shouldClearInputs: true,
      selectedTagIds: input.selectedTagIds,
      pendingTags: input.pendingTags
    };
  }
  if (input.resolution.status === 'create') {
    return {
      shouldClearInputs: true,
      selectedTagIds: input.selectedTagIds,
      pendingTags: [...input.pendingTags, input.resolution.tag]
    };
  }
  return {
    shouldClearInputs: false,
    selectedTagIds: input.selectedTagIds,
    pendingTags: input.pendingTags
  };
}

export function reconcileCreatedAssetTags(
  stagedTags: readonly CreateAssetTagDraft[],
  activeTags: readonly ActiveAssetTagReference[]
): {
  readonly createdTagIds: readonly string[];
  readonly remainingTags: readonly CreateAssetTagDraft[];
} {
  const activeByKey = new Map(activeTags.map((tag) => [tag.key, tag.id]));
  const createdTagIds: string[] = [];
  const remainingTags: CreateAssetTagDraft[] = [];
  for (const tag of stagedTags) {
    const createdTagId = activeByKey.get(assetTagKeyFromDisplayName(tag.displayName));
    if (createdTagId) {
      createdTagIds.push(createdTagId);
    } else {
      remainingTags.push(tag);
    }
  }
  return { createdTagIds, remainingTags };
}

export function reconcilePendingAssetTagDrafts(input: {
  readonly selectedTagIds: readonly string[];
  readonly pendingTags: readonly CreateAssetTagDraft[];
  readonly activeTags: readonly ActiveAssetTagReference[];
}): {
  readonly tagIds: readonly string[];
  readonly pendingTags: readonly CreateAssetTagDraft[];
} {
  const activeByKey = new Map(input.activeTags.map((tag) => [tag.key, tag.id]));
  const tagIds = new Set(input.selectedTagIds.map((tagId) => tagId.trim()).filter((tagId) => tagId.length > 0));
  const pendingTags: CreateAssetTagDraft[] = [];
  const pendingKeys = new Set<string>();

  for (const tag of input.pendingTags) {
    const displayName = tag.displayName.trim();
    const key = assetTagKeyFromDisplayName(displayName);
    if (displayName.length === 0 || key.length === 0) {
      continue;
    }
    const activeTagId = activeByKey.get(key);
    if (activeTagId) {
      tagIds.add(activeTagId);
      continue;
    }
    if (pendingKeys.has(key)) {
      continue;
    }
    pendingKeys.add(key);
    const color = normalizeTagColor(tag.color ?? '');
    pendingTags.push(color ? { displayName, color } : { displayName });
  }

  return { tagIds: Array.from(tagIds), pendingTags };
}

export async function createPendingAssetTags(
  repository: AssetTagCreateRepository,
  pendingTags: readonly CreateAssetTagDraft[]
): Promise<readonly string[]> {
  if (pendingTags.length === 0) {
    return [];
  }
  if (!repository.createAssetTag) {
    throw new Error('Tag creation is not available.');
  }
  const created = [];
  for (const tag of pendingTags) {
    const displayName = tag.displayName.trim();
    if (displayName.length === 0 || assetTagKeyFromDisplayName(displayName).length === 0) {
      continue;
    }
    const color = normalizeTagColor(tag.color ?? '');
    created.push(await repository.createAssetTag({
      displayName,
      color
    }));
  }
  return created.map((tag) => tag.id);
}

function normalizeTagColor(value: string): string | undefined {
  const trimmed = value.trim();
  if (trimmed.length === 0) {
    return undefined;
  }
  const match = /^#?([0-9a-fA-F]{6})$/.exec(trimmed);
  return match ? `#${match[1].toUpperCase()}` : undefined;
}

function utf8ByteLength(value: string): number {
  let bytes = 0;
  for (const character of value) {
    const codePoint = character.codePointAt(0) ?? 0;
    if (codePoint <= 0x7f) {
      bytes += 1;
    } else if (codePoint <= 0x7ff) {
      bytes += 2;
    } else if (codePoint <= 0xffff) {
      bytes += 3;
    } else {
      bytes += 4;
    }
  }
  return bytes;
}
