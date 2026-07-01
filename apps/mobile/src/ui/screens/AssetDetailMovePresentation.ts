import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';

export type MoveDestinationCreateKind = 'container' | 'location';

export type MoveDestinationCreateInput = {
  readonly kind: MoveDestinationCreateKind;
  readonly title: string;
  readonly description: string;
  readonly parentAssetId?: string;
};

export type MoveDestinationCreatePlacement = {
  readonly parentAssetId?: string;
  readonly parentPathLabel?: string;
};

export function canSaveMoveAsset(
  asset: Pick<AssetDetailViewModel, 'parentAssetId'>,
  selectedParent: ParentLookupResult | null
): boolean {
  return (selectedParent?.id ?? null) !== (asset.parentAssetId ?? null);
}

export function parentFromCurrentAssetPath(asset: AssetDetailViewModel): ParentLookupResult | null {
  if (!asset.parentAssetId) {
    return null;
  }
  return {
    id: asset.parentAssetId,
    title: asset.parentLocationTrailLabel,
    kind: 'container',
    subtitle: 'Current parent',
    pathLabel: asset.parentLocationTrailLabel,
    selectionHint: 'Current parent',
    willPromoteToContainer: false
  };
}

export type MovePlacementPreview = {
  readonly currentLocationLabel: string;
  readonly proposedLocationLabel: string;
  readonly hasChanged: boolean;
};

export type MoveDestinationRow = {
  readonly title: string;
  readonly kindLabel: string;
  readonly pathLabel: string;
};

export type MoveIntoEmptyState = {
  readonly title: string;
  readonly message: string;
};

export function movePlacementPreview(
  asset: Pick<AssetDetailViewModel, 'parentAssetId' | 'parentLocationTrailLabel'>,
  selectedParent: ParentLookupResult | null
): MovePlacementPreview {
  const currentLocationLabel = asset.parentLocationTrailLabel;
  const proposedLocationLabel = selectedParent
    ? selectedParent.pathLabel
    : 'Inventory root';

  return {
    currentLocationLabel,
    proposedLocationLabel,
    hasChanged: canSaveMoveAsset(asset, selectedParent)
  };
}

export function moveDestinationRow(parent: ParentLookupResult): MoveDestinationRow {
  return {
    title: parent.title,
    kindLabel: parent.kind === 'location' ? 'Location' : 'Container',
    pathLabel: parent.pathLabel
  };
}

export function moveIntoCandidateRow(asset: ParentLookupResult): MoveDestinationRow {
  return {
    title: asset.title,
    kindLabel: moveCandidateKindLabel(asset.kind),
    pathLabel: asset.pathLabel
  };
}

export function isSelectableMoveDestination(parent: ParentLookupResult): boolean {
  return parent.kind === 'container' || parent.kind === 'location';
}

export function isSelectableMoveIntoCandidate(
  candidate: Pick<ParentLookupResult, 'id' | 'parentAssetId'>,
  target: Pick<AssetDetailViewModel, 'id' | 'containedAssets'>
): boolean {
  if (candidate.id === target.id) {
    return false;
  }
  if (candidate.parentAssetId === target.id) {
    return false;
  }
  return !target.containedAssets.some((child) => child.id === candidate.id);
}

export function moveIntoEmptyState(query: string): MoveIntoEmptyState {
  if (query.trim().length === 0) {
    return {
      title: 'Search for something to move here',
      message: 'Start typing to find an item, box, or place from this inventory.'
    };
  }
  return {
    title: 'No movable matches',
    message: 'Search for something that is not already inside this place.'
  };
}

export function canCreateMoveDestination({
  kind,
  matches,
  parentAssetId,
  query
}: {
  readonly kind: MoveDestinationCreateKind;
  readonly matches: readonly Pick<ParentLookupResult, 'kind' | 'parentAssetId' | 'title'>[];
  readonly parentAssetId?: string;
  readonly query: string;
}): boolean {
  const title = query.trim();
  if (title.length === 0) {
    return false;
  }
  return !matches.some((match) =>
    match.kind === kind
      && (match.parentAssetId ?? undefined) === (parentAssetId ?? undefined)
      && normalizeForMoveDestination(match.title) === normalizeForMoveDestination(title)
  );
}

export function moveDestinationCreatePlacement(
  asset: Pick<AssetDetailViewModel, 'parentAssetId' | 'parentLocationTrailLabel'>
): MoveDestinationCreatePlacement {
  if (!asset.parentAssetId) {
    return {};
  }

  return {
    parentAssetId: asset.parentAssetId,
    parentPathLabel: asset.parentLocationTrailLabel
  };
}

export function moveDestinationCreateInput(
  kind: MoveDestinationCreateKind,
  title: string,
  placement: MoveDestinationCreatePlacement
): MoveDestinationCreateInput {
  return {
    kind,
    title: title.trim(),
    description: '',
    ...(placement.parentAssetId ? { parentAssetId: placement.parentAssetId } : {})
  };
}

export function moveDestinationCreateKindLabel(kind: MoveDestinationCreateKind): string {
  return kind === 'location' ? 'Location' : 'Container';
}

export function moveDestinationCreateKindHelp(kind: MoveDestinationCreateKind): string {
  return kind === 'location'
    ? 'Best for rooms, places, and areas.'
    : 'Best for boxes, shelves, bins, and cabinets.';
}

export function moveDestinationCreateButtonLabel(kind: MoveDestinationCreateKind, title: string): string {
  return `Create ${kind} "${title}"`;
}

export function moveDestinationCreatePlacementLabel(placement: MoveDestinationCreatePlacement): string {
  return placement.parentPathLabel ? `Creates inside ${placement.parentPathLabel}` : 'Creates at inventory root';
}

export function createdMoveDestinationParent({
  id,
  kind,
  placement,
  title
}: {
  readonly id: string;
  readonly kind: MoveDestinationCreateKind;
  readonly placement: MoveDestinationCreatePlacement;
  readonly title: string;
}): ParentLookupResult {
  const kindLabel = moveDestinationCreateKindLabel(kind);
  return {
    id,
    title,
    kind,
    parentAssetId: placement.parentAssetId,
    subtitle: `New ${kind}`,
    pathLabel: placement.parentPathLabel ? `${placement.parentPathLabel} / ${title}` : title,
    selectionHint: kindLabel,
    willPromoteToContainer: false
  };
}

function normalizeForMoveDestination(value: string): string {
  return value.trim().toLocaleLowerCase();
}

function moveCandidateKindLabel(kind: ParentLookupResult['kind']): string {
  switch (kind) {
    case 'container':
      return 'Container';
    case 'item':
      return 'Item';
    case 'location':
      return 'Location';
  }
}
