import { PhotoViewerSystemBars } from './PhotoViewerSystemBars';
import React, { createContext, useContext, useEffect, useLayoutEffect, useMemo } from 'react';
import { AccessibilityInfo, Platform, StatusBar, StyleSheet, Text, View } from 'react-native';
import { NativeCommandButton } from './NativeCommandButton';
import ImageViewing from 'react-native-image-viewing';
import { NativeActionMenu, type NativeActionMenuGroup } from './NativeActionMenu';
import { PhotoViewerActionButton } from './PhotoViewerActionButton';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { spacing } from '../theme/tokens';
import {
  fullScreenPhotoViewerActionState,
  type FullScreenPhotoViewerPhoto
} from './FullScreenPhotoViewerPresentation';

export type { FullScreenPhotoViewerPhoto } from './FullScreenPhotoViewerPresentation';

// Photo viewing intentionally uses a fixed neutral-black canvas so image
// colors and exposure do not shift with the surrounding app appearance.
const viewerColors = {
  background: '#05080A',
  foreground: '#FFFFFF'
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

  const presentation = useMemo(() => ({ active: false }), [visible]);
  useLayoutEffect(() => {
    presentation.active = visible;
    return () => { presentation.active = false; };
  }, [presentation, visible]);
  const close = () => {
    if (!presentation.active) return;
    presentation.active = false;
    onClose();
  };
  const select = (index: number) => { if (presentation.active) onSelectIndex(index); };

  return (
    <PhotoViewerToolbarContext.Provider value={{ canRemove, isRemoving, onClose: close,
      onRemove, onSelectIndex: select, photos, safeTopInset: insets.top }}>
    {visible && Platform.OS === 'ios' ? <StatusBar barStyle="light-content" /> : null}
    <ImageViewing
      HeaderComponent={PhotoViewerHeader}
      ErrorComponent={PhotoViewerLoadError}
      animationType="fade"
      backgroundColor={viewerColors.background}
      doubleTapToZoomEnabled
      imageIndex={selectedIndex}
      images={images}
      keyExtractor={(_image, index) => photos[index]?.id ?? index.toString()}
      onImageIndexChange={select}
      onRequestClose={close}
      presentationStyle="overFullScreen"
      swipeToCloseEnabled
      visible={visible}
    />
    </PhotoViewerToolbarContext.Provider>
  );
}

const PhotoViewerToolbarContext = createContext<Omit<React.ComponentProps<typeof PhotoViewerToolbar>, 'imageIndex'> | null>(null);
function PhotoViewerHeader({ imageIndex }: { imageIndex: number }) {
  const props = useContext(PhotoViewerToolbarContext);
  return <><PhotoViewerSystemBars />{props ? <PhotoViewerToolbar {...props} imageIndex={imageIndex} /> : null}</>;
}


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
  safeTopInset
}: {
  readonly canRemove: boolean;
  readonly isRemoving: boolean;
  readonly imageIndex: number;
  readonly onClose: () => void;
  readonly onRemove?: (photo: FullScreenPhotoViewerPhoto, index: number) => void;
  readonly onSelectIndex: (index: number) => void;
  readonly photos: readonly FullScreenPhotoViewerPhoto[];
  readonly safeTopInset: number;
}) {
  const canShowRemoveAction = canRemove && onRemove !== undefined;
  const state = fullScreenPhotoViewerActionState(photos, imageIndex, canShowRemoveAction);
  const currentPhoto = photos[imageIndex];
  useEffect(() => {
    if (isRemoving && Platform.OS === 'ios') AccessibilityInfo.announceForAccessibility('Removing photo…');
  }, [isRemoving]);

  const groups: readonly NativeActionMenuGroup[] = [
    { id: 'information', items: [
      { id: 'position', label: state.positionLabel, disabled: true, onPress: () => {} },
      { id: 'name', label: state.fileLabel, disabled: true, onPress: () => {} },
      ...(state.metadataLabel ? [{ id: 'metadata', label: state.metadataLabel, disabled: true, onPress: () => {} }] : [])
    ] },
    { id: 'actions', items: canShowRemoveAction ? [{ id: 'remove', label: 'Remove photo',
      systemImage: 'trash', isDestructive: true,
      disabled: isRemoving || !state.canRemove || !currentPhoto,
      onPress: () => { if (currentPhoto && !isRemoving) onRemove?.(currentPhoto, imageIndex); }
    }] : [] }
  ];
  return <View style={[styles.toolbarOuter, { paddingTop: Math.max(spacing.sm, safeTopInset) }]}>
    <View accessible accessibilityRole="adjustable" accessibilityLabel={`Photo, ${state.positionLabel}`}
      accessibilityHint="Swipe up or down to change photos"
      accessibilityActions={[{ name: 'increment', label: 'Next photo' }, { name: 'decrement', label: 'Previous photo' }]}
      onAccessibilityAction={({ nativeEvent }) => {
        if (nativeEvent.actionName === 'increment' && state.canGoNext) onSelectIndex(imageIndex + 1);
        if (nativeEvent.actionName === 'decrement' && state.canGoPrevious) onSelectIndex(imageIndex - 1);
      }} style={styles.photoAccessibilityTarget} />
    {isRemoving ? <Text accessibilityLiveRegion="polite" style={styles.progress}>Removing photo…</Text> : null}
    <NativeActionMenu accessibilityLabel="Photo options" tone="onDark" groups={groups} />
    <PhotoViewerActionButton action="close" onPress={onClose} />
  </View>;
}

const styles = StyleSheet.create({
  loadError: { flex: 1, alignItems: 'center', justifyContent: 'center', padding: spacing.lg, gap: spacing.md },
  loadErrorText: { color: viewerColors.foreground, fontSize: 20, textAlign: 'center', flexShrink: 1 },
  toolbarOuter: {
    flexDirection: 'row', alignItems: 'center', justifyContent: 'flex-end',
    paddingHorizontal: spacing.md, paddingBottom: spacing.sm, gap: spacing.sm
  },
  photoAccessibilityTarget: { flex: 1, minHeight: 44 },
  progress: { color: viewerColors.foreground, flexShrink: 1 }
});
