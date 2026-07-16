import { useRef, useState } from 'react';
import { Alert, ScrollView, Text, View } from 'react-native';
import type { ManageProviderProfileCommand } from '../../application/providerProfiles/ManageProviderProfileCommand';
import type { ProviderProfileSummary } from '../../application/providerProfiles/ProviderProfileRepository';
import type { ProviderProfileSettingsQuery } from '../../application/providerProfiles/ProviderProfileSettingsQuery';
import { recommendedProviderProfiles } from '../../application/providerProfiles/RecommendedProviderProfiles';
import type { TestProviderProfileCommand } from '../../application/providerProfiles/TestProviderProfileCommand';
import { useAppFeedback } from '../feedback/AppFeedback';
import { spacing } from '../theme/tokens';
import {
  SettingsActionRow,
  SettingsNavigationRow,
  SettingsSection,
  SettingsSeparator,
  SettingsValueRow,
  useSettingsListStyles
} from './SettingsList';
import {
  formatProviderProfileCredentialStatusLabel,
  formatProviderProfileLifecycleLabel,
  formatProviderProfileTestStatusLabel
} from './ProviderProfilesVoiceSetupPresentation';
import {
  ProviderStateView,
  readableError,
  useProviderSettings
} from './ProviderSettingsSupport';
import { stagePresentation } from './VoiceStagePresentation';

export function ProviderProfileListScreen({
  onAdd,
  onOpenProfile,
  query
}: {
  readonly onAdd: () => void;
  readonly onOpenProfile: (profileId: string) => void;
  readonly query: ProviderProfileSettingsQuery;
}) {
  const { styles } = useSettingsListStyles();
  const providers = useProviderSettings(query);
  if (providers.state.status !== 'ready') {
    return <ProviderStateView state={providers.state} onRetry={providers.load} />;
  }

  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <SettingsSection footer="Profiles are shared tenant-wide and supply one stage of the voice pipeline.">
        <SettingsActionRow accessibilityLabel="Add a provider profile" label="Add Profile" onPress={onAdd} />
      </SettingsSection>
      <SettingsSection title="Profiles">
        {providers.state.viewModel.profiles.length === 0 ? (
          <SettingsValueRow label="No profiles" value="Add one to begin" />
        ) : providers.state.viewModel.profiles.map((profile, index) => (
          <View key={profile.id}>
            {index > 0 ? <SettingsSeparator /> : null}
            <SettingsNavigationRow
              accessibilityLabel={`Open provider profile ${profile.displayName}. ${formatProviderProfileLifecycleLabel(profile.lifecycleState)}`}
              context={`${stagePresentation(profile.capability).title} · ${profile.providerKind}`}
              label={profile.displayName}
              onPress={() => onOpenProfile(profile.id)}
              value={formatProviderProfileLifecycleLabel(profile.lifecycleState)}
            />
          </View>
        ))}
      </SettingsSection>
    </ScrollView>
  );
}

export function AddProviderProfileScreen({
  manageCommand,
  onCreated
}: {
  readonly manageCommand: ManageProviderProfileCommand;
  readonly onCreated: (profileId: string) => void;
}) {
  const { styles } = useSettingsListStyles();
  const feedback = useAppFeedback();
  const [workingKey, setWorkingKey] = useState<string>();
  const workingRef = useRef(false);

  async function create(key: string): Promise<void> {
    if (workingRef.current) return;
    const template = recommendedProviderProfiles.find((item) => item.key === key);
    if (!template) return;
    workingRef.current = true;
    setWorkingKey(key);
    try {
      const profile = await manageCommand.createRecommended(template);
      feedback.showNotice({
        tone: 'success',
        title: 'Draft profile created',
        message: 'Add credentials, test it, then enable it for voice.'
      });
      onCreated(profile.id);
    } catch (error) {
      feedback.showNotice({
        tone: 'error',
        title: 'Could not create profile',
        message: readableError(error)
      });
    } finally {
      workingRef.current = false;
      setWorkingKey(undefined);
    }
  }

  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <View style={styles.detailHeader}>
        <Text accessibilityRole="header" style={styles.detailTitle}>Choose a Service</Text>
        <Text style={styles.detailSubtitle}>Stuff Stash creates a disabled draft first. You’ll add credentials, test the connection, and enable it next.</Text>
      </View>
      <SettingsSection title="Recommended">
        {recommendedProviderProfiles.map((template, index) => (
          <View key={template.key}>
            {index > 0 ? <SettingsSeparator /> : null}
            <SettingsActionRow
              accessibilityLabel={`Create draft ${template.title} provider profile`}
              disabled={workingKey !== undefined}
              label={workingKey === template.key ? `Creating ${template.title}…` : template.title}
              onPress={() => void create(template.key)}
            />
            <Text style={[styles.secondaryText, {
              paddingBottom: spacing.sm,
              paddingHorizontal: spacing.md
            }]}>{template.description}</Text>
          </View>
        ))}
      </SettingsSection>
    </ScrollView>
  );
}

