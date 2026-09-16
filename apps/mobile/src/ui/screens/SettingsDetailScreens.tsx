import { NativeCommandButton } from '../components/NativeCommandButton';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileServerQuery } from '../serverState/useMobileServerQuery';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';
import { isAccessFailure } from '../serverState/isAccessFailure';
import { SettingsRefreshNotice } from './SettingsRefreshNotice';
import { useRef, useState } from 'react';
import { Alert, ScrollView, Text, View } from 'react-native';
import { AppearancePicker } from '../components/AppearancePicker';
import type { SettingsQuery, SettingsViewModel } from '../../application/settings/SettingsQuery';
import { useAppFeedback } from '../feedback/AppFeedback';
import {
  SettingsSection,
  SettingsSeparator,
  SettingsValueRow,
  SettingsLoadingRow,
  useSettingsListStyles
} from './SettingsList';
import { serverHostname } from './SettingsScreenPresentation';
import { useTaskPresentation } from '../navigation/useTaskPresentation';

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
  const capturePresentation = useTaskPresentation(settingsQuery, principal.data?.id ?? '');
  const [working, setWorking] = useState(false);
  const workingRef = useRef(false);

  async function signOut(canPresent: () => boolean): Promise<void> {
    if (workingRef.current) return;
    workingRef.current = true;
    setWorking(true);
    try {
      await onSignOut();
    } catch (error) {
      if (canPresent()) feedback.showNotice({
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
      <SettingsRefreshNotice visible={principal.isError} message={principal.data ? undefined : 'Could not load account details. You can retry or sign out.'} onRetry={async () => { await principal.refetch(); }} />
          <SettingsSection footer="Signing out keeps this server on your device so you can sign in again quickly.">
            <SettingsValueRow label="Signed in as" value={principalLabel} />
          </SettingsSection>
          <SettingsSection>
            <NativeCommandButton
              disabled={working}
              label={working ? 'Signing Out…' : 'Sign Out'}
              onPress={() => confirmSignOut(principalLabel, ownConfirmation(capturePresentation(), signOut))}
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
  const capturePresentation = useTaskPresentation(settingsQuery, diagnostics.apiBaseUrl);
  const feedback = useAppFeedback();
  const [working, setWorking] = useState(false);
  const workingRef = useRef(false);

  async function changeServer(canPresent: () => boolean): Promise<void> {
    if (workingRef.current) return;
    workingRef.current = true;
    setWorking(true);
    try {
      await onChangeServer();
    } catch (error) {
      if (canPresent()) feedback.showNotice({
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
        <NativeCommandButton
          disabled={working}
          label={working ? 'Changing Server…' : 'Change Server'}
          onPress={() => confirmChangeServer(diagnostics.apiBaseUrl, ownConfirmation(capturePresentation(), changeServer))}
        />
      </SettingsSection>
    </ScrollView>
  );
}

export function DiagnosticsSettingsScreen({ settingsQuery }: { readonly settingsQuery: SettingsQuery }) {
  const { styles } = useSettingsListStyles();
  const diagnostics = settingsQuery.getDiagnostics();
  const principal = useMobileServerQuery({ key: mobileQueryKeys.principal, query: signal => settingsQuery.getPrincipal({ signal }) });
  const scope = useMobileInventoryServerQuery({ key: mobileQueryKeys.settingsScope, query: signal => settingsQuery.getSelectedScope({ signal }) });
  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <SettingsSection title="Connection">
        <SettingsValueRow label="API URL" value={diagnostics.apiBaseUrl} />
        <SettingsSeparator />
        <SettingsValueRow label="Authentication" value={authenticationLabel(diagnostics.authenticationMode)} />
      </SettingsSection>
      <SettingsSection title="Identity">
        <DiagnosticIdentity label="Principal ID" task="account identity"
          value={isAccessFailure(principal.error) ? undefined : principal.data?.id}
          pending={principal.isPending} failed={principal.isError} retrying={principal.isFetching}
          onRetry={async () => { await principal.refetch({ cancelRefetch: false }); }} />
        <SettingsSeparator />
        <DiagnosticIdentity label="Tenant ID" task="household identity" value={scope.data?.tenant.id}
          pending={scope.isPending} failed={scope.isError} retrying={scope.isFetching}
          onRetry={async () => { await scope.refetch({ cancelRefetch: false }); }} />
      </SettingsSection>
      <SettingsSection title="Application">
        <SettingsValueRow label="Version" value={diagnostics.appVersion} />
      </SettingsSection>
    </ScrollView>
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

function DiagnosticIdentity({ label, task, value, pending, failed, retrying, onRetry }: {
  readonly label: string; readonly task: string; readonly value?: string;
  readonly pending: boolean; readonly failed: boolean; readonly retrying: boolean;
  readonly onRetry: () => Promise<void>;
}) {
  const { styles } = useSettingsListStyles();
  return <View>
    {pending ? <SettingsLoadingRow label={`Loading ${task}`} /> : <SettingsValueRow label={label} value={value || 'Unavailable'} />}
    {failed ? <>
      <Text accessibilityRole="alert" style={styles.errorMessage}>{value ? `Could not refresh ${task}. Previously loaded value is shown.` : `Could not load ${task}.`}</Text>
      <NativeCommandButton label={retrying ? `Retrying ${task}…` : `Retry ${task}`} disabled={retrying} onPress={() => void onRetry()} />
    </> : null}
  </View>;
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


function ownConfirmation(canPresent: () => boolean, run: (canPresent: () => boolean) => Promise<void>): () => Promise<void> {
  let accepted = false;
  return async () => {
    if (accepted || !canPresent()) return;
    accepted = true;
    await run(canPresent);
  };
}
