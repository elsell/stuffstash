import { useEffect, useMemo, useState } from 'react';
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
  appearancePreferences,
  type AppearancePreference
} from '../../application/settings/AppearancePreference';
import { SettingsQuery, type SettingsViewModel } from '../../application/settings/SettingsQuery';
import { BrandMark } from '../components/BrandMark';
import { useAppearance, useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';

type SettingsScreenProps = {
  readonly settingsQuery: SettingsQuery;
  readonly onOpenProviderProfiles: () => void;
  readonly onResetConnection: () => Promise<void>;
};

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly settings: SettingsViewModel }
  | { readonly status: 'error'; readonly message: string };

export function SettingsScreen({
  settingsQuery,
  onOpenProviderProfiles,
  onResetConnection
}: SettingsScreenProps) {
  const palette = useAppearancePalette();
  const styles = useMemo(() => createStyles(palette), [palette]);
  const [screenState, setScreenState] = useState<ScreenState>({ status: 'loading' });
  const [isRefreshing, setIsRefreshing] = useState(false);

  useEffect(() => {
    let isCurrent = true;
    settingsQuery.execute().then((settings) => {
      if (isCurrent) {
        setScreenState({ status: 'ready', settings });
      }
    }).catch((error: unknown) => {
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
      setScreenState({ status: 'ready', settings: await settingsQuery.execute() });
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
          onOpenProviderProfiles={onOpenProviderProfiles}
          onRefresh={refreshSettings}
          onResetConnection={onResetConnection}
        />
      ) : null}
    </SafeAreaView>
  );
}

function LoadingState() {
  const palette = useAppearancePalette();
  const styles = useMemo(() => createStyles(palette), [palette]);
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={palette.accent} />
      <Text style={styles.stateText}>Loading settings</Text>
    </View>
  );
}

function ErrorState({ message }: { readonly message: string }) {
  const styles = useThemedStyles();
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
  onOpenProviderProfiles,
  onRefresh,
  onResetConnection
}: {
  readonly isRefreshing: boolean;
  readonly settings: SettingsViewModel;
  readonly onOpenProviderProfiles: () => void;
  readonly onRefresh: () => void;
  readonly onResetConnection: () => Promise<void>;
}) {
  const palette = useAppearancePalette();
  const styles = useMemo(() => createStyles(palette), [palette]);

  function confirmResetConnection(): void {
    Alert.alert(
      'Change instance',
      'This clears the saved Stuff Stash instance on this device and returns to onboarding.',
      [
        { text: 'Cancel', style: 'cancel' },
        { text: 'Continue', style: 'destructive', onPress: () => void onResetConnection() }
      ]
    );
  }

  return (
    <ScrollView
      contentContainerStyle={styles.content}
      refreshControl={(
        <RefreshControl refreshing={isRefreshing} tintColor={palette.action} onRefresh={onRefresh} />
      )}
    >
      <View style={styles.brandRow}>
        <BrandMark showWordmark />
      </View>
      <Text style={styles.title}>Settings</Text>

      <View style={styles.panel}>
        <Text style={styles.sectionTitle}>Appearance</Text>
        <Text style={styles.sectionDescription}>Choose how Stuff Stash looks on this device.</Text>
        <AppearancePreferenceControl />
      </View>

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
        <SettingsButton label="Voice provider profiles" onPress={onOpenProviderProfiles} />
        {settings.developerRows.map((row) => (
          <SettingsRow key={row.label} label={row.label} value={row.value} />
        ))}
        <SettingsButton label="Change instance" onPress={confirmResetConnection} />
      </View>
    </ScrollView>
  );
}

function AppearancePreferenceControl() {
  const { preference, setPreference } = useAppearance();
  const styles = useThemedStyles();

  async function select(nextPreference: AppearancePreference): Promise<void> {
    if (nextPreference === preference) {
      return;
    }
    try {
      await setPreference(nextPreference);
    } catch {
      Alert.alert('Appearance not saved', 'Stuff Stash could not save the appearance setting.');
    }
  }

  return (
    <View accessibilityRole="radiogroup" style={styles.appearanceControl}>
      {appearancePreferences.map((option) => {
        const selected = option === preference;
        return (
          <Pressable
            key={option}
            accessibilityLabel={`${appearanceLabel(option)} appearance`}
            accessibilityRole="radio"
            accessibilityState={{ checked: selected }}
            onPress={() => void select(option)}
            style={({ pressed }) => [
              styles.appearanceOption,
              selected && styles.appearanceOptionSelected,
              pressed && styles.appearanceOptionPressed
            ]}
          >
            <Text style={[styles.appearanceOptionText, selected && styles.appearanceOptionTextSelected]}>
              {appearanceLabel(option)}
            </Text>
          </Pressable>
        );
      })}
    </View>
  );
}

