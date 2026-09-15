import { useTaskPresentation } from '../navigation/useTaskPresentation';
import { SettingsRefreshNotice } from './SettingsRefreshNotice';
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
  useProviderProfileModel
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
  const providers = useProviderProfileModel(query);
  if (providers.state.status !== 'ready') {
    return <ProviderStateView state={providers.state} onRetry={providers.retry} />;
  }

  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <SettingsRefreshNotice visible={providers.hasRefreshError} onRetry={providers.retry} />
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
  const capturePresentation = useTaskPresentation(manageCommand);

  async function create(key: string): Promise<void> {
    const canPresent = capturePresentation();
    if (!canPresent() || workingRef.current) return;
    const template = recommendedProviderProfiles.find((item) => item.key === key);
    if (!template) return;
    workingRef.current = true;
    setWorkingKey(key);
    try {
      const profile = await manageCommand.createRecommended(template);
      if (!canPresent()) return;
      feedback.showNotice({
        tone: 'success',
        title: 'Draft profile created',
        message: 'Add credentials, test it, then enable it for voice.'
      });
      onCreated(profile.id);
    } catch (error) {
      if (!canPresent()) return;
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
  const providers = useProviderProfileModel(query);
  const [operation, setOperation] = useState<'test' | 'lifecycle' | 'archive'>();
  const working = operation !== undefined;
  const workingRef = useRef(false);
  const capturePresentation = useTaskPresentation(manageCommand, `${providers.ownerKey}:${profileId}`);
  if (providers.state.status !== 'ready') {
    return <ProviderStateView state={providers.state} onRetry={providers.retry} />;
  }
  const profile = providers.state.viewModel.profiles.find((item) => item.id === profileId);
  if (!profile) {
    return (
      <ProviderStateView
        state={{ status: 'error', message: 'This provider profile is no longer available.' }}
        onRetry={providers.retry}
      />
    );
  }
  const profileDisplayName = profile.displayName;

  async function act(kind: 'test' | 'lifecycle' | 'archive', action: () => Promise<unknown>, title: string): Promise<void> {
    const canPresent = capturePresentation();
    if (!canPresent() || workingRef.current) return;
    workingRef.current = true;
    setOperation(kind);
    try {
      await action();
      if (!canPresent()) return;
      feedback.showNotice({
        tone: 'success',
        title,
        message: `${profileDisplayName} was updated.`
      });
      void providers.load().catch(() => undefined);
    } catch (error) {
      if (!canPresent()) return;
      feedback.showNotice({
        tone: 'error',
        title: 'Profile action failed',
        message: readableError(error)
      });
    } finally {
      workingRef.current = false;
      setOperation(undefined);
    }
  }

  const lifecycleAction = profile.lifecycleState === 'enabled' ? 'disable' : 'enable';
  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <SettingsRefreshNotice visible={providers.hasRefreshError} onRetry={providers.retry} />
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
        {profile.credentialPurpose ? <><SettingsActionRow accessibilityLabel={`Replace credential for ${profile.displayName}`} label="Replace Credential" disabled={working} onPress={() => { if (!workingRef.current) onEditCredential(); }} /><SettingsSeparator /></> : null}
        {profile.capability === 'language_inference' ? <><SettingsActionRow accessibilityLabel={`Edit prompt guidance for ${profile.displayName}`} label="Prompt Guidance" disabled={working} onPress={() => { if (!workingRef.current) onEditPrompt(); }} /><SettingsSeparator /></> : null}
        <SettingsActionRow accessibilityLabel={`Test connection for ${profile.displayName}`} disabled={working} label={operation === 'test' ? 'Testing…' : 'Test Connection'} onPress={() => void act('test', () => testCommand.execute(profile.id), 'Connection tested')} />
        {profile.lifecycleState !== 'archived' ? <><SettingsSeparator /><SettingsActionRow accessibilityLabel={`${lifecycleAction} ${profile.displayName}`} disabled={working} label={operation === 'lifecycle' ? 'Updating…' : lifecycleAction === 'enable' ? 'Enable Profile' : 'Disable Profile'} onPress={() => void act('lifecycle', () => manageCommand.changeLifecycle(profile.id, lifecycleAction), lifecycleAction === 'enable' ? 'Profile enabled' : 'Profile disabled')} /></> : null}
      </SettingsSection>
      {profile.lifecycleState !== 'archived' ? (
        <SettingsSection footer="Archived profiles remain in history but can’t be selected for voice.">
          <SettingsActionRow accessibilityLabel={`Archive ${profile.displayName}`} destructive disabled={working} label={operation === 'archive' ? 'Archiving…' : 'Archive Profile'} onPress={() => {
            const canPresent = capturePresentation();
            if (!canPresent() || workingRef.current) return;
            confirmArchive(profile, async () => {
              if (canPresent()) await act('archive', () => manageCommand.changeLifecycle(profile.id, 'archive'), 'Profile archived');
            });
          }} />
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
