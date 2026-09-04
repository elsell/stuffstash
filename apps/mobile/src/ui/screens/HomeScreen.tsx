import { useState } from 'react';
import { router } from 'expo-router';
import { ChevronDown, Plus, UserCircle } from 'lucide-react-native';
import {
  ActivityIndicator,
  Modal,
  Pressable,
  RefreshControl,
  ScrollView,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { AssetCheckoutCommand } from '../../application/assets/AssetCheckoutCommand';
import {
  HomeDashboardQuery,
  HomeDashboardViewModel
} from '../../application/home/HomeDashboardQuery';
import type { AssetCardViewModel } from '../../application/assets/AssetViewModels';
import { AssetCard } from '../components/AssetCard';
import { AppTextInput, appKeyboardDismissMode } from '../components/AppTextInput';
import { useAppFeedback } from '../feedback/AppFeedback';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { assetDetailHref } from './AssetDetailNavigation';
import { createHomeScreenStyles } from './HomeScreen.styles';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';

type HomeScreenProps = {
  readonly dashboardQuery: HomeDashboardQuery;
  readonly assetCheckoutCommand: AssetCheckoutCommand;
};

export function HomeScreen({ assetCheckoutCommand, dashboardQuery }: HomeScreenProps) {
  const styles = createHomeScreenStyles(useAppearanceAwarePalette());
  const feedback = useAppFeedback();
  const dashboardState = useMobileInventoryServerQuery({
    key: mobileQueryKeys.home,
    query: (signal) => dashboardQuery.execute({ signal })
  });

  async function refreshDashboard(): Promise<void> {
    try {
      await dashboardState.refetch({ throwOnError: true });
    } catch (error) {
      feedback.showNotice({
        tone: 'error',
        title: 'Could not refresh Home',
        message: readableError(error, 'Stuff Stash could not refresh the mobile home screen.')
      });
    }
  }

  return (
    <SafeAreaView style={styles.shell} edges={['top', 'left', 'right']}>
      {dashboardState.isPending && !dashboardState.data ? <LoadingState /> : null}
      {dashboardState.isError && !dashboardState.data ? (
        <ErrorState
          message={readableError(dashboardState.error, 'Stuff Stash could not load the mobile home screen.')}
          onRetry={() => { void dashboardState.refetch(); }}
        />
      ) : null}
      {dashboardState.data ? (
        <Dashboard
          assetCheckoutCommand={assetCheckoutCommand}
          dashboard={dashboardState.data}
          isRefreshing={dashboardState.isRefetching}
          onRefresh={refreshDashboard}
        />
      ) : null}
    </SafeAreaView>
  );
}

function LoadingState() {
  const colors = useAppearanceAwarePalette();
  const styles = createHomeScreenStyles(colors);
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={colors.accent} />
      <Text style={styles.stateText}>Loading Stuff Stash</Text>
    </View>
  );
}

