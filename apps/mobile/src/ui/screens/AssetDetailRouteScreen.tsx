import { useCallback, useEffect, useState } from 'react';
import { router, Stack, useFocusEffect } from 'expo-router';
import {
  ActionSheetIOS,
  ActivityIndicator,
  Alert,
  Platform,
  Pressable,
  RefreshControl,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { Ellipsis } from 'lucide-react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import {
  AddAssetPhotosCommand,
  AddAssetPhotoProgressEvent,
  AddAssetPhotosCommandResult
} from '../../application/assets/AddAssetPhotosCommand';
import { AssetCheckoutCommand } from '../../application/assets/AssetCheckoutCommand';
import { AssetLifecycleCommand } from '../../application/assets/AssetLifecycleCommand';
import { UndoAssetEditCommand } from '../../application/assets/UndoAssetEditCommand';
import { DeleteAssetPhotoCommand } from '../../application/assets/DeleteAssetPhotoCommand';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { AssetDetailQuery } from '../../application/assets/AssetDetailQuery';
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
  assetLifecycleActionRows,
  assetLifecycleConfirmation,
  assetDetailLoadErrorPresentation,
  assetDetailOverflowControlState,
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
import { spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';

type AssetDetailRouteScreenProps = {
  readonly addAssetPhotosCommand: AddAssetPhotosCommand;
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetCheckoutCommand: AssetCheckoutCommand;
  readonly assetLifecycleCommand: AssetLifecycleCommand;
  readonly undoAssetEditCommand: UndoAssetEditCommand;
  readonly deleteAssetPhotoCommand: DeleteAssetPhotoCommand;
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
  assetDetailQuery,
  assetLifecycleCommand,
  undoAssetEditCommand,
  assetId,
  deleteAssetPhotoCommand,
  photoSelectionQuery
}: AssetDetailRouteScreenProps) {
  const palette = useAppearanceAwarePalette();
  const styles = createStyles(palette);
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
            } else {
              setWorkspaceStatus(assetWorkspaceSuccessStatus(completion.action, { message: completion.message }));
            }
          }
        }
      })
      .catch((error: unknown) => {
        if (isCurrent) {
          setScreenState({ status: 'error', ...assetDetailLoadErrorPresentation(error) });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [assetDetailQuery, assetId, feedback, undoAssetEditCommand]));

  async function undoSavedEdit(input: { readonly tenantId: string; readonly inventoryId: string; readonly operationId: string; readonly title: string }): Promise<void> {
    try {
      await undoAssetEditCommand.execute(input);
      await reloadAsset();
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

  async function retryLoad(): Promise<void> {
    setScreenState({ status: 'loading' });
    try {
      await reloadAsset();
    } catch (error) {
      setScreenState({ status: 'error', ...assetDetailLoadErrorPresentation(error) });
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

  function openPlacementAsset(parent: AssetDetailViewModel['parentLocationTrail'][number]): void {
    setSelectedPhotoId(undefined);
    router.push(assetDetailHref(parent.id));
  }

  function showMoreActions(asset: AssetDetailViewModel): void {
    if (assetDetailOverflowControlState(pendingAction !== undefined).disabled) {
      return;
    }
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
          if (index === overflow.checkoutHistoryIndex) {
            router.push(`/assets/${asset.id}/checkouts`);
            return;
          }
          if (index === overflow.auditIndex) {
            openHistory(asset);
            return;
          }
          const lifecycleActionIndex = overflow.lifecycleActionIndexes.indexOf(index);
          const action = lifecycleActionIndex >= 0 ? overflow.actionRows[lifecycleActionIndex] : undefined;
          if (action) {
            confirmLifecycleAction(action.kind, asset);
          }
        }
      );
      return;
    }
    Alert.alert(overflow.title, overflow.message, [
      { text: 'Checkout history', onPress: () => router.push(`/assets/${asset.id}/checkouts`) },
      { text: 'History', onPress: () => openHistory(asset) },
      ...overflow.actionRows.map((action) => ({
        text: action.label,
        style: action.isDestructive ? 'destructive' as const : 'default' as const,
        onPress: () => confirmLifecycleAction(action.kind, asset)
      })),
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
  const overflowControlState = assetDetailOverflowControlState(pendingAction !== undefined);

  return (
    <SafeAreaView style={styles.shell} edges={['left', 'right']}>
      <Stack.Screen options={{
        title: screenState.status === 'ready' ? assetDetailNavigationTitle(screenState.asset) : 'Details',
        ...(screenState.status === 'ready' ? {
          headerRight: () => (
            <Pressable
              accessibilityLabel={`More actions for ${screenState.asset.title}`}
              accessibilityRole="button"
              accessibilityState={overflowControlState.accessibilityState}
              disabled={overflowControlState.disabled}
              hitSlop={6}
              onPress={overflowControlState.disabled ? undefined : () => showMoreActions(screenState.asset)}
              style={styles.headerAction}
            >
              <Ellipsis accessible={false} color={palette.action} size={24} strokeWidth={2} />
            </Pressable>
          )
        } : {})
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
  headerAction: {
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 44,
    minWidth: 44
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
