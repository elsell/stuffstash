import { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  RefreshControl,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { SettingsQuery, SettingsViewModel } from '../../application/settings/SettingsQuery';
import { BrandMark } from '../components/BrandMark';
import { colors, radius, spacing } from '../theme/tokens';

type SettingsScreenProps = {
  readonly settingsQuery: SettingsQuery;
};

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly settings: SettingsViewModel }
  | { readonly status: 'error'; readonly message: string };

export function SettingsScreen({ settingsQuery }: SettingsScreenProps) {
  const [screenState, setScreenState] = useState<ScreenState>({ status: 'loading' });
  const [isRefreshing, setIsRefreshing] = useState(false);

  useEffect(() => {
    let isCurrent = true;

    settingsQuery
      .execute()
      .then((settings) => {
        if (isCurrent) {
          setScreenState({ status: 'ready', settings });
        }
      })
      .catch((error: unknown) => {
        if (isCurrent) {
          setScreenState({
            status: 'error',
            message: readableError(error, 'Stuff Stash could not load settings.')
          });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [settingsQuery]);

  async function refreshSettings(): Promise<void> {
    setIsRefreshing(true);

    try {
      const settings = await settingsQuery.execute();
      setScreenState({ status: 'ready', settings });
    } catch (error) {
      setScreenState({
        status: 'error',
        message: readableError(error, 'Stuff Stash could not refresh settings.')
      });
    } finally {
      setIsRefreshing(false);
    }
  }

  return (
    <SafeAreaView style={styles.shell} edges={['top', 'left', 'right']}>
      {screenState.status === 'loading' ? <LoadingState /> : null}
      {screenState.status === 'error' ? <ErrorState message={screenState.message} /> : null}
      {screenState.status === 'ready' ? (
        <SettingsContent
          isRefreshing={isRefreshing}
          settings={screenState.settings}
          onRefresh={refreshSettings}
        />
      ) : null}
    </SafeAreaView>
  );
}

function LoadingState() {
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={colors.accent} />
      <Text style={styles.stateText}>Loading settings</Text>
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

function SettingsContent({
  isRefreshing,
  settings,
  onRefresh
}: {
  readonly isRefreshing: boolean;
  readonly settings: SettingsViewModel;
  readonly onRefresh: () => void;
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
      <View style={styles.brandRow}>
        <BrandMark showWordmark />
      </View>
      <Text style={styles.title}>Settings</Text>

      <View style={styles.panel}>
        <Text style={styles.sectionTitle}>Current user</Text>
        <View style={styles.userRow}>
          <View style={styles.userAvatar}>
            <Text style={styles.userInitial}>{settings.currentUserPrimaryLabel.charAt(0)}</Text>
          </View>
          <View style={styles.userText}>
            <Text style={styles.userPrimary}>{settings.currentUserPrimaryLabel}</Text>
            <Text style={styles.userSecondary}>{settings.currentUserSecondaryLabel}</Text>
          </View>
        </View>
      </View>

      <View style={styles.panel}>
        <Text style={styles.sectionTitle}>About</Text>
        {settings.aboutRows.map((row) => (
          <SettingsRow key={row.label} label={row.label} value={row.value} />
        ))}
      </View>

      <View style={styles.panel}>
        <Text style={styles.sectionTitle}>Developer</Text>
        {settings.developerRows.map((row) => (
          <SettingsRow key={row.label} label={row.label} value={row.value} />
        ))}
      </View>
    </ScrollView>
  );
}

function SettingsRow({
  label,
  value
}: {
  readonly label: string;
  readonly value: string;
}) {
  return (
    <View style={styles.settingsRow}>
      <Text style={styles.settingsLabel}>{label}</Text>
      <Text style={styles.settingsValue}>{value}</Text>
    </View>
  );
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

const styles = StyleSheet.create({
  shell: {
    flex: 1,
    backgroundColor: colors.background
  },
  content: {
    padding: spacing.lg,
    paddingBottom: spacing.xl
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
  },
  title: {
    color: colors.text,
    fontSize: 30,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 36,
    marginBottom: spacing.md
  },
  brandRow: {
    marginBottom: spacing.md
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
  userRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.md,
    minHeight: 58
  },
  userAvatar: {
    alignItems: 'center',
    backgroundColor: colors.brandCharcoal,
    borderRadius: 22,
    height: 44,
    justifyContent: 'center',
    width: 44
  },
  userInitial: {
    color: colors.onAction,
    fontSize: 18,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  userText: {
    flex: 1,
    minWidth: 0
  },
  userPrimary: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '900',
    letterSpacing: 0
  },
  userSecondary: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '700',
    letterSpacing: 0,
    marginTop: 2
  },
  settingsRow: {
    borderBottomColor: colors.border,
    borderBottomWidth: 1,
    gap: spacing.xs,
    paddingVertical: spacing.sm
  },
  settingsLabel: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  settingsValue: {
    color: colors.text,
    fontSize: 16,
    fontWeight: '800',
    letterSpacing: 0,
    lineHeight: 22
  },
});
