import { useEffect, useMemo, useRef, useState } from 'react';
import { router } from 'expo-router';
import {
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
  TextInput,
  View
} from 'react-native';
import { SafeAreaView, useSafeAreaInsets } from 'react-native-safe-area-context';
import { Check, ChevronDown, ChevronUp, ImagePlus, X } from 'lucide-react-native';
import { CreateAssetCommand } from '../../application/add/CreateAssetCommand';
import {
  AddAssetDraft,
  AddAssetDraftContext,
  AddAssetDraftStore
} from '../../application/add/AddAssetDraftStore';
import type { AssetTagSummary } from '../../domain/assets/AssetSummary';
import {
  applyInlineAssetTagResolution,
  canResolveInlineAssetTag,
  type CreateAssetTagDraft,
  reconcileCreatedAssetTags,
  resolveInlineAssetTag
} from '../../application/assets/AssetTagDraftResolution';
import { assetTagChipStylePresentation } from '../components/AssetTagChipsPresentation';
import { TagColorPicker } from '../components/TagColorPicker';
import { AddDraftScopeQuery } from '../../application/add/AddDraftScopeQuery';
import {
  ParentLookupQuery,
  ParentLookupResult
} from '../../application/add/ParentLookupQuery';
import {
  PhotoSelectionQuery,
  SelectedAssetPhoto
} from '../../application/add/PhotoSelectionQuery';
import {
  HomeDashboardQuery,
  HomeDashboardViewModel
} from '../../application/home/HomeDashboardQuery';
import { IdentityIcon, IdentityLabel } from '../components/IdentityIcon';
import { FullScreenPhotoViewer, type FullScreenPhotoViewerPhoto } from '../components/FullScreenPhotoViewer';
import { photoMetadataLabel } from '../components/AssetPhotoWorkspacePresentation';
import { useAppFeedback } from '../feedback/AppFeedback';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
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

type AddAssetScreenProps = {
  readonly addAssetDraftStore: AddAssetDraftStore;
  readonly addDraftScopeQuery: AddDraftScopeQuery;
  readonly createAssetCommand: CreateAssetCommand;
  readonly dashboardQuery: HomeDashboardQuery;
  readonly initialParent?: ParentSelection;
  readonly parentLookupQuery: ParentLookupQuery;
  readonly photoSelectionQuery: PhotoSelectionQuery;
};

type LoadState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly dashboard: HomeDashboardViewModel }
  | { readonly status: 'error'; readonly message: string };

type SaveState =
  | { readonly status: 'idle' }
  | { readonly status: 'saving' }
  | { readonly status: 'saved'; readonly message: string }
  | { readonly status: 'error'; readonly message: string };

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
const addSheetBottomChromePadding = spacing.xl * 5;

