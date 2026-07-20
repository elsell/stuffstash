import { useEffect, useState } from 'react';
import { ChevronDown } from 'lucide-react-native';
import { Keyboard, PlatformColor, Pressable, StyleSheet, View } from 'react-native';
import { KeyboardExtender } from 'react-native-keyboard-controller';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

export function AppKeyboardAccessory() {
  const insets = useSafeAreaInsets();
  const [keyboardVisible, setKeyboardVisible] = useState(Keyboard.isVisible());

  useEffect(() => {
    const shown = Keyboard.addListener('keyboardWillShow', () => setKeyboardVisible(true));
    const hidden = Keyboard.addListener('keyboardDidHide', () => setKeyboardVisible(false));
    return () => {
      shown.remove();
      hidden.remove();
    };
  }, []);

  return (
    <View
      accessibilityElementsHidden={!keyboardVisible}
      pointerEvents={keyboardVisible ? 'box-none' : 'none'}
      style={styles.host}
      testID="app-keyboard-accessory-host"
    >
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
            onPress={Keyboard.dismiss}
            style={styles.action}
          >
            <ChevronDown color={PlatformColor('link')} size={24} strokeWidth={2.2} />
          </Pressable>
        </View>
      </KeyboardExtender>
    </View>
  );
}

const styles = StyleSheet.create({
  host: {
    bottom: 0,
    left: 0,
    position: 'absolute',
    right: 0
  },
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
