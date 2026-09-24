import { useRef, useState } from 'react';
import { ScrollView, Text } from 'react-native';
import type { ExpirationReminderPolicy } from '../src/domain/notifications/Notification';
import { ExpirationReminderEditor } from '../src/ui/components/ExpirationReminderEditor';
import { SettingsActionRow, SettingsSection, useSettingsListStyles } from '../src/ui/screens/SettingsList';

/** Isolated native composition; never changes device permissions or user data. */
export function SettingsCommandFixture() {
  const { styles } = useSettingsListStyles();
  const [policy, setPolicy] = useState<ExpirationReminderPolicy | null>(null);
  const failNext = useRef(true);
  const [activations, setActivations] = useState(0);
  return <ScrollView style={styles.shell} contentContainerStyle={styles.content}
    contentInsetAdjustmentBehavior="automatic">
    <ExpirationReminderEditor initialPolicy={policy}
      inheritedPolicy={{ enabled: true, upcoming: true, expired: true, advanceDays: 7 }}
      onSave={async value => {
        if (failNext.current) { failNext.current = false; throw new Error('Fixture save unavailable'); }
        setPolicy(value);
      }} onEditDays={() => {}} />
    <SettingsSection title="Saved state">
      <Text>{`Saved reminders: ${policy === null ? 'defaults' : policy.enabled ? 'custom' : 'off'}`}</Text>
      <SettingsActionRow label="Fail next save" onPress={() => { failNext.current = true; }} />
    </SettingsSection>
    <SettingsSection footer="Layout probe only; this does not open system settings.">
      <SettingsActionRow label="Open device settings" onPress={() => setActivations(value => value + 1)} />
      <Text>{`Device settings activations: ${activations}`}</Text>
    </SettingsSection>
  </ScrollView>;
}
