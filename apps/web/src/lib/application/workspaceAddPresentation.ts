import type { AssetKind, ParentTargetViewModel } from '$lib/domain/inventory';
import { assetKindLabel, assetKinds } from '$lib/domain/inventory';

export interface AddAssetKindCopy {
  heading: string;
  kindLabel: string;
  nameLabel: string;
  namePlaceholder: string;
  saveLabel: string;
  selectedKindLabel: string;
}

export interface AddControlOption<TValue extends string = string> {
  value: TValue;
  label: string;
}

export const quickParentKindOptions: AddControlOption<'location' | 'container'>[] = [
  { value: 'location', label: 'Location' },
  { value: 'container', label: 'Container' }
];

export function assetKindControlOptions(): AddControlOption<AssetKind>[] {
  return assetKinds.map((kind) => ({ value: kind, label: assetKindLabel(kind) }));
}

export function addAssetKindCopy(kind: AssetKind): AddAssetKindCopy {
  const kindLabel = assetKindLabel(kind);
  const selectedKindLabel = kindLabel.toLowerCase();
  return {
    heading: `Add ${selectedKindLabel}`,
    kindLabel,
    nameLabel: `${kindLabel} name`,
    namePlaceholder: addAssetNamePlaceholder(kind),
    saveLabel: `Save ${selectedKindLabel}`,
    selectedKindLabel
  };
}

export function addDestinationSummary(input: {
  quickParentEnabled: boolean;
  quickParentKind: 'location' | 'container';
  quickParentTitle: string;
  selectedParent: ParentTargetViewModel | null;
}): string {
  if (!input.quickParentEnabled) {
    return input.selectedParent?.title ?? 'Inventory root';
  }

  const parentKindLabel = assetKindLabel(input.quickParentKind);
  const parentName = input.quickParentTitle.trim() ? `New ${parentKindLabel}: ${input.quickParentTitle.trim()}` : `New ${parentKindLabel}`;
  return `${parentName} in ${quickParentContainerSummary(input.selectedParent)}`;
}

export function quickParentContainerLabel(selectedParent: ParentTargetViewModel | null): string {
  return selectedParent?.title ?? 'Inventory root';
}

export function quickParentContainerTrail(selectedParent: ParentTargetViewModel | null): string {
  return selectedParent?.containmentTrail ?? '';
}

export function quickParentContainerSummary(selectedParent: ParentTargetViewModel | null): string {
  return selectedParent ? `${selectedParent.title} / ${selectedParent.containmentTrail}` : 'Inventory root';
}

export function addPhotoCountLabel(photoCount: number): string {
  if (photoCount === 0) {
    return 'No photos';
  }
  return `${photoCount} ${photoCount === 1 ? 'photo' : 'photos'}`;
}

function addAssetNamePlaceholder(kind: AssetKind): string {
  if (kind === 'location') {
    return 'Garage shelf';
  }
  if (kind === 'container') {
    return 'Clear storage bin';
  }
  return 'Tomato fertilizer';
}
