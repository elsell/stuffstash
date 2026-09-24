import { PhotoViewerSystemBars } from './PhotoViewerSystemBars';
import React, { useEffect, useMemo } from 'react';
import { AccessibilityInfo, Platform, Pressable, StatusBar, StyleSheet, Text, View } from 'react-native';
import { NativeCommandButton } from './NativeCommandButton';
import ImageViewing from 'react-native-image-viewing';
import { ChevronLeft, ChevronRight, Trash2, X } from 'lucide-react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { radius, spacing } from '../theme/tokens';
import {
  fullScreenPhotoViewerActionState,
  type FullScreenPhotoViewerPhoto
} from './FullScreenPhotoViewerPresentation';

export type { FullScreenPhotoViewerPhoto } from './FullScreenPhotoViewerPresentation';

// Photo viewing intentionally uses a fixed neutral-black canvas so image
// colors and exposure do not shift with the surrounding app appearance.
const viewerColors = {
  background: '#05080A',
  foreground: '#FFFFFF',
  warning: '#F5B95E'
} as const;

export function FullScreenPhotoViewer({
  canRemove,
  isRemoving = false,
  currentIndex,
  onClose,
  onRemove,
  onSelectIndex,
  photos
}: {
  readonly canRemove: boolean;
  readonly isRemoving?: boolean;
  readonly currentIndex: number | undefined;
  readonly onClose: () => void;
  readonly onRemove?: (photo: FullScreenPhotoViewerPhoto, index: number) => void;
  readonly onSelectIndex: (index: number) => void;
  readonly photos: readonly FullScreenPhotoViewerPhoto[];
}) {
  const insets = useSafeAreaInsets();
  const visible = currentIndex !== undefined && photos[currentIndex] !== undefined;
  const selectedIndex = Math.max(0, currentIndex ?? 0);
  const images = useMemo(() => photos.map(photo => ({ uri: photo.uri, headers: photo.headers })), [photos]);

  return (
    <>
    {visible && Platform.OS === 'ios' ? <StatusBar barStyle="light-content" /> : null}
    <ImageViewing
      HeaderComponent={PhotoViewerHeader}
      ErrorComponent={PhotoViewerLoadError}
      animationType="fade"
      backgroundColor={viewerColors.background}
      doubleTapToZoomEnabled
      FooterComponent={({ imageIndex }) => (
        <PhotoViewerToolbar
          canRemove={canRemove}
          isRemoving={isRemoving}
          imageIndex={imageIndex}
          onClose={onClose}
          onRemove={onRemove}
          onSelectIndex={onSelectIndex}
          photos={photos}
          safeBottomInset={insets.bottom}
        />
      )}
      imageIndex={selectedIndex}
      images={images}
      keyExtractor={(_image, index) => photos[index]?.id ?? index.toString()}
      onImageIndexChange={onSelectIndex}
      onRequestClose={onClose}
      presentationStyle="overFullScreen"
      swipeToCloseEnabled
      visible={visible}
    />
    </>
  );
}

// The safe-area-aware footer owns Close; omit the library's duplicate header.
function PhotoViewerHeader() { return <PhotoViewerSystemBars />; }

export function PhotoViewerLoadError({ onRetry }: { readonly onRetry: () => void }) {
  return <View style={styles.loadError}>
    <Text accessibilityRole="alert" accessibilityLiveRegion="polite" style={styles.loadErrorText}>Photo unavailable</Text>
    <NativeCommandButton label="Retry photo" prominence="primary" onPress={onRetry} />
  </View>;
}

function PhotoViewerToolbar({
  canRemove,
  isRemoving,
  imageIndex,
  onClose,
  onRemove,
  onSelectIndex,
  photos,
  safeBottomInset
}: {
  readonly canRemove: boolean;
  readonly isRemoving: boolean;
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
  useEffect(() => {
    if (isRemoving && Platform.OS === 'ios') AccessibilityInfo.announceForAccessibility('Removing photo…');
  }, [isRemoving]);

  return (
    <View style={[styles.toolbarOuter, { paddingBottom: Math.max(spacing.md, safeBottomInset) }]}>
      <View style={styles.infoBlock}>
        <Text style={styles.positionText}>{state.positionLabel}</Text>
        {isRemoving ? <Text accessibilityLiveRegion="polite" style={styles.positionText}>Removing photo…</Text> : null}
        <Text numberOfLines={1} style={styles.fileText}>{state.fileLabel}</Text>
        {state.metadataLabel ? (
          <Text numberOfLines={1} style={styles.metadataText}>{state.metadataLabel}</Text>
        ) : null}
      </View>
      <View style={styles.toolbar}>
        <ViewerIconButton accessibilityLabel="Close photo viewer" onPress={onClose}>
          <X color={viewerColors.foreground} size={25} strokeWidth={2.5} />
        </ViewerIconButton>
        {photos.length > 1 ? (
          <>
            <ViewerIconButton
              accessibilityLabel="Previous photo"
              disabled={!state.canGoPrevious}
              onPress={() => onSelectIndex(Math.max(0, imageIndex - 1))}
            >
              <ChevronLeft color={viewerColors.foreground} size={27} strokeWidth={2.5} />
            </ViewerIconButton>
            <ViewerIconButton
              accessibilityLabel="Next photo"
              disabled={!state.canGoNext}
              onPress={() => onSelectIndex(Math.min(photos.length - 1, imageIndex + 1))}
            >
              <ChevronRight color={viewerColors.foreground} size={27} strokeWidth={2.5} />
            </ViewerIconButton>
          </>
        ) : null}
        {canShowRemoveAction ? (
          <ViewerIconButton
            accessibilityLabel="Remove photo"
            destructive
            disabled={isRemoving || !state.canRemove || !currentPhoto}
            onPress={() => {
              if (currentPhoto && !isRemoving) {
                onRemove?.(currentPhoto, imageIndex);
              }
            }}
          >
            <Trash2 color={viewerColors.warning} size={24} strokeWidth={2.4} />
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
  loadError: { flex: 1, alignItems: 'center', justifyContent: 'center', padding: spacing.lg, gap: spacing.md },
  loadErrorText: { color: viewerColors.foreground, fontSize: 20, textAlign: 'center', flexShrink: 1 },
  toolbarOuter: {
    backgroundColor: viewerColors.background,
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
    color: viewerColors.foreground,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  fileText: {
    color: viewerColors.foreground,
    fontSize: 13,
    fontWeight: '800',
    letterSpacing: 0,
    maxWidth: '92%',
    opacity: 0.84
  },
  metadataText: {
    color: viewerColors.foreground,
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
