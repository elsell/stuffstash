import { useTaskPresentation } from '../navigation/useTaskPresentation';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { usePreventRemove } from '@react-navigation/native';
import type { InventoryAssetTypesQuery } from '../../application/assets/InventoryAssetTypesQuery';
import { Fragment, ReactNode, useEffect, useLayoutEffect, useRef, useState } from 'react';
import { router, Stack, useNavigation } from 'expo-router';
import {
  ActivityIndicator,
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
  canSaveEditAsset,
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
  if (!core.data) return <AssetLoadState failed={core.isError} onRetry={() => void core.refetch()} />;
  const asset = placement.data && assetPlacementQuery ? { ...core.data.view, parentLocationTrail: placement.data.parentLocationTrail, parentLocationTrailLabel: placement.data.parentLocationTrailLabel, locationTrailLabel: placement.data.locationTrailLabel, isPlacementLoading: false } : core.data.view;
  return <Fragment key={`${asset.tenantId}:${asset.inventoryId}:${asset.id}`}>
    {assetPlacementQuery && !placement.data ? <Text accessibilityLiveRegion="polite">{placement.isError ? 'Current placement could not be loaded.' : 'Loading current placement…'}</Text> : null}
    {assetPlacementQuery && placement.isError ? <NativeCommandButton label="Retry placement" onPress={() => void placement.refetch()} /> : null}
    {children(asset)}
  </Fragment>;
}

function AssetLoadState({ failed, onRetry }: { readonly failed: boolean; readonly onRetry: () => void }) {
  const styles = useStyles();
  return <SafeAreaView style={styles.frame} edges={['left', 'right', 'bottom']}>
    {failed ? <ErrorState message="Could not load asset." onRetry={onRetry} /> : <LoadingState label="Loading asset" />}
    <NativeCommandButton label="Close" onPress={returnFromAssetAction} />
  </SafeAreaView>;
}

