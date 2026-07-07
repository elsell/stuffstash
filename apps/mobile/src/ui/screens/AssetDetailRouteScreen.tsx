import { useCallback, useEffect, useState } from 'react';
import { router, Stack, useFocusEffect } from 'expo-router';
import {
  ActionSheetIOS,
  ActivityIndicator,
  Alert,
  Platform,
  RefreshControl,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import {
  AddAssetPhotosCommand,
  AddAssetPhotoProgressEvent,
  AddAssetPhotosCommandResult
} from '../../application/assets/AddAssetPhotosCommand';
import { AssetCheckoutCommand } from '../../application/assets/AssetCheckoutCommand';
import { AssetLifecycleCommand } from '../../application/assets/AssetLifecycleCommand';
import { DeleteAssetPhotoCommand } from '../../application/assets/DeleteAssetPhotoCommand';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { AssetDetailQuery } from '../../application/assets/AssetDetailQuery';
import {
  PhotoSelectionQuery,
  SelectedAssetPhoto
} from '../../application/add/PhotoSelectionQuery';
import {
  AssetDetailView,
  AssetPhotoUploadProgressViewModel
} from '../components/AssetDetailView';
import { AssetPhotoViewerSheet } from './AssetPhotoViewerSheet';
import {
  assetPhotoViewerModel,
  isAssetPhotoId
} from '../components/AssetPhotoWorkspacePresentation';
import { addHereParams } from './AddAssetInitialParent';
import {
  assetDetailHref,
  navigateAfterDeletedAsset
} from './AssetDetailNavigation';
import {
  assetLifecycleActionRows,
  assetLifecycleConfirmation,
  assetLifecycleFailurePresentation,
  assetLifecycleOverflowMenu,
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
import { colors, spacing } from '../theme/tokens';

type AssetDetailRouteScreenProps = {
  readonly addAssetPhotosCommand: AddAssetPhotosCommand;
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetCheckoutCommand: AssetCheckoutCommand;
  readonly assetLifecycleCommand: AssetLifecycleCommand;
  readonly deleteAssetPhotoCommand: DeleteAssetPhotoCommand;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly assetId: string;
};

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly asset: AssetDetailViewModel }
  | { readonly status: 'error'; readonly message: string };

type PendingAction = 'archive' | 'restore' | 'delete' | 'edit' | 'move' | 'photos' | 'checkout' | 'return';

type PhotoUploadRow = AssetPhotoUploadProgressViewModel;

export function AssetDetailRouteScreen({
  addAssetPhotosCommand,
  assetCheckoutCommand,
  assetDetailQuery,
  assetLifecycleCommand,
  assetId,
  deleteAssetPhotoCommand,
  photoSelectionQuery
}: AssetDetailRouteScreenProps) {
  const feedback = useAppFeedback();
  const [screenState, setScreenState] = useState<ScreenState>({ status: 'loading' });
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [pendingAction, setPendingAction] = useState<PendingAction | undefined>();
  const [failedPhotoDrafts, setFailedPhotoDrafts] = useState<readonly SelectedAssetPhoto[]>([]);
  const [photoUploads, setPhotoUploads] = useState<readonly PhotoUploadRow[]>([]);
  const [photoStatus, setPhotoStatus] = useState<AddAssetPhotosCommandResult | undefined>();
  const [workspaceStatus, setWorkspaceStatus] = useState<AssetWorkspaceStatus | undefined>();
  const [selectedPhotoId, setSelectedPhotoId] = useState<string | undefined>();

  useEffect(() => {
    setSelectedPhotoId(undefined);
  }, [assetId]);

  useFocusEffect(useCallback(() => {
    let isCurrent = true;
    setPhotoUploads([]);
    setPhotoStatus(undefined);
    setWorkspaceStatus(undefined);
    setSelectedPhotoId(undefined);

    assetDetailQuery
      .execute(assetId)
      .then((asset) => {
        if (isCurrent) {
          setScreenState({ status: 'ready', asset });
          const completion = consumeAssetActionCompletion(assetId);
          if (completion) {
            setWorkspaceStatus(assetWorkspaceSuccessStatus(completion.action, { message: completion.message }));
          }
        }
      })
      .catch((error: unknown) => {
        if (isCurrent) {
          setScreenState({
            status: 'error',
            message: readableError(error, 'Could not load asset.')
          });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [assetDetailQuery, assetId]));

  async function refreshAsset(): Promise<void> {
    setIsRefreshing(true);
    setWorkspaceStatus(undefined);

    try {
      await reloadAsset();
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

  async function reloadAsset(): Promise<AssetDetailViewModel> {
    const asset = await assetDetailQuery.execute(assetId);
    setScreenState({ status: 'ready', asset });
    return asset;
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
        assetId,
        photos,
        onPhotoProgress: updatePhotoUploadProgress
      });
      setPhotoStatus(result);
      setFailedPhotoDrafts(result.failedPhotos as readonly SelectedAssetPhoto[]);
      await reloadAsset();
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
    const photos = failedPhotoDrafts;
    if (photos.length === 0) {
      return;
    }
    setPendingAction('photos');
    try {
      setPhotoStatus(undefined);
      setPhotoUploads(photoUploadRows(photos));
      const result = await addAssetPhotosCommand.execute({
        assetId,
        photos,
        onPhotoProgress: updatePhotoUploadProgress
      });
      setPhotoStatus(result);
      setFailedPhotoDrafts(result.failedPhotos as readonly SelectedAssetPhoto[]);
      await reloadAsset();
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
    setPendingAction('photos');
    try {
      const result = await deleteAssetPhotoCommand.execute({ assetId, photoId });
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
      await reloadAsset();
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

  function showMoreActions(asset: AssetDetailViewModel): void {
    const overflow = assetLifecycleOverflowMenu(asset);
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
            confirmLifecycleAction(action.kind, asset);
            return;
          }
          if (index === overflow.checkoutHistoryIndex) {
            router.push(`/assets/${asset.id}/checkouts`);
            return;
          }
          if (index === overflow.auditIndex) {
            router.push(`/assets/${asset.id}/audit`);
          }
        }
      );
      return;
    }
    Alert.alert(overflow.title, overflow.message, [
      ...overflow.actionRows.map((action) => ({
        text: action.label,
        style: action.isDestructive ? 'destructive' as const : 'default' as const,
        onPress: () => confirmLifecycleAction(action.kind, asset)
      })),
      { text: 'Checkout history', onPress: () => router.push(`/assets/${asset.id}/checkouts`) },
      { text: 'Audit history', onPress: () => router.push(`/assets/${asset.id}/audit`) },
      { text: 'Cancel', style: 'cancel' }
    ]);
  }

  function lifecycleActions(asset: AssetDetailViewModel) {
    return assetLifecycleActionRows(asset);
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
    setPendingAction(action);
    setWorkspaceStatus(undefined);

    try {
      await assetLifecycleCommand.execute({ action, assetId });

      if (action === 'delete') {
        navigateAfterDeletedAsset(router);
        return;
      }

      await reloadAsset();
      setWorkspaceStatus(assetWorkspaceSuccessStatus(action, asset));
    } catch (error) {
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
      setPendingAction(undefined);
    }
  }

  async function runCheckoutAction(action: 'checkout' | 'return', asset: AssetDetailViewModel): Promise<void> {
    setPendingAction(action);
    setWorkspaceStatus(undefined);

    try {
      await assetCheckoutCommand.execute({ action, assetId });
      await reloadAsset();
      setWorkspaceStatus(assetWorkspaceSuccessStatus(action, asset));
    } catch (error) {
      feedback.showNotice({
        tone: 'error',
        title: action === 'checkout' ? 'Could not checkout asset' : 'Could not return asset',
        message: readableError(error, 'Checkout action failed.')
      });
    } finally {
      setPendingAction(undefined);
    }
  }

  const presentedWorkspaceStatus = visibleAssetWorkspaceStatus(pendingAction, workspaceStatus);

  return (
    <SafeAreaView style={styles.shell} edges={['left', 'right']}>
      {screenState.status === 'loading' ? <LoadingState /> : null}
      {screenState.status === 'error' ? <ErrorState message={screenState.message} /> : null}
      {screenState.status === 'ready' ? (
        <>
          {(() => {
            const photoViewer = assetPhotoViewerModel(screenState.asset.photos, selectedPhotoId);
            return (
              <AssetPhotoViewerSheet
                canRemove={screenState.asset.canAddPhotos}
                model={photoViewer}
                onClose={() => setSelectedPhotoId(undefined)}
                onRemove={(photoId) => void removePhoto(photoId)}
                onSelectPhoto={setSelectedPhotoId}
                photos={screenState.asset.photos}
              />
            );
          })()}
          <Stack.Screen options={{ title: screenState.asset.title }} />
          <AssetDetailView
            asset={screenState.asset}
            canRetryPhotos={photoStatus?.canRetry}
            isActionPending={pendingAction !== undefined}
            onAddHere={screenState.asset.canAddContainedAssets ? () => router.push({
              pathname: '/add',
              params: addHereParams(screenState.asset)
            }) : undefined}
            onAddPhotos={() => choosePhotos(screenState.asset.photos.length)}
            onCheckout={() => void runCheckoutAction('checkout', screenState.asset)}
            onChildPress={openChildAsset}
            onEdit={() => router.push(`/assets/${screenState.asset.id}/edit`)}
            onMoreActions={() => showMoreActions(screenState.asset)}
            onMove={() => router.push(`/assets/${screenState.asset.id}/move`)}
            onMoveThingsHere={screenState.asset.canAddContainedAssets ? () => router.push(`/assets/${screenState.asset.id}/move-here`) : undefined}
            onPhotoPress={(photoId) => selectAssetPhoto(screenState.asset, photoId)}
            onReturn={() => void runCheckoutAction('return', screenState.asset)}
            onRetryPhotos={() => void retryPhotos()}
            photoUploads={photoUploads}
            photoStatusMessage={pendingAction === 'photos' ? 'Updating photos...' : photoStatus?.message}
            workspaceStatusKind={presentedWorkspaceStatus?.kind}
            workspaceStatusMessage={presentedWorkspaceStatus?.message}
            refreshControl={
              <RefreshControl
                refreshing={isRefreshing}
                tintColor={colors.action}
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
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={colors.accent} />
      <Text style={styles.stateText}>Loading asset</Text>
    </View>
  );
}

function ErrorState({ message }: { readonly message: string }) {
  return (
    <View style={styles.centerState}>
      <Text style={styles.errorTitle}>Could not load</Text>
      <Text style={styles.stateText}>{message}</Text>
    </View>
  );
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

const styles = StyleSheet.create({
  shell: {
    flex: 1,
    backgroundColor: colors.background
  },
  centerState: {
    alignItems: 'center',
    flex: 1,
    justifyContent: 'center',
    padding: spacing.lg
  },
  stateText: {
    color: colors.textMuted,
    fontSize: 16,
    lineHeight: 23,
    marginTop: spacing.md,
    textAlign: 'center'
  },
  errorTitle: {
    color: colors.text,
    fontSize: 24,
    fontWeight: '800',
    letterSpacing: 0
  }
});
