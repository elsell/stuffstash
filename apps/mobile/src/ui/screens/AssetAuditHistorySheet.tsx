import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import type { AssetAuditHistoryViewModel } from '../../application/assets/AssetAuditHistoryQuery';
import { colors, radius, spacing } from '../theme/tokens';

export type AssetAuditHistorySheetState =
  | { readonly status: 'closed' }
  | { readonly status: 'loading'; readonly assetTitle: string }
  | { readonly status: 'ready'; readonly assetTitle: string; readonly history: AssetAuditHistoryViewModel }
  | { readonly status: 'error'; readonly assetTitle: string; readonly message: string };

export function AssetAuditHistorySheet({
  onClose,
  state
}: {
  readonly onClose: () => void;
  readonly state: AssetAuditHistorySheetState;
}) {
  return (
    <View style={styles.sheet}>
      <View style={styles.headerRow}>
        <View style={styles.headerText}>
          <Text style={styles.sheetTitle}>Audit history</Text>
          <Text style={styles.sheetSubtitle}>
            {state.status === 'closed' ? 'Asset history' : state.assetTitle}
          </Text>
        </View>
        <Pressable accessibilityRole="button" onPress={onClose} style={styles.closeButton}>
          <Text style={styles.closeButtonText}>Close</Text>
        </Pressable>
      </View>
      {state.status === 'loading' ? <LoadingHistory /> : null}
      {state.status === 'error' ? <ErrorHistory message={state.message} /> : null}
      {state.status === 'ready' ? <ReadyHistory history={state.history} /> : null}
    </View>
  );
}

function LoadingHistory() {
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={colors.action} />
      <Text style={styles.stateText}>Loading history</Text>
    </View>
  );
}

function ErrorHistory({ message }: { readonly message: string }) {
  return (
    <View style={styles.centerState}>
      <Text style={styles.errorTitle}>Could not load history</Text>
      <Text style={styles.stateText}>{message}</Text>
    </View>
  );
}

function ReadyHistory({ history }: { readonly history: AssetAuditHistoryViewModel }) {
  if (history.records.length === 0) {
    return (
      <View style={styles.centerState}>
        <Text style={styles.emptyTitle}>{history.emptyTitle}</Text>
        <Text style={styles.stateText}>{history.emptyMessage}</Text>
      </View>
    );
  }

  return (
    <ScrollView contentContainerStyle={styles.recordList}>
      {history.records.map((record) => (
        <View key={record.id} style={styles.recordRow}>
          <View style={styles.timelineDot} />
          <View style={styles.recordContent}>
            <Text style={styles.recordTitle}>{record.title}</Text>
            <Text style={styles.recordSubtitle}>{record.subtitle}</Text>
            {record.requestLabel ? <Text style={styles.recordFinePrint}>{record.requestLabel}</Text> : null}
            {record.metadataRows.length > 0 ? (
              <View style={styles.metadataList}>
                {record.metadataRows.map((row) => (
                  <View key={row.label} style={styles.metadataRow}>
                    <Text style={styles.metadataLabel}>{row.label}</Text>
                    <Text style={styles.metadataValue}>{row.value}</Text>
                  </View>
                ))}
              </View>
            ) : null}
          </View>
        </View>
      ))}
      {history.hasMore ? (
        <Text style={styles.moreText}>More history is available in the full audit log.</Text>
      ) : null}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  sheet: {
    backgroundColor: colors.surface,
    flex: 1,
    gap: spacing.md,
    padding: spacing.lg,
    paddingTop: spacing.xl
  },
  headerRow: {
    alignItems: 'flex-start',
    flexDirection: 'row',
    gap: spacing.md
  },
  headerText: {
    flex: 1
  },
  sheetTitle: {
    color: colors.text,
    fontSize: 26,
    fontWeight: '900',
    letterSpacing: 0
  },
  sheetSubtitle: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '700',
    letterSpacing: 0,
    lineHeight: 20,
    marginTop: spacing.xs
  },
  closeButton: {
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  closeButtonText: {
    color: colors.text,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0
  },
  centerState: {
    alignItems: 'center',
    gap: spacing.sm,
    minHeight: 220,
    justifyContent: 'center',
    padding: spacing.lg
  },
  stateText: {
    color: colors.textMuted,
    fontSize: 15,
    lineHeight: 22,
    textAlign: 'center'
  },
  errorTitle: {
    color: colors.text,
    fontSize: 18,
    fontWeight: '900',
    letterSpacing: 0
  },
  emptyTitle: {
    color: colors.text,
    fontSize: 20,
    fontWeight: '900',
    letterSpacing: 0
  },
  recordList: {
    gap: spacing.xs,
    paddingBottom: spacing.lg
  },
  recordRow: {
    flexDirection: 'row',
    gap: spacing.sm,
    paddingVertical: spacing.sm
  },
  timelineDot: {
    backgroundColor: colors.action,
    borderRadius: 4,
    height: 8,
    marginTop: 7,
    width: 8
  },
  recordContent: {
    borderBottomColor: colors.border,
    borderBottomWidth: 1,
    flex: 1,
    gap: 3,
    paddingBottom: spacing.sm
  },
  recordTitle: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  recordSubtitle: {
    color: colors.textMuted,
    fontSize: 12,
    lineHeight: 17
  },
  recordFinePrint: {
    color: colors.textMuted,
    fontSize: 11,
    lineHeight: 16
  },
  metadataList: {
    gap: spacing.xs,
    marginTop: spacing.xs
  },
  metadataRow: {
    alignItems: 'baseline',
    flexDirection: 'row',
    gap: spacing.sm
  },
  metadataLabel: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '900',
    letterSpacing: 0,
    minWidth: 92,
    textTransform: 'uppercase'
  },
  metadataValue: {
    color: colors.text,
    flex: 1,
    fontSize: 13,
    lineHeight: 18
  },
  moreText: {
    color: colors.textMuted,
    fontSize: 13,
    lineHeight: 19,
    textAlign: 'center'
  }
});
