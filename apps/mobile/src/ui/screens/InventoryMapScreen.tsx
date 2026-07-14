import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { MutableRefObject } from 'react';
import { router, useFocusEffect } from 'expo-router';
import {
  ActionSheetIOS,
  ActivityIndicator,
  Alert,
  AccessibilityInfo,
  Animated,
  FlatList,
  Image,
  Modal,
  PanResponder,
  Platform,
  Pressable,
  RefreshControl,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  useWindowDimensions,
  View
} from 'react-native';
import { ChevronRight, Info, Package, Plus, Search, X } from 'lucide-react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import type {
  AddAssetPhotoProgressEvent,
  AddAssetPhotosCommand,
  AddAssetPhotosCommandResult
} from '../../application/assets/AddAssetPhotosCommand';
import type { AssetCheckoutCommand } from '../../application/assets/AssetCheckoutCommand';
import type { AssetDetailQuery } from '../../application/assets/AssetDetailQuery';
import type { AssetLifecycleCommand } from '../../application/assets/AssetLifecycleCommand';
import type { DeleteAssetPhotoCommand } from '../../application/assets/DeleteAssetPhotoCommand';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type {
  InventoryMapAssetViewModel,
  InventoryMapQuery,
  InventoryMapViewModel
} from '../../application/assets/InventoryMapQuery';
import type {
  PhotoSelectionQuery,
  SelectedAssetPhoto
} from '../../application/add/PhotoSelectionQuery';
import { radius, spacing } from '../theme/tokens';
import type { MobileColorPalette } from '../theme/tokens';
import { useAppearancePalette } from '../theme/AppearanceContext';
import {
  buildBrowseSurfaceOptions,
  buildInventoryMapEmptyColumnAction,
  buildInventoryMapBreadcrumbs,
  buildInventoryMapColumns,
  buildInventoryMapRowInteractionState,
  clampInventoryMapOffset,
  findInventoryMapSearchMatch,
  inventoryMapBranchSwipeOffset,
  inventoryMapGestureConfig,
  inventoryMapEmbeddedDetailRequest,
  InventoryMapSurface,
  mapOverviewLabel,
  nearestInventoryMapColumnForOffset,
  pathForBreadcrumbLevel,
  preserveInventoryMapHighlightForPath,
  selectInventoryMapBranch,
  shouldActivateInventoryMapPagerPan,
  shouldSelectInventoryMapBranchDuringSwipe,
  shouldSuppressInventoryMapScrollForBranchSwipe
} from './InventoryMapPresentation';
import type { InventoryMapColumnViewModel } from './InventoryMapPresentation';
import { addHereRouteParams } from './AddAssetInitialParent';
import {
  AssetDetailView,
  AssetPhotoUploadProgressViewModel
} from '../components/AssetDetailView';
import { AssetPhotoViewerSheet } from './AssetPhotoViewerSheet';
import {
  assetPhotoViewerModel,
  isAssetPhotoId
} from '../components/AssetPhotoWorkspacePresentation';
import {
  assetLifecycleActionRows,
  assetLifecycleConfirmation,
  assetLifecycleFailurePresentation,
  assetDetailOverflowControlState,
  assetLifecycleOverflowMenu,
  AssetLifecycleActionKind
} from './AssetLifecyclePresentation';
import {
  applyPhotoUploadProgress,
  photoUploadRows
} from './AssetPhotoUploadProgressPresentation';
import {
  assetWorkspaceSuccessStatus,
  visibleAssetWorkspaceStatus,
  AssetWorkspaceStatus
} from './AssetWorkspaceStatusPresentation';
import { showPhotoSourceChooser } from './PhotoSourceChooser';
import { useAppFeedback } from '../feedback/AppFeedback';

type InventoryMapScreenProps = {
  readonly addAssetPhotosCommand: AddAssetPhotosCommand;
  readonly assetCheckoutCommand: AssetCheckoutCommand;
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetLifecycleCommand: AssetLifecycleCommand;
  readonly deleteAssetPhotoCommand: DeleteAssetPhotoCommand;
  readonly inventoryMapQuery: InventoryMapQuery;
  readonly pathStore: MutableRefObject<Map<string, readonly string[]>>;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly selectedSurface: InventoryMapSurface;
  readonly onChangeSurface: (surface: InventoryMapSurface) => void;
};

type InventoryMapState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly map: InventoryMapViewModel }
  | { readonly status: 'error'; readonly message: string };

type MapSheetDetailState =
  | { readonly status: 'idle' }
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly asset: AssetDetailViewModel }
  | { readonly status: 'error'; readonly message: string };

type BranchSwipeVisualState = {
  readonly assetId: string;
  readonly dragX: number;
};

type FinishBranchSwipeOptions = {
  readonly preserveVisual?: boolean;
};

type LoadMapOptions = {
  readonly preserveSelectedAsset?: boolean;
  readonly requestId: number;
  readonly showLoading: boolean;
};

type RenderedInventoryMapColumn = {
  readonly id: string;
  readonly column: InventoryMapColumnViewModel;
  readonly exiting: boolean;
};

const columnGap = spacing.sm;
const horizontalInset = spacing.lg;
const easeOutCubic = (value: number) => 1 - Math.pow(1 - value, 3);

