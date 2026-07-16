import { useMemo, useRef, useState } from 'react';
import { Pressable, ScrollView, StyleSheet, Text, TextInput, View } from 'react-native';
import type { ManageProviderProfileCommand } from '../../application/providerProfiles/ManageProviderProfileCommand';
import type {
  ProviderCredentialPurpose,
  ProviderProfileSummary
} from '../../application/providerProfiles/ProviderProfileRepository';
import type { ProviderProfileSettingsQuery } from '../../application/providerProfiles/ProviderProfileSettingsQuery';
import { useAppFeedback } from '../feedback/AppFeedback';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useSettingsListStyles } from './SettingsList';
import { ProviderStateView, readableError, useProviderSettings } from './ProviderSettingsSupport';

type ProviderEditorProps = {
  readonly manageCommand: ManageProviderProfileCommand;
  readonly onCancel: () => void;
  readonly onSaved: () => void;
  readonly profileId: string;
  readonly query: ProviderProfileSettingsQuery;
};

export function ProviderCredentialScreen({
  manageCommand,
  onCancel,
  onSaved,
  profileId,
  query
}: ProviderEditorProps) {
  const providers = useProviderSettings(query);
  if (providers.state.status !== 'ready') {
    return <ProviderStateView state={providers.state} onRetry={providers.load} />;
  }
  const profile = providers.state.viewModel.profiles.find((item) => item.id === profileId);
  if (!profile?.credentialPurpose) {
    return <ProviderStateView state={{ status: 'error', message: 'This profile does not support mobile credential editing.' }} onRetry={providers.load} />;
  }
  return <CredentialForm manageCommand={manageCommand} onCancel={onCancel} onSaved={onSaved} profile={{ ...profile, credentialPurpose: profile.credentialPurpose }} />;
}

export function ProviderPromptScreen({
  manageCommand,
  onCancel,
  onSaved,
  profileId,
  query
}: ProviderEditorProps) {
  const providers = useProviderSettings(query);
  if (providers.state.status !== 'ready') {
    return <ProviderStateView state={providers.state} onRetry={providers.load} />;
  }
  const profile = providers.state.viewModel.profiles.find((item) => item.id === profileId);
  if (!profile) {
    return <ProviderStateView state={{ status: 'error', message: 'This provider profile is no longer available.' }} onRetry={providers.load} />;
  }
  return <PromptForm manageCommand={manageCommand} onCancel={onCancel} onSaved={onSaved} profile={profile} />;
}

function CredentialForm({
  manageCommand,
  onCancel,
  onSaved,
  profile
}: {
  readonly manageCommand: ManageProviderProfileCommand;
  readonly onCancel: () => void;
  readonly onSaved: () => void;
  readonly profile: ProviderProfileSummary & { readonly credentialPurpose: ProviderCredentialPurpose };
}) {
  const { palette, styles } = useSettingsListStyles();
  const local = useMemo(() => editorStyles(palette), [palette]);
  const feedback = useAppFeedback();
  const [value, setValue] = useState('');
  const [saving, setSaving] = useState(false);
  const savingRef = useRef(false);

  async function save(): Promise<void> {
    if (savingRef.current) return;
    savingRef.current = true;
    setSaving(true);
    try {
      await manageCommand.replaceCredential({
        providerProfileId: profile.id,
        purpose: profile.credentialPurpose,
        credential: value
      });
      setValue('');
      feedback.showNotice({ tone: 'success', title: 'Credential saved', message: `${profile.displayName} is ready to test.` });
      onSaved();
    } catch (error) {
      feedback.showNotice({ tone: 'error', title: 'Credential not saved', message: readableError(error) });
    } finally {
      savingRef.current = false;
      setSaving(false);
    }
  }

  return (
    <ScrollView contentContainerStyle={[styles.content, local.form]} keyboardShouldPersistTaps="handled" style={styles.shell}>
      <Text accessibilityRole="header" style={styles.detailTitle}>Replace Credential</Text>
      <Text style={styles.detailSubtitle}>{profile.displayName}. Secrets are sent directly to your Stuff Stash server and aren’t stored on this device.</Text>
      {profile.credentialPurpose === 'server_adc' ? (
        <Text style={local.explanation}>Use the Application Default Credentials configured by the server operator. No provider secret is entered here.</Text>
      ) : (
        <View>
          <Text style={local.label}>{credentialLabel(profile.credentialPurpose)}</Text>
          <TextInput
            accessibilityLabel={credentialLabel(profile.credentialPurpose)}
            autoCapitalize="none"
            autoCorrect={false}
            onChangeText={setValue}
            secureTextEntry
            style={local.input}
            value={value}
          />
        </View>
      )}
      <Text style={styles.secondaryText}>Leaving this screen keeps a newly created profile as a disabled draft.</Text>
      <View style={local.actions}>
        <Pressable accessibilityRole="button" disabled={saving} onPress={onCancel} style={local.secondaryButton}><Text style={styles.actionText}>Cancel</Text></Pressable>
        <Pressable accessibilityRole="button" accessibilityState={{ busy: saving, disabled: saving }} disabled={saving} onPress={() => void save()} style={local.primaryButton}><Text style={local.primaryText}>{saving ? 'Saving…' : 'Save Credential'}</Text></Pressable>
      </View>
    </ScrollView>
  );
}

