import { AssetRegionRecovery } from '../components/AssetRegionRecovery';
import { usePullRefresh } from '../serverState/usePullRefresh';
import { useCallback, useEffect, useRef, useState } from 'react';
import { router, Stack, useFocusEffect } from 'expo-router';
import {
  ActivityIndicator,
  Alert,
  RefreshControl,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import {
  AddAssetPhotosCommand,
  AddAssetPhotosCommandResult
} from '../../application/assets/AddAssetPhotosCommand';
import { AssetCheckoutCommand } from '../../application/assets/AssetCheckoutCommand';
import { AssetLifecycleCommand } from '../../application/assets/AssetLifecycleCommand';
import { UndoAssetEditCommand } from '../../application/assets/UndoAssetEditCommand';
import { DeleteAssetPhotoCommand } from '../../application/assets/DeleteAssetPhotoCommand';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { AssetCoreQuery } from '../../application/assets/AssetCoreQuery';
import { AssetContentsQuery } from '../../application/assets/AssetContentsQuery';
import { AssetPhotosQuery } from '../../application/assets/AssetPhotosQuery';
import {
  PhotoSelectionQuery,
  SelectedAssetPhoto
} from '../../application/add/PhotoSelectionQuery';
import {
  AssetDetailView,
  assetDetailNavigationTitle,
  AssetPhotoUploadProgressViewModel
} from '../components/AssetDetailView';
import { AssetPhotoViewerSheet } from './AssetPhotoViewerSheet';
import {
  assetHeaderOverflowScreenOptions
} from './AssetHeaderOverflow';
import { AssetDetailRouteErrorState } from './AssetDetailRouteErrorState';
import {
  assetPhotoViewerModel,
  isAssetPhotoId
} from '../components/AssetPhotoWorkspacePresentation';
import { addHereParams } from './AddAssetInitialParent';
import {
  assetDetailHref,
  navigateAfterDeletedAsset
} from './AssetDetailNavigation';
import { navigateToAssetTagSearch } from './AssetTagSearchNavigation';
import {
  assetLifecycleConfirmation,
  assetDetailLoadErrorPresentation,
  assetLifecycleFailurePresentation,
  AssetLifecycleActionKind
} from './AssetLifecyclePresentation';
import {
  assetWorkspaceSuccessStatus,
  visibleAssetWorkspaceStatus,
  AssetWorkspaceStatus
} from './AssetWorkspaceStatusPresentation';
import { consumeAssetActionCompletion } from './AssetActionCompletion';
import { showPhotoSourceChooser } from './PhotoSourceChooser';
import {
  applyPhotoUploadProgress,
  photoUploadRows
} from './AssetPhotoUploadProgressPresentation';
import { useAppFeedback } from '../feedback/AppFeedback';
import { spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { useProgressiveAssetDetail } from '../serverState/useProgressiveAssetDetail';
import { mergeProgressiveAssetDetail } from './AssetDetailProgressivePresentation';

type AssetDetailRouteScreenProps = {
  readonly addAssetPhotosCommand: Pick<AddAssetPhotosCommand, 'execute'>;
  readonly assetCoreQuery: Pick<AssetCoreQuery, 'execute'>;
  readonly assetContentsQuery: Pick<AssetContentsQuery, 'execute'>;
  readonly assetPhotosQuery: Pick<AssetPhotosQuery, 'execute'>;
  readonly assetCheckoutCommand: Pick<AssetCheckoutCommand, 'execute'>;
  readonly assetLifecycleCommand: Pick<AssetLifecycleCommand, 'execute'>;
  readonly undoAssetEditCommand: Pick<UndoAssetEditCommand, 'execute'>;
  readonly deleteAssetPhotoCommand: Pick<DeleteAssetPhotoCommand, 'execute'>;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly assetId: string;
};

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly asset: AssetDetailViewModel }
  | { readonly status: 'error'; readonly title: string; readonly message: string; readonly canRetry: boolean };

type PendingAction = 'archive' | 'restore' | 'delete' | 'edit' | 'move' | 'photos' | 'checkout' | 'return';

type PhotoUploadRow = AssetPhotoUploadProgressViewModel;

export function AssetDetailRouteScreen({
  addAssetPhotosCommand,
  assetCheckoutCommand,
  assetContentsQuery,
  assetCoreQuery,
  assetPhotosQuery,
  assetLifecycleCommand,
  undoAssetEditCommand,
  assetId,
  deleteAssetPhotoCommand,
  photoSelectionQuery
}: AssetDetailRouteScreenProps) {
  const palette = useAppearanceAwarePalette();
  const styles = createStyles(palette);
  const feedback = useAppFeedback();
  const progressive = useProgressiveAssetDetail(assetId, { assetCoreQuery, assetContentsQuery, assetPhotosQuery });
  const { coreAsset, assetContents, assetPhotos } = progressive;
  const screenState: ScreenState = coreAsset.data
    ? {
        status: 'ready',
        asset: mergeProgressiveAssetDetail(coreAsset.data.view, assetContents.data, assetPhotos.data)
      }
    : coreAsset.isError
      ? { status: 'error', ...assetDetailLoadErrorPresentation(coreAsset.error) }
      : { status: 'loading' };
  const [pendingAction, setPendingAction] = useState<PendingAction | undefined>();
  const [failedPhotoDrafts, setFailedPhotoDrafts] = useState<readonly SelectedAssetPhoto[]>([]);
  const [photoUploads, setPhotoUploads] = useState<readonly PhotoUploadRow[]>([]);
  const [photoStatus, setPhotoStatus] = useState<AddAssetPhotosCommandResult | undefined>();
  const [workspaceStatus, setWorkspaceStatus] = useState<AssetWorkspaceStatus | undefined>();
  const [selectedPhotoId, setSelectedPhotoId] = useState<string | undefined>();
  const [isRemovingPhoto, setIsRemovingPhoto] = useState(false);
  const assetOperation = useRef({ assetId, active: true, pending: false });
  useEffect(() => {
    if (assetOperation.current.pending) setPendingAction(undefined);
    const scope = { assetId, active: true, pending: false };
    assetOperation.current = scope;
    setIsRemovingPhoto(false);
    return () => { scope.active = false; };
  }, [assetId]);


  useEffect(() => {
    setPhotoUploads([]);
    setFailedPhotoDrafts([]);
    setPhotoStatus(undefined);
    setWorkspaceStatus(undefined);
    setSelectedPhotoId(undefined);
  }, [assetId]);

  useFocusEffect(useCallback(() => {
    const asset = screenState.status === 'ready'
      ? screenState.asset
      : coreAsset.data?.view;
    if (!asset) return;
    const completion = consumeAssetActionCompletion(assetId);
    if (!completion) return;
    if (completion.action === 'edit') {
      feedback.showNotice({
        tone: 'success',
        title: `Saved "${asset.title}"`,
        message: 'The change is now in History.',
        ...(completion.undoableOperationId ? {
          action: {
            label: 'Undo',
            onPress: () => void undoSavedEdit({
              operationId: completion.undoableOperationId!,
              tenantId: asset.tenantId ?? '',
              inventoryId: asset.inventoryId ?? '',
              title: asset.title
            })
          }
        } : {})
      });
      return;
    }
    setWorkspaceStatus(assetWorkspaceSuccessStatus(completion.action, { message: completion.message }));
  }, [assetId, coreAsset.data, feedback, screenState, undoAssetEditCommand]));

  async function undoSavedEdit(input: { readonly tenantId: string; readonly inventoryId: string; readonly operationId: string; readonly title: string }): Promise<void> {
    try {
      await undoAssetEditCommand.execute(input);
      await coreAsset.reconcile();
      feedback.showNotice({ tone: 'success', title: 'Edit undone', message: 'The previous values were reapplied.' });
    } catch (error) {
      feedback.showNotice({ tone: 'error', title: 'Could not undo edit', message: readableError(error, 'Undo failed.') });
    }
  }

  function openHistory(asset: AssetDetailViewModel): void {
    if (!asset.tenantId || !asset.inventoryId) {
      feedback.showNotice({ tone: 'error', title: 'Could not open History', message: 'The item scope is unavailable. Refresh and try again.' });
      return;
    }
    router.push({
      pathname: '/assets/[assetId]/history',
      params: { assetId: asset.id, tenantId: asset.tenantId, inventoryId: asset.inventoryId, assetTitle: asset.title }
    });
  }

  const { refreshing: isRefreshing, refresh: refreshAsset } = usePullRefresh(async () => {
    setWorkspaceStatus(undefined);

    try {
      await reloadAsset();
    } catch (error) {
      feedback.showNotice({
        tone: 'error',
        title: 'Could not refresh asset',
        message: readableError(error, 'Could not refresh asset.')
      });
    }
  });

  async function reloadAsset(): Promise<void> {
    await progressive.refresh();
  }

  async function retryLoad(): Promise<void> {
    // Query state owns initial error and retry presentation.
    await coreAsset.refetch({ cancelRefetch: false });
  }

  function choosePhotos(currentPhotoCount: number): void {
    const scope = assetOperation.current;
    showPhotoSourceChooser({
      onCamera: () => {
        if (scope.active) void addPhotos('camera', currentPhotoCount);
      },
      onLibrary: () => {
        if (scope.active) void addPhotos('library', currentPhotoCount);
      }
    });
  }

  async function addPhotos(source: 'camera' | 'library', currentPhotoCount: number): Promise<void> {
    await uploadPhotos(
      () => source === 'camera'
        ? photoSelectionQuery.captureFromCamera(currentPhotoCount)
        : photoSelectionQuery.selectFromLibrary(currentPhotoCount),
      'Could not add photos'
    );
  }

  async function retryPhotos(): Promise<void> {
    if (failedPhotoDrafts.length === 0) return;
    await uploadPhotos(async () => failedPhotoDrafts, 'Could not retry photos');
  }

  async function uploadPhotos(
    selectPhotos: () => Promise<readonly SelectedAssetPhoto[]>,
    failureTitle: string
  ): Promise<void> {
    const scope = assetOperation.current;
    if (!scope.active || scope.assetId !== assetId || scope.pending || pendingAction !== undefined) return;
    scope.pending = true;
    setPendingAction('photos');
    try {
      const photos = await selectPhotos();
      if (!scope.active || photos.length === 0) return;
      setPhotoStatus(undefined);
      setPhotoUploads(photoUploadRows(photos));
      const result = await addAssetPhotosCommand.execute({
        assetId,
        photos,
        onPhotoProgress: event => {
          if (scope.active) setPhotoUploads(current => applyPhotoUploadProgress(current, event));
        }
      });
      if (!scope.active) return;
      setPhotoStatus(result);
      setFailedPhotoDrafts(result.failedPhotos as readonly SelectedAssetPhoto[]);
      await assetPhotos.reconcile();
      if (scope.active && result.failedCount === 0) setPhotoUploads([]);
    } catch (error) {
      if (!scope.active) return;
      feedback.showNotice({
        tone: 'error',
        title: failureTitle,
        message: readableError(error, 'Photo upload failed.')
      });
    } finally {
      scope.pending = false;
      if (scope.active) setPendingAction(undefined);
    }
  }

  async function removePhoto(photoId: string): Promise<void> {
    const scope = assetOperation.current;
    if (!scope.active || scope.assetId !== assetId || scope.pending || pendingAction !== undefined) return;
    scope.pending = true;
    setIsRemovingPhoto(true);
    setPendingAction('photos');
    try {
      const result = await deleteAssetPhotoCommand.execute({ assetId, photoId });
      if (!scope.active) return;
      setPhotoStatus({
        attachedCount: 0,
        failedCount: 0,
        failedPhotos: [],
        message: result.message,
        canRetry: false
      });
      setFailedPhotoDrafts([]);
      setSelectedPhotoId(current => current === photoId ? undefined : current);
      setPhotoUploads([]);
      await assetPhotos.reconcile();
    } catch (error) {
      if (!scope.active) return;
      feedback.showDialog({
        title: 'Could not remove photo',
        message: readableError(error, 'Photo removal failed.'),
        primaryAction: { label: 'OK' }
      });
    } finally {
      scope.pending = false;
      if (scope.active) {
        setIsRemovingPhoto(false);
        setPendingAction(undefined);
      }
    }
  }

  function selectAssetPhoto(asset: AssetDetailViewModel, photoId: string): void {
    if (!isAssetPhotoId(asset.photos, photoId)) {
      return;
    }

    setSelectedPhotoId(photoId);
  }

  function openChildAsset(childId: string): void {
    setSelectedPhotoId(undefined);
    router.push(assetDetailHref(childId));
  }

  function openPlacementAsset(parent: AssetDetailViewModel['parentLocationTrail'][number]): void {
    setSelectedPhotoId(undefined);
    router.push(assetDetailHref(parent.id));
  }

  function confirmLifecycleAction(action: AssetLifecycleActionKind, asset: AssetDetailViewModel): void {
    const confirmation = assetLifecycleConfirmation(action, asset);
    Alert.alert(confirmation.title, confirmation.message, [
      { text: 'Cancel', style: 'cancel' },
      {
        text: confirmation.confirmLabel,
        style: confirmation.isDestructive ? 'destructive' : 'default',
        onPress: () => void runLifecycleAction(action, asset)
      }
    ]);
  }

  async function runLifecycleAction(action: AssetLifecycleActionKind, asset: AssetDetailViewModel): Promise<void> {
    const scope = assetOperation.current;
    if (!scope.active || scope.assetId !== assetId || scope.pending) return;
    scope.pending = true;
    setPendingAction(action);
    setWorkspaceStatus(undefined);

    try {
      await assetLifecycleCommand.execute({ action, assetId });
      if (!scope.active) return;

      if (action === 'delete') {
        navigateAfterDeletedAsset(router);
        return;
      }

      setWorkspaceStatus(assetWorkspaceSuccessStatus(action, asset));
      try {
        await coreAsset.reconcile();
      } catch {
        if (!scope.active) return;
        feedback.showNotice({
          tone: 'error',
          title: `${action === 'archive' ? 'Archive' : 'Restore'} succeeded`,
          message: 'The latest asset state could not be refreshed yet. Pull to refresh.'
        });
      }
    } catch (error) {
      if (!scope.active) return;
      const failure = assetLifecycleFailurePresentation(
        action,
        asset,
        readableError(error, 'Lifecycle action failed.')
      );
      feedback.showNotice({
        tone: 'error',
        title: failure.title,
        message: failure.message
      });
    } finally {
      scope.pending = false;
      if (scope.active) setPendingAction(undefined);
    }
  }

  async function runCheckoutAction(action: 'checkout' | 'return', asset: AssetDetailViewModel): Promise<void> {
    const scope = assetOperation.current;
    if (!scope.active || scope.assetId !== assetId || scope.pending) return;
    scope.pending = true;
    setPendingAction(action);
    setWorkspaceStatus(undefined);

    try {
      await assetCheckoutCommand.execute({ action, assetId });
      if (!scope.active) return;
      setWorkspaceStatus(assetWorkspaceSuccessStatus(action, asset));
      try {
        await coreAsset.reconcile();
      } catch {
        if (!scope.active) return;
        feedback.showNotice({
          tone: 'error',
          title: action === 'checkout' ? 'Checkout succeeded' : 'Return succeeded',
          message: 'The latest availability could not be refreshed yet. Pull to refresh.'
        });
      }
    } catch (error) {
      if (!scope.active) return;
      feedback.showNotice({
        tone: 'error',
        title: action === 'checkout' ? 'Could not checkout asset' : 'Could not return asset',
        message: readableError(error, 'Checkout action failed.')
      });
    } finally {
      scope.pending = false;
      if (scope.active) setPendingAction(undefined);
    }
  }

  const presentedWorkspaceStatus = visibleAssetWorkspaceStatus(pendingAction, workspaceStatus);
  const headerOverflow = screenState.status === 'ready' ? {
    asset: screenState.asset,
    disabled: pendingAction !== undefined,
    onCheckoutHistory: () => router.push(`/assets/${screenState.asset.id}/checkouts`),
    onHistory: () => openHistory(screenState.asset),
    onLifecycleAction: (action: AssetLifecycleActionKind) => confirmLifecycleAction(action, screenState.asset)
  } : undefined;
  return (
    <SafeAreaView style={styles.shell} edges={['left', 'right']}>
      <Stack.Screen options={{
        title: screenState.status === 'ready' ? assetDetailNavigationTitle(screenState.asset) : 'Details',
        ...(headerOverflow ? assetHeaderOverflowScreenOptions(headerOverflow) : {})
      }} />
      {screenState.status === 'loading' ? <LoadingState /> : null}
      {screenState.status === 'error' ? (
        <AssetDetailRouteErrorState
          canRetry={screenState.canRetry}
          message={screenState.message}
          onRetry={() => void retryLoad()}
          title={screenState.title}
        />
      ) : null}
      {screenState.status === 'ready' ? (
        <>
          {(() => {
            const photoViewer = assetPhotoViewerModel(screenState.asset.photos, selectedPhotoId);
            return (
              <AssetPhotoViewerSheet
                canRemove={screenState.asset.canAddPhotos}
                isRemoving={isRemovingPhoto}
                model={photoViewer}
                onClose={() => setSelectedPhotoId(undefined)}
                onRemove={(photoId) => void removePhoto(photoId)}
                onSelectPhoto={setSelectedPhotoId}
                photos={screenState.asset.photos}
              />
            );
          })()}
          <AssetDetailView
            asset={screenState.asset}
            canRetryPhotos={photoStatus?.canRetry}
            isActionPending={pendingAction !== undefined}
            isContentsLoading={!assetContents.data && assetContents.isPending}
            contentsAvailable={Boolean(assetContents.data)}
            photosAvailable={Boolean(assetPhotos.data)}
            contentsRecovery={assetContents.isError ? <AssetRegionRecovery
              region="contents" isRetrying={assetContents.isFetching}
              onRetry={() => { void assetContents.refetch({ cancelRefetch: false }); }}
            /> : undefined}
            photosRecovery={assetPhotos.isError ? <AssetRegionRecovery
              region="photos" isRetrying={assetPhotos.isFetching}
              onRetry={() => { void assetPhotos.refetch({ cancelRefetch: false }); }}
            /> : undefined}
            isPhotosLoading={!assetPhotos.data && assetPhotos.isPending}
            onAddHere={screenState.asset.canAddContainedAssets ? () => router.push({
              pathname: '/add',
              params: addHereParams(screenState.asset)
            }) : undefined}
            onAddPhotos={() => choosePhotos(screenState.asset.photos.length)}
            onCheckout={() => void runCheckoutAction('checkout', screenState.asset)}
            onChildPress={openChildAsset}
            onEdit={() => router.push(`/assets/${screenState.asset.id}/edit`)}
            onMove={() => router.push(`/assets/${screenState.asset.id}/move`)}
            onMoveThingsHere={screenState.asset.canAddContainedAssets ? () => router.push(`/assets/${screenState.asset.id}/move-here`) : undefined}
            onPhotoPress={(photoId) => selectAssetPhoto(screenState.asset, photoId)}
            onParentLocationPress={openPlacementAsset}
            onReturn={() => void runCheckoutAction('return', screenState.asset)}
            onRetryPhotos={() => void retryPhotos()}
            onTagPress={(tag) => navigateToAssetTagSearch(router, tag)}
            photoUploads={photoUploads}
            photoStatusMessage={pendingAction === 'photos' ? 'Updating photos...' : photoStatus?.message}
            workspaceStatusKind={presentedWorkspaceStatus?.kind}
            workspaceStatusMessage={presentedWorkspaceStatus?.message}
            refreshControl={
              <RefreshControl
                refreshing={isRefreshing}
                tintColor={palette.action}
                onRefresh={refreshAsset}
              />
            }
          />
        </>
      ) : null}
    </SafeAreaView>
  );
}

function LoadingState() {
  const palette = useAppearanceAwarePalette();
  const styles = createStyles(palette);
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={palette.accent} />
      <Text style={styles.stateText}>Loading asset</Text>
    </View>
  );
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

function createStyles(palette: MobileColorPalette) {
  return StyleSheet.create({
  shell: {
    flex: 1,
    backgroundColor: palette.background
  },
  centerState: {
    alignItems: 'center',
    flex: 1,
    justifyContent: 'center',
    padding: spacing.lg
  },
  stateText: {
    color: palette.textMuted,
    fontSize: 16,
    marginTop: spacing.md,
    textAlign: 'center'
  },
  });
}
