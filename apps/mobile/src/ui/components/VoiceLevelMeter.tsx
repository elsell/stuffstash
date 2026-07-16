import { StyleSheet, View } from 'react-native';
import { type MobileColorPalette } from '../theme/tokens';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { computeVoiceLevelBarHeights, type VoiceLevelMeterSize } from './VoiceLevelMeterPresentation';

export function VoiceLevelMeter({
  level,
  size
}: {
  readonly level: number;
  readonly size: VoiceLevelMeterSize;
}) {
  const styles = createStyles(useAppearanceAwarePalette());
  const heights = computeVoiceLevelBarHeights(level, size);
  return (
    <View
      accessible={false}
      pointerEvents="none"
      style={[styles.bars, size === 'compact' ? styles.compactBars : styles.regularBars]}
    >
      {heights.map((height, index) => (
        <View
          key={index.toString()}
          style={[
            styles.bar,
            size === 'compact' ? styles.compactBar : styles.regularBar,
            { height }
          ]}
        />
      ))}
    </View>
  );
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  bar: {
    backgroundColor: colors.onAction,
    borderRadius: 2,
    opacity: 0.86
  },
  bars: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'center'
  },
  compactBar: {
    width: 3
  },
  compactBars: {
    gap: 3,
    height: 18
  },
  regularBar: {
    width: 4
  },
  regularBars: {
    gap: 4,
    height: 22
  }
  });
}
