import { useEffect, useState } from 'react';
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
  InventoryAssetsQuery,
  InventoryAssetsViewModel
} from '../../application/assets/InventoryAssetsQuery';
import { AssetCard } from '../components/AssetCard';
import { IdentityLabel } from '../components/IdentityIcon';
import { assetDetailHref } from './AssetDetailNavigation';
import { colors, spacing } from '../theme/tokens';

type InventoryAssetsRouteScreenProps = {
  readonly inventoryAssetsQuery: InventoryAssetsQuery;
};

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly inventoryAssets: InventoryAssetsViewModel }
  | { readonly status: 'error'; readonly message: string };

export function InventoryAssetsRouteScreen({
  inventoryAssetsQuery
}: InventoryAssetsRouteScreenProps) {
  const [screenState, setScreenState] = useState<ScreenState>({ status: 'loading' });
  const [isRefreshing, setIsRefreshing] = useState(false);

  useEffect(() => {
    let isCurrent = true;

    inventoryAssetsQuery
      .execute()
      .then((inventoryAssets) => {
        if (isCurrent) {
          setScreenState({ status: 'ready', inventoryAssets });
        }
      })
      .catch((error: unknown) => {
        if (isCurrent) {
          setScreenState({
            status: 'error',
            message: readableError(error, 'Could not load assets.')
          });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [inventoryAssetsQuery]);

  async function refreshInventoryAssets(): Promise<void> {
    setIsRefreshing(true);

    try {
      const inventoryAssets = await inventoryAssetsQuery.execute();
      setScreenState({ status: 'ready', inventoryAssets });
    } catch (error) {
      setScreenState({
        status: 'error',
        message: readableError(error, 'Could not refresh assets.')
      });
    } finally {
      setIsRefreshing(false);
    }
  }

  return (
    <SafeAreaView style={styles.shell} edges={['left', 'right']}>
      {screenState.status === 'loading' ? <LoadingState /> : null}
      {screenState.status === 'error' ? <ErrorState message={screenState.message} /> : null}
      {screenState.status === 'ready' ? (
        <InventoryAssetList
          inventoryAssets={screenState.inventoryAssets}
          isRefreshing={isRefreshing}
          onRefresh={refreshInventoryAssets}
        />
      ) : null}
    </SafeAreaView>
  );
}

function InventoryAssetList({
  inventoryAssets,
  isRefreshing,
  onRefresh
}: {
  readonly inventoryAssets: InventoryAssetsViewModel;
  readonly isRefreshing: boolean;
  readonly onRefresh: () => void;
}) {
  return (
    <>
      <Stack.Screen options={{ title: 'Assets' }} />
      <FlatList
        data={inventoryAssets.assets}
        keyExtractor={(asset) => asset.id}
        columnWrapperStyle={styles.cardRow}
        contentContainerStyle={styles.content}
        numColumns={2}
        refreshing={isRefreshing}
        onRefresh={onRefresh}
        ListHeaderComponent={
          <View>
            <Text style={styles.title}>Recently changed</Text>
            <IdentityLabel
              iconSize="xs"
              kind="inventory"
              label={inventoryAssets.inventoryName}
              style={styles.contextLine}
              textStyle={styles.contextText}
            />
          </View>
        }
        ListEmptyComponent={<Text style={styles.emptyText}>No assets yet.</Text>}
        renderItem={({ item }) => (
          <AssetCard asset={item} onPress={() => router.push(assetDetailHref(item.id))} />
        )}
      />
    </>
  );
}

function LoadingState() {
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={colors.accent} />
      <Text style={styles.stateText}>Loading assets</Text>
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
