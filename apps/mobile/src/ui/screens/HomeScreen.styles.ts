import { StyleSheet } from 'react-native';
import { colors, radius, spacing } from '../theme/tokens';

export const styles = StyleSheet.create({
  shell: {
    flex: 1,
    backgroundColor: colors.background
  },
  content: {
    padding: spacing.lg,
    paddingBottom: spacing.xl
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
    lineHeight: 23,
    marginTop: spacing.md,
    textAlign: 'center'
  },
  errorTitle: {
    color: colors.text,
    fontSize: 24,
    fontWeight: '800',
    letterSpacing: 0
  },
  homeTopBar: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.sm,
    marginBottom: spacing.lg
  },
  contextControl: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flex: 1,
    flexDirection: 'row',
    gap: spacing.md,
    minHeight: 66,
    padding: spacing.md
  },
  settingsButton: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    height: 52,
    justifyContent: 'center',
    width: 52
  },
  contextText: {
    flex: 1,
    minWidth: 0
  },
  contextLine: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.xs,
    minWidth: 0,
    flexWrap: 'wrap'
  },
  contextInventory: {
    color: colors.text,
    fontSize: 18,
    fontWeight: '900',
    letterSpacing: 0
  },
  contextTenantPrefix: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '700',
    letterSpacing: 0
  },
  statsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm,
    marginBottom: spacing.lg
  },
  statTile: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    minHeight: 78,
    padding: spacing.md,
    width: '47%'
  },
  statValue: {
    color: colors.accentStrong,
    fontSize: 28,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 34
  },
  statLabel: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '700',
    letterSpacing: 0,
    marginTop: spacing.xs
  },
  sectionTitle: {
    color: colors.text,
    fontSize: 19,
    fontWeight: '800',
    letterSpacing: 0,
  },
  sectionHeader: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: spacing.sm
  },
  sectionAction: {
    color: colors.action,
    fontSize: 14,
    fontWeight: '800',
    letterSpacing: 0
  },
  recentTicker: {
    gap: spacing.sm,
    paddingBottom: spacing.lg
  },
  recentCard: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    minHeight: 210,
    overflow: 'hidden',
    width: 164
  },
  recentImageFrame: {
    alignItems: 'center',
    aspectRatio: 1,
    backgroundColor: colors.surfaceMuted,
    justifyContent: 'center',
    width: '100%'
  },
  recentImagePlaceholder: {
    color: colors.accentStrong,
    fontSize: 22,
    fontWeight: '900',
    letterSpacing: 0
  },
  recentImage: {
    height: '100%',
    width: '100%'
  },
  recentBody: {
    gap: spacing.xs,
    padding: spacing.sm
  },
  emptyText: {
    color: colors.textMuted,
    fontSize: 15,
    lineHeight: 22
  },
  checkedOutEmpty: {
    paddingBottom: spacing.md
  },
  returnButton: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.sm,
    justifyContent: 'center',
    margin: spacing.sm,
    minHeight: 40
  },
  returnButtonText: {
    color: colors.surface,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0
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
    fontWeight: '900',
    letterSpacing: 0
  },
  returnSheetSubtitle: {
    color: colors.textMuted,
    fontSize: 15,
    lineHeight: 22
  },
  returnDetailsInput: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    color: colors.text,
    flex: 1,
    fontSize: 16,
    lineHeight: 22,
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
    borderColor: colors.border,
    borderWidth: 1
  },
  returnSheetSaveButton: {
    backgroundColor: colors.action
  },
  returnSheetCancelText: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  returnSheetSaveText: {
    color: colors.surface,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  locationGrid: {
    gap: spacing.sm,
    marginBottom: spacing.lg
  },
  locationCard: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    overflow: 'hidden'
  },
  locationImageFrame: {
    alignItems: 'center',
    aspectRatio: 16 / 9,
    backgroundColor: colors.surfaceMuted,
    justifyContent: 'center',
    width: '100%'
  },
  locationImage: {
    height: '100%',
    width: '100%'
  },
  locationImagePlaceholder: {
    color: colors.accentStrong,
    fontSize: 22,
    fontWeight: '900',
    letterSpacing: 0
  },
  locationTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '900',
    letterSpacing: 0,
    paddingHorizontal: spacing.md,
    paddingTop: spacing.md
  },
  locationDescription: {
    color: colors.textMuted,
    fontSize: 14,
    lineHeight: 20,
    marginTop: spacing.xs,
    paddingHorizontal: spacing.md
  },
  locationFooter: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginTop: spacing.md,
    paddingHorizontal: spacing.md
  },
  locationCount: {
    color: colors.accentStrong,
    fontSize: 13,
    fontWeight: '800',
    letterSpacing: 0
  },
  recentAssetLabel: {
    color: colors.textMuted,
    fontSize: 12,
    lineHeight: 17,
    marginTop: spacing.xs,
    paddingBottom: spacing.md,
    paddingHorizontal: spacing.md
  },
  detailStack: {
    gap: spacing.md,
    marginBottom: spacing.lg
  },
  photoPanel: {
    alignItems: 'center',
    aspectRatio: 4 / 3,
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    justifyContent: 'center',
    overflow: 'hidden'
  },
  photoKind: {
    color: colors.accentStrong,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  photoTitle: {
    color: colors.textMuted,
    fontSize: 16,
    fontWeight: '700',
    letterSpacing: 0,
    marginTop: spacing.xs
  },
  detailPanel: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    padding: spacing.md
  },
  detailTitle: {
    color: colors.text,
    fontSize: 26,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 32
  },
  metadataList: {
    borderTopColor: colors.border,
    borderTopWidth: 1,
    gap: spacing.sm,
    marginTop: spacing.md,
    paddingTop: spacing.md
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
    fontSize: 15,
    fontWeight: '700',
    letterSpacing: 0,
    lineHeight: 21
  },
  actionRow: {
    borderTopColor: colors.border,
    borderTopWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    marginTop: spacing.md,
    paddingTop: spacing.md
  },
  primaryAction: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  primaryActionText: {
    color: colors.onAction,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0
  },
  secondaryAction: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  secondaryActionText: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '800',
    letterSpacing: 0
  },
  disabledAction: {
    opacity: 0.55
  },
  assetRow: {
    alignItems: 'flex-start',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.md,
    justifyContent: 'space-between',
    marginBottom: spacing.sm,
    minHeight: 98,
    padding: spacing.md
  },
  assetRowSelected: {
    borderColor: colors.action
  },
  assetText: {
    flex: 1,
    minWidth: 0
  },
  badgeRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.xs,
    marginBottom: spacing.xs
  },
  kindBadge: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  customTypeBadge: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.sm,
    borderWidth: 1,
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  checkoutBadge: {
    backgroundColor: colors.warningSurface,
    borderRadius: radius.sm,
    color: colors.warning,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  assetTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '800',
    letterSpacing: 0,
    lineHeight: 23
  },
  assetDescription: {
    color: colors.textMuted,
    fontSize: 14,
    lineHeight: 20,
    marginTop: 2
  },
  assetMeta: {
    color: colors.textMuted,
    fontSize: 12,
    lineHeight: 18,
    marginTop: spacing.xs
  },
  photoReady: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  photoNeeded: {
    backgroundColor: colors.warningSurface,
    borderRadius: radius.sm,
    color: colors.warning,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0,
    overflow: 'hidden',
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  }
});
