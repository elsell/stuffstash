import type { AssetKind } from '../../domain/assets/AssetSummary';
import type { AddAssetDraft } from '../../application/add/AddAssetDraftStore';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { ParentSelection } from './AddAssetResolution';

export type AddInitialParentParams = Readonly<Record<string, string | string[] | undefined>>;

export type AddInitialParentRouteParams = {
  readonly parentAssetId: string;
  readonly parentTitle: string;
  readonly parentKind: 'container' | 'location';
  readonly parentSubtitle: string;
  readonly parentPathLabel: string;
  readonly parentSelectionHint: string;
  readonly parentWillPromoteToContainer: 'false';
};

export function addHereParams(asset: AssetDetailViewModel): AddInitialParentRouteParams {
  if (asset.kind !== 'container' && asset.kind !== 'location') {
    throw new Error('Add-here parent must be a container or location.');
  }

  return {
    parentAssetId: asset.id,
    parentTitle: asset.title,
    parentKind: asset.kind,
    parentSubtitle: asset.parentLocationTrailLabel,
    parentPathLabel: asset.locationTrailLabel,
    parentSelectionHint: asset.kindLabel,
    parentWillPromoteToContainer: 'false'
  };
}

export function initialParentFromParams(params: AddInitialParentParams): ParentSelection | undefined {
  const id = singleParam(params.parentAssetId);
  const title = singleParam(params.parentTitle);
  const kind = assetKindParam(params.parentKind);
  const subtitle = singleParam(params.parentSubtitle);
  const pathLabel = singleParam(params.parentPathLabel);
  const selectionHint = singleParam(params.parentSelectionHint);
  const willPromoteToContainer = singleParam(params.parentWillPromoteToContainer);

  if (!id || !title || !kind || !subtitle || !pathLabel || !selectionHint || willPromoteToContainer !== 'false') {
    return undefined;
  }

  return {
    id,
    title,
    kind,
    subtitle,
    pathLabel,
    selectionHint,
    willPromoteToContainer: false
  };
}

export function applyInitialParentToDraft(
  draft: AddAssetDraft,
  initialParent: ParentSelection | undefined
): AddAssetDraft {
  if (!initialParent || draft.parentAssetId === initialParent.id) {
    return draft;
  }

  return {
    ...draft,
    parentAssetId: initialParent.id,
    parentQuery: initialParent.title,
    lastParent: initialParent
  };
}

function singleParam(value: string | string[] | undefined): string | undefined {
  const selected = Array.isArray(value) ? value[0] : value;
  const trimmed = selected?.trim() ?? '';
  return trimmed.length > 0 ? trimmed : undefined;
}

function assetKindParam(value: string | string[] | undefined): Extract<AssetKind, 'container' | 'location'> | undefined {
  const selected = singleParam(value);
  switch (selected) {
    case 'container':
    case 'location':
      return selected;
    default:
      return undefined;
  }
}
