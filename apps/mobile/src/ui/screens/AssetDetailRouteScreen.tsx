import { useEffect, useState } from 'react';
import { Stack } from 'expo-router';
import {
  ActivityIndicator,
  RefreshControl,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { AssetDetailQuery } from '../../application/assets/AssetDetailQuery';
import { AssetDetailView } from '../components/AssetDetailView';
import { colors, spacing } from '../theme/tokens';

type AssetDetailRouteScreenProps = {
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetId: string;
};

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly asset: AssetDetailViewModel }
  | { readonly status: 'error'; readonly message: string };

export function AssetDetailRouteScreen({
  assetDetailQuery,
  assetId
}: AssetDetailRouteScreenProps) {
  const [screenState, setScreenState] = useState<ScreenState>({ status: 'loading' });
  const [isRefreshing, setIsRefreshing] = useState(false);

  useEffect(() => {
    let isCurrent = true;

    assetDetailQuery
      .execute(assetId)
      .then((asset) => {
        if (isCurrent) {
          setScreenState({ status: 'ready', asset });
        }
      })
      .catch((error: unknown) => {
        if (isCurrent) {
          setScreenState({
            status: 'error',
            message: readableError(error, 'Could not load asset.')
          });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [assetDetailQuery, assetId]);

  async function refreshAsset(): Promise<void> {
    setIsRefreshing(true);

    try {
      const asset = await assetDetailQuery.execute(assetId);
      setScreenState({ status: 'ready', asset });
    } catch (error) {
      setScreenState({
        status: 'error',
        message: readableError(error, 'Could not refresh asset.')
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
        <>
          <Stack.Screen options={{ title: screenState.asset.title }} />
          <AssetDetailView
            asset={screenState.asset}
            refreshControl={
              <RefreshControl
                refreshing={isRefreshing}
                tintColor={colors.action}
                onRefresh={refreshAsset}
              />
            }
          />
        </>
      ) : null}
    </SafeAreaView>
  );
}

function LoadingState() {
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={colors.accent} />
      <Text style={styles.stateText}>Loading asset</Text>
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
  }
});
