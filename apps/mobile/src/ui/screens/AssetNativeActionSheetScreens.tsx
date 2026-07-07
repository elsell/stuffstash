import { ReactNode, useEffect, useState } from 'react';
import { router, Stack } from 'expo-router';
import {
  ActivityIndicator,
  Alert,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { AssetAuditHistoryQuery } from '../../application/assets/AssetAuditHistoryQuery';
import { AssetCheckoutHistoryQuery } from '../../application/assets/AssetCheckoutHistoryQuery';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { AssetDetailQuery } from '../../application/assets/AssetDetailQuery';
import { MoveAssetCommand } from '../../application/assets/MoveAssetCommand';
import { UpdateAssetCommand } from '../../application/assets/UpdateAssetCommand';
import { InventoryAssetTagsQuery, type AssetTagOptionViewModel } from '../../application/assets/InventoryAssetTagsQuery';
import { CreateAssetCommand } from '../../application/add/CreateAssetCommand';
import { ParentLookupQuery, ParentLookupResult } from '../../application/add/ParentLookupQuery';
import { AssetAuditHistorySheet, AssetAuditHistorySheetState } from './AssetAuditHistorySheet';
import {
  AssetCheckoutHistorySheet,
  AssetCheckoutHistorySheetState
} from './AssetCheckoutHistorySheet';
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
import { colors, spacing } from '../theme/tokens';

type LoadableAssetState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly asset: AssetDetailViewModel; readonly assetTags?: readonly AssetTagOptionViewModel[] }
  | { readonly status: 'error'; readonly message: string };

export function AssetEditSheetRouteScreen({
  assetDetailQuery,
  assetId,
  inventoryAssetTagsQuery,
  updateAssetCommand
}: {
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetId: string;
  readonly inventoryAssetTagsQuery: InventoryAssetTagsQuery;
  readonly updateAssetCommand: UpdateAssetCommand;
}) {
  const [state, setState] = useState<LoadableAssetState>({ status: 'loading' });
  const [draft, setDraft] = useState<EditDraft | undefined>();
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    let isCurrent = true;
    assetDetailQuery
      .execute(assetId)
      .then((asset) => {
        if (isCurrent) {
          setState({ status: 'ready', asset, assetTags: [] });
          setDraft({
            title: asset.title,
            description: asset.description,
            tagIds: asset.tags?.map((tag) => tag.id) ?? [],
            newTags: []
          });
        }
        return inventoryAssetTagsQuery.execute();
      })
      .then((assetTags) => {
        if (isCurrent) {
          setState((current) => current.status === 'ready' ? { ...current, assetTags } : current);
        }
      })
      .catch((error: unknown) => {
        if (isCurrent) {
          setState((current) => current.status === 'ready'
            ? current
            : { status: 'error', message: readableError(error, 'Could not load asset.') });
        }
      });
    return () => {
      isCurrent = false;
    };
  }, [assetDetailQuery, assetId, inventoryAssetTagsQuery]);

  function close(): void {
    if (state.status !== 'ready' || !hasDirtyEditAssetDraft(state.asset, draft)) {
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
        newTags: normalized.newTags
      });
      recordAssetActionCompletion({ assetId, action: 'edit', message: result.message });
      router.back();
    } catch (error) {
      await refreshEditAssetTags();
      Alert.alert('Could not save changes', readableError(error, 'Asset update failed.'));
    } finally {
      setIsSaving(false);
    }
  }

  async function refreshEditAssetTags(): Promise<void> {
    try {
      const assetTags = await inventoryAssetTagsQuery.execute();
      setState((current) => current.status === 'ready' ? { ...current, assetTags } : current);
    } catch {
      // Preserve the original save error as the visible failure.
    }
  }

  return (
    <NativeSheetFrame title="Edit asset">
      {state.status === 'loading' ? <LoadingState label="Loading asset" /> : null}
      {state.status === 'error' ? <ErrorState message={state.message} /> : null}
      {state.status === 'ready' ? (
        <EditAssetSheet
          asset={state.asset}
          assetTags={state.assetTags ?? []}
          draft={draft}
          isSaving={isSaving}
          onChange={setDraft}
          onClose={close}
          onSave={() => void save()}
        />
      ) : null}
    </NativeSheetFrame>
  );
}

