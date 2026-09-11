import { useEffect, useMemo, useRef, useState } from 'react';
import { Text, View } from 'react-native';
import { AppTextInput } from './AppTextInput';
import { SettingsChoiceRow, SettingsSection, SettingsSeparator, useSettingsListStyles } from '../screens/SettingsList';
export function readableTimeZone(zone: string) { return zone.replaceAll('_', ' ').split('/').reverse().join(' · '); }
export function TimeZonePicker({ value, disabled, onChange }: { readonly value: string; readonly disabled?: boolean; readonly onChange: (zone: string) => Promise<void> }) {
  const { palette, styles } = useSettingsListStyles();
  const [query, setQuery] = useState(''); const [error, setError] = useState(''); const [saving,setSaving]=useState(false);
  const pending=useRef(false);
  const mounted=useRef(true);
  useEffect(()=>{mounted.current=true;return()=>{mounted.current=false;};},[]);
  const zones = useMemo(() => {
    const available = (Intl as typeof Intl & { supportedValuesOf?: (key: string) => string[] }).supportedValuesOf?.('timeZone') ?? [];
    return [...new Set([value, 'UTC', Intl.DateTimeFormat().resolvedOptions().timeZone, ...available])];
  }, [value]);
  const matches = zones.filter(zone => readableTimeZone(zone).toLocaleLowerCase().includes(query.trim().toLocaleLowerCase())).slice(0, 30);
  async function select(zone: string) {
    if(disabled || pending.current)return;
    pending.current=true;setSaving(true);setError('');
    try { await onChange(zone); } catch { if(mounted.current)setError('Could not save the time zone. Try again.'); }
    finally {pending.current=false;if(mounted.current)setSaving(false);}
  }
  let validQuery = false;
  try { new Intl.DateTimeFormat(undefined, { timeZone: query.trim() }); validQuery = !!query.trim(); } catch { /* An incomplete search remains editable. */ }
  const choices=[...new Set([...matches, ...(validQuery ? [query.trim()] : [])])];
  return <>
    <SettingsSection footer="Dates end at midnight in this time zone. It stays the same when you travel.">
      <View style={styles.navigationRow}><AppTextInput accessibilityLabel="Search time zones" value={query} onChangeText={setQuery} placeholder="Search city or time zone" autoCapitalize="none" autoCorrect={false} clearButtonMode="while-editing" style={{ minHeight:44,color:palette.text,fontSize:17 }} /></View>
    </SettingsSection>
    <SettingsSection footer={matches.length === 30 ? 'Search to find another city or time zone.' : undefined}>
      {choices.map((zone,index)=><View key={zone}>{index ? <SettingsSeparator /> : null}<SettingsChoiceRow label={readableTimeZone(zone)} selected={value===zone} disabled={disabled || saving} onPress={()=>void select(zone)}/></View>)}
      {!choices.length ? <View style={styles.navigationRow}><Text style={styles.secondaryText}>No matching time zones.</Text></View> : null}
    </SettingsSection>
    {error ? <View style={styles.detailHeader}><Text accessibilityRole="alert" style={{color:palette.danger}}>{error}</Text></View> : null}
  </>;
}
