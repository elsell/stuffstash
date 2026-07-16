import { StyleSheet } from 'react-native';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';

export function createHomeScreenStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
    shell: {
      backgroundColor: colors.background,
      flex: 1
    },
    content: {
      paddingBottom: spacing.xl,
      paddingHorizontal: spacing.lg,
      paddingTop: spacing.md
    },
    centerState: {
      alignItems: 'center',
      flex: 1,
      justifyContent: 'center',
      padding: spacing.lg
    },
    stateText: {
      color: colors.textMuted,
      fontSize: 16,
      marginTop: spacing.md,
      textAlign: 'center'
    },
    errorTitle: {
      color: colors.text,
      fontSize: 24,
      fontWeight: '700',
      letterSpacing: 0
    },
    retryButton: {
      alignItems: 'center',
      backgroundColor: colors.action,
      borderRadius: radius.md,
      justifyContent: 'center',
      marginTop: spacing.lg,
      minHeight: 44,
      minWidth: 96,
      paddingHorizontal: spacing.md
    },
    retryButtonText: {
      color: colors.onAction,
      fontSize: 16,
      fontWeight: '600'
    },
    homeTopBar: {
      alignItems: 'flex-start',
      flexDirection: 'row',
      gap: spacing.sm,
      marginBottom: spacing.lg
    },
    contextControl: {
      alignItems: 'center',
      flex: 1,
      flexDirection: 'row',
      gap: spacing.sm,
      minHeight: 48,
      paddingRight: spacing.sm
    },
    contextText: {
      flex: 1,
      minWidth: 0
    },
    contextInventory: {
      color: colors.text,
      fontSize: 18,
      fontWeight: '700',
      letterSpacing: 0
    },
    contextTenantPrefix: {
      color: colors.textMuted,
      fontSize: 13,
      fontWeight: '500',
      letterSpacing: 0
    },
    topBarActions: {
      alignItems: 'center',
      alignSelf: 'flex-start',
      flexDirection: 'row',
      gap: spacing.xs
    },
    settingsButton: {
      alignItems: 'center',
      borderRadius: 22,
      justifyContent: 'center',
      minHeight: 44,
      minWidth: 44
    },
    sectionHeader: {
      alignItems: 'center',
      flexDirection: 'row',
      justifyContent: 'space-between',
      marginBottom: spacing.sm
    },
    sectionTitle: {
      color: colors.text,
      flexShrink: 1,
      fontSize: 19,
      fontWeight: '700',
      letterSpacing: 0
    },
    sectionActionButton: {
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: 44,
      minWidth: 44,
      paddingLeft: spacing.md
    },
    sectionAction: {
      color: colors.action,
      fontSize: 14,
      fontWeight: '600',
      letterSpacing: 0
    },
    recentTicker: {
      marginBottom: spacing.lg
    },
    attentionSection: {
      marginTop: spacing.xs
    },
    emptyText: {
      color: colors.textMuted,
      fontSize: 15
    },
    returnSheet: {
      backgroundColor: colors.background,
      flex: 1,
      gap: spacing.md,
      padding: spacing.lg
    },
    returnSheetHeader: {
      gap: spacing.xs
    },
    returnSheetTitle: {
      color: colors.text,
      fontSize: 24,
      fontWeight: '700',
      letterSpacing: 0
    },
    returnSheetSubtitle: {
      color: colors.textMuted,
      fontSize: 15
    },
    returnDetailsInput: {
      backgroundColor: colors.surface,
      borderColor: colors.controlBorder,
      borderRadius: radius.md,
      borderWidth: 1,
      color: colors.text,
      flex: 1,
      fontSize: 16,
      minHeight: 160,
      padding: spacing.md
    },
    returnSheetActions: {
      flexDirection: 'row',
      gap: spacing.sm
    },
    returnSheetButton: {
      alignItems: 'center',
      borderRadius: radius.md,
      flex: 1,
      justifyContent: 'center',
      minHeight: 52,
      paddingHorizontal: spacing.md
    },
    returnSheetCancelButton: {
      backgroundColor: colors.surface,
      borderColor: colors.controlBorder,
      borderWidth: 1
    },
    returnSheetSaveButton: {
      backgroundColor: colors.action
    },
    returnSheetCancelText: {
      color: colors.text,
      fontSize: 15,
      fontWeight: '600',
      letterSpacing: 0
    },
    returnSheetSaveText: {
      color: colors.onAction,
      fontSize: 15,
      fontWeight: '600',
      letterSpacing: 0
    }
  });
}
