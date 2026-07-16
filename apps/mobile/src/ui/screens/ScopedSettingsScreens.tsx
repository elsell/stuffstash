import { useEffect, useRef } from 'react';
import { AccessibilityInfo, ActivityIndicator, findNodeHandle, Pressable, ScrollView, Text, View } from 'react-native';
import { AudioLines, Braces, Share2, Tags } from 'lucide-react-native';
import type { SettingsQuery } from '../../application/settings/SettingsQuery';
import { SettingsNavigationRow, SettingsSection, SettingsSeparator, useSettingsListStyles } from './SettingsList';
import { useSettingsModel } from './SettingsScreenState';

type ScopedDestination = 'sharing' | 'tags' | 'fields' | 'asset-types' | 'voice';

export function InventorySettingsScreen({ onNavigate, settingsQuery }: { readonly onNavigate: (destination: ScopedDestination) => void; readonly settingsQuery: SettingsQuery }) {
  const model = useSettingsModel(settingsQuery);
  return <ScopeScreen model={model} onNavigate={onNavigate} scope="inventory" />;
}

export function HouseholdSettingsScreen({ onNavigate, settingsQuery }: { readonly onNavigate: (destination: ScopedDestination) => void; readonly settingsQuery: SettingsQuery }) {
  const model = useSettingsModel(settingsQuery);
  return <ScopeScreen model={model} onNavigate={onNavigate} scope="tenant" />;
}

function ScopeScreen({ model, onNavigate, scope }: { readonly model: ReturnType<typeof useSettingsModel>; readonly onNavigate: (destination: ScopedDestination) => void; readonly scope: 'tenant' | 'inventory' }) {
  const { palette, styles } = useSettingsListStyles();
  if (model.state.status === 'loading') return <View style={[styles.shell, styles.errorContainer]}><ActivityIndicator color={palette.action} /></View>;
  if (model.state.status === 'error') return <View style={[styles.shell, styles.errorContainer]}><Text accessibilityRole="header" style={styles.errorTitle}>Could not load settings</Text><Text style={styles.errorMessage}>{model.state.message}</Text><Pressable accessibilityRole="button" onPress={() => void model.load()} style={styles.retryButton}><Text style={styles.retryText}>Retry</Text></Pressable></View>;
  const settings = model.state.settings;
  const name = scope === 'tenant' ? settings.selectedTenant.name : settings.selectedInventory.name;
  const tenantCanConfigure = settings.selectedTenant.permissions.includes('configure');
  const rows: Array<{ id: ScopedDestination; label: string; context: string }> = scope === 'tenant'
    ? tenantCanConfigure ? [
        { id: 'fields', label: 'Custom fields', context: `Shared by ${name}` },
        { id: 'asset-types', label: 'Asset types', context: `Shared by ${name}` },
        { id: 'voice', label: 'Voice setup', context: `Shared by ${name}` }
      ] : []
    : [
        ...(settings.selectedInventory.permissions.includes('share') ? [{ id: 'sharing' as const, label: 'Sharing', context: name }] : []),
        { id: 'tags', label: 'Tags', context: name },
        { id: 'fields', label: 'Custom fields', context: name },
        { id: 'asset-types', label: 'Asset types', context: name }
      ];
  if (scope === 'tenant' && !tenantCanConfigure) return <DeniedSettingsState message="You don’t have permission to manage settings shared by this household." />;
  return <ScrollView contentContainerStyle={styles.content} style={styles.shell}>
    <View style={styles.detailHeader}><Text accessibilityRole="header" style={styles.detailTitle}>{name}</Text><Text style={styles.detailSubtitle}>{scope === 'tenant' ? 'Household settings' : `Inventory in ${settings.selectedTenant.name}`}</Text></View>
    <SettingsSection>{rows.map((row, index) => <View key={row.id}>{index ? <SettingsSeparator /> : null}<SettingsNavigationRow accessibilityLabel={`Open ${row.label} for ${name}`} context={row.context} icon={scopeIcon(row.id, palette.action)} label={row.label} onPress={() => onNavigate(row.id)} /></View>)}</SettingsSection>
  </ScrollView>;
}

export function DeniedSettingsState({ message }: { readonly message: string }) {
  const { styles } = useSettingsListStyles();
  const headingRef = useRef<Text>(null);
  useEffect(() => {
    const target = findNodeHandle(headingRef.current);
    if (target) AccessibilityInfo.setAccessibilityFocus(target);
    else AccessibilityInfo.announceForAccessibility(`Settings unavailable. ${message}`);
  }, [message]);
  return <View accessibilityLiveRegion="assertive" style={[styles.shell, styles.errorContainer]}><Text accessibilityRole="header" ref={headingRef} style={styles.errorTitle}>Settings unavailable</Text><Text style={styles.errorMessage}>{message}</Text></View>;
}

function scopeIcon(id: ScopedDestination, color: string) {
  const props = { color, size: 20, strokeWidth: 2.2 };
  if (id === 'sharing') return <Share2 {...props} />;
  if (id === 'tags') return <Tags {...props} />;
  if (id === 'voice') return <AudioLines {...props} />;
  return <Braces {...props} />;
}
