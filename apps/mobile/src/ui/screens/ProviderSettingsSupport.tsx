import { isAccessFailure } from '../serverState/isAccessFailure';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';
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

export function useProviderProfiles(query: ProviderProfileSettingsQuery) {
  return useMobileInventoryServerQuery({ key: mobileQueryKeys.providerProfiles, query: (signal) => query.listProfiles({ signal }) });
}

export function useProviderProfileModel(query: ProviderProfileSettingsQuery) {
  const profiles = useProviderProfiles(query);
  const state: Exclude<ProviderState, { status: 'ready' }> | { status: 'ready'; viewModel: Pick<ProviderProfileSettingsViewModel, 'profiles'> } = isAccessFailure(profiles.error) ? { status: 'error', message: 'Provider profiles are no longer available.' } : profiles.data
    ? { status: 'ready', viewModel: { profiles: profiles.data } }
    : profiles.isError ? { status: 'error', message: 'Provider profiles could not be loaded.' } : { status: 'loading' };
  return { state, ownerKey: JSON.stringify(profiles.resourceKey), load: async () => { await profiles.reconcile(); }, retry: async () => { await profiles.refetch(); }, hasRefreshError: profiles.isRefetchError };
}

export function useProviderSettings(query: ProviderProfileSettingsQuery) {
  const profiles = useProviderProfiles(query);
  const configuration = useMobileInventoryServerQuery({ key: mobileQueryKeys.voiceConfiguration, query: (signal) => query.getConfiguration({ signal }) });
  const state: ProviderState = isAccessFailure(profiles.error) || isAccessFailure(configuration.error) ? { status: 'error', message: 'Voice settings are no longer available.' } : profiles.data && configuration.data ? { status: 'ready', viewModel: {
    profiles: profiles.data, configuration: configuration.data,
    missingCapabilities: configuration.data.slots.filter((slot) => slot.readiness !== 'ready').map((slot) => slot.capability)
  } } : profiles.isError || configuration.isError ? { status: 'error', message: 'Voice settings could not be loaded.' } : { status: 'loading' };
  const load = async () => { await Promise.all([profiles.reconcile(), configuration.reconcile()]); };
  const retry = async () => { await Promise.all([profiles.refetch(), configuration.refetch()]); };
  return { state, load, retry, hasRefreshError: profiles.isRefetchError || configuration.isRefetchError };
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
