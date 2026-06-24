import { useState } from 'react';
import { router } from 'expo-router';
import {
  ActivityIndicator,
  FlatList,
  Pressable,
  StyleSheet,
  Text,
  TextInput,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import {
  SearchAssetsQuery,
  SearchAssetsViewModel
} from '../../application/search/SearchAssetsQuery';
import { AssetCard } from '../components/AssetCard';
import { colors, radius, spacing } from '../theme/tokens';

type SearchScreenProps = {
  readonly searchAssetsQuery: SearchAssetsQuery;
};

type SearchState =
  | { readonly status: 'idle'; readonly results: SearchAssetsViewModel }
  | { readonly status: 'loading'; readonly results: SearchAssetsViewModel }
  | { readonly status: 'error'; readonly message: string; readonly results: SearchAssetsViewModel };

type DetailState =
  | { readonly status: 'closed' }
  | { readonly status: 'loading' }
  | { readonly status: 'error'; readonly message: string };

const emptyResults: SearchAssetsViewModel = { query: '', assets: [], assetDetails: [] };

export function SearchScreen({ searchAssetsQuery }: SearchScreenProps) {
  const [query, setQuery] = useState('');
  const [state, setState] = useState<SearchState>({ status: 'idle', results: emptyResults });
  const [detailState, setDetailState] = useState<DetailState>({ status: 'closed' });
  const [isRefreshing, setIsRefreshing] = useState(false);

  async function submitSearch(): Promise<void> {
    setState({ status: 'loading', results: state.results });
    setDetailState({ status: 'closed' });

    try {
      const results = await searchAssetsQuery.execute(query);
      setState({ status: 'idle', results });
    } catch (error) {
      setState({
        status: 'error',
        message: readableError(error, 'Search failed.'),
        results: state.results
      });
    }
  }

  async function openAsset(assetId: string): Promise<void> {
    setDetailState({ status: 'loading' });

    if (!state.results.assetDetails.some((item) => item.id === assetId)) {
      setDetailState({ status: 'error', message: 'Could not load asset.' });
      return;
    }

    setDetailState({ status: 'closed' });
    router.push(`/assets/${assetId}`);
  }

  async function refreshResults(): Promise<void> {
    if (state.results.query.length === 0) {
      return;
    }

    setIsRefreshing(true);
    setDetailState({ status: 'closed' });

    try {
      const results = await searchAssetsQuery.execute(state.results.query);
      setState({ status: 'idle', results });
    } catch (error) {
      setState({
        status: 'error',
        message: readableError(error, 'Search refresh failed.'),
        results: state.results
      });
    } finally {
      setIsRefreshing(false);
    }
  }

  return (
    <SafeAreaView style={styles.shell} edges={['top', 'left', 'right']}>
      <FlatList
        data={state.results.assets}
        keyExtractor={(asset) => asset.id}
        columnWrapperStyle={styles.cardRow}
        contentContainerStyle={styles.content}
        keyboardShouldPersistTaps="handled"
        numColumns={2}
        refreshing={isRefreshing}
        onRefresh={refreshResults}
        ListHeaderComponent={
          <View>
            <Text style={styles.title}>Search</Text>
            <View style={styles.searchRow}>
              <TextInput
                accessibilityLabel="Search inventory"
                autoCapitalize="none"
                onChangeText={setQuery}
                onSubmitEditing={submitSearch}
                placeholder="Search assets"
                placeholderTextColor={colors.textMuted}
                returnKeyType="search"
                style={styles.searchInput}
                value={query}
              />
              <Pressable accessibilityRole="button" onPress={submitSearch} style={styles.searchButton}>
                {state.status === 'loading' ? (
                  <ActivityIndicator color={colors.onAction} />
                ) : (
                  <Text style={styles.searchButtonText}>Search</Text>
                )}
              </Pressable>
            </View>
            {state.status === 'error' ? <Text style={styles.errorText}>{state.message}</Text> : null}
            {detailState.status === 'loading' ? <Text style={styles.resultCount}>Loading asset</Text> : null}
            {detailState.status === 'error' ? <Text style={styles.errorText}>{detailState.message}</Text> : null}
            {state.results.query.length > 0 ? (
              <Text style={styles.resultCount}>
                {state.results.assets.length.toString()} results for {state.results.query}
              </Text>
            ) : null}
          </View>
        }
        ListEmptyComponent={
          <Text style={styles.emptyText}>
            {state.results.query.length > 0 ? 'No matching assets.' : ' '}
          </Text>
        }
        renderItem={({ item }) => <AssetCard asset={item} onPress={() => openAsset(item.id)} />}
      />
    </SafeAreaView>
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
  title: {
    color: colors.text,
    fontSize: 30,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 36,
    marginBottom: spacing.md
  },
  searchRow: {
    flexDirection: 'row',
    gap: spacing.sm,
    marginBottom: spacing.sm
  },
  searchInput: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    color: colors.text,
    flex: 1,
    fontSize: 16,
    minHeight: 46,
    paddingHorizontal: spacing.md
  },
  searchButton: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    justifyContent: 'center',
    minHeight: 46,
    minWidth: 86,
    paddingHorizontal: spacing.md
  },
  searchButtonText: {
    color: colors.onAction,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0
  },
  errorText: {
    color: colors.warning,
    fontSize: 14,
    lineHeight: 20,
    marginBottom: spacing.sm
  },
  resultCount: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '700',
    letterSpacing: 0,
    marginBottom: spacing.md
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
