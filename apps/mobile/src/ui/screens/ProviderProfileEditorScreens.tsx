import { Stack } from 'expo-router';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';
import { useProviderEditorExit } from './useProviderEditorExit';
import { useTaskPresentation } from '../navigation/useTaskPresentation';
import { SettingsRefreshNotice } from './SettingsRefreshNotice';
import { useEffect, useMemo, useRef, useState } from 'react';
import { AccessibilityInfo, Platform, ScrollView, StyleSheet, Text, View } from 'react-native';
import type { ManageProviderProfileCommand } from '../../application/providerProfiles/ManageProviderProfileCommand';
import type {
  ProviderCredentialPurpose,
  ProviderProfileSummary
} from '../../application/providerProfiles/ProviderProfileRepository';
import type { ProviderProfileSettingsQuery } from '../../application/providerProfiles/ProviderProfileSettingsQuery';
import { useAppFeedback } from '../feedback/AppFeedback';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useSettingsListStyles } from './SettingsList';
import { ProviderStateView, readableError, useProviderProfileModel } from './ProviderSettingsSupport';
import { AppTextInput, appKeyboardDismissMode } from '../components/AppTextInput';

type ProviderEditorProps = {
  readonly manageCommand: ManageProviderProfileCommand;
  readonly onSaved: () => void;
  readonly profileId: string;
  readonly query: ProviderProfileSettingsQuery;
};

export function ProviderCredentialScreen({
  manageCommand,
  onSaved,
  profileId,
  query
}: ProviderEditorProps) {
  const providers = useProviderProfileModel(query);
  if (providers.state.status !== 'ready') {
    return <ProviderStateView state={providers.state} onRetry={providers.retry} />;
  }
  const profile = providers.state.viewModel.profiles.find((item) => item.id === profileId);
  if (!profile?.credentialPurpose) {
    return <ProviderStateView state={{ status: 'error', message: 'This profile does not support mobile credential editing.' }} onRetry={providers.retry} />;
  }
  return <><SettingsRefreshNotice visible={providers.hasRefreshError} onRetry={providers.retry} /><CredentialForm key={`${providers.ownerKey}:${profile.id}`} manageCommand={manageCommand} onSaved={onSaved} profile={{ ...profile, credentialPurpose: profile.credentialPurpose }} /></>;
}

export function ProviderPromptScreen({
  manageCommand,
  onSaved,
  profileId,
  query
}: ProviderEditorProps) {
  const providers = useProviderProfileModel(query);
  if (providers.state.status !== 'ready') {
    return <ProviderStateView state={providers.state} onRetry={providers.retry} />;
  }
  const profile = providers.state.viewModel.profiles.find((item) => item.id === profileId);
  if (!profile) {
    return <ProviderStateView state={{ status: 'error', message: 'This provider profile is no longer available.' }} onRetry={providers.retry} />;
  }
  return <><SettingsRefreshNotice visible={providers.hasRefreshError} onRetry={providers.retry} /><PromptForm key={`${providers.ownerKey}:${profile.id}`} manageCommand={manageCommand} onSaved={onSaved} profile={profile} /></>;
}

function CredentialForm({
  manageCommand,
  onSaved,
  profile
}: {
  readonly manageCommand: ManageProviderProfileCommand;
  readonly onSaved: () => void;
  readonly profile: ProviderProfileSummary & { readonly credentialPurpose: ProviderCredentialPurpose };
}) {
  const { palette, styles } = useSettingsListStyles();
  const local = useMemo(() => editorStyles(palette), [palette]);
  const feedback = useAppFeedback();
  const [value, setValue] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string>();
  const savingRef = useRef(false);
  const capturePresentation = useTaskPresentation(manageCommand, profile.id);
  const authorizeExit = useProviderEditorExit({ dirty: value.length > 0, isSaving: () => savingRef.current, capturePresentation });

  const valid = profile.credentialPurpose === 'server_adc' || Boolean(value.trim());
  async function save(): Promise<void> {
    const canPresent = capturePresentation();
    if (!canPresent() || savingRef.current || !valid) return;
    savingRef.current = true;
    setSaving(true);
    setError(undefined);
    try {
      await manageCommand.replaceCredential({
        providerProfileId: profile.id,
        purpose: profile.credentialPurpose,
        credential: value
      });
      // This keyed form owns the submitted secret even after navigation blur.
      setValue('');
      if (!canPresent()) return;
      authorizeExit(canPresent, () => {
        feedback.showNotice({ tone: 'success', title: 'Credential saved', message: `${profile.displayName} is ready to test.` });
        onSaved();
      });
    } catch (error) {
      if (!canPresent()) return;
      setError(readableError(error));
    } finally {
      savingRef.current = false;
      setSaving(false);
    }
  }

  const saveOptions = useNativeHeaderActionOptions([{ kind: 'save', label: 'Save Credential', disabled: saving || !valid, onPress: () => void save() }]);
  const headerOptions = useMemo(() => ({ ...saveOptions, gestureEnabled: !saving, headerBackVisible: !saving }), [saveOptions, saving]);
  return (<>
    <Stack.Screen options={headerOptions} />
    <ScrollView contentContainerStyle={[styles.content, local.form]} keyboardDismissMode={appKeyboardDismissMode()} keyboardShouldPersistTaps="handled" style={styles.shell}>
      <Text style={styles.detailSubtitle}>{profile.displayName}. Secrets are sent directly to your Stuff Stash server and aren’t stored on this device.</Text>
      {profile.credentialPurpose === 'server_adc' ? (
        <Text style={local.explanation}>Use the Application Default Credentials configured by the server operator. No provider secret is entered here.</Text>
      ) : (
        <View>
          <Text style={local.label}>{credentialLabel(profile.credentialPurpose)}</Text>
          <AppTextInput
            accessibilityLabel={credentialLabel(profile.credentialPurpose)}
            autoCapitalize="none"
            autoCorrect={false}
            editable={!saving} onChangeText={next => { if (!savingRef.current) { setValue(next); setError(undefined); } }}
            secureTextEntry
            style={local.input}
            value={value}
          />
        </View>
      )}
      {error ? <ProviderEditorError message={error} /> : null}
      {saving ? <Text accessibilityLiveRegion="polite" style={styles.secondaryText}>Saving…</Text> : null}
    </ScrollView>
  </>);
}

