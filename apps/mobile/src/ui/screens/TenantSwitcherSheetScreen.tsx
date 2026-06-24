import { useEffect, useState } from 'react';
import { router } from 'expo-router';
import {
  ActivityIndicator,
  Pressable,
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
import { colors, radius, spacing } from '../theme/tokens';

type TenantSwitcherSheetScreenProps = {
  readonly dashboardQuery: HomeDashboardQuery;
  readonly selectInventoryCommand: SelectInventoryCommand;
};

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly dashboard: HomeDashboardViewModel }
  | { readonly status: 'error'; readonly message: string };

export function TenantSwitcherSheetScreen({
  dashboardQuery,
  selectInventoryCommand
}: TenantSwitcherSheetScreenProps) {
  const [screenState, setScreenState] = useState<ScreenState>({ status: 'loading' });

  useEffect(() => {
    let isCurrent = true;

    dashboardQuery
      .execute()
      .then((dashboard) => {
        if (isCurrent) {
          setScreenState({ status: 'ready', dashboard });
        }
      })
      .catch((error: unknown) => {
        if (isCurrent) {
          setScreenState({
            status: 'error',
            message: readableError(error, 'Could not load tenants.')
          });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [dashboardQuery]);

  async function selectInventory(inventoryId: string): Promise<void> {
    await selectInventoryCommand.execute(inventoryId);
    router.back();
  }

  return (
    <SafeAreaView style={styles.sheet} edges={['left', 'right', 'bottom']}>
      {screenState.status === 'loading' ? <LoadingState /> : null}
      {screenState.status === 'error' ? <ErrorState message={screenState.message} /> : null}
      {screenState.status === 'ready' ? (
        <TenantSwitcher
          dashboard={screenState.dashboard}
          onSelectInventory={selectInventory}
        />
      ) : null}
    </SafeAreaView>
  );
}

function TenantSwitcher({
  dashboard,
  onSelectInventory
}: {
  readonly dashboard: HomeDashboardViewModel;
  readonly onSelectInventory: (inventoryId: string) => Promise<void>;
}) {
  const currentTenant = dashboard.tenants.find((tenant) => tenant.name === dashboard.tenantName);
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
            kind="tenant"
            label={selectedTenant?.name ?? dashboard.tenantName}
            textStyle={styles.sheetTitle}
          />
        </View>
        <Pressable
          accessibilityRole="button"
          onPress={() => {
            if (mode === 'tenants') {
              setMode('inventories');
              return;
            }

            setMode('tenants');
          }}
          style={styles.switchButton}
        >
          <Text style={styles.switchButtonText}>
            {mode === 'tenants' ? 'Back' : 'Switch tenant'}
          </Text>
        </Pressable>
      </View>

      {mode === 'inventories' ? (
        <>
          <Text style={styles.sectionLabel}>Inventories</Text>

          {selectedTenantInventories.map((inventory, index) => {
            const isSelected = inventory.id === dashboard.inventoryId;

            return (
              <Pressable
                accessibilityRole="button"
                accessibilityState={{ selected: isSelected }}
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
          <Text style={styles.sectionLabel}>Tenants</Text>

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
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={colors.accent} />
      <Text style={styles.stateText}>Loading tenants</Text>
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

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

const styles = StyleSheet.create({
  sheet: {
    backgroundColor: colors.surface,
    padding: spacing.md
  },
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
    alignItems: 'center',
    borderBottomColor: colors.border,
    borderBottomWidth: 1,
    flexDirection: 'row',
    gap: spacing.md,
    justifyContent: 'space-between',
    paddingBottom: spacing.md
  },
  contextText: {
    flex: 1,
    minWidth: 0
  },
  sheetTitle: {
    color: colors.text,
    fontSize: 26,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 31
  },
  switchButton: {
    minHeight: 36,
    justifyContent: 'center',
    paddingHorizontal: spacing.xs
  },
  switchButtonText: {
    color: colors.action,
    fontSize: 14,
    fontWeight: '800',
    letterSpacing: 0
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
