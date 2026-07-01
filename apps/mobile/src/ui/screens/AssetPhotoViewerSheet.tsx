import { Alert, Pressable, StyleSheet, Text, View } from 'react-native';
import ImageViewing from 'react-native-image-viewing';
import type { AssetPhotoViewModel } from '../../application/assets/AssetViewModels';
import type { AssetPhotoViewerModel } from '../components/AssetPhotoWorkspacePresentation';
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
  const currentPhoto = model?.photo;

  function removeCurrentPhoto(): void {
    if (!currentPhoto?.id) {
      return;
    }

    Alert.alert('Remove photo?', 'This removes the photo from this asset.', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Remove',
        style: 'destructive',
        onPress: () => onRemove(currentPhoto.id as string)
      }
    ]);
  }

  return (
    <ImageViewing
      animationType="fade"
      backgroundColor="#05080A"
      doubleTapToZoomEnabled
      FooterComponent={({ imageIndex }) => (
        <View style={styles.previewFooter}>
          <Text style={styles.previewCount}>
            {`${(imageIndex + 1).toString()} / ${photos.length.toString()}`}
          </Text>
          <Text style={styles.previewFileName} numberOfLines={1}>
            {photos[imageIndex]?.fileName ?? photos[imageIndex]?.label ?? 'Photo'}
          </Text>
        </View>
      )}
      HeaderComponent={() => (
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
              disabled={!currentPhoto?.id}
              hitSlop={12}
              onPress={removeCurrentPhoto}
              style={styles.previewHeaderButton}
            >
              <Text style={styles.previewRemoveText}>Remove</Text>
            </Pressable>
          ) : null}
        </View>
      )}
      imageIndex={selectedIndex}
      images={photos.map((photo) => ({ uri: photo.uri, headers: photo.headers }))}
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
  previewFooter: {
    alignItems: 'center',
    gap: spacing.xs,
    paddingBottom: spacing.xl,
    paddingHorizontal: spacing.lg
  }
});
