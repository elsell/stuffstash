import type { ArchiveScope } from '../../application/archives/InventoryArchive';
import { t } from '../../presentation/localization';
import type { ExportInventoryCommand } from '../../application/exports/InventoryExport';
import { InventoryExportAction } from './InventoryExportAction';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { SettingsRefreshNotice } from './SettingsRefreshNotice';
import { useEffect, useRef } from 'react';
import { AccessibilityInfo, findNodeHandle, ScrollView, Text, View } from 'react-native';
import { AudioLines, Bell, Braces, Share2, Tags } from 'lucide-react-native';
import type { SettingsQuery } from '../../application/settings/SettingsQuery';
import { SettingsLoadingRow, SettingsNavigationRow, SettingsSection, SettingsSeparator, useSettingsListStyles } from './SettingsList';
import { useSettingsModel } from './SettingsScreenState';

type ScopedDestination = 'printers' | 'sharing' | 'tags' | 'fields' | 'asset-types' | 'voice' | 'notifications';

export function InventorySettingsScreen({ onNavigate, settingsQuery, exportCommand, onArchive }: { readonly onArchive?: (scope: ArchiveScope) => void; readonly exportCommand?: ExportInventoryCommand; readonly onNavigate: (destination: ScopedDestination) => void; readonly settingsQuery: SettingsQuery }) {
  const model = useSettingsModel(settingsQuery);
  return <ScopeScreen model={model} onNavigate={onNavigate} scope="inventory" exportCommand={exportCommand} onArchive={onArchive} />;
}

export function HouseholdSettingsScreen({ onNavigate, settingsQuery }: { readonly onNavigate: (destination: ScopedDestination) => void; readonly settingsQuery: SettingsQuery }) {
  const model = useSettingsModel(settingsQuery);
  return <ScopeScreen model={model} onNavigate={onNavigate} scope="tenant" />;
}

function ScopeScreen({ model, onNavigate, scope, exportCommand, onArchive }: { readonly onArchive?: (scope: ArchiveScope) => void; readonly exportCommand?: ExportInventoryCommand; readonly model: ReturnType<typeof useSettingsModel>; readonly onNavigate: (destination: ScopedDestination) => void; readonly scope: 'tenant' | 'inventory' }) {
  const { palette, styles } = useSettingsListStyles();
  if (model.state.status === 'loading') return <View style={[styles.shell, styles.errorContainer]}><SettingsLoadingRow label={scope === 'tenant' ? t('mobile.ScopedSettingsScreens.loadingHouseholdSettings') : t('mobile.ScopedSettingsScreens.loadingInventorySettings')} /></View>;
  if (model.state.status === 'error') return <ScrollView style={styles.shell} contentContainerStyle={styles.errorContainer}><Text accessibilityRole="header" style={styles.errorTitle}>{t('mobile.ScopedSettingsScreens.couldNotLoadSettings')}</Text><Text style={styles.errorMessage}>{model.state.message}</Text><NativeCommandButton label={t('mobile.ScopedSettingsScreens.retry')} onPress={() => void model.load()} /></ScrollView>;
  const settings = model.state.settings;
  const name = scope === 'tenant' ? settings.selectedTenant.name : settings.selectedInventory.name;
  const tenantCanConfigure = settings.selectedTenant.permissions.includes('configure');
  const rows: Array<{ id: ScopedDestination; label: string; context?: string }> = scope === 'tenant'
    ? tenantCanConfigure ? [
        { id: 'fields', label: t('mobile.ScopedSettingsScreens.customFields') },
        { id: 'asset-types', label: t('mobile.ScopedSettingsScreens.assetTypes') },
        { id: 'voice', label: t('mobile.ScopedSettingsScreens.voiceSetup') }
      ] : []
    : [
        ...(settings.selectedInventory.permissions.includes('share') ? [{ id: 'sharing' as const, label: t('mobile.ScopedSettingsScreens.sharing') }] : []),
        { id: 'printers', label: t('printing.mobile.title') },
        { id: 'notifications', label: t('mobile.ScopedSettingsScreens.notifications'), context: t('mobile.ScopedSettingsScreens.yourReminders') },
        { id: 'tags', label: t('mobile.ScopedSettingsScreens.tags') },
        { id: 'fields', label: t('mobile.ScopedSettingsScreens.customFields') },
        { id: 'asset-types', label: t('mobile.ScopedSettingsScreens.assetTypes') }
      ];
  if (scope === 'tenant' && !tenantCanConfigure) return <DeniedSettingsState message={t('mobile.ScopedSettingsScreens.youDonTHavePermissionToManageSettingsShared')} />;
  return <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
    <SettingsRefreshNotice visible={model.hasRefreshError} onRetry={model.load} />
    <View style={styles.detailHeader}><Text accessibilityRole="header" style={styles.detailTitle}>{name}</Text><Text style={styles.detailSubtitle}>{scope === 'tenant' ? t('mobile.ScopedSettingsScreens.householdSettings') : t('mobile.ScopedSettingsScreens.inventoryIn', { name: String(settings.selectedTenant.name) })}</Text></View>
    <SettingsSection>{rows.map((row, index) => <View key={row.id}>{index ? <SettingsSeparator hasLeadingIcon /> : null}<SettingsNavigationRow accessibilityLabel={t('mobile.ScopedSettingsScreens.openFor', { label: String(row.label), name: String(name) })} context={row.context} icon={scopeIcon(row.id, palette.action)} label={row.label} onPress={() => onNavigate(row.id)} /></View>)}</SettingsSection>
    {scope === 'inventory' && onArchive ? <SettingsSection><SettingsNavigationRow accessibilityLabel={t('archive.export')} label={t('archive.export')} onPress={() => onArchive({ tenantId: settings.selectedTenant.id, inventoryId: settings.selectedInventory.id })} /></SettingsSection> : null}
    {scope === 'inventory' && exportCommand ? <InventoryExportAction key={`${settings.selectedTenant.id}:${settings.selectedInventory.id}`} command={exportCommand} scope={{ tenantId: settings.selectedTenant.id, inventoryId: settings.selectedInventory.id }} /> : null}
  </ScrollView>;
}

export function DeniedSettingsState({ message }: { readonly message: string }) {
  const { styles } = useSettingsListStyles();
  const headingRef = useRef<Text>(null);
  useEffect(() => {
    const target = findNodeHandle(headingRef.current);
    if (target) AccessibilityInfo.setAccessibilityFocus(target);
    else AccessibilityInfo.announceForAccessibility(t("settings.unavailableReason", { reason: message }));
  }, [message]);
  return <ScrollView accessibilityLiveRegion="assertive" style={styles.shell} contentContainerStyle={styles.errorContainer}><Text accessibilityRole="header" ref={headingRef} style={styles.errorTitle}>{t('mobile.ScopedSettingsScreens.settingsUnavailable')}</Text><Text style={styles.errorMessage}>{message}</Text></ScrollView>;
}

function scopeIcon(id: ScopedDestination, color: string) {
  const props = { color, size: 20, strokeWidth: 2.2 };
  if (id === 'notifications') return <Bell {...props} />;
  if (id === 'sharing') return <Share2 {...props} />;
  if (id === 'tags') return <Tags {...props} />;
  if (id === 'voice') return <AudioLines {...props} />;
  return <Braces {...props} />;
}
