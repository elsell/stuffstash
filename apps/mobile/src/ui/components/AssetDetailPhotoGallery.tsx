import {
  Image,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  useWindowDimensions,
  View
} from 'react-native';
import { Camera } from 'lucide-react-native';
import type { AssetPhotoViewModel } from '../../application/assets/AssetViewModels';
import {
  radius,
  spacing,
  type MobileColorPalette
} from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';

const galleryGap = spacing.sm;
const defaultHorizontalPagePadding: number = spacing.lg;

export type AssetDetailPhotoPagePresentation = {
  readonly accessibilityLabel: string;
  readonly positionLabel: string;
};

export type AssetDetailPhotoGalleryProps = {
  readonly canAddPhotos: boolean;
  readonly contentHorizontalPadding?: number;
  readonly imagePlaceholderLabel: string;
  readonly onAddPhotos?: () => void;
  readonly onPhotoPress?: (photoId: string) => void;
  readonly palette?: MobileColorPalette;
  readonly photos: readonly AssetPhotoViewModel[];
};

export function assetDetailPhotoPages(
  photos: readonly AssetPhotoViewModel[]
): readonly AssetDetailPhotoPagePresentation[] {
  return photos.map((_, index) => {
    const positionLabel = `${(index + 1).toString()} of ${photos.length.toString()}`;
    return {
      accessibilityLabel: `Open photo ${positionLabel}`,
      positionLabel
    };
  });
}

export function assetDetailPhotoWidth(
  viewportWidth: number,
  contentHorizontalPadding = defaultHorizontalPagePadding
): number {
  return Math.max(0, viewportWidth - (contentHorizontalPadding * 2));
}

export function AssetDetailPhotoGallery({
  canAddPhotos,
  contentHorizontalPadding = defaultHorizontalPagePadding,
  imagePlaceholderLabel,
  onAddPhotos,
  onPhotoPress,
  palette: paletteOverride,
  photos
}: AssetDetailPhotoGalleryProps) {
  const appearancePalette = useAppearanceAwarePalette();
  const palette = paletteOverride ?? appearancePalette;
  const { width: viewportWidth } = useWindowDimensions();
  const photoWidth = assetDetailPhotoWidth(viewportWidth, contentHorizontalPadding);
  const pages = assetDetailPhotoPages(photos);
  const canUseAddPhotos = canAddPhotos && onAddPhotos !== undefined;

  if (photos.length === 0) {
    return (
      <View style={styles.gallery}>
        <View
          accessibilityLabel="No photos"
          style={[
            styles.mediaFrame,
            styles.emptyMedia,
            {
              backgroundColor: palette.elevatedSurface,
              borderColor: palette.border,
              width: photoWidth
            }
          ]}
        >
          <Camera color={palette.textMuted} size={28} />
          <Text style={[styles.emptyTitle, { color: palette.text }]}>{imagePlaceholderLabel}</Text>
          <Text style={[styles.emptySupporting, { color: palette.textMuted }]}>No photos</Text>
        </View>
        {canUseAddPhotos ? (
          <AddPhotosButton onAddPhotos={onAddPhotos} palette={palette} />
        ) : null}
      </View>
    );
  }

  return (
    <View style={styles.gallery}>
      <ScrollView
        accessibilityLabel={`Asset photos, ${photos.length.toString()} total`}
        contentContainerStyle={styles.photoStrip}
        decelerationRate="fast"
        horizontal
        showsHorizontalScrollIndicator={false}
        snapToAlignment="start"
        snapToInterval={photoWidth + galleryGap}
      >
        {photos.map((photo, index) => {
          const presentation = pages[index] as AssetDetailPhotoPagePresentation;
          const canOpenPhoto = photo.id !== undefined && onPhotoPress !== undefined;
          return (
            <Pressable
              accessibilityLabel={presentation.accessibilityLabel}
              accessibilityRole="imagebutton"
              accessibilityState={{ disabled: !canOpenPhoto }}
              disabled={!canOpenPhoto}
              key={photo.id ?? photo.uri}
              onPress={() => photo.id && onPhotoPress ? onPhotoPress(photo.id) : undefined}
              style={[styles.mediaFrame, { backgroundColor: palette.surfaceMuted, width: photoWidth }]}
            >
              <Image
                accessibilityIgnoresInvertColors
                accessible={false}
                resizeMode="cover"
                source={{
                  uri: photo.heroUri ?? photo.uri,
                  headers: photo.heroHeaders ?? photo.headers
                }}
                style={styles.photo}
              />
              <View
                accessible={false}
                style={[styles.positionBadge, { backgroundColor: palette.scrim }]}
              >
                <Text style={[styles.positionText, { color: palette.onScrim }]}>
                  {presentation.positionLabel}
                </Text>
              </View>
            </Pressable>
          );
        })}
      </ScrollView>

      {canUseAddPhotos ? (
        <AddPhotosButton onAddPhotos={onAddPhotos} palette={palette} />
      ) : null}
    </View>
  );
}

function AddPhotosButton({
  onAddPhotos,
  palette
}: {
  readonly onAddPhotos: () => void;
  readonly palette: MobileColorPalette;
}) {
  return (
    <Pressable
      accessibilityLabel="Add photos"
      accessibilityRole="button"
      onPress={onAddPhotos}
      style={({ pressed }) => [
        styles.addPhotoButton,
        {
          backgroundColor: pressed ? palette.selected : palette.elevatedSurface,
          borderColor: palette.controlBorder
        }
      ]}
    >
      <Camera color={palette.action} size={18} />
      <Text style={[styles.addPhotoText, { color: palette.action }]}>Add photos</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  gallery: {
    alignItems: 'flex-start',
    gap: spacing.sm
  },
  photoStrip: {
    gap: galleryGap
  },
  mediaFrame: {
    alignItems: 'center',
    aspectRatio: 4 / 3,
    borderRadius: radius.lg,
    justifyContent: 'center',
    overflow: 'hidden',
    position: 'relative'
  },
  photo: {
    height: '100%',
    width: '100%'
  },
  positionBadge: {
    borderRadius: radius.sm,
    bottom: spacing.sm,
    left: spacing.sm,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs,
    position: 'absolute'
  },
  positionText: {
    fontSize: 13,
    fontWeight: '600',
    lineHeight: 17
  },
  emptyMedia: {
    borderWidth: 1,
    gap: spacing.xs,
    padding: spacing.lg
  },
  emptyTitle: {
    fontSize: 20,
    fontWeight: '600',
    lineHeight: 25,
    textAlign: 'center'
  },
  emptySupporting: {
    fontSize: 15,
    fontWeight: '400',
    lineHeight: 20,
    textAlign: 'center'
  },
  addPhotoButton: {
    alignItems: 'center',
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.xs,
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  addPhotoText: {
    fontSize: 15,
    fontWeight: '600',
    lineHeight: 20,
    textAlign: 'center'
  }
});
