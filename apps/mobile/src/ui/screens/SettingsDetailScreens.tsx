import { useRef, useState, type ReactNode } from 'react';
import { ActivityIndicator, Alert, Pressable, ScrollView, Text, View } from 'react-native';
import { Check } from 'lucide-react-native';
import type { SettingsQuery, SettingsViewModel } from '../../application/settings/SettingsQuery';
import {
  appearancePreferences,
  type AppearancePreference
} from '../../application/settings/AppearancePreference';
import { useAppFeedback } from '../feedback/AppFeedback';
import { useAppearance } from '../theme/AppearanceContext';
import {
  SettingsActionRow,
  SettingsSection,
  SettingsSeparator,
  SettingsValueRow,
  useSettingsListStyles
} from './SettingsList';
import { appearanceLabel, serverHostname } from './SettingsScreenPresentation';
import { useSettingsModel } from './SettingsScreenState';

export function AccountSettingsScreen({
  onSignOut,
  settingsQuery
}: {
  readonly onSignOut: () => Promise<void>;
  readonly settingsQuery: SettingsQuery;
}) {
  const feedback = useAppFeedback();
  const [working, setWorking] = useState(false);
  const workingRef = useRef(false);

  async function signOut(): Promise<void> {
    if (workingRef.current) return;
    workingRef.current = true;
    setWorking(true);
    try {
      await onSignOut();
    } catch (error) {
      feedback.showNotice({
        tone: 'error',
        title: 'Could not sign out',
        message: readableError(error)
      });
      workingRef.current = false;
      setWorking(false);
    }
  }

  return (
    <SettingsModelScreen query={settingsQuery}>
      {(settings) => (
        <>
          <SettingsSection footer="Signing out keeps this server on your device so you can sign in again quickly.">
            <SettingsValueRow label="Signed in as" value={settings.principal.primaryLabel} />
          </SettingsSection>
          <SettingsSection>
            <SettingsActionRow
              accessibilityLabel={`Sign out ${settings.principal.primaryLabel}`}
              disabled={working}
              label={working ? 'Signing Out…' : 'Sign Out'}
              onPress={() => confirmSignOut(settings.principal.primaryLabel, signOut)}
            />
          </SettingsSection>
        </>
      )}
    </SettingsModelScreen>
  );
}

export function AppearanceSettingsScreen() {
  const { preference, setPreference } = useAppearance();
  const feedback = useAppFeedback();
  const { palette, styles } = useSettingsListStyles();

  async function select(next: AppearancePreference): Promise<void> {
    if (next === preference) return;
    try {
      await setPreference(next);
    } catch {
      feedback.showNotice({
        tone: 'error',
        title: 'Appearance not saved',
        message: 'Stuff Stash could not save the appearance setting.'
      });
    }
  }

  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <SettingsSection footer="System follows the appearance selected in iPhone Settings.">
        {appearancePreferences.map((option, index) => {
          const selected = option === preference;
          const label = appearanceLabel(option);
          return (
            <View key={option}>
              {index > 0 ? <SettingsSeparator /> : null}
              <Pressable
                accessibilityLabel={`${label} appearance`}
                accessibilityRole="radio"
                accessibilityState={{ checked: selected }}
                onPress={() => void select(option)}
                style={({ pressed }) => [styles.choiceRow, pressed && styles.navigationRowPressed]}
              >
                <View style={styles.navigationRowContent}>
                  <Text style={styles.rowLabel}>{label}</Text>
                  {selected ? <Check color={palette.action} size={22} /> : null}
                </View>
              </Pressable>
            </View>
          );
        })}
      </SettingsSection>
    </ScrollView>
  );
}

export function ConnectionSettingsScreen({
  onChangeServer,
  settingsQuery
}: {
  readonly onChangeServer: () => Promise<void>;
  readonly settingsQuery: SettingsQuery;
}) {
  const feedback = useAppFeedback();
  const [working, setWorking] = useState(false);
  const workingRef = useRef(false);

  async function changeServer(): Promise<void> {
    if (workingRef.current) return;
    workingRef.current = true;
    setWorking(true);
    try {
      await onChangeServer();
    } catch (error) {
      feedback.showNotice({
        tone: 'error',
        title: 'Could not change server',
        message: readableError(error)
      });
      workingRef.current = false;
      setWorking(false);
    }
  }

  return (
    <SettingsModelScreen query={settingsQuery}>
      {(settings) => (
        <>
          <SettingsSection
            footer="This server determines where the app signs in and stores your Stuff Stash data."
            title="Current Server"
          >
            <SettingsValueRow label="Server" value={serverHostname(settings.serverUrl)} />
            <SettingsSeparator />
            <SettingsValueRow label="Address" value={settings.serverUrl} />
          </SettingsSection>
          <SettingsSection footer="Changing servers signs you out and forgets this server and household selection on this device. It does not delete data from the server.">
            <SettingsActionRow
              accessibilityLabel={`Change Stuff Stash server from ${serverHostname(settings.serverUrl)}`}
              disabled={working}
              label={working ? 'Changing Server…' : 'Change Server'}
              onPress={() => confirmChangeServer(settings.serverUrl, changeServer)}
            />
          </SettingsSection>
        </>
      )}
    </SettingsModelScreen>
  );
}

