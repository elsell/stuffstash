import { useCallback, useState } from 'react';
import { router, Stack, useFocusEffect } from 'expo-router';
import {
  ActionSheetIOS,
  ActivityIndicator,
  Alert,
  Platform,
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
import { groupHistoryRecords, historyLoadError } from './AssetHistoryPresentation';
import { useAppFeedback } from '../feedback/AppFeedback';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';

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
  const [state, setState] = useState<HistoryState>({ status: 'loading' });
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [pageError, setPageError] = useState<string>();

  const load = useCallback(async (selectedView: AssetActivityView) => {
    setState({ status: 'loading' });
    setPageError(undefined);
    try {
      const result = await assetActivityQuery.execute({ tenantId, inventoryId, assetId, view: selectedView, limit: 20 });
      setState({ status: 'ready', ...result });
    } catch (error) {
      setState({ status: 'error', ...historyLoadError(error) });
    }
  }, [assetActivityQuery, assetId, inventoryId, tenantId]);

  useFocusEffect(useCallback(() => {
    void load(view);
  }, [load, view]));

  async function refresh(): Promise<void> {
    setIsRefreshing(true);
    try {
      const result = await assetActivityQuery.execute({ tenantId, inventoryId, assetId, view, limit: 20 });
      setState({ status: 'ready', ...result });
    } catch (error) {
      feedback.showNotice({ tone: 'error', title: 'Could not refresh History', message: readableError(error, 'Your existing History is still shown.') });
    } finally {
      setIsRefreshing(false);
    }
  }

  async function loadMore(): Promise<void> {
    if (state.status !== 'ready' || !state.hasMore || !state.nextCursor || isLoadingMore) return;
    setIsLoadingMore(true);
    setPageError(undefined);
    try {
      const result = await assetActivityQuery.execute({
        tenantId, inventoryId, assetId, view, limit: 20, cursor: state.nextCursor
      });
      setState({ ...result, status: 'ready', records: [...state.records, ...result.records] });
    } catch (error) {
      setPageError(readableError(error, 'Older activity could not be loaded.'));
    } finally {
      setIsLoadingMore(false);
    }
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
          {state.canRetry ? <Pressable accessibilityRole="button" onPress={() => void load(view)} style={styles.primaryButton}>
            <Text style={styles.primaryButtonText}>Try again</Text>
          </Pressable> : null}
        </View>
      ) : null}
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
              <Pressable accessibilityRole="button" disabled={isLoadingMore} onPress={() => void loadMore()} style={styles.secondaryButton}>
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
  function choose(): void {
    const options = ['Changes', 'All events', 'Cancel'];
    const select = (index?: number) => {
      if (index === 0) onChange('changes');
      if (index === 1) onChange('all');
    };
    if (Platform.OS === 'ios') {
      ActionSheetIOS.showActionSheetWithOptions({ options, cancelButtonIndex: 2, title: 'Show History' }, select);
      return;
    }
    Alert.alert('Show History', undefined, [
      { text: 'Changes', onPress: () => onChange('changes') },
      { text: 'All events', onPress: () => onChange('all') },
      { text: 'Cancel', style: 'cancel' }
    ]);
  }
  return (
    <Pressable accessibilityHint="Choose Changes or All events" accessibilityRole="button" onPress={choose} style={styles.filterButton}>
      <Text style={styles.filterLabel}>Show</Text>
      <Text style={styles.filterValue}>{value === 'changes' ? 'Changes' : 'All events'} ›</Text>
    </Pressable>
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

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
    screen: { backgroundColor: colors.background, flex: 1 },
    heading: { backgroundColor: colors.surface, borderBottomColor: colors.border, borderBottomWidth: StyleSheet.hairlineWidth, gap: spacing.md, padding: spacing.md },
    assetTitle: { color: colors.text, fontSize: 17, fontWeight: '700' },
    filterButton: { alignItems: 'center', flexDirection: 'row', justifyContent: 'space-between', minHeight: 44 },
    filterLabel: { color: colors.text, fontSize: 16, fontWeight: '600' },
    filterValue: { color: colors.action, fontSize: 16 },
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
