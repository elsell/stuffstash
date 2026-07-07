import { assetTagKeyFromDisplayName } from '../../domain/assets/AssetSummary';
import type { CreateInventoryAssetTagInput } from '../home/InventorySummaryRepository';

export type ActiveAssetTagReference = {
  readonly id: string;
  readonly key: string;
};

export type InlineAssetTagResolution =
  | { readonly status: 'empty' }
  | { readonly status: 'display_name_too_long' }
  | { readonly status: 'invalid_color' }
  | { readonly status: 'select_existing'; readonly tagId: string }
  | { readonly status: 'duplicate_pending' }
  | { readonly status: 'create'; readonly tag: CreateInventoryAssetTagInput };

export type InlineAssetTagTransition = {
  readonly shouldClearInputs: boolean;
  readonly selectedTagIds: readonly string[];
  readonly pendingTags: readonly CreateInventoryAssetTagInput[];
};

const maxDisplayNameLength = 80;

export function resolveInlineAssetTag(input: {
  readonly displayName: string;
  readonly color: string;
  readonly activeTags: readonly ActiveAssetTagReference[];
  readonly pendingTags: readonly CreateInventoryAssetTagInput[];
}): InlineAssetTagResolution {
  const displayName = input.displayName.trim();
  const key = assetTagKeyFromDisplayName(displayName);
  if (displayName.length === 0 || key.length === 0) {
    return { status: 'empty' };
  }
  if (displayName.length > maxDisplayNameLength) {
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
  readonly pendingTags: readonly CreateInventoryAssetTagInput[];
}): boolean {
  const resolution = resolveInlineAssetTag(input);
  return resolution.status === 'select_existing'
    || resolution.status === 'duplicate_pending'
    || resolution.status === 'create';
}

export function applyInlineAssetTagResolution(input: {
  readonly resolution: InlineAssetTagResolution;
  readonly selectedTagIds: readonly string[];
  readonly pendingTags: readonly CreateInventoryAssetTagInput[];
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
  stagedTags: readonly CreateInventoryAssetTagInput[],
  activeTags: readonly ActiveAssetTagReference[]
): {
  readonly createdTagIds: readonly string[];
  readonly remainingTags: readonly CreateInventoryAssetTagInput[];
} {
  const activeByKey = new Map(activeTags.map((tag) => [tag.key, tag.id]));
  const createdTagIds: string[] = [];
  const remainingTags: CreateInventoryAssetTagInput[] = [];
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

function normalizeTagColor(value: string): string | undefined {
  const trimmed = value.trim();
  if (trimmed.length === 0) {
    return undefined;
  }
  const match = /^#?([0-9a-fA-F]{6})$/.exec(trimmed);
  return match ? `#${match[1].toUpperCase()}` : undefined;
}
