import { StyleSheet } from 'react-native';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';

export function createSettingsScreenStyles(
  colors: MobileColorPalette,
  layout: { readonly stacked: boolean }
) {
  return StyleSheet.create({
    shell: { backgroundColor: colors.background, flex: 1 },
    content: { paddingBottom: spacing.xl },
    section: { marginTop: spacing.lg },
    sectionTitle: {
      color: colors.textMuted,
      fontSize: 13,
      fontWeight: '600',
      marginBottom: spacing.xs,
      marginHorizontal: spacing.lg,
      textTransform: 'uppercase'
    },
    sectionFooter: {
      color: colors.textMuted,
      fontSize: 13,
      lineHeight: 18,
      marginHorizontal: spacing.lg,
      marginTop: spacing.xs
    },
    group: {
      backgroundColor: colors.surface,
      borderRadius: radius.md,
      marginHorizontal: spacing.md,
      overflow: 'hidden'
    },
    navigationRow: {
      justifyContent: 'center',
      minHeight: 52,
      minWidth: 44,
      paddingHorizontal: spacing.md,
      paddingVertical: spacing.sm
    },
    navigationRowPressed: { backgroundColor: colors.selected },
    navigationRowContent: {
      alignItems: layout.stacked ? 'flex-start' : 'center',
      flexDirection: layout.stacked ? 'column' : 'row',
      gap: layout.stacked ? spacing.xs : spacing.sm
    },
    rowIconFrame: {
      alignItems: 'center',
      borderRadius: radius.sm,
      height: 30,
      justifyContent: 'center',
      width: 30
    },
    rowText: { flex: 1, minWidth: 0 },
    rowLabel: { color: colors.text, fontSize: 17, fontWeight: '500' },
    rowContext: { color: colors.textMuted, fontSize: 13, marginTop: 2 },
    rowValue: { color: colors.textMuted, fontSize: 16 },
    rowTrailing: { alignItems: 'center', flexDirection: 'row', gap: spacing.xs },
    separator: { backgroundColor: colors.border, height: StyleSheet.hairlineWidth, marginLeft: 56 },
    iconButton: { alignItems: 'center', justifyContent: 'center', minHeight: 44, minWidth: 44 },
    choiceRow: { justifyContent: 'center', minHeight: 52, minWidth: 44, paddingHorizontal: spacing.md, paddingVertical: spacing.sm },
    actionRow: { justifyContent: 'center', minHeight: 52, minWidth: 44, paddingHorizontal: spacing.md, paddingVertical: spacing.sm },
    actionGroup: { flexDirection: layout.stacked ? 'column' : 'row', gap: spacing.sm },
    choiceGroup: { flexDirection: layout.stacked ? 'column' : 'row', gap: spacing.sm },
    detailHeader: { paddingHorizontal: spacing.lg, paddingTop: spacing.lg },
    detailTitle: { color: colors.text, fontSize: 24, fontWeight: '700' },
    detailSubtitle: { color: colors.textMuted, fontSize: 15, lineHeight: 21, marginTop: spacing.xs },
    errorContainer: { flexGrow: 1, justifyContent: 'center', padding: spacing.lg },
    errorTitle: { color: colors.text, fontSize: 22, fontWeight: '700', textAlign: 'center' },
    errorMessage: { color: colors.textMuted, fontSize: 16, marginTop: spacing.sm, textAlign: 'center' },
    retryButton: {
      alignItems: 'center',
      alignSelf: 'center',
      backgroundColor: colors.action,
      borderRadius: radius.md,
      justifyContent: 'center',
      marginTop: spacing.md,
      minHeight: 44,
      paddingHorizontal: spacing.lg
    },
    retryText: { color: colors.onAction, fontSize: 16, fontWeight: '600' },
    dangerText: { color: colors.danger, fontSize: 17 },
    actionText: { color: colors.action, fontSize: 17 },
    valueText: { color: colors.text, fontSize: 17 },
    secondaryText: { color: colors.textMuted, fontSize: 14, lineHeight: 20 }
  });
}
