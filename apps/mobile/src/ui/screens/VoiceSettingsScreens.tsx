import { t } from '../../presentation/localization';
import { useTaskPresentation } from '../navigation/useTaskPresentation';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { SettingsPickerRow } from '../components/SettingsPickerRow';
import { SettingsRefreshNotice } from './SettingsRefreshNotice';
import { useRef, useState } from 'react';
import {
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
  SettingsLoadingRow,
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
  if (providers.state.status !== 'ready') return <ProviderStateView taskLabel={t('mobile.VoiceSettingsScreens.voiceSetup')} state={providers.state} onRetry={providers.retry} />;
  if (settings.state.status !== 'ready') return <SettingsStateBridge state={settings.state} onRetry={settings.load} />;
  const { configuration } = providers.state.viewModel;
  const tenant = settings.state.settings.selectedTenant;
  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <SettingsRefreshNotice visible={providers.hasRefreshError} onRetry={providers.retry} />
      <View style={styles.detailHeader}>
        <Text accessibilityRole="header" style={styles.detailTitle}>
          {configuration.readiness === 'ready' ? t('mobile.VoiceSettingsScreens.voiceIsReady') : t('mobile.VoiceSettingsScreens.voiceNeedsAttention')}
        </Text>
        <Text style={styles.detailSubtitle}>{t('mobile.VoiceSettingsScreens.appliesToEveryoneIn')}{tenant.name}{t('mobile.VoiceSettingsScreens.configureHowStuffStashListensUnderstandsRequestsAndSpeaks')}</Text>
      </View>
      <SettingsSection title={t('mobile.VoiceSettingsScreens.voicePipeline')}>
        {configuration.slots.map((slot, index) => {
          const presentation = stagePresentation(slot.capability);
          return (
            <View key={slot.capability}>
              {index > 0 ? <SettingsSeparator /> : null}
              <SettingsNavigationRow
                accessibilityLabel={t('mobile.VoiceSettingsScreens.openVoiceStageFor', { title: String(presentation.title), name: String(tenant.name), value: String(formatVoiceProviderReadinessLabel(slot.readiness)) })}
                context={`${presentation.description} · ${selectedProfileLabel(slot)}`}
                label={presentation.title}
                onPress={() => onOpenCapability(slot.capability)}
                value={formatVoiceProviderReadinessLabel(slot.readiness)}
              />
            </View>
          );
        })}
      </SettingsSection>
      <SettingsSection footer={t('mobile.VoiceSettingsScreens.providerProfilesAreAdvancedTenantWideServiceConfigurations')} title={t('mobile.VoiceSettingsScreens.advanced')}>
        <SettingsNavigationRow
          accessibilityLabel={t('mobile.VoiceSettingsScreens.openAdvancedProviderProfilesFor', { name: String(tenant.name) })}
          label={t('mobile.VoiceSettingsScreens.providerProfiles')}
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
  const [operation, setOperation] = useState<'select' | 'test' | 'enable'>();
  const working = operation !== undefined;
  const workingRef = useRef(false);
  const capturePresentation = useTaskPresentation(manageCommand, `${providers.ownerKey}:${capability}`);
  const stage = stagePresentation(capability);
  const taskLabel = t('mobile.VoiceSettingsScreens.settings', { value: String(stage.title.toLowerCase()) });
  if (providers.state.status !== 'ready') return <ProviderStateView taskLabel={taskLabel} state={providers.state} onRetry={providers.retry} />;
  const slot = providers.state.viewModel.configuration.slots.find((item) => item.capability === capability);
  if (!slot) return <ProviderStateView taskLabel={taskLabel} state={{ status: 'error', message: t('mobile.VoiceSettingsScreens.thisVoiceStageIsNotAvailable') }} onRetry={providers.retry} />;
  const selectedProfile = slot.selectedProfile;
  const recommendedAction = slot.recommendedAction;
  const availableProfiles = providers.state.viewModel.profiles.filter(profile =>
    profile.capability === capability && (profile.lifecycleState !== 'archived' || profile.id === slot.selectedProfileId));
  const serviceOptions = availableProfiles.map(profile => ({ value: profile.id, label: profile.displayName }));
  if (selectedProfile && !serviceOptions.some(option => option.value === selectedProfile.id)) {
    serviceOptions.unshift({ value: selectedProfile.id, label: selectedProfile.displayName });
  }
  if (!slot.selectedProfileId) serviceOptions.unshift({ value: '', label: t('mobile.VoiceSettingsScreens.notSelected') });

  async function act(kind: 'select' | 'test' | 'enable', action: () => Promise<void>, success: string): Promise<void> {
    const canPresent = capturePresentation();
    if (!canPresent() || workingRef.current) return;
    workingRef.current = true;
    setOperation(kind);
    try {
      await action();
      if (!canPresent()) return;
      feedback.showNotice({ tone: 'success', title: success, message: t('mobile.VoiceSettingsScreens.setupWasUpdated', { title: String(stage.title) }) });
      void providers.load().catch(() => undefined);
    } catch (error) {
      if (!canPresent()) return;
      feedback.showNotice({ tone: 'error', title: t('mobile.VoiceSettingsScreens.couldNotUpdateVoice'), message: readableError(error) });
    } finally {
      workingRef.current = false;
      setOperation(undefined);
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
          accessibilityLabel: t('mobile.VoiceSettingsScreens.addProviderProfileFor', { title: String(stage.title) }),
          label: t('mobile.VoiceSettingsScreens.addProfile'),
          run: onAddProfile
        };
      case 'replace_credential':
        return selectedProfile
          ? {
              accessibilityLabel: t('mobile.VoiceSettingsScreens.addCredentialForIn', { displayName: String(selectedProfile.displayName), title: String(stage.title) }),
              label: t('mobile.VoiceSettingsScreens.addCredential'),
              run: () => onEditCredential(selectedProfile.id)
            }
          : undefined;
      case 'enable_profile':
        return selectedProfile
          ? {
              accessibilityLabel: t('mobile.VoiceSettingsScreens.enableFor', { displayName: String(selectedProfile.displayName), title: String(stage.title) }),
              label: operation === 'enable' ? t('mobile.VoiceSettingsScreens.enabling') : t('mobile.VoiceSettingsScreens.enableService'),
              run: () => void act('enable', () => manageCommand.changeLifecycle(selectedProfile.id, 'enable').then(() => undefined), 'Service enabled')
            }
          : undefined;
      case 'test_profile':
        return selectedProfile
          ? {
              accessibilityLabel: t('mobile.VoiceSettingsScreens.testFor', { displayName: String(selectedProfile.displayName), title: String(stage.title) }),
              label: operation === 'test' ? t('mobile.VoiceSettingsScreens.testing') : t('mobile.VoiceSettingsScreens.testConnection'),
              run: () => void act('test', () => testCommand.execute(selectedProfile.id).then(() => undefined), 'Connection tested')
            }
          : undefined;
      default:
        return undefined;
    }
  }

  const directAction = primaryAction();

  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <SettingsRefreshNotice visible={providers.hasRefreshError} onRetry={providers.retry} />
      <View style={styles.detailHeader}>
        <Text accessibilityRole="header" style={styles.detailTitle}>{stage.title}</Text>
        <Text style={styles.detailSubtitle}>{stage.longDescription}</Text>
      </View>
      <SettingsSection title={t('mobile.VoiceSettingsScreens.selectedService')}>
        {slot.selectedProfile ? (
          <SettingsNavigationRow
            accessibilityLabel={t('mobile.VoiceSettingsScreens.openProviderProfile', { displayName: String(slot.selectedProfile.displayName) })}
            context={`${slot.selectedProfile.providerKind} · ${slot.selectedProfile.modelName || 'Default model'}`}
            label={slot.selectedProfile.displayName}
            disabled={working}
            onPress={() => { if (!workingRef.current) onEditProfile(slot.selectedProfile!.id); }}
            value={formatVoiceProviderReadinessLabel(slot.readiness)}
          />
        ) : <SettingsValueRow label={t('mobile.VoiceSettingsScreens.service')} value="Not selected" />}
      </SettingsSection>
      <SettingsSection
        footer={issueLabels.length > 0 ? issueLabels.join(' ') : undefined}
        title={t('mobile.VoiceSettingsScreens.setupStatus')}
      >
        <SettingsValueRow label={t('mobile.VoiceSettingsScreens.selection')} value={formatVoiceProviderSelectionSourceLabel(slot.selectionSource)} />
        {slot.duplicateProfiles.length > 0 ? <><SettingsSeparator /><SettingsValueRow label={t('mobile.VoiceSettingsScreens.readyChoices')} value={`${slot.duplicateProfiles.length}`} /></> : null}
      </SettingsSection>
      {directAction ? (
        <SettingsSection>
          <SettingsActionRow
            accessibilityLabel={directAction.accessibilityLabel}
            disabled={working}
            label={directAction.label}
            onPress={() => { if (!workingRef.current) directAction.run(); }}
          />
        </SettingsSection>
      ) : null}
      {availableProfiles.length > 0 ? <SettingsSection footer={t('mobile.VoiceSettingsScreens.changingTheSelectionAffectsVoiceForEveryoneInThis')}>
        <SettingsPickerRow label={t('mobile.VoiceSettingsScreens.service')} accessibilityLabel={t('mobile.VoiceSettingsScreens.chooseVoiceService')} value={slot.selectedProfileId ?? ''}
          options={serviceOptions} disabled={working} onChange={id => {
            if (workingRef.current || id === slot.selectedProfileId) return;
            const profile = availableProfiles.find(candidate => candidate.id === id);
            if (profile) void act('select', () => selectProfile(manageCommand, providers.state.status === 'ready' ? providers.state.viewModel.configuration : undefined, slot, profile), 'Voice service selected');
          }} />
        {operation === 'select' ? <Text accessibilityLiveRegion="polite" style={styles.secondaryText}>{t('mobile.VoiceSettingsScreens.selectingService')}</Text> : null}
      </SettingsSection> : null}
    </ScrollView>
  );
}

function SettingsStateBridge({ state, onRetry }: { readonly state: ReturnType<typeof useSettingsModel>['state']; readonly onRetry: () => Promise<void> }) {
  const { styles } = useSettingsListStyles();
  if (state.status === 'loading') return <View style={[styles.shell, styles.errorContainer]}><SettingsLoadingRow label={t('mobile.VoiceSettingsScreens.loadingHouseholdContext')} /></View>;
  if (state.status === 'error') return <ScrollView contentContainerStyle={styles.errorContainer} style={styles.shell}><Text style={styles.errorTitle}>{t('mobile.VoiceSettingsScreens.couldNotLoadTenantContext')}</Text><Text style={styles.errorMessage}>{state.message}</Text><NativeCommandButton label={t('mobile.VoiceSettingsScreens.retry')} onPress={() => void onRetry()} /></ScrollView>;
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
