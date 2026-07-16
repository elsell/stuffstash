import { ChevronDown } from 'lucide-react-native';
import { PlatformColor, Pressable, StyleSheet, View } from 'react-native';
import {
  KeyboardController,
  KeyboardExtender
} from 'react-native-keyboard-controller';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

export function AppKeyboardAccessory() {
  const insets = useSafeAreaInsets();

  return (
    <KeyboardExtender>
      <View
        style={[
          styles.bar,
          {
            paddingRight: Math.max(20, insets.right)
          }
        ]}
      >
        <Pressable
          accessibilityHint="Hides the keyboard without submitting"
          accessibilityLabel="Dismiss keyboard"
          accessibilityRole="button"
          hitSlop={4}
          onPress={() => { void KeyboardController.dismiss(); }}
          style={styles.action}
        >
          <ChevronDown color={PlatformColor('link')} size={24} strokeWidth={2.2} />
        </Pressable>
      </View>
    </KeyboardExtender>
  );
}

const styles = StyleSheet.create({
  bar: {
    alignItems: 'center',
    flexDirection: 'row',
    height: 44,
    justifyContent: 'flex-end',
    width: '100%'
  },
  action: {
    alignItems: 'center',
    height: 44,
    justifyContent: 'center',
    width: 44
  }
});
