import { NativeCommandButton } from './NativeCommandButton';
import { NativeChoicePicker } from './NativeChoicePicker';
import { SelectionRow } from './SelectionRow';
import { expirationMonthCalendarNotice, expirationMonthOptions, formatAssetExpiration } from '../presentation/ExpirationPresentation';
import { useState } from 'react';
import { Platform, StyleSheet, Text, View } from 'react-native';
import DateTimePicker, { type DateTimePickerEvent } from '@react-native-community/datetimepicker';
import type { AssetExpiration } from '../../domain/assets/AssetSummary';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing } from '../theme/tokens';
import { AppTextInput } from './AppTextInput';
import { NativeSegmentedControl } from './NativeSegmentedControl';

export function ExpirationField({ initialValue, initialPickerDate, disabled = false, onChange }: {
  readonly initialValue?: AssetExpiration;
  readonly initialPickerDate: Date;
  readonly disabled?: boolean;
  readonly onChange: (value: AssetExpiration | undefined, valid: boolean) => void;
}) {
  const colors = useAppearancePalette();
  const [editing, setEditing] = useState(false);
  const [precision, setPrecision] = useState<AssetExpiration['precision']>(initialValue?.precision ?? 'day');
  const [day, setDay] = useState(initialValue?.precision === 'day' ? initialValue.date : '');
  const [month, setMonth] = useState(initialValue?.precision === 'month' ? initialValue.date.slice(5) : '');
  const [year, setYear] = useState(initialValue?.precision === 'month' ? initialValue.date.slice(0, 4) : '');
  const [pickerOpen, setPickerOpen] = useState(false);
  const [pickerDate, setPickerDate] = useState(initialPickerDate);
  const monthValid = validMonthInput(month, year);
  const monthCalendarNotice = expirationMonthCalendarNotice();

  function publishMonth(nextMonth: string, nextYear: string) {
    const empty = !nextMonth && !nextYear;
    const valid = validMonthInput(nextMonth, nextYear);
    onChange(!empty && valid ? { date: `${nextYear}-${nextMonth.padStart(2, '0')}`, precision: 'month' } : undefined, valid);
  }
  function selectPrecision(next: AssetExpiration['precision']) {
    if (disabled || next === precision) return;
    setPrecision(next);
    setPickerOpen(false);
    if (next === 'month') {
      const nextMonth = day ? day.slice(5, 7) : month;
      const nextYear = day ? day.slice(0, 4) : year;
      setMonth(nextMonth); setYear(nextYear); publishMonth(nextMonth, nextYear);
    } else { setDay(''); onChange(undefined, !month && !year); }
  }
  function currentPickerDate() {
    const date = new Date(initialPickerDate);
    if (day) {
      const [y, m, d] = day.split('-').map(Number);
      date.setFullYear(y, m - 1, d);
    }
    return date;
  }
  function openPicker() {
    setPickerDate(currentPickerDate());
    setPickerOpen(true);
  }
  function commitDate(date: Date) {
    const y = date.getFullYear();
    if (!Number.isFinite(date.getTime()) || y < 1 || y > 9999) return;
    const value = `${String(y).padStart(4, '0')}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
    setDay(value);
    if (Platform.OS === 'android') setPickerOpen(false);
    onChange({ date: value, precision: 'day' }, true);
  }
  function pickerChanged(event: DateTimePickerEvent, date?: Date) {
    if (disabled) return;
    if (event.type !== 'set') { setPickerOpen(false); return; }
    if (!date) return;
    commitDate(date);
  }
  function clear() {
    if (disabled) return;
    setDay(''); setMonth(''); setYear(''); setEditing(false);
    setPickerOpen(false);
    onChange(undefined, true);
  }
  const current = precision === 'day' ? (day ? { date: day, precision } : undefined) : (month && year && monthValid ? { date: `${year}-${month.padStart(2, '0')}`, precision } : undefined);
  return <SelectionRow label="Expiration" value={current ? formatAssetExpiration(current) : 'Not set'} expanded={editing} disabled={disabled} onPress={() => setEditing(value => !value)}>
    <View style={styles.field}>
    <NativeSegmentedControl colors={colors} disabled={disabled} value={precision} onChange={selectPrecision}
      segments={[{ value: 'day', label: 'Exact date' }, { value: 'month', label: 'Month and year' }]} />
    {precision === 'month' ? <>
      {monthCalendarNotice ? <Text style={{ color: colors.textMuted }}>{monthCalendarNotice}</Text> : null}
      <NativeChoicePicker label="Month" accessibilityLabel="Expiration month" value={month ? String(Number(month)) : ''} disabled={disabled} options={expirationMonthOptions()} onChange={value => { setMonth(value); publishMonth(value, year); }} />
      <Text style={{ color: colors.text }}>Year</Text>
      <AppTextInput accessibilityLabel="Expiration year" editable={!disabled} keyboardType="number-pad" value={year} placeholder="YYYY" style={[styles.input, { color: colors.text, borderColor: colors.controlBorder }]} onChangeText={(value) => { if (disabled) return; setYear(value); publishMonth(month, value); }} />
      <Text accessibilityLiveRegion="polite" style={{ color: colors.textMuted }}>{monthValid ? 'Tracked through the end of this month.' : 'Enter a month from 1 to 12 and a four-digit year.'}</Text>
    </> : <>
      {Platform.OS === 'ios' ? day ?
        <DateTimePicker accessibilityLabel="Expiration date" disabled={disabled} mode="date" display="compact" value={currentPickerDate()} onChange={pickerChanged} /> :
        <NativeCommandButton label="Add expiration date" disabled={disabled} onPress={() => { if (!disabled) commitDate(initialPickerDate); }} /> : <>
        <NativeCommandButton label="Choose expiration date" disabled={disabled} onPress={openPicker} />
        {pickerOpen && !disabled ? <DateTimePicker mode="date" display="default" value={pickerDate} onChange={pickerChanged} /> : null}
      </>}
      <Text style={{ color: colors.textMuted }}>Tracked through the end of this day.</Text>
    </>}
    {(day || month || year) ? <NativeCommandButton label="Clear expiration" disabled={disabled} onPress={clear} /> : null}
  </View></SelectionRow>;
}
const styles = StyleSheet.create({
  field: { gap: spacing.sm, flexShrink: 0, paddingVertical: spacing.sm },
  input: { minHeight: 44, borderWidth: 1, borderRadius: radius.md, padding: spacing.sm }
});

function validMonthInput(month: string, year: string): boolean {
  return (!month && !year) || (/^\d{1,2}$/.test(month) && Number(month) >= 1 && Number(month) <= 12 && /^\d{4}$/.test(year) && Number(year) >= 1);
}
