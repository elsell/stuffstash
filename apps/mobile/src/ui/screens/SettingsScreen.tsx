import { useEffect, useMemo, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';
import { Activity, AudioLines, Info, Server, Share2, SunMedium, UserRound } from 'lucide-react-native';
import type { ProviderProfileSettingsQuery } from '../../application/providerProfiles/ProviderProfileSettingsQuery';
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
  providerProfileSettingsQuery,
  settingsQuery
}: {
  readonly onNavigate: (destination: SettingsDestination) => void;
  readonly providerProfileSettingsQuery: ProviderProfileSettingsQuery;
  readonly settingsQuery: SettingsQuery;
}) {
  const { preference } = useAppearance();
  const { palette, styles } = useSettingsListStyles();
  const { load, state } = useSettingsModel(settingsQuery);
  const [voiceReadiness, setVoiceReadiness] = useState<string>();

  const canConfigure = state.status === 'ready' &&
    state.settings.selectedTenant.permissions.includes('configure');
  useEffect(() => {
    let current = true;
    if (!canConfigure) {
      setVoiceReadiness(undefined);
      return () => { current = false; };
    }
    providerProfileSettingsQuery.execute().then((result) => {
      if (current) setVoiceReadiness(result.configuration.readiness);
    }).catch(() => {
      if (current) setVoiceReadiness('unavailable');
    });
    return () => { current = false; };
  }, [canConfigure, providerProfileSettingsQuery, state.status === 'ready' ? state.settings.selectedTenant.id : undefined]);

  const sections = useMemo(() => state.status === 'ready'
    ? buildSettingsRootSections({
        ...state.settings,
        appearance: preference,
        voiceReadiness
      })
    : [], [preference, state, voiceReadiness]);

  if (state.status === 'loading') {
    return (
      <View style={[styles.shell, styles.errorContainer]}>
        <ActivityIndicator color={palette.action} />
        <Text style={styles.errorMessage}>Loading settings</Text>
      </View>
    );
  }
  if (state.status === 'error') {
    return (
      <ScrollView contentContainerStyle={styles.errorContainer} style={styles.shell}>
        <Text accessibilityRole="header" style={styles.errorTitle}>Could not load Settings</Text>
        <Text style={styles.errorMessage}>{state.message}</Text>
        <Pressable accessibilityRole="button" onPress={() => void load()} style={styles.retryButton}>
          <Text style={styles.retryText}>Retry</Text>
        </Pressable>
      </ScrollView>
    );
  }

  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      {sections.map((section) => (
        <SettingsSection key={section.id} title={section.title}>
          {section.rows.map((row, index) => (
            <View key={row.id}>
              {index > 0 ? <SettingsSeparator /> : null}
              <SettingsNavigationRow
                accessibilityLabel={row.accessibilityLabel}
                context={row.context}
                icon={iconForRow(row.id, palette.action)}
                label={row.label}
                onPress={() => onNavigate(row.destination)}
                value={row.value}
              />
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
    case 'sharing': return <Share2 {...props} />;
    case 'voice-setup': return <AudioLines {...props} />;
    case 'server': return <Server {...props} />;
    case 'about': return <Info {...props} />;
    default: return <Activity {...props} />;
  }
}
