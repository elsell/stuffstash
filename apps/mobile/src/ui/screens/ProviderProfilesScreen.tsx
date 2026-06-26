import { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Pressable,
  RefreshControl,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { ManageProviderProfileCommand } from '../../application/providerProfiles/ManageProviderProfileCommand';
import {
  ProviderProfileSettingsQuery,
  ProviderProfileSettingsViewModel
} from '../../application/providerProfiles/ProviderProfileSettingsQuery';
import {
  ProviderProfileLifecycleAction,
  ProviderProfileSummary
} from '../../application/providerProfiles/ProviderProfileRepository';
import { recommendedProviderProfiles } from '../../application/providerProfiles/RecommendedProviderProfiles';
import { TestProviderProfileCommand } from '../../application/providerProfiles/TestProviderProfileCommand';
import { colors, radius, spacing } from '../theme/tokens';
import {
  buildCredentialEditorPresentation,
  buildPromptEditorPresentation
} from './ProviderProfilesScreenPresentation';

type ProviderProfilesScreenProps = {
  readonly manageCommand: ManageProviderProfileCommand;
  readonly query: ProviderProfileSettingsQuery;
  readonly testCommand: TestProviderProfileCommand;
};

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly viewModel: ProviderProfileSettingsViewModel }
  | { readonly status: 'error'; readonly message: string };

type CredentialEditorState = {
  readonly profileId: string;
  readonly profileName: string;
  readonly purpose: 'api_key' | 'oauth_bearer';
  readonly value: string;
};

type PromptEditorState = {
  readonly profileId: string;
  readonly profileName: string;
  readonly value: string;
};

