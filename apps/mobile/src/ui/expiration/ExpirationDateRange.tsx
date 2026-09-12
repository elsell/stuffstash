import { useState } from 'react';
import { Platform } from 'react-native';
import DateTimePicker from '@react-native-community/datetimepicker';
import { SettingsActionRow, SettingsNavigationRow, SettingsSection } from '../screens/SettingsList';
import { formatAssetExpiration } from '../presentation/ExpirationPresentation';
export function ExpirationDateRange({ fromDate, throughDate, onChange }: { readonly fromDate?: string; readonly throughDate?: string; readonly onChange: (range: { fromDate?: string; throughDate?: string }) => void }) {
 const [editing, setEditing] = useState<'fromDate' | 'throughDate' | null>(null);
 const value = editing === 'fromDate' ? fromDate : throughDate;
 const date = value ? new Date(`${value}T12:00:00`) : new Date();
 const label = (value?: string) => value ? formatAssetExpiration({ date: value, precision: 'day' }) : 'Any';
 return <SettingsSection title="Expiration dates" footer="Month-only dates are included by their month end.">
  <SettingsNavigationRow accessibilityLabel="Choose from" label="From" value={label(fromDate)} onPress={() => setEditing('fromDate')} />
  <SettingsNavigationRow accessibilityLabel="Choose through" label="Through" value={label(throughDate)} onPress={() => setEditing('throughDate')} />
  {editing ? <DateTimePicker accessibilityLabel={editing === 'fromDate' ? 'First expiration date' : 'Last expiration date'} value={date} mode="date" display={Platform.OS === 'ios' ? 'inline' : 'default'} onChange={(event, selected) => {
   if (Platform.OS !== 'ios') setEditing(null);
   if (event.type !== 'set' || !selected) return;
   const date = `${String(selected.getFullYear()).padStart(4, '0')}-${String(selected.getMonth() + 1).padStart(2, '0')}-${String(selected.getDate()).padStart(2, '0')}`;
   onChange({ fromDate, throughDate, [editing]: date });
  }} /> : null}
  <SettingsActionRow label="Clear date range" onPress={() => { setEditing(null); onChange({ fromDate: undefined, throughDate: undefined }); }} />
 </SettingsSection>;
}