type EditProps = ActionAssetQueries & {
  readonly inventoryAssetTypesQuery: Pick<InventoryAssetTypesQuery, 'execute'>;
  readonly inventoryAssetTagsQuery: Pick<InventoryAssetTagsQuery, 'execute'>;
  readonly updateAssetCommand: Pick<UpdateAssetCommand, 'execute'>;
};
export function AssetEditSheetRouteScreen(props: EditProps) {
  return <ActionAsset {...props}>{(asset) => <EditAssetForm {...props} asset={asset} />}</ActionAsset>;
}
function EditAssetForm({ asset, inventoryAssetTypesQuery, inventoryAssetTagsQuery, updateAssetCommand }: EditProps & { asset: AssetDetailViewModel }) {
  const assetId = asset.id;
  const types = useMobileInventoryServerQuery({ key: (scope, tenant, inventory) => mobileQueryKeys.customization(scope, tenant, inventory, 'inventory', 'asset-type-choices', 'active'), query: (signal) => inventoryAssetTypesQuery.execute(asset.tenantId ?? '', asset.inventoryId ?? '', { signal }) });
  const tags = useMobileInventoryServerQuery({ key: mobileQueryKeys.assetTags, query: (signal) => inventoryAssetTagsQuery.execute({ signal }) });
  const [draft, setDraft] = useState<EditDraft | undefined>(() => ({ title: asset.title, description: asset.description, tagIds: asset.tags?.map((tag) => tag.id) ?? [], newTags: [] }));
  const operation = useAssetSheetOperation(asset.canEdit, leave => close(leave));
  const isSaving = operation.busy;
  const captureDiscard = useTaskPresentation(undefined, JSON.stringify([assetId, draft, isSaving]));

  function close(leave: () => void = returnFromAssetAction): void {
    if (operation.locked()) return;
    if (!hasDirtyEditAssetDraft(asset, draft)) {
      operation.leave(leave);
      return;
    }
    const isCurrent = captureDiscard();
    if (!isCurrent()) return;
    let accepted = false;
    Alert.alert('Discard changes?', 'Your edits have not been saved.', [
      { text: 'Keep editing', style: 'cancel' },
      { text: 'Discard', style: 'destructive', onPress: () => {
        if (!isCurrent() || accepted || operation.locked()) return;
        accepted = true;
        operation.leave(leave);
      } }
    ]);
  }

  async function save(): Promise<void> {
    if (!draft || !canSaveEditAsset(asset, draft)) {
      return;
    }
    if (!operation.begin()) return;
    try {
      const normalized = normalizedEditDraft(draft);
      const activeTags = normalized.newTags?.length ? await tags.reconcile() : tags.data ?? [];
      if (!operation.canMutate()) return;
      const result = await updateAssetCommand.execute({
        assetId,
        expiration: normalized.expiration,
        customAssetTypeId: asset.customAssetTypeId ? undefined : normalized.customAssetTypeId,
        title: normalized.title,
        description: normalized.description,
        tagIds: normalized.tagIds,
        newTags: normalized.newTags,
        activeTags
      });
      if (!operation.canPresent()) return;
      recordAssetActionCompletion({
        assetId,
        action: 'edit',
        message: result.message,
        undoableOperationId: result.undoableOperationId
      });
      operation.complete(returnFromAssetAction);
    } catch (error) {
      if (!operation.canPresent()) return;
      await refreshEditAssetTags(normalizedEditDraft(draft).newTags ?? []);
      if (operation.canPresent()) Alert.alert('Could not save changes', readableError(error, 'Asset update failed.'));
    } finally {
      operation.end();
    }
  }

  async function refreshEditAssetTags(stagedTags: readonly CreateAssetTagDraft[]): Promise<void> {
    try {
      const assetTags = await tags.reconcile();
      const reconciled = reconcileCreatedAssetTags(stagedTags, assetTags);
      if (operation.canPresent() && reconciled.createdTagIds.length > 0) {
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
    <NativeSheetFrame title="Edit asset" busy={isSaving} dismissible={false}>
      <EditAssetSheet
        readOnly={!asset.canEdit}
        metadataRecovery={<>
          {types.isError ? <InlineQueryError message="Asset types could not be loaded." retryLabel="Retry asset types" onRetry={() => void types.refetch()} /> : null}
          {tags.isError ? <InlineQueryError message="Tags could not be loaded." retryLabel="Retry tags" onRetry={() => void tags.refetch()} /> : null}
        </>}
        asset={asset}
        assetTypes={types.data}
        assetTypesFailed={types.isError}
        assetTags={tags.data ?? []}
        draft={draft}
        isSaving={isSaving}
        onChange={next => operation.change(() => setDraft(next))}
        onClose={close}
        onSave={() => void save()}
      />
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
  const operation = useAssetSheetOperation(asset.canMove);
  const isSaving = operation.busy;
  const candidates = useParentCandidates(draft.query, parentLookupQuery);
  const shownDraft = { ...draft, selectedParent: draft.selectedParent?.id === asset.parentAssetId && !asset.isPlacementLoading ? parentFromCurrentAssetPath(asset) : draft.selectedParent, matches: moveDestinationMatches(candidates.data ?? [], asset) };

  async function createDestination(asset: AssetDetailViewModel): Promise<void> {
    const name = draft?.query.trim() ?? '';
    const createKind = draft?.createKind ?? 'location';
    if (name.length === 0) {
      return;
    }
    if (!operation.begin('create')) return;
    try {
      const placement = moveDestinationCreatePlacement(asset);
      const created = await createAssetCommand.execute(moveDestinationCreateInput(createKind, name, placement));
      if (!operation.canPresent()) return;
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
      if (operation.canPresent()) Alert.alert('Could not create destination', readableError(error, 'Destination creation failed.'));
    } finally {
      operation.end();
    }
  }

  async function save(): Promise<void> {
    if (!draft) {
      return;
    }
    if (!operation.begin()) return;
    try {
      const result = await moveAssetCommand.execute({
        assetId,
        parentAssetId: draft.selectedParent?.id
      });
      if (!operation.canPresent()) return;
      recordAssetActionCompletion({ assetId, action: 'move', message: result.message });
      operation.complete(returnFromAssetAction);
    } catch (error) {
      if (operation.canPresent()) Alert.alert('Could not move asset', readableError(error, 'Move failed.'));
    } finally {
      operation.end();
    }
  }

  return (
    <NativeSheetFrame title="Move asset" busy={isSaving}>
      {(
        <MoveAssetSheet
          readOnly={!asset.canMove}
          candidatesAvailable={candidates.data !== undefined}
          candidateStatus={<CandidateStatus candidates={candidates} />}
          asset={asset}
          draft={shownDraft}
          isSaving={isSaving}
          isCreatingDestination={operation.kind === 'create'}
          onChangeCreateKind={(createKind) => operation.change(() => setDraft((current) => current ? { ...current, createKind } : current))}
          onChangeQuery={(query) => operation.change(() => setDraft((current) => ({ ...current, query })))}
          onClose={() => operation.leave(returnFromAssetAction)}
          onCreateDestination={() => void createDestination(asset)}
          onSelectParent={(selectedParent) => operation.change(() => setDraft((current) => current ? { ...current, selectedParent } : current))}
          onSelectRoot={() => operation.change(() => setDraft((current) => current ? { ...current, selectedParent: null } : current))}
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
  const operation = useAssetSheetOperation(asset.canMove && asset.canContainAssets);
  const isSaving = operation.busy;
  const candidates = useParentCandidates(draft.query, parentLookupQuery);
  const shownDraft = { ...draft, matches: (candidates.data ?? []).filter((match) => isSelectableMoveIntoCandidate(match, asset)) };

  async function save(): Promise<void> {
    if (!draft?.selectedAsset) {
      return;
    }
    if (!operation.begin()) return;
    try {
      const result = await moveAssetCommand.execute({
        assetId: draft.selectedAsset.id,
        parentAssetId: draft.target.id
      });
      if (!operation.canPresent()) return;
      recordAssetActionCompletion({ assetId: draft.target.id, action: 'move', message: result.message });
      operation.complete(returnFromAssetAction);
    } catch (error) {
      if (operation.canPresent()) Alert.alert('Could not move asset here', readableError(error, 'Move failed.'));
    } finally {
      operation.end();
    }
  }

  return (
    <NativeSheetFrame title="Move something here" busy={isSaving}>
      {(
        <MoveThingsHereSheet
          readOnly={!asset.canMove || !asset.canContainAssets}
          candidatesAvailable={candidates.data !== undefined}
          candidateStatus={<CandidateStatus candidates={candidates} />}
          draft={shownDraft}
          isSaving={isSaving}
          onChangeQuery={(query) => operation.change(() => setDraft((current) => ({ ...current, query })))}
          onClose={() => operation.leave(returnFromAssetAction)}
          onSave={() => void save()}
          onSelectAsset={(selectedAsset) => operation.change(() => setDraft((current) => current ? { ...current, selectedAsset } : current))}
        />
      )}
    </NativeSheetFrame>
  );
}

function returnFromAssetAction(): void {
  if (router.canGoBack()) router.back(); else router.replace('/');
}

function NativeSheetFrame({
  children,
  title, busy, dismissible = true
}: {
  readonly children: ReactNode;
  readonly title: string;
  readonly busy: boolean;
  readonly dismissible?: boolean;
}) {
  const styles = useStyles();
  return (
    <SafeAreaView style={styles.frame} edges={['left', 'right', 'bottom']}>
      <Stack.Screen options={{ title, gestureEnabled: dismissible && !busy }} />
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
      {onRetry ? <NativeCommandButton label="Retry asset" onPress={onRetry} /> : null}
    </View>
  );
}

function InlineQueryError({ message, retryLabel, onRetry }: {
  readonly message: string; readonly retryLabel: string; readonly onRetry: () => void;
}) {
  const styles = useStyles();
  return <View style={styles.inlineError}>
    <Text accessibilityRole="alert" style={styles.inlineErrorText}>{message}</Text>
    <NativeCommandButton label={retryLabel} onPress={onRetry} />
  </View>;
}

function CandidateStatus({ candidates }: { candidates: ReturnType<typeof useParentCandidates> }) {
  if (candidates.isError) return <InlineQueryError message="Suggestions could not be loaded." retryLabel="Retry suggestions" onRetry={() => void candidates.refetch()} />;
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
  inlineError: { paddingHorizontal: spacing.md, gap: spacing.xs },
  inlineErrorText: { color: colors.text, fontSize: 16 },
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

/** One mutation owns the sheet draft until it settles. */
function useAssetSheetOperation(eligible: boolean, onRemove?: (leave: () => void) => void) {
  const eligibility = useRef(eligible);
  useLayoutEffect(() => { eligibility.current = eligible; }, [eligible]);
  const capturePresentation = useTaskPresentation();
  const completionOwner = useRef<(() => boolean) | undefined>(undefined);
  const pending = useRef(false);
  const mounted = useRef(true);
  const [kind, setKind] = useState<'save' | 'create' | null>(null);
  const navigation = useNavigation();
  const completed = useRef(false);
  usePreventRemove(kind !== null || !!onRemove, ({ data }) => {
    if (completed.current) { navigation.dispatch(data.action); return; }
    if (!pending.current) onRemove?.(() => navigation.dispatch(data.action));
  });
  useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
  return {
    kind, busy: kind !== null,
    canPresent: () => mounted.current && completionOwner.current?.() === true,
    locked: () => pending.current || !mounted.current,
    canMutate: () => eligibility.current && mounted.current,
    begin: (next: 'save' | 'create' = 'save') => {
      const isCurrent = capturePresentation();
      if (pending.current || !mounted.current || !eligibility.current || !isCurrent()) return false;
      completionOwner.current = isCurrent;
      pending.current = true; completed.current = false; setKind(next); return true;
    },
    complete: (leave: () => void) => { if (mounted.current && completionOwner.current?.()) { completed.current = true; leave(); } },
    end: () => { pending.current = false; if (mounted.current) setKind(null); },
    change: (change: () => void) => { if (!pending.current && mounted.current && eligibility.current) change(); },
    leave: (leave: () => void) => { if (!pending.current && mounted.current) { completed.current = true; leave(); } }
  };
}