function PromptForm({
  manageCommand,
  onCancel,
  onSaved,
  profile
}: {
  readonly manageCommand: ManageProviderProfileCommand;
  readonly onCancel: () => void;
  readonly onSaved: () => void;
  readonly profile: ProviderProfileSummary;
}) {
  const { palette, styles } = useSettingsListStyles();
  const local = useMemo(() => editorStyles(palette), [palette]);
  const feedback = useAppFeedback();
  const [value, setValue] = useState('');
  const [saving, setSaving] = useState(false);
  const savingRef = useRef(false);

  async function save(): Promise<void> {
    if (savingRef.current) return;
    savingRef.current = true;
    setSaving(true);
    try {
      await manageCommand.replacePromptTemplate({ providerProfileId: profile.id, promptTemplate: value });
      feedback.showNotice({ tone: 'success', title: 'Prompt guidance saved', message: `${profile.displayName} was updated.` });
      onSaved();
    } catch (error) {
      feedback.showNotice({ tone: 'error', title: 'Prompt not saved', message: readableError(error) });
    } finally {
      savingRef.current = false;
      setSaving(false);
    }
  }

  return (
    <ScrollView contentContainerStyle={[styles.content, local.form]} keyboardShouldPersistTaps="handled" style={styles.shell}>
      <Text accessibilityRole="header" style={styles.detailTitle}>Prompt Guidance</Text>
      <Text style={styles.detailSubtitle}>Optional tenant guidance for {profile.displayName}. Existing hidden prompt text is never returned to the phone.</Text>
      <Text style={local.label}>New prompt guidance</Text>
      <TextInput accessibilityLabel="New prompt guidance" multiline onChangeText={setValue} style={[local.input, local.multiline]} value={value} />
      <View style={local.actions}>
        <Pressable accessibilityRole="button" disabled={saving} onPress={onCancel} style={local.secondaryButton}><Text style={styles.actionText}>Cancel</Text></Pressable>
        <Pressable accessibilityRole="button" accessibilityState={{ busy: saving, disabled: saving }} disabled={saving} onPress={() => void save()} style={local.primaryButton}><Text style={local.primaryText}>{saving ? 'Saving…' : 'Save Guidance'}</Text></Pressable>
      </View>
    </ScrollView>
  );
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
    actions: { gap: spacing.sm },
    secondaryButton: { alignItems: 'center', backgroundColor: colors.surface, borderRadius: radius.md, justifyContent: 'center', minHeight: 48, paddingHorizontal: spacing.md },
    primaryButton: { alignItems: 'center', backgroundColor: colors.action, borderRadius: radius.md, justifyContent: 'center', minHeight: 48, paddingHorizontal: spacing.md },
    primaryText: { color: colors.onAction, fontSize: 17, fontWeight: '600' }
  });
}
