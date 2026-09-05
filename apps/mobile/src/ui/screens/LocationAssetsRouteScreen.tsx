import { useMemo } from 'react';
import { router, Stack } from 'expo-router';
import {
  ActivityIndicator,
  FlatList,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import {
  LocationAssetsQuery,
  LocationAssetsViewModel
} from '../../application/locations/LocationAssetsQuery';
import { AssetCard } from '../components/AssetCard';
import { IdentityLabel } from '../components/IdentityIcon';
import { assetDetailHref, locationAssetDetailHref } from './AssetDetailNavigation';
import { navigateToAssetTagSearch } from './AssetTagSearchNavigation';
import { spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';

type LocationAssetsRouteScreenProps = {
  readonly locationAssetsQuery: LocationAssetsQuery;
  readonly locationId: string;
};

export function LocationAssetsRouteScreen({
  locationAssetsQuery,
  locationId
}: LocationAssetsRouteScreenProps) {
  const palette = useAppearancePalette();
  const styles = useMemo(() => createStyles(palette), [palette]);
  const locationAssets = useMobileInventoryServerQuery({
    key: (scopeId, tenantId, inventoryId) =>
      mobileQueryKeys.locationAssets(scopeId, tenantId, inventoryId, locationId),
    query: (signal) => locationAssetsQuery.execute(locationId, { signal })
  });

  async function refreshLocationAssets(): Promise<void> {
    await locationAssets.refetch();
  }

  return (
    <SafeAreaView style={styles.shell} edges={['left', 'right']}>
      {locationAssets.isPending && !locationAssets.data ? <LoadingState /> : null}
      {locationAssets.isError && !locationAssets.data ? (
        <ErrorState message={readableError(locationAssets.error, 'Could not load location.')} />
      ) : null}
      {locationAssets.data ? (
        <LocationAssetList
          isRefreshing={locationAssets.isRefetching}
          locationAssets={locationAssets.data}
          onRefresh={() => { void refreshLocationAssets(); }}
        />
      ) : null}
    </SafeAreaView>
  );
}

export function LocationAssetList({
  isRefreshing,
  locationAssets,
  onRefresh
}: {
  readonly isRefreshing: boolean;
  readonly locationAssets: LocationAssetsViewModel;
  readonly onRefresh: () => void;
}) {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  return (
    <>
      <Stack.Screen options={{ title: locationAssets.locationTitle }} />
      <FlatList
        data={locationAssets.assets}
        keyExtractor={(asset) => asset.id}
        columnWrapperStyle={styles.cardRow}
        contentContainerStyle={styles.content}
        numColumns={2}
        refreshing={isRefreshing}
        onRefresh={onRefresh}
        ListHeaderComponent={
          <View>
            <Text style={styles.title}>{locationAssets.locationTitle}</Text>
            <IdentityLabel
              iconSize="xs"
              kind="inventory"
              label={locationAssets.inventoryName}
              style={styles.contextLine}
              textStyle={styles.contextText}
            />
          </View>
        }
        ListEmptyComponent={<Text style={styles.emptyText}>No assets in this location.</Text>}
        renderItem={({ item }) => (
          <AssetCard
            asset={item}
            palette={palette}
            onParentLocationPress={(location) => router.push(assetDetailHref(location.id))}
            onPress={() => router.push(locationAssetDetailHref(locationAssets.locationId, item.id))}
            onTagPress={(tag) => navigateToAssetTagSearch(router, tag)}
          />
        )}
      />
    </>
  );
}

function LoadingState() {
  const palette = useAppearancePalette();
  const styles = useMemo(() => createStyles(palette), [palette]);
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={palette.accent} />
      <Text style={styles.stateText}>Loading location</Text>
    </View>
  );
}

function ErrorState({ message }: { readonly message: string }) {
  const palette = useAppearancePalette();
  const styles = useMemo(() => createStyles(palette), [palette]);
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

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  shell: {
    flex: 1,
    backgroundColor: colors.background
  },
  content: {
    padding: spacing.lg,
    paddingBottom: spacing.xl
  },
  centerState: {
    alignItems: 'center',
    flex: 1,
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
  title: {
    color: colors.text,
    fontSize: 30,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 36
  },
  contextLine: {
    marginBottom: spacing.md,
    marginTop: spacing.xs
  },
  contextText: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '700',
    letterSpacing: 0
  },
  emptyText: {
    color: colors.textMuted,
    fontSize: 15,
    lineHeight: 22
  },
  cardRow: {
    gap: spacing.sm,
    marginBottom: spacing.sm
  }
  });
}
