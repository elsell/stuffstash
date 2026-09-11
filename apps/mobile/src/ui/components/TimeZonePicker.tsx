import { useMemo, useState } from 'react';
import { Pressable, Text, View } from 'react-native';
import { AppTextInput } from './AppTextInput';
import { SelectionRow } from './SelectionRow';
import { useAppearancePalette } from '../theme/AppearanceContext';
export function readableTimeZone(zone: string) { return zone.replaceAll('_', ' ').split('/').reverse().join(' · '); }
export function TimeZonePicker({ value, disabled, onChange }: { readonly value: string; readonly disabled?: boolean; readonly onChange: (zone: string) => Promise<void> }) {
  const colors = useAppearancePalette();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [error, setError] = useState('');
  const zones = useMemo(() => {
    const available = (Intl as typeof Intl & { supportedValuesOf?: (key: string) => string[] }).supportedValuesOf?.('timeZone') ?? [];
    return [...new Set([value, 'UTC', Intl.DateTimeFormat().resolvedOptions().timeZone, ...available])];
  }, [value]);
  const matches = zones.filter(zone => readableTimeZone(zone).toLocaleLowerCase().includes(query.toLocaleLowerCase())).slice(0, 20);
  async function select(zone: string) {
    setError('');
    try { await onChange(zone); setOpen(false); setQuery(''); }
    catch { setError('Could not save the time zone. Try again.'); }
  }
  let validQuery = false;
  try { new Intl.DateTimeFormat(undefined, { timeZone: query }); validQuery = !!query.trim(); } catch { /* Keep incomplete searches editable. */ }
  return <SelectionRow label="Time zone" value={readableTimeZone(value)} expanded={open} disabled={disabled} onPress={() => setOpen(current => !current)}>
    <Text style={{ color: colors.textMuted }}>Dates end at midnight in this time zone. It stays the same when you travel.</Text>
    <AppTextInput accessibilityLabel="Search time zones" value={query} onChangeText={setQuery} placeholder="Search city or time zone" autoCapitalize="none" autoCorrect={false} style={{ minHeight: 48, color: colors.text }} />
    <View>{[...new Set([...matches, ...(validQuery && !matches.includes(query) ? [query] : [])])].map(zone => <Pressable key={zone} accessibilityRole="radio" accessibilityLabel={readableTimeZone(zone)} accessibilityState={{ checked: value === zone }} disabled={disabled} onPress={() => void select(zone)} style={{ minHeight: 48, justifyContent: 'center' }}><Text style={{ color: colors.action }}>{value === zone ? '✓ ' : ''}{readableTimeZone(zone)}</Text></Pressable>)}</View>
    {matches.length === 20 ? <Text style={{ color: colors.textMuted }}>Type more to narrow the results.</Text> : null}
    {error ? <Text accessibilityRole="alert" style={{ color: colors.danger }}>{error}</Text> : null}
  </SelectionRow>;
}