export function InventoryMapScreen({
  addAssetPhotosCommand,
  assetCheckoutCommand,
  assetDetailQuery,
  assetLifecycleCommand,
  deleteAssetPhotoCommand,
  inventoryMapQuery,
  pathStore,
  photoSelectionQuery,
  selectedSurface,
  onChangeSurface
}: InventoryMapScreenProps) {
  const colors = useAppearancePalette();
  const styles = useMemo(() => createStyles(colors), [colors]);
  const { width } = useWindowDimensions();
  const safeAreaInsets = useSafeAreaInsets();
  const columnWidth = Math.max(292, Math.min(370, width - 72));
  const snapInterval = columnWidth + columnGap;
  const columnBottomPadding = safeAreaInsets.bottom + 150;
  const breadcrumbScrollRef = useRef<ScrollView>(null);
  const mapOffset = useRef(new Animated.Value(0)).current;
  const mapOffsetValue = useRef(0);
  const mapPanStartOffset = useRef(0);
  const requestSequence = useRef(0);
  const branchSwipeVisualClearTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const previousColumnsMapKey = useRef<string | undefined>(undefined);
  const [state, setState] = useState<InventoryMapState>({ status: 'loading' });
  const [openPath, setOpenPath] = useState<readonly string[]>([]);
  const [query, setQuery] = useState('');
  const [reduceMotionEnabled, setReduceMotionEnabled] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [pendingScrollLevel, setPendingScrollLevel] = useState<number | undefined>();
  const [highlightedAssetId, setHighlightedAssetId] = useState<string | undefined>();
  const [selectedAsset, setSelectedAsset] = useState<InventoryMapAssetViewModel | undefined>();
  const [branchSwipeVisual, setBranchSwipeVisual] = useState<BranchSwipeVisualState | undefined>();
  const [mapVerticalScrollLocked, setMapVerticalScrollLocked] = useState(false);
  const [exitingColumns, setExitingColumns] = useState<readonly RenderedInventoryMapColumn[]>([]);
  const previousColumns = useRef<readonly InventoryMapColumnViewModel[]>([]);
  const activeBranchSwipe = useRef<{
    readonly assetId: string;
    readonly fromLevel: number;
    readonly targetLevel: number;
  } | undefined>(undefined);
  const activeBranchSwipeDragX = useRef(0);

  useEffect(() => {
    let isCurrent = true;
    AccessibilityInfo.isReduceMotionEnabled().then((enabled) => {
      if (isCurrent) {
        setReduceMotionEnabled(enabled);
      }
    });
    const subscription = AccessibilityInfo.addEventListener('reduceMotionChanged', setReduceMotionEnabled);

    return () => {
      isCurrent = false;
      subscription.remove();
    };
  }, []);

  useEffect(() => () => {
    if (branchSwipeVisualClearTimer.current !== undefined) {
      clearTimeout(branchSwipeVisualClearTimer.current);
    }
  }, []);

  useEffect(() => {
    const requestId = nextRequestId(requestSequence);
    loadMap({ requestId, showLoading: true });
  }, [inventoryMapQuery]);

  useFocusEffect(useCallback(() => {
    if (state.status !== 'ready') {
      return;
    }

    const requestId = nextRequestId(requestSequence);
    loadMap({ preserveSelectedAsset: true, requestId, showLoading: false });
  }, [inventoryMapQuery, state.status]));

  const map = state.status === 'ready' ? state.map : undefined;
  const columns = useMemo(
    () => map ? buildInventoryMapColumns(map, openPath) : [],
    [map, openPath]
  );
  const assetsById = useMemo(
    () => new Map((map?.assets ?? []).map((asset) => [asset.id, asset])),
    [map]
  );
  const currentMapKey = map ? mapStorageKey(map) : undefined;
  const breadcrumbs = useMemo(
    () => map ? buildInventoryMapBreadcrumbs(map, openPath) : [],
    [map, openPath]
  );
  const maxMapLevel = Math.max(0, columns.length - 1);
  const maxMapOffset = maxMapLevel * snapInterval;
  const mapTranslateX = useMemo(() => Animated.multiply(mapOffset, -1), [mapOffset]);
  const renderedColumns = useMemo(
    () => [
      ...columns.map((column) => ({
        id: `active-${column.level.toString()}`,
        column,
        exiting: false
      })),
      ...exitingColumns
    ].sort((first, second) => {
      if (first.column.level !== second.column.level) {
        return first.column.level - second.column.level;
      }
      return first.exiting === second.exiting ? 0 : first.exiting ? 1 : -1;
    }),
    [columns, exitingColumns]
  );

  useEffect(() => {
    const listenerId = mapOffset.addListener(({ value }) => {
      mapOffsetValue.current = value;
    });

    return () => {
      mapOffset.removeListener(listenerId);
    };
  }, [mapOffset]);

  useEffect(() => {
    if (!map) {
      previousColumns.current = [];
      setExitingColumns([]);
      return;
    }
    pathStore.current.set(mapStorageKey(map), openPath);
  }, [map, openPath, pathStore]);

  useEffect(() => {
    const previousMapKey = previousColumnsMapKey.current;
    previousColumnsMapKey.current = currentMapKey;

    if (previousMapKey !== currentMapKey) {
      previousColumns.current = columns;
      setExitingColumns([]);
      return;
    }

    const previous = previousColumns.current;
    previousColumns.current = columns;

    if (!currentMapKey || reduceMotionEnabled || previous.length <= columns.length) {
      if (reduceMotionEnabled) {
        setExitingColumns([]);
      }
      return;
    }

    const removedColumns = previous.filter((column) => column.level >= columns.length && column.level > 0);
    if (removedColumns.length === 0) {
      return;
    }

    setExitingColumns((current) => [
      ...current.filter((renderedColumn) =>
        removedColumns.every((column) => column.key !== renderedColumn.column.key)
      ),
      ...removedColumns.map((column) => ({
        id: `exit-${column.level.toString()}-${column.key}`,
        column,
        exiting: true
      }))
    ]);
  }, [columns, currentMapKey, reduceMotionEnabled]);

  useEffect(() => {
    if (!map || columns.length === 0) {
      return;
    }
    if (activeBranchSwipe.current) {
      return;
    }
    const timer = setTimeout(() => {
      scrollToColumn(Math.max(0, columns.length - 1));
    }, 30);

    return () => {
      clearTimeout(timer);
    };
  }, [columns.length, map, maxMapOffset, mapOffset, reduceMotionEnabled, snapInterval]);

  useEffect(() => {
    if (pendingScrollLevel === undefined || !map) {
      return;
    }

    const timer = setTimeout(() => {
      scrollToColumn(pendingScrollLevel);
      scrollBreadcrumbsToActivePath();
      setPendingScrollLevel(undefined);
    }, 30);

    return () => {
      clearTimeout(timer);
    };
  }, [columns.length, map, maxMapOffset, mapOffset, pendingScrollLevel, reduceMotionEnabled, snapInterval]);

  useEffect(() => {
    if (activeBranchSwipe.current) {
      requestAnimationFrame(() => {
        driveBranchSwipeScroll(activeBranchSwipeDragX.current);
      });
    }
  }, [columns.length]);

  useEffect(() => {
    if (mapOffsetValue.current <= maxMapOffset) {
      return;
    }

    const clampedOffset = clampInventoryMapOffset({
      offset: mapOffsetValue.current,
      maxOffset: maxMapOffset
    });
    mapOffset.stopAnimation();
    mapOffset.setValue(clampedOffset);
    mapOffsetValue.current = clampedOffset;
  }, [mapOffset, maxMapOffset]);

  function loadMap({
    preserveSelectedAsset = false,
    requestId,
    showLoading
  }: LoadMapOptions): void {
    if (showLoading) {
      setState({ status: 'loading' });
    }

    inventoryMapQuery
      .execute()
      .then((nextMap) => {
        if (!isCurrentRequest(requestSequence, requestId)) {
          return;
        }
        setState({ status: 'ready', map: nextMap });
        setHighlightedAssetId(undefined);
        setSelectedAsset((current) => {
          if (!preserveSelectedAsset || !current) {
            return undefined;
          }
          return nextMap.assets.find((nextAsset) => nextAsset.id === current.id) ?? current;
        });
        setBranchSwipeVisual(undefined);
        setMapVerticalScrollLocked(false);
        setOpenPath(pathStore.current.get(mapStorageKey(nextMap)) ?? []);
      })
      .catch((error) => {
        if (isCurrentRequest(requestSequence, requestId)) {
          setState({
            status: 'error',
            message: error instanceof Error ? error.message : 'Inventory map failed to load.'
          });
        }
      })
      .finally(() => {
        if (isCurrentRequest(requestSequence, requestId)) {
          setIsRefreshing(false);
        }
      });
  }

  function refreshMap(options: { readonly preserveSelectedAsset?: boolean } = {}): void {
    const requestId = nextRequestId(requestSequence);
    setIsRefreshing(true);
    loadMap({
      preserveSelectedAsset: options.preserveSelectedAsset,
      requestId,
      showLoading: false
    });
  }

  function scrollToColumn(level: number): void {
    const targetOffset = clampInventoryMapOffset({
      offset: level * snapInterval,
      maxOffset: maxMapOffset
    });
    if (reduceMotionEnabled) {
      mapOffset.stopAnimation();
      mapOffset.setValue(targetOffset);
      mapOffsetValue.current = targetOffset;
      return;
    }

    Animated.timing(mapOffset, {
      duration: 240,
      easing: easeOutCubic,
      toValue: targetOffset,
      useNativeDriver: true
    }).start(({ finished }) => {
      if (finished) {
        mapOffsetValue.current = targetOffset;
      }
    });
  }

  function scrollBreadcrumbsToActivePath(): void {
    breadcrumbScrollRef.current?.scrollToEnd({ animated: !reduceMotionEnabled });
  }

  function selectBranch(asset: InventoryMapAssetViewModel): void {
    if (!map) {
      return;
    }

    if (!asset.canContainAssets) {
      setSelectedAsset(asset);
      return;
    }

    const nextPath = selectInventoryMapBranch(map, openPath, asset.id);
    setOpenPath(nextPath);
    setHighlightedAssetId(asset.id);
    setPendingScrollLevel(nextPath.length);
  }

  function beginBranchSwipe(asset: InventoryMapAssetViewModel, dragX: number): void {
    if (!map || !asset.canContainAssets || activeBranchSwipe.current?.assetId === asset.id) {
      return;
    }
    clearBranchSwipeVisualTimer();

    const nextPath = selectInventoryMapBranch(map, openPath, asset.id);
    const targetLevel = nextPath.length;
    activeBranchSwipe.current = {
      assetId: asset.id,
      fromLevel: Math.max(0, targetLevel - 1),
      targetLevel
    };
    activeBranchSwipeDragX.current = dragX;
    setBranchSwipeVisual({ assetId: asset.id, dragX });
    setMapVerticalScrollLocked(true);
    setOpenPath(nextPath);
    setHighlightedAssetId(asset.id);
    driveBranchSwipeScroll(dragX);
  }

  function driveBranchSwipeScroll(dragX: number): void {
    const activeSwipe = activeBranchSwipe.current;
    if (!activeSwipe) {
      return;
    }

    activeBranchSwipeDragX.current = dragX;
    const nextOffset = inventoryMapBranchSwipeOffset({
      dragX,
      fromLevel: activeSwipe.fromLevel,
      maxLevel: maxMapLevel,
      snapInterval
    });
    mapOffset.setValue(nextOffset);
    mapOffsetValue.current = nextOffset;
  }

  function finishBranchSwipe(options: FinishBranchSwipeOptions = {}): void {
    const activeSwipe = activeBranchSwipe.current;
    activeBranchSwipe.current = undefined;
    setMapVerticalScrollLocked(false);
    if (options.preserveVisual) {
      scheduleBranchSwipeVisualClear();
    } else {
      clearBranchSwipeVisualTimer();
      setBranchSwipeVisual(undefined);
    }
    if (!activeSwipe) {
      return;
    }

    setPendingScrollLevel(activeSwipe.targetLevel);
  }

  function scheduleBranchSwipeVisualClear(): void {
    clearBranchSwipeVisualTimer();
    branchSwipeVisualClearTimer.current = setTimeout(() => {
      branchSwipeVisualClearTimer.current = undefined;
      setBranchSwipeVisual(undefined);
    }, 280);
  }

  function clearBranchSwipeVisualTimer(): void {
    if (branchSwipeVisualClearTimer.current === undefined) {
      return;
    }

    clearTimeout(branchSwipeVisualClearTimer.current);
    branchSwipeVisualClearTimer.current = undefined;
  }

  function openBreadcrumb(level: number): void {
    const nextPath = pathForBreadcrumbLevel(openPath, level);
    setOpenPath(nextPath);
    setHighlightedAssetId(preserveInventoryMapHighlightForPath(nextPath, highlightedAssetId));
    setPendingScrollLevel(level);
  }

  function submitSearch(): void {
    if (!map) {
      return;
    }

    const match = findInventoryMapSearchMatch(map, query);
    if (!match) {
      setHighlightedAssetId(undefined);
      return;
    }

    setOpenPath(match.openPath);
    setHighlightedAssetId(match.assetId);
    setPendingScrollLevel(match.openPath.length);
  }

  function clearSearch(): void {
    setQuery('');
    setHighlightedAssetId(undefined);
  }

  function openAddHere(asset: InventoryMapAssetViewModel): void {
    router.push({
      pathname: '/add',
      params: addHereRouteParams({
        id: asset.id,
        title: asset.title,
        kind: asset.kind,
        kindLabel: asset.kindLabel,
        parentLocationTrailLabel: asset.parentPlacementLabel,
        locationTrailLabel: asset.placementLabel
      })
    });
  }

  const mapPanResponder = useMemo(
    () => PanResponder.create({
      onMoveShouldSetPanResponder: (_event, gestureState) =>
        shouldActivateInventoryMapPagerPan({
          dx: gestureState.dx,
          dy: gestureState.dy
        }),
      onPanResponderGrant: () => {
        mapOffset.stopAnimation();
        mapPanStartOffset.current = mapOffsetValue.current;
      },
      onPanResponderMove: (_event, gestureState) => {
        const nextOffset = clampInventoryMapOffset({
          offset: mapPanStartOffset.current - gestureState.dx,
          maxOffset: maxMapOffset
        });
        mapOffset.setValue(nextOffset);
        mapOffsetValue.current = nextOffset;
      },
      onPanResponderRelease: (_event, gestureState) => {
        const projectedOffset = mapOffsetValue.current
          - gestureState.vx * snapInterval * inventoryMapGestureConfig.mapPanVelocityProjection;
        const level = nearestInventoryMapColumnForOffset({
          offset: projectedOffset,
          maxLevel: maxMapLevel,
          snapInterval
        });
        scrollToColumn(level);
      },
      onPanResponderTerminate: () => {
        const level = nearestInventoryMapColumnForOffset({
          offset: mapOffsetValue.current,
          maxLevel: maxMapLevel,
          snapInterval
        });
        scrollToColumn(level);
      },
      onPanResponderTerminationRequest: () => true
    }),
    [mapOffset, maxMapLevel, maxMapOffset, snapInterval, reduceMotionEnabled]
  );

  return (
    <View style={styles.shell}>
      <View style={styles.header}>
        <View style={styles.headerTopRow}>
          <View style={styles.titleBlock}>
            <Text style={styles.title}>Browse</Text>
            {state.status === 'ready' ? (
              <Text numberOfLines={1} style={styles.overviewText}>{mapOverviewLabel(state.map)}</Text>
            ) : null}
          </View>
          <BrowseSurfaceControl palette={colors} selectedSurface={selectedSurface} onChangeSurface={onChangeSurface} />
        </View>
        <View style={styles.searchBar}>
          <Search color={colors.textMuted} size={19} strokeWidth={2.5} />
          <TextInput
            accessibilityLabel="Find in inventory map"
            autoCapitalize="none"
            onChangeText={setQuery}
            onSubmitEditing={submitSearch}
            placeholder="Find and expand path"
            placeholderTextColor={colors.textMuted}
            returnKeyType="search"
            style={styles.searchInput}
            value={query}
          />
          {query.length > 0 ? (
            <Pressable
              accessibilityLabel="Clear map search"
              accessibilityRole="button"
              hitSlop={10}
              onPress={clearSearch}
              style={styles.iconButton}
            >
              <X color={colors.textMuted} size={18} strokeWidth={2.5} />
            </Pressable>
          ) : null}
        </View>
        {state.status === 'ready' ? (
          <>
            <ScrollView
              horizontal
              ref={breadcrumbScrollRef}
              showsHorizontalScrollIndicator={false}
              contentContainerStyle={styles.breadcrumbs}
            >
              {breadcrumbs.map((breadcrumb, index) => (
                <View key={breadcrumb.key} style={styles.breadcrumbItem}>
                  {index > 0 ? <ChevronRight color={colors.textMuted} size={14} strokeWidth={2.5} /> : null}
                  <Pressable
                    accessibilityLabel={`Open location ${breadcrumb.title}`}
                    accessibilityRole="button"
                    onPress={() => openBreadcrumb(breadcrumb.level)}
                    style={({ pressed }) => [
                      styles.breadcrumbButton,
                      pressed ? styles.breadcrumbButtonPressed : null
                    ]}
                  >
                    <Text numberOfLines={1} style={styles.breadcrumbText}>{breadcrumb.title}</Text>
                  </Pressable>
                </View>
              ))}
            </ScrollView>
          </>
        ) : null}
      </View>
      {state.status === 'loading' ? (
        <View style={styles.centerState}>
          <ActivityIndicator color={colors.accent} />
          <Text style={styles.centerText}>Loading map</Text>
        </View>
      ) : null}
      {state.status === 'error' ? (
        <View style={styles.centerState}>
          <Text style={styles.errorTitle}>Map unavailable</Text>
          <Text style={styles.centerText}>{state.message}</Text>
        </View>
      ) : null}
      {state.status === 'ready' ? (
        <View style={styles.mapScroller} {...mapPanResponder.panHandlers}>
          <Animated.View
            style={[
              styles.mapContent,
              {
                columnGap,
                paddingHorizontal: horizontalInset,
                transform: [{ translateX: mapTranslateX }]
              }
            ]}
          >
            {renderedColumns.map((renderedColumn) => (
              <InventoryMapColumn
                branchSwipeVisual={branchSwipeVisual}
                column={renderedColumn.column}
                columnBottomPadding={columnBottomPadding}
                columnWidth={columnWidth}
                exiting={renderedColumn.exiting}
                highlightedAssetId={highlightedAssetId}
                isRefreshing={isRefreshing}
                key={renderedColumn.id}
                mapVerticalScrollLocked={mapVerticalScrollLocked}
                onBranchSwipeBegin={beginBranchSwipe}
                onColumnExitComplete={() => {
                  setExitingColumns((current) =>
                    current.filter((column) => column.id !== renderedColumn.id)
                  );
                }}
                onAddHere={openAddHere}
                onBranchSwipeFinish={finishBranchSwipe}
                onBranchSwipeProgress={driveBranchSwipeScroll}
                onOpenInfo={setSelectedAsset}
                onPressAsset={selectBranch}
                onRefresh={refreshMap}
                openPath={openPath}
                assetsById={assetsById}
                reduceMotionEnabled={reduceMotionEnabled}
              />
            ))}
          </Animated.View>
        </View>
      ) : null}
      <InventoryMapInfoSheet
        addAssetPhotosCommand={addAssetPhotosCommand}
        assetCheckoutCommand={assetCheckoutCommand}
        asset={selectedAsset}
        assetDetailQuery={assetDetailQuery}
        assetLifecycleCommand={assetLifecycleCommand}
        deleteAssetPhotoCommand={deleteAssetPhotoCommand}
        photoSelectionQuery={photoSelectionQuery}
        onClose={() => setSelectedAsset(undefined)}
        onMapChanged={() => refreshMap({ preserveSelectedAsset: true })}
      />
    </View>
  );
}

