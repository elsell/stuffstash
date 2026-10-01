import { t } from '../../presentation/localization';
import { useState } from 'react';
import { formatAssetExpiration } from '../presentation/ExpirationPresentation';
import { Platform, View } from 'react-native';
import DateTimePicker from '@react-native-community/datetimepicker';
import { SettingsActionRow, SettingsNavigationRow, SettingsSection, SettingsSwitchRow } from '../screens/SettingsList';
function calendarDate(date: Date) { return `${String(date.getFullYear()).padStart(4,'0')}-${String(date.getMonth()+1).padStart(2,'0')}-${String(date.getDate()).padStart(2,'0')}`; }
export function ExpirationDateRange({ fromDate, throughDate, onChange }: { readonly fromDate?: string; readonly throughDate?: string; readonly onChange: (range: { fromDate?: string; throughDate?: string }) => void }) {
 const [editing, setEditing] = useState<'fromDate' | 'throughDate' | null>(null);
 const range = { fromDate, throughDate };
 return <SettingsSection footer={t('mobile.ExpirationDateRange.monthOnlyDatesAreIncludedByTheirMonthEnd')}>
  {(['fromDate','throughDate'] as const).map(key => <View key={key}>
   <SettingsSwitchRow label={key === 'fromDate' ? t('mobile.ExpirationDateRange.fromDate') : t('mobile.ExpirationDateRange.throughDate')} value={!!range[key]} onValueChange={enabled => { onChange({...range,[key]:enabled ? calendarDate(new Date()) : undefined}); if (Platform.OS !== 'ios') setEditing(enabled ? key : null); }} />
   {range[key] && Platform.OS !== 'ios' ? <SettingsNavigationRow accessibilityLabel={key === 'fromDate' ? t('mobile.ExpirationDateRange.changeFirstExpirationDate') : t('mobile.ExpirationDateRange.changeLastExpirationDate')} label={t('mobile.ExpirationDateRange.date')} value={formatAssetExpiration({date:range[key]!,precision:'day'})} onPress={() => setEditing(key)} /> : null}
   {range[key] && (Platform.OS === 'ios' || editing === key) ? <DateTimePicker accessibilityLabel={key === 'fromDate' ? t('mobile.ExpirationDateRange.firstExpirationDate') : t('mobile.ExpirationDateRange.lastExpirationDate')} value={new Date(`${range[key]}T12:00:00`)} mode="date" display={Platform.OS === 'ios' ? 'compact' : 'default'} style={{alignSelf:'flex-end',marginHorizontal:16,marginBottom:12}} onChange={(event, selected) => {
    if (Platform.OS !== 'ios') setEditing(null);
    if (event.type === 'set' && selected) onChange({...range,[key]:calendarDate(selected)});
   }} /> : null}
  </View>)}
  <SettingsActionRow label={t('mobile.ExpirationDateRange.clearDateRange')} onPress={() => onChange({fromDate:undefined,throughDate:undefined})} />
 </SettingsSection>;
}
