import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';

export type EditDraft = {
  readonly title: string;
  readonly description: string;
};

export type AssetEditContext = {
  readonly kindLabel: string;
  readonly customTypeLabel?: string;
  readonly helperText: string;
};

export type NormalizedEditDraft = {
  readonly title: string;
  readonly description: string;
};

export function canSaveEditAsset(
  asset: Pick<AssetDetailViewModel, 'title' | 'description'>,
  draft: EditDraft | undefined
): boolean {
  return editDraftState(asset, draft).canSave;
}

export function hasDirtyEditAssetDraft(
  asset: Pick<AssetDetailViewModel, 'title' | 'description'>,
  draft: EditDraft | undefined
): boolean {
  return editDraftState(asset, draft).isDirty;
}

export function normalizedEditDraft(draft: EditDraft): NormalizedEditDraft {
  return {
    title: draft.title.trim(),
    description: draft.description.trim()
  };
}

function editDraftState(
  asset: Pick<AssetDetailViewModel, 'title' | 'description'>,
  draft: EditDraft | undefined
): {
  readonly canSave: boolean;
  readonly isDirty: boolean;
} {
  if (!draft) {
    return { canSave: false, isDirty: false };
  }

  const normalized = normalizedEditDraft(draft);
  const isDirty = normalized.title !== asset.title || normalized.description !== asset.description;

  return {
    canSave: normalized.title.length > 0 && isDirty,
    isDirty
  };
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
