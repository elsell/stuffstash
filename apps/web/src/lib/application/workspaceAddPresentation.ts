import type { AssetKind, MediaUploadPolicy, ParentTargetViewModel, SelectedPhoto } from '$lib/domain/inventory';
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

export interface AddPhotoPickerPresentation {
  actionGroupLabel: string;
  uploadLabel: string;
  cameraLabel: string;
  uploadInputLabel: string;
  cameraInputLabel: string;
  selectedListLabel: string;
}

export interface AddFormPresentation {
  summaryTypeLabel: string;
  summaryParentLabel: string;
  summaryPhotosLabel: string;
  assetKindLegend: string;
  parentPickerLegend: string;
  parentPickerGroupLabel: string;
  quickParentLegend: string;
  quickParentToggleLabel: string;
  quickParentToggleDescription: string;
  quickParentContextLabel: string;
  quickParentNameLabel: string;
  quickParentNamePlaceholder: string;
  quickParentKindLabel: string;
  descriptionLabel: string;
  descriptionPlaceholder: string;
}

export const addPhotoPickerPresentation: AddPhotoPickerPresentation = {
  actionGroupLabel: 'Photo actions',
  uploadLabel: 'Upload',
  cameraLabel: 'Camera',
  uploadInputLabel: 'Upload photos',
  cameraInputLabel: 'Take photo',
  selectedListLabel: 'Selected photos'
};

export const addFormPresentation: AddFormPresentation = {
  summaryTypeLabel: 'Type',
  summaryParentLabel: 'Parent',
  summaryPhotosLabel: 'Photos',
  assetKindLegend: 'Asset kind',
  parentPickerLegend: 'Place in existing parent',
  parentPickerGroupLabel: 'Parent target',
  quickParentLegend: 'Create missing parent',
  quickParentToggleLabel: 'Create a parent first',
  quickParentToggleDescription: 'Use this when the shelf, box, or location does not exist yet.',
  quickParentContextLabel: 'Created under',
  quickParentNameLabel: 'Parent name',
  quickParentNamePlaceholder: 'Laundry shelf',
  quickParentKindLabel: 'New parent kind',
  descriptionLabel: 'Description',
  descriptionPlaceholder: 'Optional notes'
};

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

export function quickParentMissingNameMessage(): string {
  return 'Enter a parent name or turn this option off.';
}

export function addPhotoCountLabel(photoCount: number): string {
  if (photoCount === 0) {
    return 'No photos';
  }
  return `${photoCount} ${photoCount === 1 ? 'photo' : 'photos'}`;
}

export function addSupportedImageTypes(mediaPolicy: MediaUploadPolicy): SelectedPhoto['contentType'][] {
  return mediaPolicy.supportedContentTypes.filter((type): type is SelectedPhoto['contentType'] => type.startsWith('image/'));
}

export function addPhotoAcceptTypes(supportedImageTypes: string[]): string {
  return supportedImageTypes.join(',');
}

export function addPhotoSupportedTypeLabel(types: string[]): string {
  if (types.length === 0) {
    return 'No image formats';
  }
  const labels = types.map(formatImageContentType);
  if (labels.length === 1) {
    return labels[0] ?? '';
  }
  if (labels.length === 2) {
    return `${labels[0]} or ${labels[1]}`;
  }
  return `${labels.slice(0, -1).join(', ')}, or ${labels[labels.length - 1]}`;
}

export function addPhotoHelpText(supportedTypeLabel: string, maxBytesLabel: string): string {
  return `Optional ${supportedTypeLabel} up to ${maxBytesLabel}.`;
}

export function addPhotoRemoveLabel(photo: Pick<SelectedPhoto, 'name'>): string {
  return `Remove ${photo.name}`;
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

function formatImageContentType(type: string): string {
  if (type === 'image/jpeg') return 'JPEG';
  if (type === 'image/png') return 'PNG';
  if (type === 'image/webp') return 'WebP';
  return type.replace(/^image\//, '').toUpperCase();
}
