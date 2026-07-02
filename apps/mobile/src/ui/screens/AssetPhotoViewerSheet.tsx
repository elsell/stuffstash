import { Alert, Pressable, StyleSheet, Text, View } from 'react-native';
import ImageViewing from 'react-native-image-viewing';
import { ChevronLeft, ChevronRight } from 'lucide-react-native';
import type { AssetPhotoViewModel } from '../../application/assets/AssetViewModels';
import {
  assetPhotoViewerControls,
  assetPhotoViewerModelAtIndex,
  type AssetPhotoViewerModel
} from '../components/AssetPhotoWorkspacePresentation';
import { colors, spacing } from '../theme/tokens';

export function AssetPhotoViewerSheet({
  canRemove,
  model,
  onClose,
  onRemove,
  onSelectPhoto,
  photos
}: {
  readonly canRemove: boolean;
  readonly model: AssetPhotoViewerModel | undefined;
  readonly onClose: () => void;
  readonly onRemove: (photoId: string) => void;
  readonly onSelectPhoto: (photoId: string) => void;
  readonly photos: readonly AssetPhotoViewModel[];
}) {
  const selectedIndex = model
    ? Math.max(0, photos.findIndex((photo) => photo.id === model.photo.id))
    : 0;

  function removeCurrentPhoto(photoId: string | undefined): void {
    if (!photoId) {
      return;
    }

    Alert.alert('Remove photo?', 'This removes the photo from this asset.', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Remove',
        style: 'destructive',
        onPress: () => onRemove(photoId)
      }
    ]);
  }

  return (
    <ImageViewing
      animationType="fade"
      backgroundColor="#05080A"
      doubleTapToZoomEnabled
      FooterComponent={({ imageIndex }) => (
        <PhotoViewerFooter
          canRemove={canRemove}
          imageIndex={imageIndex}
          onSelectPhoto={onSelectPhoto}
          photos={photos}
        />
      )}
      HeaderComponent={({ imageIndex }) => (
        <PhotoViewerHeader
          canRemove={canRemove}
          imageIndex={imageIndex}
          onClose={onClose}
          onRemove={removeCurrentPhoto}
          photos={photos}
        />
      )}
      imageIndex={selectedIndex}
      images={photos.map((photo) => ({
        uri: photo.viewerUri ?? photo.heroUri ?? photo.uri,
        headers: photo.viewerHeaders ?? photo.heroHeaders ?? photo.headers
      }))}
      keyExtractor={(_image, index) => photos[index]?.id ?? index.toString()}
      onImageIndexChange={(index) => {
        const photoId = photos[index]?.id;
        if (photoId) {
          onSelectPhoto(photoId);
        }
      }}
      onRequestClose={onClose}
      presentationStyle="overFullScreen"
      swipeToCloseEnabled
      visible={model !== undefined}
    />
  );
}

function PhotoViewerHeader({
  canRemove,
  imageIndex,
  onClose,
  onRemove,
  photos
}: {
  readonly canRemove: boolean;
  readonly imageIndex: number;
  readonly onClose: () => void;
  readonly onRemove: (photoId: string | undefined) => void;
  readonly photos: readonly AssetPhotoViewModel[];
}) {
  const currentModel = assetPhotoViewerModelAtIndex(photos, imageIndex);
  const controls = assetPhotoViewerControls(currentModel, canRemove);
  return (
    <View style={styles.previewHeader}>
      <Pressable
        accessibilityRole="button"
        hitSlop={12}
        onPress={onClose}
        style={styles.previewHeaderButton}
      >
        <Text style={styles.previewHeaderText}>Close</Text>
      </Pressable>
      {canRemove ? (
        <Pressable
          accessibilityRole="button"
          disabled={!controls.canRemove}
          hitSlop={12}
          onPress={() => onRemove(currentModel?.photo.id)}
          style={[styles.previewHeaderButton, !controls.canRemove ? styles.previewDisabledButton : null]}
        >
          <Text style={styles.previewRemoveText}>Remove</Text>
        </Pressable>
      ) : null}
    </View>
  );
}

