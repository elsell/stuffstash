import { t } from '../../presentation/localization';
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
        <Text style={styles.errorMessage}>{t('mobile.InventorySharingGuard.checkingSharingAccess')}</Text>
      </View>
    );
  }
  if (decision.status === 'allowed') return children(decision.scope);

  const unavailable = decision.status === 'unavailable';
  return (
    <ScrollView contentContainerStyle={styles.errorContainer} style={styles.shell}>
      <Text accessibilityRole="header" style={styles.errorTitle}>
        {unavailable ? t('mobile.InventorySharingGuard.sharingUnavailable') : t('mobile.InventorySharingGuard.couldNotVerifySharingAccess')}
      </Text>
      <Text style={styles.errorMessage}>
        {unavailable
          ? t('mobile.InventorySharingGuard.youDonTHavePermissionToManageInvitationsFor', { inventoryName: String(decision.inventoryName) })
          : decision.message}
      </Text>
      <NativeCommandButton label={unavailable ? t('mobile.InventorySharingGuard.checkAgain') : t('mobile.InventorySharingGuard.retry')} onPress={() => void load()} />
    </ScrollView>
  );
}