export function ProviderProfileDetailScreen({
  manageCommand,
  onEditCredential,
  onEditPrompt,
  profileId,
  query,
  testCommand
}: {
  readonly manageCommand: ManageProviderProfileCommand;
  readonly onEditCredential: () => void;
  readonly onEditPrompt: () => void;
  readonly profileId: string;
  readonly query: ProviderProfileSettingsQuery;
  readonly testCommand: TestProviderProfileCommand;
}) {
  const { styles } = useSettingsListStyles();
  const feedback = useAppFeedback();
  const providers = useProviderSettings(query);
  const [working, setWorking] = useState(false);
  const workingRef = useRef(false);
  if (providers.state.status !== 'ready') {
    return <ProviderStateView state={providers.state} onRetry={providers.load} />;
  }
  const profile = providers.state.viewModel.profiles.find((item) => item.id === profileId);
  if (!profile) {
    return (
      <ProviderStateView
        state={{ status: 'error', message: 'This provider profile is no longer available.' }}
        onRetry={providers.load}
      />
    );
  }
  const profileDisplayName = profile.displayName;

  async function act(action: () => Promise<unknown>, title: string): Promise<void> {
    if (workingRef.current) return;
    workingRef.current = true;
    setWorking(true);
    try {
      await action();
      feedback.showNotice({
        tone: 'success',
        title,
        message: `${profileDisplayName} was updated.`
      });
      await providers.load();
    } catch (error) {
      feedback.showNotice({
        tone: 'error',
        title: 'Profile action failed',
        message: readableError(error)
      });
    } finally {
      workingRef.current = false;
      setWorking(false);
    }
  }

  const lifecycleAction = profile.lifecycleState === 'enabled' ? 'disable' : 'enable';
  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <View style={styles.detailHeader}>
        <Text accessibilityRole="header" style={styles.detailTitle}>{profile.displayName}</Text>
        <Text style={styles.detailSubtitle}>{stagePresentation(profile.capability).title} · {profile.providerKind}</Text>
      </View>
      <SettingsSection title="Configuration">
        <SettingsValueRow label="Model" value={profile.modelName || 'Default'} />
        <SettingsSeparator />
        <SettingsValueRow label="Status" value={formatProviderProfileLifecycleLabel(profile.lifecycleState)} />
        <SettingsSeparator />
        <SettingsValueRow label="Credential" value={formatProviderProfileCredentialStatusLabel(profile.credentialStatus)} />
        <SettingsSeparator />
        <SettingsValueRow label="Last tested" value={formatProviderProfileTestStatusLabel(profile.lastTestedAt)} />
      </SettingsSection>
      <SettingsSection title="Actions">
        {profile.credentialPurpose ? <><SettingsActionRow accessibilityLabel={`Replace credential for ${profile.displayName}`} label="Replace Credential" onPress={onEditCredential} /><SettingsSeparator /></> : null}
        {profile.capability === 'language_inference' ? <><SettingsActionRow accessibilityLabel={`Edit prompt guidance for ${profile.displayName}`} label="Prompt Guidance" onPress={onEditPrompt} /><SettingsSeparator /></> : null}
        <SettingsActionRow accessibilityLabel={`Test connection for ${profile.displayName}`} disabled={working} label={working ? 'Testing…' : 'Test Connection'} onPress={() => void act(() => testCommand.execute(profile.id), 'Connection tested')} />
        {profile.lifecycleState !== 'archived' ? <><SettingsSeparator /><SettingsActionRow accessibilityLabel={`${lifecycleAction} ${profile.displayName}`} disabled={working} label={working ? 'Updating…' : lifecycleAction === 'enable' ? 'Enable Profile' : 'Disable Profile'} onPress={() => void act(() => manageCommand.changeLifecycle(profile.id, lifecycleAction), lifecycleAction === 'enable' ? 'Profile enabled' : 'Profile disabled')} /></> : null}
      </SettingsSection>
      {profile.lifecycleState !== 'archived' ? (
        <SettingsSection footer="Archived profiles remain in history but can’t be selected for voice.">
          <SettingsActionRow accessibilityLabel={`Archive ${profile.displayName}`} destructive disabled={working} label={working ? 'Archiving…' : 'Archive Profile'} onPress={() => confirmArchive(profile, () => act(() => manageCommand.changeLifecycle(profile.id, 'archive'), 'Profile archived'))} />
        </SettingsSection>
      ) : null}
    </ScrollView>
  );
}

function confirmArchive(
  profile: ProviderProfileSummary,
  archive: () => Promise<void>
): void {
  Alert.alert(
    'Archive provider profile?',
    `${profile.displayName} will no longer be available for voice.`,
    [
      { text: 'Cancel', style: 'cancel' },
      { text: 'Archive', style: 'destructive', onPress: () => void archive() }
    ]
  );
}