function ErrorState({ message, onRetry }: { readonly message: string; readonly onRetry: () => void }) {
  const styles = createHomeScreenStyles(useAppearanceAwarePalette());
  return (
    <View style={styles.centerState}>
      <Text style={styles.errorTitle}>Could not load</Text>
      <Text style={styles.stateText}>{message}</Text>
      <Pressable
        accessibilityLabel="Retry loading Home"
        accessibilityRole="button"
        onPress={onRetry}
        style={styles.retryButton}
      >
        <Text style={styles.retryButtonText}>Retry</Text>
      </Pressable>
    </View>
  );
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

function Dashboard({
  assetCheckoutCommand,
  dashboard,
  isRefreshing,
  onRefresh
}: {
  readonly assetCheckoutCommand: AssetCheckoutCommand;
  readonly dashboard: HomeDashboardViewModel;
  readonly isRefreshing: boolean;
  readonly onRefresh: () => void | Promise<void>;
}) {
  const colors = useAppearanceAwarePalette();
  const styles = createHomeScreenStyles(colors);
  return (
    <ScrollView
      contentInsetAdjustmentBehavior="automatic"
      contentContainerStyle={styles.content}
      keyboardDismissMode={appKeyboardDismissMode()}
      keyboardShouldPersistTaps="handled"
      refreshControl={
        <RefreshControl
          refreshing={isRefreshing}
          tintColor={colors.action}
          onRefresh={onRefresh}
        />
      }
    >
      <DashboardHeader
        assetCheckoutCommand={assetCheckoutCommand}
        dashboard={dashboard}
        onDashboardChanged={onRefresh}
      />
    </ScrollView>
  );
}

type PendingReturnState = {
  readonly asset: AssetCardViewModel;
  readonly checkoutId: string;
  readonly undoableOperationId: string | undefined;
  readonly details: string;
  readonly isSaving: boolean;
};

function DashboardHeader({
  assetCheckoutCommand,
  dashboard,
  onDashboardChanged
}: {
  readonly assetCheckoutCommand: AssetCheckoutCommand;
  readonly dashboard: HomeDashboardViewModel;
  readonly onDashboardChanged: () => void | Promise<void>;
}) {
  const colors = useAppearanceAwarePalette();
  const styles = createHomeScreenStyles(colors);
  const feedback = useAppFeedback();
  const [returningAssetId, setReturningAssetId] = useState<string | undefined>();
  const [pendingReturn, setPendingReturn] = useState<PendingReturnState | undefined>();

  async function returnAsset(asset: AssetCardViewModel): Promise<void> {
    setReturningAssetId(asset.id);

    try {
      const checkout = await assetCheckoutCommand.execute({ action: 'return', assetId: asset.id });
      if (!checkout.undoableOperationId) {
        feedback.showNotice({
          tone: 'warning',
          title: 'Return completed without undo',
          message: 'The asset was returned, but this return cannot be canceled.'
        });
      }
      setPendingReturn({
        asset,
        checkoutId: checkout.id,
        undoableOperationId: checkout.undoableOperationId,
        details: '',
        isSaving: false
      });
      void onDashboardChanged();
    } catch (error) {
      feedback.showNotice({
        tone: 'error',
        title: 'Could not return asset',
        message: readableError(error, 'The asset was not returned.')
      });
    } finally {
      setReturningAssetId(undefined);
    }
  }

  async function saveReturnDetails(): Promise<void> {
    if (!pendingReturn) {
      return;
    }
    setPendingReturn({ ...pendingReturn, isSaving: true });

    try {
      await assetCheckoutCommand.updateReturnedCheckoutDetails({
        assetId: pendingReturn.asset.id,
        checkoutId: pendingReturn.checkoutId,
        details: pendingReturn.details
      });
      setPendingReturn(undefined);
      await onDashboardChanged();
    } catch (error) {
      setPendingReturn({ ...pendingReturn, isSaving: false });
      feedback.showNotice({
        tone: 'error',
        title: 'Could not save return details',
        message: readableError(error, 'Return details were not saved.')
      });
    }
  }

  async function cancelReturn(): Promise<void> {
    if (!pendingReturn) {
      return;
    }
    if (!pendingReturn.undoableOperationId) {
      setPendingReturn(undefined);
      return;
    }
    setPendingReturn({ ...pendingReturn, isSaving: true });

    try {
      await assetCheckoutCommand.undoOperation({ operationId: pendingReturn.undoableOperationId });
      setPendingReturn(undefined);
      await onDashboardChanged();
    } catch (error) {
      setPendingReturn({ ...pendingReturn, isSaving: false });
      feedback.showNotice({
        tone: 'error',
        title: 'Could not cancel return',
        message: readableError(error, 'The asset is still returned.')
      });
    }
  }

  return (
    <View>
      <View style={styles.homeTopBar}>
        <Pressable
          accessibilityLabel={`Current inventory ${dashboard.inventoryName}, tenant ${dashboard.tenantName}. Switch inventory`}
          accessibilityRole="button"
          onPress={() => router.push('/tenant-switcher')}
          style={styles.contextControl}
        >
          <View style={styles.contextText}>
            <Text style={styles.contextInventory}>{dashboard.inventoryName}</Text>
            <Text style={styles.contextTenantPrefix}>{dashboard.tenantName}</Text>
          </View>
          <ChevronDown color={colors.textMuted} size={18} strokeWidth={2} />
        </Pressable>
        <View style={styles.topBarActions}>
          {dashboard.canAdd ? (
            <Pressable
              accessibilityLabel="Add an asset"
              accessibilityRole="button"
              onPress={() => router.push('/add')}
              style={styles.settingsButton}
            >
              <Plus color={colors.action} size={24} strokeWidth={2.2} />
            </Pressable>
          ) : null}
          <Pressable
            accessibilityLabel="Open account and settings"
            accessibilityRole="button"
            onPress={() => router.push('/settings')}
            style={styles.settingsButton}
          >
            <UserCircle color={colors.text} size={24} strokeWidth={2} />
          </Pressable>
        </View>
      </View>

      <View style={styles.sectionHeader}>
        <Text accessibilityRole="header" style={styles.sectionTitle}>Recently changed</Text>
        <Pressable
          accessibilityLabel="View all recently changed assets"
          accessibilityRole="button"
          onPress={() => router.push('/assets')}
          style={styles.sectionActionButton}
        >
          <Text style={styles.sectionAction}>See all</Text>
        </Pressable>
      </View>
      <View style={styles.recentTicker}>
        {dashboard.recentAssets.slice(0, 3).map((asset) => (
          <AssetCard
            asset={asset}
            density="row"
            key={asset.id}
            palette={colors}
            showUpdatedAt
            onParentLocationPress={(location) => router.push(assetDetailHref(location.id))}
            onPress={() => router.push(assetDetailHref(asset.id))}
          />
        ))}
        {dashboard.recentAssets.length === 0 ? (
          <Text style={styles.emptyText}>No assets yet.</Text>
        ) : null}
      </View>

      {dashboard.checkedOutAssets.length > 0 ? (
        <View style={styles.attentionSection}>
          <View style={styles.sectionHeader}>
            <Text accessibilityRole="header" style={styles.sectionTitle}>Checked out</Text>
            <Pressable
              accessibilityLabel="View all checked-out assets"
              accessibilityRole="button"
              onPress={() => router.navigate({ pathname: '/search', params: { checkoutState: 'checked_out' } })}
              style={styles.sectionActionButton}
            >
              <Text style={styles.sectionAction}>View all</Text>
            </Pressable>
          </View>
          <View style={styles.recentTicker}>
            {dashboard.checkedOutAssets.slice(0, 3).map((asset) => (
              <AssetCard
                asset={asset}
                density="row"
                footerAction={{
                  accessibilityLabel: `Return ${asset.title}`,
                  disabled: returningAssetId === asset.id,
                  label: returningAssetId === asset.id ? 'Returning...' : 'Return',
                  onPress: () => void returnAsset(asset)
                }}
                key={asset.id}
                palette={colors}
                onParentLocationPress={(location) => router.push(assetDetailHref(location.id))}
                onPress={() => router.push(assetDetailHref(asset.id))}
                showTags={false}
              />
            ))}
          </View>
        </View>
      ) : null}

      <ReturnDetailsSheet
        pendingReturn={pendingReturn}
        onCancel={() => void cancelReturn()}
        onChangeDetails={(details) => {
          if (pendingReturn) {
            setPendingReturn({ ...pendingReturn, details });
          }
        }}
        onSave={() => void saveReturnDetails()}
      />

    </View>
  );
}

function ReturnDetailsSheet({
  pendingReturn,
  onCancel,
  onChangeDetails,
  onSave
}: {
  readonly pendingReturn: PendingReturnState | undefined;
  readonly onCancel: () => void;
  readonly onChangeDetails: (details: string) => void;
  readonly onSave: () => void;
}) {
  const colors = useAppearanceAwarePalette();
  const styles = createHomeScreenStyles(colors);
  return (
    <Modal
      animationType="slide"
      onRequestClose={onCancel}
      presentationStyle="pageSheet"
      transparent={false}
      visible={pendingReturn !== undefined}
    >
      <SafeAreaView style={styles.returnSheet} edges={['top', 'left', 'right', 'bottom']}>
        <View style={styles.returnSheetHeader}>
          <Text style={styles.returnSheetTitle}>Return details</Text>
          <Text style={styles.returnSheetSubtitle} numberOfLines={2}>
            {pendingReturn?.asset.title}
          </Text>
        </View>
        <AppTextInput
          multiline
          editable={!pendingReturn?.isSaving}
          onChangeText={onChangeDetails}
          placeholder="Optional details"
          placeholderTextColor={colors.textMuted}
          style={styles.returnDetailsInput}
          textAlignVertical="top"
          value={pendingReturn?.details ?? ''}
        />
        <View style={styles.returnSheetActions}>
          <Pressable
            accessibilityRole="button"
            disabled={pendingReturn?.isSaving}
            onPress={onCancel}
            style={[styles.returnSheetButton, styles.returnSheetCancelButton]}
          >
            <Text style={styles.returnSheetCancelText}>
              {pendingReturn?.undoableOperationId ? 'Cancel return' : 'Close'}
            </Text>
          </Pressable>
          <Pressable
            accessibilityRole="button"
            disabled={pendingReturn?.isSaving}
            onPress={onSave}
            style={[styles.returnSheetButton, styles.returnSheetSaveButton]}
          >
            <Text style={styles.returnSheetSaveText}>{pendingReturn?.isSaving ? 'Saving...' : 'Save'}</Text>
          </Pressable>
        </View>
      </SafeAreaView>
    </Modal>
  );
}
