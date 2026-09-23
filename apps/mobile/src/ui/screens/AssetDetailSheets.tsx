import { Stack } from 'expo-router';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';
import { DraftTextField } from '../components/DraftTextField';
import { useFocusedSheetActions } from '../components/useFocusedSheetActions';
import { AssetActionKeyboardFrame } from './AssetActionKeyboardFrame';
import { AssetTagSelectionField } from '../components/AssetTagSelectionField';
import { NativeSheetActions } from '../components/NativeSheetActions';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { NativeChoicePicker } from '../components/NativeChoicePicker';
import { AssetExpirationEditor } from '../components/AssetExpirationEditor';
import type { CustomAssetTypeDefinition } from '../../domain/customization/Customization';
import { useMemo, useState, type ReactNode } from 'react';
import {
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { AssetTagOptionViewModel } from '../../application/assets/InventoryAssetTagsQuery';
import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';
import {
  applyInlineAssetTagResolution,
  canApplyInlineAssetTagResolution,
  type CreateAssetTagDraft,
  resolveInlineAssetTag
} from '../../application/assets/AssetTagDraftResolution';
import { assetTagChipStylePresentation } from '../components/AssetTagChipsPresentation';
import { TagColorPicker } from '../components/TagColorPicker';
import { AppTextInput, appKeyboardDismissMode } from '../components/AppTextInput';
import {
  assetEditContext,
  canSaveEditAsset,
  hasUnstagedEditTag,
  EditDraft
} from './AssetDetailEditPresentation';
import {
  canCreateMoveDestination,
  canSaveMoveAsset,
  moveIntoCandidateRow,
  moveIntoEmptyState,
  moveDestinationRow,
  moveDestinationCreateButtonLabel,
  moveDestinationCreateKindHelp,
  moveDestinationCreatePlacement,
  moveDestinationCreatePlacementLabel,
  type MoveDestinationCreateKind,
  type MoveDestinationRow,
  movePlacementPreview
} from './AssetDetailMovePresentation';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { minimumTouchTargetSize, radius, spacing, type MobileColorPalette } from '../theme/tokens';

export type MoveDraft = {
  readonly creationName?: string;
  readonly queryRevision?: number;
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
          entry={draft?.inlineTag ?? { name: '', color: '' }}
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
  entry
}: {
  readonly scope: string;
  readonly disabled: boolean;
  readonly newTags: readonly CreateAssetTagDraft[];
  readonly entry: NonNullable<EditDraft['inlineTag']>;
  readonly onChange: (tagIds: readonly string[], newTags: readonly CreateAssetTagDraft[], entry: NonNullable<EditDraft['inlineTag']>) => void;
  readonly selectedTagIds: readonly string[];
  readonly tags: readonly AssetTagOptionViewModel[];
}) {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  const { name: newTagName, color: newTagColor } = entry;
  function setNewTagName(name: string): void { if (!disabled) onChange(selectedTagIds, newTags, { ...entry, name }); }
  function setNewTagColor(color: string): void { if (!disabled) onChange(selectedTagIds, newTags, { ...entry, color }); }
  const [creatingTag, setCreatingTag] = useState(false);
  const creationVisible = creatingTag || hasUnstagedEditTag(entry);
  const [tagEntryRevision, setTagEntryRevision] = useState(0);
  const creationActions = useFocusedSheetActions({
    primaryLabel: 'New tag', secondaryLabel: 'Cancel new tag',
    disabled: disabled || creationVisible, secondaryDisabled: disabled || !creationVisible,
    onApply: () => setCreatingTag(true),
    onBack: () => {
      setCreatingTag(false);
      setTagEntryRevision(current => current + 1);
      onChange(selectedTagIds, newTags, { name: '', color: '' });
    }
  });

  function addNewTag(): void {
    const displayName = newTagName.trim();
    if (disabled || displayName.length === 0) {
      return;
    }
    const transition = applyInlineAssetTagResolution({
      resolution: tagResolution,
      selectedTagIds,
      pendingTags: newTags
    });
    if (transition.shouldClearInputs) setTagEntryRevision(current => current + 1);
    onChange(transition.selectedTagIds, transition.pendingTags, transition.shouldClearInputs ? { name: '', color: '' } : entry);
  }

  const tagResolution = resolveInlineAssetTag({
    displayName: newTagName,
    color: newTagColor,
    activeTags: tags,
    pendingTags: newTags
  });
  const canAddNewTag = canApplyInlineAssetTagResolution(tagResolution);

  return (
    <View style={styles.tagPicker}>
      <Text style={styles.inputLabel}>Tags</Text>
      <AssetTagSelectionField scope={scope} disabled={disabled} tags={tags} selectedIds={selectedTagIds}
        onChange={ids => onChange(ids, newTags, entry)} />
      <View style={styles.tagOptions}>
        {newTags.map((tag, index) => {
          const colorStyle = assetTagChipStylePresentation(tag);
          return (
            <Pressable
              accessibilityRole="button"
              accessibilityLabel={`Remove new tag ${tag.displayName}`}
              accessibilityState={{ disabled }}
              disabled={disabled}
              key={`${tag.displayName}-${index.toString()}`}
              onPress={() => onChange(selectedTagIds, newTags.filter((_, currentIndex) => currentIndex !== index), entry)}
              style={[
                styles.tagOption,
                colorStyle.colored ? { backgroundColor: colorStyle.backgroundColor, borderColor: colorStyle.borderColor } : null,
                styles.tagOptionSelected,
                disabled ? styles.disabledAction : null
              ]}
            >
              <Text style={[styles.tagOptionText, styles.tagOptionTextSelected]} numberOfLines={1}>
                {tag.displayName}
              </Text>
            </Pressable>
          );
        })}
      </View>
      {creationVisible ? <>
        <View style={styles.newTagRow}>
          <View style={styles.newTagNameInput}>
            <DraftTextField
              key={Platform.OS === 'ios' ? tagEntryRevision : 'tag-name'}
              accessibilityLabel="New tag name"
              editable={!disabled}
              onChangeText={setNewTagName}
              placeholder="New tag"
              placeholderTextColor={palette.textMuted}
              style={styles.input}
              value={newTagName}
            />
          </View>
          <AppTextInput
            accessibilityLabel="New tag color"
            autoCapitalize="characters"
            editable={!disabled}
            onChangeText={setNewTagColor}
            placeholder="#2F80ED"
            placeholderTextColor={palette.textMuted}
            style={[styles.input, styles.newTagColorInput]}
            value={newTagColor}
          />
        </View>
        {tagResolution.status === 'display_name_too_long' ? <Text accessibilityRole="alert" style={styles.sheetSubtitle}>Use a shorter tag name.</Text> : null}
        <TagColorPicker disabled={disabled} palette={palette} value={newTagColor} onChange={setNewTagColor} />
        <NativeCommandButton label="Add tag" disabled={disabled || !canAddNewTag} onPress={addNewTag} />
        {hasUnstagedEditTag(entry) ? <Text style={styles.sheetSubtitle}>Add this tag or clear its name and color before saving.</Text> : null}
        <NativeCommandButton label="Cancel new tag" disabled={creationActions.secondaryDisabled} onPress={creationActions.onBack} />
      </> : <NativeCommandButton label="New tag" disabled={creationActions.disabled} onPress={creationActions.onApply} />}
    </View>
  );
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
  readonly candidateStatus?: ReactNode;
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
  const cancelOptions = useNativeHeaderActionOptions([{ kind: 'close', label: 'Cancel',
    disabled: isSaving, onPress: actions.onBack }], 'left');
  const moveOptions = useNativeHeaderActionOptions([{ kind: 'save', label: 'Move',
    disabled: !canSaveMove, onPress: actions.onApply }]);
  const headerOptions = useMemo(() => ({ headerShown: true, headerBackVisible: false,
    ...cancelOptions, ...moveOptions }), [cancelOptions, moveOptions]);
  const Frame = Platform.OS === 'ios' ? View : AssetActionKeyboardFrame;
  return (
    <Frame style={styles.editor}>
      <Stack.Screen options={headerOptions} />
      <ScrollView style={{ flex: 1 }} contentInsetAdjustmentBehavior="automatic" automaticallyAdjustKeyboardInsets contentContainerStyle={styles.formScrollContent} keyboardDismissMode={appKeyboardDismissMode()} keyboardShouldPersistTaps="handled">
        {readOnly ? <ActionEligibilityNotice /> : null}
        <Text style={styles.moveSubject}>{asset.title}</Text>
        {placement ? <Text style={styles.sheetSubtitle}>{`Current location: ${placement.currentLocationLabel || 'Inventory root'}`}</Text> : null}
        {placement?.hasChanged ? <Text style={styles.sheetSubtitle}>{`Selected: ${placement.proposedLocationLabel}`}</Text> : null}
        <Text style={styles.inputLabel}>Put in</Text>
        <DraftTextField
          key={Platform.OS === 'ios' ? draft?.queryRevision ?? 0 : 'move-query'}
          accessibilityLabel="Put in"
          editable={!disabled}
          onChangeText={query => { if (!disabled) { onCancelCreation(); onChangeQuery(query); } }}
          placeholder="Search places, boxes, shelves"
          placeholderTextColor={palette.textMuted}
          style={styles.input}
          value={draft?.query ?? ''}
        />
        {creationExpanded ? null : candidateStatus}
        {isSaving && !isCreatingDestination ? <Text accessibilityLiveRegion="polite" style={styles.sheetSubtitle}>Moving…</Text> : null}
        <ParentRow
          disabled={disabled}
          isSelected={draft?.selectedParent === null}
          row={{
            title: 'Inventory root',
            kindLabel: 'Top level',
            pathLabel: 'No containing location'
          }}
          onPress={onSelectRoot}
        />
        {draft?.matches.map((match) => (
          <ParentRow
            disabled={disabled}
            key={match.id}
            isSelected={draft.selectedParent?.id === match.id}
            row={moveDestinationRow(match)}
            onPress={() => onSelectParent(match)}
          />
        ))}
        {candidatesAvailable && !creationExpanded ? <NativeCommandButton label="New destination" disabled={disabled}
          onPress={() => { if (!disabled) onBeginCreation(); }} /> : null}
        {creationExpanded ? (
          <View style={styles.createDestinationPanel}>
            <Text style={styles.inputLabel}>Name</Text>
            <DraftTextField accessibilityLabel="New destination name" value={draft?.creationName ?? ''}
              editable={!disabled} placeholder="Place or container name" style={styles.input}
              onChangeText={name => { if (!disabled) onChangeCreationName(name); }} />
            {creationStatus}
            <NativeChoicePicker label="Kind" accessibilityLabel="Choose destination kind"
              value={createKind} options={[{ value: 'location', label: 'Location' }, { value: 'container', label: 'Container' }]}
              includeEmptyOption={false} disabled={disabled}
              onChange={value => { if (!disabled && (value === 'location' || value === 'container')) onChangeCreateKind(value); }} />
            <Text style={styles.createKindHelp}>{moveDestinationCreateKindHelp(createKind)}</Text>
            <Text style={styles.createPlacementText}>
              {moveDestinationCreatePlacementLabel(createPlacement)}
            </Text>
            <NativeCommandButton label={moveDestinationCreateButtonLabel(createKind, createTitle)}
              disabled={creationActions.disabled} onPress={creationActions.onApply} />
            <Text style={styles.parentSubtitle}>The new destination will be selected for this move.</Text>
            {isCreatingDestination ? <Text accessibilityLiveRegion="polite" style={styles.sheetSubtitle}>Creating destination…</Text> : null}
            <NativeCommandButton label="Cancel new destination" disabled={creationActions.secondaryDisabled} onPress={creationActions.onBack} />
          </View>
        ) : null}
      </ScrollView>
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
  readonly candidateStatus?: ReactNode;
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
  return (
    <AssetActionKeyboardFrame style={styles.sheet}>
      <ScrollView automaticallyAdjustKeyboardInsets contentContainerStyle={styles.formScrollContent} keyboardDismissMode={appKeyboardDismissMode()} keyboardShouldPersistTaps="handled">
        {Platform.OS !== 'android' ? <Text style={styles.sheetTitle}>Move something here</Text> : null}
        {readOnly ? <ActionEligibilityNotice /> : null}
        <Text style={styles.sheetSubtitle}>Choose an existing asset to put inside {draft?.target.title ?? 'this place'}.</Text>
        <Text style={styles.inputLabel}>Find item, box, or place</Text>
        <DraftTextField
          accessibilityLabel="Find item, box, or place"
          editable={!disabled}
          onChangeText={onChangeQuery}
          placeholder="Search your inventory"
          placeholderTextColor={palette.textMuted}
          style={styles.input}
          value={draft?.query ?? ''}
        />
        {candidateStatus}
        {candidatesAvailable && draft?.matches.length === 0 ? (
          <View style={styles.parentEmptyState}>
            <Text style={styles.parentTitle}>{emptyState.title}</Text>
            <Text style={styles.parentSubtitle}>{emptyState.message}</Text>
          </View>
        ) : null}
        {draft?.matches.map((match) => (
          <ParentRow
            disabled={disabled}
            key={match.id}
            isSelected={draft.selectedAsset?.id === match.id}
            row={moveIntoCandidateRow(match)}
            onPress={() => onSelectAsset(match)}
          />
        ))}
        <MovePreview
          left={draft?.selectedAsset?.title ?? 'Choose something'}
          right={draft?.target.title ?? 'Here'}
        />
      </ScrollView>
      <SheetActions
        busy={isSaving}
        disabled={!canSave}
        primaryLabel={isSaving ? 'Moving' : 'Move here'}
        onClose={onClose}
        onSave={onSave}
      />
    </AssetActionKeyboardFrame>
  );
}


function ActionEligibilityNotice() {
  const styles = useStyles();
  return <Text accessibilityRole="alert" style={styles.sheetSubtitle}>
    This item cannot be changed here. Your draft is kept while this screen is open.
  </Text>;
}

function ParentRow({
  disabled, isSelected,
  onPress,
  row
}: {
  readonly disabled: boolean;
  readonly isSelected: boolean;
  readonly onPress: () => void;
  readonly row: MoveDestinationRow;
}) {
  const styles = useStyles();
  return (
    <Pressable accessibilityRole="button" disabled={disabled} accessibilityState={{ disabled, selected: isSelected }} onPress={() => { if (!disabled) onPress(); }} style={[styles.parentRow, isSelected ? styles.parentRowSelected : null, disabled && styles.disabledAction]}>
      <View style={styles.parentTextColumn}>
        <View style={styles.parentTitleRow}>
          <Text style={styles.parentTitle}>{row.title}</Text>
          <Text style={styles.parentKindPill}>{row.kindLabel}</Text>
        </View>
        <Text style={styles.parentPath}>{row.pathLabel}</Text>
      </View>
      {isSelected ? <Text style={styles.parentSelected}>Selected</Text> : null}
    </Pressable>
  );
}

function MovePreview({ left, right }: { readonly left: string; readonly right: string }) {
  const styles = useStyles();
  return (
    <View style={styles.movePreview}>
      <Text style={styles.movePreviewLabel}>Move preview</Text>
      <Text style={styles.movePreviewText}>{left}{' -> '}{right}</Text>
    </View>
  );
}

function SheetActions({
  busy, disabled,
  onClose,
  onSave,
  primaryLabel
}: {
  readonly disabled: boolean;
  readonly onClose: () => void;
  readonly onSave: () => void;
  readonly primaryLabel: string;
  readonly busy: boolean;
}) {
  const actions = useFocusedSheetActions({
    primaryLabel, secondaryLabel: 'Cancel', disabled, secondaryDisabled: busy,
    onApply: onSave, onBack: onClose
  });
  return <NativeSheetActions keyboardAvoidance="container" {...actions} />;
}

function useStyles() {
  return createStyles(useAppearancePalette());
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  editor: {
    backgroundColor: colors.surface,
    flex: 1,
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.sm
  },
  sheet: {
    backgroundColor: colors.surface,
    flex: 1,
    gap: spacing.sm,
    padding: spacing.lg,
    paddingTop: spacing.xl
  },
  sheetTitle: {
    color: colors.text,
    fontSize: 26,
    fontWeight: '900',
    letterSpacing: 0
  },
  moveSubject: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '500'
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
  tagPicker: {
    gap: spacing.xs
  },
  tagOptions: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.xs
  },
  tagOption: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 999,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.xs,
    maxWidth: '100%',
    minHeight: minimumTouchTargetSize,
    minWidth: minimumTouchTargetSize,
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
    gap: spacing.xs
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

  createDestinationPanel: {
    backgroundColor: colors.brandDustyBlueSoft,
    borderRadius: radius.md,
    gap: spacing.sm,
    marginVertical: spacing.xs,
    padding: spacing.sm
  },
  createKindHelp: {
    color: colors.textMuted,
    fontSize: 13,
    lineHeight: 18,
    paddingHorizontal: spacing.xs
  },
  createPlacementText: {
    color: colors.accentStrong,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 18,
    paddingHorizontal: spacing.xs
  },
  parentRow: {
    alignItems: 'center',
    borderBottomColor: colors.border,
    borderBottomWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    justifyContent: 'space-between',
    minHeight: 64,
    paddingVertical: spacing.sm
  },
  parentRowSelected: {
    backgroundColor: colors.selected
  },
  parentEmptyState: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    gap: spacing.xs,
    marginVertical: spacing.xs,
    padding: spacing.md
  },
  parentTextColumn: {
    flex: 1,
    gap: spacing.xs,
    minWidth: 0
  },
  parentTitleRow: {
    alignItems: 'center',
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.xs
  },
  parentTitle: {
    color: colors.text,
    flexShrink: 1,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  parentKindPill: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    color: colors.accentStrong,
    fontSize: 11,
    fontWeight: '900',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.xs,
    paddingVertical: 3
  },
  parentPath: {
    color: colors.textMuted,
    fontSize: 13,
    lineHeight: 18
  },
  parentSubtitle: {
    color: colors.textMuted,
    fontSize: 13,
    lineHeight: 18
  },
  parentSelected: {
    color: colors.action,
    flexShrink: 0,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0
  },
  movePreview: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    padding: spacing.md
  },
  movePreviewLabel: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  movePreviewText: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0,
    marginTop: spacing.xs
  },
  disabledAction: {
    opacity: 0.55
  }
  });
}
