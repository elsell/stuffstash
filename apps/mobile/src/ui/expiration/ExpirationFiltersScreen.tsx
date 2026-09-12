import { useState } from 'react';
import { Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import type { ExpirationFilter } from '../../application/expiration/ExpirationRepository';
import { AppTextInput } from '../components/AppTextInput';
import { SettingsChoiceRow, SettingsNavigationRow, SettingsSection, SettingsSwitchRow, useSettingsListStyles } from '../screens/SettingsList';
import { ExpirationDateRange } from './ExpirationDateRange';
export type ExpirationChoices = { readonly types: readonly Choice[]; readonly tags: readonly Choice[]; readonly locations: readonly Choice[] };
type Choice = { readonly id: string; readonly label: string };
type Page = 'overview' | 'types' | 'tags' | 'locations' | 'dates' | 'kinds' | 'availability';
export function ExpirationFiltersScreen({ initial, choices, onApply, onCancel }: { readonly initial: ExpirationFilter; readonly choices: ExpirationChoices; readonly onApply: (filter: ExpirationFilter) => void; readonly onCancel: () => void }) {
 const [draft, setDraft] = useState(initial); const [page, setPage] = useState<Page>('overview'); const [search, setSearch] = useState('');
 const { palette } = useSettingsListStyles();
 const rangeError = !!draft.fromDate && !!draft.throughDate && draft.fromDate > draft.throughDate;
 const open = (next: Page) => { setSearch(''); setPage(next); };
 const label = (items: readonly Choice[], id?: string) => items.find(item => item.id === id)?.label ?? (id ? 'Selected' : 'Any');
 const action = (title: string, accessibilityLabel: string, onPress: () => void, disabled = false) => <Pressable accessibilityRole="button" accessibilityLabel={accessibilityLabel} disabled={disabled} onPress={onPress} style={styles.button}><Text style={{ color: disabled ? palette.textMuted : palette.action, fontSize: 17 }}>{title}</Text></Pressable>;
 return <View style={[styles.shell, { backgroundColor: palette.background }]}>
  <View style={styles.toolbar}>{action(page === 'overview' ? 'Cancel' : 'Back', 'Cancel or return to filters', () => page === 'overview' ? onCancel() : open('overview'))}<Text accessibilityRole="header" style={{ color: palette.text, fontSize: 18, fontWeight: '600' }}>{page === 'overview' ? 'Filters' : page === 'dates' ? 'Date range' : page[0].toUpperCase() + page.slice(1)}</Text>{action('Apply', 'Apply expiration filters', () => onApply(draft), rangeError)}</View>
  <ScrollView contentContainerStyle={styles.content} keyboardShouldPersistTaps="handled" automaticallyAdjustKeyboardInsets>
   {page === 'overview' ? <>
    <SettingsSection>
     <SettingsNavigationRow accessibilityLabel="Choose item kind" label="Kind" value={draft.kind ?? 'Any'} onPress={() => open('kinds')} />
     <SettingsNavigationRow accessibilityLabel="Choose availability" label="Availability" value={draft.checkoutState === 'checked_out' ? 'Checked out' : draft.checkoutState === 'available' ? 'Available' : 'Any'} onPress={() => open('availability')} />
     <SettingsNavigationRow accessibilityLabel="Choose type" label="Type" value={label(choices.types, draft.typeId)} onPress={() => open('types')} />
     <SettingsNavigationRow accessibilityLabel="Choose tags" label="Tags" value={draft.tagIds?.length ? `${draft.tagIds.length} selected` : 'Any'} onPress={() => open('tags')} />
     <SettingsNavigationRow accessibilityLabel="Choose location" label="Location" value={label(choices.locations, draft.locationId)} onPress={() => open('locations')} />
     <SettingsNavigationRow accessibilityLabel="Choose date range" label="Date range" value={draft.fromDate || draft.throughDate ? 'Custom' : 'Any date'} onPress={() => open('dates')} />
    </SettingsSection>
    {action('Clear filters', 'Clear expiration filters', () => setDraft({ mode: draft.mode }))}
   </> : page === 'kinds' || page === 'availability' ? <SettingsSection>{(page === 'kinds' ? [{id:'',label:'Any kind'},{id:'item',label:'Items'},{id:'container',label:'Containers'},{id:'location',label:'Places'}] : [{id:'',label:'Any availability'},{id:'available',label:'Available'},{id:'checked_out',label:'Checked out'}]).map(option => <SettingsChoiceRow key={option.id} label={option.label} selected={option.id === (page === 'kinds' ? draft.kind ?? '' : draft.checkoutState ?? '')} onPress={() => {setDraft({...draft,...(page === 'kinds' ? {kind:option.id as ExpirationFilter['kind'] || undefined} : {checkoutState:option.id as ExpirationFilter['checkoutState'] || undefined})});open('overview');}} />)}</SettingsSection> : page === 'dates' ? <ExpirationDateRange fromDate={draft.fromDate} throughDate={draft.throughDate} onChange={range => setDraft({ ...draft, ...range })} /> : <>
    <AppTextInput accessibilityLabel={`Search ${page}`} value={search} onChangeText={setSearch} placeholder={`Search ${page}`} style={[styles.search, { color: palette.text, borderColor: palette.border }]} />
    <SettingsSection>
     {page !== 'tags' ? <SettingsChoiceRow label="Any" selected={page === 'types' ? !draft.typeId : !draft.locationId} onPress={() => { setDraft({ ...draft, ...(page === 'types' ? { typeId: undefined } : { locationId: undefined }) }); open('overview'); }} /> : null}
     {choices[page].filter(item => item.label.toLocaleLowerCase().includes(search.toLocaleLowerCase())).map(item => page === 'tags' ? <SettingsSwitchRow key={item.id} label={item.label} value={draft.tagIds?.includes(item.id) ?? false} onValueChange={selected => setDraft({ ...draft, tagIds: selected ? [...(draft.tagIds ?? []), item.id] : draft.tagIds?.filter(id => id !== item.id) })} /> : <SettingsChoiceRow key={item.id} label={item.label} selected={page === 'types' ? draft.typeId === item.id : draft.locationId === item.id} onPress={() => { setDraft({ ...draft, ...(page === 'types' ? { typeId: item.id } : { locationId: item.id }) }); open('overview'); }} />)}
    </SettingsSection>
    {!choices[page].some(item => item.label.toLocaleLowerCase().includes(search.toLocaleLowerCase())) ? <Text style={{ color: palette.textMuted }}>No matches</Text> : null}
   </>}
   {rangeError ? <Text accessibilityRole="alert" style={{ color: palette.text }}>The end date must be on or after the start date.</Text> : null}
  </ScrollView>
 </View>;
}
const styles = StyleSheet.create({ shell: { flex: 1 }, toolbar: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', padding: 12 }, content: { padding: 20 }, button: { minHeight: 44, justifyContent: 'center', paddingHorizontal: 8 }, search: { minHeight: 44, borderWidth: StyleSheet.hairlineWidth, paddingHorizontal: 12, borderRadius: 10, fontSize: 17, marginBottom: 16 } });
