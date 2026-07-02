import { useCallback, useEffect, useRef, useState } from 'react';
import { router, useFocusEffect } from 'expo-router';
import { Settings } from 'lucide-react-native';
import {
  ActivityIndicator,
  Image,
  Pressable,
  RefreshControl,
  ScrollView,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import {
  HomeDashboardLocationViewModel,
  HomeDashboardQuery,
  HomeDashboardViewModel
} from '../../application/home/HomeDashboardQuery';
import type { AssetCardViewModel } from '../../application/assets/AssetViewModels';
import { BrandMark } from '../components/BrandMark';
import { IdentityLabel } from '../components/IdentityIcon';
import { colors } from '../theme/tokens';
import { assetDetailHref } from './AssetDetailNavigation';
import { styles } from './HomeScreen.styles';

type HomeScreenProps = {
  readonly dashboardQuery: HomeDashboardQuery;
};

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly dashboard: HomeDashboardViewModel }
  | { readonly status: 'error'; readonly message: string };

export function HomeScreen({ dashboardQuery }: HomeScreenProps) {
  const [screenState, setScreenState] = useState<ScreenState>({ status: 'loading' });
  const [isRefreshing, setIsRefreshing] = useState(false);
  const didInitialLoadRef = useRef(false);

  useEffect(() => {
    let isCurrent = true;

    dashboardQuery
      .execute()
      .then((dashboard) => {
        if (isCurrent) {
          didInitialLoadRef.current = true;
          setScreenState({ status: 'ready', dashboard });
        }
      })
      .catch((error: unknown) => {
        if (isCurrent) {
          setScreenState({
            status: 'error',
            message: readableError(error, 'Stuff Stash could not load the mobile home screen.')
          });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [dashboardQuery]);

  useFocusEffect(
    useCallback(() => {
      if (!didInitialLoadRef.current) {
        return;
      }

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
              message: readableError(error, 'Stuff Stash could not refresh the mobile home screen.')
            });
          }
        });

      return () => {
        isCurrent = false;
      };
    }, [dashboardQuery])
  );

  async function refreshDashboard(): Promise<void> {
    setIsRefreshing(true);

    try {
      const dashboard = await dashboardQuery.execute();
      setScreenState({ status: 'ready', dashboard });
    } catch (error) {
      setScreenState({
        status: 'error',
        message: readableError(error, 'Stuff Stash could not refresh the mobile home screen.')
      });
    } finally {
      setIsRefreshing(false);
    }
  }

  return (
    <SafeAreaView style={styles.shell} edges={['top', 'left', 'right']}>
      {screenState.status === 'loading' ? <LoadingState /> : null}
      {screenState.status === 'error' ? <ErrorState message={screenState.message} /> : null}
      {screenState.status === 'ready' ? (
        <Dashboard
          dashboard={screenState.dashboard}
          isRefreshing={isRefreshing}
          onRefresh={refreshDashboard}
        />
      ) : null}
    </SafeAreaView>
  );
}

