import { Fragment, ReactNode, useState } from 'react';
import { router, Stack } from 'expo-router';
import {
  ActivityIndicator,
  Pressable,
  Alert,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { AssetCoreQuery } from '../../application/assets/AssetCoreQuery';
import type { AssetPlacementQuery } from '../../application/assets/AssetPlacementQuery';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';
import { useParentCandidates } from '../serverState/useParentCandidates';
import { MoveAssetCommand } from '../../application/assets/MoveAssetCommand';
import { UpdateAssetCommand } from '../../application/assets/UpdateAssetCommand';
import { InventoryAssetTagsQuery } from '../../application/assets/InventoryAssetTagsQuery';
import { CreateAssetCommand } from '../../application/add/CreateAssetCommand';
import { ParentLookupQuery, ParentLookupResult } from '../../application/add/ParentLookupQuery';
import { reconcileCreatedAssetTags, type CreateAssetTagDraft } from '../../application/assets/AssetTagDraftResolution';
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
import { recordAssetActionCompletion } from './AssetActionCompletion';
import {
  createdMoveDestinationParent,
  isSelectableMoveDestination,
  isSelectableMoveIntoCandidate,
  moveDestinationCreateInput,
  moveDestinationCreatePlacement,
  parentFromCurrentAssetPath
} from './AssetDetailMovePresentation';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing, type MobileColorPalette } from '../theme/tokens';

type ActionAssetQueries = {
  readonly assetId: string;
  readonly assetCoreQuery: Pick<AssetCoreQuery, 'execute'>;
  readonly assetPlacementQuery?: Pick<AssetPlacementQuery, 'execute'>;
};

function ActionAsset({ children, assetId, assetCoreQuery, assetPlacementQuery }: ActionAssetQueries & { children: (asset: AssetDetailViewModel) => ReactNode }) {
  const core = useMobileInventoryServerQuery({
    key: (scope, tenant, inventory) => mobileQueryKeys.assetCore(scope, tenant, inventory, assetId),
    query: (signal) => assetCoreQuery.execute(assetId, { signal })
  });
  const placement = useMobileInventoryServerQuery({
    key: (scope, tenant, inventory) => mobileQueryKeys.assetPlacement(scope, tenant, inventory, assetId, core.data?.view.parentAssetId ?? 'root'),
    query: (signal) => assetPlacementQuery!.execute(core.data!.snapshot, { signal }),
    enabled: Boolean(assetPlacementQuery && core.data)
  });
  if (!core.data) return core.isError ? <ErrorState message="Could not load asset." onRetry={() => void core.refetch()} /> : <LoadingState label="Loading asset" />;
  const asset = placement.data && assetPlacementQuery ? { ...core.data.view, parentLocationTrail: placement.data.parentLocationTrail, parentLocationTrailLabel: placement.data.parentLocationTrailLabel, locationTrailLabel: placement.data.locationTrailLabel, isPlacementLoading: false } : core.data.view;
  return <Fragment key={`${asset.tenantId}:${asset.inventoryId}:${asset.id}`}>
    {assetPlacementQuery && !placement.data ? <Text accessibilityLiveRegion="polite">{placement.isError ? 'Current placement could not be loaded.' : 'Loading current placement…'}</Text> : null}
    {assetPlacementQuery && placement.isError ? <Pressable accessibilityRole="button" onPress={() => void placement.refetch()}><Text>Retry placement</Text></Pressable> : null}
    {children(asset)}
  </Fragment>;
}

