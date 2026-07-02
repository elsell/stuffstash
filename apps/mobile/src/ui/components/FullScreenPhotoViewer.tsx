import type React from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import ImageViewing from 'react-native-image-viewing';
import { ChevronLeft, ChevronRight, Trash2, X } from 'lucide-react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { colors, radius, spacing } from '../theme/tokens';
import {
  fullScreenPhotoViewerActionState,
  type FullScreenPhotoViewerPhoto
} from './FullScreenPhotoViewerPresentation';

export type { FullScreenPhotoViewerPhoto } from './FullScreenPhotoViewerPresentation';

export function FullScreenPhotoViewer({
  canRemove,
  currentIndex,
  onClose,
  onRemove,
  onSelectIndex,
  photos
}: {
  readonly canRemove: boolean;
  readonly currentIndex: number | undefined;
  readonly onClose: () => void;
  readonly onRemove?: (photo: FullScreenPhotoViewerPhoto, index: number) => void;
  readonly onSelectIndex: (index: number) => void;
  readonly photos: readonly FullScreenPhotoViewerPhoto[];
}) {
  const insets = useSafeAreaInsets();
  const visible = currentIndex !== undefined && photos[currentIndex] !== undefined;
  const selectedIndex = Math.max(0, currentIndex ?? 0);

  return (
    <ImageViewing
      animationType="fade"
      backgroundColor="#05080A"
      doubleTapToZoomEnabled
      FooterComponent={({ imageIndex }) => (
        <PhotoViewerToolbar
          canRemove={canRemove}
          imageIndex={imageIndex}
          onClose={onClose}
          onRemove={onRemove}
          onSelectIndex={onSelectIndex}
          photos={photos}
          safeBottomInset={insets.bottom}
        />
      )}
      imageIndex={selectedIndex}
      images={photos.map((photo) => ({ uri: photo.uri, headers: photo.headers }))}
      keyExtractor={(_image, index) => photos[index]?.id ?? index.toString()}
      onImageIndexChange={onSelectIndex}
      onRequestClose={onClose}
      presentationStyle="overFullScreen"
      swipeToCloseEnabled
      visible={visible}
    />
  );
}

function PhotoViewerToolbar({
  canRemove,
  imageIndex,
  onClose,
  onRemove,
  onSelectIndex,
  photos,
  safeBottomInset
}: {
  readonly canRemove: boolean;
  readonly imageIndex: number;
  readonly onClose: () => void;
  readonly onRemove?: (photo: FullScreenPhotoViewerPhoto, index: number) => void;
  readonly onSelectIndex: (index: number) => void;
  readonly photos: readonly FullScreenPhotoViewerPhoto[];
  readonly safeBottomInset: number;
}) {
  const canShowRemoveAction = canRemove && onRemove !== undefined;
  const state = fullScreenPhotoViewerActionState(photos, imageIndex, canShowRemoveAction);
  const currentPhoto = photos[imageIndex];

  return (
    <View style={[styles.toolbarOuter, { paddingBottom: Math.max(spacing.md, safeBottomInset) }]}>
      <View style={styles.infoBlock}>
        <Text style={styles.positionText}>{state.positionLabel}</Text>
        <Text numberOfLines={1} style={styles.fileText}>{state.fileLabel}</Text>
        {state.metadataLabel ? (
          <Text numberOfLines={1} style={styles.metadataText}>{state.metadataLabel}</Text>
        ) : null}
      </View>
      <View style={styles.toolbar}>
        <ViewerIconButton accessibilityLabel="Close photo viewer" onPress={onClose}>
          <X color={colors.onAction} size={25} strokeWidth={2.5} />
        </ViewerIconButton>
        {photos.length > 1 ? (
          <>
            <ViewerIconButton
              accessibilityLabel="Previous photo"
              disabled={!state.canGoPrevious}
              onPress={() => onSelectIndex(Math.max(0, imageIndex - 1))}
            >
              <ChevronLeft color={colors.onAction} size={27} strokeWidth={2.5} />
            </ViewerIconButton>
            <ViewerIconButton
              accessibilityLabel="Next photo"
              disabled={!state.canGoNext}
              onPress={() => onSelectIndex(Math.min(photos.length - 1, imageIndex + 1))}
            >
              <ChevronRight color={colors.onAction} size={27} strokeWidth={2.5} />
            </ViewerIconButton>
          </>
        ) : null}
        {canShowRemoveAction ? (
          <ViewerIconButton
            accessibilityLabel="Remove photo"
            destructive
            disabled={!state.canRemove || !currentPhoto}
            onPress={() => {
              if (currentPhoto) {
                onRemove?.(currentPhoto, imageIndex);
              }
            }}
          >
            <Trash2 color={colors.brandAmber} size={24} strokeWidth={2.4} />
          </ViewerIconButton>
        ) : null}
      </View>
    </View>
  );
}

function ViewerIconButton({
  accessibilityLabel,
  children,
  destructive,
  disabled,
  onPress
}: {
  readonly accessibilityLabel: string;
  readonly children: React.ReactNode;
  readonly destructive?: boolean;
  readonly disabled?: boolean;
  readonly onPress: () => void;
}) {
  return (
    <Pressable
      accessibilityLabel={accessibilityLabel}
      accessibilityRole="button"
      accessibilityState={{ disabled }}
      disabled={disabled}
      hitSlop={10}
      onPress={onPress}
      style={[
        styles.iconButton,
        destructive ? styles.destructiveButton : null,
        disabled ? styles.disabledButton : null
      ]}
    >
      {children}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  toolbarOuter: {
    gap: spacing.sm,
    paddingHorizontal: spacing.md,
    paddingTop: spacing.sm
  },
  infoBlock: {
    alignItems: 'center',
    gap: 2,
    paddingHorizontal: spacing.lg
  },
  positionText: {
    color: colors.onAction,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  fileText: {
    color: colors.onAction,
    fontSize: 13,
    fontWeight: '800',
    letterSpacing: 0,
    maxWidth: '92%',
    opacity: 0.84
  },
  metadataText: {
    color: colors.onAction,
    fontSize: 12,
    fontWeight: '700',
    letterSpacing: 0,
    maxWidth: '92%',
    opacity: 0.62
  },
  toolbar: {
    alignItems: 'center',
    alignSelf: 'center',
    backgroundColor: 'rgba(13, 18, 22, 0.82)',
    borderColor: 'rgba(255, 255, 255, 0.14)',
    borderRadius: radius.lg,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.xs,
    justifyContent: 'center',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  iconButton: {
    alignItems: 'center',
    borderRadius: radius.md,
    height: 46,
    justifyContent: 'center',
    width: 54
  },
  destructiveButton: {
    backgroundColor: 'rgba(249, 189, 73, 0.12)'
  },
  disabledButton: {
    opacity: 0.35
  }
});
