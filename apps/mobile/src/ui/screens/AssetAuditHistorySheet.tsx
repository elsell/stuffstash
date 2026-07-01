import {
  ActivityIndicator,
  Modal,
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
    <Modal
      animationType="slide"
      transparent
      visible={state.status !== 'closed'}
      onRequestClose={onClose}
    >
      <View style={styles.modalShell}>
        <View style={styles.sheet}>
          <View style={styles.sheetHandle} />
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
      </View>
    </Modal>
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
        <View key={record.id} style={styles.recordCard}>
          <Text style={styles.recordTitle}>{record.title}</Text>
          <Text style={styles.recordSubtitle}>{record.occurredAtLabel}</Text>
          <View style={styles.recordMetaRow}>
            <Text style={styles.recordPill}>{record.sourceLabel}</Text>
            <Text style={styles.recordPill}>{record.principalLabel}</Text>
          </View>
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
      ))}
      {history.hasMore ? (
        <Text style={styles.moreText}>More history is available in the full audit log.</Text>
      ) : null}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  modalShell: {
    backgroundColor: colors.scrim,
    flex: 1,
    justifyContent: 'flex-end'
  },
  sheet: {
    backgroundColor: colors.surface,
    borderTopLeftRadius: radius.lg,
    borderTopRightRadius: radius.lg,
    gap: spacing.md,
    maxHeight: '88%',
    padding: spacing.lg
  },
  sheetHandle: {
    alignSelf: 'center',
    backgroundColor: colors.border,
    borderRadius: radius.sm,
    height: 5,
    width: 44
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
    gap: spacing.sm,
    paddingBottom: spacing.lg
  },
  recordCard: {
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    gap: spacing.xs,
    padding: spacing.md
  },
  recordTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '900',
    letterSpacing: 0
  },
  recordSubtitle: {
    color: colors.textMuted,
    fontSize: 13,
    lineHeight: 18
  },
  recordMetaRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.xs,
    marginTop: spacing.xs
  },
  recordPill: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    color: colors.text,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  recordFinePrint: {
    color: colors.textMuted,
    fontSize: 12,
    lineHeight: 18
  },
  metadataList: {
    borderTopColor: colors.border,
    borderTopWidth: 1,
    gap: spacing.xs,
    marginTop: spacing.sm,
    paddingTop: spacing.sm
  },
  metadataRow: {
    gap: 2
  },
  metadataLabel: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  metadataValue: {
    color: colors.text,
    fontSize: 14,
    lineHeight: 20
  },
  moreText: {
    color: colors.textMuted,
    fontSize: 13,
    lineHeight: 19,
    textAlign: 'center'
  }
});
