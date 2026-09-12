import { useRef, useState } from 'react';
import { KeyboardAvoidingView, Platform, ScrollView, StyleSheet, Text, View } from 'react-native';
import { Stack } from 'expo-router';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { SearchBarCommands } from 'react-native-screens';
import { NativeRefinementButton } from '../components/NativeRefinementButton';
import type { ExpirationFilter } from '../../application/expiration/ExpirationRepository';
import { SettingsActionRow, SettingsChoiceRow, SettingsNavigationRow, SettingsSection, useSettingsListStyles } from '../screens/SettingsList';
import { ExpirationDateRange } from './ExpirationDateRange';
export type ExpirationChoices = { readonly types: readonly Choice[]; readonly tags: readonly Choice[]; readonly locations: readonly Choice[] };
type Choice = { readonly id: string; readonly label: string };
type Page = 'overview' | 'types' | 'tags' | 'locations' | 'dates' | 'kinds' | 'availability';
export function ExpirationFiltersScreen({ initial, choices, onApply, onCancel }: { readonly initial: ExpirationFilter; readonly choices: ExpirationChoices; readonly onApply: (filter: ExpirationFilter) => void; readonly onCancel: () => void }) {
 const [draft, setDraft] = useState(initial); const [page, setPage] = useState<Page>('overview'); const [search, setSearch] = useState('');
 const { palette } = useSettingsListStyles();
 const rangeError = !!draft.fromDate && !!draft.throughDate && draft.fromDate > draft.throughDate;
 const searchRef = useRef<SearchBarCommands | null>(null);
 const open = (next: Page) => { searchRef.current?.clearText(); setSearch(''); setPage(next); };
 const searchable = page === 'types' || page === 'tags' || page === 'locations';
 const label = (items: readonly Choice[], id?: string) => items.find(item => item.id === id)?.label ?? (id ? 'Selected' : 'Any');
 return <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : undefined} style={[styles.shell, { backgroundColor: palette.background }]}>
  <Stack.Screen options={{ headerShown: true, title: page === 'overview' ? 'Filters' : page === 'dates' ? 'Date range' : page[0].toUpperCase() + page.slice(1),
   headerSearchBarOptions: searchable ? { ref:searchRef, placeholder:`Search ${page}`, placement:'stacked', hideWhenScrolling:false, hideNavigationBar:false, obscureBackground:false, autoCapitalize:'none', onChangeText:event=>setSearch(event.nativeEvent.text), onCancelButtonPress:()=>setSearch('') } : undefined,
  }} />
  <ScrollView style={styles.shell} contentContainerStyle={styles.content} keyboardShouldPersistTaps="handled" keyboardDismissMode="on-drag" contentInsetAdjustmentBehavior="automatic">
   {page === 'overview' ? <>
    <SettingsSection>
     <SettingsNavigationRow accessibilityLabel="Choose item kind" label="Kind" value={draft.kind ?? 'Any'} onPress={() => open('kinds')} />
     <SettingsNavigationRow accessibilityLabel="Choose availability" label="Availability" value={draft.checkoutState === 'checked_out' ? 'Checked out' : draft.checkoutState === 'available' ? 'Available' : 'Any'} onPress={() => open('availability')} />
     <SettingsNavigationRow accessibilityLabel="Choose type" label="Type" value={label(choices.types, draft.typeId)} onPress={() => open('types')} />
     <SettingsNavigationRow accessibilityLabel="Choose tags" label="Tags" value={draft.tagIds?.length ? `${draft.tagIds.length} selected` : 'Any'} onPress={() => open('tags')} />
     <SettingsNavigationRow accessibilityLabel="Choose location" label="Location" value={label(choices.locations, draft.locationId)} onPress={() => open('locations')} />
     <SettingsNavigationRow accessibilityLabel="Choose date range" label="Date range" value={draft.fromDate || draft.throughDate ? 'Custom' : 'Any date'} onPress={() => open('dates')} />
    </SettingsSection>
    <SettingsSection><SettingsActionRow label="Clear filters" accessibilityLabel="Clear expiration filters" onPress={() => setDraft({ mode: draft.mode })} /></SettingsSection>
   </> : page === 'kinds' || page === 'availability' ? <SettingsSection>{(page === 'kinds' ? [{id:'',label:'Any kind'},{id:'item',label:'Items'},{id:'container',label:'Containers'},{id:'location',label:'Places'}] : [{id:'',label:'Any availability'},{id:'available',label:'Available'},{id:'checked_out',label:'Checked out'}]).map(option => <SettingsChoiceRow key={option.id} label={option.label} selected={option.id === (page === 'kinds' ? draft.kind ?? '' : draft.checkoutState ?? '')} onPress={() => {setDraft({...draft,...(page === 'kinds' ? {kind:option.id as ExpirationFilter['kind'] || undefined} : {checkoutState:option.id as ExpirationFilter['checkoutState'] || undefined})});open('overview');}} />)}</SettingsSection> : page === 'dates' ? <ExpirationDateRange fromDate={draft.fromDate} throughDate={draft.throughDate} onChange={range => setDraft({ ...draft, ...range })} /> : <>
    <SettingsSection>
     {page !== 'tags' ? <SettingsChoiceRow label="Any" selected={page === 'types' ? !draft.typeId : !draft.locationId} onPress={() => { setDraft({ ...draft, ...(page === 'types' ? { typeId: undefined } : { locationId: undefined }) }); open('overview'); }} /> : null}
     {choices[page].filter(item => item.label.toLocaleLowerCase().includes(search.toLocaleLowerCase())).map(item => page === 'tags' ? <SettingsChoiceRow key={item.id} label={item.label} selected={draft.tagIds?.includes(item.id) ?? false} onPress={() => setDraft({ ...draft, tagIds: !draft.tagIds?.includes(item.id) ? [...(draft.tagIds ?? []), item.id] : draft.tagIds?.filter(id => id !== item.id) })} /> : <SettingsChoiceRow key={item.id} label={item.label} selected={page === 'types' ? draft.typeId === item.id : draft.locationId === item.id} onPress={() => { setDraft({ ...draft, ...(page === 'types' ? { typeId: item.id } : { locationId: item.id }) }); open('overview'); }} />)}
    </SettingsSection>
    {!choices[page].some(item => item.label.toLocaleLowerCase().includes(search.toLocaleLowerCase())) ? <Text style={{ color: palette.textMuted }}>No matches</Text> : null}
   </>}
   {rangeError ? <Text accessibilityRole="alert" style={{ color: palette.text }}>The end date must be on or after the start date.</Text> : null}
  </ScrollView>
  <SafeAreaView edges={['bottom']} style={{backgroundColor:palette.background}}>
   <View testID="expiration-filter-footer" style={styles.footer}>
    <NativeRefinementButton label={page === 'overview' ? 'Cancel' : 'Back'} accessibilityLabel="Cancel or return to filters" onPress={() => page === 'overview' ? onCancel() : open('overview')} />
    <NativeRefinementButton label="Apply filters" accessibilityLabel="Apply expiration filters" disabled={rangeError} onPress={() => onApply(draft)} />
   </View>
  </SafeAreaView>
 </KeyboardAvoidingView>;
}
const styles = StyleSheet.create({ shell: { flex: 1 }, content: { paddingBottom: 20 }, footer: { paddingHorizontal: 20, paddingVertical: 12, flexDirection: 'row', flexWrap: 'wrap', justifyContent: 'space-between', gap: 12 } });