export function AssetMoveSheetRouteScreen({
  assetDetailQuery,
  assetId,
  createAssetCommand,
  moveAssetCommand,
  parentLookupQuery
}: {
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetId: string;
  readonly createAssetCommand: CreateAssetCommand;
  readonly moveAssetCommand: MoveAssetCommand;
  readonly parentLookupQuery: ParentLookupQuery;
}) {
  const [state, setState] = useState<LoadableAssetState>({ status: 'loading' });
  const [draft, setDraft] = useState<MoveDraft | undefined>();
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    let isCurrent = true;
    async function load(): Promise<void> {
      try {
        const asset = await assetDetailQuery.execute(assetId);
        const matches = await parentLookupQuery.execute('');
        const safeMatches = moveDestinationMatches(matches, asset);
        const currentParent = asset.parentAssetId
          ? safeMatches.find((match) => match.id === asset.parentAssetId) ?? parentFromCurrentAssetPath(asset)
          : null;
        if (isCurrent) {
          setState({ status: 'ready', asset });
          setDraft({
            createKind: 'location',
            query: currentParent?.title ?? '',
            matches: safeMatches,
            selectedParent: currentParent
          });
        }
      } catch (error) {
        if (isCurrent) {
          setState({ status: 'error', message: readableError(error, 'Could not load move options.') });
        }
      }
    }
    void load();
    return () => {
      isCurrent = false;
    };
  }, [assetDetailQuery, assetId, parentLookupQuery]);

  async function updateQuery(query: string, asset: AssetDetailViewModel): Promise<void> {
    setDraft((current) => current ? { ...current, query } : current);
    const matches = await parentLookupQuery.execute(query);
    setDraft((current) => current && current.query === query
      ? { ...current, matches: moveDestinationMatches(matches, asset) }
      : current);
  }

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
      {state.status === 'loading' ? <LoadingState label="Loading move options" /> : null}
      {state.status === 'error' ? <ErrorState message={state.message} /> : null}
      {state.status === 'ready' ? (
        <MoveAssetSheet
          asset={state.asset}
          draft={draft}
          isSaving={isSaving}
          onChangeCreateKind={(createKind) => setDraft((current) => current ? { ...current, createKind } : current)}
          onChangeQuery={(query) => void updateQuery(query, state.asset)}
          onClose={() => router.back()}
          onCreateDestination={() => void createDestination(state.asset)}
          onSelectParent={(selectedParent) => setDraft((current) => current ? { ...current, selectedParent } : current)}
          onSelectRoot={() => setDraft((current) => current ? { ...current, selectedParent: null } : current)}
          onSave={() => void save()}
        />
      ) : null}
    </NativeSheetFrame>
  );
}

export function AssetMoveHereSheetRouteScreen({
  assetDetailQuery,
  assetId,
  moveAssetCommand,
  parentLookupQuery
}: {
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetId: string;
  readonly moveAssetCommand: MoveAssetCommand;
  readonly parentLookupQuery: ParentLookupQuery;
}) {
  const [state, setState] = useState<LoadableAssetState>({ status: 'loading' });
  const [draft, setDraft] = useState<MoveIntoDraft | undefined>();
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    let isCurrent = true;
    async function load(): Promise<void> {
      try {
        const asset = await assetDetailQuery.execute(assetId);
        const matches = await parentLookupQuery.execute('');
        if (isCurrent) {
          setState({ status: 'ready', asset });
          setDraft({
            target: asset,
            query: '',
            matches: matches.filter((match) => isSelectableMoveIntoCandidate(match, asset)),
            selectedAsset: undefined
          });
        }
      } catch (error) {
        if (isCurrent) {
          setState({ status: 'error', message: readableError(error, 'Could not load move options.') });
        }
      }
    }
    void load();
    return () => {
      isCurrent = false;
    };
  }, [assetDetailQuery, assetId, parentLookupQuery]);

  async function updateQuery(query: string, target: AssetDetailViewModel): Promise<void> {
    setDraft((current) => current ? { ...current, query } : current);
    const matches = await parentLookupQuery.execute(query);
    setDraft((current) => current && current.query === query
      ? { ...current, matches: matches.filter((match) => isSelectableMoveIntoCandidate(match, target)) }
      : current);
  }

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
      {state.status === 'loading' ? <LoadingState label="Loading move options" /> : null}
      {state.status === 'error' ? <ErrorState message={state.message} /> : null}
      {state.status === 'ready' ? (
        <MoveThingsHereSheet
          draft={draft}
          isSaving={isSaving}
          onChangeQuery={(query) => void updateQuery(query, state.asset)}
          onClose={() => router.back()}
          onSave={() => void save()}
          onSelectAsset={(selectedAsset) => setDraft((current) => current ? { ...current, selectedAsset } : current)}
        />
      ) : null}
    </NativeSheetFrame>
  );
}

