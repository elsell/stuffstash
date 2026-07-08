import { useState } from 'react';
import {
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View
} from 'react-native';
import { Check } from 'lucide-react-native';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { AssetTagOptionViewModel } from '../../application/assets/InventoryAssetTagsQuery';
import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';
import {
  applyInlineAssetTagResolution,
  canResolveInlineAssetTag,
  type CreateAssetTagDraft,
  resolveInlineAssetTag
} from '../../application/assets/AssetTagDraftResolution';
import { assetTagChipStylePresentation } from '../components/AssetTagChipsPresentation';
import { TagColorPicker } from '../components/TagColorPicker';
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
  moveDestinationRow,
  moveDestinationCreateButtonLabel,
  moveDestinationCreateKindHelp,
  moveDestinationCreateKindLabel,
  moveDestinationCreatePlacement,
  moveDestinationCreatePlacementLabel,
  type MoveDestinationCreateKind,
  type MoveDestinationRow,
  movePlacementPreview,
  MovePlacementPreview
} from './AssetDetailMovePresentation';
import { colors, radius, spacing } from '../theme/tokens';

export type MoveDraft = {
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
  asset,
  assetTags,
  draft,
  isSaving,
  onChange,
  onClose,
  onSave
}: {
  readonly asset: AssetDetailViewModel;
  readonly assetTags: readonly AssetTagOptionViewModel[];
  readonly draft: EditDraft | undefined;
  readonly isSaving: boolean;
  readonly onChange: (draft: EditDraft) => void;
  readonly onClose: () => void;
  readonly onSave: () => void;
}) {
  const editContext = assetEditContext(asset);
  const canSave = canSaveEditAsset(asset, draft) && !isSaving;
  return (
    <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : undefined} style={styles.sheet}>
      <Text style={styles.sheetTitle}>Edit asset</Text>
      <ScrollView contentContainerStyle={styles.editScrollContent} keyboardShouldPersistTaps="handled">
        <View style={styles.readOnlyContextPanel}>
          <Text style={styles.readOnlyContextLabel}>Kind</Text>
          <Text style={styles.readOnlyContextValue}>
            {editContext.customTypeLabel
              ? `${editContext.kindLabel} / ${editContext.customTypeLabel}`
              : editContext.kindLabel}
          </Text>
          <Text style={styles.readOnlyContextHelp}>{editContext.helperText}</Text>
        </View>
        <Text style={styles.inputLabel}>Name</Text>
        <TextInput
          autoCapitalize="sentences"
          editable={!isSaving}
          onChangeText={(title) => onChange({ title, description: draft?.description ?? '', tagIds: draft?.tagIds ?? [], newTags: draft?.newTags ?? [] })}
          style={styles.input}
          value={draft?.title ?? ''}
        />
        <Text style={styles.inputLabel}>Description</Text>
        <TextInput
          editable={!isSaving}
          multiline
          onChangeText={(description) => onChange({ title: draft?.title ?? '', description, tagIds: draft?.tagIds ?? [], newTags: draft?.newTags ?? [] })}
          style={[styles.input, styles.multilineInput]}
          value={draft?.description ?? ''}
        />
        <EditTagPicker
          disabled={isSaving}
          tags={assetTags}
          selectedTagIds={draft?.tagIds ?? []}
          newTags={draft?.newTags ?? []}
          onChange={(tagIds) => onChange({ title: draft?.title ?? '', description: draft?.description ?? '', tagIds, newTags: draft?.newTags ?? [] })}
          onNewTagsChange={(newTags) => onChange({ title: draft?.title ?? '', description: draft?.description ?? '', tagIds: draft?.tagIds ?? [], newTags })}
        />
      </ScrollView>
      <SheetActions
        disabled={!canSave}
        primaryLabel={isSaving ? 'Saving' : 'Save'}
        onClose={onClose}
        onSave={onSave}
      />
    </KeyboardAvoidingView>
  );
}

