import type { SearchBarCommands, SearchBarProps } from 'react-native-screens';
import { SettingsPickerRow } from '../components/SettingsPickerRow';
import { useMemo, useRef, useState } from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';
import { Stack } from 'expo-router';
import { SafeAreaView } from 'react-native-safe-area-context';
import { NativeSheetActions } from '../components/NativeSheetActions';
import type { ExpirationFilter } from '../../application/expiration/ExpirationRepository';
import { SettingsActionRow, SettingsChoiceRow, SettingsNavigationRow, SettingsSection, useSettingsListStyles } from '../screens/SettingsList';
import { useSheetKeyboardInset } from './useSheetKeyboardInset';
import { ExpirationDateRange } from './ExpirationDateRange';
export type ExpirationChoices = { readonly types: readonly Choice[]; readonly tags: readonly Choice[]; readonly locations: readonly Choice[] };
type Choice = { readonly id: string; readonly label: string };
type Page = 'overview' | 'types' | 'tags' | 'locations' | 'dates';
export function ExpirationFiltersScreen({ initial, choices, onApply, onCancel }: { readonly initial: ExpirationFilter; readonly choices: ExpirationChoices; readonly onApply: (filter: ExpirationFilter) => void; readonly onCancel: () => void }) {
 const [draft, setDraft] = useState(initial); const [page, setPage] = useState<Page>('overview'); const [search, setSearch] = useState('');
 const [footerHeight, setFooterHeight] = useState(0);
 const boundaryRef = useRef<View>(null);
 const keyboard = useSheetKeyboardInset(boundaryRef);
 const { palette } = useSettingsListStyles();
 const rangeError = !!draft.fromDate && !!draft.throughDate && draft.fromDate > draft.throughDate;
 const searchRef = useRef<SearchBarCommands | null>(null);
 const open = (next: Page) => { searchRef.current?.clearText(); setSearch(''); setPage(next); };
 const searchable = page === 'types' || page === 'tags' || page === 'locations';
 const label = (items: readonly Choice[], id?: string) => items.find(item => item.id === id)?.label ?? (id ? 'Selected' : 'Any');
 const headerOptions = useMemo(() => ({ headerShown: true, title: page === 'overview' ? 'Filters' : page === 'dates' ? 'Date range' : page[0].toUpperCase() + page.slice(1),
   headerSearchBarOptions: searchable ? { ref:searchRef, placeholder:`Search ${page}`, placement:'stacked', hideWhenScrolling:false, hideNavigationBar:false, obscureBackground:false, autoCapitalize:'none', onChangeText:event=>setSearch(event.nativeEvent.text), onCancelButtonPress:()=>setSearch('') } satisfies SearchBarProps : undefined,
  }), [page, searchable]);
 return <>
  <Stack.Screen options={headerOptions} />
  <ScrollView automaticallyAdjustKeyboardInsets style={[styles.shell, { backgroundColor: palette.background }]} contentContainerStyle={{ paddingBottom: footerHeight + 20 }} scrollIndicatorInsets={{ bottom: footerHeight }} keyboardShouldPersistTaps="handled" keyboardDismissMode="on-drag" contentInsetAdjustmentBehavior="automatic">
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
  </ScrollView>
  <View ref={boundaryRef} collapsable={false} pointerEvents="none" onLayout={keyboard.measure} style={styles.bottomBoundary} />
  <SafeAreaView edges={keyboard.bottomInset > 0 ? [] : ['bottom']} onLayout={event => setFooterHeight(event.nativeEvent.layout.height)} style={[styles.footerOverlay, { bottom: keyboard.bottomInset, backgroundColor: palette.background }]}>
   <View testID="expiration-filter-footer" style={styles.footer}>
    <NativeSheetActions primaryLabel="Apply filters" primaryAccessibilityLabel="Apply expiration filters" secondaryAccessibilityLabel="Cancel or return to filters" secondaryLabel={page === 'overview' ? 'Cancel' : 'Back'} disabled={rangeError}
      onBack={() => page === 'overview' ? onCancel() : open('overview')} onApply={() => onApply(draft)} />
   </View>
  </SafeAreaView>
 </>;
}
const styles = StyleSheet.create({ shell: { flex: 1 }, bottomBoundary: { position: 'absolute', bottom: 0, left: 0, right: 0, height: 0 }, footerOverlay: { position: 'absolute', left: 0, right: 0, bottom: 0 }, footer: { paddingHorizontal: 20, paddingVertical: 12 } });
