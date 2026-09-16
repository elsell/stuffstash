import { NativeCommandButton } from '../components/NativeCommandButton';
import { useCallback, useMemo, useRef, useState } from 'react';
import { router, Stack, useFocusEffect } from 'expo-router';
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import {
  HomeDashboardQuery,
  HomeDashboardViewModel
} from '../../application/home/HomeDashboardQuery';
import { SelectInventoryCommand } from '../../application/home/SelectInventoryCommand';
import { IdentityIcon, IdentityLabel } from '../components/IdentityIcon';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';
import { useMobileServerStateScopeId } from '../navigation/MobileServerStateProvider';

type TenantSwitcherSheetScreenProps = {
  readonly dashboardQuery: HomeDashboardQuery;
  readonly selectInventoryCommand: SelectInventoryCommand;
};

export function TenantSwitcherSheetScreen({
  dashboardQuery,
  selectInventoryCommand
}: TenantSwitcherSheetScreenProps) {
  const styles = useStyles();
  const [selecting, setSelecting] = useState(false);
  const [selectionError, setSelectionError] = useState('');
  const pending = useRef<AbortController | undefined>(undefined);
  const focused = useRef(true);
  const scopeId = useMobileServerStateScopeId();
  const [visit, setVisit] = useState<{ active: boolean } | null>(null);
  useFocusEffect(useCallback(() => {
    const owner = { active: true };
    focused.current = true; setVisit(owner); setSelecting(Boolean(pending.current));
    return () => { owner.active = false; focused.current = false; pending.current?.abort(); };
  }, [scopeId, selectInventoryCommand]));
  const dashboard = useMobileInventoryServerQuery({
    key: mobileQueryKeys.home,
    query: (signal) => dashboardQuery.execute({ signal })
  });

  async function selectInventory(inventoryId: string): Promise<void> {
    if (!visit?.active || pending.current) return;
    const request = new AbortController(); pending.current = request;
    setSelecting(true); setSelectionError('');
    try {
      await selectInventoryCommand.execute(inventoryId, { signal: request.signal });
      if (focused.current && !request.signal.aborted) router.back();
    } catch {
      if (focused.current && !request.signal.aborted) setSelectionError('Could not switch inventories. Try again.');
    } finally {
      if (pending.current === request) pending.current = undefined;
      if (focused.current) setSelecting(false);
    }
  }

  const actionOptions = useNativeHeaderActionOptions([{ kind: 'close', label: 'Close inventory switcher', onPress: () => {
    if (!visit?.active) return;
    visit.active = false;
    pending.current?.abort(); router.back();
  } }]);
  const headerOptions = useMemo(() => ({ title: 'Inventories', ...actionOptions }), [actionOptions]);

  return (
    <SafeAreaView style={styles.sheet} edges={['left', 'right', 'bottom']}>
      <Stack.Screen options={headerOptions} />
      <ScrollView contentInsetAdjustmentBehavior="automatic" contentContainerStyle={styles.content}>
      {dashboard.isPending && !dashboard.data ? <LoadingState /> : null}
      {dashboard.isError && !dashboard.data ? (
        <ErrorState onRetry={() => { void dashboard.refetch(); }} />
      ) : null}
      {selectionError ? <Text accessibilityRole="alert" style={styles.errorMessage}>{selectionError}</Text> : null}
      {dashboard.data ? (
        <TenantSwitcher
          dashboard={dashboard.data}
          selecting={selecting}
          onSelectInventory={selectInventory}
        />
      ) : null}
      </ScrollView>
    </SafeAreaView>
  );
}

