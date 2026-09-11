import { AssetBreadcrumbTrail } from '../components/AssetCard';
import { formatAssetExpiration } from '../presentation/ExpirationPresentation';
import { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import type { NotificationInboxQueries } from '../../application/notifications/NotificationInboxQueries';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';
import type { ExpirationNotification } from '../../domain/notifications/Notification';
import { NativeSegmentedControl } from '../components/NativeSegmentedControl';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing } from '../theme/tokens';

type Filter = 'all' | 'unread';
export function NotificationInboxScreen({ tenantId, inventoryId, queries, onOpenAsset, onChanged, onSettings }: {
  readonly tenantId: string; readonly inventoryId: string;
  readonly queries: Pick<NotificationInboxQueries, 'list' | 'open' | 'markAllRead'>;
  readonly onOpenAsset: (assetId: string) => void;
  readonly onChanged: () => void;
  readonly onSettings: () => void;
}) {
  const colors = useAppearancePalette();
  const [filter, setFilter] = useState<Filter>('all');
  const [rows, setRows] = useState<ExpirationNotification[]>([]);
  const [cursor, setCursor] = useState<string | null>(null);
  const [locallyRead, setLocallyRead] = useState<ReadonlySet<string>>(new Set());
  const [loaded, setLoaded] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const pending = useRef(false);
  const mounted = useRef(true);
  const controller = useRef<AbortController | null>(null);
  useEffect(() => {
    let active = true; mounted.current = true;
    void Promise.resolve().then(() => { if (active) void load('all'); });
    return () => { active = false; mounted.current = false; controller.current?.abort(); };
  }, []);
  async function run(operation: (signal: AbortSignal) => Promise<void>, fallback: string) {
    if (pending.current) return;
    pending.current = true; setBusy(true); setError('');
    const request = new AbortController(); controller.current = request;
    try { await operation(request.signal); }
    catch (caught) { if (mounted.current && !request.signal.aborted) setError(caught instanceof NotificationFailure ? caught.message : fallback); }
    finally { pending.current = false; if (mounted.current) setBusy(false); }
  }
  async function fetchPage(selected: Filter, signal: AbortSignal, after?: string) {
    const page = await queries.list(tenantId, inventoryId, { unreadOnly: selected === 'unread', cursor: after, limit: 20, signal });
    if (!mounted.current || signal.aborted) return;
    setRows((existing) => after ? [...new Map([...existing, ...page.items].map((row) => [row.id, row])).values()] : page.items);
    setCursor(page.pagination.hasMore ? page.pagination.nextCursor : null); setLoaded(true); setFilter(selected);
  }
  function load(selected: Filter, after?: string) { return run((signal) => fetchPage(selected, signal, after), 'Notifications could not be loaded. Try refreshing.'); }
  function open(row: ExpirationNotification) {
    return run(async (signal) => {
      const assetId = await queries.open(tenantId, inventoryId, row.id, { signal });
      if (mounted.current && !signal.aborted) {
        setLocallyRead((previous) => new Set([...previous, row.id]));
        if (filter === 'unread') setRows((previous) => previous.filter((entry) => entry.id !== row.id));
        onChanged(); onOpenAsset(assetId);
      }
    }, 'This notification could not be opened. Refresh and try again.');
  }
  function markAll() {
    return run(async (signal) => {
      await queries.markAllRead(tenantId, inventoryId, { signal });
      if (!mounted.current || signal.aborted) return;
      onChanged(); await fetchPage(filter, signal);
    }, 'Could not finish marking notifications read. Refresh and try again.');
  }
  const button = (label: string, action: () => void, disabled = busy) => <Pressable accessibilityRole="button" accessibilityLabel={label} disabled={disabled} onPress={action} style={[styles.button, { borderColor: colors.controlBorder, opacity: disabled ? 0.5 : 1 }]}><Text style={{ color: colors.text }}>{label}</Text></Pressable>;
  return <ScrollView style={{ backgroundColor: colors.background }} contentContainerStyle={styles.content}>
    <Text accessibilityRole="header" style={[styles.heading, { color: colors.text }]}>Notifications</Text>
    <NativeSegmentedControl colors={colors} value={filter} disabled={busy} segments={[{ label: 'All', value: 'all' }, { label: 'Unread', value: 'unread' }]} onChange={(value) => void load(value)} />
    <View style={styles.actions}>{button('Refresh notifications', () => void load(filter))}{button('Reminder settings', onSettings)}</View>
    {error ? <Text accessibilityRole="alert" style={{ color: colors.danger }}>{error}</Text> : null}
    {busy ? <ActivityIndicator accessibilityLabel="Updating notifications" color={colors.action} /> : null}
    {rows.map((row) => <View key={row.id} style={[styles.card, { borderColor: colors.controlBorder }]}><Pressable accessibilityRole="button" accessibilityLabel={`Open ${row.title}`} accessibilityValue={{text:`${row.milestone === 'expired' ? 'Expired' : 'Expires'} ${formatAssetExpiration(row.expiration)}. ${row.readAt || locallyRead.has(row.id) ? 'Read' : 'Unread'}`}} disabled={busy} onPress={() => void open(row)}>
      <Text style={[styles.title, { color: colors.text }]}>{row.title}</Text>
      <Text style={{ color: colors.text }}>{row.milestone === 'expired' ? 'Expired' : 'Expires'} {formatAssetExpiration(row.expiration)}</Text>
      <Text style={{ color: colors.textMuted }}>{row.readAt || locallyRead.has(row.id) ? 'Read' : 'Unread'}</Text>
    </Pressable>
      {row.parentTrailIncomplete ? <Text style={{color:colors.textMuted}}>{row.parentTrail?.length ? 'Partial location path' : 'Location unavailable'}</Text> : null}
      <AssetBreadcrumbTrail palette={colors} disabled={busy} segments={(row.parentTrail ?? []).map((entry,index)=>({id:entry.assetId,title:entry.title,isImmediateParent:index===(row.parentTrail?.length ?? 0)-1}))} onSegmentPress={entry=>{if(!busy)onOpenAsset(entry.id);}} />
    </View>)}
    {loaded && !rows.length && !cursor && !error ? <Text style={{ color: colors.textMuted }}>{filter === 'unread' ? 'No unread notifications.' : 'No notifications yet.'}</Text> : null}
    {cursor ? button('Load more notifications', () => void load(filter, cursor)) : null}
    {rows.some((row) => !row.readAt && !locallyRead.has(row.id)) ? button('Mark all read', () => void markAll()) : null}
  </ScrollView>;
}
const styles = StyleSheet.create({
  content: { padding: spacing.lg, gap: spacing.md }, heading: { fontSize: 24, fontWeight: '700' }, title: { fontSize: 18, fontWeight: '600' },
  actions: { gap: spacing.sm }, card: { borderWidth: 1, borderRadius: radius.md, padding: spacing.md, gap: spacing.sm },
  button: { minHeight: 44, borderWidth: 1, borderRadius: radius.md, padding: spacing.sm, justifyContent: 'center' }
});
