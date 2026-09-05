import { isAccessFailure } from '../serverState/isAccessFailure';
import { useInfiniteQuery } from '@tanstack/react-query';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileServerStateScopeId } from '../navigation/MobileServerStateProvider';
import { useState } from 'react';
import { router, Stack } from 'expo-router';
import {
  ActivityIndicator,
  Pressable,
  RefreshControl,
  SectionList,
  StyleSheet,
  Text,
  View
} from 'react-native';
import {
  AssetActivityQuery,
  AssetActivityRecordViewModel,
  AssetActivityView
} from '../../application/assets/AssetActivityQuery';
import { groupHistoryRecords, historyFilterMenuGroups, historyLoadError } from './AssetHistoryPresentation';
import { useAppFeedback } from '../feedback/AppFeedback';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { NativeActionMenu } from '../components/NativeActionMenu';

type HistoryState =
  | { readonly status: 'loading' }
  | { readonly status: 'error'; readonly title: string; readonly message: string; readonly canRetry: boolean }
  | {
      readonly status: 'ready';
      readonly records: readonly AssetActivityRecordViewModel[];
      readonly nextCursor?: string;
      readonly hasMore: boolean;
      readonly emptyTitle: string;
      readonly emptyMessage: string;
    };

export function AssetHistoryRouteScreen({
  assetActivityQuery,
  tenantId,
  inventoryId,
  assetId,
  assetTitle
}: {
  readonly assetActivityQuery: AssetActivityQuery;
  readonly tenantId: string;
  readonly inventoryId: string;
  readonly assetId: string;
  readonly assetTitle: string;
}) {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  const feedback = useAppFeedback();
  const [view, setView] = useState<AssetActivityView>('changes');
  const scopeId = useMobileServerStateScopeId();
  const history = useInfiniteQuery({
    queryKey: mobileQueryKeys.assetHistory(scopeId, tenantId, inventoryId, assetId, view),
    initialPageParam: undefined as string | undefined,
    queryFn: ({ signal, pageParam }) => assetActivityQuery.execute({ tenantId, inventoryId, assetId, view, limit: 20, cursor: pageParam, signal }),
    getNextPageParam: (page) => page.hasMore ? page.nextCursor : undefined
  });
  const firstPage = isAccessFailure(history.error) ? undefined : history.data?.pages[0];
  const state: HistoryState = firstPage
    ? { ...firstPage, status: 'ready', records: history.data!.pages.flatMap((page) => page.records), hasMore: history.hasNextPage }
    : history.isError ? { status: 'error', ...historyLoadError(history.error) } : { status: 'loading' };
  const [isRefreshing, setIsRefreshing] = useState(false);
  const isLoadingMore = history.isFetchingNextPage;
  const pageError = history.isFetchNextPageError ? 'Older activity could not be loaded.' : undefined;

  async function refresh(): Promise<void> {
    setIsRefreshing(true);
    try {
      await history.refetch({ throwOnError: true });
    } catch {
      feedback.showNotice({ tone: 'error', title: 'Could not refresh History', message: 'Please try again when access and connectivity are available.' });
    } finally {
      setIsRefreshing(false);
    }
  }

  async function loadMore(): Promise<void> {
    if (history.hasNextPage && !history.isFetching) await history.fetchNextPage();
  }

  function openDetail(record: AssetActivityRecordViewModel): void {
    router.push({
      pathname: '/assets/[assetId]/history/[activityId]',
      params: { assetId, activityId: record.id, assetTitle, tenantId, inventoryId }
    });
  }

  return (
    <View style={styles.screen}>
      <Stack.Screen options={{ title: 'History' }} />
      <View style={styles.heading}>
        <Text accessibilityRole="header" style={styles.assetTitle}>{assetTitle}</Text>
        <HistoryFilter value={view} onChange={setView} styles={styles} />
      </View>
      {state.status === 'loading' ? <CenteredState label="Loading history" palette={palette} styles={styles} /> : null}
      {state.status === 'error' ? (
        <View style={styles.centerState}>
          <Text accessibilityRole="header" style={styles.stateTitle}>{state.title}</Text>
          <Text style={styles.stateMessage}>{state.message}</Text>
          {state.canRetry ? <Pressable accessibilityRole="button" onPress={() => void history.refetch()} style={styles.primaryButton}>
            <Text style={styles.primaryButtonText}>Try again</Text>
          </Pressable> : null}
        </View>
      ) : null}
      {state.status === 'ready' && history.isRefetchError ? <View style={styles.heading}>
        <Text accessibilityRole="alert" style={styles.pageError}>History could not be refreshed. Previously loaded activity is shown.</Text>
        <Pressable accessibilityRole="button" onPress={() => void refresh()} style={styles.secondaryButton}><Text style={styles.secondaryButtonText}>Try refreshing again</Text></Pressable>
      </View> : null}
      {state.status === 'ready' ? (
        <SectionList
          contentContainerStyle={state.records.length === 0 ? styles.emptyList : styles.list}
          sections={groupHistoryRecords(state.records)}
          keyExtractor={(record) => record.id}
          refreshControl={<RefreshControl refreshing={isRefreshing} onRefresh={() => void refresh()} tintColor={palette.action} />}
          renderItem={({ item }) => (
            <HistoryRow record={item} onPress={() => openDetail(item)} styles={styles} />
          )}
          renderSectionHeader={({ section }) => <Text accessibilityRole="header" style={styles.dateHeader}>{section.title}</Text>}
          ListEmptyComponent={(
            <View style={styles.centerState}>
              <Text accessibilityRole="header" style={styles.stateTitle}>{state.emptyTitle}</Text>
              <Text style={styles.stateMessage}>{state.emptyMessage}</Text>
            </View>
          )}
          ListFooterComponent={state.hasMore || pageError ? (
            <View style={styles.footer}>
              {pageError ? <Text accessibilityRole="alert" style={styles.pageError}>{pageError}</Text> : null}
              <Pressable accessibilityLabel="Load older activity" accessibilityRole="button" disabled={history.isFetching} onPress={() => void loadMore()} style={styles.secondaryButton}>
                {isLoadingMore ? <ActivityIndicator color={palette.action} /> : <Text style={styles.secondaryButtonText}>{pageError ? 'Try older activity again' : 'Load older activity'}</Text>}
              </Pressable>
            </View>
          ) : null}
        />
      ) : null}
    </View>
  );
}

