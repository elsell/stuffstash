import { StyleSheet } from 'react-native';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { inventoryMapGestureConfig } from './InventoryMapPresentation';

export function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  shell: {
    backgroundColor: colors.background,
    flex: 1
  },
  header: {
    paddingHorizontal: spacing.md,
    paddingTop: spacing.sm
  },
  headerTopRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.md,
    marginBottom: spacing.xs
  },
  headerActions: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.xs
  },
  headerAddButton: {
    alignItems: 'center',
    borderRadius: 22,
    justifyContent: 'center',
    minHeight: 44,
    minWidth: 44
  },
  titleBlock: {
    flex: 1,
    minWidth: 0
  },
  title: {
    color: colors.text,
    fontSize: 25,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 30
  },
  searchBar: {
    alignItems: 'center',
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 44,
    paddingHorizontal: spacing.sm
  },
  searchInput: {
    color: colors.text,
    flex: 1,
    fontSize: 15,
    minHeight: 44,
    paddingVertical: 0
  },
  iconButton: {
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 32,
    minWidth: 32
  },
  breadcrumbs: {
    alignItems: 'center',
    paddingTop: spacing.xs
  },
  breadcrumbItem: {
    alignItems: 'center',
    flexDirection: 'row'
  },
  breadcrumbButton: {
    justifyContent: 'center',
    minHeight: 44,
    maxWidth: 150,
    paddingHorizontal: spacing.xs
  },
  breadcrumbButtonPressed: {
    opacity: 0.62
  },
  breadcrumbText: {
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0
  },
  overviewText: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0,
    marginTop: 0
  },
  mapScroller: {
    flex: 1,
    marginTop: spacing.xs,
    overflow: 'hidden'
  },
  mapContent: {
    alignItems: 'stretch',
    flexDirection: 'row',
    height: '100%',
    paddingBottom: spacing.xl,
    paddingTop: spacing.sm
  },
  column: {
    backgroundColor: 'transparent',
    flexShrink: 0,
    height: '100%',
    overflow: 'hidden'
  },
  columnTitle: {
    color: colors.text,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0,
    paddingHorizontal: spacing.xs,
    paddingVertical: spacing.sm
  },
  columnList: {
    gap: spacing.xs,
    paddingTop: 2
  },
  columnListSurface: {
    flex: 1
  },
  mapRow: {
    alignItems: 'center',
    backgroundColor: 'transparent',
    borderColor: 'transparent',
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    minHeight: 74,
    overflow: 'hidden',
    position: 'relative',
    shadowColor: '#000000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.04,
    shadowRadius: 4
  },
  mapRowExpanded: {
    backgroundColor: colors.brandDustyBlueSoft,
    borderColor: colors.border
  },
  mapRowHighlighted: {
    borderColor: colors.focusRing,
    shadowColor: colors.focusRing,
    shadowOpacity: 0.22,
    shadowRadius: 8
  },
  rowSwipeUnderlay: {
    alignItems: 'center',
    backgroundColor: colors.action,
    bottom: 0,
    justifyContent: 'center',
    position: 'absolute',
    right: 0,
    top: 0,
    width: inventoryMapGestureConfig.branchSwipeRevealWidth
  },
  rowCard: {
    alignItems: 'center',
    alignSelf: 'stretch',
    backgroundColor: colors.elevatedSurface,
    flex: 1,
    flexDirection: 'row',
    minWidth: 0
  },
  rowMainGesture: {
    alignSelf: 'stretch',
    flex: 1,
    minWidth: 0
  },
  rowMain: {
    alignItems: 'center',
    flex: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 72,
    minWidth: 0,
    paddingLeft: spacing.sm,
    paddingVertical: spacing.xs
  },
  rowImageWrap: {
    position: 'relative'
  },
  rowImageFrame: {
    alignItems: 'center',
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.sm,
    height: 52,
    justifyContent: 'center',
    overflow: 'hidden',
    width: 52
  },
  childCountBadge: {
    alignItems: 'center',
    backgroundColor: colors.accentStrong,
    borderColor: colors.elevatedSurface,
    borderRadius: 9,
    borderWidth: 1,
    bottom: -3,
    flexDirection: 'row',
    gap: 2,
    minHeight: 18,
    paddingHorizontal: 5,
    position: 'absolute',
    right: -4
  },
  childCountBadgeText: {
    color: colors.onAction,
    fontSize: 10,
    fontWeight: '900',
    letterSpacing: 0
  },
  rowImage: {
    height: '100%',
    width: '100%'
  },
  rowImageLabel: {
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0
  },
  rowText: {
    flex: 1,
    gap: 2,
    minWidth: 0
  },
  rowTitleLine: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.xs
  },
  rowTitle: {
    color: colors.text,
    flex: 1,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  rowMeta: {
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '800',
    letterSpacing: 0
  },
  rowTrail: {
    color: colors.textMuted,
    fontSize: 12,
    lineHeight: 16
  },
  rowInfoButton: {
    alignItems: 'center',
    alignSelf: 'stretch',
    backgroundColor: colors.elevatedSurface,
    justifyContent: 'center',
    minWidth: 48
  },
  emptyColumn: {
    alignItems: 'center',
    gap: spacing.xs,
    justifyContent: 'center',
    minHeight: 130,
    padding: spacing.md
  },
  emptyColumnText: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '800',
    letterSpacing: 0,
    textAlign: 'center'
  },
  emptyColumnAction: {
    alignItems: 'center',
    backgroundColor: colors.elevatedSurface,
    borderColor: colors.border,
    borderRadius: 999,
    borderWidth: StyleSheet.hairlineWidth,
    flexDirection: 'row',
    gap: spacing.xs,
    marginTop: spacing.xs,
    minHeight: 40,
    paddingHorizontal: spacing.md
  },
  emptyColumnActionText: {
    color: colors.action,
    fontSize: 14,
    fontWeight: '900',
    letterSpacing: 0
  },
  centerState: {
    alignItems: 'center',
    flex: 1,
    justifyContent: 'center',
    padding: spacing.lg
  },
  centerText: {
    color: colors.textMuted,
    fontSize: 15,
    fontWeight: '700',
    lineHeight: 22,
    marginTop: spacing.sm,
    textAlign: 'center'
  },
  errorTitle: {
    color: colors.text,
    fontSize: 20,
    fontWeight: '900',
    letterSpacing: 0
  },
  sheet: {
    backgroundColor: colors.background,
    flex: 1,
    padding: spacing.lg
  },
  sheetHandle: {
    alignSelf: 'center',
    backgroundColor: colors.border,
    borderRadius: 2,
    height: 4,
    marginBottom: spacing.xs,
    width: 44
  },
  sheetTopBar: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'flex-end',
    minHeight: 44
  },
  sheetLoadingState: {
    alignItems: 'center',
    flex: 1,
    justifyContent: 'center',
    padding: spacing.lg
  },
  sheetCloseButton: {
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.xs
  },
  sheetCloseText: {
    color: colors.action,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  }
  });
}
