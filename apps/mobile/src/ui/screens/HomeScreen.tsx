import { type ReactNode } from 'react';
import { useHomeReturnActions } from './useHomeReturnActions';
import { router } from 'expo-router';
import { usePullRefresh } from '../serverState/usePullRefresh';
import { useHomeReturnTaskPresentation } from '../navigation/HomeReturnTaskPresentation';
import { HomeReturnDetailsSheet } from './HomeReturnDetailsSheet';
import { HomeNavigationHeader } from './HomeNavigationHeader';
import type { NativeHeaderAction } from '../components/NativeHeaderActions.types';
import {
  ActivityIndicator,
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
import { appKeyboardDismissMode } from '../components/AppTextInput';
import { useAppFeedback } from '../feedback/AppFeedback';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { assetDetailHref } from './AssetDetailNavigation';
import { createHomeScreenStyles } from './HomeScreen.styles';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';

type HomeScreenProps = {
  readonly notificationAction?: NativeHeaderAction;
  readonly expirationSection?: ReactNode;
  readonly onRefreshAdditional?: () => Promise<void>;
  readonly dashboardQuery: HomeDashboardQuery;
  readonly assetCheckoutCommand: AssetCheckoutCommand;
};

export function HomeScreen({ assetCheckoutCommand, dashboardQuery, notificationAction, expirationSection, onRefreshAdditional }: HomeScreenProps) {
  const styles = createHomeScreenStyles(useAppearanceAwarePalette());
  const feedback = useAppFeedback();
  const dashboardState = useMobileInventoryServerQuery({
    key: mobileQueryKeys.home,
    query: (signal) => dashboardQuery.execute({ signal })
  });

  async function refreshDashboard(shouldNotify: () => boolean = () => true): Promise<void> {
    try {
      await Promise.all([dashboardState.refetch({ throwOnError: true }), onRefreshAdditional?.()]);
    } catch (error) {
      if (!shouldNotify()) return;
      feedback.showNotice({
        tone: 'error',
        title: 'Could not refresh Home',
        message: readableError(error, 'Stuff Stash could not refresh the mobile home screen.')
      });
    }
  }

  const pullRefresh = usePullRefresh(refreshDashboard);

  return (
    <SafeAreaView style={styles.shell} edges={['left', 'right']}>
      <HomeNavigationHeader dashboard={dashboardState.data} notificationAction={notificationAction} />
      {dashboardState.isPending && !dashboardState.data ? <LoadingState /> : null}
      {dashboardState.isError && !dashboardState.data ? (
        <ErrorState
          message={readableError(dashboardState.error, 'Stuff Stash could not load the mobile home screen.')}
          onRetry={() => { void dashboardState.refetch(); }}
        />
      ) : null}
      {dashboardState.data ? (
        <Dashboard
          expirationSection={expirationSection}
          assetCheckoutCommand={assetCheckoutCommand}
          dashboard={dashboardState.data}
          isRefreshing={pullRefresh.refreshing}
          onRefresh={pullRefresh.refresh}
          onDashboardChanged={refreshDashboard}
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
  expirationSection,
  assetCheckoutCommand,
  dashboard,
  isRefreshing,
  onRefresh,
  onDashboardChanged
}: {
  readonly expirationSection?: ReactNode;
  readonly assetCheckoutCommand: AssetCheckoutCommand;
  readonly dashboard: HomeDashboardViewModel;
  readonly isRefreshing: boolean;
  readonly onRefresh: () => void | Promise<void>;
  readonly onDashboardChanged: (shouldNotify: () => boolean) => void | Promise<void>;
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
        key={`${dashboard.tenantId}:${dashboard.inventoryId}`}
          expirationSection={expirationSection}
        assetCheckoutCommand={assetCheckoutCommand}
        dashboard={dashboard}
        onDashboardChanged={onDashboardChanged}
      />
    </ScrollView>
  );
}

function DashboardHeader({
  expirationSection,
  assetCheckoutCommand,
  dashboard,
  onDashboardChanged
}: {
  readonly expirationSection?: ReactNode;
  readonly assetCheckoutCommand: AssetCheckoutCommand;
  readonly dashboard: HomeDashboardViewModel;
  readonly onDashboardChanged: (shouldNotify: () => boolean) => void | Promise<void>;
}) {
  const colors = useAppearanceAwarePalette();
  const styles = createHomeScreenStyles(colors);
  const { returningAssetId, pendingReturn, returnAsset, isReturnDisabled, saveReturnDetails, cancelReturn, closeReturn, changeDetails, setEditorFocused } = useHomeReturnActions(assetCheckoutCommand, onDashboardChanged, dashboard.checkedOutAssets, dashboard.canReturn);

  useHomeReturnTaskPresentation(pendingReturn ? {
    content: <HomeReturnDetailsSheet
        canReturn={dashboard.canReturn}
        onClose={closeReturn}
        pendingReturn={pendingReturn}
        onCancel={() => void cancelReturn()}
        onChangeDetails={changeDetails}
        onSave={() => void saveReturnDetails()}
      />,
    requestClose: () => { if (dashboard.canReturn) void cancelReturn(); else closeReturn(); },
    focusChanged: setEditorFocused
  } : undefined);

  return (
    <View>

      {expirationSection}
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
                footerAction={dashboard.canReturn ? {
                  accessibilityLabel: `Return ${asset.title}`,
                  disabled: isReturnDisabled(asset),
                  label: returningAssetId === asset.id ? 'Returning...' : 'Return',
                  onPress: () => void returnAsset(asset)
                } : undefined}
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

    </View>
  );
}