function EditTagPicker({
  disabled,
  newTags,
  onChange,
  onNewTagsChange,
  selectedTagIds,
  tags
}: {
  readonly disabled: boolean;
  readonly newTags: readonly CreateAssetTagDraft[];
  readonly onChange: (tagIds: readonly string[]) => void;
  readonly onNewTagsChange: (tags: readonly CreateAssetTagDraft[]) => void;
  readonly selectedTagIds: readonly string[];
  readonly tags: readonly AssetTagOptionViewModel[];
}) {
  const [newTagName, setNewTagName] = useState('');
  const [newTagColor, setNewTagColor] = useState('');
  const selected = new Set(selectedTagIds);

  function toggleTag(tagId: string): void {
    if (disabled) {
      return;
    }
    if (selected.has(tagId)) {
      onChange(selectedTagIds.filter((current) => current !== tagId));
      return;
    }
    onChange([...selectedTagIds, tagId]);
  }

  function addNewTag(): void {
    const displayName = newTagName.trim();
    if (disabled || displayName.length === 0) {
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
      <Text style={styles.inputLabel}>Tags</Text>
      <View style={styles.tagOptions}>
        {newTags.map((tag, index) => {
          const colorStyle = assetTagChipStylePresentation(tag);
          return (
            <Pressable
              accessibilityRole="button"
              accessibilityState={{ disabled, selected: true }}
              disabled={disabled}
              key={`${tag.displayName}-${index.toString()}`}
              onPress={() => onNewTagsChange(newTags.filter((_, currentIndex) => currentIndex !== index))}
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
        {tags.map((tag) => {
          const isSelected = selected.has(tag.id);
          const colorStyle = assetTagChipStylePresentation(tag);
          return (
            <Pressable
              accessibilityRole="button"
              accessibilityState={{ disabled, selected: isSelected }}
              disabled={disabled}
              key={tag.id}
              onPress={() => toggleTag(tag.id)}
              style={[
                styles.tagOption,
                colorStyle.colored ? { backgroundColor: colorStyle.backgroundColor, borderColor: colorStyle.borderColor } : null,
                isSelected ? styles.tagOptionSelected : null,
                disabled ? styles.disabledAction : null
              ]}
            >
              <Text style={[styles.tagOptionText, isSelected ? styles.tagOptionTextSelected : null]} numberOfLines={1}>
                {tag.label}
              </Text>
              {isSelected ? <Check color={colors.action} size={14} strokeWidth={2.4} /> : null}
            </Pressable>
          );
        })}
      </View>
      <View style={styles.newTagRow}>
        <TextInput
          accessibilityLabel="New tag name"
          editable={!disabled}
          onChangeText={setNewTagName}
          placeholder="New tag"
          placeholderTextColor={colors.textMuted}
          style={[styles.input, styles.newTagNameInput]}
          value={newTagName}
        />
        <TextInput
          accessibilityLabel="New tag color"
          autoCapitalize="characters"
          editable={!disabled}
          onChangeText={setNewTagColor}
          placeholder="#2F80ED"
          placeholderTextColor={colors.textMuted}
          style={[styles.input, styles.newTagColorInput]}
          value={newTagColor}
        />
        <Pressable
          accessibilityRole="button"
          disabled={disabled || !canAddNewTag}
          onPress={addNewTag}
          style={[styles.newTagButton, disabled || !canAddNewTag ? styles.disabledAction : null]}
        >
          <Text style={styles.newTagButtonText}>Add</Text>
        </Pressable>
      </View>
      <TagColorPicker disabled={disabled} value={newTagColor} onChange={setNewTagColor} />
    </View>
  );
}

export function MoveAssetSheet({
  asset,
  draft,
  isSaving,
  onChangeQuery,
  onChangeCreateKind,
  onClose,
  onCreateDestination,
  onSave,
  onSelectParent,
  onSelectRoot
}: {
  readonly asset: AssetDetailViewModel;
  readonly draft: MoveDraft | undefined;
  readonly isSaving: boolean;
  readonly onChangeCreateKind: (kind: MoveDestinationCreateKind) => void;
  readonly onChangeQuery: (query: string) => void;
  readonly onClose: () => void;
  readonly onCreateDestination: () => void;
  readonly onSave: () => void;
  readonly onSelectParent: (parent: ParentLookupResult) => void;
  readonly onSelectRoot: () => void;
}) {
  const canSaveMove = draft ? canSaveMoveAsset(asset, draft.selectedParent) && !isSaving : false;
  const placement = draft ? movePlacementPreview(asset, draft.selectedParent) : undefined;
  const createPlacement = moveDestinationCreatePlacement(asset);
  const createTitle = draft?.query.trim() ?? '';
  const createKind = draft?.createKind ?? 'location';
  const canCreate = draft
    ? canCreateMoveDestination({
        kind: createKind,
        matches: draft.matches,
        parentAssetId: createPlacement.parentAssetId,
        query: draft.query
      }) && !isSaving
    : false;
  return (
    <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : undefined} style={styles.sheet}>
      <Text style={styles.sheetTitle}>Move {asset.title}</Text>
      <Text style={styles.sheetSubtitle}>Choose the place, box, shelf, or top level where this belongs.</Text>
      {placement ? <PlacementPanel preview={placement} /> : null}
      <Text style={styles.inputLabel}>Put in</Text>
      <TextInput
        autoCapitalize="sentences"
        editable={!isSaving}
        onChangeText={onChangeQuery}
        placeholder="Search places, boxes, shelves"
        placeholderTextColor={colors.textMuted}
        style={styles.input}
        value={draft?.query ?? ''}
      />
      <ScrollView style={styles.parentList} keyboardShouldPersistTaps="handled">
        {canCreate ? (
          <View style={styles.createDestinationPanel}>
            <View style={styles.createKindSegment} accessibilityRole="tablist">
              <CreateKindOption
                kind="location"
                disabled={isSaving}
                selectedKind={createKind}
                onPress={onChangeCreateKind}
              />
              <CreateKindOption
                kind="container"
                disabled={isSaving}
                selectedKind={createKind}
                onPress={onChangeCreateKind}
              />
            </View>
            <Text style={styles.createKindHelp}>{moveDestinationCreateKindHelp(createKind)}</Text>
            <Text style={styles.createPlacementText}>
              {moveDestinationCreatePlacementLabel(createPlacement)}
            </Text>
            <Pressable
              accessibilityRole="button"
              accessibilityState={{ disabled: isSaving }}
              disabled={isSaving}
              onPress={onCreateDestination}
              style={[styles.parentCreateRow, isSaving ? styles.disabledAction : null]}
            >
              <Text style={styles.parentTitle}>{moveDestinationCreateButtonLabel(createKind, createTitle)}</Text>
              <Text style={styles.parentSubtitle}>Then select it as the destination</Text>
            </Pressable>
          </View>
        ) : null}
        <ParentRow
          isSelected={draft?.selectedParent === null}
          row={{
            title: 'No parent',
            kindLabel: 'Top level',
            pathLabel: 'Inventory root'
          }}
          onPress={onSelectRoot}
        />
        {draft?.matches.map((match) => (
          <ParentRow
            key={match.id}
            isSelected={draft.selectedParent?.id === match.id}
            row={moveDestinationRow(match)}
            onPress={() => onSelectParent(match)}
          />
        ))}
      </ScrollView>
      <SheetActions
        disabled={!canSaveMove}
        primaryLabel={isSaving ? 'Moving' : 'Move'}
        onClose={onClose}
        onSave={onSave}
      />
    </KeyboardAvoidingView>
  );
}

export function MoveThingsHereSheet({
  draft,
  isSaving,
  onChangeQuery,
  onClose,
  onSave,
  onSelectAsset
}: {
  readonly draft: MoveIntoDraft | undefined;
  readonly isSaving: boolean;
  readonly onChangeQuery: (query: string) => void;
  readonly onClose: () => void;
  readonly onSave: () => void;
  readonly onSelectAsset: (asset: ParentLookupResult) => void;
}) {
  const canSave = draft?.selectedAsset !== undefined && !isSaving;
  const emptyState = moveIntoEmptyState(draft?.query ?? '');
  return (
    <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : undefined} style={styles.sheet}>
      <Text style={styles.sheetTitle}>Move something here</Text>
      <Text style={styles.sheetSubtitle}>Choose an existing asset to put inside {draft?.target.title ?? 'this place'}.</Text>
      <Text style={styles.inputLabel}>Find item, box, or place</Text>
      <TextInput
        autoCapitalize="sentences"
        editable={!isSaving}
        onChangeText={onChangeQuery}
        placeholder="Search your inventory"
        placeholderTextColor={colors.textMuted}
        style={styles.input}
        value={draft?.query ?? ''}
      />
      <ScrollView style={styles.parentList} keyboardShouldPersistTaps="handled">
        {draft?.matches.length === 0 ? (
          <View style={styles.parentEmptyState}>
            <Text style={styles.parentTitle}>{emptyState.title}</Text>
            <Text style={styles.parentSubtitle}>{emptyState.message}</Text>
          </View>
        ) : null}
        {draft?.matches.map((match) => (
          <ParentRow
            key={match.id}
            isSelected={draft.selectedAsset?.id === match.id}
            row={moveIntoCandidateRow(match)}
            onPress={() => onSelectAsset(match)}
          />
        ))}
      </ScrollView>
      <MovePreview
        left={draft?.selectedAsset?.title ?? 'Choose something'}
        right={draft?.target.title ?? 'Here'}
      />
      <SheetActions
        disabled={!canSave}
        primaryLabel={isSaving ? 'Moving' : 'Move here'}
        onClose={onClose}
        onSave={onSave}
      />
    </KeyboardAvoidingView>
  );
}

function CreateKindOption({
  disabled,
  kind,
  onPress,
  selectedKind
}: {
  readonly disabled: boolean;
  readonly kind: MoveDestinationCreateKind;
  readonly onPress: (kind: MoveDestinationCreateKind) => void;
  readonly selectedKind: MoveDestinationCreateKind;
}) {
  const isSelected = kind === selectedKind;
  return (
    <Pressable
      accessibilityRole="tab"
      accessibilityState={{ disabled, selected: isSelected }}
      disabled={disabled}
      onPress={() => onPress(kind)}
      style={[styles.createKindOption, isSelected ? styles.createKindOptionSelected : null, disabled ? styles.disabledAction : null]}
    >
      <Text style={[styles.createKindOptionText, isSelected ? styles.createKindOptionTextSelected : null]}>
        {moveDestinationCreateKindLabel(kind)}
      </Text>
    </Pressable>
  );
}

function ParentRow({
  isSelected,
  onPress,
  row
}: {
  readonly isSelected: boolean;
  readonly onPress: () => void;
  readonly row: MoveDestinationRow;
}) {
  return (
    <Pressable accessibilityRole="button" onPress={onPress} style={[styles.parentRow, isSelected ? styles.parentRowSelected : null]}>
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

function PlacementPanel({ preview }: { readonly preview: MovePlacementPreview }) {
  return (
    <View style={styles.placementPanel}>
      <PlacementRow label="From" value={preview.currentLocationLabel} />
      <PlacementRow label="To" value={preview.proposedLocationLabel} isEmphasized={preview.hasChanged} />
    </View>
  );
}

function PlacementRow({
  isEmphasized = false,
  label,
  value
}: {
  readonly isEmphasized?: boolean;
  readonly label: string;
  readonly value: string;
}) {
  return (
    <View style={styles.placementRow}>
      <Text style={styles.placementLabel}>{label}</Text>
      <Text style={[styles.placementValue, isEmphasized ? styles.placementValueEmphasized : null]}>{value}</Text>
    </View>
  );
}

function MovePreview({ left, right }: { readonly left: string; readonly right: string }) {
  return (
    <View style={styles.movePreview}>
      <Text style={styles.movePreviewLabel}>Move preview</Text>
      <Text style={styles.movePreviewText}>{left}{' -> '}{right}</Text>
    </View>
  );
}

function SheetActions({
  disabled,
  onClose,
  onSave,
  primaryLabel
}: {
  readonly disabled: boolean;
  readonly onClose: () => void;
  readonly onSave: () => void;
  readonly primaryLabel: string;
}) {
  return (
    <View style={styles.sheetActions}>
      <Pressable accessibilityRole="button" onPress={onClose} style={styles.sheetSecondary}>
        <Text style={styles.sheetSecondaryText}>Cancel</Text>
      </Pressable>
      <Pressable
        accessibilityRole="button"
        disabled={disabled}
        onPress={onSave}
        style={[styles.sheetPrimary, disabled ? styles.disabledAction : null]}
      >
        <Text style={styles.sheetPrimaryText}>{primaryLabel}</Text>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
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
  sheetSubtitle: {
    color: colors.textMuted,
    fontSize: 14,
    lineHeight: 20
  },
  editScrollContent: {
    gap: spacing.sm,
    paddingBottom: spacing.sm
  },
  readOnlyContextPanel: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    gap: spacing.xs,
    padding: spacing.md
  },
  readOnlyContextLabel: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  readOnlyContextValue: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  readOnlyContextHelp: {
    color: colors.textMuted,
    fontSize: 13,
    lineHeight: 18
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
    minHeight: 34,
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
  sheetActions: {
    flexDirection: 'row',
    gap: spacing.sm,
    marginTop: spacing.md
  },
  sheetPrimary: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    flex: 1,
    justifyContent: 'center',
    minHeight: 48
  },
  sheetPrimaryText: {
    color: colors.onAction,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  sheetSecondary: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flex: 1,
    justifyContent: 'center',
    minHeight: 48
  },
  sheetSecondaryText: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  parentList: {
    maxHeight: 280
  },
  createDestinationPanel: {
    backgroundColor: colors.brandDustyBlueSoft,
    borderRadius: radius.md,
    gap: spacing.sm,
    marginVertical: spacing.xs,
    padding: spacing.sm
  },
  createKindSegment: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.xs,
    padding: 4
  },
  createKindOption: {
    alignItems: 'center',
    borderRadius: radius.sm,
    flex: 1,
    justifyContent: 'center',
    minHeight: 36
  },
  createKindOptionSelected: {
    backgroundColor: colors.action
  },
  createKindOptionText: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0
  },
  createKindOptionTextSelected: {
    color: colors.onAction
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
  parentCreateRow: {
    backgroundColor: colors.surface,
    borderRadius: radius.md,
    padding: spacing.md
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
  placementPanel: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    gap: spacing.sm,
    padding: spacing.md
  },
  placementRow: {
    gap: 2
  },
  placementLabel: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  placementValue: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '700',
    letterSpacing: 0,
    lineHeight: 21
  },
  placementValueEmphasized: {
    color: colors.action,
    fontWeight: '900'
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
