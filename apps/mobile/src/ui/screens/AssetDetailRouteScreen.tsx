import { useEffect, useRef, useState } from 'react';
import { router, Stack } from 'expo-router';
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
import { AssetAuditHistoryQuery } from '../../application/assets/AssetAuditHistoryQuery';
import { AssetLifecycleCommand } from '../../application/assets/AssetLifecycleCommand';
import { DeleteAssetPhotoCommand } from '../../application/assets/DeleteAssetPhotoCommand';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { AssetDetailQuery } from '../../application/assets/AssetDetailQuery';
import { MoveAssetCommand } from '../../application/assets/MoveAssetCommand';
import { UpdateAssetCommand } from '../../application/assets/UpdateAssetCommand';
import { CreateAssetCommand } from '../../application/add/CreateAssetCommand';
import { ParentLookupQuery } from '../../application/add/ParentLookupQuery';
import {
  PhotoSelectionQuery,
  SelectedAssetPhoto
} from '../../application/add/PhotoSelectionQuery';
import {
  AssetDetailView,
  AssetPhotoUploadProgressViewModel
} from '../components/AssetDetailView';
import {
  AssetAuditHistorySheet,
  AssetAuditHistorySheetState
} from './AssetAuditHistorySheet';
import { isCurrentAuditHistoryRequest } from './AssetAuditHistoryPresentation';
import { AssetPhotoViewerSheet } from './AssetPhotoViewerSheet';
import {
  assetPhotoViewerModel,
  moveAssetPhotoOrder,
  orderedAssetPhotos,
  resetLocalAssetPhotoOrder
} from '../components/AssetPhotoWorkspacePresentation';
import {
  EditAssetSheet,
  MoveAssetSheet,
  MoveDraft,
  MoveIntoDraft,
  MoveThingsHereSheet
} from './AssetDetailSheets';
import {
  EditDraft,
  hasDirtyEditAssetDraft,
  normalizedEditDraft
} from './AssetDetailEditPresentation';
import { addHereParams } from './AddAssetInitialParent';
import { navigateAfterDeletedAsset } from './AssetDetailNavigation';
import {
  assetLifecycleActionRows,
  assetLifecycleConfirmation,
  assetLifecycleFailurePresentation,
  assetLifecycleOverflowMenu,
  AssetLifecycleActionKind
} from './AssetLifecyclePresentation';
import {
  createdMoveDestinationParent,
  isSelectableMoveDestination,
  moveDestinationCreateInput,
  parentFromCurrentAssetPath
} from './AssetDetailMovePresentation';
import { showPhotoSourceChooser } from './PhotoSourceChooser';
import { colors, radius, spacing } from '../theme/tokens';

type AssetDetailRouteScreenProps = {
  readonly addAssetPhotosCommand: AddAssetPhotosCommand;
  readonly assetAuditHistoryQuery: AssetAuditHistoryQuery;
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetLifecycleCommand: AssetLifecycleCommand;
  readonly createAssetCommand: CreateAssetCommand;
  readonly deleteAssetPhotoCommand: DeleteAssetPhotoCommand;
  readonly moveAssetCommand: MoveAssetCommand;
  readonly parentLookupQuery: ParentLookupQuery;
  readonly photoSelectionQuery: PhotoSelectionQuery;
  readonly updateAssetCommand: UpdateAssetCommand;
  readonly assetId: string;
};

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly asset: AssetDetailViewModel }
  | { readonly status: 'error'; readonly message: string };

type PendingAction = 'archive' | 'restore' | 'delete' | 'edit' | 'move' | 'photos';

type PhotoUploadRow = AssetPhotoUploadProgressViewModel;

