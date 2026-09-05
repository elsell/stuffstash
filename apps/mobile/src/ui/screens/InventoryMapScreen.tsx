import { InventoryMapInfoSheet } from './InventoryMapInfoSheet';
import { createStyles } from './InventoryMapScreen.styles';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import type { ProgressiveAssetDetailQueries } from '../serverState/useProgressiveAssetDetail';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { MutableRefObject } from 'react';
import { router } from 'expo-router';
import {
  ActivityIndicator,
  AccessibilityInfo,
  Animated,
  FlatList,
  Image,
  PanResponder,
  Pressable,
  RefreshControl,
  ScrollView,
  Text,
  useWindowDimensions,
  View
} from 'react-native';
import { ChevronRight, Info, Package, Plus, Search, X } from 'lucide-react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import type { AddAssetPhotosCommand } from '../../application/assets/AddAssetPhotosCommand';
import type { AssetCheckoutCommand } from '../../application/assets/AssetCheckoutCommand';
import type { AssetLifecycleCommand } from '../../application/assets/AssetLifecycleCommand';
import type { DeleteAssetPhotoCommand } from '../../application/assets/DeleteAssetPhotoCommand';
import type {
  InventoryMapAssetViewModel,
  InventoryMapQuery,
  InventoryMapViewModel
} from '../../application/assets/InventoryMapQuery';
import type { PhotoSelectionQuery } from '../../application/add/PhotoSelectionQuery';
import { spacing } from '../theme/tokens';
import type { MobileColorPalette } from '../theme/tokens';
import { useAppearancePalette } from '../theme/AppearanceContext';
import {
  buildInventoryMapEmptyColumnAction,
  buildInventoryMapBreadcrumbs,
  buildInventoryMapColumns,
  buildInventoryMapRowInteractionState,
  clampInventoryMapOffset,
  findInventoryMapSearchMatch,
  inventoryMapBranchSwipeOffset,
  inventoryMapGestureConfig,
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
import { BrowseSurfaceControl } from './BrowseSurfaceControl';
import type { InventoryMapColumnViewModel } from './InventoryMapPresentation';
import { addHereRouteParams } from './AddAssetInitialParent';
import { useAppFeedback } from '../feedback/AppFeedback';
import { AppTextInput, appKeyboardDismissMode } from '../components/AppTextInput';

type InventoryMapScreenProps = {
  readonly addAssetPhotosCommand: Pick<AddAssetPhotosCommand, 'execute'>;
  readonly assetCheckoutCommand: Pick<AssetCheckoutCommand, 'execute'>;
  readonly assetCoreQuery: ProgressiveAssetDetailQueries['assetCoreQuery'];
  readonly assetContentsQuery: ProgressiveAssetDetailQueries['assetContentsQuery'];
  readonly assetPhotosQuery: ProgressiveAssetDetailQueries['assetPhotosQuery'];
  readonly assetLifecycleCommand: Pick<AssetLifecycleCommand, 'execute'>;
  readonly canAdd: boolean;
  readonly deleteAssetPhotoCommand: Pick<DeleteAssetPhotoCommand, 'execute'>;
  readonly inventoryMapQuery: Pick<InventoryMapQuery, 'execute'>;
  readonly pathStore: MutableRefObject<Map<string, readonly string[]>>;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly selectedSurface: InventoryMapSurface;
  readonly onAdd: () => void;
  readonly onChangeSurface: (surface: InventoryMapSurface) => void;
};

type InventoryMapState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly map: InventoryMapViewModel }
  | { readonly status: 'error'; readonly message: string };

type BranchSwipeVisualState = {
  readonly assetId: string;
  readonly dragX: number;
};

