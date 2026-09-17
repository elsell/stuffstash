import { NativeCommandButton } from './NativeCommandButton';
import { useLayoutEffect, useMemo, useRef, useState } from 'react';
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
          <NativeCommandButton label="Add photos" onPress={onAddPhotos} />
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
          return <GalleryPreview key={photo.id ?? photo.uri} photo={photo} palette={palette}
            presentation={presentation} width={photoWidth} onPhotoPress={onPhotoPress} />;
        })}
      </ScrollView>

      {canUseAddPhotos ? (
        <NativeCommandButton label="Add photos" onPress={onAddPhotos} />
      ) : null}
    </View>
  );
}

function GalleryPreview({ photo, palette, presentation, width, onPhotoPress }: {
  readonly photo: AssetPhotoViewModel;
  readonly palette: MobileColorPalette;
  readonly presentation: AssetDetailPhotoPagePresentation;
  readonly width: number;
  readonly onPhotoPress?: (photoId: string) => void;
}) {
  const canOpen = photo.id !== undefined && onPhotoPress !== undefined;
  const uri = photo.heroUri ?? photo.uri;
  const headers = photo.heroHeaders ?? photo.headers;
  const source = useMemo(() => ({ uri, headers }), [uri, headers]);
  const currentSource = useRef<typeof source | undefined>(source);
  const [failedSource, setFailedSource] = useState<typeof source>();
  useLayoutEffect(() => {
    currentSource.current = source;
    return () => { currentSource.current = undefined; };
  }, [source]);
  const failed = failedSource === source;
  return <Pressable accessibilityLabel={presentation.accessibilityLabel}
    accessibilityValue={failed ? { text: 'Preview unavailable' } : undefined}
    accessibilityHint={canOpen ? 'Opens the original photo' : undefined}
    accessibilityRole="imagebutton" accessibilityState={{ disabled: !canOpen }} disabled={!canOpen}
    onPress={() => { if (photo.id && onPhotoPress) onPhotoPress(photo.id); }}
    style={[styles.mediaFrame, styles.photoMediaFrame, { backgroundColor: palette.surfaceMuted, width }]}>
    {failed ? <View style={styles.previewFailure}>
      <Text style={[styles.emptySupporting, { color: palette.text }]}>Preview unavailable</Text>
      {canOpen ? <Text style={[styles.emptySupporting, { color: palette.text }]}>Open photo</Text> : null}
    </View> : <Image accessibilityIgnoresInvertColors accessible={false} resizeMode="cover"
    source={source} style={styles.photo}
    onError={() => { if (currentSource.current === source) setFailedSource(source); }} />}
    <View accessible={false} style={[styles.positionBadge, { backgroundColor: palette.scrim }]}>
      <Text style={[styles.positionText, { color: palette.onScrim }]}>{presentation.positionLabel}</Text>
    </View>
  </Pressable>;
}

const styles = StyleSheet.create({
  previewFailure: { padding: spacing.lg, gap: spacing.sm },
  gallery: {
    alignItems: 'flex-start',
    gap: spacing.sm
  },
  photoStrip: {
    gap: galleryGap
  },
  mediaFrame: {
    alignItems: 'center',
    borderRadius: radius.lg,
    justifyContent: 'center',
    position: 'relative'
  },
  photoMediaFrame: {
    aspectRatio: 4 / 3,
    overflow: 'hidden'
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
    fontWeight: '600'
  },
  emptyMedia: {
    borderWidth: 1,
    gap: spacing.xs,
    minHeight: 240,
    padding: spacing.lg
  },
  emptyTitle: {
    fontSize: 20,
    fontWeight: '600',
    textAlign: 'center'
  },
  emptySupporting: {
    fontSize: 15,
    fontWeight: '400',
    textAlign: 'center'
  },

});