export function BrowseSurfaceControl({
  palette,
  selectedSurface,
  onChangeSurface
}: {
  readonly palette: MobileColorPalette;
  readonly selectedSurface: InventoryMapSurface;
  readonly onChangeSurface: (surface: InventoryMapSurface) => void;
}) {
  const styles = createStyles(palette);
  return (
    <View accessibilityLabel="Browse view" accessibilityRole="tablist" style={[styles.surfaceControl, { backgroundColor: palette.surfaceMuted }]}>
      {buildBrowseSurfaceOptions().map((option) => {
        const selected = option.value === selectedSurface;
        return (
          <Pressable
            accessibilityRole="tab"
            accessibilityState={{ selected }}
            key={option.value}
            onPress={() => onChangeSurface(option.value)}
            style={({ pressed }) => [
              styles.surfaceButton,
              selected ? [styles.surfaceButtonSelected, { backgroundColor: palette.surface }] : null,
              pressed ? { backgroundColor: palette.selected } : null
            ]}
          >
            <Text style={[styles.surfaceText, { color: selected ? palette.text : palette.textMuted }]}>
              {option.label}
            </Text>
          </Pressable>
        );
      })}
    </View>
  );
}

function InventoryMapColumn({
  branchSwipeVisual,
  column,
  columnBottomPadding,
  columnWidth,
  exiting,
  highlightedAssetId,
  isRefreshing,
  mapVerticalScrollLocked,
  onAddHere,
  onBranchSwipeBegin,
  onColumnExitComplete,
  onBranchSwipeFinish,
  onBranchSwipeProgress,
  onOpenInfo,
  onPressAsset,
  onRefresh,
  openPath,
  assetsById,
  reduceMotionEnabled
}: {
  readonly branchSwipeVisual?: BranchSwipeVisualState;
  readonly column: InventoryMapColumnViewModel;
  readonly columnBottomPadding: number;
  readonly columnWidth: number;
  readonly exiting: boolean;
  readonly highlightedAssetId?: string;
  readonly isRefreshing: boolean;
  readonly mapVerticalScrollLocked: boolean;
  readonly onAddHere: (asset: InventoryMapAssetViewModel) => void;
  readonly onBranchSwipeBegin: (asset: InventoryMapAssetViewModel, dragX: number) => void;
  readonly onColumnExitComplete: () => void;
  readonly onBranchSwipeFinish: (options?: FinishBranchSwipeOptions) => void;
  readonly onBranchSwipeProgress: (dragX: number) => void;
  readonly onOpenInfo: (asset: InventoryMapAssetViewModel) => void;
  readonly onPressAsset: (asset: InventoryMapAssetViewModel) => void;
  readonly onRefresh: () => void;
  readonly openPath: readonly string[];
  readonly assetsById: ReadonlyMap<string, InventoryMapAssetViewModel>;
  readonly reduceMotionEnabled: boolean;
}) {
  const colors = useAppearancePalette();
  const styles = useMemo(() => createStyles(colors), [colors]);
  const entryProgress = useRef(new Animated.Value(reduceMotionEnabled ? 1 : 0)).current;
  const transitionSequence = useRef(0);
  const onColumnExitCompleteRef = useRef(onColumnExitComplete);
  const [displayedColumn, setDisplayedColumn] = useState(column);
  const [motionActive, setMotionActive] = useState(!reduceMotionEnabled && column.level > 0);
  const entryTranslateX = entryProgress.interpolate({
    inputRange: [0, 1],
    outputRange: [displayedColumn.level === 0 || reduceMotionEnabled ? 0 : 18, 0]
  });
  const displayedParentAsset = displayedColumn.parentId ? assetsById.get(displayedColumn.parentId) : undefined;
  const emptyAction = buildInventoryMapEmptyColumnAction(displayedParentAsset);

  useEffect(() => {
    onColumnExitCompleteRef.current = onColumnExitComplete;
  }, [onColumnExitComplete]);

  useEffect(() => {
    if (exiting) {
      return;
    }

    if (reduceMotionEnabled) {
      setMotionActive(false);
      entryProgress.setValue(1);
      return;
    }

    setMotionActive(column.level > 0);
    entryProgress.setValue(0);
    Animated.timing(entryProgress, {
      duration: 180,
      easing: easeOutCubic,
      toValue: 1,
      useNativeDriver: true
    }).start(({ finished }) => {
      if (finished) {
        setMotionActive(false);
      }
    });
  }, [column.level, entryProgress, exiting, reduceMotionEnabled]);

  useEffect(() => {
    if (exiting) {
      return;
    }

    if (displayedColumn.key === column.key) {
      if (displayedColumn !== column) {
        setDisplayedColumn(column);
      }
      return;
    }

    const transitionId = transitionSequence.current + 1;
    transitionSequence.current = transitionId;

    if (reduceMotionEnabled) {
      setDisplayedColumn(column);
      setMotionActive(false);
      entryProgress.setValue(1);
      return;
    }

    setMotionActive(true);
    entryProgress.stopAnimation();
    Animated.timing(entryProgress, {
      duration: 110,
      easing: easeOutCubic,
      toValue: 0,
      useNativeDriver: true
    }).start(({ finished }) => {
      if (!finished || transitionSequence.current !== transitionId) {
        return;
      }

      setDisplayedColumn(column);
      entryProgress.setValue(0);
      Animated.timing(entryProgress, {
        duration: 170,
        easing: easeOutCubic,
        toValue: 1,
        useNativeDriver: true
      }).start(({ finished: enterFinished }) => {
        if (enterFinished && transitionSequence.current === transitionId) {
          setMotionActive(false);
        }
      });
    });
  }, [column, displayedColumn.key, entryProgress, exiting, reduceMotionEnabled]);

  useEffect(() => {
    if (!exiting) {
      return;
    }

    if (reduceMotionEnabled) {
      setMotionActive(false);
      onColumnExitCompleteRef.current();
      return;
    }

    setMotionActive(true);
    const transitionId = transitionSequence.current + 1;
    transitionSequence.current = transitionId;
    entryProgress.stopAnimation();
    entryProgress.setValue(1);
    Animated.timing(entryProgress, {
      duration: 150,
      easing: easeOutCubic,
      toValue: 0,
      useNativeDriver: true
    }).start(({ finished }) => {
      if (finished && transitionSequence.current === transitionId) {
        onColumnExitCompleteRef.current();
      }
    });
  }, [entryProgress, exiting, reduceMotionEnabled]);

  return (
    <Animated.View
      pointerEvents={exiting ? 'none' : 'auto'}
      style={[
        styles.column,
        motionActive ? {
          opacity: entryProgress,
          transform: [{ translateX: entryTranslateX }]
        } : null,
        {
          width: columnWidth
        }
      ]}
    >
      <Text numberOfLines={1} style={styles.columnTitle}>{displayedColumn.title}</Text>
      <FlatList
        data={displayedColumn.assets}
        keyExtractor={(asset) => asset.id}
        keyboardShouldPersistTaps="handled"
        scrollEnabled={!mapVerticalScrollLocked && !exiting}
        showsVerticalScrollIndicator={false}
        style={styles.columnListSurface}
        contentContainerStyle={[
          styles.columnList,
          { paddingBottom: columnBottomPadding }
        ]}
        ListEmptyComponent={
          <View style={styles.emptyColumn}>
            <Package color={colors.accent} size={22} strokeWidth={2.4} />
            <Text style={styles.emptyColumnText}>{displayedColumn.emptyLabel}</Text>
            {emptyAction && displayedParentAsset ? (
              <Pressable
                accessibilityRole="button"
                onPress={() => onAddHere(displayedParentAsset)}
                style={styles.emptyColumnAction}
              >
                <Plus color={colors.action} size={16} strokeWidth={2.6} />
                <Text style={styles.emptyColumnActionText}>{emptyAction.label}</Text>
              </Pressable>
            ) : null}
          </View>
        }
        renderItem={({ item: asset }) => {
          const rowState = buildInventoryMapRowInteractionState(openPath, asset.id, highlightedAssetId);

          return (
            <InventoryMapRow
              asset={asset}
              expanded={rowState.expanded}
              highlighted={rowState.highlighted}
              onBranchSwipeBegin={(dragX) => onBranchSwipeBegin(asset, dragX)}
              onBranchSwipeFinish={onBranchSwipeFinish}
              onBranchSwipeProgress={onBranchSwipeProgress}
              onOpenInfo={() => onOpenInfo(asset)}
              onPress={() => onPressAsset(asset)}
              reduceMotionEnabled={reduceMotionEnabled}
              swipeDragX={branchSwipeVisual?.assetId === asset.id ? branchSwipeVisual.dragX : undefined}
            />
          );
        }}
        refreshControl={
          <RefreshControl
            refreshing={isRefreshing}
            tintColor={colors.action}
            onRefresh={onRefresh}
          />
        }
      />
    </Animated.View>
  );
}

