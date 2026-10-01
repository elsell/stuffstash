import { t } from '../../presentation/localization';
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
  scope: [{ value: 'all', label: t('mobile.BrowseFiltersScreen.allTypes') }, { value: 'places', label: t('mobile.BrowseFiltersScreen.places') }, { value: 'containers', label: t('mobile.BrowseFiltersScreen.containers') }, { value: 'items', label: t('mobile.BrowseFiltersScreen.items') }],
  lifecycleState: [{ value: 'active', label: t('mobile.BrowseFiltersScreen.active') }, { value: 'archived', label: t('mobile.BrowseFiltersScreen.archived') }, { value: 'all', label: t('mobile.BrowseFiltersScreen.allStatuses') }],
  checkoutState: [{ value: 'any', label: t('mobile.BrowseFiltersScreen.anyAvailability') }, { value: 'available', label: t('mobile.BrowseFiltersScreen.available') }, { value: 'checked_out', label: t('mobile.BrowseFiltersScreen.checkedOut') }],
  sort: [{ value: 'updated_desc', label: t('mobile.BrowseFiltersScreen.recentlyChanged') }, { value: 'id_asc', label: t('mobile.BrowseFiltersScreen.defaultOrder') }]
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
    <NativeFilterSheet title={titles[page]} search={page === 'tags' ? { query: search, placeholder: t('mobile.BrowseFiltersScreen.searchTags'), onChange: setSearch, onSubmit: setSearch, onClear: () => setSearch('') } : undefined} footerTestID="browse-filter-footer" actions={{
      primaryLabel: 'Show results', secondaryLabel: page === 'overview' ? 'Cancel' : 'Back',
      secondaryAccessibilityLabel: page === 'overview' ? 'Cancel filters' : 'Back to filters', disabled: busy,
      onApply: () => onApply(draft), onBack: () => { if (page === 'overview') onCancel(); else { onCancelPending?.(); open('overview'); } }
    }}>
      {error ? <Text accessibilityRole="alert" style={styles.errorMessage}>{error}</Text> : null}
      {page === 'overview' ? <>
        <SettingsSection>
          <SettingsPickerRow label={t('mobile.BrowseFiltersScreen.type')} accessibilityLabel={t('mobile.BrowseFiltersScreen.chooseType')} value={draft.scope} options={choices.scope} disabled={busy} onChange={value => setDraft({ ...draft, scope: value })} />
          <SettingsPickerRow label={t('mobile.BrowseFiltersScreen.status')} accessibilityLabel={t('mobile.BrowseFiltersScreen.chooseStatus')} value={draft.lifecycleState} options={choices.lifecycleState} disabled={busy} onChange={value => setDraft({ ...draft, lifecycleState: value })} />
          <SettingsPickerRow label={t('mobile.BrowseFiltersScreen.availability')} accessibilityLabel={t('mobile.BrowseFiltersScreen.chooseAvailability')} value={draft.checkoutState} options={choices.checkoutState} disabled={busy} onChange={value => setDraft({ ...draft, checkoutState: value })} />
          <SettingsNavigationRow label={t('mobile.BrowseFiltersScreen.tags')} context={draft.tagIds.length ? draft.tagIds.length + ' selected' : 'Any tags'} accessibilityLabel={t('mobile.BrowseFiltersScreen.chooseTags')} onPress={() => open('tags')} />
          {searchMode ? <SettingsValueRow label={t('mobile.BrowseFiltersScreen.sort')} value="Relevance while searching" /> : <SettingsPickerRow label={t('mobile.BrowseFiltersScreen.sort')} accessibilityLabel={t('mobile.BrowseFiltersScreen.chooseSort')} value={draft.sort} options={choices.sort} disabled={busy} onChange={value => setDraft({ ...draft, sort: value })} />}
        </SettingsSection>
        <SettingsSection footer="Reviews active items only.">
          <View style={styles.navigationRow}>
            <NativeActionMenu accessibilityLabel={t('mobile.BrowseFiltersScreen.chooseExpirationReview')} disabled={busy}
              trigger={{ kind: 'row', label: t('mobile.BrowseFiltersScreen.reviewExpiration') }} groups={[{ id: 'expiration', items: [
                { id: 'soon', label: t('mobile.BrowseFiltersScreen.expiringSoon'), onPress: () => onExpiration('soon', draft) },
                { id: 'expired', label: t('mobile.BrowseFiltersScreen.expired'), onPress: () => onExpiration('expired', draft) },
                { id: 'all', label: t('mobile.BrowseFiltersScreen.allDates'), onPress: () => onExpiration('all', draft) }
              ] }]} />
          </View>
        </SettingsSection>
        <SettingsSection><SettingsActionRow label={t('mobile.BrowseFiltersScreen.resetAll')} accessibilityLabel={t('mobile.BrowseFiltersScreen.resetAllFilters')} onPress={() => setDraft(defaults)} /></SettingsSection>
      </> : <SettingsSection footer={visibleTags.length ? undefined : tags.length ? 'No matching tags' : 'No tags available'}>
        {visibleTags.map(tag =>
          <SettingsChoiceRow key={tag.id} multiple label={tag.label} accessibilityLabel={'Filter by tag ' + tag.label} selected={draft.tagIds.includes(tag.id)}
            onPress={() => setDraft({ ...draft, tagIds: draft.tagIds.includes(tag.id) ? draft.tagIds.filter(id => id !== tag.id) : [...draft.tagIds, tag.id] })} />
        )}
      </SettingsSection>}
    </NativeFilterSheet>
  </>;
}
