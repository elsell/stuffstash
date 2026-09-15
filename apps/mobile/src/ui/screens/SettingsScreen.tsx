import { NativeCommandButton } from '../components/NativeCommandButton';
import { AppearancePicker } from '../components/AppearancePicker';
import { SettingsRefreshNotice } from './SettingsRefreshNotice';
import { useMemo } from 'react';
import { ActivityIndicator, ScrollView, Text, View } from 'react-native';
import { Activity, Boxes, House, Info, Server, SunMedium, UserRound } from 'lucide-react-native';
import type { SettingsQuery } from '../../application/settings/SettingsQuery';
import { useAppearance } from '../theme/AppearanceContext';
import {
  SettingsNavigationRow,
  SettingsSection,
  SettingsSeparator,
  useSettingsListStyles
} from './SettingsList';
import {
  buildSettingsRootSections,
  type SettingsDestination
} from './SettingsScreenPresentation';
import { useSettingsModel } from './SettingsScreenState';

export function SettingsScreen({
  onNavigate,
  settingsQuery
}: {
  readonly onNavigate: (destination: SettingsDestination) => void;
  readonly settingsQuery: SettingsQuery;
}) {
  const { preference } = useAppearance();
  const { palette, styles } = useSettingsListStyles();
  const { load, state, hasRefreshError } = useSettingsModel(settingsQuery);
  const sections = useMemo(() => state.status === 'ready'
    ? buildSettingsRootSections({
        ...state.settings,
        appearance: preference
      })
    : [], [preference, state]);

  if (state.status === 'loading') {
    return (
      <ScrollView style={styles.shell} contentContainerStyle={styles.errorContainer}>
        <ActivityIndicator color={palette.action} />
        <Text style={styles.errorMessage}>Loading settings</Text>
        <SettingsRecoveryLinks onNavigate={onNavigate} />
      </ScrollView>
    );
  }
  if (state.status === 'error') {
    return (
      <ScrollView contentContainerStyle={styles.errorContainer} style={styles.shell}>
        <Text accessibilityRole="header" style={styles.errorTitle}>Could not load Settings</Text>
        <Text style={styles.errorMessage}>{state.message}</Text>
        <NativeCommandButton label="Retry" onPress={() => void load()} />
        <SettingsRecoveryLinks onNavigate={onNavigate} />
      </ScrollView>
    );
  }

  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <SettingsRefreshNotice visible={hasRefreshError} onRetry={load} />
      {sections.map((section) => (
        <SettingsSection key={section.id} title={section.title}>
          {section.rows.map((row, index) => (
            <View key={row.id}>
              {index > 0 ? <SettingsSeparator /> : null}
              {row.id === 'appearance' ? <AppearancePicker /> : <SettingsNavigationRow
                accessibilityLabel={row.accessibilityLabel}
                context={row.context}
                icon={iconForRow(row.id, palette.action)}
                label={row.label}
                onPress={() => onNavigate(row.destination)}
                value={row.value}
              />}
            </View>
          ))}
        </SettingsSection>
      ))}
    </ScrollView>
  );
}

function iconForRow(id: string, color: string) {
  const props = { color, size: 20, strokeWidth: 2.2 };
  switch (id) {
    case 'account': return <UserRound {...props} />;
    case 'appearance': return <SunMedium {...props} />;
    case 'tenant-settings': return <House {...props} />;
    case 'inventory-settings': return <Boxes {...props} />;
    case 'server': return <Server {...props} />;
    case 'about': return <Info {...props} />;
    default: return <Activity {...props} />;
  }
}

function SettingsRecoveryLinks({ onNavigate }: { readonly onNavigate: (destination: SettingsDestination) => void }) {
  return <View style={{ alignSelf: 'stretch' }}><SettingsSection>
    <SettingsNavigationRow accessibilityLabel="Open Account settings" label="Account" onPress={() => onNavigate('account')} />
    <SettingsSeparator />
    <SettingsNavigationRow accessibilityLabel="Open Stuff Stash server settings" label="Stuff Stash server" onPress={() => onNavigate('connection')} />
  </SettingsSection></View>;
}
