import { Stack } from 'expo-router';
import { CheckCheck, Settings, Mail, MailOpen } from 'lucide-react-native';
import { AssetBreadcrumbTrail } from '../components/AssetCard';
import { formatAssetExpiration } from '../presentation/ExpirationPresentation';
import { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, Pressable, RefreshControl, ScrollView, StyleSheet, Text, View } from 'react-native';
import type { NotificationInboxQueries } from '../../application/notifications/NotificationInboxQueries';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';
import type { ExpirationNotification } from '../../domain/notifications/Notification';
import { NativeSegmentedControl } from '../components/NativeSegmentedControl';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing } from '../theme/tokens';

type Filter = 'all' | 'unread';
export function NotificationInboxScreen({ tenantId, inventoryId, queries, onOpenAsset, onChanged, onSettings }: {
  readonly tenantId: string; readonly inventoryId: string;
  readonly queries: Pick<NotificationInboxQueries, 'list' | 'open' | 'markAllRead' | 'setRead'>;
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
    setLocallyRead(previous => new Set([...previous].filter(id => !page.items.some(row => row.id === id))));
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
  function toggleRead(row: ExpirationNotification) {
    const read = !!row.readAt || locallyRead.has(row.id);
    return run(async (signal) => {
      await queries.setRead(tenantId, inventoryId, row.id, !read, { signal });
      if (!mounted.current || signal.aborted) return;
      setLocallyRead((previous) => { const next = new Set(previous); next.delete(row.id); return next; });
      onChanged();
      await fetchPage(filter, signal);
    }, 'Could not update this notification. Try again.');
  }
  const button = (label: string, action: () => void, disabled = busy) => <Pressable accessibilityRole="button" accessibilityLabel={label} disabled={disabled} onPress={action} style={[styles.button, { borderColor: colors.controlBorder, opacity: disabled ? 0.5 : 1 }]}><Text style={{ color: colors.text }}>{label}</Text></Pressable>;
  return <>
    <Stack.Screen options={{ title: 'Notifications', headerRight: () => <View style={{ flexDirection: 'row' }}>
      <Pressable accessibilityRole="button" accessibilityLabel="Mark all read" disabled={busy || (!cursor && !rows.some(row => !row.readAt && !locallyRead.has(row.id)))} onPress={() => void markAll()} style={styles.toolbarButton}><CheckCheck size={22} color={colors.action} /></Pressable>
      <Pressable accessibilityRole="button" accessibilityLabel="Reminder settings" onPress={onSettings} style={styles.toolbarButton}><Settings size={22} color={colors.action} /></Pressable>
    </View> }} />
    <ScrollView style={{ backgroundColor: colors.background }} contentContainerStyle={styles.content} alwaysBounceVertical contentInsetAdjustmentBehavior="automatic" refreshControl={<RefreshControl refreshing={busy && loaded} onRefresh={() => void load(filter)} tintColor={colors.action} />}>
    <NativeSegmentedControl colors={colors} value={filter} disabled={busy} segments={[{ label: 'All', value: 'all' }, { label: 'Unread', value: 'unread' }]} onChange={(value) => void load(value)} />

    {error ? <View><Text accessibilityRole="alert" style={{ color: colors.danger }}>{error}</Text>{button('Retry notifications', () => void load(filter))}</View> : null}
    {busy && !loaded ? <ActivityIndicator accessibilityLabel="Updating notifications" color={colors.action} /> : null}
    {rows.map((row) => <View key={row.id} style={[styles.card, { borderColor: colors.border }]}><Pressable accessibilityRole="button" accessibilityLabel={`Open ${row.title}`} accessibilityValue={{text:`${row.milestone === 'expired' ? 'Expired' : 'Expires'} ${formatAssetExpiration(row.expiration)}. ${row.readAt || locallyRead.has(row.id) ? 'Read' : 'Unread'}`}} disabled={busy} onPress={() => void open(row)}>
      <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}>
        {!row.readAt && !locallyRead.has(row.id) ? <View accessibilityElementsHidden style={{ width: 8, height: 8, borderRadius: 4, backgroundColor: colors.action }} /> : null}
        <Text style={[styles.title, { color: colors.text, fontWeight: row.readAt || locallyRead.has(row.id) ? '400' : '600', flexShrink: 1 }]}>{row.title}</Text>
      </View>
      <Text style={{ color: colors.text }}>{row.milestone === 'expired' ? 'Expired' : 'Expires'} {formatAssetExpiration(row.expiration)}</Text>

    </Pressable>
      <Pressable accessibilityRole="button" accessibilityLabel={`Mark ${row.title} ${row.readAt || locallyRead.has(row.id) ? 'unread' : 'read'}`} disabled={busy} onPress={() => void toggleRead(row)} style={styles.readAction}>
        {row.readAt || locallyRead.has(row.id) ? <Mail size={20} color={colors.action} /> : <MailOpen size={20} color={colors.action} />}
      </Pressable>
      {row.parentTrailIncomplete ? <Text style={{color:colors.textMuted}}>{row.parentTrail?.length ? 'Partial location path' : 'Location unavailable'}</Text> : null}
      <AssetBreadcrumbTrail palette={colors} disabled={busy} segments={(row.parentTrail ?? []).map((entry,index)=>({id:entry.assetId,title:entry.title,isImmediateParent:index===(row.parentTrail?.length ?? 0)-1}))} onSegmentPress={entry=>{if(!busy)onOpenAsset(entry.id);}} />
    </View>)}
    {loaded && !rows.length && !cursor && !error ? <Text style={{ color: colors.textMuted }}>{filter === 'unread' ? 'No unread notifications.' : 'No notifications yet.'}</Text> : null}
    {cursor ? button('Load more notifications', () => void load(filter, cursor)) : null}
  </ScrollView></>;
}
const styles = StyleSheet.create({
  content: { padding: spacing.lg, gap: spacing.md }, heading: { fontSize: 24, fontWeight: '700' }, title: { fontSize: 18, fontWeight: '600' },
  actions: { gap: spacing.sm }, card: { borderBottomWidth: StyleSheet.hairlineWidth, paddingVertical: spacing.md, paddingRight: 44, gap: spacing.sm },
  toolbarButton: { minWidth: 44, minHeight: 44, alignItems: 'center', justifyContent: 'center' },
  readAction: { position: 'absolute', right: 0, top: spacing.md, minWidth: 44, minHeight: 44, alignItems: 'center', justifyContent: 'center' },
  button: { minHeight: 44, padding: spacing.sm, justifyContent: 'center' }
});
