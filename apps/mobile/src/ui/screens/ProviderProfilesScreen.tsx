import { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Pressable,
  RefreshControl,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import {
  ProviderProfileSettingsQuery,
  ProviderProfileSettingsViewModel
} from '../../application/providerProfiles/ProviderProfileSettingsQuery';
import { ProviderProfileSummary } from '../../application/providerProfiles/ProviderProfileRepository';
import { TestProviderProfileCommand } from '../../application/providerProfiles/TestProviderProfileCommand';
import { colors, radius, spacing } from '../theme/tokens';

type ProviderProfilesScreenProps = {
  readonly query: ProviderProfileSettingsQuery;
  readonly testCommand: TestProviderProfileCommand;
};

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly viewModel: ProviderProfileSettingsViewModel }
  | { readonly status: 'error'; readonly message: string };

export function ProviderProfilesScreen({ query, testCommand }: ProviderProfilesScreenProps) {
  const [screenState, setScreenState] = useState<ScreenState>({ status: 'loading' });
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [testingProfileId, setTestingProfileId] = useState<string | undefined>();
  const [lastResultByProfileId, setLastResultByProfileId] = useState<Record<string, string>>({});

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
          isRefreshing={isRefreshing}
          lastResultByProfileId={lastResultByProfileId}
          testingProfileId={testingProfileId}
          viewModel={screenState.viewModel}
          onRefresh={refresh}
          onTestProfile={testProfile}
        />
      ) : null}
    </SafeAreaView>
  );
}

function ProviderProfileContent({
  isRefreshing,
  lastResultByProfileId,
  testingProfileId,
  viewModel,
  onRefresh,
  onTestProfile
}: {
  readonly isRefreshing: boolean;
  readonly lastResultByProfileId: Record<string, string>;
  readonly testingProfileId?: string;
  readonly viewModel: ProviderProfileSettingsViewModel;
  readonly onRefresh: () => void;
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
          <Text style={styles.readyText}>Speech, language, and voice output profiles are enabled.</Text>
        </View>
      )}

      {viewModel.profiles.length === 0 ? (
        <View style={styles.emptyPanel}>
          <Text style={styles.emptyTitle}>No profiles yet</Text>
          <Text style={styles.emptyText}>Create tenant provider profiles through the API before using tenant-managed voice.</Text>
        </View>
      ) : (
        viewModel.profiles.map((profile) => (
          <ProviderProfileCard
            key={profile.id}
            lastResult={lastResultByProfileId[profile.id]}
            profile={profile}
            isTesting={testingProfileId === profile.id}
            onTest={() => onTestProfile(profile)}
          />
        ))
      )}
    </ScrollView>
  );
}

function ProviderProfileCard({
  isTesting,
  lastResult,
  profile,
  onTest
}: {
  readonly isTesting: boolean;
  readonly lastResult?: string;
  readonly profile: ProviderProfileSummary;
  readonly onTest: () => void;
}) {
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