type EditProps = ActionAssetQueries & {
  readonly inventoryAssetTagsQuery: Pick<InventoryAssetTagsQuery, 'execute'>;
  readonly updateAssetCommand: Pick<UpdateAssetCommand, 'execute'>;
};
export function AssetEditSheetRouteScreen(props: EditProps) {
  return <ActionAsset {...props}>{(asset) => <EditAssetForm {...props} asset={asset} />}</ActionAsset>;
}
function EditAssetForm({ asset, inventoryAssetTagsQuery, updateAssetCommand }: EditProps & { asset: AssetDetailViewModel }) {
  const assetId = asset.id;
  const tags = useMobileInventoryServerQuery({ key: mobileQueryKeys.assetTags, query: (signal) => inventoryAssetTagsQuery.execute({ signal }) });
  const [draft, setDraft] = useState<EditDraft | undefined>(() => ({ title: asset.title, description: asset.description, tagIds: asset.tags?.map((tag) => tag.id) ?? [], newTags: [] }));
  const [isSaving, setIsSaving] = useState(false);

  function close(): void {
    if (!hasDirtyEditAssetDraft(asset, draft)) {
      router.back();
      return;
    }
    Alert.alert('Discard changes?', 'Your edits have not been saved.', [
      { text: 'Keep editing', style: 'cancel' },
      { text: 'Discard', style: 'destructive', onPress: () => router.back() }
    ]);
  }

  async function save(): Promise<void> {
    if (!draft) {
      return;
    }
    setIsSaving(true);
    try {
      const normalized = normalizedEditDraft(draft);
      const result = await updateAssetCommand.execute({
        assetId,
        title: normalized.title,
        description: normalized.description,
        tagIds: normalized.tagIds,
        newTags: normalized.newTags,
        activeTags: normalized.newTags?.length ? await tags.reconcile() : tags.data ?? []
      });
      recordAssetActionCompletion({
        assetId,
        action: 'edit',
        message: result.message,
        undoableOperationId: result.undoableOperationId
      });
      router.back();
    } catch (error) {
      await refreshEditAssetTags(normalizedEditDraft(draft).newTags ?? []);
      Alert.alert('Could not save changes', readableError(error, 'Asset update failed.'));
    } finally {
      setIsSaving(false);
    }
  }

  async function refreshEditAssetTags(stagedTags: readonly CreateAssetTagDraft[]): Promise<void> {
    try {
      const assetTags = await tags.reconcile();
      const reconciled = reconcileCreatedAssetTags(stagedTags, assetTags);
      if (reconciled.createdTagIds.length > 0) {
        setDraft((current) => current
          ? {
              ...current,
              tagIds: uniqueStrings([...(current.tagIds ?? []), ...reconciled.createdTagIds]),
              newTags: reconciled.remainingTags
            }
          : current);
      }
    } catch {
      // Preserve the original save error as the visible failure.
    }
  }

  return (
    <NativeSheetFrame title="Edit asset">
      {tags.isError ? <ErrorState message="Tags could not be loaded." onRetry={() => void tags.refetch()} /> : null}
      {(
        <EditAssetSheet
          asset={asset}
          assetTags={tags.data ?? []}
          draft={draft}
          isSaving={isSaving}
          onChange={setDraft}
          onClose={close}
          onSave={() => void save()}
        />
      )}
    </NativeSheetFrame>
  );
}

function uniqueStrings(values: readonly string[]): readonly string[] {
  return Array.from(new Set(values));
}

