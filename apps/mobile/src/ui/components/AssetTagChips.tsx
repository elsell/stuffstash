import { StyleSheet, Text, View } from 'react-native';
import type { AssetTagViewModel } from '../../application/assets/AssetViewModels';
import { colors, radius, spacing } from '../theme/tokens';

type AssetTagChipsProps = {
  readonly tags?: readonly AssetTagViewModel[];
};

export function AssetTagChips({ tags }: AssetTagChipsProps) {
  const visibleTags = tags ?? [];
  if (visibleTags.length === 0) {
    return null;
  }

  return (
    <View accessibilityLabel="Asset tags" style={styles.tagRow}>
      {visibleTags.map((tag) => (
        <View key={tag.id} style={styles.tagChip}>
          {tag.color ? <View style={[styles.tagSwatch, { backgroundColor: tag.color, borderColor: tag.color }]} /> : null}
          <Text numberOfLines={1} style={styles.tagLabel}>{tag.label}</Text>
        </View>
      ))}
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
  tagSwatch: {
    borderRadius: radius.lg,
    borderWidth: 1,
    height: 8,
    width: 8
  },
  tagLabel: {
    color: colors.text,
    flexShrink: 1,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0
  }
});
