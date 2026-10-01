import { t } from '../../presentation/localization';
export const photoViewerActions = {
  close: { label: t('mobile.PhotoViewerActionButtontypes.closePhotoViewer'), symbol: 'xmark' },
  previous: { label: t('mobile.PhotoViewerActionButtontypes.previousPhoto'), symbol: 'chevron.left' },
  next: { label: t('mobile.PhotoViewerActionButtontypes.nextPhoto'), symbol: 'chevron.right' },
  remove: { label: t('mobile.PhotoViewerActionButtontypes.removePhoto'), symbol: 'trash' }
} as const;

export type PhotoViewerActionButtonProps = {
  readonly action: keyof typeof photoViewerActions;
  readonly disabled?: boolean;
  readonly onPress: () => void;
};