type MoveProps = ActionAssetQueries & {
  readonly createAssetCommand: Pick<CreateAssetCommand, 'execute'>;
  readonly moveAssetCommand: Pick<MoveAssetCommand, 'execute'>;
  readonly parentLookupQuery: Pick<ParentLookupQuery, 'execute'>;
};
export function AssetMoveSheetRouteScreen(props: MoveProps) {
  return <ActionAsset {...props}>{(asset) => <MoveAssetForm {...props} asset={asset} />}</ActionAsset>;
}
function MoveAssetForm({ asset, createAssetCommand, moveAssetCommand, parentLookupQuery }: MoveProps & { asset: AssetDetailViewModel }) {
  const assetId = asset.id;
  const [draft, setDraft] = useState<MoveDraft>(() => ({ createKind: 'location', query: '', matches: [], selectedParent: parentFromCurrentAssetPath(asset) }));
  const [isSaving, setIsSaving] = useState(false);
  const candidates = useParentCandidates(draft.query, parentLookupQuery);
  const shownDraft = { ...draft, selectedParent: draft.selectedParent?.id === asset.parentAssetId && !asset.isPlacementLoading ? parentFromCurrentAssetPath(asset) : draft.selectedParent, matches: moveDestinationMatches(candidates.data ?? [], asset) };

  async function createDestination(asset: AssetDetailViewModel): Promise<void> {
    const name = draft?.query.trim() ?? '';
    const createKind = draft?.createKind ?? 'location';
    if (name.length === 0) {
      return;
    }
    setIsSaving(true);
    try {
      const placement = moveDestinationCreatePlacement(asset);
      const created = await createAssetCommand.execute(moveDestinationCreateInput(createKind, name, placement));
      const createdParent = createdMoveDestinationParent({
        id: created.id,
        kind: createKind,
        placement,
        title: created.title
      });
      setDraft({
        createKind,
        query: created.title,
        matches: [createdParent, ...(draft?.matches ?? []).filter((match) => match.id !== asset.id)],
        selectedParent: createdParent
      });
    } catch (error) {
      Alert.alert('Could not create destination', readableError(error, 'Destination creation failed.'));
    } finally {
      setIsSaving(false);
    }
  }

  async function save(): Promise<void> {
    if (!draft) {
      return;
    }
    setIsSaving(true);
    try {
      const result = await moveAssetCommand.execute({
        assetId,
        parentAssetId: draft.selectedParent?.id
      });
      recordAssetActionCompletion({ assetId, action: 'move', message: result.message });
      router.back();
    } catch (error) {
      Alert.alert('Could not move asset', readableError(error, 'Move failed.'));
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <NativeSheetFrame title="Move asset">
      <CandidateStatus candidates={candidates} />
      {(
        <MoveAssetSheet
          asset={asset}
          draft={shownDraft}
          isSaving={isSaving}
          onChangeCreateKind={(createKind) => setDraft((current) => current ? { ...current, createKind } : current)}
          onChangeQuery={(query) => setDraft((current) => ({ ...current, query }))}
          onClose={() => router.back()}
          onCreateDestination={() => void createDestination(asset)}
          onSelectParent={(selectedParent) => setDraft((current) => current ? { ...current, selectedParent } : current)}
          onSelectRoot={() => setDraft((current) => current ? { ...current, selectedParent: null } : current)}
          onSave={() => void save()}
        />
      )}
    </NativeSheetFrame>
  );
}

type MoveHereProps = ActionAssetQueries & {
  readonly moveAssetCommand: Pick<MoveAssetCommand, 'execute'>;
  readonly parentLookupQuery: Pick<ParentLookupQuery, 'execute'>;
};
export function AssetMoveHereSheetRouteScreen(props: MoveHereProps) {
  return <ActionAsset {...props}>{(asset) => <MoveHereForm {...props} asset={asset} />}</ActionAsset>;
}
function MoveHereForm({ asset, moveAssetCommand, parentLookupQuery }: MoveHereProps & { asset: AssetDetailViewModel }) {
  const [draft, setDraft] = useState<MoveIntoDraft>({ target: asset, query: '', matches: [], selectedAsset: undefined });
  const [isSaving, setIsSaving] = useState(false);
  const candidates = useParentCandidates(draft.query, parentLookupQuery);
  const shownDraft = { ...draft, matches: (candidates.data ?? []).filter((match) => isSelectableMoveIntoCandidate(match, asset)) };

  async function save(): Promise<void> {
    if (!draft?.selectedAsset) {
      return;
    }
    setIsSaving(true);
    try {
      const result = await moveAssetCommand.execute({
        assetId: draft.selectedAsset.id,
        parentAssetId: draft.target.id
      });
      recordAssetActionCompletion({ assetId: draft.target.id, action: 'move', message: result.message });
      router.back();
    } catch (error) {
      Alert.alert('Could not move asset here', readableError(error, 'Move failed.'));
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <NativeSheetFrame title="Move something here">
      <CandidateStatus candidates={candidates} />
      {(
        <MoveThingsHereSheet
          draft={shownDraft}
          isSaving={isSaving}
          onChangeQuery={(query) => setDraft((current) => ({ ...current, query }))}
          onClose={() => router.back()}
          onSave={() => void save()}
          onSelectAsset={(selectedAsset) => setDraft((current) => current ? { ...current, selectedAsset } : current)}
        />
      )}
    </NativeSheetFrame>
  );
}

function NativeSheetFrame({
  children,
  title
}: {
  readonly children: ReactNode;
  readonly title: string;
}) {
  const styles = useStyles();
  return (
    <SafeAreaView style={styles.frame} edges={['left', 'right', 'bottom']}>
      <Stack.Screen options={{ title }} />
      {children}
    </SafeAreaView>
  );
}

function LoadingState({ label }: { readonly label: string }) {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={palette.action} />
      <Text style={styles.stateText}>{label}</Text>
    </View>
  );
}

function ErrorState({ message, onRetry }: { readonly message: string; readonly onRetry?: () => void }) {
  const styles = useStyles();
  return (
    <View style={styles.centerState}>
      <Text style={styles.errorTitle}>Could not load</Text>
      <Text style={styles.stateText}>{message}</Text>
      {onRetry ? <Pressable accessibilityRole="button" onPress={onRetry}><Text style={styles.stateText}>Try again</Text></Pressable> : null}
    </View>
  );
}

function CandidateStatus({ candidates }: { candidates: ReturnType<typeof useParentCandidates> }) {
  if (candidates.isError) return <View><Text accessibilityRole="alert">Suggestions could not be loaded.</Text><Pressable accessibilityRole="button" onPress={() => void candidates.refetch()}><Text>Retry suggestions</Text></Pressable></View>;
  if (!candidates.data) return <Text accessibilityLiveRegion="polite">Loading suggestions…</Text>;
  return null;
}

function moveDestinationMatches(
  matches: readonly ParentLookupResult[],
  asset: AssetDetailViewModel
): readonly ParentLookupResult[] {
  return matches.filter((match) => match.id !== asset.id && isSelectableMoveDestination(match));
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

function useStyles() {
  return createStyles(useAppearancePalette());
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  frame: {
    backgroundColor: colors.surface,
    flex: 1
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
    fontSize: 22,
    fontWeight: '900',
    letterSpacing: 0
  }
  });
}
