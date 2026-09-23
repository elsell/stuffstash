import { SelectionRow } from '../components/SelectionRow';
import { useAddDestinationPresentation } from '../navigation/AddDestinationTask';
import { AddDestinationSelectionScreen } from './AddDestinationSelectionScreen';
import { useFocusedSheetActions } from '../components/useFocusedSheetActions';
import { AssetTagSelectionField } from '../components/AssetTagSelectionField';
import { useTaskPresentation } from '../navigation/useTaskPresentation';
import { useHeaderHeight } from '@react-navigation/elements';
import { AddDraftNameField } from './AddDraftNameField';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { usePreventRemove } from '@react-navigation/native';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';
import { AssetExpirationEditor } from '../components/AssetExpirationEditor';
import type { InventoryAssetTypesQuery } from '../../application/assets/InventoryAssetTypesQuery';
import type { AssetExpiration } from '../../domain/assets/AssetSummary';
import { useQuery } from '@tanstack/react-query';
import { useMobileServerStateScopeId } from '../navigation/MobileServerStateProvider';
import { isAccessFailure } from '../serverState/isAccessFailure';
import { useParentCandidates } from '../serverState/useParentCandidates';
import { useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { router, Stack } from 'expo-router';
import {
  AccessibilityInfo,
  ActivityIndicator,
  Alert,
  Image,
  Keyboard,
  PanResponder,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView, useSafeAreaInsets } from 'react-native-safe-area-context';
import { ChevronDown, ChevronUp, ImagePlus, X } from 'lucide-react-native';
import { CreateAssetCommand } from '../../application/add/CreateAssetCommand';
import {
  AddAssetDraft,
  AddAssetDraftContext,
  AddAssetDraftStore
} from '../../application/add/AddAssetDraftStore';
import type { AssetTagSummary } from '../../domain/assets/AssetSummary';
import {
  applyInlineAssetTagResolution,
  canApplyInlineAssetTagResolution,
  type CreateAssetTagDraft,
  reconcileCreatedAssetTags,
  resolveInlineAssetTag
} from '../../application/assets/AssetTagDraftResolution';
import { assetTagChipStylePresentation } from '../components/AssetTagChipsPresentation';
import { TagColorPicker } from '../components/TagColorPicker';
import { AppTextInput, appKeyboardDismissMode } from '../components/AppTextInput';
import { AddDraftScopeQuery } from '../../application/add/AddDraftScopeQuery';
import {
  ParentLookupQuery,
  ParentLookupResult
} from '../../application/add/ParentLookupQuery';
import {
  PhotoSelectionQuery,
  SelectedAssetPhoto
} from '../../application/add/PhotoSelectionQuery';
import { AddAssetContextQuery, type AddAssetContext } from '../../application/add/AddAssetContextQuery';
import { IdentityIcon, IdentityLabel } from '../components/IdentityIcon';
import { DraftPhotoPreviewModal } from './DraftPhotoPreviewModal';
import { useAppFeedback } from '../feedback/AppFeedback';
import { minimumTouchTargetSize, radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';
import {
  assertSelectableParent,
  ParentSelection,
  resolveParentAssetId,
  resolveSelectedParent
} from './AddAssetResolution';
import { applyInitialParentToDraft } from './AddAssetInitialParent';
import { assetDetailHref } from './AssetDetailNavigation';
import { showPhotoSourceChooser } from './PhotoSourceChooser';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';

type AddAssetScreenProps = {
  readonly inventoryAssetTypesQuery: Pick<InventoryAssetTypesQuery, 'execute'>;
  readonly addAssetDraftStore: AddAssetDraftStore;
  readonly addDraftScopeQuery: AddDraftScopeQuery;
  readonly createAssetCommand: Pick<CreateAssetCommand, 'execute'>;
  readonly addAssetContextQuery: AddAssetContextQuery;
  readonly initialParent?: ParentSelection;
  readonly onDismiss?: () => void;
  readonly parentLookupQuery: ParentLookupQuery;
  readonly photoSelectionQuery: PhotoSelectionQuery;
};

type LoadState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly context: AddAssetContext }
  | { readonly status: 'error'; readonly message: string };

type SaveState =
  | { readonly status: 'idle' }
  | { readonly status: 'saving' }
  | { readonly status: 'saved'; readonly message: string }
  | { readonly status: 'error'; readonly title: string; readonly message: string };

const emptyDraft: AddAssetDraft = {
  title: '',
  description: '',
  parentAssetId: undefined,
  parentQuery: '',
  selectedPhotos: [],
  selectedTagIds: [],
  newTags: [],
  showDetails: false,
  lastParent: undefined
};
export function AddAssetScreen(props: AddAssetScreenProps) {
  const scopeId = useMobileServerStateScopeId();
  const addContext = useMobileInventoryServerQuery({ key: mobileQueryKeys.addContext, query: signal => props.addAssetContextQuery.execute({ signal }) });
  const principal = useQuery({ queryKey: mobileQueryKeys.principal(scopeId), queryFn: ({ signal }) => props.addDraftScopeQuery.getPrincipal({ signal }) });
  return <ScopedAddAssetScreen key={JSON.stringify([addContext.resourceKey, principal.data?.id])} {...props} addContext={addContext} principalId={principal.data?.id} principalError={principal.error} onRetry={() => { if (addContext.isError) void addContext.refetch({ cancelRefetch: false }); if (principal.isError) void principal.refetch({ cancelRefetch: false }); }} />;
}

function ScopedAddAssetScreen({
  inventoryAssetTypesQuery, addAssetDraftStore, createAssetCommand, initialParent, onDismiss = () => router.back(), parentLookupQuery, photoSelectionQuery, addContext, principalId, principalError, onRetry
}: AddAssetScreenProps & { readonly addContext: ReturnType<typeof useMobileInventoryServerQuery<AddAssetContext>>; readonly principalId?: string; readonly principalError: Error | null; readonly onRetry: () => void }) {
  const types = useMobileInventoryServerQuery({ key: (scope, tenant, inventory) => mobileQueryKeys.customization(scope, tenant, inventory, 'inventory', 'asset-type-choices', 'active'), enabled: !!addContext.data, query: (signal) => inventoryAssetTypesQuery.execute(addContext.data!.tenantId, addContext.data!.inventoryId, { signal }) });
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  const feedback = useAppFeedback();
  const capturePhotoChooserVisit = useTaskPresentation(undefined, principalId);
  const restoredDraft = useRef(false);
  const safeAreaInsets = useSafeAreaInsets();
  const formScrollRef = useRef<ScrollView>(null);
  const navigationHeaderHeight = useHeaderHeight();
  const [loadState, setLoadState] = useState<LoadState>({ status: 'loading' });
  const [draftContext, setDraftContext] = useState<AddAssetDraftContext | undefined>();
  const [expiration, setExpiration] = useState<AssetExpiration | undefined>();
  const [customAssetTypeId, setCustomAssetTypeId] = useState<string | undefined>();
  const [expirationValid, setExpirationValid] = useState(true);
  const [expirationRevision, setExpirationRevision] = useState(0);
  const [nameRevision, setNameRevision] = useState(0);
  const [title, setTitle] = useState(emptyDraft.title);
  const [description, setDescription] = useState(emptyDraft.description);
  const [parentAssetId, setParentAssetId] = useState<string | undefined>(emptyDraft.parentAssetId);
  const [parentQuery, setParentQuery] = useState(emptyDraft.parentQuery);
  const [parentSearchQuery, setParentSearchQuery] = useState('');

  const [isCreatingParent, setIsCreatingParent] = useState(false);
  const [selectedPhotos, setSelectedPhotos] = useState<readonly SelectedAssetPhoto[]>(
    emptyDraft.selectedPhotos
  );
  const [selectedTagIds, setSelectedTagIds] = useState<readonly string[]>(emptyDraft.selectedTagIds ?? []);
  const [inlineTag, setInlineTag] = useState({ name: '', color: '' });
  const hasUnstagedTag = Boolean(inlineTag.name.trim() || inlineTag.color.trim());
  const [newTags, setNewTags] = useState<readonly CreateAssetTagDraft[]>(emptyDraft.newTags ?? []);
  const [showDetails, setShowDetails] = useState(emptyDraft.showDetails);
  const [lastParent, setLastParent] = useState<ParentSelection | undefined>(emptyDraft.lastParent);
  const [isParentMenuOpen, setIsParentMenuOpen] = useState(false);
  const [createdParent, setCreatedParent] = useState<ParentSelection | undefined>();
  const [draggingPhotoId, setDraggingPhotoId] = useState<string | undefined>();
  const [previewPhotoIndex, setPreviewPhotoIndex] = useState<number | undefined>();
  const [saveState, setSaveState] = useState<SaveState>({ status: 'idle' });
  const draftOperation = useRef<'save' | 'parent' | 'photo' | null>(null);
  const [draftBusy, setDraftBusy] = useState(false);
  usePreventRemove(draftBusy, () => {});
  function beginDraftOperation(operation: 'save' | 'parent' | 'photo') {
    if (draftOperation.current) return false;
    draftOperation.current = operation; setDraftBusy(true); Keyboard.dismiss();
    if (saveState.status === 'error') setSaveState({ status: 'idle' });
    return true;
  }
  function endDraftOperation() { draftOperation.current = null; setDraftBusy(false); }
  function showDraftError(title: string, message: string) {
    setSaveState({ status: 'error', title, message });
    if (Platform.OS === 'ios') AccessibilityInfo.announceForAccessibility(`${title}. ${message}`);
  }
  function editDraft(change: () => void) {
    if (draftOperation.current) return;
    if (saveState.status === 'error') setSaveState({ status: 'idle' });
    change();
  }

  const [keyboardBar, setKeyboardBar] = useState({ isVisible: false, keyboardHeight: 0 });

  const candidates = useParentCandidates(isParentMenuOpen ? parentSearchQuery : parentQuery, parentLookupQuery, isParentMenuOpen);
  const parentMatches = createdParent && createdParent.title === parentSearchQuery
    ? [createdParent, ...(candidates.data ?? []).filter((parent) => parent.id !== createdParent.id)]
    : candidates.data ?? [];
  const normalizedParentQuery = normalizeParentName(parentSearchQuery);
  const canCreateParent = candidates.data !== undefined && !candidates.isError && normalizedParentQuery.length > 0
    && ![createdParent, ...parentMatches].filter(isParentSelection).some(parent => normalizeParentName(parent.title) === normalizedParentQuery);

  const canChooseDestination = loadState.status === 'ready' && loadState.context.canAdd
    && Boolean(addContext.data?.canAdd) && !isAccessFailure(addContext.error) && !principalError
    && Boolean(principalId && principalId === draftContext?.principalId);
  const destinationOwner = useRef<object | undefined>(undefined);
  useLayoutEffect(() => {
    destinationOwner.current = canChooseDestination ? {} : undefined;
    if (!canChooseDestination) setIsParentMenuOpen(false);
    return () => { destinationOwner.current = undefined; };
  }, [canChooseDestination]);

  useEffect(() => {
    if (!addContext.data) {
      if (addContext.isError) {
        setLoadState({
          status: 'error',
          message: readableError(addContext.error, 'Could not load inventory context.')
        });
      }
      return;
    }
    if (isAccessFailure(addContext.error) || principalError) {
      setLoadState({ status: 'error', message: readableError(addContext.error ?? principalError, 'Could not load inventory context.') });
      return;
    }
    if (!principalId) return;
    const context = addContext.data;
    setLoadState({ status: 'ready', context });
    if (!restoredDraft.current) {
      restoredDraft.current = true;
      const nextContext = { tenantId: context.tenantId, inventoryId: context.inventoryId, principalId };
      setDraftContext(nextContext);
      applyDraft(applyInitialParentToDraft(addAssetDraftStore.load(nextContext) ?? emptyDraft, initialParent));
    }
  }, [addAssetDraftStore, addContext.data, addContext.error, addContext.isError, principalId, principalError, initialParent]);

  useEffect(() => {
    if (!draftContext) {
      return;
    }

    addAssetDraftStore.save(draftContext, {
      expiration,
      customAssetTypeId,
      title,
      description,
      parentAssetId,
      parentQuery,
      selectedPhotos,
      selectedTagIds,
      newTags,
      inlineTag,
      showDetails,
      lastParent
    });
  }, [
    addAssetDraftStore,
    expiration,
    customAssetTypeId,
    description,
    draftContext,
    lastParent,
    parentAssetId,
    parentQuery,
    selectedPhotos,
    selectedTagIds,
    newTags,
    inlineTag,
    showDetails,
    title
  ]);

  useEffect(() => {
    const showEvent = Platform.OS === 'ios' ? 'keyboardWillShow' : 'keyboardDidShow';
    const changeEvent = Platform.OS === 'ios' ? 'keyboardWillChangeFrame' : 'keyboardDidShow';
    const hideEvent = Platform.OS === 'ios' ? 'keyboardWillHide' : 'keyboardDidHide';

    const showSubscription = Keyboard.addListener(showEvent, (event) => {
      setKeyboardBar({
        isVisible: true,
        keyboardHeight: event.endCoordinates.height
      });
    });
    const changeSubscription = Keyboard.addListener(changeEvent, (event) => {
      setKeyboardBar({
        isVisible: true,
        keyboardHeight: event.endCoordinates.height
      });
    });
    const hideSubscription = Keyboard.addListener(hideEvent, () => {
      setKeyboardBar({ isVisible: false, keyboardHeight: 0 });
    });

    return () => {
      showSubscription.remove();
      changeSubscription.remove();
      hideSubscription.remove();
    };
  }, []);

  async function saveAsset(): Promise<void> {
    if (hasUnstagedTag || !expirationValid || !beginDraftOperation('save')) return;
    setSaveState({ status: 'saving' });

    try {
      const selectedParent = resolveSelectedParent(
        parentMatches,
        parentAssetId,
        parentQuery,
        lastParent
      );
      if (parentAssetId && !selectedParent) {
        throw new Error('Choose this parent again before saving.');
      }
      assertSelectableParent(selectedParent);
      const resolvedParentAssetId = resolveParentAssetId(
        parentMatches,
        parentQuery,
        parentAssetId
      );
      const result = await createAssetCommand.execute({
        expiration,
        customAssetTypeId,
        title,
        description,
        parentAssetId: resolvedParentAssetId,
        tagIds: selectedTagIds,
        newTags,
        activeTags: loadState.status === 'ready' ? loadState.context.assetTags : [],
        photos: selectedPhotos.map((photo) => ({
          fileName: photo.fileName,
          contentType: photo.contentType,
          contentBase64: photo.contentBase64,
          uri: photo.uri,
          sizeBytes: photo.sizeBytes
        }))
      });
      const nextParent = resolveSelectedParent(
        parentMatches,
        resolvedParentAssetId,
        parentQuery,
        lastParent
      );
      setExpiration(undefined);
      setCustomAssetTypeId(undefined);
      setExpirationValid(true);
      setExpirationRevision(value => value + 1);
      setNameRevision(value => value + 1);
      setTitle('');
      setDescription('');
      setParentAssetId(nextParent?.id);
      setParentQuery(nextParent?.title ?? '');
      setSelectedPhotos([]);
      setSelectedTagIds([]);
      setNewTags([]);
      setInlineTag({ name: '', color: '' });
      setShowDetails(false);
      setLastParent(nextParent);
      if (draftContext) {
        addAssetDraftStore.save(draftContext, {
          ...emptyDraft,
          parentAssetId: nextParent?.id,
          parentQuery: nextParent?.title ?? '',
          newTags: [],
          lastParent: nextParent
        });
      }
      setSaveState({ status: 'saved', message: result.message });
      feedback.showNotice({
        tone: 'success',
        title: 'Asset saved',
        message: result.message,
        action: {
          label: 'View',
          onPress: () => router.push(assetDetailHref(result.id))
        }
      });
    } catch (error) {
      const message = readableError(error, 'Could not save asset.');
      showDraftError('Could not save asset', message);
      await refreshDashboardAfterTagCreation(newTags);
    } finally { endDraftOperation(); }
  }

  async function refreshDashboardAfterTagCreation(stagedTags: readonly CreateAssetTagDraft[]): Promise<void> {
    try {
      const refreshed = await addContext.refetch({ throwOnError: true });
      const context = refreshed.data;
      if (!context) {
        throw new Error('Could not refresh inventory context.');
      }
      setLoadState({ status: 'ready', context });
      const reconciled = reconcileCreatedAssetTags(stagedTags, context.assetTags);
      if (reconciled.createdTagIds.length > 0) {
        setSelectedTagIds((current) => uniqueStrings([...current, ...reconciled.createdTagIds]));
        setNewTags(reconciled.remainingTags);
      }
    } catch {
      // The original save error is the user-facing failure.
    }
  }

  async function createParent(): Promise<void> {
    const parentName = parentSearchQuery.trim();
    if (!destinationOwner.current || !isParentMenuOpen || !canCreateParent) {
      return;
    }

    const owner = destinationOwner.current;
    if (!beginDraftOperation('parent')) return;
    setIsCreatingParent(true);
    setSaveState({ status: 'idle' });
    try {
      const result = await createAssetCommand.execute({
        kind: 'location',
        title: parentName,
        description: ''
      });
      if (destinationOwner.current !== owner) return;
      const createdParent = {
        id: result.id,
        title: result.title,
        kind: 'location' as const,
        subtitle: 'New location',
        pathLabel: result.title,
        selectionHint: 'Location',
        willPromoteToContainer: false
      };
      setParentAssetId(result.id);
      setParentQuery(result.title);
      setParentSearchQuery(result.title);
      setLastParent(createdParent);
      setCreatedParent(createdParent);
      setIsParentMenuOpen(false);

    } catch (error) {
      if (destinationOwner.current !== owner) return;
      const message = readableError(error, 'Could not create parent.');
      showDraftError('Could not create parent', message);
    } finally {
      setIsCreatingParent(false); endDraftOperation();
    }
  }

  async function addPhotosFromLibrary(): Promise<void> {
    if (!beginDraftOperation('photo')) return;
    try {
      const photos = await photoSelectionQuery.selectFromLibrary(selectedPhotos.length);
      if (photos.length === 0) {
        return;
      }

      setPreviewPhotoIndex(undefined);
      setSelectedPhotos((current) => [...current, ...photos]);
      setSaveState({ status: 'idle' });
    } catch (error) {
      const message = readableError(error, 'Could not select photos.');
      showDraftError('Could not select photos', message);
    } finally { endDraftOperation(); }
  }

  async function takePhoto(): Promise<void> {
    if (!beginDraftOperation('photo')) return;
    try {
      const photos = await photoSelectionQuery.captureFromCamera(selectedPhotos.length);
      if (photos.length === 0) {
        return;
      }

      setPreviewPhotoIndex(undefined);
      setSelectedPhotos((current) => [...current, ...photos]);
      setSaveState({ status: 'idle' });
    } catch (error) {
      const message = readableError(error, 'Could not take photo.');
      showDraftError('Could not take photo', message);
    } finally { endDraftOperation(); }
  }

  function removePhoto(photoId: string): void {
    if (draftOperation.current) return;
    setSelectedPhotos((current) => current.filter((photo) => photo.id !== photoId));
    setPreviewPhotoIndex((current) => {
      if (current === undefined) {
        return undefined;
      }

      const nextPhotos = selectedPhotos.filter((photo) => photo.id !== photoId);
      if (nextPhotos.length === 0) {
        return undefined;
      }

      return Math.min(current, nextPhotos.length - 1);
    });
  }

  function choosePhotoSource(): void {
    if (draftOperation.current) return;
    const canPresent = capturePhotoChooserVisit();
    if (!canPresent()) return;
    showPhotoSourceChooser({
      isCurrent: canPresent,
      onCamera: () => void takePhoto(),
      onLibrary: () => void addPhotosFromLibrary()
    });
  }

  function movePhoto(photoId: string, direction: number): void {
    if (draftOperation.current) return;
    setSelectedPhotos((current) => {
      const index = current.findIndex((photo) => photo.id === photoId);
      const targetIndex = index + direction;
      if (index < 0 || targetIndex < 0 || targetIndex >= current.length) {
        return current;
      }

      const next = [...current];
      const [photo] = next.splice(index, 1);
      next.splice(targetIndex, 0, photo);
      return next;
    });
  }

  function clearDraft(): void {
    if (draftOperation.current) return;
    const clearedDraft = { ...emptyDraft };
    applyDraft(clearedDraft);
    if (draftContext) {
      addAssetDraftStore.save(draftContext, clearedDraft);
    }
    setSaveState({ status: 'idle' });
  }

  function applyDraft(draft: AddAssetDraft): void {
    setExpiration(draft.expiration);
    setCustomAssetTypeId(draft.customAssetTypeId);
    setExpirationValid(true);
    setExpirationRevision(value => value + 1);
    setNameRevision(value => value + 1);
    setTitle(draft.title);
    setDescription(draft.description);
    setParentAssetId(draft.parentAssetId);
    setParentQuery(draft.parentQuery);
    setSelectedPhotos(draft.selectedPhotos);
    setSelectedTagIds(draft.selectedTagIds ?? []);
    setNewTags(draft.newTags ?? []);
    setInlineTag(draft.inlineTag ?? { name: '', color: '' });
    setShowDetails(draft.showDetails);
    setLastParent(draft.lastParent);
  }

  const selectedParent = resolveSelectedParent(parentMatches, parentAssetId, parentQuery, lastParent);
  const destinationActions = useFocusedSheetActions({
    primaryLabel: 'Choose destination', secondaryLabel: 'Cancel', disabled: draftBusy || !canChooseDestination,
    onApply: () => editDraft(() => { setParentSearchQuery(''); setIsParentMenuOpen(true); }), onBack: () => {}
  });
  useAddDestinationPresentation(isParentMenuOpen && canChooseDestination ? {
    blocked: draftBusy,
    onClose: () => setIsParentMenuOpen(false),
    content: <AddDestinationSelectionScreen query={parentSearchQuery} selected={selectedParent}
      unresolvedSelection={parentAssetId || parentQuery.trim() ? parentQuery || 'Selected destination' : undefined}
      matches={parentMatches} disabled={draftBusy} loading={!candidates.data && !candidates.isError}
      failed={candidates.isError} creating={isCreatingParent} canCreate={canCreateParent}
      error={saveState.status === 'error' ? saveState.message : undefined}
      onQuery={value => { if (destinationOwner.current) editDraft(() => { setParentSearchQuery(value); setCreatedParent(undefined); }); }}
      onRetry={() => { if (destinationOwner.current && !draftOperation.current) void candidates.refetch(); }}
      onCreate={() => void createParent()}
      onClose={() => { if (!draftOperation.current) setIsParentMenuOpen(false); }}
      onSelect={parent => {
        if (!destinationOwner.current || draftOperation.current) return;
        setParentAssetId(parent?.id); setParentQuery(parent?.title ?? ''); setLastParent(parent);
        setCreatedParent(undefined); setIsParentMenuOpen(false);
      }} />
  } : undefined);

  const closeOptions = useNativeHeaderActionOptions([{ kind: 'close', label: 'Close Add', disabled: draftBusy, onPress: () => editDraft(() => onDismiss?.()) }], 'left');
  const saveOptions = useNativeHeaderActionOptions([{ kind: 'save', label: 'Save item',
    disabled: draftBusy || hasUnstagedTag || !title.trim() || !expirationValid || loadState.status !== 'ready' || !loadState.context.canAdd,
    onPress: () => void saveAsset() }]);
  const headerOptions = useMemo(() => ({ headerShown: true, headerBackVisible: false, gestureEnabled: !draftBusy, title: 'Add item',
    ...closeOptions, ...saveOptions }), [draftBusy, closeOptions, saveOptions]);

  return (
    <SafeAreaView style={styles.shell} edges={['left', 'right']}>
      <Stack.Screen options={headerOptions} />
      <ScrollView
        ref={formScrollRef}
        style={{ flex: 1 }}
        contentInsetAdjustmentBehavior="automatic"
        scrollToOverflowEnabled={Platform.OS === 'ios'}
        contentContainerStyle={[
          styles.content,
          { paddingBottom: safeAreaInsets.bottom + spacing.lg }
        ]}
        automaticallyAdjustKeyboardInsets
        keyboardDismissMode={appKeyboardDismissMode()}
        keyboardShouldPersistTaps="handled"
      >
        {saveState.status === 'saving' ? <ActivityIndicator accessibilityLabel="Saving item" color={colors.action} /> : null}
        {saveState.status === 'error' ? <View accessibilityLiveRegion="assertive" onLayout={() => formScrollRef.current?.scrollTo({ y: Platform.OS === 'ios' ? -navigationHeaderHeight : 0, animated: false })}>
          <Text accessibilityRole="header" style={styles.errorText}>{saveState.title}</Text>
          <Text style={styles.errorText}>{saveState.message}</Text>
        </View> : null}
        {loadState.status === 'loading' ? (
          <View style={styles.centerState}>
            <ActivityIndicator color={colors.accent} />
            <Text style={styles.stateText}>Loading inventory</Text>
          </View>
        ) : null}
        {loadState.status === 'error' ? (
          <View style={styles.centerState}>
            <Text style={styles.errorTitle}>Could not load</Text>
            <Text style={styles.stateText}>{loadState.message}</Text>
            <NativeCommandButton label="Retry Add context" onPress={onRetry} />
          </View>
        ) : null}
        {loadState.status === 'ready' ? (
          <View>
            <View style={styles.contextLine}>
              <IdentityLabel
                iconSize="xs"
                kind="inventory"
                label={loadState.context.inventoryName}
                textStyle={styles.contextText}
              />
              <IdentityLabel
                iconSize="xs"
                kind="tenant"
                label={loadState.context.tenantName}
                textStyle={styles.contextText}
              />
            </View>

            {!loadState.context.canAdd ? (
              <View style={styles.unavailablePanel}>
                <Text style={styles.unavailableTitle}>Add is unavailable</Text>
                <Text style={styles.unavailableText}>
                  This inventory does not allow you to create assets.
                </Text>
              </View>
            ) : (
              <View>
                <PhotoCapture disabled={draftBusy}
                  draggingPhotoId={draggingPhotoId}
                  onBeginPhotoDrag={id => editDraft(() => setDraggingPhotoId(id))}
                  onEndPhotoDrag={() => setDraggingPhotoId(undefined)}
                  onAddPhotos={choosePhotoSource}
                  onMovePhoto={movePhoto}
                  onOpenPhoto={(index) => editDraft(() => setPreviewPhotoIndex(index))}
                  onRemovePhoto={removePhoto}
                  photos={selectedPhotos}
                />

                <Text style={styles.fieldLabel}>Name</Text>
                <AddDraftNameField key={Platform.OS === 'ios' ? `name-${nameRevision}` : 'name'}
                  accessibilityLabel="Asset name"
                  editable={!draftBusy}
                  onChangeText={value => editDraft(() => setTitle(value))}
                  placeholder="Furnace filter, passport, camping bin"
                  placeholderTextColor={colors.textMuted}
                  style={styles.input}
                  value={title}
                />

                <View style={styles.parentPicker}>
                  <SelectionRow label="Put in" accessibilityLabel="Choose destination" value={selectedParent?.title ?? (parentQuery.trim() || 'Top level')}
                    disabled={destinationActions.disabled} onPress={destinationActions.onApply} />
                  {selectedParent ? <Text style={styles.parentMeta}>{selectedParent.pathLabel || selectedParent.subtitle}</Text> : null}
                  {selectedParent?.willPromoteToContainer ? <Text style={styles.parentPromotionText}>Stuff Stash will turn {selectedParent.title} into a container for this item.</Text> : null}
                </View>

                <Pressable
                  accessibilityRole="button"
                  disabled={draftBusy}
                  accessibilityState={{ expanded: showDetails, disabled: draftBusy }}
                  onPress={() => editDraft(() => setShowDetails((current) => !current))}
                  style={styles.moreDetailsButton}
                >
                  <Text style={styles.moreDetailsText}>More details</Text>
                  {showDetails ? (
                    <ChevronUp color={colors.textMuted} size={18} strokeWidth={2.2} />
                  ) : (
                    <ChevronDown color={colors.textMuted} size={18} strokeWidth={2.2} />
                  )}
                </Pressable>

                {types.isError ? <View><Text accessibilityRole="alert" style={{ color: colors.text }}>Asset types could not be loaded.</Text><NativeCommandButton label="Retry asset types" disabled={draftBusy} onPress={() => { if (!draftOperation.current) void types.refetch(); }} /></View> : null}
                {types.data || !types.isError ? <AssetExpirationEditor key={expirationRevision} asset={{ id: 'new-item', title, description }} types={types.data} disabled={draftBusy}
                  draft={{ title, description, expiration, customAssetTypeId, expirationValid }}
                  onChange={(draft) => { if (draftOperation.current) return; setCustomAssetTypeId(draft.customAssetTypeId); if (draft.expirationValid !== false) setExpiration(draft.expiration ?? undefined); setExpirationValid(draft.expirationValid !== false); }} /> : null}
                {hasUnstagedTag && !showDetails ? <Text style={styles.parentPromotionText}>Open More details to add or clear the unfinished tag before saving.</Text> : null}
                {showDetails ? (
                  <View>
                    <AppTextInput
                      accessibilityLabel="Asset description"
                      multiline
                      editable={!draftBusy}
                      onChangeText={value => editDraft(() => setDescription(value))}
                      placeholder="Description"
                      placeholderTextColor={colors.textMuted}
                      style={[styles.input, styles.textArea]}
                      value={description}
                    />
                    <AssetTagPicker key={nameRevision} disabled={draftBusy} scope={JSON.stringify([loadState.context.tenantId, loadState.context.inventoryId, nameRevision])}
                      tags={loadState.context.assetTags}
                      selectedTagIds={selectedTagIds}
                      newTags={newTags}
                      entry={inlineTag}
                      onChange={(ids, tags, entry) => editDraft(() => { setSelectedTagIds(ids); setNewTags(tags); setInlineTag(entry); })}
                    />
                    <NativeCommandButton label="Clear draft" role="destructive"
                      disabled={draftBusy} onPress={clearDraft} />
                  </View>
                ) : null}


              </View>
            )}
          </View>
        ) : null}
      </ScrollView>
      <KeyboardDismissBar
        keyboardHeight={keyboardBar.keyboardHeight}
        visible={keyboardBar.isVisible}
      />
      <DraftPhotoPreviewModal
        disabled={draftBusy}
        currentIndex={previewPhotoIndex}
        onClose={() => setPreviewPhotoIndex(undefined)}
        onRemovePhoto={removePhoto}
        onSetIndex={setPreviewPhotoIndex}
        photos={selectedPhotos}
      />
    </SafeAreaView>
  );
}

function PhotoCapture({
  disabled,
  draggingPhotoId,
  onAddPhotos,
  onBeginPhotoDrag,
  onEndPhotoDrag,
  onMovePhoto,
  onOpenPhoto,
  onRemovePhoto,
  photos
}: {
  readonly disabled: boolean;
  readonly draggingPhotoId: string | undefined;
  readonly onAddPhotos: () => void;
  readonly onBeginPhotoDrag: (photoId: string) => void;
  readonly onEndPhotoDrag: () => void;
  readonly onMovePhoto: (photoId: string, direction: number) => void;
  readonly onOpenPhoto: (index: number) => void;
  readonly onRemovePhoto: (photoId: string) => void;
  readonly photos: readonly SelectedAssetPhoto[];
}) {
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  return (
    <View style={styles.photoPanel}>
      <Text style={styles.photoSectionTitle}>Photos</Text>
      <ScrollView
        horizontal
        scrollEnabled={draggingPhotoId === undefined}
        showsHorizontalScrollIndicator={false}
        style={styles.photoStrip}
      >
        <Pressable
          accessibilityLabel="Add photos"
          accessibilityHint="Choose camera or photo library"
          accessibilityRole="button"
          disabled={disabled}
          onPress={onAddPhotos}
          style={styles.addPhotoTile}
        >
          <ImagePlus color={colors.action} size={28} strokeWidth={2.2} />
        </Pressable>
        {photos.map((photo, index) => (
          <PhotoPreviewItem disabled={disabled}
            draggingPhotoId={draggingPhotoId}
            index={index}
            key={photo.id}
            onBeginPhotoDrag={onBeginPhotoDrag}
            onEndPhotoDrag={onEndPhotoDrag}
            onMovePhoto={onMovePhoto}
            onOpenPhoto={onOpenPhoto}
            onRemovePhoto={onRemovePhoto}
            photo={photo}
            photoCount={photos.length}
          />
        ))}
      </ScrollView>
    </View>
  );
}

function PhotoPreviewItem({
  disabled,
  draggingPhotoId,
  index,
  onBeginPhotoDrag,
  onEndPhotoDrag,
  onMovePhoto,
  onOpenPhoto,
  onRemovePhoto,
  photo,
  photoCount
}: {
  readonly disabled: boolean;
  readonly draggingPhotoId: string | undefined;
  readonly index: number;
  readonly onBeginPhotoDrag: (photoId: string) => void;
  readonly onEndPhotoDrag: () => void;
  readonly onMovePhoto: (photoId: string, direction: number) => void;
  readonly onOpenPhoto: (index: number) => void;
  readonly onRemovePhoto: (photoId: string) => void;
  readonly photo: SelectedAssetPhoto;
  readonly photoCount: number;
}) {
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  const isDragging = draggingPhotoId === photo.id;
  const dragState = useRef({ isDragging: false, didMove: false });
  const suppressNextPress = useRef(false);
  const panResponder = useMemo(
    () =>
      PanResponder.create({
        onStartShouldSetPanResponder: () => false,
        onMoveShouldSetPanResponder: (_event, gestureState) =>
          !disabled && dragState.current.isDragging &&
          Math.abs(gestureState.dx) > 6 &&
          Math.abs(gestureState.dx) > Math.abs(gestureState.dy),
        onPanResponderMove: (_event, gestureState) => {
          if (Math.abs(gestureState.dx) > 8 || Math.abs(gestureState.dy) > 8) {
            dragState.current = { ...dragState.current, didMove: true };
          }
        },
        onPanResponderRelease: (_event, gestureState) => {
          if (dragState.current.isDragging) {
            const slots = Math.trunc(gestureState.dx / 88);
            const clampedSlots = Math.max(-index, Math.min(photoCount - index - 1, slots));
            if (clampedSlots !== 0) {
              onMovePhoto(photo.id, clampedSlots);
            }
            dragState.current = { isDragging: false, didMove: false };
            onEndPhotoDrag();
            return;
          }
          dragState.current = { isDragging: false, didMove: false };
        },
        onPanResponderTerminate: () => {
          dragState.current = { isDragging: false, didMove: false };
          onEndPhotoDrag();
        }
      }),
    [disabled, index, onBeginPhotoDrag, onEndPhotoDrag, onMovePhoto, onOpenPhoto, photo.id, photoCount]
  );

  return (
    <View style={styles.photoPreviewShell}>
      <Pressable
        {...panResponder.panHandlers}
        disabled={disabled}
        accessibilityState={{ disabled }}
        accessibilityActions={[
          { name: 'activate', label: 'Preview photo' },
          { name: 'decrement', label: 'Move earlier' },
          { name: 'increment', label: 'Move later' },
          { name: 'delete', label: 'Remove photo' }
        ]}
        accessibilityHint="Tap to preview. Hold and drag to reorder."
        accessibilityRole="adjustable"
        accessibilityValue={{ text: `${(index + 1).toString()} of ${photoCount.toString()}` }}
        delayLongPress={220}
        onAccessibilityAction={(event) => {
          if (disabled) return;
          if (event.nativeEvent.actionName === 'activate') {
            onOpenPhoto(index);
          }
          if (event.nativeEvent.actionName === 'decrement') {
            onMovePhoto(photo.id, -1);
          }
          if (event.nativeEvent.actionName === 'increment') {
            onMovePhoto(photo.id, 1);
          }
          if (event.nativeEvent.actionName === 'delete') {
            onRemovePhoto(photo.id);
          }
        }}
        onLongPress={() => {
          if (disabled) return;
          dragState.current = { isDragging: true, didMove: false };
          suppressNextPress.current = true;
          onBeginPhotoDrag(photo.id);
        }}
        onPress={() => {
          if (disabled) return;
          if (suppressNextPress.current) {
            return;
          }

          onOpenPhoto(index);
        }}
        onPressOut={() => {
          if (!dragState.current.isDragging) {
            suppressNextPress.current = false;
            return;
          }

          dragState.current = { isDragging: false, didMove: false };
          suppressNextPress.current = false;
          onEndPhotoDrag();
        }}
        style={[styles.photoPreview, isDragging ? styles.photoPreviewDragging : null]}
      >
        <Image
          accessibilityIgnoresInvertColors
          source={{ uri: photo.uri }}
          style={styles.photoPreviewImage}
        />
        <Text style={styles.photoOrdinal}>{(index + 1).toString()}</Text>
        <Text style={styles.photoDragHint}>{isDragging ? 'Drag' : 'Hold'}</Text>
      </Pressable>
      <NativeCommandButton label={`Remove photo ${index + 1}`}
        disabled={disabled} onPress={() => onRemovePhoto(photo.id)} />
    </View>
  );
}


function AssetTagPicker({
  scope,
  disabled,
  newTags,
  entry,
  tags,
  selectedTagIds,
  onChange
}: {
  readonly scope: string;
  readonly disabled: boolean;
  readonly newTags: readonly CreateAssetTagDraft[];
  readonly entry: NonNullable<AddAssetDraft['inlineTag']>;
  readonly tags: readonly AssetTagSummary[];
  readonly selectedTagIds: readonly string[];
  readonly onChange: (tagIds: readonly string[], tags: readonly CreateAssetTagDraft[], entry: NonNullable<AddAssetDraft['inlineTag']>) => void;
}) {
  const [tagNameRevision, setTagNameRevision] = useState(0);
  const [creatingTag, setCreatingTag] = useState(false);
  const creationVisible = creatingTag || Boolean(entry.name.trim() || entry.color.trim());
  const creationActions = useFocusedSheetActions({
    primaryLabel: 'New tag', secondaryLabel: 'Cancel new tag',
    disabled: disabled || creationVisible, secondaryDisabled: disabled || !creationVisible,
    onApply: () => setCreatingTag(true),
    onBack: () => {
      setCreatingTag(false);
      setTagNameRevision(current => current + 1);
      onChange(selectedTagIds, newTags, { name: '', color: '' });
    }
  });
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  const { name: newTagName, color: newTagColor } = entry;
  function setNewTagName(name: string): void { if (!disabled) onChange(selectedTagIds, newTags, { ...entry, name }); }
  function setNewTagColor(color: string): void { if (!disabled) onChange(selectedTagIds, newTags, { ...entry, color }); }

  function addNewTag(): void {
    if (disabled) return;
    const displayName = newTagName.trim();
    if (displayName.length === 0) {
      return;
    }
    const transition = applyInlineAssetTagResolution({
      resolution,
      selectedTagIds,
      pendingTags: newTags
    });
    onChange(transition.selectedTagIds, transition.pendingTags, transition.shouldClearInputs ? { name: '', color: '' } : entry);
    if (transition.shouldClearInputs) setTagNameRevision(value => value + 1);
  }

  const resolution = resolveInlineAssetTag({
    displayName: newTagName,
    color: newTagColor,
    activeTags: tags,
    pendingTags: newTags
  });
  const canAddNewTag = canApplyInlineAssetTagResolution(resolution);

  return (
    <View style={styles.tagPicker}>
      <Text style={styles.tagPickerTitle}>Tags</Text>
      <AssetTagSelectionField scope={scope} disabled={disabled} tags={tags.map(tag => ({ id: tag.id, label: tag.displayName }))} selectedIds={selectedTagIds}
        onChange={ids => onChange(ids, newTags, entry)} />
      <View style={styles.tagOptions}>
        {newTags.map((tag, index) => {
          const colorStyle = assetTagChipStylePresentation(tag);
          return (
            <Pressable
              disabled={disabled}
              accessibilityRole="button"
              accessibilityLabel={`Remove new tag ${tag.displayName}`}
              accessibilityState={{ disabled }}
              key={`${tag.displayName}-${index.toString()}`}
              onPress={() => onChange(selectedTagIds, newTags.filter((_, currentIndex) => currentIndex !== index), entry)}
              style={[
                styles.tagOption,
                colorStyle.colored ? { backgroundColor: colorStyle.backgroundColor, borderColor: colorStyle.borderColor } : null,
                styles.tagOptionSelected
              ]}
            >
              <Text style={[styles.tagOptionText, styles.tagOptionTextSelected]} numberOfLines={1}>
                {tag.displayName}
              </Text>
              <X color={colors.action} size={14} strokeWidth={2.4} />
            </Pressable>
          );
        })}
      </View>
      {creationVisible ? <>
      <View style={styles.newTagRow}>
        <AddDraftNameField key={Platform.OS === 'ios' ? tagNameRevision : 'tag-name'} editable={!disabled}
          accessibilityLabel="New tag name"
          onChangeText={setNewTagName}
          placeholder="New tag"
          placeholderTextColor={colors.textMuted}
          style={[styles.input, styles.newTagNameInput]}
          value={newTagName}
        />
      </View>
      {resolution.status === 'display_name_too_long' ? <Text accessibilityRole="alert" style={styles.parentPromotionText}>Use a shorter tag name.</Text> : null}
      <TagColorPicker disabled={disabled} palette={colors} value={newTagColor} onChange={setNewTagColor} />
      <NativeCommandButton label="Add tag" disabled={disabled || !canAddNewTag} onPress={addNewTag} />
      {newTagName.trim() || newTagColor.trim() ? <Text style={styles.parentPromotionText}>Add this tag or clear its name and color before saving.</Text> : null}
      <NativeCommandButton label="Cancel new tag" disabled={creationActions.secondaryDisabled} onPress={creationActions.onBack} />
      </> : <NativeCommandButton label="New tag" disabled={creationActions.disabled} onPress={creationActions.onApply} />}
    </View>
  );
}

function uniqueStrings(values: readonly string[]): readonly string[] {
  return Array.from(new Set(values));
}

function KeyboardDismissBar({
  keyboardHeight,
  visible
}: {
  readonly keyboardHeight: number;
  readonly visible: boolean;
}) {
  const styles = createStyles(useAppearanceAwarePalette());
  if (Platform.OS !== 'ios' || !visible) {
    return null;
  }

  return (
    <View style={[styles.keyboardDismissBar, { bottom: keyboardHeight }]}>
      <Pressable
        accessibilityLabel="Dismiss keyboard"
        accessibilityRole="button"
        hitSlop={8}
        onPress={Keyboard.dismiss}
        style={styles.keyboardDoneButton}
      >
        <Text style={styles.keyboardDoneText}>Done</Text>
      </Pressable>
    </View>
  );
}

function normalizeParentName(value: string): string {
  return value.trim().toLocaleLowerCase();
}

function isParentSelection(
  value: ParentSelection | ParentLookupResult | undefined
): value is ParentSelection | ParentLookupResult {
  return value !== undefined;
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  shell: {
    flex: 1,
    backgroundColor: colors.background
  },
  content: {
    padding: spacing.lg,
    paddingBottom: spacing.xl
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
  },
  unavailablePanel: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    padding: spacing.md
  },
  unavailableTitle: {
    color: colors.text,
    fontSize: 18,
    fontWeight: '900',
    letterSpacing: 0
  },
  unavailableText: {
    color: colors.textMuted,
    fontSize: 14,
    lineHeight: 20,
    marginTop: spacing.xs
  },
  contextLine: {
    alignItems: 'center',
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm,
    marginBottom: spacing.md,
    marginTop: spacing.xs
  },
  contextText: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '700',
    letterSpacing: 0
  },
  input: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    color: colors.text,
    fontSize: 16,
    marginBottom: spacing.sm,
    minHeight: 48,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  fieldLabel: {
    color: colors.text,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0,
    marginBottom: spacing.xs
  },
  textArea: {
    minHeight: 96,
    textAlignVertical: 'top'
  },
  keyboardDismissBar: {
    alignItems: 'flex-end',
    backgroundColor: colors.surface,
    borderTopColor: colors.border,
    borderTopWidth: 1,
    justifyContent: 'center',
    left: 0,
    minHeight: 44,
    paddingHorizontal: spacing.md,
    position: 'absolute',
    right: 0
  },
  keyboardDoneButton: {
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 36,
    minWidth: 56
  },
  keyboardDoneText: {
    color: colors.action,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  },
  sectionTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '900',
    letterSpacing: 0,
    marginBottom: spacing.sm,
    marginTop: spacing.sm
  },
  photoPanel: {
    marginBottom: spacing.md
  },
  photoSectionTitle: {
    color: colors.text,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0,
    marginBottom: spacing.xs
  },
  photoStrip: {
    marginTop: spacing.xs
  },
  addPhotoTile: {
    alignItems: 'center',
    aspectRatio: 1,
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderStyle: 'dashed',
    borderWidth: 1,
    justifyContent: 'center',
    marginRight: spacing.sm,
    width: 108
  },
  parentPicker: {
    marginTop: spacing.xs
  },
  parentMeta: {
    color: colors.textMuted,
    fontSize: 12,
    letterSpacing: 0,
    marginTop: 2
  },
  parentPromotionText: {
    color: colors.textMuted,
    fontSize: 13,
    lineHeight: 18,
    marginBottom: spacing.sm
  },
  photoPreviewShell: {
    marginRight: spacing.sm,
    width: 108
  },
  photoPreview: {
    aspectRatio: 1,
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    overflow: 'hidden',
    width: '100%'
  },
  photoPreviewDragging: {
    borderColor: colors.action,
    borderWidth: 2,
    transform: [{ scale: 0.98 }]
  },
  photoPreviewImage: {
    height: '100%',
    width: '100%'
  },
  photoOrdinal: {
    backgroundColor: colors.surface,
    borderRadius: 999,
    color: colors.text,
    fontSize: 12,
    fontWeight: '900',
    left: 6,
    overflow: 'hidden',
    paddingHorizontal: 7,
    paddingVertical: 3,
    position: 'absolute',
    top: 6
  },
  photoDragHint: {
    backgroundColor: colors.surface,
    borderRadius: radius.sm,
    bottom: 6,
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '900',
    left: 6,
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: 7,
    paddingVertical: 4,
    position: 'absolute'
  },
  moreDetailsButton: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.xs,
    justifyContent: 'space-between',
    minHeight: 44,
    paddingVertical: spacing.xs
  },
  moreDetailsText: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  tagPicker: {
    marginTop: spacing.sm
  },
  tagPickerTitle: {
    color: colors.text,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0,
    marginBottom: spacing.xs
  },
  tagOptions: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.xs,
    marginBottom: spacing.sm
  },
  tagOption: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 999,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.xs,
    minHeight: minimumTouchTargetSize,
    minWidth: minimumTouchTargetSize,
    maxWidth: '100%',
    paddingHorizontal: spacing.sm,
    paddingVertical: 6
  },
  tagOptionSelected: {
    borderColor: colors.action
  },
  tagOptionText: {
    color: colors.textMuted,
    flexShrink: 1,
    fontSize: 13,
    fontWeight: '800',
    letterSpacing: 0,
    maxWidth: 180
  },
  tagOptionTextSelected: {
    color: colors.text
  },
  newTagRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.xs,
    marginBottom: spacing.sm
  },
  newTagNameInput: {
    flex: 1,
    minHeight: minimumTouchTargetSize,
    minWidth: 0
  },
  newTagColorInput: {
    minHeight: minimumTouchTargetSize,
    width: 96
  },

  savedText: {
    color: colors.accentStrong,
    fontSize: 14,
    fontWeight: '800',
    letterSpacing: 0,
    marginBottom: spacing.sm,
    marginTop: spacing.sm
  },
  errorText: {
    color: colors.warning,
    fontSize: 14,
    lineHeight: 20,
    marginBottom: spacing.sm,
    marginTop: spacing.sm
  },
  saveButton: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    justifyContent: 'center',
    minHeight: 52,
    marginTop: spacing.sm
  },
  saveButtonText: {
    color: colors.onAction,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  }
  });
}