function LoadingState() {
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={colors.accent} />
      <Text style={styles.stateText}>Loading Stuff Stash</Text>
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

function Dashboard({
  dashboard,
  isRefreshing,
  onRefresh
}: {
  readonly dashboard: HomeDashboardViewModel;
  readonly isRefreshing: boolean;
  readonly onRefresh: () => void;
}) {
  return (
    <ScrollView
      contentContainerStyle={styles.content}
      refreshControl={
        <RefreshControl
          refreshing={isRefreshing}
          tintColor={colors.action}
          onRefresh={onRefresh}
        />
      }
    >
      <DashboardHeader dashboard={dashboard} />
    </ScrollView>
  );
}

function DashboardHeader({ dashboard }: {
  readonly dashboard: HomeDashboardViewModel;
}) {
  return (
    <View>
      <View style={styles.homeTopBar}>
        <Pressable
          accessibilityRole="button"
          onPress={() => router.push('/tenant-switcher')}
          style={styles.contextControl}
        >
          <BrandMark size="sm" />
          <View style={styles.contextText}>
            <View style={styles.contextLine}>
              <Text numberOfLines={1} style={styles.contextTenantPrefix}>
                {dashboard.tenantName} /
              </Text>
              <IdentityLabel
                kind="inventory"
                label={dashboard.inventoryName}
                textStyle={styles.contextInventory}
              />
            </View>
          </View>
        </Pressable>
        <Pressable
          accessibilityLabel="Open Settings"
          accessibilityRole="button"
          onPress={() => router.push('/settings')}
          style={styles.settingsButton}
        >
          <Settings color={colors.text} size={22} strokeWidth={2.2} />
        </Pressable>
      </View>

      <View style={styles.sectionHeader}>
        <Text style={styles.sectionTitle}>Recently changed</Text>
        <Pressable accessibilityRole="button" onPress={() => router.push('/assets')}>
          <Text style={styles.sectionAction}>See all</Text>
        </Pressable>
      </View>
      <ScrollView
        contentContainerStyle={styles.recentTicker}
        horizontal
        showsHorizontalScrollIndicator={false}
      >
        {dashboard.recentAssets.map((asset) => (
          <RecentAssetCard
            asset={asset}
            key={asset.id}
            onPress={() => router.push(assetDetailHref(asset.id))}
          />
        ))}
        {dashboard.recentAssets.length === 0 ? (
          <Text style={styles.emptyText}>No assets yet.</Text>
        ) : null}
      </ScrollView>

      <View style={styles.sectionHeader}>
        <Text style={styles.sectionTitle}>Locations</Text>
        <Pressable accessibilityRole="button" onPress={() => router.navigate({ pathname: '/search', params: { scope: 'places' } })}>
          <Text style={styles.sectionAction}>View all</Text>
        </Pressable>
      </View>
      <View style={styles.locationGrid}>
        {dashboard.topLocations.map((location) => (
          <LocationCard key={location.id} location={location} />
        ))}
        {dashboard.topLocations.length === 0 ? (
          <Text style={styles.emptyText}>No locations yet.</Text>
        ) : null}
      </View>
    </View>
  );
}

function LocationCard({ location }: { readonly location: HomeDashboardLocationViewModel }) {
  return (
    <Pressable
      accessibilityRole="button"
      onPress={() => router.push(`/locations/${location.id}`)}
      style={styles.locationCard}
    >
      <View style={styles.locationImageFrame}>
        {location.photo ? (
          <Image
            accessibilityIgnoresInvertColors
            source={{ uri: location.photo.uri, headers: location.photo.headers }}
            style={styles.locationImage}
          />
        ) : (
          <Text style={styles.locationImagePlaceholder}>Place</Text>
        )}
      </View>
      <Text style={styles.locationTitle}>{location.title}</Text>
      <Text style={styles.locationDescription}>{location.description}</Text>
      <View style={styles.locationFooter}>
        <Text style={styles.locationCount}>{location.containedAssetCountLabel} assets</Text>
        <Text style={location.photoLabel === 'Photo ready' ? styles.photoReady : styles.photoNeeded}>
          {location.photoLabel}
        </Text>
      </View>
      <Text style={styles.recentAssetLabel}>{location.recentAssetLabel || 'No recent assets'}</Text>
    </Pressable>
  );
}

function RecentAssetCard({
  asset,
  onPress
}: {
  readonly asset: AssetCardViewModel;
  readonly onPress: () => void;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      onPress={onPress}
      style={styles.recentCard}
    >
      <View style={styles.recentImageFrame}>
        {asset.photo ? (
          <Image
            accessibilityIgnoresInvertColors
            source={{ uri: asset.photo.uri, headers: asset.photo.headers }}
            style={styles.recentImage}
          />
        ) : (
          <Text style={styles.recentImagePlaceholder}>{asset.imagePlaceholderLabel}</Text>
        )}
      </View>
      <View style={styles.recentBody}>
        <View style={styles.badgeRow}>
          <Text style={styles.kindBadge}>{asset.kindLabel}</Text>
          {asset.customTypeLabel ? <Text style={styles.customTypeBadge}>{asset.customTypeLabel}</Text> : null}
        </View>
        <Text numberOfLines={2} style={styles.assetTitle}>
          {asset.title}
        </Text>
        <Text numberOfLines={1} style={styles.assetMeta}>
          {asset.locationTrailLabel}
        </Text>
        <Text numberOfLines={1} style={styles.assetMeta}>
          {asset.updatedAtLabel}
        </Text>
      </View>
    </Pressable>
  );
}