export function AssetDetailRouteScreen({
  addAssetPhotosCommand,
  assetAuditHistoryQuery,
  assetDetailQuery,
  assetLifecycleCommand,
  assetId,
  createAssetCommand,
  deleteAssetPhotoCommand,
  moveAssetCommand,
  parentLookupQuery,
  photoSelectionQuery,
  updateAssetCommand
}: AssetDetailRouteScreenProps) {
  const [screenState, setScreenState] = useState<ScreenState>({ status: 'loading' });
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [pendingAction, setPendingAction] = useState<PendingAction | undefined>();
  const [editDraft, setEditDraft] = useState<EditDraft | undefined>();
  const [moveDraft, setMoveDraft] = useState<MoveDraft | undefined>();
  const [moveIntoDraft, setMoveIntoDraft] = useState<MoveIntoDraft | undefined>();
  const [failedPhotoDrafts, setFailedPhotoDrafts] = useState<readonly SelectedAssetPhoto[]>([]);
  const [photoUploads, setPhotoUploads] = useState<readonly PhotoUploadRow[]>([]);
  const [photoStatus, setPhotoStatus] = useState<AddAssetPhotosCommandResult | undefined>();
  const [selectedPhotoId, setSelectedPhotoId] = useState<string | undefined>();
  const [photoOrder, setPhotoOrder] = useState<readonly string[]>([]);
  const [auditHistoryState, setAuditHistoryState] = useState<AssetAuditHistorySheetState>({
    status: 'closed'
  });
  const auditHistoryRequestRef = useRef(0);

  useEffect(() => {
    let isCurrent = true;
    auditHistoryRequestRef.current += 1;
    setAuditHistoryState({ status: 'closed' });
    setPhotoUploads([]);
    setPhotoStatus(undefined);
    setSelectedPhotoId(undefined);
    setPhotoOrder(resetLocalAssetPhotoOrder());

    assetDetailQuery
      .execute(assetId)
      .then((asset) => {
        if (isCurrent) {
          setScreenState({ status: 'ready', asset });
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
  }, [assetDetailQuery, assetId]);

  async function refreshAsset(): Promise<void> {
    setIsRefreshing(true);

    try {
      await reloadAsset();
    } catch (error) {
      setScreenState({
        status: 'error',
        message: readableError(error, 'Could not refresh asset.')
      });
    } finally {
      setIsRefreshing(false);
    }
  }

  async function reloadAsset(): Promise<AssetDetailViewModel> {
    const asset = await assetDetailQuery.execute(assetId);
    setScreenState({ status: 'ready', asset });
    setPhotoOrder(resetLocalAssetPhotoOrder());
    return asset;
  }

  function openEdit(asset: AssetDetailViewModel): void {
    setEditDraft({ title: asset.title, description: asset.description });
  }

  function requestCloseEdit(asset: AssetDetailViewModel): void {
    if (!hasDirtyEditAssetDraft(asset, editDraft)) {
      setEditDraft(undefined);
      return;
    }
    Alert.alert('Discard changes?', 'Your edits have not been saved.', [
      { text: 'Keep editing', style: 'cancel' },
      { text: 'Discard', style: 'destructive', onPress: () => setEditDraft(undefined) }
    ]);
  }

  async function saveEdit(): Promise<void> {
    if (!editDraft) {
      return;
    }
    setPendingAction('edit');
    try {
      const normalized = normalizedEditDraft(editDraft);
      await updateAssetCommand.execute({
        assetId,
        title: normalized.title,
        description: normalized.description
      });
      setEditDraft(undefined);
      await reloadAsset();
    } catch (error) {
      Alert.alert('Could not save changes', readableError(error, 'Asset update failed.'));
    } finally {
      setPendingAction(undefined);
    }
  }

  async function openMove(asset: AssetDetailViewModel): Promise<void> {
    const matches = await parentLookupQuery.execute('');
    const safeMatches = matches.filter((match) => match.id !== asset.id && isSelectableMoveDestination(match));
    const currentParent = asset.parentAssetId
      ? safeMatches.find((match) => match.id === asset.parentAssetId) ?? parentFromCurrentAssetPath(asset)
      : null;
    setMoveDraft({
      createKind: 'location',
      query: currentParent?.title ?? '',
      matches: safeMatches,
      selectedParent: currentParent
    });
  }

  async function updateMoveQuery(query: string, asset: AssetDetailViewModel): Promise<void> {
    setMoveDraft((current) => current ? { ...current, query } : current);
    const matches = await parentLookupQuery.execute(query);
    setMoveDraft((current) => current
      && current.query === query
      ? {
          ...current,
          matches: matches.filter((match) => match.id !== asset.id && isSelectableMoveDestination(match))
        }
      : current);
  }

  async function createMoveDestination(asset: AssetDetailViewModel): Promise<void> {
    const name = moveDraft?.query.trim() ?? '';
    const createKind = moveDraft?.createKind ?? 'location';
    if (name.length === 0) {
      return;
    }
    setPendingAction('move');
    try {
      const created = await createAssetCommand.execute(moveDestinationCreateInput(createKind, name));
      const createdParent = createdMoveDestinationParent({
        id: created.id,
        kind: createKind,
        title: created.title
      });
      setMoveDraft({
        createKind,
        query: created.title,
        matches: [createdParent, ...(moveDraft?.matches ?? []).filter((match) => match.id !== asset.id)],
        selectedParent: createdParent
      });
    } catch (error) {
      Alert.alert('Could not create destination', readableError(error, 'Destination creation failed.'));
    } finally {
      setPendingAction(undefined);
    }
  }

  async function saveMove(): Promise<void> {
    if (!moveDraft) {
      return;
    }
    setPendingAction('move');
    try {
      await moveAssetCommand.execute({
        assetId,
        parentAssetId: moveDraft.selectedParent?.id
      });
      setMoveDraft(undefined);
      await reloadAsset();
    } catch (error) {
      Alert.alert('Could not move asset', readableError(error, 'Move failed.'));
    } finally {
      setPendingAction(undefined);
    }
  }

  async function openMoveThingsHere(target: AssetDetailViewModel): Promise<void> {
    const matches = await parentLookupQuery.execute('');
    setMoveIntoDraft({
      target,
      query: '',
      matches: matches.filter((match) => match.id !== target.id),
      selectedAsset: undefined
    });
  }

  async function updateMoveIntoQuery(query: string, target: AssetDetailViewModel): Promise<void> {
    setMoveIntoDraft((current) => current ? { ...current, query } : current);
    const matches = await parentLookupQuery.execute(query);
    setMoveIntoDraft((current) => current
      && current.query === query
      ? {
          ...current,
          matches: matches.filter((match) => match.id !== target.id)
        }
      : current);
  }

  async function saveMoveInto(): Promise<void> {
    if (!moveIntoDraft?.selectedAsset) {
      return;
    }
    setPendingAction('move');
    try {
      await moveAssetCommand.execute({
        assetId: moveIntoDraft.selectedAsset.id,
        parentAssetId: moveIntoDraft.target.id
      });
      setMoveIntoDraft(undefined);
      await reloadAsset();
    } catch (error) {
      Alert.alert('Could not move asset here', readableError(error, 'Move failed.'));
    } finally {
      setPendingAction(undefined);
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
      Alert.alert('Could not add photos', readableError(error, 'Photo upload failed.'));
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
      Alert.alert('Could not retry photos', readableError(error, 'Photo retry failed.'));
    } finally {
      setPendingAction(undefined);
    }
  }

  function confirmRemovePhoto(photoId: string): void {
    Alert.alert('Remove photo?', 'This removes the photo from this asset.', [
      { text: 'Cancel', style: 'cancel' },
      { text: 'Remove', style: 'destructive', onPress: () => void removePhoto(photoId) }
    ]);
  }

  function movePhoto(photoId: string, direction: -1 | 1): void {
    if (screenState.status !== 'ready') {
      return;
    }
    setPhotoOrder((current) => moveAssetPhotoOrder({
      direction,
      photoId,
      photoOrder: current,
      photos: screenState.asset.photos
    }));
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
      Alert.alert('Could not remove photo', readableError(error, 'Photo removal failed.'));
    } finally {
      setPendingAction(undefined);
    }
  }

  function updatePhotoUploadProgress(event: AddAssetPhotoProgressEvent): void {
    setPhotoUploads((current) =>
      current.map((upload) =>
        upload.index === event.index
          ? { ...upload, status: event.status }
          : upload
      )
    );
  }

  function photoUploadRows(photos: readonly SelectedAssetPhoto[]): readonly PhotoUploadRow[] {
    return photos.map((photo, index) => ({
      index,
      fileName: photo.fileName,
      status: 'pending'
    }));
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
          if (index === overflow.auditIndex) {
            void openAuditHistory(asset);
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
      { text: 'Audit history', onPress: () => void openAuditHistory(asset) },
      { text: 'Cancel', style: 'cancel' }
    ]);
  }

  async function openAuditHistory(asset: AssetDetailViewModel): Promise<void> {
    const requestId = auditHistoryRequestRef.current + 1;
    auditHistoryRequestRef.current = requestId;
    setAuditHistoryState({ status: 'loading', assetTitle: asset.title });
    try {
      const history = await assetAuditHistoryQuery.execute({ assetId: asset.id, limit: 20 });
      if (isCurrentAuditHistoryRequest(auditHistoryRequestRef.current, requestId)) {
        setAuditHistoryState({ status: 'ready', assetTitle: asset.title, history });
      }
    } catch (error) {
      if (isCurrentAuditHistoryRequest(auditHistoryRequestRef.current, requestId)) {
        setAuditHistoryState({
          status: 'error',
          assetTitle: asset.title,
          message: readableError(error, 'Audit history failed.')
        });
      }
    }
  }

  function closeAuditHistory(): void {
    auditHistoryRequestRef.current += 1;
    setAuditHistoryState({ status: 'closed' });
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

    try {
      await assetLifecycleCommand.execute({ action, assetId });

      if (action === 'delete') {
        navigateAfterDeletedAsset(router);
        return;
      }

      await reloadAsset();
    } catch (error) {
      const failure = assetLifecycleFailurePresentation(
        action,
        asset,
        readableError(error, 'Lifecycle action failed.')
      );
      Alert.alert(failure.title, failure.message);
    } finally {
      setPendingAction(undefined);
    }
  }

  return (
    <SafeAreaView style={styles.shell} edges={['left', 'right']}>
      {screenState.status === 'loading' ? <LoadingState /> : null}
      {screenState.status === 'error' ? <ErrorState message={screenState.message} /> : null}
      {screenState.status === 'ready' ? (
        <>
          {(() => {
            const visiblePhotos = orderedAssetPhotos(screenState.asset.photos, photoOrder);
            const photoViewer = assetPhotoViewerModel(visiblePhotos, selectedPhotoId);
            return (
              <AssetPhotoViewerSheet
                canRemove={screenState.asset.canAddPhotos}
                model={photoViewer}
                onClose={() => setSelectedPhotoId(undefined)}
                onRemove={(photoId) => void removePhoto(photoId)}
                onSelectPhoto={setSelectedPhotoId}
                photos={visiblePhotos}
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
            onChildPress={(childId) => router.push(`/assets/${childId}`)}
            onEdit={() => openEdit(screenState.asset)}
            onMoreActions={() => showMoreActions(screenState.asset)}
            onMove={() => void openMove(screenState.asset)}
            onMoveThingsHere={screenState.asset.canAddContainedAssets ? () => void openMoveThingsHere(screenState.asset) : undefined}
            onMovePhoto={movePhoto}
            onPhotoPress={setSelectedPhotoId}
            onRemovePhoto={confirmRemovePhoto}
            onRetryPhotos={() => void retryPhotos()}
            photoOrder={photoOrder}
            photoUploads={photoUploads}
            photoStatusMessage={pendingAction === 'photos' ? 'Updating photos...' : photoStatus?.message}
            refreshControl={
              <RefreshControl
                refreshing={isRefreshing}
                tintColor={colors.action}
                onRefresh={refreshAsset}
              />
            }
          />
          <EditAssetSheet
            asset={screenState.asset}
            draft={editDraft}
            isSaving={pendingAction === 'edit'}
            onChange={setEditDraft}
            onClose={() => requestCloseEdit(screenState.asset)}
            onSave={() => void saveEdit()}
          />
          <MoveAssetSheet
            asset={screenState.asset}
            draft={moveDraft}
            isSaving={pendingAction === 'move'}
            onChangeCreateKind={(createKind) => setMoveDraft((current) => current ? { ...current, createKind } : current)}
            onChangeQuery={(query) => void updateMoveQuery(query, screenState.asset)}
            onClose={() => setMoveDraft(undefined)}
            onCreateDestination={() => void createMoveDestination(screenState.asset)}
            onSelectParent={(selectedParent) => setMoveDraft((current) => current ? { ...current, selectedParent } : current)}
            onSelectRoot={() => setMoveDraft((current) => current ? { ...current, selectedParent: null } : current)}
            onSave={() => void saveMove()}
          />
          <MoveThingsHereSheet
            draft={moveIntoDraft}
            isSaving={pendingAction === 'move'}
            onChangeQuery={(query) => void updateMoveIntoQuery(query, screenState.asset)}
            onClose={() => setMoveIntoDraft(undefined)}
            onSave={() => void saveMoveInto()}
            onSelectAsset={(selectedAsset) => setMoveIntoDraft((current) => current ? { ...current, selectedAsset } : current)}
          />
          <AssetAuditHistorySheet
            state={auditHistoryState}
            onClose={closeAuditHistory}
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
