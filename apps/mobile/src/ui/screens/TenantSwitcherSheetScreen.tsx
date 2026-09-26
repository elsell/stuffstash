import { NativeActionRow } from '../components/NativeActionRow';
import type { CreateWorkspace, CreatedHousehold, CreatedInventory } from '../../application/inventories/CreateWorkspace';
import { WorkspaceCreationForm, type WorkspaceCreationTask } from './WorkspaceCreationForm';
import { returnToPreviousOrHome } from '../navigation/returnToPreviousOrHome';
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
  readonly createWorkspace?: CreateWorkspace;
  readonly dashboardQuery: HomeDashboardQuery;
  readonly selectInventoryCommand: SelectInventoryCommand;
};

export function TenantSwitcherSheetScreen(props: TenantSwitcherSheetScreenProps) {
  const scopeId = useMobileServerStateScopeId();
  return <TenantSwitcherVisit key={scopeId} {...props} />;
}

function TenantSwitcherVisit({
  dashboardQuery,
  selectInventoryCommand,
  createWorkspace
}: TenantSwitcherSheetScreenProps) {
  const styles = useStyles();
  const [creation, setCreation] = useState<WorkspaceCreationTask>();
  const [creating, setCreating] = useState(false);
  const [createdHouseholds, setCreatedHouseholds] = useState<readonly CreatedHousehold[]>([]);
  const [createdInventories, setCreatedInventories] = useState<readonly CreatedInventory[]>([]);
  const [preferredTenantId, setPreferredTenantId] = useState<string>();
  const [selecting, setSelecting] = useState(false);
  const [selectionError, setSelectionError] = useState('');
  const pending = useRef<AbortController | undefined>(undefined);
  const focused = useRef(true);
  const scopeId = useMobileServerStateScopeId();
  const [visit, setVisit] = useState<{ active: boolean } | null>(null);
  useFocusEffect(useCallback(() => {
    const owner = { active: true };
    focused.current = true; setVisit(owner); setSelecting(Boolean(pending.current));
    return () => { owner.active = false; focused.current = false; pending.current?.abort(); setCreation(undefined); setCreating(false); };
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
      if (focused.current && !request.signal.aborted) returnToPreviousOrHome(router);
    } catch {
      if (focused.current && !request.signal.aborted) setSelectionError('Could not switch inventories. Try again.');
    } finally {
      if (pending.current === request) pending.current = undefined;
      if (focused.current) setSelecting(false);
    }
  }

  const data = dashboard.data;
  const displayed = data ? { ...data,
    tenants: [...data.tenants, ...createdHouseholds.filter(item => !data.tenants.some(tenant => tenant.id === item.id))],
    inventories: [...data.inventories, ...createdInventories.filter(item => !data.inventories.some(inventory => inventory.id === item.id)).map(item => ({ ...item,
      tenantName: createdHouseholds.find(tenant => tenant.id === item.tenantId)?.name ?? '', roleLabel: 'Owner', updatedAtLabel: 'Just created' }))]
  } : undefined;
  const actionOptions = useNativeHeaderActionOptions([{ kind: 'close', label: creation ? 'Cancel creation' : 'Close inventory switcher', disabled: creating, onPress: () => {
    if (creating) return;
    if (creation) { setCreation(undefined); return; }
    if (!visit?.active) return;
    visit.active = false;
    pending.current?.abort(); returnToPreviousOrHome(router);
  } }]);
  const headerOptions = useMemo(() => ({ title: creation ? creation.kind === 'household' ? 'New household' : 'New inventory' : 'Inventories', ...actionOptions }), [actionOptions, creation]);

  return (
    <SafeAreaView style={styles.sheet} edges={['left', 'right', 'bottom']}>
      <Stack.Screen options={headerOptions} />
      {creation && createWorkspace ? <WorkspaceCreationForm task={creation} command={createWorkspace}
        onBusy={busy => { if (visit?.active) setCreating(busy); }} onCancel={() => setCreation(undefined)} onCreated={result => {
          if (!visit?.active) return;
          if (result.kind === 'household') {
            setCreatedHouseholds(current => [...current, result.value]); setPreferredTenantId(result.value.id);
          } else {
            setCreatedInventories(current => [...current, result.value]); setPreferredTenantId(result.value.tenantId);
          }
          setCreating(false); setCreation(undefined);
        }} /> : <ScrollView contentInsetAdjustmentBehavior="automatic" contentContainerStyle={styles.content}>
      {dashboard.isPending && !dashboard.data ? <LoadingState /> : null}
      {dashboard.isError && !dashboard.data ? (
        <ErrorState onRetry={() => { void dashboard.refetch(); }} />
      ) : null}
      {selectionError ? <Text accessibilityRole="alert" style={styles.errorMessage}>{selectionError}</Text> : null}
      {displayed ? (
        <TenantSwitcher
          dashboard={displayed}
          preferredTenantId={preferredTenantId}
          onCreate={createWorkspace ? task => { if (visit?.active && !pending.current) setCreation(task); } : undefined}
          selecting={selecting}
          onSelectInventory={selectInventory}
        />
      ) : null}
      </ScrollView>}
    </SafeAreaView>
  );
}

function TenantSwitcher({
  dashboard,
  selecting,
  onSelectInventory,
  preferredTenantId,
  onCreate
}: {
  readonly preferredTenantId?: string;
  readonly onCreate?: (task: WorkspaceCreationTask) => void;
  readonly dashboard: HomeDashboardViewModel;
  readonly selecting: boolean;
  readonly onSelectInventory: (inventoryId: string) => Promise<void>;
}) {
  const styles = useStyles();
  const currentTenant = dashboard.tenants.find((tenant) => tenant.id === dashboard.tenantId);
  const [selectedTenantId, setSelectedTenantId] = useState(preferredTenantId ?? currentTenant?.id ?? dashboard.tenants[0]?.id);
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
        <View style={styles.switchAction}><NativeCommandButton prominence="standard" label={mode === 'tenants' ? 'Back' : 'Switch household'}
          disabled={selecting} onPress={() => setMode(mode === 'tenants' ? 'inventories' : 'tenants')} /></View>
      </View>

      {mode === 'inventories' ? (
        <>
          <Text style={styles.sectionLabel}>Inventories</Text>
          {onCreate && selectedTenant?.canCreateInventory ? <NativeActionRow label="New inventory" disabled={selecting}
            onPress={() => onCreate({ kind: 'inventory', household: selectedTenant })} /> : null}
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
          {onCreate ? <NativeActionRow label="New household" disabled={selecting} onPress={() => onCreate({ kind: 'household' })} /> : null}

          {dashboard.tenants.map((tenant, index) => {
            const isSelected = tenant.id === selectedTenant?.id;
            const inventoryCount = dashboard.inventories.filter((inventory) => inventory.tenantId === tenant.id).length;

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
                    {`${inventoryCount} ${inventoryCount === 1 ? 'inventory' : 'inventories'}`}
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
    alignItems: 'center',
    flexDirection: 'row',
    borderBottomColor: colors.border,
    borderBottomWidth: 1,
    gap: spacing.sm,
    paddingBottom: spacing.md
  },
  switchAction: { width: '44%', flexShrink: 0, alignItems: 'flex-end' },
  contextText: {
    flex: 1,
    minWidth: 0
  },
  sheetTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '600',
    letterSpacing: 0,
    lineHeight: 22
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
