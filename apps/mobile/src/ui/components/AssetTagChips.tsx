import { StyleSheet, Text, View } from 'react-native';
import type { AssetTagViewModel } from '../../application/assets/AssetViewModels';
import { colors, radius, spacing } from '../theme/tokens';
import { assetTagChipLayoutPresentation, assetTagChipPresentation, assetTagChipStylePresentation } from './AssetTagChipsPresentation';

type AssetTagChipsProps = {
  readonly tags?: readonly AssetTagViewModel[];
  readonly compact?: boolean;
  readonly overflowLimit?: number;
};

export function AssetTagChips({ tags, compact = false, overflowLimit }: AssetTagChipsProps) {
  const presentation = assetTagChipPresentation(tags, overflowLimit);
  const layout = assetTagChipLayoutPresentation(compact);
  if (!presentation.shouldRender) {
    return null;
  }

  return (
    <View accessibilityLabel="Asset tags" style={[styles.tagRow, layout.compactRow ? styles.compactTagRow : null]}>
      {presentation.visibleTags.map((tag) => {
        const colorStyle = assetTagChipStylePresentation(tag);
        return (
          <View
            key={tag.id}
            style={[
              styles.tagChip,
              colorStyle.colored ? { backgroundColor: colorStyle.backgroundColor, borderColor: colorStyle.borderColor } : null,
              layout.shrinkVisibleChips ? styles.compactTagChip : null
            ]}
          >
            <Text numberOfLines={1} style={styles.tagLabel}>{tag.label}</Text>
          </View>
        );
      })}
      {presentation.hiddenCount > 0 ? (
        <View accessibilityLabel={`${presentation.hiddenCount} more tags`} style={[styles.tagChip, styles.overflowChip]}>
          <Text numberOfLines={1} style={[styles.tagLabel, styles.overflowLabel]}>+{presentation.hiddenCount}</Text>
        </View>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  tagRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.xs,
    minHeight: 0
  },
  compactTagRow: {
    flexWrap: 'nowrap',
    overflow: 'hidden'
  },
  tagChip: {
    alignItems: 'center',
    backgroundColor: colors.surfaceMuted,
    borderColor: colors.border,
    borderRadius: radius.lg,
    borderWidth: 1,
    flexDirection: 'row',
    gap: 5,
    maxWidth: '100%',
    minHeight: 24,
    paddingHorizontal: spacing.xs,
    paddingVertical: 3
  },
  compactTagChip: {
    flexShrink: 1,
    minWidth: 0
  },
  tagLabel: {
    color: colors.text,
    flexShrink: 1,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0
  },
  overflowChip: {
    flexShrink: 0
  },
  overflowLabel: {
    color: colors.textMuted
  }
});