function TenantSwitcher({
  dashboard,
  selecting,
  onSelectInventory
}: {
  readonly dashboard: HomeDashboardViewModel;
  readonly selecting: boolean;
  readonly onSelectInventory: (inventoryId: string) => Promise<void>;
}) {
  const styles = useStyles();
  const currentTenant = dashboard.tenants.find((tenant) => tenant.id === dashboard.tenantId);
  const [selectedTenantId, setSelectedTenantId] = useState(currentTenant?.id ?? dashboard.tenants[0]?.id);
  const [mode, setMode] = useState<'inventories' | 'tenants'>('inventories');
  const selectedTenant =
    dashboard.tenants.find((tenant) => tenant.id === selectedTenantId) ??
    dashboard.tenants[0];
  const selectedTenantInventories = selectedTenant
    ? dashboard.inventories.filter((inventory) => inventory.tenantId === selectedTenant.id)
    : [];

  return (
    <View>
      <View style={styles.sheetHeader}>
        <View style={styles.contextText}>
          <IdentityLabel
            iconSize="md"
            numberOfLines={0}
            kind="tenant"
            label={selectedTenant?.name ?? dashboard.tenantName}
            textStyle={styles.sheetTitle}
          />
        </View>
        <NativeCommandButton label={mode === 'tenants' ? 'Back' : 'Switch household'}
          disabled={selecting} onPress={() => setMode(mode === 'tenants' ? 'inventories' : 'tenants')} />
      </View>

      {mode === 'inventories' ? (
        <>
          <Text style={styles.sectionLabel}>Inventories</Text>
          {selectedTenantInventories.length === 0 ? <Text style={styles.stateText}>No inventories are available in this household.</Text> : null}

          {selectedTenantInventories.map((inventory, index) => {
            const isSelected = inventory.id === dashboard.inventoryId;

            return (
              <Pressable
                accessibilityRole="button"
                accessibilityLabel={`Switch to inventory ${inventory.name}`}
                accessibilityState={{ selected: isSelected, disabled: selecting, busy: selecting }}
                disabled={selecting}
                key={inventory.id}
                onPress={() => onSelectInventory(inventory.id)}
                style={[
                  styles.optionRow,
                  index === selectedTenantInventories.length - 1 ? styles.optionRowLast : null
                ]}
              >
                <Text style={styles.optionCheck}>{isSelected ? '✓' : ''}</Text>
                <IdentityIcon kind="inventory" size="md" />
                <View style={styles.optionText}>
                  <Text style={styles.optionName}>{inventory.name}</Text>
                  <Text style={styles.optionMeta}>{inventory.updatedAtLabel}</Text>
                </View>
                <Text style={styles.rolePill}>{inventory.roleLabel}</Text>
              </Pressable>
            );
          })}
        </>
      ) : (
        <>
          <Text style={styles.sectionLabel}>Households</Text>

          {dashboard.tenants.map((tenant, index) => {
            const isSelected = tenant.id === selectedTenant?.id;

            return (
              <Pressable
                accessibilityRole="button"
                accessibilityState={{ selected: isSelected }}
                key={tenant.id}
                onPress={() => {
                  setSelectedTenantId(tenant.id);
                  setMode('inventories');
                }}
                style={[
                  styles.optionRow,
                  index === dashboard.tenants.length - 1 ? styles.optionRowLast : null
                ]}
              >
                <Text style={styles.optionCheck}>{isSelected ? '✓' : ''}</Text>
                <IdentityIcon kind="tenant" size="md" />
                <View style={styles.optionText}>
                  <Text style={styles.optionName}>{tenant.name}</Text>
                  <Text style={styles.optionMeta}>
                    {dashboard.inventories.filter((inventory) => inventory.tenantId === tenant.id).length.toString()} inventories
                  </Text>
                </View>
              </Pressable>
            );
          })}
        </>
      )}
    </View>
  );
}

function LoadingState() {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={palette.accent} />
      <Text style={styles.stateText}>Loading inventories</Text>
    </View>
  );
}

function ErrorState({ onRetry }: { readonly onRetry: () => void }) {
  const styles = useStyles();
  return (
    <View style={styles.centerState}>
      <Text style={styles.errorTitle}>Could not load</Text>
      <Text style={styles.stateText}>Inventories could not be loaded. Try again.</Text>
      <NativeCommandButton label="Retry inventories" onPress={onRetry} />
    </View>
  );
}

function useStyles() {
  return createStyles(useAppearancePalette());
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  sheet: {
    flex: 1,
    backgroundColor: colors.surface,
  },
  content: {
    padding: spacing.md
  },
  errorMessage: { color: colors.danger, fontSize: 17, paddingVertical: spacing.md },
  centerState: {
    alignItems: 'center',
    minHeight: 140,
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
  sheetHeader: {
    alignItems: 'stretch',
    borderBottomColor: colors.border,
    borderBottomWidth: 1,
    gap: spacing.sm,
    paddingBottom: spacing.md
  },
  contextText: {
    minWidth: 0
  },
  sheetTitle: {
    color: colors.text,
    fontSize: 26,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 31
  },
  sectionLabel: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0,
    paddingTop: spacing.md,
    paddingBottom: spacing.xs,
    textTransform: 'uppercase'
  },
  optionRow: {
    alignItems: 'center',
    borderBottomColor: colors.border,
    borderBottomWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 62,
    paddingVertical: spacing.sm
  },
  optionRowLast: {
    borderBottomWidth: 0
  },
  optionCheck: {
    color: colors.action,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0,
    width: 20
  },
  optionText: {
    flex: 1,
    minWidth: 0
  },
  optionName: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0
  },
  optionMeta: {
    color: colors.textMuted,
    fontSize: 12,
    letterSpacing: 0,
    marginTop: 2
  },
  rolePill: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  }
  });
}
