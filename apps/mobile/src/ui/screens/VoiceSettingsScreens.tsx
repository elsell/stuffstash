import { useRef, useState } from 'react';
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  Text,
  View
} from 'react-native';
import type { ManageProviderProfileCommand } from '../../application/providerProfiles/ManageProviderProfileCommand';
import type {
  ProviderProfileCapability,
  ProviderProfileSummary,
  VoiceProviderConfiguration,
  VoiceProviderSlot
} from '../../application/providerProfiles/ProviderProfileRepository';
import type { ProviderProfileSettingsQuery, ProviderProfileSettingsViewModel } from '../../application/providerProfiles/ProviderProfileSettingsQuery';
import type { TestProviderProfileCommand } from '../../application/providerProfiles/TestProviderProfileCommand';
import type { SettingsQuery } from '../../application/settings/SettingsQuery';
import { useAppFeedback } from '../feedback/AppFeedback';
import {
  SettingsActionRow,
  SettingsNavigationRow,
  SettingsSection,
  SettingsSeparator,
  SettingsValueRow,
  useSettingsListStyles
} from './SettingsList';
import {
  formatVoiceProviderReadinessLabel,
  formatVoiceProviderSelectionSourceLabel,
  voiceProviderSetupIssueLabels
} from './ProviderProfilesVoiceSetupPresentation';
import { useSettingsModel } from './SettingsScreenState';
import { ProviderStateView, readableError, useProviderSettings } from './ProviderSettingsSupport';
import { stagePresentation } from './VoiceStagePresentation';

export { ProviderCredentialScreen, ProviderPromptScreen } from './ProviderProfileEditorScreens';
export {
  AddProviderProfileScreen,
  ProviderProfileDetailScreen,
  ProviderProfileListScreen
} from './ProviderProfileScreens';

export function VoiceSetupScreen({
  onOpenCapability,
  onOpenProfiles,
  query,
  settingsQuery
}: {
  readonly onOpenCapability: (capability: ProviderProfileCapability) => void;
  readonly onOpenProfiles: () => void;
  readonly query: ProviderProfileSettingsQuery;
  readonly settingsQuery: SettingsQuery;
}) {
  const { styles } = useSettingsListStyles();
  const providers = useProviderSettings(query);
  const settings = useSettingsModel(settingsQuery);
  if (providers.state.status !== 'ready') return <ProviderStateView state={providers.state} onRetry={providers.load} />;
  if (settings.state.status !== 'ready') return <SettingsStateBridge state={settings.state} onRetry={settings.load} />;
  const { configuration } = providers.state.viewModel;
  const tenant = settings.state.settings.selectedTenant;
  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <View style={styles.detailHeader}>
        <Text accessibilityRole="header" style={styles.detailTitle}>
          {configuration.readiness === 'ready' ? 'Voice is ready' : 'Voice needs attention'}
        </Text>
        <Text style={styles.detailSubtitle}>Applies to everyone in {tenant.name}. Configure how Stuff Stash listens, understands requests, and speaks.</Text>
      </View>
      <SettingsSection title="Voice Pipeline">
        {configuration.slots.map((slot, index) => {
          const presentation = stagePresentation(slot.capability);
          return (
            <View key={slot.capability}>
              {index > 0 ? <SettingsSeparator /> : null}
              <SettingsNavigationRow
                accessibilityLabel={`Open ${presentation.title} voice stage for ${tenant.name}. ${formatVoiceProviderReadinessLabel(slot.readiness)}`}
                context={`${presentation.description} · ${selectedProfileLabel(slot)}`}
                label={presentation.title}
                onPress={() => onOpenCapability(slot.capability)}
                value={formatVoiceProviderReadinessLabel(slot.readiness)}
              />
            </View>
          );
        })}
      </SettingsSection>
      <SettingsSection footer="Provider profiles are advanced tenant-wide service configurations." title="Advanced">
        <SettingsNavigationRow
          accessibilityLabel={`Open advanced provider profiles for ${tenant.name}`}
          label="Provider Profiles"
          onPress={onOpenProfiles}
          value={`${providers.state.viewModel.profiles.length}`}
        />
      </SettingsSection>
    </ScrollView>
  );
}

