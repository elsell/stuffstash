import { useHeaderHeight } from '@react-navigation/elements';
import { NativeFilterSearch } from './NativeFilterSearch.android';
import { KeyboardAvoidingView, ScrollView, StyleSheet, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { NativeSheetActions } from './NativeSheetActions';
import type { NativeFilterSheetProps } from './NativeFilterSheet.types';
import { useFocusedSheetActions } from './useFocusedSheetActions';

/** A native stack route keeps search and actions outside its flexible scroll body. */
export function NativeFilterSheet({ title, search, children, actions, footerTestID }: NativeFilterSheetProps) {
  const palette = useAppearancePalette();
  const headerHeight = useHeaderHeight();
  const footerActions = useFocusedSheetActions(actions);
  return <KeyboardAvoidingView style={styles.screen} behavior="height" keyboardVerticalOffset={headerHeight}>
    <SafeAreaView edges={['bottom']} style={[styles.screen, { backgroundColor: palette.background }]}>
    {search ? <NativeFilterSearch key={title} {...search} /> : null}
    <ScrollView style={styles.body} contentContainerStyle={styles.content}
      keyboardShouldPersistTaps="handled" keyboardDismissMode="on-drag">
      {children}
    </ScrollView>
    <View testID={footerTestID} style={styles.actions}>
      <NativeSheetActions {...footerActions} keyboardAvoidance="container" />
    </View>
  </SafeAreaView>
  </KeyboardAvoidingView>;
}

const styles = StyleSheet.create({
  screen: { flex: 1 },
  body: { flex: 1 },
  content: { paddingBottom: 20 },
  actions: { paddingHorizontal: 20, paddingVertical: 12 }
});
