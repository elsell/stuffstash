import { useEffect, useMemo, useState } from 'react';
import { router } from 'expo-router';
import { ActivityIndicator, Alert, Modal, Pressable, RefreshControl, Text, View } from 'react-native';
import type { AddAssetPhotoProgressEvent, AddAssetPhotosCommand, AddAssetPhotosCommandResult } from '../../application/assets/AddAssetPhotosCommand';
import type { AssetCheckoutCommand } from '../../application/assets/AssetCheckoutCommand';
import type { AssetLifecycleCommand } from '../../application/assets/AssetLifecycleCommand';
import type { DeleteAssetPhotoCommand } from '../../application/assets/DeleteAssetPhotoCommand';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { InventoryMapAssetViewModel } from '../../application/assets/InventoryMapQuery';
import type { PhotoSelectionQuery, SelectedAssetPhoto } from '../../application/add/PhotoSelectionQuery';
import { useProgressiveAssetDetail, type ProgressiveAssetDetailQueries } from '../serverState/useProgressiveAssetDetail';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { useAppFeedback } from '../feedback/AppFeedback';
import { AssetDetailView, type AssetPhotoUploadProgressViewModel } from '../components/AssetDetailView';
import { assetPhotoViewerModel, isAssetPhotoId } from '../components/AssetPhotoWorkspacePresentation';
import { AssetPhotoViewerSheet } from './AssetPhotoViewerSheet';
import { AssetOverflowMenu } from './AssetOverflowMenu';
import { assetLifecycleConfirmation, assetLifecycleFailurePresentation, type AssetLifecycleActionKind } from './AssetLifecyclePresentation';
import { applyPhotoUploadProgress, photoUploadRows } from './AssetPhotoUploadProgressPresentation';
import { assetWorkspaceSuccessStatus, visibleAssetWorkspaceStatus, type AssetWorkspaceStatus } from './AssetWorkspaceStatusPresentation';
import { showPhotoSourceChooser } from './PhotoSourceChooser';
import { addHereRouteParams } from './AddAssetInitialParent';
import { createStyles } from './InventoryMapScreen.styles';

type MapSheetDetailState =
  | { readonly status: 'idle' }
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly asset: AssetDetailViewModel }
  | { readonly status: 'error'; readonly message: string };

