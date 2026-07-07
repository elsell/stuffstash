import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { CreateAssetTagDraft } from '../../application/assets/AssetTagDraftResolution';

export type EditDraft = {
  readonly title: string;
  readonly description: string;
  readonly tagIds?: readonly string[];
  readonly newTags?: readonly CreateAssetTagDraft[];
};

export type AssetEditContext = {
  readonly kindLabel: string;
  readonly customTypeLabel?: string;
  readonly helperText: string;
};

export type NormalizedEditDraft = {
  readonly title: string;
  readonly description: string;
  readonly tagIds?: readonly string[];
  readonly newTags?: readonly CreateAssetTagDraft[];
};

export function canSaveEditAsset(
  asset: Pick<AssetDetailViewModel, 'title' | 'description' | 'tags'>,
  draft: EditDraft | undefined
): boolean {
  return editDraftState(asset, draft).canSave;
}

export function hasDirtyEditAssetDraft(
  asset: Pick<AssetDetailViewModel, 'title' | 'description' | 'tags'>,
  draft: EditDraft | undefined
): boolean {
  return editDraftState(asset, draft).isDirty;
}

export function normalizedEditDraft(draft: EditDraft): NormalizedEditDraft {
  return {
    title: draft.title.trim(),
    description: draft.description.trim(),
    tagIds: normalizeTagIds(draft.tagIds),
    newTags: normalizeNewTags(draft.newTags)
  };
}

function editDraftState(
  asset: Pick<AssetDetailViewModel, 'title' | 'description' | 'tags'>,
  draft: EditDraft | undefined
): {
  readonly canSave: boolean;
  readonly isDirty: boolean;
} {
  if (!draft) {
    return { canSave: false, isDirty: false };
  }

  const normalized = normalizedEditDraft(draft);
  const isDirty = normalized.title !== asset.title
    || normalized.description !== asset.description
    || !sameTagAssignments(normalized.tagIds ?? [], asset.tags?.map((tag) => tag.id) ?? [])
    || (normalized.newTags?.length ?? 0) > 0;

  return {
    canSave: normalized.title.length > 0 && isDirty,
    isDirty
  };
}

function normalizeTagIds(tagIds: readonly string[] | undefined): readonly string[] {
  return (tagIds ?? []).map((tagId) => tagId.trim()).filter((tagId) => tagId.length > 0);
}

function normalizeNewTags(newTags: readonly CreateAssetTagDraft[] | undefined): readonly CreateAssetTagDraft[] {
  return (newTags ?? [])
    .map((tag) => {
      const displayName = tag.displayName.trim();
      const color = tag.color?.trim();
      return color ? { displayName, color } : { displayName };
    })
    .filter((tag) => tag.displayName.length > 0);
}

function sameTagAssignments(left: readonly string[], right: readonly string[]): boolean {
  if (left.length !== right.length) {
    return false;
  }
  const rightSet = new Set(right);
  return left.every((tagId) => rightSet.has(tagId));
}

export function assetEditContext(
  asset: Pick<AssetDetailViewModel, 'kindLabel' | 'customTypeLabel'>
): AssetEditContext {
  return {
    kindLabel: asset.kindLabel,
    customTypeLabel: asset.customTypeLabel,
    helperText: 'Kind and type changes need a future conversion flow.'
  };
}
