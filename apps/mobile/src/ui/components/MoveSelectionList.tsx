import { ScrollView, Text, View } from 'react-native';
import { SettingsChoiceRow, SettingsSection, useSettingsListStyles } from '../screens/SettingsList';
import { useFocusedSheetActions } from './useFocusedSheetActions';
import type { MoveSelectionListProps, MoveSelectionRowModel } from './MoveSelectionList.types';

export function MoveSelectionList(props: MoveSelectionListProps) {
  const { styles } = useSettingsListStyles();
  return <ScrollView style={{ flex: 1 }} contentInsetAdjustmentBehavior="automatic" keyboardShouldPersistTaps="handled">
    <SettingsSection title={props.subjectLabel}>
      <View style={styles.navigationRow}>
        <Text style={styles.rowLabel}>{props.subject}</Text>
        <Text style={styles.rowContext}>{props.context}</Text>
      </View>
    </SettingsSection>
    {props.retainedSelection ? <SettingsSection title="Selected"><Choice row={props.retainedSelection} /></SettingsSection> : null}
    <SettingsSection title={props.title}>
      {props.status}
      {props.rows.map(row => <Choice key={row.id} row={row} />)}
    </SettingsSection>
  </ScrollView>;
}
function Choice({ row }: { readonly row: MoveSelectionRowModel }) {
  const actions = useFocusedSheetActions({ primaryLabel: row.accessibilityLabel, secondaryLabel: '',
    disabled: !!row.disabled, onApply: row.onPress, onBack: () => {} });
  return <SettingsChoiceRow label={row.label} context={row.context} accessibilityLabel={row.accessibilityLabel}
    disabled={row.disabled} selected={row.selected} onPress={actions.onApply} />;
}
