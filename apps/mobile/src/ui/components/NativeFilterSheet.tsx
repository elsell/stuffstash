import { useRef, useState } from 'react';
import { ScrollView, StyleSheet, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { NativeSheetActions } from './NativeSheetActions';
import type { NativeFilterSheetProps } from './NativeFilterSheet.types';
import { useSheetKeyboardInset } from './useSheetKeyboardInset';

/** Keep the native scroll body direct; reserve the measured, opaque action area. */
export function NativeFilterSheet({ children, actions, footerTestID }: NativeFilterSheetProps) {
  const palette = useAppearancePalette();
  const [footerHeight, setFooterHeight] = useState(0);
  const boundaryRef = useRef<View>(null);
  const keyboard = useSheetKeyboardInset(boundaryRef);
  return <>
    <ScrollView automaticallyAdjustKeyboardInsets style={[styles.body, { backgroundColor: palette.background }]}
      contentContainerStyle={{ paddingBottom: footerHeight + 20 }} scrollIndicatorInsets={{ bottom: footerHeight }}
      keyboardShouldPersistTaps="handled" keyboardDismissMode="on-drag" contentInsetAdjustmentBehavior="automatic">
      {children}
    </ScrollView>
    <View ref={boundaryRef} collapsable={false} pointerEvents="none" onLayout={keyboard.measure} style={styles.boundary} />
    <SafeAreaView edges={keyboard.bottomInset > 0 ? [] : ['bottom']}
      onLayout={event => setFooterHeight(event.nativeEvent.layout.height)}
      style={[styles.footer, { bottom: keyboard.bottomInset, backgroundColor: palette.background }]}>
      <View testID={footerTestID} style={styles.actions}>
        <NativeSheetActions {...actions} keyboardAvoidance="container" />
      </View>
    </SafeAreaView>
  </>;
}

const styles = StyleSheet.create({
  body: { flex: 1 },
  boundary: { position: 'absolute', bottom: 0, left: 0, right: 0, height: 0 },
  footer: { position: 'absolute', left: 0, right: 0, bottom: 0 },
  actions: { paddingHorizontal: 20, paddingVertical: 12 }
});
