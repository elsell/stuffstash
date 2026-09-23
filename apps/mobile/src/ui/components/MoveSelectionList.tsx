import { ScrollView } from 'react-native';
import { SettingsChoiceRow, SettingsSection, SettingsValueRow } from '../screens/SettingsList';
import { useFocusedSheetActions } from './useFocusedSheetActions';
import type { MoveSelectionListProps, MoveSelectionRowModel } from './MoveSelectionList.types';

export function MoveSelectionList(props: MoveSelectionListProps) {
  return <ScrollView style={{ flex: 1 }} contentInsetAdjustmentBehavior="automatic" keyboardShouldPersistTaps="handled">
    <SettingsSection title={props.subjectLabel}>
      <SettingsValueRow label={props.subject} value={props.context} />
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
