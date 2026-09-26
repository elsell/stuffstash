import { SettingsPickerRow } from '../components/SettingsPickerRow';
import { useState } from 'react';
import { Text, View } from 'react-native';
import { NativeActionMenu } from '../components/NativeActionMenu';
import { Stack } from 'expo-router';
import type { AssetTagOptionViewModel } from '../../application/assets/InventoryAssetTagsQuery';
import type { AssetBrowseSort } from '../../application/home/InventorySummaryRepository';
import type { BrowseDraftFilters } from './BrowseFilterState';
import { SettingsActionRow, SettingsChoiceRow, SettingsNavigationRow, SettingsSection, SettingsValueRow, useSettingsListStyles } from './SettingsList';
import { NativeFilterSheet } from '../components/NativeFilterSheet';
import type { ExpirationMode } from '../../application/expiration/ExpirationRepository';

export type BrowseFilterDraft = BrowseDraftFilters & { readonly sort: AssetBrowseSort };
type Page = 'overview' | 'tags';
const defaults: BrowseFilterDraft = { scope: 'all', lifecycleState: 'active', checkoutState: 'any', tagIds: [], sort: 'updated_desc' };
const choices = {
  scope: [{ value: 'all', label: 'All types' }, { value: 'places', label: 'Places' }, { value: 'containers', label: 'Containers' }, { value: 'items', label: 'Items' }],
  lifecycleState: [{ value: 'active', label: 'Active' }, { value: 'archived', label: 'Archived' }, { value: 'all', label: 'All statuses' }],
  checkoutState: [{ value: 'any', label: 'Any availability' }, { value: 'available', label: 'Available' }, { value: 'checked_out', label: 'Checked out' }],
  sort: [{ value: 'updated_desc', label: 'Recently changed' }, { value: 'id_asc', label: 'Default order' }]
} as const;
const titles: Record<Page, string> = { overview: 'Filters', tags: 'Tags' };

export function BrowseFiltersScreen({ initial, query, tags, busy = false, error, onApply, onCancel, onCancelPending, onExpiration }: {
  readonly initial: BrowseFilterDraft; readonly query: string; readonly tags: readonly AssetTagOptionViewModel[];
  readonly error?: string;
  readonly busy?: boolean; readonly onApply: (draft: BrowseFilterDraft) => void; readonly onCancel: () => void;
  readonly onCancelPending?: () => void;
  readonly onExpiration: (mode: ExpirationMode, draft: BrowseFilterDraft) => void;
}) {
  const { styles } = useSettingsListStyles();
  const [draft, setDraft] = useState(initial);
  const [page, setPage] = useState<Page>('overview');
  const [search, setSearch] = useState('');
  const open = (next: Page) => { setSearch(''); setPage(next); };
  const searchMode = !!query.trim() || draft.tagIds.length > 0;
  const visibleTags = [...tags].sort((a, b) => a.label.localeCompare(b.label))
    .filter(tag => tag.label.toLocaleLowerCase().includes(search.toLocaleLowerCase()));
  return <>
    <Stack.Screen options={{ title: titles[page] }} />
    <NativeFilterSheet title={titles[page]} search={page === 'tags' ? { query: search, placeholder: 'Search tags', onChange: setSearch, onSubmit: setSearch, onClear: () => setSearch('') } : undefined} footerTestID="browse-filter-footer" actions={{
      primaryLabel: 'Show results', secondaryLabel: page === 'overview' ? 'Cancel' : 'Back',
      secondaryAccessibilityLabel: page === 'overview' ? 'Cancel filters' : 'Back to filters', disabled: busy,
      onApply: () => onApply(draft), onBack: () => { if (page === 'overview') onCancel(); else { onCancelPending?.(); open('overview'); } }
    }}>
      {error ? <Text accessibilityRole="alert" style={styles.errorMessage}>{error}</Text> : null}
      {page === 'overview' ? <>
        <SettingsSection>
          <SettingsPickerRow label="Type" accessibilityLabel="Choose type" value={draft.scope} options={choices.scope} disabled={busy} onChange={value => setDraft({ ...draft, scope: value })} />
          <SettingsPickerRow label="Status" accessibilityLabel="Choose status" value={draft.lifecycleState} options={choices.lifecycleState} disabled={busy} onChange={value => setDraft({ ...draft, lifecycleState: value })} />
          <SettingsPickerRow label="Availability" accessibilityLabel="Choose availability" value={draft.checkoutState} options={choices.checkoutState} disabled={busy} onChange={value => setDraft({ ...draft, checkoutState: value })} />
          <SettingsNavigationRow label="Tags" context={draft.tagIds.length ? draft.tagIds.length + ' selected' : 'Any tags'} accessibilityLabel="Choose tags" onPress={() => open('tags')} />
          {searchMode ? <SettingsValueRow label="Sort" value="Relevance while searching" /> : <SettingsPickerRow label="Sort" accessibilityLabel="Choose sort" value={draft.sort} options={choices.sort} disabled={busy} onChange={value => setDraft({ ...draft, sort: value })} />}
        </SettingsSection>
        <SettingsSection footer="Reviews active items only.">
          <View style={styles.navigationRow}>
            <NativeActionMenu accessibilityLabel="Choose expiration review" disabled={busy}
              trigger={{ kind: 'row', label: 'Review expiration' }} groups={[{ id: 'expiration', items: [
                { id: 'soon', label: 'Expiring soon', onPress: () => onExpiration('soon', draft) },
                { id: 'expired', label: 'Expired', onPress: () => onExpiration('expired', draft) },
                { id: 'all', label: 'All dates', onPress: () => onExpiration('all', draft) }
              ] }]} />
          </View>
        </SettingsSection>
        <SettingsSection><SettingsActionRow label="Reset all" accessibilityLabel="Reset all filters" onPress={() => setDraft(defaults)} /></SettingsSection>
      </> : <SettingsSection footer={visibleTags.length ? undefined : tags.length ? 'No matching tags' : 'No tags available'}>
        {visibleTags.map(tag =>
          <SettingsChoiceRow key={tag.id} multiple label={tag.label} accessibilityLabel={'Filter by tag ' + tag.label} selected={draft.tagIds.includes(tag.id)}
            onPress={() => setDraft({ ...draft, tagIds: draft.tagIds.includes(tag.id) ? draft.tagIds.filter(id => id !== tag.id) : [...draft.tagIds, tag.id] })} />
        )}
      </SettingsSection>}
    </NativeFilterSheet>
  </>;
}
