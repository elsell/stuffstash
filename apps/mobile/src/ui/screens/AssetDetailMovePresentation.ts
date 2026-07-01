import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';

export type MoveDestinationCreateKind = 'container' | 'location';

export type MoveDestinationCreateInput = {
  readonly kind: MoveDestinationCreateKind;
  readonly title: string;
  readonly description: string;
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

export function canCreateMoveDestination({
  kind,
  matches,
  query
}: {
  readonly kind: MoveDestinationCreateKind;
  readonly matches: readonly Pick<ParentLookupResult, 'kind' | 'title'>[];
  readonly query: string;
}): boolean {
  const title = query.trim();
  if (title.length === 0) {
    return false;
  }
  return !matches.some((match) =>
    match.kind === kind && normalizeForMoveDestination(match.title) === normalizeForMoveDestination(title)
  );
}

export function moveDestinationCreateInput(kind: MoveDestinationCreateKind, title: string): MoveDestinationCreateInput {
  return {
    kind,
    title: title.trim(),
    description: ''
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

export function createdMoveDestinationParent({
  id,
  kind,
  title
}: {
  readonly id: string;
  readonly kind: MoveDestinationCreateKind;
  readonly title: string;
}): ParentLookupResult {
  const kindLabel = moveDestinationCreateKindLabel(kind);
  return {
    id,
    title,
    kind,
    subtitle: `New ${kind}`,
    pathLabel: title,
    selectionHint: kindLabel,
    willPromoteToContainer: false
  };
}

function normalizeForMoveDestination(value: string): string {
  return value.trim().toLocaleLowerCase();
}