export function DiagnosticsSettingsScreen({ settingsQuery }: { readonly settingsQuery: SettingsQuery }) {
  return (
    <SettingsModelScreen query={settingsQuery}>
      {(settings) => (
        <>
          <SettingsSection title="Connection">
            <SettingsValueRow label="API URL" value={settings.serverUrl} />
            <SettingsSeparator />
            <SettingsValueRow label="Authentication" value={authenticationLabel(settings.authenticationMode)} />
          </SettingsSection>
          <SettingsSection title="Identity">
            <SettingsValueRow label="Principal ID" value={settings.principal.id} />
            <SettingsSeparator />
            <SettingsValueRow label="Tenant ID" value={settings.selectedTenant.id} />
          </SettingsSection>
          <SettingsSection title="Application">
            <SettingsValueRow label="Version" value={settings.appVersion} />
          </SettingsSection>
        </>
      )}
    </SettingsModelScreen>
  );
}

export function AboutSettingsScreen({ settingsQuery }: { readonly settingsQuery: SettingsQuery }) {
  const { styles } = useSettingsListStyles();
  return (
    <SettingsModelScreen query={settingsQuery}>
      {(settings) => (
        <>
          <View style={styles.detailHeader}>
            <Text accessibilityRole="header" style={styles.detailTitle}>Stuff Stash</Text>
            <Text style={styles.detailSubtitle}>A calm, flexible home inventory for knowing what you have and where it lives.</Text>
          </View>
          <SettingsSection>
            <SettingsValueRow label="Version" value={settings.appVersion} />
          </SettingsSection>
        </>
      )}
    </SettingsModelScreen>
  );
}

function SettingsModelScreen({
  children,
  query
}: {
  readonly children: (settings: SettingsViewModel) => ReactNode;
  readonly query: SettingsQuery;
}) {
  const { palette, styles } = useSettingsListStyles();
  const { load, state } = useSettingsModel(query);
  if (state.status === 'loading') {
    return <View style={[styles.shell, styles.errorContainer]}><ActivityIndicator color={palette.action} /></View>;
  }
  if (state.status === 'error') {
    return (
      <ScrollView contentContainerStyle={styles.errorContainer} style={styles.shell}>
        <Text accessibilityRole="header" style={styles.errorTitle}>Could not load this setting</Text>
        <Text style={styles.errorMessage}>{state.message}</Text>
        <Pressable accessibilityRole="button" onPress={() => void load()} style={styles.retryButton}>
          <Text style={styles.retryText}>Retry</Text>
        </Pressable>
      </ScrollView>
    );
  }
  return <ScrollView contentContainerStyle={styles.content} style={styles.shell}>{children(state.settings)}</ScrollView>;
}

function confirmSignOut(label: string, onSignOut: () => Promise<void>): void {
  Alert.alert('Sign out?', `You’ll need to sign in again as ${label}. This Stuff Stash server will stay saved on your device.`, [
    { text: 'Cancel', style: 'cancel' },
    { text: 'Sign Out', onPress: () => void onSignOut() }
  ]);
}

function confirmChangeServer(serverUrl: string, onChangeServer: () => Promise<void>): void {
  Alert.alert(
    'Change Stuff Stash server?',
    `You’ll be signed out of ${serverHostname(serverUrl)}, and this device will forget its saved server and household selection. Your Stuff Stash data won’t be deleted.`,
    [
      { text: 'Cancel', style: 'cancel' },
      { text: 'Change Server', onPress: () => void onChangeServer() }
    ]
  );
}

function authenticationLabel(value: SettingsViewModel['authenticationMode']): string {
  return value === 'oidc-sso' ? 'OIDC SSO' : 'Not configured';
}

function readableError(error: unknown): string {
  return error instanceof Error ? error.message : 'The action failed safely. Try again.';
}