export function ProviderProfilesScreen({
  manageCommand,
  query,
  testCommand
}: ProviderProfilesScreenProps) {
  const [screenState, setScreenState] = useState<ScreenState>({ status: 'loading' });
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [testingProfileId, setTestingProfileId] = useState<string | undefined>();
  const [workingAction, setWorkingAction] = useState<string | undefined>();
  const [lastResultByProfileId, setLastResultByProfileId] = useState<Record<string, string>>({});
  const [credentialEditor, setCredentialEditor] = useState<CredentialEditorState | undefined>();
  const [promptEditor, setPromptEditor] = useState<PromptEditorState | undefined>();

  useEffect(() => {
    let isCurrent = true;

    query
      .execute()
      .then((viewModel) => {
        if (isCurrent) {
          setScreenState({ status: 'ready', viewModel });
        }
      })
      .catch((error: unknown) => {
        if (isCurrent) {
          setScreenState({
            status: 'error',
            message: readableError(error, 'Stuff Stash could not load provider profiles.')
          });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [query]);

  async function refresh(): Promise<void> {
    setIsRefreshing(true);
    try {
      const viewModel = await query.execute();
      setScreenState({ status: 'ready', viewModel });
    } catch (error) {
      setScreenState({
        status: 'error',
        message: readableError(error, 'Stuff Stash could not refresh provider profiles.')
      });
    } finally {
      setIsRefreshing(false);
    }
  }

  async function runWorkingAction(actionKey: string, action: () => Promise<void>): Promise<void> {
    setWorkingAction(actionKey);
    try {
      await action();
    } catch (error) {
      Alert.alert('Provider profile action failed', readableError(error, 'The action failed safely.'));
    } finally {
      setWorkingAction(undefined);
    }
  }

  async function createRecommendedProfile(templateKey: string): Promise<void> {
    const template = recommendedProviderProfiles.find((item) => item.key === templateKey);
    if (!template) {
      return;
    }

    await runWorkingAction(`create-${template.key}`, async () => {
      const profile = await manageCommand.createRecommended(template);
      setCredentialEditor({
        profileId: profile.id,
        profileName: profile.displayName,
        purpose: template.credentialPurpose,
        value: ''
      });
      await refresh();
    });
  }

  async function changeLifecycle(
    profile: ProviderProfileSummary,
    action: ProviderProfileLifecycleAction
  ): Promise<void> {
    const perform = async () => {
      await runWorkingAction(`${action}-${profile.id}`, async () => {
        await manageCommand.changeLifecycle(profile.id, action);
        await refresh();
      });
    };

    if (action === 'archive') {
      Alert.alert('Archive provider profile', `Archive ${profile.displayName}?`, [
        { text: 'Cancel', style: 'cancel' },
        { text: 'Archive', style: 'destructive', onPress: () => void perform() }
      ]);
      return;
    }

    await perform();
  }

  async function replaceCredential(): Promise<void> {
    if (!credentialEditor) {
      return;
    }

    const editor = credentialEditor;
    await runWorkingAction(`credential-${editor.profileId}`, async () => {
      await manageCommand.replaceCredential({
        providerProfileId: editor.profileId,
        purpose: editor.purpose,
        credential: editor.value
      });
      setCredentialEditor(undefined);
      await refresh();
    });
  }

  async function savePromptTemplate(): Promise<void> {
    if (!promptEditor) {
      return;
    }

    const editor = promptEditor;
    await runWorkingAction(`prompt-${editor.profileId}`, async () => {
      await manageCommand.replacePromptTemplate({
        providerProfileId: editor.profileId,
        promptTemplate: editor.value
      });
      setPromptEditor(undefined);
      await refresh();
    });
  }

  async function testProfile(profile: ProviderProfileSummary): Promise<void> {
    setTestingProfileId(profile.id);
    try {
      const result = await testCommand.execute(profile.id);
      setLastResultByProfileId((current) => ({
        ...current,
        [profile.id]: `${result.status}: ${result.message}`
      }));
      await refresh();
    } catch (error) {
      Alert.alert('Provider test failed', readableError(error, 'The provider test failed safely.'));
    } finally {
      setTestingProfileId(undefined);
    }
  }

  return (
    <SafeAreaView style={styles.shell} edges={['left', 'right']}>
      {screenState.status === 'loading' ? <LoadingState /> : null}
      {screenState.status === 'error' ? <ErrorState message={screenState.message} /> : null}
      {screenState.status === 'ready' ? (
        <ProviderProfileContent
          credentialEditor={credentialEditor}
          isRefreshing={isRefreshing}
          lastResultByProfileId={lastResultByProfileId}
          promptEditor={promptEditor}
          testingProfileId={testingProfileId}
          viewModel={screenState.viewModel}
          workingAction={workingAction}
          onCancelCredential={() => setCredentialEditor(undefined)}
          onCancelPrompt={() => setPromptEditor(undefined)}
          onChangeCredential={(value) =>
            setCredentialEditor((current) => (current ? { ...current, value } : current))
          }
          onChangeLifecycle={changeLifecycle}
          onChangePrompt={(value) =>
            setPromptEditor((current) => (current ? { ...current, value } : current))
          }
          onCreateRecommendedProfile={createRecommendedProfile}
          onEditCredential={(profile) => {
            if (!profile.credentialPurpose) {
              Alert.alert(
                'Credential purpose unknown',
                'This profile was not created from the mobile recommended path. Replace its credential through the API for now.'
              );
              return;
            }
            setCredentialEditor(buildCredentialEditorPresentation(profile));
          }}
          onEditPrompt={(profile) =>
            setPromptEditor(buildPromptEditorPresentation(profile))
          }
          onRefresh={refresh}
          onReplaceCredential={replaceCredential}
          onSavePrompt={savePromptTemplate}
          onTestProfile={testProfile}
        />
      ) : null}
    </SafeAreaView>
  );
}

function ProviderProfileContent({
  credentialEditor,
  isRefreshing,
  lastResultByProfileId,
  promptEditor,
  testingProfileId,
  viewModel,
  workingAction,
  onCancelCredential,
  onCancelPrompt,
  onChangeCredential,
  onChangeLifecycle,
  onChangePrompt,
  onCreateRecommendedProfile,
  onEditCredential,
  onEditPrompt,
  onRefresh,
  onReplaceCredential,
  onSavePrompt,
  onTestProfile
}: {
  readonly credentialEditor?: CredentialEditorState;
  readonly isRefreshing: boolean;
  readonly lastResultByProfileId: Record<string, string>;
  readonly promptEditor?: PromptEditorState;
  readonly testingProfileId?: string;
  readonly viewModel: ProviderProfileSettingsViewModel;
  readonly workingAction?: string;
  readonly onCancelCredential: () => void;
  readonly onCancelPrompt: () => void;
  readonly onChangeCredential: (value: string) => void;
  readonly onChangeLifecycle: (profile: ProviderProfileSummary, action: ProviderProfileLifecycleAction) => void;
  readonly onChangePrompt: (value: string) => void;
  readonly onCreateRecommendedProfile: (templateKey: string) => void;
  readonly onEditCredential: (profile: ProviderProfileSummary) => void;
  readonly onEditPrompt: (profile: ProviderProfileSummary) => void;
  readonly onRefresh: () => void;
  readonly onReplaceCredential: () => void;
  readonly onSavePrompt: () => void;
  readonly onTestProfile: (profile: ProviderProfileSummary) => void;
}) {
  return (
    <ScrollView
      contentContainerStyle={styles.content}
      refreshControl={
        <RefreshControl
          refreshing={isRefreshing}
          tintColor={colors.action}
          onRefresh={onRefresh}
        />
      }
    >
      <Text style={styles.title}>Voice providers</Text>
      {viewModel.missingCapabilities.length > 0 ? (
        <View style={styles.warningPanel}>
          <Text style={styles.warningTitle}>Voice setup incomplete</Text>
          <Text style={styles.warningText}>
            Missing, disabled, or untested profiles: {viewModel.missingCapabilities.map(formatCapability).join(', ')}
          </Text>
        </View>
      ) : (
        <View style={styles.readyPanel}>
          <Text style={styles.readyTitle}>Ready for voice testing</Text>
          <Text style={styles.readyText}>Speech, language, and voice output profiles are enabled and tested.</Text>
        </View>
      )}

      {viewModel.profiles.length === 0 ? (
        <View style={styles.emptyPanel}>
          <Text style={styles.emptyTitle}>No profiles yet</Text>
          <Text style={styles.emptyText}>Add the recommended profiles below to configure tenant-managed voice.</Text>
        </View>
      ) : null}

      <View style={styles.panel}>
        <Text style={styles.sectionTitle}>Add recommended profile</Text>
        {recommendedProviderProfiles.map((template) => (
          <Pressable
            accessibilityRole="button"
            disabled={Boolean(workingAction)}
            key={template.key}
            onPress={() => onCreateRecommendedProfile(template.key)}
            style={styles.templateButton}
          >
            <Text style={styles.templateTitle}>{template.title}</Text>
            <Text style={styles.templateText}>{template.description}</Text>
          </Pressable>
        ))}
      </View>

      {credentialEditor ? (
        <InlineSecretEditor
          editor={credentialEditor}
          isSaving={workingAction === `credential-${credentialEditor.profileId}`}
          onCancel={onCancelCredential}
          onChange={onChangeCredential}
          onSave={onReplaceCredential}
        />
      ) : null}

      {promptEditor ? (
        <InlinePromptEditor
          editor={promptEditor}
          isSaving={workingAction === `prompt-${promptEditor.profileId}`}
          onCancel={onCancelPrompt}
          onChange={onChangePrompt}
          onSave={onSavePrompt}
        />
      ) : null}

      {viewModel.profiles.map((profile) => (
        <ProviderProfileCard
          key={profile.id}
          lastResult={lastResultByProfileId[profile.id]}
          profile={profile}
          isTesting={testingProfileId === profile.id}
          workingAction={workingAction}
          onChangeLifecycle={(action) => onChangeLifecycle(profile, action)}
          onEditCredential={() => onEditCredential(profile)}
          onEditPrompt={() => onEditPrompt(profile)}
          onTest={() => onTestProfile(profile)}
        />
      ))}
    </ScrollView>
  );
}

function ProviderProfileCard({
  isTesting,
  lastResult,
  profile,
  workingAction,
  onChangeLifecycle,
  onEditCredential,
  onEditPrompt,
  onTest
}: {
  readonly isTesting: boolean;
  readonly lastResult?: string;
  readonly profile: ProviderProfileSummary;
  readonly workingAction?: string;
  readonly onChangeLifecycle: (action: ProviderProfileLifecycleAction) => void;
  readonly onEditCredential: () => void;
  readonly onEditPrompt: () => void;
  readonly onTest: () => void;
}) {
  const lifecycleAction = profile.lifecycleState === 'enabled' ? 'disable' : 'enable';

  return (
    <View style={styles.card}>
      <View style={styles.cardHeader}>
        <View style={styles.cardTitleGroup}>
          <Text style={styles.cardTitle}>{profile.displayName}</Text>
          <Text style={styles.cardSubtitle}>{formatCapability(profile.capability)} / {profile.providerKind}</Text>
        </View>
        <StatusPill label={profile.lifecycleState} tone={profile.lifecycleState === 'enabled' ? 'success' : 'muted'} />
      </View>
      <InfoRow label="Credential" value={profile.credentialStatus} />
      <InfoRow label="Model" value={profile.modelName || 'Not set'} />
      <InfoRow label="Last tested" value={profile.lastTestedAt ?? 'Not tested'} />
      {profile.capability === 'language_inference' ? (
        <InfoRow label="Prompt template" value={profile.hasPromptTemplate ? 'Configured' : 'Default'} />
      ) : null}
      {lastResult ? <Text style={styles.lastResult}>{lastResult}</Text> : null}
      <View style={styles.actionGrid}>
        {profile.credentialPurpose ? (
          <SecondaryActionButton label="Credential" onPress={onEditCredential} />
        ) : null}
        {profile.capability === 'language_inference' ? (
          <SecondaryActionButton label="Prompt" onPress={onEditPrompt} />
        ) : null}
        {profile.lifecycleState !== 'archived' ? (
          <SecondaryActionButton
            label={lifecycleAction === 'enable' ? 'Enable' : 'Disable'}
            isBusy={workingAction === `${lifecycleAction}-${profile.id}`}
            onPress={() => onChangeLifecycle(lifecycleAction)}
          />
        ) : null}
        {profile.lifecycleState !== 'archived' ? (
          <SecondaryActionButton
            label="Archive"
            isDestructive
            isBusy={workingAction === `archive-${profile.id}`}
            onPress={() => onChangeLifecycle('archive')}
          />
        ) : null}
      </View>
      <Pressable
        accessibilityRole="button"
        disabled={isTesting}
        onPress={onTest}
        style={[styles.testButton, isTesting ? styles.testButtonDisabled : null]}
      >
        {isTesting ? (
          <ActivityIndicator color={colors.onAction} />
        ) : (
          <Text style={styles.testButtonText}>Test profile</Text>
        )}
      </Pressable>
    </View>
  );
}

function InlineSecretEditor({
  editor,
  isSaving,
  onCancel,
  onChange,
  onSave
}: {
  readonly editor: CredentialEditorState;
  readonly isSaving: boolean;
  readonly onCancel: () => void;
  readonly onChange: (value: string) => void;
  readonly onSave: () => void;
}) {
  return (
    <View style={styles.panel}>
      <Text style={styles.sectionTitle}>Replace credential</Text>
      <Text style={styles.editorHint}>{editor.profileName} / {editor.purpose}</Text>
      <TextInput
        autoCapitalize="none"
        autoCorrect={false}
        placeholder="Paste credential"
        placeholderTextColor={colors.textMuted}
        secureTextEntry
        style={styles.textInput}
        value={editor.value}
        onChangeText={onChange}
      />
      <EditorActions isSaving={isSaving} onCancel={onCancel} onSave={onSave} />
    </View>
  );
}

function InlinePromptEditor({
  editor,
  isSaving,
  onCancel,
  onChange,
  onSave
}: {
  readonly editor: PromptEditorState;
  readonly isSaving: boolean;
  readonly onCancel: () => void;
  readonly onChange: (value: string) => void;
  readonly onSave: () => void;
}) {
  return (
    <View style={styles.panel}>
      <Text style={styles.sectionTitle}>Replace prompt template</Text>
      <Text style={styles.editorHint}>{editor.profileName}</Text>
      <TextInput
        multiline
        placeholder="Optional tenant guidance for this language model"
        placeholderTextColor={colors.textMuted}
        style={[styles.textInput, styles.promptInput]}
        value={editor.value}
        onChangeText={onChange}
      />
      <EditorActions isSaving={isSaving} onCancel={onCancel} onSave={onSave} />
    </View>
  );
}

function EditorActions({
  isSaving,
  onCancel,
  onSave
}: {
  readonly isSaving: boolean;
  readonly onCancel: () => void;
  readonly onSave: () => void;
}) {
  return (
    <View style={styles.editorActions}>
      <SecondaryActionButton label="Cancel" onPress={onCancel} />
      <Pressable
        accessibilityRole="button"
        disabled={isSaving}
        onPress={onSave}
        style={[styles.editorSaveButton, isSaving ? styles.testButtonDisabled : null]}
      >
        {isSaving ? (
          <ActivityIndicator color={colors.onAction} />
        ) : (
          <Text style={styles.testButtonText}>Save</Text>
        )}
      </Pressable>
    </View>
  );
}

function SecondaryActionButton({
  isBusy,
  isDestructive,
  label,
  onPress
}: {
  readonly isBusy?: boolean;
  readonly isDestructive?: boolean;
  readonly label: string;
  readonly onPress: () => void;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      disabled={isBusy}
      onPress={onPress}
      style={styles.secondaryActionButton}
    >
      {isBusy ? (
        <ActivityIndicator color={colors.action} />
      ) : (
        <Text style={[styles.secondaryActionText, isDestructive ? styles.destructiveActionText : null]}>{label}</Text>
      )}
    </Pressable>
  );
}

function StatusPill({ label, tone }: { readonly label: string; readonly tone: 'success' | 'muted' }) {
  return (
    <View style={[styles.statusPill, tone === 'success' ? styles.statusPillSuccess : null]}>
      <Text style={[styles.statusPillText, tone === 'success' ? styles.statusPillTextSuccess : null]}>{label}</Text>
    </View>
  );
}

function InfoRow({ label, value }: { readonly label: string; readonly value: string }) {
  return (
    <View style={styles.infoRow}>
      <Text style={styles.infoLabel}>{label}</Text>
      <Text style={styles.infoValue}>{value}</Text>
    </View>
  );
}

function LoadingState() {
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={colors.accent} />
      <Text style={styles.stateText}>Loading provider profiles</Text>
    </View>
  );
}

function ErrorState({ message }: { readonly message: string }) {
  return (
    <View style={styles.centerState}>
      <Text style={styles.errorTitle}>Could not load</Text>
      <Text style={styles.stateText}>{message}</Text>
    </View>
  );
}

function formatCapability(value: string): string {
  return value.replace(/_/g, ' ');
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

const styles = StyleSheet.create({
  shell: {
    backgroundColor: colors.background,
    flex: 1
  },
  content: {
    padding: spacing.lg,
    paddingBottom: spacing.xl
  },
  title: {
    color: colors.text,
    fontSize: 30,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 36,
    marginBottom: spacing.md
  },
  warningPanel: {
    backgroundColor: colors.warningSurface,
    borderColor: colors.brandAmber,
    borderRadius: radius.md,
    borderWidth: 1,
    marginBottom: spacing.md,
    padding: spacing.md
  },
  warningTitle: {
    color: colors.warning,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  },
  warningText: {
    color: colors.warning,
    fontSize: 14,
    fontWeight: '700',
    letterSpacing: 0,
    lineHeight: 20,
    marginTop: spacing.xs
  },
  readyPanel: {
    backgroundColor: colors.brandDustyBlueSoft,
    borderColor: colors.accent,
    borderRadius: radius.md,
    borderWidth: 1,
    marginBottom: spacing.md,
    padding: spacing.md
  },
  readyTitle: {
    color: colors.text,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  },
  readyText: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '700',
    lineHeight: 20,
    marginTop: spacing.xs
  },
  emptyPanel: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    marginBottom: spacing.md,
    padding: spacing.md
  },
  emptyTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '900',
    letterSpacing: 0
  },
  emptyText: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '700',
    lineHeight: 20,
    marginTop: spacing.xs
  },
  panel: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    marginBottom: spacing.md,
    padding: spacing.md
  },
  sectionTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '900',
    letterSpacing: 0,
    marginBottom: spacing.sm
  },
  templateButton: {
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    marginTop: spacing.sm,
    padding: spacing.md
  },
  templateTitle: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  templateText: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '700',
    lineHeight: 19,
    marginTop: spacing.xs
  },
  editorHint: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '800',
    lineHeight: 19,
    marginBottom: spacing.sm
  },
  textInput: {
    backgroundColor: colors.background,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    color: colors.text,
    fontSize: 16,
    fontWeight: '700',
    minHeight: 46,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  promptInput: {
    minHeight: 112,
    textAlignVertical: 'top'
  },
  editorActions: {
    flexDirection: 'row',
    gap: spacing.sm,
    marginTop: spacing.md
  },
  editorSaveButton: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    flex: 1,
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  card: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    marginBottom: spacing.md,
    padding: spacing.md
  },
  cardHeader: {
    alignItems: 'flex-start',
    flexDirection: 'row',
    gap: spacing.sm,
    justifyContent: 'space-between',
    marginBottom: spacing.sm
  },
  cardTitleGroup: {
    flex: 1,
    minWidth: 0
  },
  cardTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '900',
    letterSpacing: 0
  },
  cardSubtitle: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '800',
    letterSpacing: 0,
    marginTop: 2,
    textTransform: 'capitalize'
  },
  statusPill: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  statusPillSuccess: {
    backgroundColor: colors.brandDustyBlueSoft
  },
  statusPillText: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  statusPillTextSuccess: {
    color: colors.success
  },
  infoRow: {
    borderTopColor: colors.border,
    borderTopWidth: 1,
    gap: spacing.xs,
    paddingVertical: spacing.sm
  },
  infoLabel: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  infoValue: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0,
    lineHeight: 21
  },
  lastResult: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '700',
    lineHeight: 20,
    marginTop: spacing.xs
  },
  actionGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm,
    marginTop: spacing.sm
  },
  secondaryActionButton: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flexGrow: 1,
    justifyContent: 'center',
    minHeight: 38,
    minWidth: 92,
    paddingHorizontal: spacing.sm
  },
  secondaryActionText: {
    color: colors.action,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0
  },
  destructiveActionText: {
    color: colors.danger
  },
  testButton: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    justifyContent: 'center',
    marginTop: spacing.md,
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  testButtonDisabled: {
    opacity: 0.68
  },
  testButtonText: {
    color: colors.onAction,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  centerState: {
    alignItems: 'center',
    flex: 1,
    justifyContent: 'center',
    padding: spacing.lg
  },
  stateText: {
    color: colors.textMuted,
    fontSize: 16,
    lineHeight: 23,
    marginTop: spacing.md,
    textAlign: 'center'
  },
  errorTitle: {
    color: colors.text,
    fontSize: 24,
    fontWeight: '800',
    letterSpacing: 0
  }
});