export function AddAssetScreen({
  addAssetDraftStore,
  addDraftScopeQuery,
  createAssetCommand,
  dashboardQuery,
  initialParent,
  parentLookupQuery,
  photoSelectionQuery
}: AddAssetScreenProps) {
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  const feedback = useAppFeedback();
  const safeAreaInsets = useSafeAreaInsets();
  const bottomChromeAllowance = safeAreaInsets.bottom + addSheetBottomChromePadding;
  const [loadState, setLoadState] = useState<LoadState>({ status: 'loading' });
  const [draftContext, setDraftContext] = useState<AddAssetDraftContext | undefined>();
  const [title, setTitle] = useState(emptyDraft.title);
  const [description, setDescription] = useState(emptyDraft.description);
  const [parentAssetId, setParentAssetId] = useState<string | undefined>(emptyDraft.parentAssetId);
  const [parentQuery, setParentQuery] = useState(emptyDraft.parentQuery);
  const [parentMatches, setParentMatches] = useState<readonly ParentLookupResult[]>([]);
  const [isCreatingParent, setIsCreatingParent] = useState(false);
  const [selectedPhotos, setSelectedPhotos] = useState<readonly SelectedAssetPhoto[]>(
    emptyDraft.selectedPhotos
  );
  const [selectedTagIds, setSelectedTagIds] = useState<readonly string[]>(emptyDraft.selectedTagIds ?? []);
  const [newTags, setNewTags] = useState<readonly CreateAssetTagDraft[]>(emptyDraft.newTags ?? []);
  const [showDetails, setShowDetails] = useState(emptyDraft.showDetails);
  const [lastParent, setLastParent] = useState<ParentSelection | undefined>(emptyDraft.lastParent);
  const [isParentMenuOpen, setIsParentMenuOpen] = useState(false);
  const [createdParent, setCreatedParent] = useState<ParentSelection | undefined>();
  const [draggingPhotoId, setDraggingPhotoId] = useState<string | undefined>();
  const [previewPhotoIndex, setPreviewPhotoIndex] = useState<number | undefined>();
  const [saveState, setSaveState] = useState<SaveState>({ status: 'idle' });
  const [keyboardBar, setKeyboardBar] = useState({ isVisible: false, keyboardHeight: 0 });

  useEffect(() => {
    let isCurrent = true;

    Promise.all([dashboardQuery.execute(), addDraftScopeQuery.execute()])
      .then(([dashboard, scope]) => {
        if (isCurrent) {
          const nextContext = {
            tenantId: dashboard.tenantId,
            inventoryId: dashboard.inventoryId,
            principalId: scope.principalId
          };
          setLoadState({ status: 'ready', dashboard });
          setDraftContext(nextContext);
          applyDraft(applyInitialParentToDraft(addAssetDraftStore.load(nextContext) ?? emptyDraft, initialParent));
        }
      })
      .catch((error: unknown) => {
        if (isCurrent) {
          setLoadState({
            status: 'error',
            message: readableError(error, 'Could not load inventory context.')
          });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [addAssetDraftStore, addDraftScopeQuery, dashboardQuery, initialParent]);

  useEffect(() => {
    let isCurrent = true;
    if (!isParentMenuOpen && parentQuery.trim().length === 0) {
      setParentMatches([]);
      return;
    }

    parentLookupQuery
      .execute(parentQuery)
      .then((matches) => {
        if (isCurrent) {
          setParentMatches(matches);
        }
      })
      .catch(() => {
        if (isCurrent) {
          setParentMatches([]);
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [isParentMenuOpen, parentLookupQuery, parentQuery]);

  useEffect(() => {
    if (!draftContext) {
      return;
    }

    addAssetDraftStore.save(draftContext, {
      title,
      description,
      parentAssetId,
      parentQuery,
      selectedPhotos,
      selectedTagIds,
      newTags,
      showDetails,
      lastParent
    });
  }, [
    addAssetDraftStore,
    description,
    draftContext,
    lastParent,
    parentAssetId,
    parentQuery,
    selectedPhotos,
    selectedTagIds,
    newTags,
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
        title,
        description,
        parentAssetId: resolvedParentAssetId,
        tagIds: selectedTagIds,
        newTags,
        activeTags: loadState.status === 'ready' ? loadState.dashboard.assetTags : [],
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
      setTitle('');
      setDescription('');
      setParentAssetId(nextParent?.id);
      setParentQuery(nextParent?.title ?? '');
      setSelectedPhotos([]);
      setSelectedTagIds([]);
      setNewTags([]);
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
      setSaveState({ status: 'idle' });
      await refreshDashboardAfterTagCreation(newTags);
      feedback.showNotice({ tone: 'error', title: 'Could not save asset', message });
    }
  }

  async function refreshDashboardAfterTagCreation(stagedTags: readonly CreateAssetTagDraft[]): Promise<void> {
    try {
      const dashboard = await dashboardQuery.execute();
      setLoadState({ status: 'ready', dashboard });
      const reconciled = reconcileCreatedAssetTags(stagedTags, dashboard.assetTags);
      if (reconciled.createdTagIds.length > 0) {
        setSelectedTagIds((current) => uniqueStrings([...current, ...reconciled.createdTagIds]));
        setNewTags(reconciled.remainingTags);
      }
    } catch {
      // The original save error is the user-facing failure.
    }
  }

  async function createParent(): Promise<void> {
    const parentName = parentQuery.trim();
    if (loadState.status !== 'ready' || parentName.length === 0) {
      return;
    }

    setIsCreatingParent(true);
    setSaveState({ status: 'idle' });
    try {
      const result = await createAssetCommand.execute({
        kind: 'location',
        title: parentName,
        description: ''
      });
      const dashboard = await dashboardQuery.execute();
      const createdParent = {
        id: result.id,
        title: result.title,
        kind: 'location' as const,
        subtitle: 'New location',
        pathLabel: result.title,
        selectionHint: 'Location',
        willPromoteToContainer: false
      };
      setLoadState({ status: 'ready', dashboard });
      setParentAssetId(result.id);
      setParentQuery(result.title);
      setLastParent(createdParent);
      setCreatedParent(createdParent);
      setIsParentMenuOpen(true);
      setParentMatches((current) => [
        createdParent,
        ...current.filter((parent) => parent.id !== result.id)
      ]);
    } catch (error) {
      const message = readableError(error, 'Could not create parent.');
      setSaveState({ status: 'idle' });
      feedback.showNotice({ tone: 'error', title: 'Could not create parent', message });
    } finally {
      setIsCreatingParent(false);
    }
  }

  async function addPhotosFromLibrary(): Promise<void> {
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
      setSaveState({ status: 'idle' });
      feedback.showNotice({ tone: 'error', title: 'Could not select photos', message });
    }
  }

  async function takePhoto(): Promise<void> {
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
      setSaveState({ status: 'idle' });
      feedback.showNotice({ tone: 'error', title: 'Could not take photo', message });
    }
  }

  function removePhoto(photoId: string): void {
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
    showPhotoSourceChooser({
      onCamera: () => void takePhoto(),
      onLibrary: () => void addPhotosFromLibrary()
    });
  }

  function movePhoto(photoId: string, direction: number): void {
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
    const clearedDraft = { ...emptyDraft };
    applyDraft(clearedDraft);
    if (draftContext) {
      addAssetDraftStore.save(draftContext, clearedDraft);
    }
    setSaveState({ status: 'idle' });
  }

  function applyDraft(draft: AddAssetDraft): void {
    setTitle(draft.title);
    setDescription(draft.description);
    setParentAssetId(draft.parentAssetId);
    setParentQuery(draft.parentQuery);
    setSelectedPhotos(draft.selectedPhotos);
    setSelectedTagIds(draft.selectedTagIds ?? []);
    setNewTags(draft.newTags ?? []);
    setShowDetails(draft.showDetails);
    setLastParent(draft.lastParent);
  }

  return (
    <SafeAreaView style={styles.shell} edges={['top', 'left', 'right']}>
      <ScrollView
        contentContainerStyle={[
          styles.content,
          { paddingBottom: bottomChromeAllowance }
        ]}
        keyboardDismissMode="interactive"
        keyboardShouldPersistTaps="handled"
      >
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
          </View>
        ) : null}
        {loadState.status === 'ready' ? (
          <View>
            <Text style={styles.title}>Add</Text>
            <View style={styles.contextLine}>
              <IdentityLabel
                iconSize="xs"
                kind="inventory"
                label={loadState.dashboard.inventoryName}
                textStyle={styles.contextText}
              />
              <IdentityLabel
                iconSize="xs"
                kind="tenant"
                label={loadState.dashboard.tenantName}
                textStyle={styles.contextText}
              />
            </View>

            {!loadState.dashboard.canAdd ? (
              <View style={styles.unavailablePanel}>
                <Text style={styles.unavailableTitle}>Add is unavailable</Text>
                <Text style={styles.unavailableText}>
                  This inventory does not allow you to create assets.
                </Text>
              </View>
            ) : (
              <View>
                <PhotoCapture
                  draggingPhotoId={draggingPhotoId}
                  onBeginPhotoDrag={setDraggingPhotoId}
                  onEndPhotoDrag={() => setDraggingPhotoId(undefined)}
                  onAddPhotos={choosePhotoSource}
                  onMovePhoto={movePhoto}
                  onOpenPhoto={(index) => setPreviewPhotoIndex(index)}
                  onRemovePhoto={removePhoto}
                  photos={selectedPhotos}
                />

                <Text style={styles.fieldLabel}>What is it?</Text>
                <TextInput
                  accessibilityLabel="Asset name"
                  onChangeText={setTitle}
                  placeholder="Furnace filter, passport, camping bin"
                  placeholderTextColor={colors.textMuted}
                  style={styles.input}
                  value={title}
                />

                <ParentPicker
                  isCreatingParent={isCreatingParent}
                  createdParent={createdParent}
                  matches={parentMatches}
                  isOpen={isParentMenuOpen}
                  lastParent={lastParent}
                  onChangeQuery={(value) => {
                    setParentQuery(value);
                    setParentAssetId(undefined);
                    setCreatedParent(undefined);
                    setIsParentMenuOpen(true);
                  }}
                  onCreateParent={createParent}
                  onOpenChange={setIsParentMenuOpen}
                  onSelectParent={(parent) => {
                    setParentAssetId(parent?.id);
                    setParentQuery(parent?.title ?? '');
                    setLastParent(parent);
                    setCreatedParent(undefined);
                    setIsParentMenuOpen(false);
                  }}
                  parentAssetId={parentAssetId}
                  query={parentQuery}
                />

                <Pressable
                  accessibilityRole="button"
                  accessibilityState={{ expanded: showDetails }}
                  onPress={() => setShowDetails((current) => !current)}
                  style={styles.moreDetailsButton}
                >
                  <Text style={styles.moreDetailsText}>More details</Text>
                  {showDetails ? (
                    <ChevronUp color={colors.textMuted} size={18} strokeWidth={2.2} />
                  ) : (
                    <ChevronDown color={colors.textMuted} size={18} strokeWidth={2.2} />
                  )}
                </Pressable>

                {showDetails ? (
                  <View>
                    <TextInput
                      accessibilityLabel="Asset description"
                      multiline
                      onChangeText={setDescription}
                      placeholder="Description"
                      placeholderTextColor={colors.textMuted}
                      style={[styles.input, styles.textArea]}
                      value={description}
                    />
                    <AssetTagPicker
                      tags={loadState.dashboard.assetTags}
                      selectedTagIds={selectedTagIds}
                      onChange={setSelectedTagIds}
                      newTags={newTags}
                      onNewTagsChange={setNewTags}
                    />
                    <Pressable
                      accessibilityRole="button"
                      onPress={clearDraft}
                      style={styles.clearDraftButton}
                    >
                      <Text style={styles.clearDraftText}>Clear draft</Text>
                    </Pressable>
                  </View>
                ) : null}

                <Pressable
                  accessibilityRole="button"
                  disabled={saveState.status === 'saving'}
                  onPress={saveAsset}
                  style={[styles.saveButton, saveState.status === 'saving' ? styles.disabledButton : null]}
                >
                  {saveState.status === 'saving' ? (
                    <ActivityIndicator color={colors.onAction} />
                  ) : (
                    <Text style={styles.saveButtonText}>Save</Text>
                  )}
                </Pressable>
              </View>
            )}
          </View>
        ) : null}
      </ScrollView>
      <KeyboardDismissBar
        keyboardHeight={keyboardBar.keyboardHeight}
        visible={keyboardBar.isVisible}
      />
      <PhotoPreviewModal
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
  draggingPhotoId,
  onAddPhotos,
  onBeginPhotoDrag,
  onEndPhotoDrag,
  onMovePhoto,
  onOpenPhoto,
  onRemovePhoto,
  photos
}: {
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
          accessibilityHint="Choose camera or photo library"
          accessibilityRole="button"
          onPress={onAddPhotos}
          style={styles.addPhotoTile}
        >
          <ImagePlus color={colors.action} size={28} strokeWidth={2.2} />
        </Pressable>
        {photos.map((photo, index) => (
          <PhotoPreviewItem
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
          dragState.current.isDragging &&
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
    [index, onBeginPhotoDrag, onEndPhotoDrag, onMovePhoto, onOpenPhoto, photo.id, photoCount]
  );

  return (
    <View style={styles.photoPreviewShell}>
      <Pressable
        {...panResponder.panHandlers}
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
          dragState.current = { isDragging: true, didMove: false };
          suppressNextPress.current = true;
          onBeginPhotoDrag(photo.id);
        }}
        onPress={() => {
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
      <Pressable
        accessibilityLabel={`Remove ${photo.fileName}`}
        accessibilityRole="button"
        onPress={() => onRemovePhoto(photo.id)}
        style={styles.removePhotoButton}
      >
        <X color={colors.text} size={16} strokeWidth={2.4} />
      </Pressable>
    </View>
  );
}

function PhotoPreviewModal({
  currentIndex,
  onClose,
  onRemovePhoto,
  onSetIndex,
  photos
}: {
  readonly currentIndex: number | undefined;
  readonly onClose: () => void;
  readonly onRemovePhoto: (photoId: string) => void;
  readonly onSetIndex: (index: number | undefined) => void;
  readonly photos: readonly SelectedAssetPhoto[];
}) {
  function removeCurrentPhoto(photo: FullScreenPhotoViewerPhoto, index: number): void {
    if (!photo.id) {
      return;
    }

    Alert.alert('Remove photo?', 'This removes the photo from this new item draft.', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Remove',
        style: 'destructive',
        onPress: () => {
          onRemovePhoto(photo.id as string);
          if (photos.length <= 1) {
            onClose();
            return;
          }

          onSetIndex(Math.min(index, photos.length - 2));
        }
      }
    ]);
  }

  return (
    <FullScreenPhotoViewer
      canRemove
      currentIndex={currentIndex}
      onClose={onClose}
      onRemove={removeCurrentPhoto}
      onSelectIndex={onSetIndex}
      photos={photos.map((photo) => ({
        id: photo.id,
        label: photo.fileName,
        metadataLabel: photoMetadataLabel(photo),
        uri: photo.uri
      }))}
    />
  );
}

function ParentPicker({
  createdParent,
  isCreatingParent,
  isOpen,
  lastParent,
  matches,
  onChangeQuery,
  onCreateParent,
  onOpenChange,
  onSelectParent,
  parentAssetId,
  query
}: {
  readonly createdParent: ParentSelection | undefined;
  readonly isCreatingParent: boolean;
  readonly isOpen: boolean;
  readonly lastParent: ParentSelection | undefined;
  readonly matches: readonly ParentLookupResult[];
  readonly onChangeQuery: (value: string) => void;
  readonly onCreateParent: () => void;
  readonly onOpenChange: (isOpen: boolean) => void;
  readonly onSelectParent: (parent: ParentSelection | undefined) => void;
  readonly parentAssetId: string | undefined;
  readonly query: string;
}) {
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  const normalizedQuery = normalizeParentName(query);
  const exactParent = [createdParent, ...matches].filter(isParentSelection).find(
    (parent) => normalizeParentName(parent.title) === normalizedQuery
  );
  const canCreateParent = normalizedQuery.length > 0 && !exactParent;
  const selectedParent = resolveSelectedParent(matches, parentAssetId, query, lastParent);
  const createdParentId = createdParent?.id;

  return (
    <View style={styles.parentPicker}>
      <Text style={styles.sectionTitle}>Put in</Text>
      <Pressable
        accessibilityRole="button"
        accessibilityState={{ expanded: isOpen }}
        onPress={() => onOpenChange(!isOpen)}
        style={styles.parentSelectButton}
      >
        <View style={styles.parentSelectText}>
          <Text style={styles.parentTitle}>{selectedParent?.title ?? (query.trim() || 'No parent')}</Text>
          <Text style={styles.parentMeta}>
            {selectedParent
              ? `${selectedParent.selectionHint} · ${selectedParent.subtitle}`
              : 'Top level in this inventory'}
          </Text>
        </View>
        {isOpen ? (
          <ChevronUp color={colors.textMuted} size={18} strokeWidth={2.2} />
        ) : (
          <ChevronDown color={colors.textMuted} size={18} strokeWidth={2.2} />
        )}
      </Pressable>
      {isOpen ? (
        <View style={styles.parentMenu}>
          <TextInput
            accessibilityLabel="Search parent"
            autoFocus
            onChangeText={onChangeQuery}
            placeholder="Search or type new place"
            placeholderTextColor={colors.textMuted}
            style={styles.input}
            value={query}
          />
          {canCreateParent ? (
            <Pressable
              accessibilityRole="button"
              disabled={isCreatingParent}
              onPress={onCreateParent}
              style={[styles.createParentButton, isCreatingParent ? styles.disabledButton : null]}
            >
              {isCreatingParent ? (
                <ActivityIndicator color={colors.action} />
              ) : (
                <Text style={styles.createParentText}>Create "{query.trim()}" as a place</Text>
              )}
            </Pressable>
          ) : null}
          {createdParent ? (
            <ParentOption
              isSelected
              label={createdParent.title}
              leading="created"
              meta="Place created"
              onPress={() => onSelectParent(createdParent)}
            />
          ) : null}
          <ParentOption
            identityKind="inventory"
            isSelected={parentAssetId === undefined && query.trim().length === 0}
            label="No parent"
            meta="Top level in this inventory"
            onPress={() => onSelectParent(undefined)}
          />
          {matches.filter((parent) => parent.id !== createdParentId).map((parent) => (
            <ParentOption
              disabled={parent.canSelectAsParent === false}
              isSelected={parentAssetId === parent.id}
              key={parent.id}
              label={parent.title}
              meta={parent.disabledReason ?? `${parent.selectionHint} · ${parent.subtitle}`}
              onPress={() => {
                if (parent.canSelectAsParent === false) {
                  return;
                }
                onSelectParent(parent);
              }}
            />
          ))}
        </View>
      ) : null}
      {selectedParent?.willPromoteToContainer ? (
        <Text style={styles.parentPromotionText}>
          Stuff Stash will turn {selectedParent.title} into a container for this item.
        </Text>
      ) : null}
    </View>
  );
}

function AssetTagPicker({
  newTags,
  onNewTagsChange,
  tags,
  selectedTagIds,
  onChange
}: {
  readonly newTags: readonly CreateAssetTagDraft[];
  readonly onNewTagsChange: (tags: readonly CreateAssetTagDraft[]) => void;
  readonly tags: readonly AssetTagSummary[];
  readonly selectedTagIds: readonly string[];
  readonly onChange: (tagIds: readonly string[]) => void;
}) {
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  const [newTagName, setNewTagName] = useState('');
  const [newTagColor, setNewTagColor] = useState('');
  const selected = new Set(selectedTagIds);

  function toggleTag(tagId: string): void {
    if (selected.has(tagId)) {
      onChange(selectedTagIds.filter((current) => current !== tagId));
      return;
    }
    onChange([...selectedTagIds, tagId]);
  }

  function addNewTag(): void {
    const displayName = newTagName.trim();
    if (displayName.length === 0) {
      return;
    }
    const resolution = resolveInlineAssetTag({
      displayName,
      color: newTagColor,
      activeTags: tags,
      pendingTags: newTags
    });
    const transition = applyInlineAssetTagResolution({
      resolution,
      selectedTagIds,
      pendingTags: newTags
    });
    onChange(transition.selectedTagIds);
    onNewTagsChange(transition.pendingTags);
    if (transition.shouldClearInputs) {
      setNewTagName('');
      setNewTagColor('');
    }
  }

  const canAddNewTag = canResolveInlineAssetTag({
    displayName: newTagName,
    color: newTagColor,
    activeTags: tags,
    pendingTags: newTags
  });

  return (
    <View style={styles.tagPicker}>
      <Text style={styles.tagPickerTitle}>Tags</Text>
      <View style={styles.tagOptions}>
        {newTags.map((tag, index) => {
          const colorStyle = assetTagChipStylePresentation(tag);
          return (
            <Pressable
              accessibilityRole="button"
              key={`${tag.displayName}-${index.toString()}`}
              onPress={() => onNewTagsChange(newTags.filter((_, currentIndex) => currentIndex !== index))}
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
        {tags.map((tag) => {
          const isSelected = selected.has(tag.id);
          const colorStyle = assetTagChipStylePresentation(tag);
          return (
            <Pressable
              accessibilityRole="button"
              accessibilityState={{ selected: isSelected }}
              key={tag.id}
              onPress={() => toggleTag(tag.id)}
              style={[
                styles.tagOption,
                colorStyle.colored ? { backgroundColor: colorStyle.backgroundColor, borderColor: colorStyle.borderColor } : null,
                isSelected ? styles.tagOptionSelected : null
              ]}
            >
              <Text style={[styles.tagOptionText, isSelected ? styles.tagOptionTextSelected : null]} numberOfLines={1}>
                {tag.displayName}
              </Text>
              {isSelected ? <Check color={colors.action} size={14} strokeWidth={2.4} /> : null}
            </Pressable>
          );
        })}
      </View>
      <View style={styles.newTagRow}>
        <TextInput
          accessibilityLabel="New tag name"
          onChangeText={setNewTagName}
          placeholder="New tag"
          placeholderTextColor={colors.textMuted}
          style={[styles.input, styles.newTagNameInput]}
          value={newTagName}
        />
        <TextInput
          accessibilityLabel="New tag color"
          autoCapitalize="characters"
          onChangeText={setNewTagColor}
          placeholder="#2F80ED"
          placeholderTextColor={colors.textMuted}
          style={[styles.input, styles.newTagColorInput]}
          value={newTagColor}
        />
        <Pressable
          accessibilityRole="button"
          disabled={!canAddNewTag}
          onPress={addNewTag}
          style={[styles.newTagButton, !canAddNewTag ? styles.disabledButton : null]}
        >
          <Text style={styles.newTagButtonText}>Add</Text>
        </Pressable>
      </View>
      <TagColorPicker palette={colors} value={newTagColor} onChange={setNewTagColor} />
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

function ParentOption({
  disabled = false,
  identityKind,
  isSelected,
  label,
  leading,
  meta,
  onPress
}: {
  readonly disabled?: boolean;
  readonly identityKind?: 'inventory';
  readonly isSelected: boolean;
  readonly label: string;
  readonly leading?: 'created';
  readonly meta: string;
  readonly onPress: () => void;
}) {
  const colors = useAppearanceAwarePalette();
  const styles = createStyles(colors);
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ disabled, selected: isSelected }}
      disabled={disabled}
      onPress={onPress}
      style={[
        styles.parentOption,
        isSelected ? styles.parentOptionSelected : null,
        disabled ? styles.parentOptionDisabled : null
      ]}
    >
      <View style={styles.parentCheck}>
        {leading === 'created' ? (
          <Check color={colors.success} size={16} strokeWidth={2.6} />
        ) : (
          <Text style={styles.parentCheckText}>{isSelected ? '✓' : ''}</Text>
        )}
      </View>
      {identityKind ? <IdentityIcon kind={identityKind} size="sm" /> : null}
      <View style={styles.parentText}>
        <Text style={styles.parentTitle}>{label}</Text>
        <Text style={styles.parentMeta}>{meta}</Text>
      </View>
    </Pressable>
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
  title: {
    color: colors.text,
    fontSize: 30,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 36
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
  parentSelectButton: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    justifyContent: 'space-between',
    minHeight: 56,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  parentSelectText: {
    flex: 1,
    minWidth: 0
  },
  parentMenu: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    marginTop: spacing.xs,
    padding: spacing.xs
  },
  parentOption: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.sm,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    marginBottom: spacing.xs,
    minHeight: 48,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  parentOptionSelected: {
    borderColor: colors.action
  },
  parentOptionDisabled: {
    opacity: 0.58
  },
  createParentButton: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.action,
    borderRadius: radius.sm,
    borderWidth: 1,
    justifyContent: 'center',
    marginBottom: spacing.xs,
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  createParentText: {
    color: colors.action,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  parentCheck: {
    alignItems: 'center',
    justifyContent: 'center',
    width: 20
  },
  parentCheckText: {
    color: colors.action,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  },
  parentText: {
    flex: 1,
    minWidth: 0
  },
  parentTitle: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0
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
    aspectRatio: 1,
    marginRight: spacing.sm,
    position: 'relative',
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
  removePhotoButton: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderRadius: 999,
    height: 28,
    justifyContent: 'center',
    position: 'absolute',
    right: 6,
    top: 6,
    width: 28
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
    minHeight: 34,
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
    minHeight: 40,
    minWidth: 0
  },
  newTagColorInput: {
    minHeight: 40,
    width: 96
  },
  newTagButton: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    justifyContent: 'center',
    minHeight: 40,
    paddingHorizontal: spacing.sm
  },
  newTagButtonText: {
    color: colors.onAction,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0
  },
  clearDraftButton: {
    alignItems: 'center',
    alignSelf: 'center',
    justifyContent: 'center',
    minHeight: 40,
    paddingHorizontal: spacing.md
  },
  clearDraftText: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '800',
    letterSpacing: 0
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
  disabledButton: {
    opacity: 0.65
  },
  saveButtonText: {
    color: colors.onAction,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  }
  });
}
