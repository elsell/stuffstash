import { useMemo, useState } from 'react';
import { Text } from 'react-native';
import { Stack } from 'expo-router';
import { NativeFilterSheet } from '../components/NativeFilterSheet';
import { NativeSegmentedControl } from '../components/NativeSegmentedControl';
import { SettingsChoiceRow, SettingsSection, useSettingsListStyles } from './SettingsList';

export type AssetTagSelectionOption = { readonly id: string; readonly label: string };

/** One selection visit. Its owner remounts this screen for a new draft or scope. */
export function AssetTagSelectionScreen({ tags, initialSelectedIds, available = true, onDone, onCancel }: {
  readonly tags: readonly AssetTagSelectionOption[];
  readonly initialSelectedIds: readonly string[];
  readonly available?: boolean;
  readonly onDone: (ids: readonly string[]) => void;
  readonly onCancel: () => void;
}) {
  const { palette, styles } = useSettingsListStyles();
  const [selected, setSelected] = useState(() => [...new Set(initialSelectedIds)]);
  const [query, setQuery] = useState('');
  const [view, setView] = useState<'all' | 'selected'>('all');
  const ordered = useMemo(() => [...tags].sort((a, b) => a.label.localeCompare(b.label, undefined, { numeric: true, sensitivity: 'base' })), [tags]);
  const visible = ordered.filter(tag => view === 'selected' ? selected.includes(tag.id) : tag.label.toLocaleLowerCase().includes(query.trim().toLocaleLowerCase()));
  const unavailable = view === 'selected' ? selected.filter(id => !tags.some(tag => tag.id === id)) : [];
  const empty = view === 'selected' ? 'No tags selected' : tags.length === 0 ? 'No tags available' : 'No matching tags';
  return <>
    <Stack.Screen options={{ title: 'Tags' }} />
    <NativeFilterSheet title="Tags" footerTestID="asset-tag-selection-actions"
      search={available && view === 'all' ? { query, placeholder: 'Search tags', onChange: setQuery, onSubmit: setQuery, onClear: () => setQuery('') } : undefined}
      actions={{ primaryLabel: 'Done', primaryAccessibilityLabel: 'Done selecting tags', secondaryLabel: 'Cancel', secondaryAccessibilityLabel: 'Cancel selecting tags', disabled: !available, onApply: () => onDone(selected), onBack: onCancel }}>
      <SettingsSection>
        <NativeSegmentedControl colors={palette} disabled={!available} value={view} onChange={setView}
          segments={[{ value: 'all', label: 'All tags' }, { value: 'selected', label: 'Selected' }]} />
        <Text accessibilityLiveRegion="polite" style={styles.rowContext}>{selected.length} selected</Text>
      </SettingsSection>
      {!available ? <SettingsSection><Text accessibilityRole="alert" style={styles.rowContext}>Tag selection is no longer available for this draft.</Text></SettingsSection> : (
        <SettingsSection footer={visible.length || unavailable.length ? undefined : empty}>
          {visible.map(tag => <SettingsChoiceRow key={tag.id} multiple label={tag.label}
            accessibilityLabel={`Select tag ${tag.label}`} selected={selected.includes(tag.id)}
            onPress={() => setSelected(current => current.includes(tag.id) ? current.filter(id => id !== tag.id) : [...current, tag.id])} />)}
          {unavailable.map((id, index) => <SettingsChoiceRow key={id} multiple selected
            label="Unavailable tag" context="Not in the current tag list"
            accessibilityLabel={`Remove unavailable tag ${index + 1}`}
            onPress={() => setSelected(current => current.filter(selectedId => selectedId !== id))} />)}
        </SettingsSection>
      )}
    </NativeFilterSheet>
  </>;
}