type FinishBranchSwipeOptions = {
  readonly preserveVisual?: boolean;
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
  assetCoreQuery,
  assetContentsQuery,
  assetPhotosQuery,
  assetLifecycleCommand,
  canAdd,
  deleteAssetPhotoCommand,
  inventoryMapQuery,
  pathStore,
  photoSelectionQuery,
  selectedSurface,
  onAdd,
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
  const branchSwipeVisualClearTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const previousColumnsMapKey = useRef<string | undefined>(undefined);
  const feedback = useAppFeedback();
  const mapQuery = useMobileInventoryServerQuery({
    key: mobileQueryKeys.inventoryMap,
    query: (signal) => inventoryMapQuery.execute({ signal })
  });
  const state: InventoryMapState = mapQuery.data
    ? { status: 'ready', map: mapQuery.data }
    : mapQuery.isError
      ? { status: 'error', message: 'Inventory map could not load.' }
      : { status: 'loading' };
  const [openPath, setOpenPath] = useState<readonly string[]>([]);
  const [query, setQuery] = useState('');
  const [reduceMotionEnabled, setReduceMotionEnabled] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [pendingScrollLevel, setPendingScrollLevel] = useState<number | undefined>();
  const [highlightedAssetId, setHighlightedAssetId] = useState<string | undefined>();
  const [selection, setSelection] = useState<{ scope: string; asset: InventoryMapAssetViewModel }>();
  const selectedAsset = mapQuery.data && selection?.scope === mapStorageKey(mapQuery.data) ? selection.asset : undefined;
  function setSelectedAsset(asset: InventoryMapAssetViewModel | undefined): void {
    setSelection(asset && mapQuery.data ? { scope: mapStorageKey(mapQuery.data), asset } : undefined);
  }
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
    if (!mapQuery.data) return;
    const nextMap = mapQuery.data;
    setHighlightedAssetId(undefined);
    setSelection((current) => current?.scope === mapStorageKey(nextMap)
      ? { ...current, asset: nextMap.assets.find((next) => next.id === current.asset.id) ?? current.asset }
      : undefined);
    setBranchSwipeVisual(undefined);
    setMapVerticalScrollLocked(false);
    setOpenPath(pathStore.current.get(mapStorageKey(nextMap)) ?? []);
  }, [mapQuery.data, pathStore]);

  useEffect(() => {
    if (mapQuery.isError && mapQuery.data) {
      feedback.showNotice({ tone: 'error', title: 'Could not refresh Map', message: 'Your last loaded map is still available. Pull to refresh.' });
    }
  }, [mapQuery.isError, mapQuery.error, feedback]);

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

  async function refreshMap(): Promise<void> {
    setIsRefreshing(true);
    try { await mapQuery.refetch({ cancelRefetch: false }); }
    finally { setIsRefreshing(false); }
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
          <InventoryMapHeaderActions
            canAdd={canAdd}
            palette={colors}
            selectedSurface={selectedSurface}
            onAdd={onAdd}
            onChangeSurface={onChangeSurface}
          />
        </View>
        <View style={styles.searchBar}>
          <Search color={colors.textMuted} size={19} strokeWidth={2.5} />
          <AppTextInput
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
          <Pressable accessibilityRole="button" accessibilityLabel="Retry map" onPress={() => void refreshMap()}>
            <Text style={styles.sheetCloseText}>Retry</Text>
          </Pressable>
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
        assetCoreQuery={assetCoreQuery}
        assetContentsQuery={assetContentsQuery}
        assetPhotosQuery={assetPhotosQuery}
        assetLifecycleCommand={assetLifecycleCommand}
        deleteAssetPhotoCommand={deleteAssetPhotoCommand}
        photoSelectionQuery={photoSelectionQuery}
        onClose={() => setSelectedAsset(undefined)}
        onMapChanged={() => { void mapQuery.reconcile().catch(() => undefined); }}
      />
    </View>
  );
}

export function InventoryMapHeaderActions({
  canAdd,
  palette,
  selectedSurface,
  onAdd,
  onChangeSurface
}: {
  readonly canAdd: boolean;
  readonly palette: MobileColorPalette;
  readonly selectedSurface: InventoryMapSurface;
  readonly onAdd: () => void;
  readonly onChangeSurface: (surface: InventoryMapSurface) => void;
}) {
  const styles = createStyles(palette);
  return (
    <View style={styles.headerActions}>
      {canAdd ? (
        <Pressable
          accessibilityLabel="Add an asset"
          accessibilityRole="button"
          onPress={onAdd}
          style={styles.headerAddButton}
        >
          <Plus color={palette.action} size={24} strokeWidth={2.2} />
        </Pressable>
      ) : null}
      <BrowseSurfaceControl
        palette={palette}
        selectedSurface={selectedSurface}
        onChangeSurface={onChangeSurface}
      />
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
        keyboardDismissMode={appKeyboardDismissMode()}
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

function mapStorageKey(map: InventoryMapViewModel): string {
  return `${map.sessionScopeId}:${map.tenantId}:${map.inventoryId}`;
}
