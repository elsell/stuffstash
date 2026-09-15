import { NativeNavigationSearch } from '../components/NativeNavigationSearch';
import { SettingsPickerRow } from '../components/SettingsPickerRow';
import { useMemo, useState } from 'react';
import { Text } from 'react-native';
import { Stack } from 'expo-router';
import { NativeFilterSheet } from '../components/NativeFilterSheet';
import type { ExpirationFilter } from '../../application/expiration/ExpirationRepository';
import { SettingsActionRow, SettingsChoiceRow, SettingsNavigationRow, SettingsSection, useSettingsListStyles } from '../screens/SettingsList';
import { ExpirationDateRange } from './ExpirationDateRange';
export type ExpirationChoices = { readonly types: readonly Choice[]; readonly tags: readonly Choice[]; readonly locations: readonly Choice[] };
type Choice = { readonly id: string; readonly label: string };
type Page = 'overview' | 'types' | 'tags' | 'locations' | 'dates';
export function ExpirationFiltersScreen({ initial, choices, onApply, onCancel }: { readonly initial: ExpirationFilter; readonly choices: ExpirationChoices; readonly onApply: (filter: ExpirationFilter) => void; readonly onCancel: () => void }) {
 const [draft, setDraft] = useState(initial); const [page, setPage] = useState<Page>('overview'); const [search, setSearch] = useState('');
 const { palette } = useSettingsListStyles();
 const rangeError = !!draft.fromDate && !!draft.throughDate && draft.fromDate > draft.throughDate;
 const open = (next: Page) => { setSearch(''); setPage(next); };
 const searchable = page === 'types' || page === 'tags' || page === 'locations';
 const label = (items: readonly Choice[], id?: string) => items.find(item => item.id === id)?.label ?? (id ? 'Selected' : 'Any');
 const headerOptions = useMemo(() => ({ headerShown: true, title: page === 'overview' ? 'Filters' : page === 'dates' ? 'Date range' : page[0].toUpperCase() + page.slice(1),
  }), [page]);
 return <>
  <Stack.Screen options={headerOptions} />
  <NativeNavigationSearch key={page} enabled={searchable} query={search} placeholder={`Search ${page}`} onChange={setSearch} onSubmit={setSearch} onClear={() => setSearch('')} />
  <NativeFilterSheet footerTestID="expiration-filter-footer" actions={{
   primaryLabel: 'Apply filters', primaryAccessibilityLabel: 'Apply expiration filters', secondaryAccessibilityLabel: 'Cancel or return to filters',
   secondaryLabel: page === 'overview' ? 'Cancel' : 'Back', disabled: rangeError,
   onBack: () => page === 'overview' ? onCancel() : open('overview'), onApply: () => onApply(draft)
  }}>
   {page === 'overview' ? <>
    <SettingsSection>
     <SettingsPickerRow label="Kind" accessibilityLabel="Choose item kind" value={draft.kind ?? ''} options={[{value:'',label:'Any kind'},{value:'item',label:'Items'},{value:'container',label:'Containers'},{value:'location',label:'Places'}] as const} onChange={value => setDraft({...draft,kind:value || undefined})} />
     <SettingsPickerRow label="Availability" accessibilityLabel="Choose availability" value={draft.checkoutState ?? ''} options={[{value:'',label:'Any availability'},{value:'available',label:'Available'},{value:'checked_out',label:'Checked out'}] as const} onChange={value => setDraft({...draft,checkoutState:value || undefined})} />
     <SettingsNavigationRow accessibilityLabel="Choose type" label="Type" value={label(choices.types, draft.typeId)} onPress={() => open('types')} />
     <SettingsNavigationRow accessibilityLabel="Choose tags" label="Tags" value={draft.tagIds?.length ? `${draft.tagIds.length} selected` : 'Any'} onPress={() => open('tags')} />
     <SettingsNavigationRow accessibilityLabel="Choose location" label="Location" value={label(choices.locations, draft.locationId)} onPress={() => open('locations')} />
     <SettingsNavigationRow accessibilityLabel="Choose date range" label="Date range" value={draft.fromDate || draft.throughDate ? 'Custom' : 'Any date'} onPress={() => open('dates')} />
    </SettingsSection>
    <SettingsSection><SettingsActionRow label="Clear filters" accessibilityLabel="Clear expiration filters" onPress={() => setDraft({ mode: draft.mode })} /></SettingsSection>
   </> : page === 'dates' ? <ExpirationDateRange fromDate={draft.fromDate} throughDate={draft.throughDate} onChange={range => setDraft({ ...draft, ...range })} /> : <>
    <SettingsSection>
     {page !== 'tags' ? <SettingsChoiceRow label="Any" selected={page === 'types' ? !draft.typeId : !draft.locationId} onPress={() => { setDraft({ ...draft, ...(page === 'types' ? { typeId: undefined } : { locationId: undefined }) }); open('overview'); }} /> : null}
     {choices[page].filter(item => item.label.toLocaleLowerCase().includes(search.toLocaleLowerCase())).map(item => page === 'tags' ? <SettingsChoiceRow multiple key={item.id} label={item.label} selected={draft.tagIds?.includes(item.id) ?? false} onPress={() => setDraft({ ...draft, tagIds: !draft.tagIds?.includes(item.id) ? [...(draft.tagIds ?? []), item.id] : draft.tagIds?.filter(id => id !== item.id) })} /> : <SettingsChoiceRow key={item.id} label={item.label} selected={page === 'types' ? draft.typeId === item.id : draft.locationId === item.id} onPress={() => { setDraft({ ...draft, ...(page === 'types' ? { typeId: item.id } : { locationId: item.id }) }); open('overview'); }} />)}
    </SettingsSection>
    {!choices[page].some(item => item.label.toLocaleLowerCase().includes(search.toLocaleLowerCase())) ? <Text style={{ color: palette.textMuted }}>No matches</Text> : null}
   </>}
   {rangeError ? <Text accessibilityRole="alert" style={{ color: palette.text }}>The end date must be on or after the start date.</Text> : null}
  </NativeFilterSheet>
 </>;
}
