import { Pressable, StyleSheet, Text, View } from 'react-native';
import type { StyleProp, ViewStyle } from 'react-native';
import type { AssetTagViewModel } from '../../application/assets/AssetViewModels';
import { lightPalette, radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { assetTagChipLayoutPresentation, assetTagChipPresentation, assetTagChipStylePresentation } from './AssetTagChipsPresentation';

type AssetTagChipsProps = {
  readonly tags?: readonly AssetTagViewModel[];
  readonly compact?: boolean;
  readonly overflowLimit?: number;
  readonly onTagPress?: (tag: AssetTagViewModel) => void;
  readonly palette?: MobileColorPalette;
};

export function AssetTagChips({ tags, compact = false, overflowLimit, onTagPress, palette: paletteOverride }: AssetTagChipsProps) {
  const appearancePalette = useAppearanceAwarePalette();
  const palette = paletteOverride ?? appearancePalette;
  const styles = createStyles(palette);
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
          <TagChip
            key={tag.id}
            tag={tag}
            onTagPress={onTagPress}
            palette={palette}
            style={[
              styles.tagChip,
              colorStyle.colored ? { backgroundColor: colorStyle.backgroundColor, borderColor: colorStyle.borderColor } : null,
              layout.shrinkVisibleChips ? styles.compactTagChip : null
            ]}
          />
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

export function TagChip({
  tag,
  onTagPress,
  palette,
  style
}: {
  readonly tag: AssetTagViewModel;
  readonly onTagPress?: (tag: AssetTagViewModel) => void;
  readonly palette?: MobileColorPalette;
  readonly style: StyleProp<ViewStyle>;
}) {
  const styles = createStyles(palette ?? lightPalette);
  if (!onTagPress) {
    return (
      <View style={style}>
        <Text numberOfLines={1} style={styles.tagLabel}>{tag.label}</Text>
      </View>
    );
  }

  return (
    <Pressable
      accessibilityLabel={`Search for tag ${tag.label}`}
      accessibilityRole="button"
      hitSlop={6}
      onPress={() => onTagPress(tag)}
      style={style}
    >
      <Text numberOfLines={1} style={styles.tagLabel}>{tag.label}</Text>
    </Pressable>
  );
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
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
    // The six-point hit slop brings interactive chips to a 44-point target.
    minHeight: 32,
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
}
