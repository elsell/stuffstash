import { useEffect, useState } from 'react';
import { router, Stack } from 'expo-router';
import {
  ActivityIndicator,
  Alert,
  RefreshControl,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { AssetLifecycleCommand } from '../../application/assets/AssetLifecycleCommand';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { AssetDetailQuery } from '../../application/assets/AssetDetailQuery';
import { AssetDetailView } from '../components/AssetDetailView';
import { navigateAfterDeletedAsset } from './AssetDetailNavigation';
import { colors, spacing } from '../theme/tokens';

type AssetDetailRouteScreenProps = {
  readonly assetDetailQuery: AssetDetailQuery;
  readonly assetLifecycleCommand: AssetLifecycleCommand;
  readonly assetId: string;
};

type ScreenState =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly asset: AssetDetailViewModel }
  | { readonly status: 'error'; readonly message: string };

export function AssetDetailRouteScreen({
  assetDetailQuery,
  assetLifecycleCommand,
  assetId
}: AssetDetailRouteScreenProps) {
  const [screenState, setScreenState] = useState<ScreenState>({ status: 'loading' });
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [pendingAction, setPendingAction] = useState<'archive' | 'restore' | 'delete' | undefined>();

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

  function confirmArchive(): void {
    Alert.alert(
      'Archive asset?',
      'Archived assets are hidden from normal inventory work and can be restored later.',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Archive',
          style: 'destructive',
          onPress: () => {
            void runLifecycleAction('archive');
          }
        }
      ]
    );
  }

  function confirmRestore(): void {
    Alert.alert('Restore asset?', 'This returns the asset to active inventory work.', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Restore',
        onPress: () => {
          void runLifecycleAction('restore');
        }
      }
    ]);
  }

  function confirmDelete(): void {
    Alert.alert(
      'Delete permanently?',
      'This removes the asset permanently. Audit history is preserved.',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: () => {
            void runLifecycleAction('delete');
          }
        }
      ]
    );
  }

  async function runLifecycleAction(action: 'archive' | 'restore' | 'delete'): Promise<void> {
    setPendingAction(action);

    try {
      await assetLifecycleCommand.execute({ action, assetId });

      if (action === 'delete') {
        navigateAfterDeletedAsset(router);
        return;
      }

      const asset = await assetDetailQuery.execute(assetId);
      setScreenState({ status: 'ready', asset });
    } catch (error) {
      Alert.alert('Could not update asset', readableError(error, 'Lifecycle action failed.'));
    } finally {
      setPendingAction(undefined);
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
            isLifecycleActionPending={pendingAction !== undefined}
            onArchive={confirmArchive}
            onRestore={confirmRestore}
            onDeletePermanently={confirmDelete}
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