function PhotoViewerFooter({
  canRemove,
  imageIndex,
  onSelectPhoto,
  photos
}: {
  readonly canRemove: boolean;
  readonly imageIndex: number;
  readonly onSelectPhoto: (photoId: string) => void;
  readonly photos: readonly AssetPhotoViewModel[];
}) {
  const currentModel = assetPhotoViewerModelAtIndex(photos, imageIndex);
  const controls = assetPhotoViewerControls(currentModel, canRemove);
  return (
    <View style={styles.previewFooter}>
      <Text style={styles.previewCount}>{controls.positionLabel}</Text>
      <Text style={styles.previewFileName} numberOfLines={1}>
        {controls.fileLabel}
      </Text>
      {controls.metadataLabel ? (
        <Text style={styles.previewMetadata} numberOfLines={1}>
          {controls.metadataLabel}
        </Text>
      ) : null}
      {photos.length > 1 ? (
        <View style={styles.previewNavigation}>
          <PhotoNavigationButton
            accessibilityLabel="Previous photo"
            disabled={!controls.canGoPrevious}
            direction="previous"
            onPress={() => {
              if (currentModel?.previousPhotoId) {
                onSelectPhoto(currentModel.previousPhotoId);
              }
            }}
          />
          <PhotoNavigationButton
            accessibilityLabel="Next photo"
            disabled={!controls.canGoNext}
            direction="next"
            onPress={() => {
              if (currentModel?.nextPhotoId) {
                onSelectPhoto(currentModel.nextPhotoId);
              }
            }}
          />
        </View>
      ) : null}
    </View>
  );
}

function PhotoNavigationButton({
  accessibilityLabel,
  direction,
  disabled,
  onPress
}: {
  readonly accessibilityLabel: string;
  readonly direction: 'next' | 'previous';
  readonly disabled: boolean;
  readonly onPress: () => void;
}) {
  return (
    <Pressable
      accessibilityLabel={accessibilityLabel}
      accessibilityRole="button"
      accessibilityState={{ disabled }}
      disabled={disabled}
      onPress={onPress}
      style={[styles.previewNavButton, disabled ? styles.previewDisabledButton : null]}
    >
      {direction === 'previous'
        ? <ChevronLeft color={colors.onAction} size={24} />
        : <ChevronRight color={colors.onAction} size={24} />}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  previewHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.xl
  },
  previewHeaderButton: {
    minHeight: 44,
    minWidth: 72,
    justifyContent: 'center'
  },
  previewDisabledButton: {
    opacity: 0.45
  },
  previewHeaderText: {
    color: colors.onAction,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  },
  previewRemoveText: {
    color: colors.brandAmber,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0,
    textAlign: 'right'
  },
  previewCount: {
    color: colors.onAction,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0,
    textAlign: 'center'
  },
  previewFileName: {
    color: colors.onAction,
    fontSize: 13,
    fontWeight: '700',
    letterSpacing: 0,
    opacity: 0.8,
    textAlign: 'center'
  },
  previewMetadata: {
    color: colors.onAction,
    fontSize: 12,
    fontWeight: '700',
    letterSpacing: 0,
    opacity: 0.65,
    textAlign: 'center'
  },
  previewFooter: {
    alignItems: 'center',
    gap: spacing.xs,
    paddingBottom: spacing.xl,
    paddingHorizontal: spacing.lg
  },
  previewNavigation: {
    flexDirection: 'row',
    gap: spacing.md,
    marginTop: spacing.sm
  },
  previewNavButton: {
    alignItems: 'center',
    backgroundColor: 'rgba(255, 255, 255, 0.14)',
    borderRadius: 22,
    height: 44,
    justifyContent: 'center',
    width: 44
  }
});