function InventoryMapRow({
  asset,
  expanded,
  highlighted,
  onBranchSwipeBegin,
  onBranchSwipeFinish,
  onBranchSwipeProgress,
  onOpenInfo,
  onPress,
  reduceMotionEnabled,
  swipeDragX
}: {
  readonly asset: InventoryMapAssetViewModel;
  readonly expanded: boolean;
  readonly highlighted: boolean;
  readonly onBranchSwipeBegin: (dragX: number) => void;
  readonly onBranchSwipeFinish: (options?: FinishBranchSwipeOptions) => void;
  readonly onBranchSwipeProgress: (dragX: number) => void;
  readonly onOpenInfo: () => void;
  readonly onPress: () => void;
  readonly reduceMotionEnabled: boolean;
  readonly swipeDragX?: number;
}) {
  const colors = useAppearancePalette();
  const styles = useMemo(() => createStyles(colors), [colors]);
  const controlledSwipeOffset = swipeDragX === undefined
    ? 0
    : Math.max(-inventoryMapGestureConfig.branchSwipeRevealWidth, Math.min(0, swipeDragX));
  const rowOffset = useRef(new Animated.Value(controlledSwipeOffset)).current;
  const [rowSwipeActive, setRowSwipeActive] = useState(false);
  const rowBranchSelected = useRef(false);
  const hadControlledSwipe = useRef(swipeDragX !== undefined);
  const onBranchSwipeBeginRef = useRef(onBranchSwipeBegin);
  const onBranchSwipeFinishRef = useRef(onBranchSwipeFinish);
  const onBranchSwipeProgressRef = useRef(onBranchSwipeProgress);
  const canSwipeBranch = asset.canContainAssets;
  const rowAccessibilityLabel = asset.canContainAssets
    ? `${asset.title}, ${asset.kindLabel}, ${asset.childCount.toString()} inside`
    : `${asset.title}, ${asset.kindLabel}`;
  const rowAccessibilityHint = asset.canContainAssets
    ? 'Opens the next containment column. Swipe left to open this branch.'
    : 'Shows item details.';

  const resetRowOffset = useCallback(() => {
    if (reduceMotionEnabled) {
      rowOffset.setValue(0);
      return;
    }

    Animated.spring(rowOffset, {
      damping: 18,
      mass: 0.8,
      stiffness: 220,
      toValue: 0,
      useNativeDriver: true
    }).start();
  }, [reduceMotionEnabled, rowOffset]);

  useEffect(() => {
    onBranchSwipeBeginRef.current = onBranchSwipeBegin;
    onBranchSwipeFinishRef.current = onBranchSwipeFinish;
    onBranchSwipeProgressRef.current = onBranchSwipeProgress;
  }, [onBranchSwipeBegin, onBranchSwipeFinish, onBranchSwipeProgress]);

  useEffect(() => {
    if (swipeDragX !== undefined) {
      hadControlledSwipe.current = true;
      rowOffset.setValue(controlledSwipeOffset);
      return;
    }

    if (!hadControlledSwipe.current) {
      return;
    }

    hadControlledSwipe.current = false;
    resetRowOffset();
  }, [controlledSwipeOffset, resetRowOffset, rowOffset, swipeDragX]);

  const panResponder = useMemo(
    () => PanResponder.create({
      onMoveShouldSetPanResponder: (_event, gestureState) =>
        shouldSuppressInventoryMapScrollForBranchSwipe({
          canContainAssets: canSwipeBranch,
          dx: gestureState.dx,
          dy: gestureState.dy
        }),
      onMoveShouldSetPanResponderCapture: (_event, gestureState) =>
        shouldSuppressInventoryMapScrollForBranchSwipe({
          canContainAssets: canSwipeBranch,
          dx: gestureState.dx,
          dy: gestureState.dy
        }),
      onPanResponderGrant: () => {
        rowBranchSelected.current = false;
        rowOffset.stopAnimation();
        setRowSwipeActive(true);
      },
      onPanResponderMove: (_event, gestureState) => {
        const dragX = Math.min(0, gestureState.dx);
        rowOffset.setValue(Math.max(-inventoryMapGestureConfig.branchSwipeRevealWidth, dragX));

        if (!rowBranchSelected.current && shouldSelectInventoryMapBranchDuringSwipe({ dx: dragX })) {
          rowBranchSelected.current = true;
          onBranchSwipeBeginRef.current(dragX);
          return;
        }

        if (rowBranchSelected.current) {
          onBranchSwipeProgressRef.current(dragX);
        }
      },
      onPanResponderRelease: (_event, gestureState) => {
        setRowSwipeActive(false);
        if (rowBranchSelected.current) {
          rowBranchSelected.current = false;
          onBranchSwipeFinishRef.current({ preserveVisual: true });
          return;
        }

        resetRowOffset();
      },
      onPanResponderTerminate: () => {
        if (rowBranchSelected.current) {
          rowBranchSelected.current = false;
          setRowSwipeActive(false);
          onBranchSwipeFinishRef.current({ preserveVisual: true });
          return;
        }

        rowBranchSelected.current = false;
        setRowSwipeActive(false);
        onBranchSwipeFinishRef.current();
        resetRowOffset();
      },
      onPanResponderTerminationRequest: () => false
    }),
    [
      canSwipeBranch,
      resetRowOffset,
      rowOffset
    ]
  );

  return (
    <Animated.View
      style={[
        styles.mapRow,
        expanded ? styles.mapRowExpanded : null,
        highlighted ? styles.mapRowHighlighted : null
      ]}
    >
      {canSwipeBranch ? (
        <View pointerEvents="none" style={styles.rowSwipeUnderlay}>
          <ChevronRight color={colors.onAction} size={22} strokeWidth={3} />
        </View>
      ) : null}
      <Animated.View
        style={[
          styles.rowCard,
          { transform: [{ translateX: rowOffset }] }
        ]}
      >
        <Animated.View
          {...panResponder.panHandlers}
          onTouchStart={() => {
            if (canSwipeBranch) {
              rowOffset.stopAnimation();
            }
          }}
          style={styles.rowMainGesture}
        >
          <Pressable
            accessibilityHint={rowAccessibilityHint}
            accessibilityLabel={rowAccessibilityLabel}
            accessibilityRole="button"
            accessibilityState={asset.canContainAssets ? { expanded, selected: highlighted } : { selected: highlighted }}
            disabled={rowSwipeActive}
            onPress={onPress}
            style={styles.rowMain}
          >
            <View style={styles.rowImageWrap}>
              <View style={styles.rowImageFrame}>
                {asset.photo ? (
                  <Image
                    accessibilityIgnoresInvertColors
                    source={{ uri: asset.photo.uri, headers: asset.photo.headers }}
                    style={styles.rowImage}
                  />
                ) : (
                  <Text style={styles.rowImageLabel}>{asset.imagePlaceholderLabel}</Text>
                )}
              </View>
              {asset.childCount > 0 ? (
                <View style={styles.childCountBadge}>
                  <Package color={colors.onAction} size={11} strokeWidth={2.7} />
                  <Text style={styles.childCountBadgeText}>{asset.childCount.toString()}</Text>
                </View>
              ) : null}
            </View>
            <View style={styles.rowText}>
              <View style={styles.rowTitleLine}>
                <Text numberOfLines={1} style={styles.rowTitle}>{asset.title}</Text>
              </View>
              <Text numberOfLines={1} style={styles.rowMeta}>
                {asset.kindLabel}{asset.customTypeLabel ? ` · ${asset.customTypeLabel}` : ''}{asset.checkedOutLabel ? ` · ${asset.checkedOutLabel}` : ''}
              </Text>
              <Text numberOfLines={1} style={styles.rowTrail}>{asset.placementLabel}</Text>
            </View>
          </Pressable>
        </Animated.View>
        <Pressable
          accessibilityLabel={`Show details for ${asset.title}`}
          accessibilityRole="button"
          hitSlop={8}
          onPress={onOpenInfo}
          style={styles.rowInfoButton}
        >
          <Info color={colors.textMuted} size={20} strokeWidth={2.5} />
        </Pressable>
      </Animated.View>
    </Animated.View>
  );
}

