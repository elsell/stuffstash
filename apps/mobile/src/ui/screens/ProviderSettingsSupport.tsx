import { useCallback, useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';
import type {
  ProviderProfileSettingsQuery,
  ProviderProfileSettingsViewModel
} from '../../application/providerProfiles/ProviderProfileSettingsQuery';
import { useSettingsListStyles } from './SettingsList';

export type ProviderState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly viewModel: ProviderProfileSettingsViewModel }
  | { readonly status: 'error'; readonly message: string };

export function useProviderSettings(query: ProviderProfileSettingsQuery) {
  const [state, setState] = useState<ProviderState>({ status: 'loading' });
  const load = useCallback(async () => {
    setState({ status: 'loading' });
    try {
      setState({ status: 'ready', viewModel: await query.execute() });
    } catch (error) {
      setState({ status: 'error', message: readableError(error) });
    }
  }, [query]);
  useEffect(() => { void load(); }, [load]);
  return { load, state };
}

export function ProviderStateView({
  state,
  onRetry
}: {
  readonly state: Exclude<ProviderState, { status: 'ready' }>;
  readonly onRetry: () => Promise<void>;
}) {
  const { palette, styles } = useSettingsListStyles();
  if (state.status === 'loading') {
    return <View style={[styles.shell, styles.errorContainer]}><ActivityIndicator color={palette.action} /></View>;
  }
  return (
    <ScrollView contentContainerStyle={styles.errorContainer} style={styles.shell}>
      <Text accessibilityRole="header" style={styles.errorTitle}>Could not load Voice Setup</Text>
      <Text style={styles.errorMessage}>{state.message}</Text>
      <Pressable accessibilityRole="button" onPress={() => void onRetry()} style={styles.retryButton}>
        <Text style={styles.retryText}>Retry</Text>
      </Pressable>
    </ScrollView>
  );
}

export function readableError(error: unknown): string {
  return error instanceof Error ? error.message : 'The action failed safely.';
}
