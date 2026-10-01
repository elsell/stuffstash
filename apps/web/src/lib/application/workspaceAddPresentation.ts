import { t } from '$lib/presentation/localization';
import type { AssetKind, MediaUploadPolicy, ParentTargetViewModel, SelectedPhoto } from '$lib/domain/inventory';
import { assetKinds } from '$lib/domain/inventory';
import { assetKindLabel } from '$lib/presentation/assetKindLabel';

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
  actionGroupLabel: t('web.workspaceAddPresentation.photoActions'),
  uploadLabel: t('web.workspaceAddPresentation.choosePhotos'),
  cameraLabel: t('web.workspaceAddPresentation.takePhoto'),
  uploadInputLabel: t('web.workspaceAddPresentation.choosePhotos'),
  cameraInputLabel: t('web.workspaceAddPresentation.takePhoto'),
  selectedListLabel: t('web.workspaceAddPresentation.selectedPhotos')
};

export const addFormPresentation: AddFormPresentation = {
  summaryTypeLabel: t('web.workspaceAddPresentation.type'),
  summaryParentLabel: t('web.workspaceAddPresentation.parent'),
  summaryPhotosLabel: t('web.workspaceAddPresentation.photos'),
  assetKindLegend: t('web.workspaceAddPresentation.assetKind'),
  parentPickerLegend: t('web.workspaceAddPresentation.placeInExistingParent'),
  parentPickerGroupLabel: t('web.workspaceAddPresentation.parentTarget'),
  quickParentLegend: t('web.workspaceAddPresentation.createMissingParent'),
  quickParentToggleLabel: t('web.workspaceAddPresentation.createAParentFirst'),
  quickParentToggleDescription: t('web.workspaceAddPresentation.useThisWhenTheShelfBoxOrLocationDoes'),
  quickParentContextLabel: t('web.workspaceAddPresentation.createdUnder'),
  quickParentNameLabel: t('web.workspaceAddPresentation.parentName'),
  quickParentNamePlaceholder: t('web.workspaceAddPresentation.laundryShelf'),
  quickParentKindLabel: t('web.workspaceAddPresentation.newParentKind'),
  descriptionLabel: t('web.workspaceAddPresentation.description'),
  descriptionPlaceholder: t('web.workspaceAddPresentation.optionalNotes')
};

export const quickParentKindOptions: AddControlOption<'location' | 'container'>[] = [
  { value: 'location', label: t('web.workspaceAddPresentation.location') },
  { value: 'container', label: t('web.workspaceAddPresentation.container') }
];

export function assetKindControlOptions(): AddControlOption<AssetKind>[] {
  return assetKinds.map((kind) => ({ value: kind, label: assetKindLabel(kind) }));
}

export function addAssetKindCopy(kind: AssetKind): AddAssetKindCopy {
  const kindLabel = assetKindLabel(kind);
  const selectedKindLabel = kindLabel.toLowerCase();
  return {
    heading: t(`asset.add.${kind}`),
    kindLabel,
    nameLabel: t(`asset.name.${kind}`),
    namePlaceholder: addAssetNamePlaceholder(kind),
    saveLabel: t(`asset.save.${kind}`),
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
    return input.selectedParent?.title ?? t('web.workspaceAddPresentation.inventoryRoot');
  }

  const parentName = input.quickParentTitle.trim() ? t(`asset.newNamed.${input.quickParentKind}`, { name: input.quickParentTitle.trim() }) : t(`asset.new.${input.quickParentKind}`);
  return t('web.workspaceAddPresentation.in', { parentName: String(parentName), value: String(quickParentContainerSummary(input.selectedParent)) });
}

export function quickParentContainerLabel(selectedParent: ParentTargetViewModel | null): string {
  return selectedParent?.title ?? t('web.workspaceAddPresentation.inventoryRoot');
}

export function quickParentContainerTrail(selectedParent: ParentTargetViewModel | null): string {
  return selectedParent?.containmentTrail ?? '';
}

export function quickParentContainerSummary(selectedParent: ParentTargetViewModel | null): string {
  return selectedParent ? `${selectedParent.title} / ${selectedParent.containmentTrail}` : t('web.workspaceAddPresentation.inventoryRoot');
}

export function quickParentMissingNameMessage(): string {
  return t('web.workspaceAddPresentation.enterAParentNameOrTurnThisOptionOff');
}

export function addPhotoCountLabel(photoCount: number): string {
  if (photoCount === 0) {
    return t('web.workspaceAddPresentation.noPhotos');
  }
  return t('photos.count', { count: photoCount });
}

export function addSupportedImageTypes(mediaPolicy: MediaUploadPolicy): SelectedPhoto['contentType'][] {
  return mediaPolicy.supportedContentTypes.filter((type): type is SelectedPhoto['contentType'] => type.startsWith('image/'));
}

export function addPhotoAcceptTypes(supportedImageTypes: string[]): string {
  return supportedImageTypes.join(',');
}

export function addPhotoSupportedTypeLabel(types: string[]): string {
  if (types.length === 0) {
    return t('web.workspaceAddPresentation.noImageFormats');
  }
  const labels = types.map(formatImageContentType);
  if (labels.length === 1) {
    return labels[0] ?? '';
  }
  if (labels.length === 2) {
    return t('web.workspaceAddPresentation.or', { value: String(labels[0]), value2: String(labels[1]) });
  }
  return t('web.workspaceAddPresentation.or2', { value: String(labels.slice(0, -1).join(', ')), value2: String(labels[labels.length - 1]) });
}

export function addPhotoHelpText(supportedTypeLabel: string, maxBytesLabel: string): string {
  return t('web.workspaceAddPresentation.optionalUpTo', { supportedTypeLabel: String(supportedTypeLabel), maxBytesLabel: String(maxBytesLabel) });
}

export function addPhotoRemoveLabel(photo: Pick<SelectedPhoto, 'name'>): string {
  return t('web.workspaceAddPresentation.remove', { name: String(photo.name) });
}

function addAssetNamePlaceholder(kind: AssetKind): string {
  if (kind === 'location') {
    return t('web.workspaceAddPresentation.garageShelf');
  }
  if (kind === 'container') {
    return t('web.workspaceAddPresentation.clearStorageBin');
  }
  return t('web.workspaceAddPresentation.tomatoFertilizer');
}

function formatImageContentType(type: string): string {
  if (type === 'image/jpeg') return 'JPEG';
  if (type === 'image/png') return 'PNG';
  if (type === 'image/webp') return 'WebP';
  return type.replace(/^image\//, '').toUpperCase();
}
