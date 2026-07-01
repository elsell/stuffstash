import {
  KeyboardAvoidingView,
  Modal,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View
} from 'react-native';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';
import {
  canSaveMoveAsset,
  movePlacementPreview,
  MovePlacementPreview
} from './AssetDetailMovePresentation';
import { colors, radius, spacing } from '../theme/tokens';

export type EditDraft = {
  readonly title: string;
  readonly description: string;
};

export type MoveDraft = {
  readonly query: string;
  readonly matches: readonly ParentLookupResult[];
  readonly selectedParent: ParentLookupResult | null;
};

export type MoveIntoDraft = {
  readonly target: AssetDetailViewModel;
  readonly query: string;
  readonly matches: readonly ParentLookupResult[];
  readonly selectedAsset?: ParentLookupResult;
};

export function EditAssetSheet({
  draft,
  isSaving,
  onChange,
  onClose,
  onSave
}: {
  readonly draft: EditDraft | undefined;
  readonly isSaving: boolean;
  readonly onChange: (draft: EditDraft) => void;
  readonly onClose: () => void;
  readonly onSave: () => void;
}) {
  return (
    <Modal animationType="slide" transparent visible={draft !== undefined} onRequestClose={onClose}>
      <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : undefined} style={styles.modalShell}>
        <View style={styles.sheet}>
          <View style={styles.sheetHandle} />
          <Text style={styles.sheetTitle}>Edit asset</Text>
          <Text style={styles.sheetSubtitle}>Kind is managed by Stuff Stash rules for now.</Text>
          <Text style={styles.inputLabel}>Name</Text>
          <TextInput
            autoCapitalize="sentences"
            editable={!isSaving}
            onChangeText={(title) => onChange({ title, description: draft?.description ?? '' })}
            style={styles.input}
            value={draft?.title ?? ''}
          />
          <Text style={styles.inputLabel}>Description</Text>
          <TextInput
            editable={!isSaving}
            multiline
            onChangeText={(description) => onChange({ title: draft?.title ?? '', description })}
            style={[styles.input, styles.multilineInput]}
            value={draft?.description ?? ''}
          />
          <SheetActions
            disabled={isSaving}
            primaryLabel={isSaving ? 'Saving' : 'Save'}
            onClose={onClose}
            onSave={onSave}
          />
        </View>
      </KeyboardAvoidingView>
    </Modal>
  );
}

export function MoveAssetSheet({
  asset,
  draft,
  isSaving,
  onChangeQuery,
  onClose,
  onCreateDestination,
  onSave,
  onSelectParent,
  onSelectRoot
}: {
  readonly asset: AssetDetailViewModel;
  readonly draft: MoveDraft | undefined;
  readonly isSaving: boolean;
  readonly onChangeQuery: (query: string) => void;
  readonly onClose: () => void;
  readonly onCreateDestination: () => void;
  readonly onSave: () => void;
  readonly onSelectParent: (parent: ParentLookupResult) => void;
  readonly onSelectRoot: () => void;
}) {
  const exactMatch = draft?.matches.some((match) => normalize(match.title) === normalize(draft.query)) ?? false;
  const canCreate = (draft?.query.trim().length ?? 0) > 0 && !exactMatch;
  const canSaveMove = draft ? canSaveMoveAsset(asset, draft.selectedParent) && !isSaving : false;
  const placement = draft ? movePlacementPreview(asset, draft.selectedParent) : undefined;
  return (
    <Modal animationType="slide" transparent visible={draft !== undefined} onRequestClose={onClose}>
      <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : undefined} style={styles.modalShell}>
        <View style={styles.sheet}>
          <View style={styles.sheetHandle} />
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
              <Pressable accessibilityRole="button" onPress={onCreateDestination} style={styles.parentCreateRow}>
                <Text style={styles.parentTitle}>Create new location "{draft?.query.trim()}"</Text>
                <Text style={styles.parentSubtitle}>Then select it as the destination</Text>
              </Pressable>
            ) : null}
            <ParentRow
              isSelected={draft?.selectedParent === null}
              subtitle="Top level"
              title="No parent"
              onPress={onSelectRoot}
            />
            {draft?.matches.map((match) => (
              <ParentRow
                key={match.id}
                isSelected={draft.selectedParent?.id === match.id}
                subtitle={`${match.selectionHint} / ${match.subtitle}`}
                title={match.title}
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
        </View>
      </KeyboardAvoidingView>
    </Modal>
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
  return (
    <Modal animationType="slide" transparent visible={draft !== undefined} onRequestClose={onClose}>
      <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : undefined} style={styles.modalShell}>
        <View style={styles.sheet}>
          <View style={styles.sheetHandle} />
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
            {draft?.matches.map((match) => (
              <ParentRow
                key={match.id}
                isSelected={draft.selectedAsset?.id === match.id}
                subtitle={`${match.selectionHint} / ${match.subtitle}`}
                title={match.title}
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
        </View>
      </KeyboardAvoidingView>
    </Modal>
  );
}

function ParentRow({
  isSelected,
  onPress,
  subtitle,
  title
}: {
  readonly isSelected: boolean;
  readonly onPress: () => void;
  readonly subtitle: string;
  readonly title: string;
}) {
  return (
    <Pressable accessibilityRole="button" onPress={onPress} style={[styles.parentRow, isSelected ? styles.parentRowSelected : null]}>
      <View>
        <Text style={styles.parentTitle}>{title}</Text>
        <Text style={styles.parentSubtitle}>{subtitle}</Text>
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

function normalize(value: string | undefined): string {
  return (value ?? '').trim().toLocaleLowerCase();
}

const styles = StyleSheet.create({
  modalShell: {
    backgroundColor: colors.scrim,
    flex: 1,
    justifyContent: 'flex-end'
  },
  sheet: {
    backgroundColor: colors.surface,
    borderTopLeftRadius: radius.lg,
    borderTopRightRadius: radius.lg,
    gap: spacing.sm,
    maxHeight: '88%',
    padding: spacing.lg
  },
  sheetHandle: {
    alignSelf: 'center',
    backgroundColor: colors.border,
    borderRadius: radius.sm,
    height: 5,
    marginBottom: spacing.xs,
    width: 44
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
    backgroundColor: colors.brandDustyBlueSoft,
    borderRadius: radius.md,
    marginVertical: spacing.xs,
    padding: spacing.md
  },
  parentTitle: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  parentSubtitle: {
    color: colors.textMuted,
    fontSize: 13,
    lineHeight: 18
  },
  parentSelected: {
    color: colors.action,
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
