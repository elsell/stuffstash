import { NativeSegmentedControl } from '../components/NativeSegmentedControl';
import { Stack } from 'expo-router';
import { SettingsSection, useSettingsListStyles } from './SettingsList';
import { useHeaderHeight } from '@react-navigation/elements';
import { MoveSelectionList } from '../components/MoveSelectionList';
import type { MoveSelectionRowModel, MoveSelectionStatus } from '../components/MoveSelectionList.types';
import { NativeNavigationSearch } from '../components/NativeNavigationSearch';
import { NativeFilterSearch } from '../components/NativeFilterSearch.android';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';
import { DraftTextField } from '../components/DraftTextField';
import { useFocusedSheetActions } from '../components/useFocusedSheetActions';
import { AssetActionKeyboardFrame } from './AssetActionKeyboardFrame';
import { AssetTagSelectionField } from '../components/AssetTagSelectionField';
import { NativeChoicePicker } from '../components/NativeChoicePicker';
import { AssetExpirationEditor } from '../components/AssetExpirationEditor';
import type { CustomAssetTypeDefinition } from '../../domain/customization/Customization';
import { useMemo, type ReactNode } from 'react';
import {
  Platform,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { AssetTagOptionViewModel } from '../../application/assets/InventoryAssetTagsQuery';
import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';
import {
  type CreateAssetTagDraft
} from '../../application/assets/AssetTagDraftResolution';
import { AppTextInput, appKeyboardDismissMode } from '../components/AppTextInput';
import {
  assetEditContext,
  canSaveEditAsset,
  EditDraft
} from './AssetDetailEditPresentation';
import {
  canCreateMoveDestination,
  canSaveMoveAsset,
  moveIntoCandidateRow,
  moveIntoEmptyState,
  moveDestinationCreateKindHelp,
  moveDestinationCreatePlacement,
  moveDestinationCreatePlacementLabel,
  type MoveDestinationCreateKind,
  movePlacementPreview
} from './AssetDetailMovePresentation';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';

export type MoveDraft = {
  readonly creationName?: string;
  readonly query: string;
  readonly matches: readonly ParentLookupResult[];
  readonly selectedParent: ParentLookupResult | null;
  readonly createKind: MoveDestinationCreateKind;
};

export type MoveIntoDraft = {
  readonly target: AssetDetailViewModel;
  readonly query: string;
  readonly matches: readonly ParentLookupResult[];
  readonly selectedAsset?: ParentLookupResult;
};

export function EditAssetSheet({
  readOnly = false,
  asset,
  assetTypes,
  assetTypesFailed = false,
  assetTags,
  metadataRecovery,
  draft,
  isSaving,
  onChange,
  onClose,
  onSave
}: {
  readonly readOnly?: boolean;
  readonly asset: AssetDetailViewModel;
  readonly assetTypes?: readonly CustomAssetTypeDefinition[];
  readonly assetTypesFailed?: boolean;
  readonly assetTags: readonly AssetTagOptionViewModel[];
  readonly metadataRecovery?: ReactNode;
  readonly draft: EditDraft | undefined;
  readonly isSaving: boolean;
  readonly onChange: (draft: EditDraft) => void;
  readonly onClose: () => void;
  readonly onSave: () => void;
}) {
  const styles = useStyles();
  const editContext = assetEditContext(asset);
  const disabled = isSaving || readOnly;
  const canSave = canSaveEditAsset(asset, draft) && !disabled;
  const actions = useFocusedSheetActions({
    primaryLabel: 'Save', secondaryLabel: 'Cancel', disabled: !canSave,
    secondaryDisabled: isSaving, onApply: onSave, onBack: onClose
  });
  const cancelOptions = useNativeHeaderActionOptions([{ kind: 'close', label: 'Cancel',
    disabled: isSaving, onPress: actions.onBack }], 'left');
  const saveOptions = useNativeHeaderActionOptions([{ kind: 'save', label: 'Save',
    disabled: !canSave, onPress: actions.onApply }]);
  const headerOptions = useMemo(() => ({ headerShown: true, headerBackVisible: false,
    ...cancelOptions, ...saveOptions }), [cancelOptions, saveOptions]);
  const Frame = Platform.OS === 'ios' ? View : AssetActionKeyboardFrame;
  return (
    <Frame style={styles.editor}>
      <Stack.Screen options={headerOptions} />
      <ScrollView style={{ flex: 1 }} contentInsetAdjustmentBehavior="automatic" automaticallyAdjustKeyboardInsets contentContainerStyle={styles.formScrollContent} keyboardDismissMode={appKeyboardDismissMode()} keyboardShouldPersistTaps="handled">
        {readOnly ? <ActionEligibilityNotice /> : null}
        {metadataRecovery}
        {isSaving ? <Text accessibilityLiveRegion="polite" style={styles.sheetSubtitle}>Saving changes…</Text> : null}
        <Text style={styles.sheetSubtitle}>
          {editContext.customTypeLabel ? `${editContext.kindLabel} · ${editContext.customTypeLabel}` : editContext.kindLabel}
        </Text>
        <Text style={styles.inputLabel}>Name</Text>
        <AppTextInput
          accessibilityLabel="Asset name"
          autoCapitalize="sentences"
          editable={!disabled}
          onChangeText={(title) => onChange({ ...draft, title, description: draft?.description ?? '', tagIds: draft?.tagIds ?? [], newTags: draft?.newTags ?? [] })}
          style={styles.input}
          value={draft?.title ?? ''}
        />
        <Text style={styles.inputLabel}>Description</Text>
        <AppTextInput
          accessibilityLabel="Description"
          editable={!disabled}
          multiline
          onChangeText={(description) => onChange({ ...draft, title: draft?.title ?? '', description, tagIds: draft?.tagIds ?? [], newTags: draft?.newTags ?? [] })}
          style={[styles.input, styles.multilineInput]}
          value={draft?.description ?? ''}
        />
        {assetTypes || !assetTypesFailed ? <AssetExpirationEditor asset={asset} draft={draft} types={assetTypes} disabled={disabled} onChange={onChange} /> : null}
        <EditTagPicker
          scope={JSON.stringify([asset.tenantId, asset.inventoryId, asset.id])}
          disabled={disabled}
          tags={assetTags}
          selectedTagIds={draft?.tagIds ?? []}
          newTags={draft?.newTags ?? []}
          onChange={(tagIds, newTags, inlineTag) => onChange({ ...draft, title: draft?.title ?? '', description: draft?.description ?? '', tagIds, newTags, inlineTag })}
        />
      </ScrollView>
    </Frame>
  );
}

function EditTagPicker({
  scope,
  disabled,
  newTags,
  onChange,
  selectedTagIds,
  tags,
}: {
  readonly scope: string;
  readonly disabled: boolean;
  readonly newTags: readonly CreateAssetTagDraft[];
  readonly onChange: (tagIds: readonly string[], newTags: readonly CreateAssetTagDraft[], entry: NonNullable<EditDraft['inlineTag']>) => void;
  readonly selectedTagIds: readonly string[];
  readonly tags: readonly AssetTagOptionViewModel[];
}) {
  return <AssetTagSelectionField scope={scope} disabled={disabled} tags={tags} selectedIds={selectedTagIds} newTags={newTags}
    onChange={(ids, pending) => onChange(ids, pending ?? newTags, { name: '', color: '' })} />;
}

export function MoveAssetSheet({
  readOnly = false,
  isCreatingDestination = false, asset,
  draft,
  candidatesAvailable = true,
  creationCandidatesAvailable = false,
  creationMatches = [],
  creationStatus,
  onBeginCreation,
  onCancelCreation,
  onChangeCreationName,
  candidateStatus,
  isSaving,
  onChangeQuery,
  onChangeCreateKind,
  onClose,
  onCreateDestination,
  onSave,
  onSelectParent,
  onSelectRoot
}: {
  readonly readOnly?: boolean;
  readonly asset: AssetDetailViewModel;
  readonly draft: MoveDraft | undefined;
  readonly candidatesAvailable?: boolean;
  readonly creationCandidatesAvailable?: boolean;
  readonly creationMatches?: readonly ParentLookupResult[];
  readonly creationStatus?: ReactNode;
  readonly onBeginCreation: () => void;
  readonly onCancelCreation: () => void;
  readonly onChangeCreationName: (name: string) => void;
  readonly candidateStatus?: MoveSelectionStatus;
  readonly isSaving: boolean;
  readonly isCreatingDestination?: boolean;
  readonly onChangeCreateKind: (kind: MoveDestinationCreateKind) => void;
  readonly onChangeQuery: (query: string) => void;
  readonly onClose: () => void;
  readonly onCreateDestination: () => void;
  readonly onSave: () => void;
  readonly onSelectParent: (parent: ParentLookupResult) => void;
  readonly onSelectRoot: () => void;
}) {
  const headerHeight = useHeaderHeight();
  const { styles: settingsStyles } = useSettingsListStyles();
  const creationExpanded = draft?.creationName !== undefined;
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  const disabled = isSaving || readOnly;
  const canSaveMove = draft ? canSaveMoveAsset(asset, draft.selectedParent) && !disabled : false;
  const placement = draft ? movePlacementPreview(asset, draft.selectedParent) : undefined;
  const createPlacement = moveDestinationCreatePlacement(asset);
  const createTitle = draft?.creationName?.trim() ?? '';
  const createKind = draft?.createKind ?? 'location';
  const canCreate = draft && creationCandidatesAvailable
    ? canCreateMoveDestination({
        kind: createKind,
        matches: creationMatches,
        parentAssetId: createPlacement.parentAssetId,
        query: createTitle
      })
    : false;
  const creationActions = useFocusedSheetActions({
    primaryLabel: 'Create destination', secondaryLabel: 'Cancel new destination',
    disabled: disabled || !creationExpanded || !canCreate,
    secondaryDisabled: disabled || !creationExpanded,
    onApply: onCreateDestination, onBack: onCancelCreation
  });
  const actions = useFocusedSheetActions({ primaryLabel: 'Move', secondaryLabel: 'Cancel',
    disabled: !canSaveMove, secondaryDisabled: isSaving, onApply: onSave, onBack: onClose });
  const cancelOptions = useNativeHeaderActionOptions([{ kind: 'close',
    label: creationExpanded ? 'Cancel new destination' : 'Cancel',
    disabled: isSaving, onPress: creationExpanded ? creationActions.onBack : actions.onBack }], 'left');
  const moveOptions = useNativeHeaderActionOptions(creationExpanded ? [{ kind: 'save', label: 'Create destination',
    emphasis: 'primary', disabled: creationActions.disabled, onPress: creationActions.onApply }] : [
    { kind: 'add', label: 'New destination', disabled: disabled || !candidatesAvailable, onPress: onBeginCreation },
    { kind: 'save', label: 'Move', emphasis: 'primary', disabled: !canSaveMove, onPress: actions.onApply }
  ]);
  const headerOptions = useMemo(() => ({ title: creationExpanded ? 'New destination' : 'Move',
    headerShown: true, headerBackVisible: false, ...cancelOptions, ...moveOptions }),
    [creationExpanded, cancelOptions, moveOptions]);
  const searchEnabled = !disabled && !creationExpanded;
  const search = { query: draft?.query ?? '', placeholder: 'Search places, boxes, shelves',
    onChange: onChangeQuery, onSubmit: onChangeQuery, onClear: () => onChangeQuery('') };
  function toRow(match: ParentLookupResult): MoveSelectionRowModel {
    return { id: match.id, label: match.title,
      context: match.disabledReason ?? `${match.kind === 'location' ? 'Location' : 'Container'} · ${match.pathLabel || match.title}`,
      kind: match.kind, selected: draft?.selectedParent?.id === match.id,
      disabled: disabled || match.canSelectAsParent === false,
      accessibilityLabel: `Choose destination ${match.title}`, onPress: () => onSelectParent(match) };
  }
  const Frame = Platform.OS === 'ios' ? View : AssetActionKeyboardFrame;
  return (
    <Frame style={[Platform.OS === 'ios' && !creationExpanded ? styles.nativeSelection : styles.editor,
      creationExpanded ? { backgroundColor: palette.background, paddingHorizontal: 0 } : undefined,
      Platform.OS === 'ios' && creationExpanded ? { paddingTop: headerHeight + spacing.sm } : undefined]}>
      <Stack.Screen options={headerOptions} />
      {Platform.OS === 'ios' ? <NativeNavigationSearch {...search} placement="stacked" enabled={searchEnabled} />
        : searchEnabled ? <NativeFilterSearch {...search} /> : null}
      {!creationExpanded ? <MoveSelectionList subjectLabel="Moving" subject={asset.title}
        context={`Current location: ${placement?.currentLocationLabel || 'Inventory root'}`} title="Destinations"
        destinationLabel={draft?.selectedParent === null ? 'Inventory root' : draft?.selectedParent?.pathLabel || draft?.selectedParent?.title || 'Choose a destination'}
        statuses={moveSelectionStatuses(readOnly, isSaving, candidateStatus)}
        rows={[{ id: 'inventory-root', label: 'Inventory root', context: 'Top level', kind: 'root',
          selected: draft?.selectedParent === null, disabled, accessibilityLabel: 'Choose inventory root', onPress: onSelectRoot },
          ...(draft?.matches ?? []).map(toRow)]}
        retainedSelection={draft?.selectedParent && !draft.matches.some(match => match.id === draft.selectedParent?.id)
          ? toRow(draft.selectedParent) : undefined} /> :
      <ScrollView style={{ flex: 1 }} contentInsetAdjustmentBehavior="never" automaticallyAdjustKeyboardInsets contentContainerStyle={styles.formScrollContent} keyboardDismissMode={appKeyboardDismissMode()} keyboardShouldPersistTaps="handled">
        {creationExpanded ? (
          <>
            <SettingsSection title="Name">
              <View style={settingsStyles.navigationRow}>
                <DraftTextField accessibilityLabel="New destination name" value={draft?.creationName ?? ''}
                  editable={!disabled} placeholder="Place or container name"
                  style={[settingsStyles.rowLabel, { minHeight: 48, flex: 1 }]}
                  onChangeText={name => { if (!disabled) onChangeCreationName(name); }} />
              </View>
            </SettingsSection>
            <SettingsSection title="Kind" footer={`${moveDestinationCreateKindHelp(createKind)} ${moveDestinationCreatePlacementLabel(createPlacement)}. The new destination will be selected for this move.`}>
              <View style={settingsStyles.navigationRow}>
                <NativeSegmentedControl colors={palette} style={{ flex: 1, height: 48 }}
                  value={createKind} segments={[{ value: 'location', label: 'Location' }, { value: 'container', label: 'Container' }]}
                  disabled={disabled}
                  onChange={value => { if (!disabled && (value === 'location' || value === 'container')) onChangeCreateKind(value); }} />
              </View>
            </SettingsSection>
            {creationStatus}
            {isCreatingDestination ? <Text accessibilityLiveRegion="polite" style={styles.sheetSubtitle}>Creating destination…</Text> : null}
          </>
        ) : null}
      </ScrollView>}
    </Frame>
  );
}

export function MoveThingsHereSheet({
  readOnly = false,
  draft,
  candidatesAvailable = true,
  candidateStatus,
  isSaving,
  onChangeQuery,
  onClose,
  onSave,
  onSelectAsset
}: {
  readonly readOnly?: boolean;
  readonly draft: MoveIntoDraft | undefined;
  readonly candidatesAvailable?: boolean;
  readonly candidateStatus?: MoveSelectionStatus;
  readonly isSaving: boolean;
  readonly onChangeQuery: (query: string) => void;
  readonly onClose: () => void;
  readonly onSave: () => void;
  readonly onSelectAsset: (asset: ParentLookupResult) => void;
}) {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  const disabled = isSaving || readOnly;
  const canSave = draft?.selectedAsset !== undefined && !disabled;
  const emptyState = moveIntoEmptyState(draft?.query ?? '');
  const actions = useFocusedSheetActions({ primaryLabel: 'Move here', secondaryLabel: 'Cancel',
    disabled: !canSave, secondaryDisabled: isSaving, onApply: onSave, onBack: onClose });
  const cancelOptions = useNativeHeaderActionOptions([{ kind: 'close', label: 'Cancel',
    disabled: isSaving, onPress: actions.onBack }], 'left');
  const moveOptions = useNativeHeaderActionOptions([{ kind: 'save', label: 'Move here',
    emphasis: 'primary', disabled: !canSave, onPress: actions.onApply }]);
  const headerOptions = useMemo(() => ({ headerShown: true, headerBackVisible: false,
    ...cancelOptions, ...moveOptions }), [cancelOptions, moveOptions]);
  const search = { query: draft?.query ?? '', placeholder: 'Search your inventory',
    onChange: onChangeQuery, onSubmit: onChangeQuery, onClear: () => onChangeQuery('') };
  function toRow(match: ParentLookupResult): MoveSelectionRowModel {
    const row = moveIntoCandidateRow(match);
    return { id: match.id, label: row.title, context: `${row.kindLabel} · ${row.pathLabel || row.title}`,
      kind: match.kind, selected: draft?.selectedAsset?.id === match.id, disabled,
      accessibilityLabel: `Choose item ${row.title}`, onPress: () => onSelectAsset(match) };
  }
  const Frame = Platform.OS === 'ios' ? View : AssetActionKeyboardFrame;
  return (
    <Frame style={Platform.OS === 'ios' ? styles.nativeSelection : styles.editor}>
      <Stack.Screen options={headerOptions} />
      {Platform.OS === 'ios' ? <NativeNavigationSearch {...search} placement="stacked" enabled={!disabled} />
        : !disabled ? <NativeFilterSearch {...search} /> : null}
      <MoveSelectionList subjectLabel="Destination" subject={draft?.target.title ?? 'This place'}
        context="Choose an item to move here." title="Items"
        statuses={[...moveSelectionStatuses(readOnly, isSaving, candidateStatus),
          ...(candidatesAvailable && draft?.matches.length === 0 ? [{ title: emptyState.title, message: emptyState.message }] : [])]}
        rows={(draft?.matches ?? []).map(toRow)}
        retainedSelection={draft?.selectedAsset && !draft.matches.some(match => match.id === draft.selectedAsset?.id)
          ? toRow(draft.selectedAsset) : undefined} />
    </Frame>
  );
}


function ActionEligibilityNotice() {
  const styles = useStyles();
  return <Text accessibilityRole="alert" style={styles.sheetSubtitle}>
    This item cannot be changed here. Your draft is kept while this screen is open.
  </Text>;
}


function useStyles() {
  return createStyles(useAppearancePalette());
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  nativeSelection: { flex: 1 },
  editor: {
    backgroundColor: colors.surface,
    flex: 1,
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.sm
  },
  sheetSubtitle: {
    color: colors.textMuted,
    fontSize: 14,
    lineHeight: 20
  },
  formScrollContent: {
    gap: spacing.sm,
    paddingBottom: spacing.sm
  },
  inputLabel: {
    color: colors.text,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0,
    marginTop: spacing.sm
  },
  input: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    color: colors.text,
    fontSize: 16,
    minHeight: 48,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  multilineInput: {
    minHeight: 104,
    textAlignVertical: 'top'
  },
  disabledAction: {
    opacity: 0.55
  }
  });
}

function moveSelectionStatuses(readOnly: boolean, isSaving: boolean, candidate?: MoveSelectionStatus): readonly MoveSelectionStatus[] {
  return [
    ...(readOnly ? [{ message: 'This item cannot be changed here. Your draft is kept while this screen is open.' }] : []),
    ...(candidate ? [{ ...candidate, retry: candidate.retry ? { ...candidate.retry, disabled: readOnly || isSaving } : undefined }] : []),
    ...(isSaving ? [{ message: 'Moving…' }] : [])
  ];
}
