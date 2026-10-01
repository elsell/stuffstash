import { t } from '../../presentation/localization';
import { NativeCommandButton } from '../components/NativeCommandButton';
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
    return <ProviderStateView taskLabel={t('mobile.ProviderProfileScreens.providerProfiles')} state={providers.state} onRetry={providers.retry} />;
  }

  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
      <SettingsRefreshNotice visible={providers.hasRefreshError} onRetry={providers.retry} />
      <SettingsSection footer={t('mobile.ProviderProfileScreens.profilesAreSharedTenantWideAndSupplyOneStage')}>
        <SettingsNavigationRow accessibilityLabel={t('mobile.ProviderProfileScreens.addAProviderProfile')} label={t('mobile.ProviderProfileScreens.addProfile')} onPress={onAdd} />
      </SettingsSection>
      <SettingsSection title={t('mobile.ProviderProfileScreens.profiles')}>
        {providers.state.viewModel.profiles.length === 0 ? (
          <SettingsValueRow label={t('mobile.ProviderProfileScreens.noProfiles')} value={t('mobile.ProviderProfileScreens.addOneToBegin')} />
        ) : providers.state.viewModel.profiles.map((profile, index) => (
          <View key={profile.id}>
            {index > 0 ? <SettingsSeparator /> : null}
            <SettingsNavigationRow
              accessibilityLabel={t('mobile.ProviderProfileScreens.openProviderProfile', { displayName: String(profile.displayName), value: String(formatProviderProfileLifecycleLabel(profile.lifecycleState)) })}
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
        title: t('mobile.ProviderProfileScreens.draftProfileCreated'),
        message: t('mobile.ProviderProfileScreens.addCredentialsTestItThenEnableItForVoice')
      });
      onCreated(profile.id);
    } catch (error) {
      if (!canPresent()) return;
      feedback.showNotice({
        tone: 'error',
        title: t('mobile.ProviderProfileScreens.couldNotCreateProfile'),
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
        <Text accessibilityRole="header" style={styles.detailTitle}>{t('mobile.ProviderProfileScreens.chooseAService')}</Text>
        <Text style={styles.detailSubtitle}>{t('mobile.ProviderProfileScreens.stuffStashCreatesADisabledDraftFirstYouLl')}</Text>
      </View>
      <SettingsSection title={t('mobile.ProviderProfileScreens.recommended')}>
        {recommendedProviderProfiles.map((template, index) => (
          <View key={template.key}>
            {index > 0 ? <SettingsSeparator /> : null}
            <NativeCommandButton
              disabled={workingKey !== undefined}
              label={workingKey === template.key ? t('mobile.ProviderProfileScreens.creating', { title: String(template.title) }) : t('mobile.ProviderProfileScreens.create', { title: String(template.title) })}
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
    return <ProviderStateView taskLabel={t('mobile.ProviderProfileScreens.providerProfile')} state={providers.state} onRetry={providers.retry} />;
  }
  const profile = providers.state.viewModel.profiles.find((item) => item.id === profileId);
  if (!profile) {
    return (
      <ProviderStateView
        taskLabel={t('mobile.ProviderProfileScreens.providerProfile')}
        state={{ status: 'error', message: t('mobile.ProviderProfileScreens.thisProviderProfileIsNoLongerAvailable') }}
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
        message: t('mobile.ProviderProfileScreens.wasUpdated', { profileDisplayName: String(profileDisplayName) })
      });
      void providers.load().catch(() => undefined);
    } catch (error) {
      if (!canPresent()) return;
      feedback.showNotice({
        tone: 'error',
        title: t('mobile.ProviderProfileScreens.profileActionFailed'),
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
      <SettingsSection title={t('mobile.ProviderProfileScreens.configuration')}>
        <SettingsValueRow label={t('mobile.ProviderProfileScreens.model')} value={profile.modelName || t('mobile.ProviderProfileScreens.default')} />
        <SettingsSeparator />
        <SettingsValueRow label={t('mobile.ProviderProfileScreens.status')} value={formatProviderProfileLifecycleLabel(profile.lifecycleState)} />
        <SettingsSeparator />
        <SettingsValueRow label={t('mobile.ProviderProfileScreens.credential')} value={formatProviderProfileCredentialStatusLabel(profile.credentialStatus)} />
        <SettingsSeparator />
        <SettingsValueRow label={t('mobile.ProviderProfileScreens.lastTested')} value={formatProviderProfileTestStatusLabel(profile.lastTestedAt)} />
      </SettingsSection>
      <SettingsSection title={t('mobile.ProviderProfileScreens.actions')}>
        {profile.credentialPurpose ? <><SettingsNavigationRow accessibilityLabel={t('mobile.ProviderProfileScreens.replaceCredentialFor', { displayName: String(profile.displayName) })} label={t('mobile.ProviderProfileScreens.replaceCredential')} disabled={working} onPress={() => { if (!workingRef.current) onEditCredential(); }} /><SettingsSeparator /></> : null}
        {profile.capability === 'language_inference' ? <><SettingsNavigationRow accessibilityLabel={t('mobile.ProviderProfileScreens.editPromptGuidanceFor', { displayName: String(profile.displayName) })} label={t('mobile.ProviderProfileScreens.promptGuidance')} disabled={working} onPress={() => { if (!workingRef.current) onEditPrompt(); }} /><SettingsSeparator /></> : null}
        <NativeCommandButton disabled={working} label={operation === 'test' ? t('mobile.ProviderProfileScreens.testing') : t('mobile.ProviderProfileScreens.testConnection')} onPress={() => void act('test', () => testCommand.execute(profile.id), t('mobile.ProviderProfileScreens.connectionTested'))} />
        {profile.lifecycleState !== 'archived' ? <><SettingsSeparator /><NativeCommandButton disabled={working} label={operation === 'lifecycle' ? t('mobile.ProviderProfileScreens.updating') : lifecycleAction === 'enable' ? t('mobile.ProviderProfileScreens.enableProfile') : t('mobile.ProviderProfileScreens.disableProfile')} onPress={() => void act('lifecycle', () => manageCommand.changeLifecycle(profile.id, lifecycleAction), lifecycleAction === 'enable' ? t('mobile.ProviderProfileScreens.profileEnabled') : t('mobile.ProviderProfileScreens.profileDisabled'))} /></> : null}
      </SettingsSection>
      {profile.lifecycleState !== 'archived' ? (
        <SettingsSection footer={t('mobile.ProviderProfileScreens.archivedProfilesRemainInHistoryButCanTBe')}>
          <NativeCommandButton disabled={working} label={operation === 'archive' ? t('mobile.ProviderProfileScreens.archiving') : t('mobile.ProviderProfileScreens.archiveProfile')} onPress={() => {
            const canPresent = capturePresentation();
            if (!canPresent() || workingRef.current) return;
            let confirmed = false;
            confirmArchive(profile, async () => {
              if (confirmed || !canPresent() || workingRef.current) return;
              confirmed = true;
              await act('archive', () => manageCommand.changeLifecycle(profile.id, 'archive'), t('mobile.ProviderProfileScreens.profileArchived'));
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
    t('mobile.ProviderProfileScreens.archiveProviderProfile'),
    t('mobile.ProviderProfileScreens.willNoLongerBeAvailableForVoice', { displayName: String(profile.displayName) }),
    [
      { text: t('mobile.ProviderProfileScreens.cancel'), style: 'cancel' },
      { text: t('mobile.ProviderProfileScreens.archive'), style: 'destructive', onPress: () => void archive() }
    ]
  );
}