export function VoiceCapabilityScreen({
  capability,
  manageCommand,
  onAddProfile,
  onEditCredential,
  onEditProfile,
  query,
  testCommand
}: {
  readonly capability: ProviderProfileCapability;
  readonly manageCommand: ManageProviderProfileCommand;
  readonly onAddProfile: () => void;
  readonly onEditCredential: (profileId: string) => void;
  readonly onEditProfile: (profileId: string) => void;
  readonly query: ProviderProfileSettingsQuery;
  readonly testCommand: TestProviderProfileCommand;
}) {
  const { styles } = useSettingsListStyles();
  const feedback = useAppFeedback();
  const providers = useProviderSettings(query);
  const [working, setWorking] = useState(false);
  const workingRef = useRef(false);
  if (providers.state.status !== 'ready') return <ProviderStateView state={providers.state} onRetry={providers.load} />;
  const slot = providers.state.viewModel.configuration.slots.find((item) => item.capability === capability);
  if (!slot) return <ProviderStateView state={{ status: 'error', message: 'This voice stage is not available.' }} onRetry={providers.load} />;
  const stage = stagePresentation(capability);
  const selectedProfile = slot.selectedProfile;
  const recommendedAction = slot.recommendedAction;
  const alternatives = providers.state.viewModel.profiles.filter((profile) =>
    profile.capability === capability && profile.id !== slot.selectedProfileId && profile.lifecycleState !== 'archived');

  async function act(action: () => Promise<void>, success: string): Promise<void> {
    if (workingRef.current) return;
    workingRef.current = true;
    setWorking(true);
    try {
      await action();
      feedback.showNotice({ tone: 'success', title: success, message: `${stage.title} setup was updated.` });
      await providers.load();
    } catch (error) {
      feedback.showNotice({ tone: 'error', title: 'Could not update voice', message: readableError(error) });
    } finally {
      workingRef.current = false;
      setWorking(false);
    }
  }

  const issueLabels = voiceProviderSetupIssueLabels(slot.readiness, recommendedAction);

  function primaryAction(): {
    readonly accessibilityLabel: string;
    readonly label: string;
    readonly run: () => void;
  } | undefined {
    switch (recommendedAction) {
      case 'add_profile':
        return {
          accessibilityLabel: `Add provider profile for ${stage.title}`,
          label: 'Add Profile',
          run: onAddProfile
        };
      case 'replace_credential':
        return selectedProfile
          ? {
              accessibilityLabel: `Add credential for ${selectedProfile.displayName} in ${stage.title}`,
              label: 'Add Credential',
              run: () => onEditCredential(selectedProfile.id)
            }
          : undefined;
      case 'enable_profile':
        return selectedProfile
          ? {
              accessibilityLabel: `Enable ${selectedProfile.displayName} for ${stage.title}`,
              label: working ? 'Enabling…' : 'Enable Service',
              run: () => void act(() => manageCommand.changeLifecycle(selectedProfile.id, 'enable').then(() => undefined), 'Service enabled')
            }
          : undefined;
      case 'test_profile':
        return selectedProfile
          ? {
              accessibilityLabel: `Test ${selectedProfile.displayName} for ${stage.title}`,
              label: working ? 'Testing…' : 'Test Connection',
              run: () => void act(() => testCommand.execute(selectedProfile.id).then(() => undefined), 'Connection tested')
            }
          : undefined;
      default:
        return undefined;
    }
  }

  const directAction = primaryAction();

  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <View style={styles.detailHeader}>
        <Text accessibilityRole="header" style={styles.detailTitle}>{stage.title}</Text>
        <Text style={styles.detailSubtitle}>{stage.longDescription}</Text>
      </View>
      <SettingsSection title="Selected Service">
        {slot.selectedProfile ? (
          <SettingsNavigationRow
            accessibilityLabel={`Open provider profile ${slot.selectedProfile.displayName}`}
            context={`${slot.selectedProfile.providerKind} · ${slot.selectedProfile.modelName || 'Default model'}`}
            label={slot.selectedProfile.displayName}
            onPress={() => onEditProfile(slot.selectedProfile!.id)}
            value={formatVoiceProviderReadinessLabel(slot.readiness)}
          />
        ) : <SettingsValueRow label="Service" value="Not selected" />}
      </SettingsSection>
      <SettingsSection
        footer={issueLabels.length > 0 ? issueLabels.join(' ') : undefined}
        title="Setup Status"
      >
        <SettingsValueRow label="Selection" value={formatVoiceProviderSelectionSourceLabel(slot.selectionSource)} />
        {slot.duplicateProfiles.length > 0 ? <><SettingsSeparator /><SettingsValueRow label="Ready choices" value={`${slot.duplicateProfiles.length}`} /></> : null}
      </SettingsSection>
      {directAction ? (
        <SettingsSection>
          <SettingsActionRow
            accessibilityLabel={directAction.accessibilityLabel}
            disabled={working}
            label={directAction.label}
            onPress={directAction.run}
          />
        </SettingsSection>
      ) : null}
      {alternatives.length > 0 ? (
        <SettingsSection footer="Changing the selection affects voice for everyone in this tenant." title="Other Services">
          {alternatives.map((profile, index) => (
            <View key={profile.id}>
              {index > 0 ? <SettingsSeparator /> : null}
              <SettingsActionRow
                accessibilityLabel={`Use ${profile.displayName} for ${stage.title}`}
                disabled={working}
                label={`Use ${profile.displayName}`}
                onPress={() => void act(() => selectProfile(manageCommand, providers.state.status === 'ready' ? providers.state.viewModel.configuration : undefined, slot, profile), 'Voice service selected')}
              />
            </View>
          ))}
        </SettingsSection>
      ) : null}
    </ScrollView>
  );
}

function SettingsStateBridge({ state, onRetry }: { readonly state: ReturnType<typeof useSettingsModel>['state']; readonly onRetry: () => Promise<void> }) {
  const { palette, styles } = useSettingsListStyles();
  if (state.status === 'loading') return <View style={[styles.shell, styles.errorContainer]}><ActivityIndicator color={palette.action} /></View>;
  if (state.status === 'error') return <ScrollView contentContainerStyle={styles.errorContainer} style={styles.shell}><Text style={styles.errorTitle}>Could not load tenant context</Text><Text style={styles.errorMessage}>{state.message}</Text><Pressable accessibilityRole="button" onPress={() => void onRetry()} style={styles.retryButton}><Text style={styles.retryText}>Retry</Text></Pressable></ScrollView>;
  return null;
}

function selectedProfileLabel(slot: VoiceProviderSlot): string { return slot.selectedProfile?.displayName ?? 'No service selected'; }

async function selectProfile(manageCommand: ManageProviderProfileCommand, configuration: VoiceProviderConfiguration | undefined, slot: VoiceProviderSlot, profile: ProviderProfileSummary): Promise<void> {
  if (!configuration) return;
  await manageCommand.updateVoiceProviderConfiguration({
    speechToTextProfileId: slot.capability === 'speech_to_text' ? profile.id : configuration.profileIds.speechToText,
    languageInferenceProfileId: slot.capability === 'language_inference' ? profile.id : configuration.profileIds.languageInference,
    textToSpeechProfileId: slot.capability === 'text_to_speech' ? profile.id : configuration.profileIds.textToSpeech
  });
}
