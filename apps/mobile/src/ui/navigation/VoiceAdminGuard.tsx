import type { ReactNode } from 'react';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';
import type { SettingsQuery } from '../../application/settings/SettingsQuery';
import { useSettingsListStyles } from '../screens/SettingsList';
import {
  type SettingsLoadState,
  useSettingsModel
} from '../screens/SettingsScreenState';

export type VoiceAdminAccessDecision =
  | { readonly status: 'loading' }
  | { readonly status: 'allowed' }
  | { readonly status: 'unavailable'; readonly tenantName: string }
  | { readonly status: 'error'; readonly message: string };

export type VoiceAdminGuardPresentation = {
  readonly title: string;
  readonly message: string;
  readonly retryLabel: string;
};

export function decideVoiceAdminAccess(state: SettingsLoadState): VoiceAdminAccessDecision {
  if (state.status !== 'ready') {
    return state;
  }

  if (!state.settings.selectedTenant.permissions.includes('configure')) {
    return {
      status: 'unavailable',
      tenantName: state.settings.selectedTenant.name
    };
  }

  return { status: 'allowed' };
}

export function voiceAdminGuardPresentation(
  decision: Extract<VoiceAdminAccessDecision, { status: 'error' | 'unavailable' }>
): VoiceAdminGuardPresentation {
  if (decision.status === 'unavailable') {
    return {
      title: 'Voice settings unavailable',
      message: `Only tenant administrators can configure Voice for ${decision.tenantName}.`,
      retryLabel: 'Check Again'
    };
  }

  return {
    title: 'Could not verify Voice settings access',
    message: decision.message,
    retryLabel: 'Retry'
  };
}

export function VoiceAdminGuard({
  children,
  settingsQuery
}: {
  readonly children: ReactNode;
  readonly settingsQuery: SettingsQuery;
}) {
  const { load, state } = useSettingsModel(settingsQuery);
  const { palette, styles } = useSettingsListStyles();
  const decision = decideVoiceAdminAccess(state);

  if (decision.status === 'loading') {
    return (
      <View style={[styles.shell, styles.errorContainer]}>
        <ActivityIndicator color={palette.action} />
        <Text style={styles.errorMessage}>Checking Voice settings access</Text>
      </View>
    );
  }

  if (decision.status === 'allowed') {
    return children;
  }

  const presentation = voiceAdminGuardPresentation(decision);
  return (
    <ScrollView contentContainerStyle={styles.errorContainer} style={styles.shell}>
      <Text accessibilityRole="header" style={styles.errorTitle}>{presentation.title}</Text>
      <Text style={styles.errorMessage}>{presentation.message}</Text>
      <Pressable
        accessibilityLabel={presentation.retryLabel}
        accessibilityRole="button"
        onPress={() => void load()}
        style={styles.retryButton}
      >
        <Text style={styles.retryText}>{presentation.retryLabel}</Text>
      </Pressable>
    </ScrollView>
  );
}