function PromptForm({
  manageCommand,
  onSaved,
  profile
}: {
  readonly manageCommand: ManageProviderProfileCommand;
  readonly onSaved: () => void;
  readonly profile: ProviderProfileSummary;
}) {
  const { palette, styles } = useSettingsListStyles();
  const local = useMemo(() => editorStyles(palette), [palette]);
  const feedback = useAppFeedback();
  const [value, setValue] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string>();
  const savingRef = useRef(false);
  const capturePresentation = useTaskPresentation(manageCommand, profile.id);
  const authorizeExit = useProviderEditorExit({ dirty: value.length > 0, isSaving: () => savingRef.current, capturePresentation });

  const valid = Boolean(value.trim());
  async function save(): Promise<void> {
    const canPresent = capturePresentation();
    if (!canPresent() || savingRef.current || !valid) return;
    savingRef.current = true;
    setSaving(true);
    setError(undefined);
    try {
      await manageCommand.replacePromptTemplate({ providerProfileId: profile.id, promptTemplate: value });
      setValue('');
      if (!canPresent()) return;
      authorizeExit(canPresent, () => {
        feedback.showNotice({ tone: 'success', title: 'Prompt guidance saved', message: `${profile.displayName} was updated.` });
        onSaved();
      });
    } catch (error) {
      if (!canPresent()) return;
      setError(readableError(error));
    } finally {
      savingRef.current = false;
      setSaving(false);
    }
  }

  const saveOptions = useNativeHeaderActionOptions([{ kind: 'save', label: 'Save Guidance', disabled: saving || !valid, onPress: () => void save() }]);
  const headerOptions = useMemo(() => ({ ...saveOptions, gestureEnabled: !saving, headerBackVisible: !saving }), [saveOptions, saving]);
  return (<>
    <Stack.Screen options={headerOptions} />
    <ScrollView contentContainerStyle={[styles.content, local.form]} keyboardDismissMode={appKeyboardDismissMode()} keyboardShouldPersistTaps="handled" style={styles.shell}>
      <Text style={styles.detailSubtitle}>Optional tenant guidance for {profile.displayName}. Existing hidden prompt text is never returned to the phone.</Text>
      <Text style={local.label}>New prompt guidance</Text>
      <AppTextInput accessibilityLabel="New prompt guidance" multiline editable={!saving} onChangeText={next => { if (!savingRef.current) { setValue(next); setError(undefined); } }} style={[local.input, local.multiline]} value={value} />
      {error ? <ProviderEditorError message={error} /> : null}
      {saving ? <Text accessibilityLiveRegion="polite" style={styles.secondaryText}>Saving…</Text> : null}
    </ScrollView>
  </>);
}

function ProviderEditorError({ message }: { readonly message: string }) {
  const { palette } = useSettingsListStyles();
  const styles = useMemo(() => editorStyles(palette), [palette]);
  useEffect(() => {
    if (Platform.OS === 'ios') AccessibilityInfo.announceForAccessibility(message);
  }, [message]);
  return <Text accessibilityLiveRegion="polite" style={styles.error}>{message}</Text>;
}

function credentialLabel(purpose: ProviderCredentialPurpose): string {
  return purpose === 'api_key'
    ? 'API key'
    : purpose === 'oauth_bearer'
      ? 'OAuth bearer token'
      : 'Server credentials';
}

function editorStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
    form: { gap: spacing.md, padding: spacing.lg },
    label: { color: colors.text, fontSize: 15, fontWeight: '600', marginBottom: spacing.xs },
    input: { backgroundColor: colors.surface, borderColor: colors.controlBorder, borderRadius: radius.md, borderWidth: 1, color: colors.text, fontSize: 17, minHeight: 48, paddingHorizontal: spacing.md, paddingVertical: spacing.sm },
    multiline: { minHeight: 160, textAlignVertical: 'top' },
    explanation: { backgroundColor: colors.surface, borderRadius: radius.md, color: colors.text, fontSize: 16, lineHeight: 23, padding: spacing.md },
    error: { color: colors.danger, fontSize: 15, lineHeight: 22 }
  });
}
