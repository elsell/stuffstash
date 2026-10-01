import { t } from '../../presentation/localization';
import type { CreateAssetTagDraft } from '../../application/assets/AssetTagDraftResolution';
import { NewAssetTagScreen } from './NewAssetTagScreen';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { useMemo, useState } from 'react';
import { Text } from 'react-native';
import { Stack } from 'expo-router';
import { NativeFilterSheet } from '../components/NativeFilterSheet';
import { NativeSegmentedControl } from '../components/NativeSegmentedControl';
import { SettingsChoiceRow, SettingsSection, useSettingsListStyles } from './SettingsList';

export type AssetTagSelectionOption = { readonly id: string; readonly label: string; readonly key?: string };

/** One selection visit. Its owner remounts this screen for a new draft or scope. */
export function AssetTagSelectionScreen({ tags, initialSelectedIds, initialNewTags, available = true, onDone, onCancel }: {
  readonly tags: readonly AssetTagSelectionOption[];
  readonly initialSelectedIds: readonly string[];
  readonly initialNewTags?: readonly CreateAssetTagDraft[];
  readonly available?: boolean;
  readonly onDone: (ids: readonly string[], newTags?: readonly CreateAssetTagDraft[]) => void;
  readonly onCancel: () => void;
}) {
  const { palette, styles } = useSettingsListStyles();
  const [selected, setSelected] = useState(() => [...new Set(initialSelectedIds)]);
  const [newTags, setNewTags] = useState<readonly CreateAssetTagDraft[]>(initialNewTags ?? []);
  const [creating, setCreating] = useState(false);
  const [query, setQuery] = useState('');
  const [view, setView] = useState<'all' | 'selected'>('all');
  const ordered = useMemo(() => [...tags].sort((a, b) => a.label.localeCompare(b.label, undefined, { numeric: true, sensitivity: 'base' })), [tags]);
  const visible = ordered.filter(tag => view === 'selected' ? selected.includes(tag.id) : tag.label.toLocaleLowerCase().includes(query.trim().toLocaleLowerCase()));
  const unavailable = view === 'selected' ? selected.filter(id => !tags.some(tag => tag.id === id)) : [];
  const empty = view === 'selected' ? t('tags.noneSelected') : tags.length === 0 ? t('tags.noneAvailable') : t('tags.noneMatching');
  if (creating) return <NewAssetTagScreen tags={tags} selectedIds={selected} newTags={newTags} available={available}
    onCancel={() => setCreating(false)} onDone={(ids, pending) => { setSelected([...ids]); setNewTags(pending); setCreating(false); }} />;
  return <>
    <Stack.Screen options={{ title: t('mobile.AssetTagSelectionScreen.tags') }} />
    <NativeFilterSheet title={t('mobile.AssetTagSelectionScreen.tags')} footerTestID="asset-tag-selection-actions"
      search={available && view === 'all' ? { query, placeholder: t('mobile.AssetTagSelectionScreen.searchTags'), onChange: setQuery, onSubmit: setQuery, onClear: () => setQuery('') } : undefined}
      actions={{ primaryLabel: t('mobile.AssetTagSelectionScreen.done'), primaryAccessibilityLabel: t('mobile.AssetTagSelectionScreen.doneSelectingTags'), secondaryLabel: t('mobile.AssetTagSelectionScreen.cancel'), secondaryAccessibilityLabel: t('mobile.AssetTagSelectionScreen.cancelSelectingTags'), disabled: !available, onApply: () => onDone(selected, initialNewTags === undefined ? undefined : newTags), onBack: onCancel }}>
      <SettingsSection>
        <NativeSegmentedControl colors={palette} disabled={!available} value={view} onChange={setView}
          segments={[{ value: 'all', label: t('mobile.AssetTagSelectionScreen.allTags') }, { value: 'selected', label: t('mobile.AssetTagSelectionScreen.selected') }]} />
        <Text accessibilityLiveRegion="polite" style={styles.rowContext}>{selected.length + newTags.length}{t('mobile.AssetTagSelectionScreen.selected2')}</Text>
      </SettingsSection>
      {initialNewTags !== undefined && available ? <SettingsSection>
        <NativeCommandButton label={t('mobile.AssetTagSelectionScreen.newTag')} onPress={() => setCreating(true)} />
        {newTags.map((tag, index) => <SettingsChoiceRow key={tag.displayName} multiple selected label={tag.displayName}
          context={t('mobile.AssetTagSelectionScreen.newTag')} accessibilityLabel={t('mobile.AssetTagSelectionScreen.removeNewTag', { displayName: String(tag.displayName) })}
          onPress={() => setNewTags(current => current.filter((_, currentIndex) => currentIndex !== index))} />)}
      </SettingsSection> : null}
      {!available ? <SettingsSection><Text accessibilityRole="alert" style={styles.rowContext}>{t('mobile.AssetTagSelectionScreen.tagSelectionIsNoLongerAvailableForThisDraft')}</Text></SettingsSection> : (
        <SettingsSection footer={visible.length || unavailable.length ? undefined : empty}>
          {visible.map(tag => <SettingsChoiceRow key={tag.id} multiple label={tag.label}
            accessibilityLabel={t('mobile.AssetTagSelectionScreen.selectTag', { label: String(tag.label) })} selected={selected.includes(tag.id)}
            onPress={() => setSelected(current => current.includes(tag.id) ? current.filter(id => id !== tag.id) : [...current, tag.id])} />)}
          {unavailable.map((id, index) => <SettingsChoiceRow key={id} multiple selected
            label={t('mobile.AssetTagSelectionScreen.unavailableTag')} context={t('mobile.AssetTagSelectionScreen.notInTheCurrentTagList')}
            accessibilityLabel={t('mobile.AssetTagSelectionScreen.removeUnavailableTag', { value: String(index + 1) })}
            onPress={() => setSelected(current => current.filter(selectedId => selectedId !== id))} />)}
        </SettingsSection>
      )}
    </NativeFilterSheet>
  </>;
}
