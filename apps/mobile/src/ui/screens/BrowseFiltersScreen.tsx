import { useState } from 'react';
import { Stack } from 'expo-router';
import { ScrollView, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { AssetTagOptionViewModel } from '../../application/assets/InventoryAssetTagsQuery';
import type { AssetBrowseSort } from '../../application/home/InventorySummaryRepository';
import type { BrowseDraftFilters } from './BrowseFilterState';
import { SettingsActionRow, SettingsChoiceRow, SettingsNavigationRow, SettingsSection, useSettingsListStyles } from './SettingsList';
import { NativeSheetActions } from '../components/NativeSheetActions';
import { NativeNavigationSearch } from '../components/NativeNavigationSearch';
import type { ExpirationMode } from '../../application/expiration/ExpirationRepository';

export type BrowseFilterDraft = BrowseDraftFilters & { readonly sort: AssetBrowseSort };
type Page = 'overview' | 'scope' | 'lifecycleState' | 'checkoutState' | 'tags' | 'sort' | 'expiration';
const defaults: BrowseFilterDraft = { scope: 'all', lifecycleState: 'active', checkoutState: 'any', tagIds: [], sort: 'updated_desc' };
const choices = {
  scope: [{ value: 'all', label: 'All types' }, { value: 'places', label: 'Places' }, { value: 'containers', label: 'Containers' }, { value: 'items', label: 'Items' }],
  lifecycleState: [{ value: 'active', label: 'Active' }, { value: 'archived', label: 'Archived' }, { value: 'all', label: 'All statuses' }],
  checkoutState: [{ value: 'any', label: 'Any availability' }, { value: 'available', label: 'Available' }, { value: 'checked_out', label: 'Checked out' }],
  sort: [{ value: 'updated_desc', label: 'Recently changed' }, { value: 'id_asc', label: 'Default order' }]
} as const;
const titles: Record<Page, string> = { overview: 'Filters', scope: 'Type', lifecycleState: 'Status', checkoutState: 'Availability', tags: 'Tags', sort: 'Sort', expiration: 'Expiration' };

export function BrowseFiltersScreen({ initial, query, tags, busy = false, onApply, onCancel, onCancelPending, onExpiration }: {
  readonly initial: BrowseFilterDraft; readonly query: string; readonly tags: readonly AssetTagOptionViewModel[];
  readonly busy?: boolean; readonly onApply: (draft: BrowseFilterDraft) => void; readonly onCancel: () => void;
  readonly onCancelPending?: () => void;
  readonly onExpiration: (mode: ExpirationMode, draft: BrowseFilterDraft) => void;
}) {
  const [draft, setDraft] = useState(initial);
  const [page, setPage] = useState<Page>('overview');
  const [search, setSearch] = useState('');
  const { palette } = useSettingsListStyles();
  const open = (next: Page) => { setSearch(''); setPage(next); };
  const searchMode = !!query.trim() || draft.tagIds.length > 0;
  const summary = (key: keyof typeof choices) => choices[key].find(choice => choice.value === draft[key])?.label ?? '';
  return <View style={{ flex: 1, backgroundColor: palette.background }}>
    <Stack.Screen options={{ title: titles[page], headerSearchBarOptions: undefined }} />
    {page === 'tags' ? <NativeNavigationSearch query={search} placeholder="Search tags" onChange={setSearch} onSubmit={setSearch} onClear={() => setSearch('')} /> : null}
    <ScrollView contentInsetAdjustmentBehavior="automatic" keyboardShouldPersistTaps="handled" keyboardDismissMode="on-drag" style={{ flex: 1 }}>
      {page === 'overview' ? <>
        <SettingsSection>
          <SettingsNavigationRow label="Type" context={summary('scope')} accessibilityLabel="Choose type" onPress={() => open('scope')} />
          <SettingsNavigationRow label="Status" context={summary('lifecycleState')} accessibilityLabel="Choose status" onPress={() => open('lifecycleState')} />
          <SettingsNavigationRow label="Availability" context={summary('checkoutState')} accessibilityLabel="Choose availability" onPress={() => open('checkoutState')} />
          <SettingsNavigationRow label="Tags" context={draft.tagIds.length ? draft.tagIds.length + ' selected' : 'Any tags'} accessibilityLabel="Choose tags" onPress={() => open('tags')} />
          <SettingsNavigationRow label="Sort" context={searchMode ? 'Relevance while searching' : summary('sort')} disabled={searchMode} accessibilityLabel="Choose sort" onPress={() => open('sort')} />
          <SettingsNavigationRow label="Expiration" context="Review active items by date" accessibilityLabel="Choose expiration review" onPress={() => open('expiration')} />
        </SettingsSection>
        <SettingsSection><SettingsActionRow label="Reset all" accessibilityLabel="Reset all filters" onPress={() => setDraft(defaults)} /></SettingsSection>
      </> : page === 'tags' ? <SettingsSection>
        {[...tags].sort((a, b) => a.label.localeCompare(b.label)).filter(tag => tag.label.toLocaleLowerCase().includes(search.toLocaleLowerCase())).map(tag =>
          <SettingsChoiceRow key={tag.id} multiple label={tag.label} accessibilityLabel={'Filter by tag ' + tag.label} selected={draft.tagIds.includes(tag.id)}
            onPress={() => setDraft({ ...draft, tagIds: draft.tagIds.includes(tag.id) ? draft.tagIds.filter(id => id !== tag.id) : [...draft.tagIds, tag.id] })} />
        )}
      </SettingsSection> : page === 'expiration' ? <SettingsSection footer="Expiration reviews active items, regardless of the Browse status filter.">
        <SettingsNavigationRow label="Expiring soon" accessibilityLabel="Review expiring soon items" onPress={() => onExpiration('soon', draft)} />
        <SettingsNavigationRow label="Expired" accessibilityLabel="Review expired items" onPress={() => onExpiration('expired', draft)} />
        <SettingsNavigationRow label="All dates" accessibilityLabel="Review all expiration dates" onPress={() => onExpiration('all', draft)} />
      </SettingsSection> : <SettingsSection>
        {choices[page].map(choice => <SettingsChoiceRow key={choice.value} label={choice.label} selected={draft[page] === choice.value} onPress={() => {
          setDraft({ ...draft, [page]: choice.value }); open('overview');
        }} />)}
      </SettingsSection>}
    </ScrollView>
    <SafeAreaView edges={['bottom']} style={{ paddingHorizontal: 20, paddingTop: 12 }}>
      <NativeSheetActions primaryLabel="Show results" secondaryLabel={page === 'overview' ? 'Cancel' : 'Back'}
        secondaryAccessibilityLabel={page === 'overview' ? 'Cancel filters' : 'Back to filters'} disabled={busy}
        onApply={() => onApply(draft)} onBack={() => { if (page === 'overview') onCancel(); else { onCancelPending?.(); open('overview'); } }} />
    </SafeAreaView>
  </View>;
}