export function InventoryMapInfoSheet({
  addAssetPhotosCommand,
  asset,
  assetCheckoutCommand,
  assetCoreQuery,
  assetContentsQuery,
  assetPhotosQuery,
  assetLifecycleCommand,
  deleteAssetPhotoCommand,
  photoSelectionQuery,
  onClose,
  onMapChanged
}: {
  readonly addAssetPhotosCommand: Pick<AddAssetPhotosCommand, 'execute'>;
  readonly assetCheckoutCommand: Pick<AssetCheckoutCommand, 'execute'>;
  readonly asset?: InventoryMapAssetViewModel;
  readonly assetCoreQuery: ProgressiveAssetDetailQueries['assetCoreQuery'];
  readonly assetContentsQuery: ProgressiveAssetDetailQueries['assetContentsQuery'];
  readonly assetPhotosQuery: ProgressiveAssetDetailQueries['assetPhotosQuery'];
  readonly assetLifecycleCommand: Pick<AssetLifecycleCommand, 'execute'>;
  readonly deleteAssetPhotoCommand: Pick<DeleteAssetPhotoCommand, 'execute'>;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly onClose: () => void;
  readonly onMapChanged: () => void;
}) {
  const colors = useAppearancePalette();
  const styles = useMemo(() => createStyles(colors), [colors]);
  const feedback = useAppFeedback();
  const [selection, setSelection] = useState<{ rootId: string; assetId: string }>();
  const activeDetailAssetId = asset
    ? selection?.rootId === asset.id ? selection.assetId : asset.id
    : undefined;
  const progressive = useProgressiveAssetDetail(activeDetailAssetId, { assetCoreQuery, assetContentsQuery, assetPhotosQuery });
  const detailState: MapSheetDetailState = !asset ? { status: 'idle' }
    : progressive.asset ? { status: 'ready', asset: progressive.asset }
    : progressive.coreAsset.isError ? { status: 'error', message: 'Could not load asset details.' }
    : { status: 'loading' };
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [pendingAction, setPendingAction] = useState<PendingMapSheetAction | undefined>();
  const [failedPhotoDrafts, setFailedPhotoDrafts] = useState<readonly SelectedAssetPhoto[]>([]);
  const [photoUploads, setPhotoUploads] = useState<readonly AssetPhotoUploadProgressViewModel[]>([]);
  const [photoStatus, setPhotoStatus] = useState<AddAssetPhotosCommandResult | undefined>();
  const [workspaceStatus, setWorkspaceStatus] = useState<AssetWorkspaceStatus | undefined>();
  const [selectedPhotoId, setSelectedPhotoId] = useState<string | undefined>();

  useEffect(() => { setSelection(undefined); }, [asset?.id]);

  useEffect(() => {
    setIsRefreshing(false);
    setPendingAction(undefined);
    setFailedPhotoDrafts([]);
    setPhotoUploads([]);
    setPhotoStatus(undefined);
    setWorkspaceStatus(undefined);
    setSelectedPhotoId(undefined);
  }, [activeDetailAssetId]);

  useEffect(() => {
    if (progressive.asset && (progressive.assetContents.isError || progressive.assetPhotos.isError)) {
      feedback.showNotice({ tone: 'error', title: 'Some details could not load', message: 'The asset is available. Pull to refresh its photos and contents.' });
    }
  }, [progressive.assetContents.error, progressive.assetPhotos.error, feedback]);

  async function reconcileCore(): Promise<void> {
    try { await progressive.coreAsset.reconcile(); }
    catch { feedback.showNotice({ tone: 'error', title: 'Change saved', message: 'The latest asset state could not be refreshed. Pull to refresh.' }); }
  }

  async function reconcilePhotos(): Promise<void> {
    try { await progressive.assetPhotos.reconcile(); }
    catch { feedback.showNotice({ tone: 'error', title: 'Photos updated', message: 'The latest photos could not be refreshed. Pull to refresh.' }); }
  }

  async function refreshDetail(): Promise<void> {
    setIsRefreshing(true);
    setWorkspaceStatus(undefined);

    try {
      await progressive.refresh();
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
      await reconcilePhotos();
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
      await reconcilePhotos();
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
      await reconcilePhotos();
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
    if (asset) setSelection({ rootId: asset.id, assetId });
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

      setWorkspaceStatus(assetWorkspaceSuccessStatus(action, detail));
      onMapChanged();
      await reconcileCore();
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
      setWorkspaceStatus(assetWorkspaceSuccessStatus(action, detail));
      onMapChanged();
      await reconcileCore();
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
              <Pressable accessibilityRole="button" accessibilityLabel="Retry details" onPress={() => void progressive.coreAsset.refetch({ cancelRefetch: false })}>
                <Text style={styles.sheetCloseText}>Retry</Text>
              </Pressable>
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
                isPhotosLoading={!progressive.assetPhotos.data && progressive.assetPhotos.isPending}
                isContentsLoading={!progressive.assetContents.data && progressive.assetContents.isPending}
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
                overflowMenu={(
                  <AssetOverflowMenu
                    asset={detailState.asset}
                    disabled={pendingAction !== undefined}
                    onCheckoutHistory={() => {
                      onClose();
                      router.push(`/assets/${detailState.asset.id}/checkouts`);
                    }}
                    onHistory={() => {
                      onClose();
                      router.push({
                        pathname: '/assets/[assetId]/history',
                        params: {
                          assetId: detailState.asset.id,
                          tenantId: detailState.asset.tenantId ?? '',
                          inventoryId: detailState.asset.inventoryId ?? '',
                          assetTitle: detailState.asset.title
                        }
                      });
                    }}
                    onLifecycleAction={(action) => confirmLifecycleAction(action, detailState.asset)}
                  />
                )}
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

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}
