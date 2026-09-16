import { useLayoutEffect, useMemo, useRef, useState } from 'react';
import { ScrollView, StyleSheet, Text, View } from 'react-native';
import { Stack, useRouter } from 'expo-router';
import { SafeAreaView, useSafeAreaInsets } from 'react-native-safe-area-context';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { NativeSheetActions } from './NativeSheetActions';
import type { NativeFilterSheetProps } from './NativeFilterSheet.types';

/** Android's native footer follows the current detent instead of expanded body bounds. */
export function NativeFilterSheet({ title, children, actions, footerTestID }: NativeFilterSheetProps) {
  const palette = useAppearancePalette();
  const hasUnderlyingRoute = useRouter().canGoBack();
  const insets = useSafeAreaInsets();
  const [footerHeight, setFooterHeight] = useState(0);
  const current = useRef<typeof actions | undefined>(actions);
  useLayoutEffect(() => {
    current.current = actions;
    return () => { current.current = undefined; };
  }, [actions]);
  const { onApply: _apply, onBack: _back, ...presentation } = actions;
  const signature = JSON.stringify(presentation);
  const renderFooter = useMemo(() => () => <SafeAreaView edges={['bottom']}
      style={{ backgroundColor: palette.background }}
      onLayout={event => setFooterHeight(event.nativeEvent.layout.height)}>
      <View testID={footerTestID} style={styles.actions}>
        <NativeSheetActions {...presentation} keyboardAvoidance="container"
          onApply={() => { const latest = current.current; if (latest && !latest.disabled) latest.onApply(); }}
          onBack={() => { const latest = current.current; if (latest && !latest.secondaryDisabled) latest.onBack(); }} />
      </View>
    </SafeAreaView>, [signature, palette.background, footerTestID]);
  const options = useMemo(() => ({ headerShown: false,
    unstable_sheetFooter: hasUnderlyingRoute ? renderFooter : undefined
  }), [hasUnderlyingRoute, renderFooter]);
  return <>
    <Stack.Screen options={options} />
    <ScrollView style={[styles.body, { backgroundColor: palette.background }]}
      contentContainerStyle={{ paddingTop: hasUnderlyingRoute ? 0 : insets.top, paddingBottom: footerHeight + 20 }}
      scrollIndicatorInsets={{ bottom: footerHeight }} keyboardShouldPersistTaps="handled"
      keyboardDismissMode="on-drag">
      <Text accessibilityRole="header" style={[styles.title, { color: palette.text }]}>{title}</Text>
      {children}
    </ScrollView>
    {!hasUnderlyingRoute ? <View style={styles.rootFooter}>{renderFooter()}</View> : null}
  </>;
}

const styles = StyleSheet.create({
  body: { flex: 1 },
  title: { fontSize: 24, fontWeight: '600', paddingHorizontal: 20, paddingTop: 20 },
  actions: { paddingHorizontal: 20, paddingVertical: 12 },
  rootFooter: { position: 'absolute', bottom: 0, left: 0, right: 0 }
});
