import { t } from '../../presentation/localization';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileServerQuery } from '../serverState/useMobileServerQuery';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';
import { isAccessFailure } from '../serverState/isAccessFailure';
import { SettingsRefreshNotice } from './SettingsRefreshNotice';
import { useRef, useState } from 'react';
import { Alert, ScrollView, Text, View } from 'react-native';
import type { SettingsQuery, SettingsViewModel } from '../../application/settings/SettingsQuery';
import { useAppFeedback } from '../feedback/AppFeedback';
import {
  SettingsSection,
  SettingsActionRow,
  SettingsAppearanceRow,
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
        title: t('mobile.SettingsDetailScreens.couldNotSignOut'),
        message: readableError(error)
      });
      workingRef.current = false;
      setWorking(false);
    }
  }

  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <SettingsRefreshNotice visible={principal.isError} message={principal.data ? undefined : t('mobile.SettingsDetailScreens.couldNotLoadAccountDetailsYouCanRetryOr')} onRetry={async () => { await principal.refetch(); }} />
          <SettingsSection footer="Signing out keeps this server on your device so you can sign in again quickly.">
            <SettingsValueRow label={t('mobile.SettingsDetailScreens.signedInAs')} value={principalLabel} />
          </SettingsSection>
          <SettingsSection>
            <SettingsActionRow
              disabled={working}
              label={working ? t('mobile.SettingsDetailScreens.signingOut') : t('mobile.SettingsDetailScreens.signOut')}
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
      <SettingsAppearanceRow />
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
        title: t('mobile.SettingsDetailScreens.couldNotChangeServer'),
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
        title={t('mobile.SettingsDetailScreens.currentServer')}
      >
        <SettingsValueRow label={t('mobile.SettingsDetailScreens.server')} value={serverHostname(diagnostics.apiBaseUrl)} />
        <SettingsSeparator />
        <SettingsValueRow label={t('mobile.SettingsDetailScreens.address')} value={diagnostics.apiBaseUrl} />
      </SettingsSection>
      <SettingsSection footer="Changing servers signs you out and forgets this server and household selection on this device. It does not delete data from the server.">
        <SettingsActionRow
          disabled={working}
          label={working ? t('mobile.SettingsDetailScreens.changingServer') : t('mobile.SettingsDetailScreens.changeServer')}
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
      <SettingsSection title={t('mobile.SettingsDetailScreens.connection')}>
        <SettingsValueRow label={t('mobile.SettingsDetailScreens.aPIURL')} value={diagnostics.apiBaseUrl} />
        <SettingsSeparator />
        <SettingsValueRow label={t('mobile.SettingsDetailScreens.authentication')} value={authenticationLabel(diagnostics.authenticationMode)} />
      </SettingsSection>
      <SettingsSection title={t('mobile.SettingsDetailScreens.identity')}>
        <DiagnosticIdentity label={t('mobile.SettingsDetailScreens.principalID')} task="account identity"
          value={isAccessFailure(principal.error) ? undefined : principal.data?.id}
          pending={principal.isPending} failed={principal.isError} retrying={principal.isFetching}
          onRetry={async () => { await principal.refetch({ cancelRefetch: false }); }} />
        <SettingsSeparator />
        <DiagnosticIdentity label={t('mobile.SettingsDetailScreens.tenantID')} task="household identity" value={scope.data?.tenant.id}
          pending={scope.isPending} failed={scope.isError} retrying={scope.isFetching}
          onRetry={async () => { await scope.refetch({ cancelRefetch: false }); }} />
      </SettingsSection>
      <SettingsSection title={t('mobile.SettingsDetailScreens.application')}>
        <SettingsValueRow label={t('mobile.SettingsDetailScreens.version')} value={diagnostics.appVersion} />
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
        <Text accessibilityRole="header" style={styles.detailTitle}>{t('mobile.SettingsDetailScreens.stuffStash')}</Text>
        <Text style={styles.detailSubtitle}>{t('mobile.SettingsDetailScreens.aCalmFlexibleHomeInventoryForKnowingWhatYou')}</Text>
      </View>
      <SettingsSection>
        <SettingsValueRow label={t('mobile.SettingsDetailScreens.version')} value={diagnostics.appVersion} />
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
    {pending ? <SettingsLoadingRow label={t('mobile.SettingsDetailScreens.loading', { task: String(task) })} /> : <SettingsValueRow label={label} value={value || 'Unavailable'} />}
    {failed ? <>
      <Text accessibilityRole="alert" style={styles.errorMessage}>{value ? t('mobile.SettingsDetailScreens.couldNotRefreshPreviouslyLoadedValueIsShown', { task: String(task) }) : t('mobile.SettingsDetailScreens.couldNotLoad', { task: String(task) })}</Text>
      <NativeCommandButton label={retrying ? t('mobile.SettingsDetailScreens.retrying', { task: String(task) }) : t('mobile.SettingsDetailScreens.retry', { task: String(task) })} disabled={retrying} onPress={() => void onRetry()} />
    </> : null}
  </View>;
}

function confirmSignOut(label: string, onSignOut: () => Promise<void>): void {
  Alert.alert(t('mobile.SettingsDetailScreens.signOut2'), t('mobile.SettingsDetailScreens.youLlNeedToSignInAgainAsThis', { label: String(label) }), [
    { text: t('mobile.SettingsDetailScreens.cancel'), style: 'cancel' },
    { text: t('mobile.SettingsDetailScreens.signOut'), onPress: () => void onSignOut() }
  ]);
}

function confirmChangeServer(serverUrl: string, onChangeServer: () => Promise<void>): void {
  Alert.alert(
    t('mobile.SettingsDetailScreens.changeStuffStashServer'),
    t('mobile.SettingsDetailScreens.youLlBeSignedOutOfAndThisDevice', { value: String(serverHostname(serverUrl)) }),
    [
      { text: t('mobile.SettingsDetailScreens.cancel'), style: 'cancel' },
      { text: t('mobile.SettingsDetailScreens.changeServer'), onPress: () => void onChangeServer() }
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