function InventoryMapInfoSheet({
  addAssetPhotosCommand,
  asset,
  assetCheckoutCommand,
  assetDetailQuery,
  assetLifecycleCommand,
  deleteAssetPhotoCommand,
  photoSelectionQuery,
  onClose,
  onMapChanged
}: {
  readonly addAssetPhotosCommand: AddAssetPhotosCommand;
  readonly assetCheckoutCommand: AssetCheckoutCommand;
  readonly asset?: InventoryMapAssetViewModel;
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetLifecycleCommand: AssetLifecycleCommand;
  readonly deleteAssetPhotoCommand: DeleteAssetPhotoCommand;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly onClose: () => void;
  readonly onMapChanged: () => void;
}) {
  const colors = useAppearancePalette();
  const styles = useMemo(() => createStyles(colors), [colors]);
  const feedback = useAppFeedback();
  const [detailState, setDetailState] = useState<MapSheetDetailState>({ status: 'idle' });
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [pendingAction, setPendingAction] = useState<PendingMapSheetAction | undefined>();
  const [failedPhotoDrafts, setFailedPhotoDrafts] = useState<readonly SelectedAssetPhoto[]>([]);
  const [photoUploads, setPhotoUploads] = useState<readonly AssetPhotoUploadProgressViewModel[]>([]);
  const [photoStatus, setPhotoStatus] = useState<AddAssetPhotosCommandResult | undefined>();
  const [workspaceStatus, setWorkspaceStatus] = useState<AssetWorkspaceStatus | undefined>();
  const [selectedPhotoId, setSelectedPhotoId] = useState<string | undefined>();

  const activeDetailAssetId = detailState.status === 'ready' ? detailState.asset.id : asset?.id;

  useEffect(() => {
    if (!asset) {
      setDetailState({ status: 'idle' });
      setIsRefreshing(false);
      setPendingAction(undefined);
      setFailedPhotoDrafts([]);
      setPhotoUploads([]);
      setPhotoStatus(undefined);
      setWorkspaceStatus(undefined);
      setSelectedPhotoId(undefined);
      return;
    }

    let isCurrent = true;
    setDetailState({ status: 'loading' });
    assetDetailQuery
      .execute(asset.id, { source: 'map' })
      .then((detail) => {
        if (isCurrent) {
          setDetailState({ status: 'ready', asset: detail });
        }
      })
      .catch((error) => {
        if (isCurrent) {
          setDetailState({
            status: 'error',
            message: error instanceof Error ? error.message : 'Could not load asset details.'
          });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [asset, assetDetailQuery]);

  async function reloadDetail(): Promise<AssetDetailViewModel> {
    if (!activeDetailAssetId) {
      throw new Error('No asset selected.');
    }

    const detail = await assetDetailQuery.execute(activeDetailAssetId, { source: 'map' });
    setDetailState({ status: 'ready', asset: detail });
    return detail;
  }

  async function refreshDetail(): Promise<void> {
    setIsRefreshing(true);
    setWorkspaceStatus(undefined);

    try {
      await reloadDetail();
      onMapChanged();
    } catch (error) {
      feedback.showNotice({
        tone: 'error',
        title: 'Could not refresh asset',
        message: readableError(error, 'Could not refresh asset.')
      });
    } finally {
      setIsRefreshing(false);
    }
  }

  function choosePhotos(currentPhotoCount: number): void {
    showPhotoSourceChooser({
      onCamera: () => {
        void addPhotos('camera', currentPhotoCount);
      },
      onLibrary: () => {
        void addPhotos('library', currentPhotoCount);
      }
    });
  }

  async function addPhotos(source: 'camera' | 'library', currentPhotoCount: number): Promise<void> {
    if (!activeDetailAssetId) {
      return;
    }

    setPendingAction('photos');
    try {
      const photos = source === 'camera'
        ? await photoSelectionQuery.captureFromCamera(currentPhotoCount)
        : await photoSelectionQuery.selectFromLibrary(currentPhotoCount);
      if (photos.length === 0) {
        return;
      }
      setPhotoStatus(undefined);
      setPhotoUploads(photoUploadRows(photos));
      const result = await addAssetPhotosCommand.execute({
        assetId: activeDetailAssetId,
        photos,
        onPhotoProgress: updatePhotoUploadProgress
      });
      setPhotoStatus(result);
      setFailedPhotoDrafts(result.failedPhotos as readonly SelectedAssetPhoto[]);
      await reloadDetail();
      onMapChanged();
      if (result.failedCount === 0) {
        setPhotoUploads([]);
      }
    } catch (error) {
      feedback.showNotice({
        tone: 'error',
        title: 'Could not add photos',
        message: readableError(error, 'Photo upload failed.')
      });
    } finally {
      setPendingAction(undefined);
    }
  }

  async function retryPhotos(): Promise<void> {
    if (!activeDetailAssetId || failedPhotoDrafts.length === 0) {
      return;
    }

    setPendingAction('photos');
    try {
      setPhotoStatus(undefined);
      setPhotoUploads(photoUploadRows(failedPhotoDrafts));
      const result = await addAssetPhotosCommand.execute({
        assetId: activeDetailAssetId,
        photos: failedPhotoDrafts,
        onPhotoProgress: updatePhotoUploadProgress
      });
      setPhotoStatus(result);
      setFailedPhotoDrafts(result.failedPhotos as readonly SelectedAssetPhoto[]);
      await reloadDetail();
      onMapChanged();
      if (result.failedCount === 0) {
        setPhotoUploads([]);
      }
    } catch (error) {
      feedback.showNotice({
        tone: 'error',
        title: 'Could not retry photos',
        message: readableError(error, 'Photo retry failed.')
      });
    } finally {
      setPendingAction(undefined);
    }
  }

  async function removePhoto(photoId: string): Promise<void> {
    if (!activeDetailAssetId) {
      return;
    }

    setPendingAction('photos');
    try {
      const result = await deleteAssetPhotoCommand.execute({ assetId: activeDetailAssetId, photoId });
      setPhotoStatus({
        attachedCount: 0,
        failedCount: 0,
        failedPhotos: [],
        message: result.message,
        canRetry: false
      });
      setFailedPhotoDrafts([]);
      setSelectedPhotoId(undefined);
      setPhotoUploads([]);
      await reloadDetail();
      onMapChanged();
    } catch (error) {
      feedback.showNotice({
        tone: 'error',
        title: 'Could not remove photo',
        message: readableError(error, 'Photo removal failed.')
      });
    } finally {
      setPendingAction(undefined);
    }
  }

  function updatePhotoUploadProgress(event: AddAssetPhotoProgressEvent): void {
    setPhotoUploads((current) => applyPhotoUploadProgress(current, event));
  }

  function selectAssetPhoto(detail: AssetDetailViewModel, photoId: string): void {
    if (!isAssetPhotoId(detail.photos, photoId)) {
      return;
    }

    setSelectedPhotoId(photoId);
  }

  function openEmbeddedDetail(assetId: string): void {
    const request = inventoryMapEmbeddedDetailRequest(assetId);
    setSelectedPhotoId(undefined);
    setPhotoUploads([]);
    setPhotoStatus(undefined);
    setWorkspaceStatus(undefined);
    setDetailState({ status: 'loading' });
    assetDetailQuery
      .execute(request.assetId, request.options)
      .then((nextDetail) => setDetailState({ status: 'ready', asset: nextDetail }))
      .catch((error) => setDetailState({
        status: 'error',
        message: readableError(error, 'Could not load asset details.')
      }));
  }

  function showMoreActions(detail: AssetDetailViewModel): void {
    if (assetDetailOverflowControlState(pendingAction !== undefined).disabled) {
      return;
    }
    const overflow = assetLifecycleOverflowMenu(detail);
    if (Platform.OS === 'ios') {
      ActionSheetIOS.showActionSheetWithOptions(
        {
          title: overflow.title,
          message: overflow.message,
          options: [...overflow.options],
          cancelButtonIndex: overflow.cancelIndex,
          destructiveButtonIndex: overflow.destructiveIndex
        },
        (index) => {
          const action = overflow.actionRows[index];
          if (action) {
            confirmLifecycleAction(action.kind, detail);
            return;
          }
          if (index === overflow.checkoutHistoryIndex) {
            onClose();
            router.push(`/assets/${detail.id}/checkouts`);
            return;
          }
          if (index === overflow.auditIndex) {
            onClose();
            router.push(`/assets/${detail.id}/audit`);
          }
        }
      );
      return;
    }

    Alert.alert(overflow.title, overflow.message, [
      ...assetLifecycleActionRows(detail).map((action) => ({
        text: action.label,
        style: action.isDestructive ? 'destructive' as const : 'default' as const,
        onPress: () => confirmLifecycleAction(action.kind, detail)
      })),
      {
        text: 'Checkout history',
        onPress: () => {
          onClose();
          router.push(`/assets/${detail.id}/checkouts`);
        }
      },
      {
        text: 'Audit history',
        onPress: () => {
          onClose();
          router.push(`/assets/${detail.id}/audit`);
        }
      },
      { text: 'Cancel', style: 'cancel' }
    ]);
  }

  function confirmLifecycleAction(action: AssetLifecycleActionKind, detail: AssetDetailViewModel): void {
    const confirmation = assetLifecycleConfirmation(action, detail);
    Alert.alert(confirmation.title, confirmation.message, [
      { text: 'Cancel', style: 'cancel' },
      {
        text: confirmation.confirmLabel,
        style: confirmation.isDestructive ? 'destructive' : 'default',
        onPress: () => void runLifecycleAction(action, detail)
      }
    ]);
  }

  async function runLifecycleAction(action: AssetLifecycleActionKind, detail: AssetDetailViewModel): Promise<void> {
    setPendingAction(action);
    setWorkspaceStatus(undefined);

    try {
      await assetLifecycleCommand.execute({ action, assetId: detail.id });
      if (action === 'delete') {
        onMapChanged();
        onClose();
        feedback.showNotice({
          tone: 'success',
          title: 'Asset deleted',
          message: `${detail.title} was permanently deleted.`
        });
        return;
      }

      await reloadDetail();
      onMapChanged();
      setWorkspaceStatus(assetWorkspaceSuccessStatus(action, detail));
    } catch (error) {
      const failure = assetLifecycleFailurePresentation(
        action,
        detail,
        readableError(error, 'Lifecycle action failed.')
      );
      feedback.showNotice({
        tone: 'error',
        title: failure.title,
        message: failure.message
      });
    } finally {
      setPendingAction(undefined);
    }
  }

  async function runCheckoutAction(action: 'checkout' | 'return', detail: AssetDetailViewModel): Promise<void> {
    setPendingAction(action);
    setWorkspaceStatus(undefined);

    try {
      await assetCheckoutCommand.execute({ action, assetId: detail.id });
      await reloadDetail();
      onMapChanged();
      setWorkspaceStatus(assetWorkspaceSuccessStatus(action, detail));
    } catch (error) {
      feedback.showNotice({
        tone: 'error',
        title: action === 'checkout' ? 'Checkout failed' : 'Return failed',
        message: readableError(error, action === 'checkout'
          ? 'Could not check out this asset.'
          : 'Could not return this asset.')
      });
    } finally {
      setPendingAction(undefined);
    }
  }

  return (
    <Modal
      animationType="slide"
      onRequestClose={onClose}
      presentationStyle="pageSheet"
      visible={asset !== undefined}
    >
      {asset ? (
        <View style={styles.sheet}>
          <View style={styles.sheetHandle} />
          <View style={styles.sheetTopBar}>
            <Pressable accessibilityRole="button" onPress={onClose} style={styles.sheetCloseButton}>
              <Text style={styles.sheetCloseText}>Done</Text>
            </Pressable>
          </View>
          {detailState.status === 'loading' ? (
            <View style={styles.sheetLoadingState}>
              <ActivityIndicator color={colors.accent} />
              <Text style={styles.centerText}>Loading details</Text>
            </View>
          ) : null}
          {detailState.status === 'error' ? (
            <View style={styles.sheetLoadingState}>
              <Text style={styles.errorTitle}>Details unavailable</Text>
              <Text style={styles.centerText}>{detailState.message}</Text>
            </View>
          ) : null}
          {detailState.status === 'ready' ? (
            <>
              {(() => {
                const photoViewer = assetPhotoViewerModel(detailState.asset.photos, selectedPhotoId);
                return (
                  <AssetPhotoViewerSheet
                    canRemove={detailState.asset.canAddPhotos}
                    model={photoViewer}
                    onClose={() => setSelectedPhotoId(undefined)}
                    onRemove={(photoId) => void removePhoto(photoId)}
                    onSelectPhoto={setSelectedPhotoId}
                    photos={detailState.asset.photos}
                  />
                );
              })()}
              <AssetDetailView
                asset={detailState.asset}
                canRetryPhotos={photoStatus?.canRetry}
                isActionPending={pendingAction !== undefined}
                onAddHere={detailState.asset.canAddContainedAssets ? () => {
                  onClose();
                  router.push({
                    pathname: '/add',
                    params: addHereRouteParams(detailState.asset)
                  });
                } : undefined}
                onAddPhotos={() => choosePhotos(detailState.asset.photos.length)}
                onChildPress={openEmbeddedDetail}
                onEdit={() => {
                  onClose();
                  router.push(`/assets/${detailState.asset.id}/edit`);
                }}
                onCheckout={() => void runCheckoutAction('checkout', detailState.asset)}
                onMoreActions={() => showMoreActions(detailState.asset)}
                onMove={() => {
                  onClose();
                  router.push(`/assets/${detailState.asset.id}/move`);
                }}
                onMoveThingsHere={detailState.asset.canAddContainedAssets ? () => {
                  onClose();
                  router.push(`/assets/${detailState.asset.id}/move-here`);
                } : undefined}
                onPhotoPress={(photoId) => selectAssetPhoto(detailState.asset, photoId)}
                onParentLocationPress={(parent) => openEmbeddedDetail(parent.id)}
                onRetryPhotos={() => void retryPhotos()}
                onReturn={() => void runCheckoutAction('return', detailState.asset)}
                photoUploads={photoUploads}
                photoStatusMessage={pendingAction === 'photos' ? 'Updating photos...' : photoStatus?.message}
                refreshControl={
                  <RefreshControl
                    refreshing={isRefreshing}
                    tintColor={colors.action}
                    onRefresh={refreshDetail}
                  />
                }
                workspaceStatusKind={visibleAssetWorkspaceStatus(pendingAction, workspaceStatus)?.kind}
                workspaceStatusMessage={visibleAssetWorkspaceStatus(pendingAction, workspaceStatus)?.message}
              />
            </>
          ) : null}
        </View>
      ) : null}
    </Modal>
  );
}

type PendingMapSheetAction = 'archive' | 'restore' | 'delete' | 'photos' | 'checkout' | 'return';

function mapStorageKey(map: InventoryMapViewModel): string {
  return `${map.sessionScopeId}:${map.tenantId}:${map.inventoryId}`;
}

function nextRequestId(requestSequence: { current: number }): number {
  requestSequence.current += 1;
  return requestSequence.current;
}

function isCurrentRequest(
  requestSequence: { readonly current: number },
  requestId: number
): boolean {
  return requestSequence.current === requestId;
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  shell: {
    backgroundColor: colors.background,
    flex: 1
  },
  header: {
    paddingHorizontal: spacing.md,
    paddingTop: spacing.sm
  },
  headerTopRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.md,
    marginBottom: spacing.xs
  },
  titleBlock: {
    flex: 1,
    minWidth: 0
  },
  title: {
    color: colors.text,
    fontSize: 25,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 30
  },
  surfaceControl: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    flexDirection: 'row',
    gap: 2,
    minWidth: 142,
    padding: 2
  },
  surfaceButton: {
    alignItems: 'center',
    borderRadius: radius.sm,
    flex: 1,
    minHeight: 44,
    justifyContent: 'center',
    paddingHorizontal: spacing.xs
  },
  surfaceButtonSelected: {
    backgroundColor: colors.elevatedSurface
  },
  surfaceText: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0
  },
  surfaceTextSelected: {
    color: colors.text
  },
  searchBar: {
    alignItems: 'center',
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 44,
    paddingHorizontal: spacing.sm
  },
  searchInput: {
    color: colors.text,
    flex: 1,
    fontSize: 15,
    minHeight: 44,
    paddingVertical: 0
  },
  iconButton: {
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 32,
    minWidth: 32
  },
  breadcrumbs: {
    alignItems: 'center',
    paddingTop: spacing.xs
  },
  breadcrumbItem: {
    alignItems: 'center',
    flexDirection: 'row'
  },
  breadcrumbButton: {
    justifyContent: 'center',
    minHeight: 44,
    maxWidth: 150,
    paddingHorizontal: spacing.xs
  },
  breadcrumbButtonPressed: {
    opacity: 0.62
  },
  breadcrumbText: {
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0
  },
  overviewText: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0,
    marginTop: 0
  },
  mapScroller: {
    flex: 1,
    marginTop: spacing.xs,
    overflow: 'hidden'
  },
  mapContent: {
    alignItems: 'stretch',
    flexDirection: 'row',
    height: '100%',
    paddingBottom: spacing.xl,
    paddingTop: spacing.sm
  },
  column: {
    backgroundColor: 'transparent',
    flexShrink: 0,
    height: '100%',
    overflow: 'hidden'
  },
  columnTitle: {
    color: colors.text,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0,
    paddingHorizontal: spacing.xs,
    paddingVertical: spacing.sm
  },
  columnList: {
    gap: spacing.xs,
    paddingTop: 2
  },
  columnListSurface: {
    flex: 1
  },
  mapRow: {
    alignItems: 'center',
    backgroundColor: 'transparent',
    borderColor: 'transparent',
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    minHeight: 74,
    overflow: 'hidden',
    position: 'relative',
    shadowColor: '#000000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.04,
    shadowRadius: 4
  },
  mapRowExpanded: {
    backgroundColor: colors.brandDustyBlueSoft,
    borderColor: colors.border
  },
  mapRowHighlighted: {
    borderColor: colors.focusRing,
    shadowColor: colors.focusRing,
    shadowOpacity: 0.22,
    shadowRadius: 8
  },
  rowSwipeUnderlay: {
    alignItems: 'center',
    backgroundColor: colors.action,
    bottom: 0,
    justifyContent: 'center',
    position: 'absolute',
    right: 0,
    top: 0,
    width: inventoryMapGestureConfig.branchSwipeRevealWidth
  },
  rowCard: {
    alignItems: 'center',
    alignSelf: 'stretch',
    backgroundColor: colors.elevatedSurface,
    flex: 1,
    flexDirection: 'row',
    minWidth: 0
  },
  rowMainGesture: {
    alignSelf: 'stretch',
    flex: 1,
    minWidth: 0
  },
  rowMain: {
    alignItems: 'center',
    flex: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 72,
    minWidth: 0,
    paddingLeft: spacing.sm,
    paddingVertical: spacing.xs
  },
  rowImageWrap: {
    position: 'relative'
  },
  rowImageFrame: {
    alignItems: 'center',
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    height: 52,
    justifyContent: 'center',
    overflow: 'hidden',
    width: 52
  },
  childCountBadge: {
    alignItems: 'center',
    backgroundColor: colors.accentStrong,
    borderColor: colors.elevatedSurface,
    borderRadius: 9,
    borderWidth: 1,
    bottom: -3,
    flexDirection: 'row',
    gap: 2,
    minHeight: 18,
    paddingHorizontal: 5,
    position: 'absolute',
    right: -4
  },
  childCountBadgeText: {
    color: colors.onAction,
    fontSize: 10,
    fontWeight: '900',
    letterSpacing: 0
  },
  rowImage: {
    height: '100%',
    width: '100%'
  },
  rowImageLabel: {
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0
  },
  rowText: {
    flex: 1,
    gap: 2,
    minWidth: 0
  },
  rowTitleLine: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.xs
  },
  rowTitle: {
    color: colors.text,
    flex: 1,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  rowMeta: {
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0
  },
  rowTrail: {
    color: colors.textMuted,
    fontSize: 12,
    lineHeight: 16
  },
  rowInfoButton: {
    alignItems: 'center',
    alignSelf: 'stretch',
    backgroundColor: colors.elevatedSurface,
    justifyContent: 'center',
    minWidth: 48
  },
  emptyColumn: {
    alignItems: 'center',
    gap: spacing.xs,
    justifyContent: 'center',
    minHeight: 130,
    padding: spacing.md
  },
  emptyColumnText: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '800',
    letterSpacing: 0,
    textAlign: 'center'
  },
  emptyColumnAction: {
    alignItems: 'center',
    backgroundColor: colors.elevatedSurface,
    borderColor: colors.border,
    borderRadius: 999,
    borderWidth: StyleSheet.hairlineWidth,
    flexDirection: 'row',
    gap: spacing.xs,
    marginTop: spacing.xs,
    minHeight: 40,
    paddingHorizontal: spacing.md
  },
  emptyColumnActionText: {
    color: colors.action,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0
  },
  centerState: {
    alignItems: 'center',
    flex: 1,
    justifyContent: 'center',
    padding: spacing.lg
  },
  centerText: {
    color: colors.textMuted,
    fontSize: 15,
    fontWeight: '700',
    lineHeight: 22,
    marginTop: spacing.sm,
    textAlign: 'center'
  },
  errorTitle: {
    color: colors.text,
    fontSize: 20,
    fontWeight: '900',
    letterSpacing: 0
  },
  sheet: {
    backgroundColor: colors.background,
    flex: 1,
    padding: spacing.lg
  },
  sheetHandle: {
    alignSelf: 'center',
    backgroundColor: colors.border,
    borderRadius: 2,
    height: 4,
    marginBottom: spacing.xs,
    width: 44
  },
  sheetTopBar: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'flex-end',
    minHeight: 44
  },
  sheetLoadingState: {
    alignItems: 'center',
    flex: 1,
    justifyContent: 'center',
    padding: spacing.lg
  },
  sheetCloseButton: {
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.xs
  },
  sheetCloseText: {
    color: colors.action,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  }
  });
}
