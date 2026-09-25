import { ScrollView, Text, View } from 'react-native';
import { SettingsChoiceRow, SettingsSection, useSettingsListStyles } from '../screens/SettingsList';
import { NativeCommandButton } from './NativeCommandButton';
import { useFocusedSheetActions } from './useFocusedSheetActions';
import type { MoveSelectionListProps, MoveSelectionRowModel, MoveSelectionStatus } from './MoveSelectionList.types';

export function MoveSelectionList(props: MoveSelectionListProps) {
  const { styles } = useSettingsListStyles();
  return <ScrollView style={{ flex: 1 }} contentInsetAdjustmentBehavior="automatic" keyboardShouldPersistTaps="handled">
    <SettingsSection title={props.subjectLabel}>
      <View style={styles.navigationRow}>
        <Text style={styles.rowLabel}>{props.subject}</Text>
        <Text style={styles.rowContext}>{props.context}</Text>
      </View>
    </SettingsSection>
    {props.destinationLabel ? <SettingsSection title="Move to"><View style={styles.navigationRow}><Text style={styles.rowLabel}>{props.destinationLabel}</Text></View></SettingsSection> : null}
    <SettingsSection title={props.title}>
      {props.statuses?.map((status, index) => <Status key={index} status={status} />)}
      {props.retainedSelection ? <Choice row={props.retainedSelection} /> : null}
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

function Status({ status }: { readonly status: MoveSelectionStatus }) {
  const { styles } = useSettingsListStyles();
  const actions = useFocusedSheetActions({ primaryLabel: status.retry?.label ?? '', secondaryLabel: '',
    disabled: !status.retry || !!status.retry.disabled, onApply: () => status.retry?.onPress(), onBack: () => {} });
  return <View style={styles.navigationRow}>
    {status.title ? <Text style={styles.rowLabel}>{status.title}</Text> : null}
    <Text accessibilityLiveRegion="polite" style={styles.rowContext}>{status.message}</Text>
    {status.retry ? <NativeCommandButton label={status.retry.label} disabled={actions.disabled} onPress={actions.onApply} /> : null}
  </View>;
}