export function AssetAuditSheetRouteScreen({
  assetAuditHistoryQuery,
  assetDetailQuery,
  assetId
}: {
  readonly assetAuditHistoryQuery: AssetAuditHistoryQuery;
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetId: string;
}) {
  const [state, setState] = useState<AssetAuditHistorySheetState>({ status: 'loading', assetTitle: 'Asset' });

  useEffect(() => {
    let isCurrent = true;
    async function load(): Promise<void> {
      let loadedTitle = 'Asset';
      try {
        const asset = await assetDetailQuery.execute(assetId);
        loadedTitle = asset.title;
        if (isCurrent) {
          setState({ status: 'loading', assetTitle: asset.title });
        }
        const history = await assetAuditHistoryQuery.execute({ assetId, limit: 20 });
        if (isCurrent) {
          setState({ status: 'ready', assetTitle: asset.title, history });
        }
      } catch (error) {
        if (isCurrent) {
          setState({
            status: 'error',
            assetTitle: loadedTitle,
            message: readableError(error, 'Audit history failed.')
          });
        }
      }
    }
    void load();
    return () => {
      isCurrent = false;
    };
  }, [assetAuditHistoryQuery, assetDetailQuery, assetId]);

  return (
    <NativeSheetFrame title="Audit history">
      <AssetAuditHistorySheet state={state} onClose={() => router.back()} />
    </NativeSheetFrame>
  );
}

export function AssetCheckoutHistorySheetRouteScreen({
  assetCheckoutHistoryQuery,
  assetDetailQuery,
  assetId
}: {
  readonly assetCheckoutHistoryQuery: AssetCheckoutHistoryQuery;
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetId: string;
}) {
  const [state, setState] = useState<AssetCheckoutHistorySheetState>({ status: 'loading', assetTitle: 'Asset' });

  useEffect(() => {
    let isCurrent = true;
    async function load(): Promise<void> {
      let loadedTitle = 'Asset';
      try {
        const asset = await assetDetailQuery.execute(assetId);
        loadedTitle = asset.title;
        if (isCurrent) {
          setState({ status: 'loading', assetTitle: asset.title });
        }
        const history = await assetCheckoutHistoryQuery.execute({ assetId, limit: 20 });
        if (isCurrent) {
          setState({ status: 'ready', assetTitle: asset.title, history });
        }
      } catch (error) {
        if (isCurrent) {
          setState({
            status: 'error',
            assetTitle: loadedTitle,
            message: readableError(error, 'Checkout history failed.')
          });
        }
      }
    }
    void load();
    return () => {
      isCurrent = false;
    };
  }, [assetCheckoutHistoryQuery, assetDetailQuery, assetId]);

  return (
    <NativeSheetFrame title="Checkout history">
      <AssetCheckoutHistorySheet state={state} onClose={() => router.back()} />
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
  return (
    <SafeAreaView style={styles.frame} edges={['left', 'right', 'bottom']}>
      <Stack.Screen options={{ title }} />
      {children}
    </SafeAreaView>
  );
}

function LoadingState({ label }: { readonly label: string }) {
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={colors.action} />
      <Text style={styles.stateText}>{label}</Text>
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

function moveDestinationMatches(
  matches: readonly ParentLookupResult[],
  asset: AssetDetailViewModel
): readonly ParentLookupResult[] {
  return matches.filter((match) => match.id !== asset.id && isSelectableMoveDestination(match));
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

const styles = StyleSheet.create({
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