function HistoryFilter({ value, onChange, styles }: { readonly value: AssetActivityView; readonly onChange: (view: AssetActivityView) => void; readonly styles: ReturnType<typeof createStyles> }) {
  return (
    <View style={styles.filterButton}>
      <Text style={styles.filterLabel}>Show</Text>
      <NativeActionMenu
        accessibilityLabel={`Show History, ${value === 'changes' ? 'Changes' : 'All events'}`}
        groups={historyFilterMenuGroups(value, onChange)}
        trigger={{ kind: 'label', label: value === 'changes' ? 'Changes' : 'All events' }}
      />
    </View>
  );
}

function HistoryRow({ record, onPress, styles }: { readonly record: AssetActivityRecordViewModel; readonly onPress: () => void; readonly styles: ReturnType<typeof createStyles> }) {
  return (
    <View style={styles.row}>
      <Pressable accessibilityHint="Shows exact time and technical details" accessibilityRole="button" onPress={onPress} style={styles.rowMain}>
        <Text style={styles.rowTitle}>{record.title}</Text>
        <Text style={styles.rowSummary}>{record.summary}</Text>
        <Text style={styles.rowMeta}>{record.occurredAtLabel} · {record.actorLabel} · {record.sourceLabel}</Text>
      </Pressable>
    </View>
  );
}

function CenteredState({ label, palette, styles }: { readonly label: string; readonly palette: MobileColorPalette; readonly styles: ReturnType<typeof createStyles> }) {
  return <View style={styles.centerState}><ActivityIndicator color={palette.action} /><Text style={styles.stateMessage}>{label}</Text></View>;
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
    screen: { backgroundColor: colors.background, flex: 1 },
    heading: { backgroundColor: colors.surface, borderBottomColor: colors.border, borderBottomWidth: StyleSheet.hairlineWidth, gap: spacing.md, padding: spacing.md },
    assetTitle: { color: colors.text, fontSize: 17, fontWeight: '700' },
    filterButton: { alignItems: 'center', flexDirection: 'row', justifyContent: 'space-between', minHeight: 44 },
    filterLabel: { color: colors.text, fontSize: 16, fontWeight: '600' },
    dateHeader: { backgroundColor: colors.background, color: colors.textMuted, fontSize: 13, fontWeight: '700', paddingHorizontal: spacing.md, paddingVertical: spacing.sm },
    list: { paddingBottom: spacing.xl },
    emptyList: { flexGrow: 1 },
    row: { alignItems: 'center', backgroundColor: colors.surface, borderBottomColor: colors.border, borderBottomWidth: StyleSheet.hairlineWidth, flexDirection: 'row', minHeight: 76, paddingHorizontal: spacing.md },
    rowMain: { flex: 1, justifyContent: 'center', minHeight: 76, paddingVertical: spacing.sm },
    rowTitle: { color: colors.text, fontSize: 16, fontWeight: '700' },
    rowSummary: { color: colors.text, fontSize: 15, lineHeight: 21, marginTop: 2 },
    rowMeta: { color: colors.textMuted, fontSize: 12, lineHeight: 18, marginTop: 2 },
    centerState: { alignItems: 'center', flex: 1, gap: spacing.sm, justifyContent: 'center', padding: spacing.xl },
    stateTitle: { color: colors.text, fontSize: 20, fontWeight: '800', textAlign: 'center' },
    stateMessage: { color: colors.textMuted, fontSize: 15, lineHeight: 22, textAlign: 'center' },
    primaryButton: { alignItems: 'center', backgroundColor: colors.action, borderRadius: radius.md, justifyContent: 'center', marginTop: spacing.sm, minHeight: 44, paddingHorizontal: spacing.lg },
    primaryButtonText: { color: colors.onAction, fontSize: 15, fontWeight: '800' },
    footer: { alignItems: 'center', gap: spacing.sm, padding: spacing.md },
    pageError: { color: colors.danger, fontSize: 14, textAlign: 'center' },
    secondaryButton: { alignItems: 'center', borderColor: colors.border, borderRadius: radius.md, borderWidth: 1, justifyContent: 'center', minHeight: 44, minWidth: 200, paddingHorizontal: spacing.md },
    secondaryButtonText: { color: colors.action, fontSize: 15, fontWeight: '700' }
  });
}
