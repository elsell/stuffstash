import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import type { AssetCheckoutHistoryViewModel } from '../../application/assets/AssetCheckoutHistoryQuery';
import { colors, radius, spacing } from '../theme/tokens';

export type AssetCheckoutHistorySheetState =
  | { readonly status: 'closed' }
  | { readonly status: 'loading'; readonly assetTitle: string }
  | { readonly status: 'ready'; readonly assetTitle: string; readonly history: AssetCheckoutHistoryViewModel }
  | { readonly status: 'error'; readonly assetTitle: string; readonly message: string };

export function AssetCheckoutHistorySheet({
  onClose,
  state
}: {
  readonly onClose: () => void;
  readonly state: AssetCheckoutHistorySheetState;
}) {
  return (
    <View style={styles.sheet}>
      <View style={styles.headerRow}>
        <View style={styles.headerText}>
          <Text style={styles.sheetTitle}>Checkout history</Text>
          <Text numberOfLines={1} style={styles.sheetSubtitle}>
            {state.status === 'closed' ? 'Asset checkout history' : state.assetTitle}
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
      <Text style={styles.stateText}>Loading checkout history</Text>
    </View>
  );
}

function ErrorHistory({ message }: { readonly message: string }) {
  return (
    <View style={styles.centerState}>
      <Text style={styles.errorTitle}>Could not load checkout history</Text>
      <Text style={styles.stateText}>{message}</Text>
    </View>
  );
}

function ReadyHistory({ history }: { readonly history: AssetCheckoutHistoryViewModel }) {
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
          <View style={styles.timelineRail}>
            <View style={[
              styles.timelineDot,
              record.statusLabel === 'Checked out' ? styles.timelineDotOpen : null
            ]} />
          </View>
          <View style={styles.recordContent}>
            <View style={styles.recordTitleRow}>
              <Text style={styles.recordTitle}>{record.title}</Text>
              <Text style={styles.statusPill}>{record.statusLabel}</Text>
            </View>
            <Text style={styles.recordSubtitle}>{record.subtitle}</Text>
            {record.returnedLabel ? <Text style={styles.recordFinePrint}>{record.returnedLabel}</Text> : null}
            {record.checkoutDetails ? (
              <View style={styles.detailBlock}>
                <Text style={styles.detailLabel}>Checkout details</Text>
                <Text style={styles.detailValue}>{record.checkoutDetails}</Text>
              </View>
            ) : null}
            {record.returnDetails ? (
              <View style={styles.detailBlock}>
                <Text style={styles.detailLabel}>Return details</Text>
                <Text style={styles.detailValue}>{record.returnDetails}</Text>
              </View>
            ) : null}
          </View>
        </View>
      ))}
      {history.hasMore ? (
        <Text style={styles.moreText}>More checkout history is available.</Text>
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
    justifyContent: 'center',
    minHeight: 220,
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
    letterSpacing: 0,
    textAlign: 'center'
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
  timelineRail: {
    alignItems: 'center',
    width: 12
  },
  timelineDot: {
    backgroundColor: colors.border,
    borderRadius: 5,
    height: 10,
    marginTop: 7,
    width: 10
  },
  timelineDotOpen: {
    backgroundColor: colors.action
  },
  recordContent: {
    borderBottomColor: colors.border,
    borderBottomWidth: 1,
    flex: 1,
    gap: 5,
    paddingBottom: spacing.md
  },
  recordTitleRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.sm
  },
  recordTitle: {
    color: colors.text,
    flex: 1,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  statusPill: {
    backgroundColor: colors.surfaceMuted,
    borderColor: colors.border,
    borderRadius: radius.sm,
    borderWidth: 1,
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '900',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: 3
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
  detailBlock: {
    backgroundColor: colors.surfaceMuted,
    borderColor: colors.border,
    borderRadius: radius.sm,
    borderWidth: 1,
    gap: 3,
    marginTop: spacing.xs,
    padding: spacing.sm
  },
  detailLabel: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  detailValue: {
    color: colors.text,
    fontSize: 13,
    lineHeight: 18
  },
  moreText: {
    color: colors.textMuted,
    fontSize: 12,
    lineHeight: 18,
    paddingVertical: spacing.sm,
    textAlign: 'center'
  }
});
