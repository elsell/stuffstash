import { useEffect, useRef, useState } from 'react';
import { router, Stack } from 'expo-router';
import { Alert, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { AssetActivityQuery } from '../../application/assets/AssetActivityQuery';
import { RevertAssetChangeCommand } from '../../application/assets/RevertAssetChangeCommand';
import { useAppFeedback } from '../feedback/AppFeedback';
import { historyLoadError, technicalDetailRows } from './AssetHistoryPresentation';
import { applyHistoryRevert, requestHistoryRevertConfirmation } from './AssetHistoryRevertAction';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';

export function AssetHistoryDetailRouteScreen({
  assetActivityQuery,
  revertAssetChangeCommand,
  activityId,
  assetId,
  tenantId,
  inventoryId,
  assetTitle
}: {
  readonly assetActivityQuery: AssetActivityQuery;
  readonly revertAssetChangeCommand: RevertAssetChangeCommand;
  readonly activityId: string;
  readonly assetId: string;
  readonly tenantId: string;
  readonly inventoryId: string;
  readonly assetTitle: string;
}) {
  const activityScope = { tenantId, inventoryId, assetId, activityId };
  const [entry, setEntry] = useState(() => assetActivityQuery.cachedEntry(activityScope));
  const [isLoading, setIsLoading] = useState(!entry);
  const [loadFailure, setLoadFailure] = useState<ReturnType<typeof historyLoadError>>();
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  const feedback = useAppFeedback();
  const [showsTechnical, setShowsTechnical] = useState(false);
  const [isReverting, setIsReverting] = useState(false);
  const isRevertingRef = useRef(false);
  const [isRevertUnavailable, setIsRevertUnavailable] = useState(false);

  useEffect(() => {
    if (entry) return;
    let current = true;
    assetActivityQuery.loadEntry({ tenantId, inventoryId, assetId, activityId })
      .then((loaded) => { if (current) setEntry(loaded); })
      .catch((error) => { if (current) setLoadFailure(historyLoadError(error)); })
      .finally(() => { if (current) setIsLoading(false); });
    return () => { current = false; };
  }, [activityId, assetActivityQuery, assetId, entry, inventoryId, tenantId]);

  async function retryLoad(): Promise<void> {
    setIsLoading(true);
    setLoadFailure(undefined);
    try {
      const loaded = await assetActivityQuery.loadEntry({ tenantId, inventoryId, assetId, activityId });
      setEntry(loaded);
    } catch (error) {
      setLoadFailure(historyLoadError(error));
    } finally {
      setIsLoading(false);
    }
  }

  if (isLoading) {
    return <View style={styles.centerState}><Stack.Screen options={{ title: 'History detail' }} /><Text style={styles.muted}>Loading activity…</Text></View>;
  }

  if (loadFailure) {
    return (
      <View style={styles.centerState}>
        <Stack.Screen options={{ title: 'History detail' }} />
        <Text accessibilityRole="header" style={styles.title}>{loadFailure.title}</Text>
        <Text style={styles.muted}>{loadFailure.message}</Text>
        {loadFailure.canRetry ? <Pressable accessibilityRole="button" onPress={() => void retryLoad()} style={styles.button}>
          <Text style={styles.buttonText}>Try again</Text>
        </Pressable> : null}
      </View>
    );
  }

  if (!entry) {
    return (
      <View style={styles.centerState}>
        <Stack.Screen options={{ title: 'History detail' }} />
        <Text accessibilityRole="header" style={styles.title}>Activity is no longer available</Text>
        <Text style={styles.muted}>Return to History and open it again.</Text>
        <Pressable accessibilityRole="button" onPress={() => router.back()} style={styles.button}>
          <Text style={styles.buttonText}>Back to History</Text>
        </Pressable>
      </View>
    );
  }

  function confirmRevert(): void {
    if (!entry?.undo || entry.undo.status !== 'available') return;
    requestHistoryRevertConfirmation(
      entry,
      (confirmation, confirm) => Alert.alert(confirmation.title, confirmation.message, [
        { text: 'Cancel', style: 'cancel' },
        { text: confirmation.confirmLabel, onPress: confirm }
      ]),
      () => void revertChange()
    );
  }

  async function revertChange(): Promise<void> {
    if (!entry?.undo || entry.undo.status !== 'available' || isRevertingRef.current) return;
    isRevertingRef.current = true;
    setIsReverting(true);
    const result = await applyHistoryRevert(
      revertAssetChangeCommand,
      { tenantId, inventoryId, operationId: entry.undo.operationId },
      {
        invalidateActivity: () => assetActivityQuery.invalidateEntry(activityScope),
        showSuccess: () => feedback.showNotice({ tone: 'success', title: 'Change reverted', message: `“${assetTitle}” was updated. The reversal is now in History.` }),
        navigateBack: () => router.back()
      }
    );
    if (result.status === 'failed') {
      const failure = result.failure;
      if (failure.isTerminal) setIsRevertUnavailable(true);
      feedback.showNotice({ tone: 'error', title: failure.title, message: failure.message });
    }
    isRevertingRef.current = false;
    setIsReverting(false);
  }

  return (
    <ScrollView contentContainerStyle={styles.content} style={styles.screen}>
      <Stack.Screen options={{ title: 'History detail' }} />
      <View style={styles.section}>
        <Text accessibilityRole="header" style={styles.title}>{detailTitle(entry.action)}</Text>
        <Text style={styles.timestamp}>{exactTime(entry.occurredAt)}</Text>
        <Text style={styles.muted}>{entry.principal?.email?.trim() || entry.principalId || 'Someone with access'} · {sourceLabel(entry.source)}</Text>
      </View>

      {entry.changes.length > 0 ? (
        <View style={styles.section}>
          <Text accessibilityRole="header" style={styles.sectionTitle}>What changed</Text>
          {entry.changes.map((change, index) => (
            <View key={`${change.field}-${index.toString()}`} style={styles.detailRow}>
              <Text style={styles.label}>{fieldLabel(change.field)}</Text>
              <Text style={styles.value}>{changeSummary(change.previousValue, change.currentValue)}</Text>
            </View>
          ))}
        </View>
      ) : null}

      {entry.undo?.status === 'available' && !isRevertUnavailable ? (
        <Pressable accessibilityRole="button" accessibilityState={{ disabled: isReverting }} disabled={isReverting} onPress={confirmRevert} style={styles.button}>
          <Text style={styles.buttonText}>{isReverting ? 'Reverting…' : 'Revert change'}</Text>
        </Pressable>
      ) : null}
      {isRevertUnavailable ? <Text accessibilityRole="alert" style={styles.muted}>This change can no longer be safely reverted.</Text> : null}

      <View style={styles.section}>
        <Pressable accessibilityRole="button" accessibilityState={{ expanded: showsTechnical }} onPress={() => setShowsTechnical((value) => !value)} style={styles.disclosureButton}>
          <Text style={styles.sectionTitle}>Technical details</Text><Text style={styles.disclosureText}>{showsTechnical ? '−' : '+'}</Text>
        </Pressable>
        {showsTechnical ? <>
          {technicalDetailRows(entry).map((row) => <TechnicalRow key={row.label} label={row.label} value={row.value} styles={styles} />)}
        </> : null}
      </View>
    </ScrollView>
  );
}

function TechnicalRow({ label, value, styles }: { readonly label: string; readonly value: string; readonly styles: ReturnType<typeof createStyles> }) {
  return <View style={styles.detailRow}><Text style={styles.label}>{label}</Text><Text selectable style={styles.technicalValue}>{value}</Text></View>;
}

function detailTitle(action: string): string {
  switch (action) {
    case 'asset.created': return 'Item added';
    case 'asset.archived': return 'Item archived';
    case 'asset.restored': return 'Item restored';
    case 'asset.checked_out': return 'Item checked out';
    case 'asset.returned': return 'Item returned';
    case 'asset.viewed': return 'Item viewed';
    default: return action.startsWith('asset.') ? 'Item updated' : 'Item activity';
  }
}

function fieldLabel(field: string): string {
  switch (field) {
    case 'title': return 'Name';
    case 'description': return 'Description';
    case 'tags': return 'Tags';
    case 'parent': return 'Location';
    case 'lifecycle_state': return 'Status';
    case 'checkout_state': return 'Checkout';
    default: return field;
  }
}

function changeSummary(previousValue: string | undefined, currentValue: string | undefined): string {
  const previous = previousValue?.trim();
  const current = currentValue?.trim();
  if (!previous && !current) return 'Changed';
  return `${previous || 'None'} → ${current || 'None'}`;
}

function sourceLabel(source: string): string {
  if (source === 'api') return 'App';
  if (source === 'conversation' || source === 'voice') return 'Voice';
  if (source === 'import') return 'Import';
  return 'Stuff Stash';
}

function exactTime(value: string): string {
  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) return value;
  return new Intl.DateTimeFormat('en-US', { dateStyle: 'long', timeStyle: 'long' }).format(new Date(timestamp));
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
    screen: { backgroundColor: colors.background, flex: 1 },
    content: { gap: spacing.md, padding: spacing.md, paddingBottom: spacing.xl },
    section: { backgroundColor: colors.surface, borderRadius: radius.md, gap: spacing.sm, padding: spacing.md },
    title: { color: colors.text, fontSize: 24, fontWeight: '800' },
    timestamp: { color: colors.text, fontSize: 16, lineHeight: 23 },
    muted: { color: colors.textMuted, fontSize: 14, lineHeight: 21 },
    sectionTitle: { color: colors.text, fontSize: 17, fontWeight: '800' },
    detailRow: { borderTopColor: colors.border, borderTopWidth: StyleSheet.hairlineWidth, gap: 3, paddingTop: spacing.sm },
    label: { color: colors.textMuted, fontSize: 13, fontWeight: '700' },
    value: { color: colors.text, fontSize: 16, lineHeight: 23 },
    technicalValue: { color: colors.text, fontFamily: 'Courier', fontSize: 13, lineHeight: 19 },
    disclosureButton: { alignItems: 'center', flexDirection: 'row', justifyContent: 'space-between', minHeight: 44 },
    disclosureText: { color: colors.action, fontSize: 22 },
    button: { alignItems: 'center', backgroundColor: colors.action, borderRadius: radius.md, justifyContent: 'center', minHeight: 48, paddingHorizontal: spacing.md },
    buttonText: { color: colors.onAction, fontSize: 16, fontWeight: '800' },
    centerState: { alignItems: 'center', backgroundColor: colors.background, flex: 1, gap: spacing.md, justifyContent: 'center', padding: spacing.xl }
  });
}