function SettingsButton({ label, onPress }: { readonly label: string; readonly onPress: () => void }) {
  const styles = useThemedStyles();
  return (
    <Pressable accessibilityRole="button" onPress={onPress} style={styles.secondaryButton}>
      <Text style={styles.secondaryButtonText}>{label}</Text>
    </Pressable>
  );
}

function SettingsRow({ label, value }: { readonly label: string; readonly value: string }) {
  const styles = useThemedStyles();
  return (
    <View style={styles.settingsRow}>
      <Text style={styles.settingsLabel}>{label}</Text>
      <Text style={styles.settingsValue}>{value}</Text>
    </View>
  );
}

function useThemedStyles() {
  const palette = useAppearancePalette();
  return useMemo(() => createStyles(palette), [palette]);
}

function appearanceLabel(preference: AppearancePreference): string {
  return preference.charAt(0).toUpperCase() + preference.slice(1);
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
    shell: { flex: 1, backgroundColor: colors.background },
    content: { padding: spacing.lg, paddingBottom: spacing.xl },
    centerState: { alignItems: 'center', flex: 1, justifyContent: 'center', padding: spacing.lg },
    stateText: { color: colors.textMuted, fontSize: 16, lineHeight: 23, marginTop: spacing.md, textAlign: 'center' },
    errorTitle: { color: colors.text, fontSize: 24, fontWeight: '800' },
    title: { color: colors.text, fontSize: 30, fontWeight: '900', lineHeight: 36, marginBottom: spacing.md },
    brandRow: { marginBottom: spacing.md },
    panel: { backgroundColor: colors.surface, borderColor: colors.border, borderRadius: radius.md, borderWidth: 1, marginBottom: spacing.md, padding: spacing.md },
    sectionTitle: { color: colors.text, fontSize: 17, fontWeight: '800', marginBottom: spacing.xs },
    sectionDescription: { color: colors.textMuted, fontSize: 14, lineHeight: 20, marginBottom: spacing.md },
    appearanceControl: { backgroundColor: colors.surfaceMuted, borderRadius: radius.md, flexDirection: 'row', padding: 3 },
    appearanceOption: { alignItems: 'center', borderRadius: radius.sm, flex: 1, justifyContent: 'center', minHeight: 44, paddingHorizontal: spacing.sm },
    appearanceOptionSelected: { backgroundColor: colors.surface, elevation: 1, shadowColor: colors.brandCharcoalDeep, shadowOffset: { width: 0, height: 1 }, shadowOpacity: 0.12, shadowRadius: 2 },
    appearanceOptionPressed: { backgroundColor: colors.selected },
    appearanceOptionText: { color: colors.textMuted, fontSize: 15, fontWeight: '700' },
    appearanceOptionTextSelected: { color: colors.text, fontWeight: '800' },
    userRow: { alignItems: 'center', flexDirection: 'row', gap: spacing.md, minHeight: 58 },
    userAvatar: { alignItems: 'center', backgroundColor: colors.brandCharcoal, borderRadius: 22, height: 44, justifyContent: 'center', width: 44 },
    userInitial: { color: colors.onAction, fontSize: 18, fontWeight: '900', textTransform: 'uppercase' },
    userText: { flex: 1, minWidth: 0 },
    userPrimary: { color: colors.text, fontSize: 17, fontWeight: '800' },
    userSecondary: { color: colors.textMuted, fontSize: 13, fontWeight: '700', marginTop: 2 },
    settingsRow: { borderBottomColor: colors.border, borderBottomWidth: 1, gap: spacing.xs, paddingVertical: spacing.sm },
    settingsLabel: { color: colors.textMuted, fontSize: 12, fontWeight: '800', textTransform: 'uppercase' },
    settingsValue: { color: colors.text, fontSize: 16, fontWeight: '700', lineHeight: 22 },
    secondaryButton: { alignItems: 'center', borderColor: colors.controlBorder, borderRadius: radius.md, borderWidth: 1, justifyContent: 'center', marginTop: spacing.md, minHeight: 44, paddingHorizontal: spacing.md },
    secondaryButtonText: { color: colors.action, fontSize: 15, fontWeight: '800' }
  });
}
