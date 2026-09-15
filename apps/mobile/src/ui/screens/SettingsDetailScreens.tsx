import { NativeCommandButton } from '../components/NativeCommandButton';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileServerQuery } from '../serverState/useMobileServerQuery';
import { SettingsRefreshNotice } from './SettingsRefreshNotice';
import { useRef, useState, type ReactNode } from 'react';
import { ActivityIndicator, Alert, ScrollView, Text, View } from 'react-native';
import { AppearancePicker } from '../components/AppearancePicker';
import type { SettingsQuery, SettingsViewModel } from '../../application/settings/SettingsQuery';
import { useAppFeedback } from '../feedback/AppFeedback';
import {
  SettingsActionRow,
  SettingsSection,
  SettingsSeparator,
  SettingsValueRow,
  useSettingsListStyles
} from './SettingsList';
import { serverHostname } from './SettingsScreenPresentation';
import { useSettingsModel } from './SettingsScreenState';

export function AccountSettingsScreen({
  onSignOut,
  settingsQuery
}: {
  readonly onSignOut: () => Promise<void>;
  readonly settingsQuery: SettingsQuery;
}) {
  const feedback = useAppFeedback();
  const { styles } = useSettingsListStyles();
  const principal = useMobileServerQuery({ key: mobileQueryKeys.principal, query: signal => settingsQuery.getPrincipal({ signal }) });
  const principalLabel = principal.data?.email ?? 'Current account';
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
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <SettingsRefreshNotice visible={principal.isError} onRetry={async () => { await principal.refetch(); }} />
          <SettingsSection footer="Signing out keeps this server on your device so you can sign in again quickly.">
            <SettingsValueRow label="Signed in as" value={principalLabel} />
          </SettingsSection>
          <SettingsSection>
            <SettingsActionRow
              accessibilityLabel={`Sign out ${principalLabel}`}
              disabled={working}
              label={working ? 'Signing Out…' : 'Sign Out'}
              onPress={() => confirmSignOut(principalLabel, signOut)}
            />
          </SettingsSection>
    </ScrollView>
  );
}

export function AppearanceSettingsScreen() {
  const { styles } = useSettingsListStyles();
  return <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
    <SettingsSection footer="System follows your device’s appearance setting.">
      <AppearancePicker />
    </SettingsSection>
  </ScrollView>;
}

export function ConnectionSettingsScreen({
  onChangeServer,
  settingsQuery
}: {
  readonly onChangeServer: () => Promise<void>;
  readonly settingsQuery: SettingsQuery;
}) {
  const { styles } = useSettingsListStyles();
  const diagnostics = settingsQuery.getDiagnostics();
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
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <SettingsSection
        footer="This server determines where the app signs in and stores your Stuff Stash data."
        title="Current Server"
      >
        <SettingsValueRow label="Server" value={serverHostname(diagnostics.apiBaseUrl)} />
        <SettingsSeparator />
        <SettingsValueRow label="Address" value={diagnostics.apiBaseUrl} />
      </SettingsSection>
      <SettingsSection footer="Changing servers signs you out and forgets this server and household selection on this device. It does not delete data from the server.">
        <SettingsActionRow
          accessibilityLabel={`Change Stuff Stash server from ${serverHostname(diagnostics.apiBaseUrl)}`}
          disabled={working}
          label={working ? 'Changing Server…' : 'Change Server'}
          onPress={() => confirmChangeServer(diagnostics.apiBaseUrl, changeServer)}
        />
      </SettingsSection>
    </ScrollView>
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
  const diagnostics = settingsQuery.getDiagnostics();
  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <View style={styles.detailHeader}>
        <Text accessibilityRole="header" style={styles.detailTitle}>Stuff Stash</Text>
        <Text style={styles.detailSubtitle}>A calm, flexible home inventory for knowing what you have and where it lives.</Text>
      </View>
      <SettingsSection>
        <SettingsValueRow label="Version" value={diagnostics.appVersion} />
      </SettingsSection>
    </ScrollView>
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
  const { load, state, hasRefreshError } = useSettingsModel(query);
  if (state.status === 'loading') {
    return <View style={[styles.shell, styles.errorContainer]}><ActivityIndicator color={palette.action} /></View>;
  }
  if (state.status === 'error') {
    return (
      <ScrollView contentContainerStyle={styles.errorContainer} style={styles.shell}>
        <Text accessibilityRole="header" style={styles.errorTitle}>Could not load this setting</Text>
        <Text style={styles.errorMessage}>{state.message}</Text>
        <NativeCommandButton label="Retry" onPress={() => void load()} />
      </ScrollView>
    );
  }
  return <ScrollView contentContainerStyle={styles.content} style={styles.shell}><SettingsRefreshNotice visible={hasRefreshError} onRetry={load} />{children(state.settings)}</ScrollView>;
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
