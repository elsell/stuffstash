import { useState } from 'react';
import { formatAssetExpiration } from '../presentation/ExpirationPresentation';
import { Platform, View } from 'react-native';
import DateTimePicker from '@react-native-community/datetimepicker';
import { SettingsActionRow, SettingsNavigationRow, SettingsSection, SettingsSwitchRow } from '../screens/SettingsList';
function calendarDate(date: Date) { return `${String(date.getFullYear()).padStart(4,'0')}-${String(date.getMonth()+1).padStart(2,'0')}-${String(date.getDate()).padStart(2,'0')}`; }
export function ExpirationDateRange({ fromDate, throughDate, onChange }: { readonly fromDate?: string; readonly throughDate?: string; readonly onChange: (range: { fromDate?: string; throughDate?: string }) => void }) {
 const [editing, setEditing] = useState<'fromDate' | 'throughDate' | null>(null);
 const range = { fromDate, throughDate };
 return <SettingsSection footer="Month-only dates are included by their month end.">
  {(['fromDate','throughDate'] as const).map(key => <View key={key}>
   <SettingsSwitchRow label={key === 'fromDate' ? 'From date' : 'Through date'} value={!!range[key]} onValueChange={enabled => { onChange({...range,[key]:enabled ? calendarDate(new Date()) : undefined}); if (Platform.OS !== 'ios') setEditing(enabled ? key : null); }} />
   {range[key] && Platform.OS !== 'ios' ? <SettingsNavigationRow accessibilityLabel={key === 'fromDate' ? 'Change first expiration date' : 'Change last expiration date'} label="Date" value={formatAssetExpiration({date:range[key]!,precision:'day'})} onPress={() => setEditing(key)} /> : null}
   {range[key] && (Platform.OS === 'ios' || editing === key) ? <DateTimePicker accessibilityLabel={key === 'fromDate' ? 'First expiration date' : 'Last expiration date'} value={new Date(`${range[key]}T12:00:00`)} mode="date" display={Platform.OS === 'ios' ? 'compact' : 'default'} style={{alignSelf:'flex-end',marginHorizontal:16,marginBottom:12}} onChange={(event, selected) => {
    if (Platform.OS !== 'ios') setEditing(null);
    if (event.type === 'set' && selected) onChange({...range,[key]:calendarDate(selected)});
   }} /> : null}
  </View>)}
  <SettingsActionRow label="Clear date range" onPress={() => onChange({fromDate:undefined,throughDate:undefined})} />
 </SettingsSection>;
}
