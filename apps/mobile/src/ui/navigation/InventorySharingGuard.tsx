import { NativeCommandButton } from '../components/NativeCommandButton';
import type { ReactNode } from 'react';
import { ActivityIndicator, ScrollView, Text, View } from 'react-native';
import type { InventorySharingScope } from '../../application/sharing/InventorySharing';
import type { SettingsQuery } from '../../application/settings/SettingsQuery';
import { useSettingsListStyles } from '../screens/SettingsList';
import { useSettingsModel } from '../screens/SettingsScreenState';
import { decideInventorySharingAccess } from './InventorySharingAccess';

export function InventorySharingGuard({
  children,
  settingsQuery
}: {
  readonly children: (scope: InventorySharingScope) => ReactNode;
  readonly settingsQuery: SettingsQuery;
}) {
  const { load, state } = useSettingsModel(settingsQuery);
  const { palette, styles } = useSettingsListStyles();
  const decision = decideInventorySharingAccess(state);

  if (decision.status === 'loading') {
    return (
      <View style={[styles.shell, styles.errorContainer]}>
        <ActivityIndicator color={palette.action} />
        <Text style={styles.errorMessage}>Checking Sharing access</Text>
      </View>
    );
  }
  if (decision.status === 'allowed') return children(decision.scope);

  const unavailable = decision.status === 'unavailable';
  return (
    <ScrollView contentContainerStyle={styles.errorContainer} style={styles.shell}>
      <Text accessibilityRole="header" style={styles.errorTitle}>
        {unavailable ? 'Sharing unavailable' : 'Could not verify Sharing access'}
      </Text>
      <Text style={styles.errorMessage}>
        {unavailable
          ? `You don’t have permission to manage invitations for ${decision.inventoryName}.`
          : decision.message}
      </Text>
      <NativeCommandButton label={unavailable ? 'Check Again' : 'Retry'} onPress={() => void load()} />
    </ScrollView>
  );
}
