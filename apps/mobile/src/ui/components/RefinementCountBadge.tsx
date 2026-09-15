import { StyleSheet, Text, View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';

/** Supplementary count; the refinement command owns its accessible name. */
export function RefinementCountBadge({ count }: { readonly count?: number }) {
  const palette = useAppearanceAwarePalette();
  if (!count || count <= 0) return null;
  return <View
    accessible={false}
    importantForAccessibility="no-hide-descendants"
    pointerEvents="none"
    style={[styles.badge, { backgroundColor: palette.action }]}
  >
    <Text style={[styles.text, { color: palette.onAction }]}>{count > 9 ? '9+' : count.toString()}</Text>
  </View>;
}

const styles = StyleSheet.create({
  badge: {
    alignItems: 'center',
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 16,
    minWidth: 16,
    paddingHorizontal: 3,
    position: 'absolute',
    right: -3,
    top: -3
  },
  text: { fontSize: 10, fontWeight: '700', lineHeight: 12 }
});
